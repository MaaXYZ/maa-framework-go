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

**补充说明**：`OverrideNext` 现改为接收 `[]NextItem`。

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
| 查询方法 | `GetLatestNode`, `GetTaskDetail` |
| 设置方法 | `BindResource`, `BindController`, `ClearCache` |

#### Resource

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 构造函数 | `NewResource` |
| 设置方法 | `UseCPU`, `UseDirectml`, `UseCoreml`, `UseAutoExecutionProvider`, `RegisterCustomRecognition`, `UnregisterCustomRecognition`, `ClearCustomRecognition`, `RegisterCustomAction`, `UnregisterCustomAction`, `ClearCustomAction`, `OverridePipeline`, `OverrideNext`, `OverrideImage`, `Clear` |
| 查询方法 | `GetNodeJSON`, `GetHash`, `GetNodeList`, `GetCustomRecognitionList`, `GetCustomActionList`, `GetDefaultRecognitionParam`, `GetDefaultActionParam` |

**补充说明**：`OverrideNext` 现改为接收 `[]NextItem`。

#### Custom Action and Recognition

| 变更类型 | 旧 API | 新 API |
|---------|--------|--------|
| 类型别名 | `CustomAction` | `CustomActionRunner` |
| 类型别名 | `CustomRecognition` | `CustomRecognitionRunner` |
| 回调参数 | `CustomActionArg.TaskDetail *TaskDetail` | `CustomActionArg.TaskID int64` |
| 回调参数 | `CustomRecognitionArg.TaskDetail *TaskDetail` | `CustomRecognitionArg.TaskID int64` |

**补充说明**：自定义识别与动作回调默认不再预取任务详情。若确有需要，请通过 `Tasker.GetTaskDetail(taskId int64)` 按需查询。

#### Global Configuration

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 设置方法 | `SetLogDir`, `SetSaveDraw`, `SetStdoutLevel`, `SetDebugMode`, `SetSaveOnError`, `SetDrawQuality`, `SetRecoImageCacheLimit`, `LoadPlugin` |

**补充说明**：
- `InitConfig` 已改为私有类型 `initConfig`，不再对外暴露。
- `InitOption` 签名改为 `type InitOption func(*initConfig)`，由于参数类型私有，包外不再支持自定义 `InitOption`，请使用 `WithXxx` 函数。
- `Init()` 不再隐式应用默认全局配置，仅在显式传入对应 `WithXxx` 时才会调用设置。
- `defaultInitConfig()` 已移除，`Init()` 现在直接使用 `initConfig{}` 初始化。
- `WithPluginPaths` 会对输入切片进行拷贝，避免外部后续修改影响已构建的选项。

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

### 类型与 API 重命名（refactor/node）

| 旧名称 | 新名称 |
|--------|--------|
| `NodeNextItem` | `NextItem` |
| `NodeMultiSwipeItem` | `MultiSwipeItem` |
| `NodeAction` | `Action` |
| `NodeRecognition` | `Recognition` |
| 各 `Node*Param`（如 `NodeCustomActionParam`） | 去掉 `Node` 前缀（如 `CustomActionParam`、`ClickParam`、`OCRParam` 等） |

**Action 相关**：动作定义由 `node_action.go` 迁移至 `action.go`；构造函数统一为单参数（如 `ActClick(p ClickParam)`），不再使用 variadic。`ActMultiSwipe` 使用 `MultiSwipeItem`（原 `NodeMultiSwipeItem`）。

**Recognition 相关**：识别定义由 `node_recognition.go` 迁移至 `recognition.go`。`WithBoxIndex` 重命名为链式方法 `SetBoxIndex`（如 `RecAnd(...).SetBoxIndex(2)`）。`RecOCR` 由 variadic 改为单参：`RecOCR(p OCRParam)`。各算法的 `OrderBy` 枚举按算法拆分为独立类型（与 C++ 对齐）。

**Context**：`Context.WaitFreezes` 参数收窄为 `*WaitFreezesParam`。

**行为说明**：动作/识别构造函数会对 slice 等参数做 clone，避免与调用方共享底层数组。

### Node Anchor API 变更

- `Node.Anchor`：`[]string` → `map[string]string`（与 C++ `GetNodeData` 输出一致，`anchor` 为对象）
- `Node.SetAnchor`：`SetAnchor([]string)` → `SetAnchor(map[string]string)`
- 不再兼容旧的 `anchor` 字符串/字符串数组语义，统一为对象语义：
  - `{"A":"CurrentNode"}` 表示锚点指向目标节点
  - `{"A":""}` 表示显式清除锚点
- `Node.AddAnchor(anchor)` 语义明确为快捷写法：设置 `anchor -> 当前节点名`
- `Node.RemoveAnchor(anchor)` 保持为删除该配置项（移除 key）

### NodeRecognition API 变更

#### And/Or 识别：SubRecognitionItem 与 C++ GetNodeData 对齐

与 C++ 端 `GetNodeData` 输出一致：`all_of` / `any_of` 数组元素为 **节点名字符串** 或 **内联识别对象**。Go 侧引入统一类型并调整 And/Or 构造方式。

| 变更类型 | 旧 API | 新 API |
|---------|--------|--------|
| 子项类型（And） | `AllOf []*NodeAndRecognitionItem` | `AllOf []SubRecognitionItem` |
| 子项类型（Or） | `AnyOf []*NodeRecognition` | `AnyOf []SubRecognitionItem` |
| 内联项类型名 | `NodeAndRecognitionItem` | `InlineSubRecognition`（And/Or 通用） |
| RecAnd 签名 | `RecAnd([]*NodeAndRecognitionItem, opts ...)` | `RecAnd(items ...SubRecognitionItem)`，BoxIndex 用链式 `.SetBoxIndex(n)` |
| RecOr 签名 | `RecOr(anyOf []SubRecognitionItem)` | `RecOr(anyOf ...SubRecognitionItem)` |

**新增类型与函数**：
- `SubRecognitionItem`：表示一项子识别，可为节点名引用（`NodeName`）或内联识别（`Inline *InlineSubRecognition`），JSON 为 string 或 object。
- `InlineSubRecognition`：内联子识别（含 `sub_name`、`type`、`param`），与 C++ `InlineSubRecognition` 对应。
- `Ref(nodeName string) SubRecognitionItem`：按节点名引用。
- `Inline(rec *Recognition, name ...string) SubRecognitionItem`：内联识别，`name` 可选（Or 常省略）。

**受影响的方法与字段**：
- `RecAnd(items ...SubRecognitionItem)`、`RecOr(anyOf ...SubRecognitionItem)`
- `AndRecognitionParam.AllOf`、`OrRecognitionParam.AnyOf`
- `Ref` / `Inline` 为子识别项构造的推荐写法（原 `AndItem`、`SubRecognitionRef`/`SubRecognitionInline` 已移除）。

### 迁移示例

#### Init 选项迁移（隐式默认 -> 显式传参）

```go
// 旧行为：Init() 会隐式应用部分默认全局配置
_ = maa.Init()

// 新行为：如需保持旧默认配置，请显式传入 WithXxx
err := maa.Init(
    maa.WithLogDir("./debug"),
    maa.WithStdoutLevel(maa.LoggingLevelInfo),
    maa.WithSaveDraw(false),
    maa.WithDebugMode(false),
)
if err != nil {
    // 处理错误
}
```

#### InitOption 迁移（包外自定义 -> 内置 WithXxx）

原来在包外自定义 `InitOption` 的代码需迁移为内置 `WithXxx` 函数。

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

#### OverrideNext 迁移（[]string → []NextItem）

```go
// 旧 API
err := ctx.OverrideNext("Entry", []string{"TaskA", "[JumpBack]TaskB"})

// 新 API
err := ctx.OverrideNext("Entry", []maa.NextItem{
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

#### And/Or Recognition 迁移（SubRecognitionItem + Ref/Inline）

```go
// 旧 API（指针数组 + AndItem）
rec := maa.RecAnd([]*maa.NodeAndRecognitionItem{
    maa.AndItem("template", maa.RecTemplateMatch(...)),
    maa.AndItem("color", maa.RecColorMatch(...)),
}, maa.WithAndRecognitionBoxIndex(0))

orRec := maa.RecOr([]maa.SubRecognitionItem{
    maa.SubRecognitionInline(maa.AndItem("", maa.RecTemplateMatch(...))),
})

// 新 API（variadic + Ref/Inline）
rec := maa.RecAnd(
    maa.Ref("OtherNode"),                           // 节点名引用
    maa.Inline(maa.RecTemplateMatch(...), "template"),
    maa.Inline(maa.RecColorMatch(...), "color"),
).SetBoxIndex(0)

orRec := maa.RecOr(
    maa.Inline(maa.RecTemplateMatch(...)),   // 无 sub_name 时省略第二参数
    maa.Inline(maa.RecColorMatch(...)),
)
```

#### Node Anchor 迁移（[]string → map[string]string）

```go
// 旧 API
node.SetAnchor([]string{"X", "Y"})

// 新 API（指向当前节点）
node.SetAnchor(map[string]string{
    "X": node.Name,
    "Y": node.Name,
})

// 新 API（指向指定节点）
node.SetAnchorTarget("X", "TargetNode")

// 新 API（显式清除锚点）
node.ClearAnchor("X") // 等价于 node.SetAnchorTarget("X", "")
```

### RecognitionResults.Best 类型修正

`RecognitionResults.Best` 字段从 `[]*RecognitionResult` 修正为 `*RecognitionResult`，与 C++ 端 `best_result_`（`std::optional<Result>`）对齐。JSON 中 `best` 为单个对象或 `null`，而非数组。

```go
// 旧 API
best := results.Best[0] // 按数组索引访问

// 新 API
best := results.Best // 直接使用，可能为 nil
if best != nil {
    // 使用 best
}
```

### 字段名与 JSON Tag 对齐 C++

以下字段名和 JSON tag 修正为与 C++ 序列化输出一致：

| 结构体 | 旧字段 / JSON tag | 新字段 / JSON tag | C++ 对照 |
|--------|-------------------|-------------------|----------|
| `ShellActionResult` | `Timeout` / `"timeout"` | `ShellTimeout` / `"shell_timeout"` | `Actuator.cpp` |
| `NodeNextListDetail` | `NextList` / `"next_list"` | `List` / `"list"` | `PipelineTask.cpp` |

### NeuralNetworkClassifyResult 移除多余字段

移除 `Raw []float64` 和 `Probs []float64` 字段。C++ 端 `NeuralNetworkClassifierResult` 的 `MEO_JSONIZATION` 仅导出 `cls_index, label, box, score`，`raw` 和 `probs` 不参与 JSON 序列化，Go 侧保留会导致永远为零值。

## Added

- `Resource.GetNode`
- `Pipeline.GetNode`
- `Pipeline.HasNode`
- `Pipeline.RemoveNode`
- `Pipeline.Len`
- `Node.SetAnchorTarget`
- `Node.ClearAnchor`
- And/Or 识别：`SubRecognitionItem`、`InlineSubRecognition`、`Ref`、`Inline`（与 C++ GetNodeData 的 all_of/any_of 对齐；`RecAnd`/`RecOr` 均为 variadic）
- `Recognition.SetBoxIndex`：链式方法，替代原 `WithBoxIndex`，指定 And 识别使用哪个子结果的 box
- `WaitFreezesParam` 与 `Context.WaitFreezes(duration, box, *WaitFreezesParam)`：等待画面稳定
- OCR 颜色过滤：`OCRParam.ColorFilter` 字段 & `WithOCRColorFilter` 选项函数，指定 ColorMatch 节点名对图像进行颜色二值化后再送入 OCR 识别（适配 [MaaFramework#1145](https://github.com/MaaXYZ/MaaFramework/pull/1145)）
