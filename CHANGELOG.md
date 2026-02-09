## Breaking Change

### API 变更概览

本次重大变更将所有方法的返回类型从 `bool` 或 `(T, bool)` 改为标准的 Go 错误处理模式：

- **构造函数**：`*T` → `(*T, error)`
- **设置方法**：`bool` → `error`
- **查询方法**：`(T, bool)` → `(T, error)`
- **运行方法**：`T` → `(T, error)`

### 受影响的组件

#### AgentClient

| 变更类型 | 旧 API | 新 API |
|---------|--------|--------|
| 构造函数 | `NewAgentClient(string)` <br> `NewAgentClientTcp(uint16)` | `NewAgentClient(opts ...AgentClientOption)` |
| 设置方法 | `bool` 返回型 | `error` 返回型 |
| 查询方法 | `(T, bool)` 返回型 | `(T, error)` 返回型 |

**受影响的方法**：
- 设置方法：`BindResource`, `Connect`, `Disconnect`, `SetTimeout`, `RegisterResourceSink`, `RegisterControllerSink`, `RegisterTaskerSink`
- 查询方法：`Identifier`, `GetCustomRecognitionList`, `GetCustomActionList`

**新增选项函数**：
- `WithIdentifier(identifier string) AgentClientOption`
- `WithTcpPort(port uint16) AgentClientOption`

#### Context

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 运行方法 | `RunTask`, `RunRecognition`, `RunAction`, `RunRecognitionDirect`, `RunActionDirect` |
| 设置方法 | `OverridePipeline`, `OverrideNext`, `OverrideImage`, `SetAnchor`, `ClearHitCount` |
| 查询方法 | `GetNodeJSON`, `GetAnchor`, `GetHitCount` |

**补充说明**：`OverrideNext` 现改为接收 `[]NodeNextItem`。

#### TaskJob

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 查询方法 | `GetDetail` |
| 设置方法 | `OverridePipeline` |
| 新增方法 | `Error() error` |

**错误处理增强**：当任务创建过程中发生错误（如 JSON 序列化失败）时，`TaskJob` 会保存该错误而非静默忽略。此时：
- `Status()` 返回 `StatusFailure`
- `Error()` 返回具体的错误信息
- `Wait()` 会跳过等待直接返回
- `GetDetail()` 和 `OverridePipeline()` 会返回保存的错误

#### Controller

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 构造函数 | `NewAdbController`, `NewPlayCoverController`, `NewWin32Controller`, `NewGamepadController`, `NewCustomController`, `NewCarouselImageController`, `NewBlankController` |
| 设置方法 | `SetScreenshot`（改用 Option 模式） |
| 查询方法 | `GetShellOutput`, `CacheImage`, `GetUUID`, `GetResolution` |

**移除方法**：`SetScreenshotTargetLongSide`, `SetScreenshotTargetShortSide`, `SetScreenshotUseRawSize`
**新增**：`SetScreenshot(opts ...ScreenshotOption) error` 与配套选项函数

#### Tasker

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 构造函数 | `NewTasker` |
| 查询方法 | `GetLatestNode` |
| 设置方法 | `BindResource`, `BindController`, `ClearCache` |

#### Resource

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 构造函数 | `NewResource` |
| 设置方法 | `UseCPU`, `UseDirectml`, `UseCoreml`, `UseAutoExecutionProvider`, `RegisterCustomRecognition`, `UnregisterCustomRecognition`, `ClearCustomRecognition`, `RegisterCustomAction`, `UnregisterCustomAction`, `ClearCustomAction`, `OverridePipeline`, `OverrideNext`, `OverrideImage`, `Clear` |
| 查询方法 | `GetNodeJSON`, `GetHash`, `GetNodeList`, `GetCustomRecognitionList`, `GetCustomActionList`, `GetDefaultRecognitionParam`, `GetDefaultActionParam` |

**补充说明**：`OverrideNext` 现改为接收 `[]NodeNextItem`。

#### Custom Action and Recognition

| 变更类型 | 旧 API | 新 API |
|---------|--------|--------|
| 类型别名 | `CustomAction` | `CustomActionRunner` |
| 类型别名 | `CustomRecognition` | `CustomRecognitionRunner` |

#### Global Configuration

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 设置方法 | `SetLogDir`, `SetSaveDraw`, `SetStdoutLevel`, `SetDebugMode`, `SetSaveOnError`, `SetDrawQuality`, `SetRecoImageCacheLimit`, `LoadPlugin` |

#### Toolkit

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 设置方法 | `ConfigInitOption` |
| 查询方法 | `FindAdbDevices`, `FindDesktopWindows` |

### 类型名称修正

| 旧 API | 新 API |
|--------|--------|
| `InterenceDevice` | `InferenceDevice` |
| `InterenceDeviceAuto` | `InferenceDeviceAuto` |
| `OverriderImage` | `OverrideImage` |

### 方法重命名

- `Context.GetNodeData` → `Context.GetNode`

### 迁移示例

#### 构造函数迁移

```go
// 旧 API
client := maa.NewAgentClient("7788")

// 新 API
client, err := maa.NewAgentClient(maa.WithIdentifier("7788"))
if err != nil {
    // 处理错误
}
```

#### 设置方法迁移（bool → error）

```go
// 旧 API
ok := maa.SetLogDir("./logs")

// 新 API
err := maa.SetLogDir("./logs")
if err != nil {
    // 处理错误
}
```

#### 查询方法迁移（(T, bool) → (T, error)）

```go
// 旧 API
id, ok := client.Identifier()

// 新 API
id, err := client.Identifier()
if err != nil {
    // 处理错误
}
```

#### 运行方法迁移（T → (T, error)）

```go
// 旧 API
detail := ctx.RunTask("MyTask", pipeline)

// 新 API
detail, err := ctx.RunTask("MyTask", pipeline)
if err != nil {
    // 处理错误
}
```

#### OverrideNext 迁移（[]string → []NodeNextItem）

```go
// 旧 API
err := ctx.OverrideNext("Entry", []string{"TaskA", "[JumpBack]TaskB"})

// 新 API
err := ctx.OverrideNext("Entry", []maa.NodeNextItem{
    {Name: "TaskA"},
    {Name: "TaskB", JumpBack: true},
})
```

#### 任务创建错误处理

```go
// 新增：检查任务创建阶段的错误
job := tasker.PostTask("entry", invalidOverride)
if err := job.Error(); err != nil {
    // 处理任务创建错误（如 JSON 序列化失败）
}
```

## Added

- `Resource.GetNode`
- `Pipeline.GetNode`
- `Pipeline.HasNode`
- `Pipeline.RemoveNode`
- `Pipeline.Len`
