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

// 旧 API
detail := ctx.RunRecognition("MyRec", img, pipeline)

// 新 API
detail, err := ctx.RunRecognition("MyRec", img, pipeline)
if err != nil {
    // 处理错误
}

// 旧 API
detail := ctx.RunAction("MyAct", box, recDetail, pipeline)

// 新 API
detail, err := ctx.RunAction("MyAct", box, recDetail, pipeline)
if err != nil {
    // 处理错误
}

// 旧 API
detail := ctx.RunRecognitionDirect(NodeRecognitionTypeDirectHit, param, img)

// 新 API
detail, err := ctx.RunRecognitionDirect(NodeRecognitionTypeDirectHit, param, img)
if err != nil {
    // 处理错误
}

// 旧 API
detail := ctx.RunActionDirect(NodeActionTypeClick, param, box, recDetail)

// 新 API
detail, err := ctx.RunActionDirect(NodeActionTypeClick, param, box, recDetail)
if err != nil {
    // 处理错误
}
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
// 旧 API
ok := ctx.OverridePipeline(pipeline)

// 新 API
err := ctx.OverridePipeline(pipeline)
if err != nil {
    // 处理错误
}

// 旧 API
ok := ctx.OverrideNext(name, nextList)

// 新 API
err := ctx.OverrideNext(name, nextList)
if err != nil {
    // 处理错误
}

// 旧 API
ok := ctx.OverrideImage(imageName, image)

// 新 API
err := ctx.OverrideImage(imageName, image)
if err != nil {
    // 处理错误
}

// 旧 API
data, ok := ctx.GetNodeJSON(name)

// 新 API
data, err := ctx.GetNodeJSON(name)
if err != nil {
    // 处理错误
}

// 旧 API
ok := ctx.SetAnchor(anchorName, nodeName)

// 新 API
err := ctx.SetAnchor(anchorName, nodeName)
if err != nil {
    // 处理错误
}

// 旧 API
anchor, ok := ctx.GetAnchor(anchorName)

// 新 API
anchor, err := ctx.GetAnchor(anchorName)
if err != nil {
    // 处理错误
}

// 旧 API
count, ok := ctx.GetHitCount(nodeName)

// 新 API
count, err := ctx.GetHitCount(nodeName)
if err != nil {
    // 处理错误
}

// 旧 API
ok := ctx.ClearHitCount(nodeName)

// 新 API
err := ctx.ClearHitCount(nodeName)
if err != nil {
    // 处理错误
}
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
img := ctrl.CacheImage()

// 新 API
img, err := ctrl.CacheImage()
if err != nil {
    // 处理错误
}

// 旧 API
uuid, ok := ctrl.GetUUID()

// 新 API
uuid, err := ctrl.GetUUID()
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
