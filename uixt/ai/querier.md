# HttpRunner AI Querier - 自定义输出格式功能

## 功能概述

HttpRunner 的 AI Querier 模块支持自定义输出格式功能，允许用户指定特定的数据结构，让 AI 模型返回结构化的数据响应。适用于：

- **UI 元素分析**：自动化测试中的界面元素提取
- **游戏界面分析**：网格类游戏（连连看、消消乐、2048等）数据提取
- **表单数据提取**：从表单截图中提取结构化信息
- **图像内容分析**：任何需要从截图中提取结构化信息的场景

## 核心数据结构

```go
// QueryOptions - 查询选项
type QueryOptions struct {
    Query        string      `json:"query"`                  // 查询文本
    Screenshot   string      `json:"screenshot"`             // Base64编码的截图
    Size         types.Size  `json:"size"`                   // 屏幕尺寸
    OutputSchema interface{} `json:"outputSchema,omitempty"` // 自定义输出格式（可选）
}

// QueryResult - 查询结果
type QueryResult struct {
    Content string      `json:"content"`        // 人类可读的分析结果
    Thought string      `json:"thought"`        // AI 推理过程
    Data    interface{} `json:"data,omitempty"` // 结构化数据（使用OutputSchema时自动转换为指定类型）
}
```

## 基本用法

### 标准查询

```go
// 创建查询器
modelConfig, err := ai.GetModelConfig(option.OPENAI_GPT_4O)
querier, err := ai.NewQuerier(ctx, modelConfig)

// 执行查询
result, err := querier.Query(ctx, &ai.QueryOptions{
    Query:      "请分析这张截图中的内容",
    Screenshot: screenshot,
    Size:       size,
    // 不指定 OutputSchema
})

fmt.Printf("分析结果: %s\n", result.Content)
fmt.Printf("推理过程: %s\n", result.Thought)
// result.Data 为 nil
```

### 自定义格式查询

```go
// 定义输出结构
type GameAnalysis struct {
    Content string   `json:"content"` // 分析描述
    Thought string   `json:"thought"` // 思考过程
    Rows    int      `json:"rows"`    // 行数
    Cols    int      `json:"cols"`    // 列数
    Icons   []string `json:"icons"`   // 图标类型
}

// 执行查询
result, err := querier.Query(ctx, &ai.QueryOptions{
    Query:        "分析这个游戏界面的网格结构和图标类型",
    Screenshot:   screenshot,
    Size:         size,
    OutputSchema: GameAnalysis{}, // 指定输出格式
})

// 直接类型断言获取结构化数据
if gameData, ok := result.Data.(*GameAnalysis); ok {
    fmt.Printf("行数: %d, 列数: %d\n", gameData.Rows, gameData.Cols)
    fmt.Printf("图标类型: %v\n", gameData.Icons)
}
```

## 应用场景示例

### UI 元素分析

```go
type UIAnalysis struct {
    Content  string      `json:"content"`
    Thought  string      `json:"thought"`
    Elements []UIElement `json:"elements"`
}

type UIElement struct {
    Type     string      `json:"type"`        // button, text, input等
    Text     string      `json:"text"`        // 文本内容
    BoundBox BoundingBox `json:"boundBox"`    // 位置坐标
    Clickable bool       `json:"clickable"`   // 是否可点击
}

type BoundingBox struct {
    X, Y, Width, Height int `json:"x,y,width,height"`
}
```

### 网格游戏分析

```go
type GridGame struct {
    Content string     `json:"content"`
    Thought string     `json:"thought"`
    Grid    [][]Cell   `json:"grid"`       // 网格数据
    Stats   Statistics `json:"statistics"` // 统计信息
}

type Cell struct {
    Type  string `json:"type"`  // 单元格类型
    Value string `json:"value"` // 单元格值
    Row   int    `json:"row"`   // 行索引
    Col   int    `json:"col"`   // 列索引
}

type Statistics struct {
    TotalCells  int `json:"totalCells"`
    UniqueTypes int `json:"uniqueTypes"`
}
```

### 表单数据提取

```go
type FormAnalysis struct {
    Content string      `json:"content"`
    Thought string      `json:"thought"`
    Fields  []FormField `json:"fields"`
    Actions []Action    `json:"actions"`
}

type FormField struct {
    Label    string      `json:"label"`    // 字段标签
    Type     string      `json:"type"`     // 字段类型
    Value    string      `json:"value"`    // 当前值
    Required bool        `json:"required"` // 是否必填
    BoundBox BoundingBox `json:"boundBox"` // 位置
}
```

## 核心特性

### 自动类型转换
- 指定 `OutputSchema` 时，`QueryResult.Data` 自动转换为指定类型
- 支持直接类型断言：`result.Data.(*YourType)`
- 无需手动调用转换函数

### 多级回退机制
1. 优先解析为指定的结构化类型
2. 失败时尝试通用JSON解析
3. 最终回退到纯文本响应

### 向后兼容
- 不指定 `OutputSchema` 时行为不变
- 现有代码无需修改

## 最佳实践

### 1. 结构体设计

```go
// 推荐：包含标准字段
type YourSchema struct {
    Content string `json:"content"` // 必须：人类可读描述
    Thought string `json:"thought"` // 必须：AI推理过程
    // 自定义字段...
    Data    CustomData `json:"data"`
}

// 使用描述性的JSON标签
type Element struct {
    Type     string `json:"elementType"`   // 清晰的字段名
    Position Point  `json:"gridPosition"`  // 描述性标签
    Visible  bool   `json:"isVisible"`     // 布尔值清晰性
}
```

### 2. 查询指令

```go
// 推荐：详细的查询指令
opts := &ai.QueryOptions{
    Query: `分析这张截图并提供结构化信息：
1. 识别界面类型和主要元素
2. 提取所有可交互元素的位置和属性
3. 统计各类元素的数量`,
    Screenshot:   screenshot,
    Size:         size,
    OutputSchema: YourSchema{},
}
```

### 3. 错误处理

```go
result, err := querier.Query(ctx, opts)
if err != nil {
    return err
}

// 类型断言
if data, ok := result.Data.(*YourSchema); ok {
    // 使用结构化数据
    processData(data)
} else {
    // 回退到文本结果
    log.Printf("结构化解析失败，使用文本结果: %s", result.Content)
}
```

## 完整示例

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/httprunner/httprunner/v5/internal/builtin"
    "github.com/httprunner/httprunner/v5/uixt/ai"
    "github.com/httprunner/httprunner/v5/uixt/option"
)

type ScreenAnalysis struct {
    Content    string   `json:"content"`
    Thought    string   `json:"thought"`
    Elements   []string `json:"elements"`
    Categories []string `json:"categories"`
    Count      int      `json:"count"`
}

func main() {
    ctx := context.Background()

    // 创建查询器
    modelConfig, err := ai.GetModelConfig(option.OPENAI_GPT_4O)
    if err != nil {
        log.Fatal(err)
    }

    querier, err := ai.NewQuerier(ctx, modelConfig)
    if err != nil {
        log.Fatal(err)
    }

    // 加载截图
    screenshot, size, err := builtin.LoadImage("screenshot.png")
    if err != nil {
        log.Fatal(err)
    }

    // 执行结构化查询
    result, err := querier.Query(ctx, &ai.QueryOptions{
        Query:        "分析截图中的UI元素，提取元素类型和分类信息",
        Screenshot:   screenshot,
        Size:         size,
        OutputSchema: ScreenAnalysis{},
    })
    if err != nil {
        log.Fatal(err)
    }

    // 使用结构化数据
    if analysis, ok := result.Data.(*ScreenAnalysis); ok {
        fmt.Printf("发现 %d 个元素\n", analysis.Count)
        fmt.Printf("元素类型: %v\n", analysis.Elements)
        fmt.Printf("分类: %v\n", analysis.Categories)
    } else {
        fmt.Printf("文本结果: %s\n", result.Content)
    }
}
```

## 辅助函数

对于特殊情况，提供了类型转换辅助函数：

```go
// 手动类型转换（通常不需要）
converted, err := ai.ConvertQueryResultData[YourType](result)
if err != nil {
    return err
}
```

**注意**：使用 `OutputSchema` 时，`Data` 字段已自动转换为正确类型，通常不需要手动调用此函数。

## 技术限制

- 需要支持结构化输出的AI模型（如 OpenAI GPT-4）
- 复杂嵌套结构需要清晰的查询指令
- AI模型可能不总是严格遵循指定格式
- UI-TARS 模型使用不同的响应格式处理