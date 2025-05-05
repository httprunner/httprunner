### UI操作标注

针对 UI 操作的位置进行标注，帮助用户直观地了解操作发生的位置。

#### 功能说明

- 点击操作（tap）：使用红色矩形框标注点击位置
- 滑动操作（swipe）：使用红色箭头标注滑动方向，从起始点指向结束点

#### 使用方法

只需在操作函数中添加 `WithMarkOperationEnabled(true)` 选项即可启用操作标注功能：

```go
// 启用操作标注功能
opts := []option.ActionOption{option.WithMarkOperationEnabled(true)}

// 执行点击操作，会自动用红色矩形标注点击位置
err := driver.TapXY(0.5, 0.5, opts...)

// 执行滑动操作，会自动用红色箭头标注滑动方向
err = driver.Swipe(0.2, 0.5, 0.8, 0.5, opts...)

// 可以同时使用其他选项
opts = append(opts, option.WithScreenShotFileName("custom_name"))
err = driver.TapXY(0.3, 0.7, opts...)
```

#### 标注结果

标注后的图片会保存在截图目录中，文件名格式为：`{timestamp}_{tap|swipe}_marked.png`
