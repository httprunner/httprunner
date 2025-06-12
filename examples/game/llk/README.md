# LianLianKan (连连看) Game Bot

基于 HttpRunner @/uixt 模块实现的连连看小游戏自动游玩机器人。

## 功能特性

### 核心功能
- **智能界面分析**: 使用 AI 模型分析游戏界面，自动识别游戏元素类型和位置
- **完整求解算法**: 实现符合连连看规则的完整求解算法，支持直线、一次转弯、两次转弯连接
- **静态分析求解**: 基于初始游戏状态进行静态分析，预先计算所有有效配对
- **跨平台支持**: 支持 Android、iOS、HarmonyOS、Browser 等多种平台

### 连连看算法
- **直线连接**: 检测水平和垂直直线连接（0次转弯）
- **L形连接**: 支持一次转弯的 L 形路径连接（1次转弯）
- **Z形连接**: 支持两次转弯的 Z 形路径连接（2次转弯）
- **路径验证**: 确保连接路径无阻挡
- **游戏规则验证**: 严格按照连连看游戏规则验证配对有效性

## 项目结构

```
examples/game/llk/
├── main.go          # 主要实现文件，包含游戏机器人
├── solver.go        # 连连看求解器实现
├── main_test.go     # 游戏机器人测试
├── solver_test.go   # 求解器测试
├── testdata/        # 测试数据
├── results/         # 运行结果
├── cmd/             # 命令行工具
└── README.md        # 项目说明
```

### 主要组件

#### 数据结构
- `GameElement`: 游戏元素信息，包含维度、元素列表等
- `Element`: 单个游戏元素，包含类型和位置信息
- `Position`: 网格位置，包含行列坐标
- `Dimensions`: 网格维度，包含行数和列数
- `LLKGameBot`: 游戏机器人，集成 XTDriver 和 AI 服务
- `LLKSolver`: 连连看求解器，实现完整的游戏求解逻辑

#### 核心方法

**LLKGameBot 方法**:
- `NewLLKGameBot()`: 创建游戏机器人实例
- `AnalyzeGameInterface()`: 分析游戏界面，提取游戏元素
- `TakeScreenshot()`: 截取屏幕截图
- `SolveGame()`: 求解整个游戏
- `Play()`: 执行游戏操作
- `Close()`: 关闭机器人并清理资源

**LLKSolver 方法**:
- `NewLLKSolver()`: 创建求解器实例
- `FindAllPairs()`: 查找所有有效的匹配对
- `canConnect()`: 检查两个位置是否可以连接
- `canConnectDirect()`: 检查直线连接
- `canConnectWithOneTurn()`: 检查一次转弯连接
- `canConnectWithTwoTurns()`: 检查两次转弯连接

## 环境配置

需要配置 AI 服务密钥：

```bash
# doubao-1.6-seed-250615，用作分析游戏界面
DOUBAO_SEED_1_6_250615_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
DOUBAO_SEED_1_6_250615_API_KEY=<your_api_key>

# doubao-1.5-ui-tars-250328，用作执行游戏操作
DOUBAO_1_5_UI_TARS_250328_BASE_URL=https://ark.cn-beijing.volces.com/api/v3
DOUBAO_1_5_UI_TARS_250328_API_KEY=<your_api_key>

```

## 使用示例

### 基本使用

```go
// 创建游戏机器人
bot, err := NewLLKGameBot("android", "")
if err != nil {
    log.Fatal(err)
}
defer bot.Close()

// 分析游戏界面
gameElement, err := bot.AnalyzeGameInterface()
if err != nil {
    log.Fatal(err)
}

// 创建求解器并查找配对
solver := NewLLKSolver(gameElement)
pairs := solver.FindAllPairs()

// 求解完整游戏
solution, err := bot.SolveGame(gameElement)
if err != nil {
    log.Fatal(err)
}

// 执行游戏
err = bot.Play()
if err != nil {
    log.Fatal(err)
}
```

### 求解器独立使用

```go
// 直接使用求解器
solver := NewLLKSolver(gameElement)
allPairs := solver.FindAllPairs()

// 打印解决方案
for i, pair := range allPairs {
    fmt.Printf("Pair %d: (%d,%d) -> (%d,%d) [%s]\n",
        i+1,
        pair[0].Position.Row, pair[0].Position.Col,
        pair[1].Position.Row, pair[1].Position.Col,
        pair[0].Type)
}
```

## 测试

### 运行测试

```bash
# 运行所有测试
go test -v

# 运行游戏机器人测试
go test -v -run TestLLKGameBot

# 运行求解器测试
go test -v -run TestLLKSolver

# 运行基准测试
go test -v -bench=.
```

### 测试覆盖

- **AI 分析测试**: 测试 AI 模型的界面分析能力
- **求解器测试**: 测试连连看算法的正确性和性能
- **连接规则测试**: 验证各种连接规则的实现
- **完整集成测试**: 测试游戏机器人的完整流程

### 测试数据

项目包含完整的测试数据集，包括：
- 14x8 游戏板，共 112 个元素
- 25 种不同的游戏元素类型
- 完整的求解路径验证

## 技术特点

### AI 集成
- 使用先进的 AI 模型进行图像分析
- 支持结构化输出 Schema
- 自动提取游戏元素的类型、位置、坐标信息
- 支持多种 AI 服务提供商

### 算法优化
- **静态分析**: 基于初始游戏状态进行分析，避免动态状态管理的复杂性
- **完全遵循游戏规则**: 严格按照连连看规则验证连接有效性
- **高效路径检测**: 支持 0-2 次转弯的路径连接算法
- **智能配对查找**: 预先计算所有有效配对，提高执行效率

### 代码质量
- 完整的单元测试覆盖
- 详细的英文代码注释
- 清晰的错误处理和日志记录
- 完善的资源管理和清理
- 模块化设计，职责分离

## 许可证

本项目遵循 HttpRunner 项目的许可证。