## Breaking Change

### AgentClient

- `NewAgentClient` 和 `NewAgentClientTcp` 函数已移除
- 新增 `NewAgentClient(opts ...AgentClientOption) (*AgentClient, error)` 函数
- 新增 `WithIdentifier(identifier string) AgentClientOption` 选项函数
- 新增 `WithTcpPort(port uint16) AgentClientOption` 选项函数

迁移示例：

```go
// 旧 API
client := maa.NewAgentClient("7788")

// 新 API
client, err := maa.NewAgentClient(maa.WithIdentifier("7788"))

// 旧 API
client := maa.NewAgentClientTcp(7788)

// 新 API
client, err := maa.NewAgentClient(maa.WithTcpPort(7788))
```

### Context

**Run 方法返回值变更**

- `RunTask` 现在返回 `(*TaskDetail, error)` 而非 `*TaskDetail`
- `RunRecognition` 现在返回 `(*RecognitionDetail, error)` 而非 `*RecognitionDetail`
- `RunAction` 现在返回 `(*ActionDetail, error)` 而非 `*ActionDetail`
- `RunRecognitionDirect` 现在返回 `(*RecognitionDetail, error)` 而非 `*RecognitionDetail`
- `RunActionDirect` 现在返回 `(*ActionDetail, error)` 而非 `*ActionDetail`

迁移示例：

```go
// 旧 API
detail := ctx.RunTask("MyTask", pipeline)

// 新 API
detail, err := ctx.RunTask("MyTask", pipeline)
if err != nil {
    // 处理错误
}
// 其他 Run 系列方法迁移方式相同
```

**Override 方法返回值变更**

- `OverridePipeline` 现在返回 `error` 而非 `bool`
- `OverrideNext` 现在返回 `error` 而非 `bool`
- `OverrideImage` 现在返回 `error` 而非 `bool`
- `GetNodeJSON` 现在返回 `(string, error)` 而非 `(string, bool)`
- `SetAnchor` 现在返回 `error` 而非 `bool`
- `GetAnchor` 现在返回 `(string, error)` 而非 `(string, bool)`
- `GetHitCount` 现在返回 `(uint64, error)` 而非 `(uint64, bool)`
- `ClearHitCount` 现在返回 `error` 而非 `bool`

迁移示例：

```go
// 旧 API (error 返回类型)
ok := ctx.OverridePipeline(pipeline)

// 新 API
err := ctx.OverridePipeline(pipeline)
if err != nil {
    // 处理错误
}

// 旧 API ((T, error) 返回类型)
data, ok := ctx.GetNodeJSON(name)

// 新 API
data, err := ctx.GetNodeJSON(name)
if err != nil {
    // 处理错误
}
// 其他 Override 系列方法迁移方式相同
```

### TaskJob

- `GetDetail` 现在返回 `(*TaskDetail, error)` 而非 `*TaskDetail`

迁移示例：

```go
// 旧 API
detail := taskJob.Wait().GetDetail()

// 新 API
detail, err := taskJob.Wait().GetDetail()
if err != nil {
    // 处理错误
}
```

### Controller

- `NewAdbController` 现在返回 `(*Controller, error)` 而非 `*Controller`
- `NewPlayCoverController` 现在返回 `(*Controller, error)` 而非 `*Controller`
- `NewWin32Controller` 现在返回 `(*Controller, error)` 而非 `*Controller`
- `NewGamepadController` 现在返回 `(*Controller, error)` 而非 `*Controller`
- `NewCustomController` 现在返回 `(*Controller, error)` 而非 `*Controller`
- `NewCarouselImageController` 现在返回 `(*Controller, error)` 而非 `*Controller`
- `NewBlankController` 现在返回 `(*Controller, error)` 而非 `*Controller`
- `SetScreenshotTargetLongSide`、`SetScreenshotTargetShortSide`、`SetScreenshotUseRawSize` 已移除
- 新增 `SetScreenshot(opts ...ScreenshotOption) error` 与配套选项函数

迁移示例：

```go
// 旧 API
ctrl := maa.NewAdbController(adbPath, address, screencapMethod, inputMethod, config, agentPath)

// 新 API
ctrl, err := maa.NewAdbController(adbPath, address, screencapMethod, inputMethod, config, agentPath)
// 其他 New*Controller 迁移方式相同

// 旧 API
ok := ctrl.SetScreenshotTargetLongSide(1280)

// 新 API
err := ctrl.SetScreenshot(maa.WithScreenshotTargetLongSide(1280))
```

- `GetShellOutput` 现在返回 `(string, error)` 而非 `(string, bool)`
- `CacheImage` 现在返回 `(image.Image, error)` 而非 `image.Image`
- `GetUUID` 现在返回 `(string, error)` 而非 `(string, bool)`
- `GetResolution` 现在返回 `(width, height int32, error)` 而非 `(width, height int32, bool)`

迁移示例：

```go
// 旧 API
output, ok := ctrl.GetShellOutput()

// 新 API
output, err := ctrl.GetShellOutput()
if err != nil {
    // 处理错误
}

// 旧 API
width, height, ok := ctrl.GetResolution()

// 新 API
width, height, err := ctrl.GetResolution()
if err != nil {
    // 处理错误
}
// 其他 Get 系列方法迁移方式相同
```

### Custom Action and Recognition

- `CustomAction` 类型别名已移除，直接使用 `CustomActionRunner`
- `CustomRecognition` 类型别名已移除，直接使用 `CustomRecognitionRunner`

迁移示例：

```go
// 旧 API
var _ maa.CustomAction = &MyAction{}
var _ maa.CustomRecognition = &MyRecognition{}

// 新 API
var _ maa.CustomActionRunner = &MyAction{}
var _ maa.CustomRecognitionRunner = &MyRecognition{}
```

### Tasker

- `GetLatestNode` 现在返回 `(*NodeDetail, error)` 而非 `*NodeDetail`

迁移示例：

```go
// 旧 API
detail := tasker.GetLatestNode("MyTaskName")

// 新 API
detail, err := tasker.GetLatestNode("MyTaskName")
if err != nil {
    // 处理错误
}
```

### Resource

- `NewResource` 现在返回 `(*Resource, error)` 而非 `*Resource`

迁移示例：

```go
// 旧 API
res := maa.NewResource()

// 新 API
res, err := maa.NewResource()
if err != nil {
    // 处理错误
}
```

**设置方法返回值变更**

- `UseCPU` 现在返回 `error` 而非 `bool`
- `UseDirectml` 现在返回 `error` 而非 `bool`
- `UseCoreml` 现在返回 `error` 而非 `bool`
- `UseAutoExecutionProvider` 现在返回 `error` 而非 `bool`

迁移示例：

```go
// 旧 API
ok := res.UseCPU()

// 新 API
err := res.UseCPU()
if err != nil {
    // 处理错误
}
// 其他设置方法迁移方式相同
```

**自定义识别和操作方法返回值变更**

- `RegisterCustomRecognition` 现在返回 `error` 而非 `bool`
- `UnregisterCustomRecognition` 现在返回 `error` 而非 `bool`
- `ClearCustomRecognition` 现在返回 `error` 而非 `bool`
- `RegisterCustomAction` 现在返回 `error` 而非 `bool`
- `UnregisterCustomAction` 现在返回 `error` 而非 `bool`
- `ClearCustomAction` 现在返回 `error` 而非 `bool`

迁移示例：

```go
// 旧 API
ok := res.RegisterCustomRecognition("MyRec", &MyRecognition{})

// 新 API
err := res.RegisterCustomRecognition("MyRec", &MyRecognition{})
if err != nil {
    // 处理错误
}
// 其他自定义方法迁移方式相同
```

**覆盖和查询方法返回值变更**

- `OverridePipeline` 现在返回 `error` 而非 `bool`
- `OverrideNext` 现在返回 `error` 而非 `bool`
- `OverrideImage` 现在返回 `error` 而非 `bool`

**方法重命名**

- `OverriderImage` 已重命名为 `OverrideImage`（修正拼写错误）

迁移示例：

```go
// 旧 API
err := res.OverriderImage("name", img)

// 新 API
err := res.OverrideImage("name", img)
```
- `GetNodeJSON` 现在返回 `(string, error)` 而非 `(string, bool)`
- `GetHash` 现在返回 `(string, error)` 而非 `(string, bool)`
- `GetNodeList` 现在返回 `([]string, error)` 而非 `([]string, bool)`
- `GetCustomRecognitionList` 现在返回 `([]string, error)` 而非 `([]string, bool)`
- `GetCustomActionList` 现在返回 `([]string, error)` 而非 `([]string, bool)`
- `GetDefaultRecognitionParam` 现在返回 `(NodeRecognitionParam, error)` 而非 `(aNodeRecognitionParam, bool)`
- `GetDefaultActionParam` 现在返回 `(NodeActionParam, error)` 而非 `(NodeActionParam, bool)`
- `Clear` 现在返回 `error` 而非 `bool`

迁移示例：

```go
// 旧 API (error 返回类型)
ok := res.OverridePipeline(pipeline)

// 新 API
err := res.OverridePipeline(pipeline)
if err != nil {
    // 处理错误
}

// 旧 API ((T, error) 返回类型)
hash, ok := res.GetHash()

// 新 API
hash, err := res.GetHash()
if err != nil {
    // 处理错误
}
// 其他查询方法迁移方式相同
```