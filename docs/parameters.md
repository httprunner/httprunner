# HttpRunner 参数化功能 (Parameters)

## 概述

HttpRunner 支持强大的**数据驱动测试**能力，允许用户在**测试用例（Testcase）**和**测试步骤（Step）**两个层级上进行参数化。这使得测试用例可以与外部数据文件解耦，实现更灵活、可维护性更高的自动化测试。

- **测试用例层级参数化**：对整个测试流程使用多组不同的数据重复执行。适用于需要验证完整业务流程的场景，例如使用不同用户登录并执行相同操作。
- **测试步骤层级参数化**：仅在单个步骤内使用不同的参数重复执行。适用于需要验证单个功能点的场景，例如在搜索框中输入不同的关键词。

## 测试用例层级参数化 (TestCase-Level)

当您需要使用不同的数据集完整地运行整个测试流程时，应使用测试用例层级的参数化。

### 使用方法

通过在 `hrp.TestCase` 的 `Parameters` 字段中定义参数，并可选择使用 `WithParametersSetting` 进行策略配置。

```go
// testcase_parameters_test.go
func TestTestcaseParameters(t *testing.T) {
    testcase := &hrp.TestCase{
        Config: hrp.NewConfig("测试用例层级参数化").
            WithParameters(map[string]interface{}{
                "username-password": [][]interface{}{
                    {"user1", "pass1"},
                    {"user2", "pass2"},
                    {"user3", "pass3"},
                },
            }).
            WithParametersSetting(
                hrp.WithRandomOrder(), // 随机选择参数
                hrp.WithLimit(2),      // 只执行2次
            ),
        TestSteps: []hrp.IStep{
            hrp.NewStep("登录").
                POST("/api/login").
                WithBody(map[string]interface{}{
                    "username": "$username",
                    "password": "$password",
                }),
            hrp.NewStep("获取用户信息").
                GET("/api/user/info"),
        },
    }

    err := hrp.NewRunner(t).Run(testcase)
    assert.Nil(t, err)
}
```

### 执行结果

- 整个测试用例将运行两次。
- 第一次运行时，`$username` 为 `user1`，`$password` 为 `pass1`。
- 第二次运行时，`$username` 为 `user2`，`$password` 为 `pass2`。
- 测试报告中会生成两个独立的测试结果，每个结果对应一组参数。

---

## 测试步骤层级参数化 (TestStep-Level)

当您只需要在一个测试流程中对某个特定步骤使用不同参数进行多次测试时，应使用测试步骤层级的参数化。

### 使用场景

- 需要对同一个操作使用不同参数进行多次测试
- 减少重复的步骤定义代码
- 在 UI 自动化测试中，对同一个界面操作使用不同的输入数据

### 核心设计

#### 1. 架构改动

##### StepConfig 结构扩展
```go
type StepConfig struct {
    StepName           string                 `json:"name" yaml:"name"`
    Variables          map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
    Parameters         map[string]interface{} `json:"parameters,omitempty" yaml:"parameters,omitempty"`          // 新增
    ParametersSetting  *TParamsConfig         `json:"parameters_setting,omitempty" yaml:"parameters_setting,omitempty"` // 新增
    Loops              int                    `json:"loops,omitempty" yaml:"loops,omitempty"`                   // 已弃用，建议使用 Loop() 方法
    // ... 其他字段保持不变
}
```

##### TParamsConfig 参数配置
```go
type TParamsConfig struct {
    PickOrder  iteratorPickOrder           `json:"pick_order,omitempty" yaml:"pick_order,omitempty"`
    Strategies map[string]IteratorStrategy `json:"strategies,omitempty" yaml:"strategies,omitempty"`
    Limit      int                         `json:"limit,omitempty" yaml:"limit,omitempty"`
}
```

#### 2. Limit 和 Loops 的关系

**重要**：`Limit` 和 `Loops` 不是替代关系，而是可以同时存在的不同控制机制：

1. **Limit（总数限制）**：
   - 作用：限制参数迭代的**总执行次数**
   - 位置：`TParamsConfig.Limit`
   - 示例：`Limit = 4` 表示最多执行4次，无论有多少参数

2. **Loops（循环次数）**：
   - 作用：设置单个 parameter 的**循环使用次数**
   - 位置：`StepConfig.Loops`（通过 `Loop()` 方法设置）
   - 示例：`Loops = 2` 表示每个参数都循环执行2次

##### 执行顺序优化

用户期望的执行顺序（外层循环）：
```go
// 参数：["A", "B"]，Loops = 2
// 执行顺序：A, B, A, B（先遍历所有参数，再进行下一轮循环）
// 步骤命名：[loop_1_params_1], [loop_1_params_2], [loop_2_params_1], [loop_2_params_2]
```

#### 3. 统一执行流程

```
1. 收集所有参数组合
2. 外层循环：循环次数 (1 到 loopTimes)
3. 内层循环：遍历所有参数
4. 执行步骤并收集结果
```

### API 设计

#### 统一的参数配置方法

为了避免代码重复和保持API一致性，参数配置方法只在 `StepRequest` 中定义：

- `WithParameters()` - 设置步骤级参数
- `WithParametersSetting()` - 设置参数选择策略（支持 option 模式）
- `Loop()` - 设置循环次数
- `WithVariables()` - 设置步骤变量

这些方法返回 `*StepRequest`，可以与其他步骤类型（如 Mobile、HTTP 等）配合使用。

#### Option API（推荐）

新的 option API 提供更简洁和灵活的参数配置方式：

```go
// 单个 option
hrp.NewStep("测试").
    WithParameters(map[string]interface{}{
        "query": []interface{}{"成都", "北京"},
    }).
    WithParametersSetting(hrp.WithSequentialOrder()).
    Loop(2).
    Android()

// 多个 options 组合
hrp.NewStep("测试").
    WithParameters(...).
    WithParametersSetting(
        hrp.WithRandomOrder(),           // 设置随机选择顺序
        hrp.WithLimit(4),                // 限制总执行次数
        hrp.WithStrategy("param1", hrp.IteratorStrategy{
            Name:      "custom",
            PickOrder: "random",
        }),
    ).
    Loop(2).
    Android()
```

##### 可用的 Option 函数

###### 参数选择顺序
- `hrp.WithSequentialOrder()` - 设置参数按顺序执行（默认）
- `hrp.WithRandomOrder()` - 设置参数随机选择
- `hrp.WithUniqueOrder()` - 设置参数唯一选择，避免重复

###### 其他配置
- `hrp.WithLimit(limit int)` - 设置参数迭代总数限制
- `hrp.WithStrategy(paramName string, strategy IteratorStrategy)` - 为特定参数设置策略

#### 正确的调用顺序

**重要**：参数配置必须在指定步骤类型之前调用：

```go
// ✅ 推荐：使用 Loop 方法
hrp.NewStep("搜索测试").
    WithParameters(map[string]interface{}{
        "query": []interface{}{"成都", "北京", "重庆"},
    }).
    Loop(3).   // 使用 Loop 方法
    Android().
    StartToGoal("进入搜索框，输入查询词「$query」")

// ✅ 推荐：使用新的 Option API
hrp.NewStep("搜索测试").
    WithParameters(map[string]interface{}{
        "query": []interface{}{"成都", "北京", "重庆"},
    }).
    WithParametersSetting(
        hrp.WithSequentialOrder(),
        hrp.WithLimit(3),
    ).
    Android().
    StartToGoal("进入搜索框，输入查询词「$query」")

// ✅ 兼容：使用传统 TParamsConfig
hrp.NewStep("搜索测试").
    WithParameters(map[string]interface{}{
        "query": []interface{}{"成都", "北京", "重庆"},
    }).
    WithParametersSetting(&hrp.TParamsConfig{
        PickOrder: "sequential",
        Limit:     3,
    }).
    Android().
    StartToGoal("进入搜索框，输入查询词「$query」")

// ❌ 错误：在指定步骤类型后无法配置参数
hrp.NewStep("搜索测试").
    Android().
    WithParameters(...) // 这里会编译错误
```

### 基本用法

#### 1. 简单参数列表

```go
hrp.NewStep("搜索查询词").
    WithParameters(map[string]interface{}{
        "query": []interface{}{"成都", "北京", "重庆"},
    }).
    Android().
    StartToGoal("进入搜索框，输入 query「$query」，等待加载出 sug 提示词")
```

#### 2. 复合参数

```go
hrp.NewStep("登录测试").
    WithParameters(map[string]interface{}{
        "username-password": [][]interface{}{
            {"user1", "pass1"},
            {"user2", "pass2"},
            {"user3", "pass3"},
        },
    }).
    POST("/api/login").
    WithBody(map[string]interface{}{
        "username": "$username",
        "password": "$password",
    })
```

#### 3. 参数策略配置

```go
hrp.NewStep("随机测试数据").
    WithParameters(map[string]interface{}{
        "data": []interface{}{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
    }).
    WithParametersSetting(
        hrp.WithRandomOrder(), // 随机选择
        hrp.WithLimit(3),      // 只执行3次
    ).
    POST("/api/test").
    WithBody(map[string]interface{}{
        "value": "$data",
    })
```

#### 4. 循环配置

```go
// 推荐方式：使用 Loop 方法
hrp.NewStep("循环测试").
    WithParameters(map[string]interface{}{
        "query": []interface{}{"成都", "北京"},
    }).
    Loop(3).  // 每个参数重复执行3次
    Android().
    StartToGoal("搜索「$query」")

// 总执行次数：2 × 3 = 6 次
// 执行顺序：成都, 北京, 成都, 北京, 成都, 北京
```

### 参数策略选项

- `sequential`: 顺序执行（默认）
- `random`: 随机选择
- `unique`: 唯一值，避免重复

### 高级特性

#### 1. 与现有功能兼容

步骤级参数化与现有功能完全兼容：

- ✅ 支持 Loop 循环
- ✅ 支持变量提取和校验
- ✅ 支持 SetupHooks 和 TeardownHooks
- ✅ 支持 failfast 控制

#### 2. 步骤命名规则

- 无参数，无循环：`步骤名称`
- 无参数，有循环：`步骤名称_loop_N`
- 有参数，无循环：`步骤名称 [params_N]`
- 有参数，有循环：`步骤名称 [loop_M_params_N]`（外层循环优化）

#### 3. 变量恢复机制

每次执行完成后，会自动恢复原始的步骤变量，避免副作用影响后续执行。

### 完整示例

```go
func TestStepParametersExample(t *testing.T) {
    testCase := &hrp.TestCase{
        Config: hrp.NewConfig("步骤级参数化示例"),
        TestSteps: []hrp.IStep{
            hrp.NewStep("多查询词搜索测试").
                WithParameters(map[string]interface{}{
                    "query": []interface{}{"成都", "北京", "重庆"},
                }).
                WithParametersSetting(
                    hrp.WithSequentialOrder(),
                    hrp.WithLimit(3),
                ).
                Android().
                StartToGoal("搜索「$query」并等待结果加载").
                AIQuery("提取搜索结果"),
            hrp.NewStep("循环执行测试").
                WithParameters(map[string]interface{}{
                    "data": []interface{}{1, 2, 3},
                }).
                Loop(2).  // 每个参数执行2次
                GET("/api/test").
                WithParams(map[string]interface{}{
                    "value": "$data",
                }),
        },
    }

    // 预期执行：
    // 第一个步骤：成都, 北京, 重庆 = 3次
    // 第二个步骤：1, 2, 3, 1, 2, 3 = 6次

    err := hrp.NewRunner(t).Run(testCase)
    assert.Nil(t, err)
}
```

## 对比与选择

| 特性 | 测试用例层级 (Testcase-Level) | 测试步骤层级 (Step-Level) |
| :--- | :--- | :--- |
| **适用场景** | 完整的业务流程验证 | 单个功能点或API的验证 |
| **数据作用域** | 整个 `TestCase` 中的所有 `Step` | 单个 `Step` |
| **报告结果** | 每组参数生成一个独立的测试报告 | 所有参数在同一个测试报告中 |
| **性能** | 开销较大，每次都需执行完整流程 | 开销较小，仅重复执行单个步骤 |
| **配置方法** | `hrp.TestCase{ Parameters: ... }` | `hrp.Step{}.WithParameters(...)` |

**选择建议**：
- 当您需要用多组数据测试一个完整的端到端流程时（例如，不同的用户配置），请使用**测试用例层级**参数化。
- 当您只需要在一个流程中，针对某个特定步骤进行多次数据测试时（例如，测试接口对不同输入的响应），请使用**测试步骤层级**参数化。

## 变量优先级

HttpRunner 中的变量优先级遵循覆盖原则，顺序如下（从高到低）：

1.  **步骤级参数化变量** (`WithParameters`)
2.  **用例级参数化变量** (`TestCase.Parameters`)
3.  **步骤级变量** (`WithVariables`)
4.  **会话变量** (前面步骤提取的变量)
5.  **配置变量** (`Config.Variables`)

## 注意事项

1. **API 调用顺序**：步骤级参数化方法必须在步骤类型方法（如 `Android()`、`POST()` 等）之前调用
2. **参数继承**：`StepMobile` 等步骤类型继承 `StepConfig`，参数会正确传递
3. **编译时检查**：错误的调用顺序会在编译时报错，确保API使用正确性
4. **变量覆盖**：Parameters 会覆盖同名的 Variables，但不会影响其他变量
5. **变量恢复**：每次执行后自动恢复原始变量，避免副作用
6. **执行顺序**：步骤级参数化采用外层循环优化，先遍历所有参数，再进行下一轮循环
7. **Option API 推荐**：新项目建议使用 option 模式的 API，更加灵活和简洁