## Breaking Change

### AgentClient

- `NewAgentClient` 和 `NewAgentClientTcp` 函数已移除
- 新增 `NewAgentClient(opts ...AgentClientOption) (*AgentClient, error)` 函数
- 新增 `WithIdentifier(identifier string) AgentClientOption` 选项函数
- 新增 `WithTcpPort(port uint16) AgentClientOption` 选项函数
- `NewAgentClient` 现在返回 `(*AgentClient, error)` 而非 `*AgentClient`

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


