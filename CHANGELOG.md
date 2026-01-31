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

迁移示例：

```go
// 旧 API
ctrl := maa.NewAdbController(
    adbPath,
    address,
    screencapMethod,
    inputMethod,
    config,
    agentPath,
)

// 新 API
ctrl, err := maa.NewAdbController(
    adbPath,
    address,
    screencapMethod,
    inputMethod,
    config,
    agentPath,
)
if err != nil {
    // 处理错误
}

// 旧 API
ctrl := maa.NewPlayCoverController(address, uuid)

// 新 API
ctrl, err := maa.NewPlayCoverController(address, uuid)
if err != nil {
    // 处理错误
}

// 旧 API
ctrl := maa.NewWin32Controller(
    hWnd,
    screencapMethod,
    mouseMethod,
    keyboardMethod,
)

// 新 API
ctrl, err := maa.NewWin32Controller(
    hWnd,
    screencapMethod,
    mouseMethod,
    keyboardMethod,
)
if err != nil {
    // 处理错误
}

// 旧 API
ctrl := maa.NewGamepadController(hWnd, gamepadType, screencapMethod)

// 新 API
ctrl, err := maa.NewGamepadController(hWnd, gamepadType, screencapMethod)
if err != nil {
    // 处理错误
}

// 旧 API
ctrl := maa.NewCustomController(customCtrl)

// 新 API
ctrl, err := maa.NewCustomController(customCtrl)
if err != nil {
    // 处理错误
}

// 旧 API
ctrl := maa.NewCarouselImageController(path)

// 新 API
ctrl, err := maa.NewCarouselImageController(path)
if err != nil {
    // 处理错误
}

// 旧 API
ctrl := maa.NewBlankController()

// 新 API
ctrl, err := maa.NewBlankController()
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
