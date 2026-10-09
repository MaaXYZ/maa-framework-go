## Breaking Change

### API 变更概览

本次重大变更将所有方法的返回类型从 `bool` 或 `(T, bool)` 改为标准的 Go 错误处理模式：

- **构造函数**：`*T` → `(*T, error)`
- **设置方法**：`bool` → `error`
- **查询方法**：`(T, bool)` → `(T, error)`
- **运行方法**：`T` → `(T, error)`
- **提交方法**：`*Job` / `*TaskJob` → `(*Job, error)` / `(*TaskJob, error)`

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

#### Tasker / Controller / Resource

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 提交方法 | Tasker：`PostTask`, `PostRecognition`, `PostAction`, `PostStop`；Controller：`PostConnect`, `PostClick`, `PostSwipe`, `PostScreencap` 等全部 `Post*`；Resource：`PostBundle`, `PostOcrModel`, `PostPipeline`, `PostImage` |

**提交错误语义**：所有 `Post*` 方法统一返回 `(Job, error)`。提交失败分为两类：
- wrapper 预检失败（如 JSON 序列化失败、对象已关闭）：不会调用原生提交接口，直接返回一个终态失败的 Job 和非 nil 的 error
- 原生提交接口返回 invalid ID：原生接口已被调用但拒绝了本次提交，Go 侧将其转换为同样的终态失败 Job 和 error

两种失败的共同行为：
- 返回的 `error` 非 nil 当且仅当提交失败
- 忽略 error 的调用方在 `Status()` / `Wait()` 上得到失败终态，而不是一个永远 pending 的 Job
- `Error()` 保留为镜像访问器：对提交失败的 Job 读取同一提交错误；成功提交的 Job 在拥有者关闭后也会返回 `ErrClosed`（Job 不再可用）

Context 的运行方法（`RunTask` / `RunRecognition` / `RunAction`）与 `WaitFreezes` 的参数序列化失败同样返回错误，此时不会提交到原生层。

**图像参数校验**：`Tasker.PostRecognition`、`Context.RunRecognition`、`Context.RunRecognitionDirect`、`Context.OverrideImage`、`Resource.OverrideImage` 现在会校验图像参数，图像为 nil 或宽高为 0 时返回错误，不会调用原生接口（旧版对空图静默清空 buffer，对 nil 图直接 panic）。

**并发与回调约定**：

- `Job` / `TaskJob` 的等待与状态查询可以并发执行，多个等待者共享完成结果；对象不可复制。Tasker 绑定 getter 可与绑定变更并发调用。
- sink 与自定义识别、动作的注册变更必须在实例及关联 tasker 静止时执行，不得在回调中变更。配置事务会串行化，原生注册失败会回滚 Go 回调；`Add*Sink` 失败仍返回 0。
- 自定义 Controller 的回调保留至原生析构完成，析构期间的 `KeyUp` / `TouchUp` 可正常执行。`Destroy` 成功返回后不再调用用户回调；回调内销毁返回 `ErrInCallback`。
- stop 使旧 Job ID 失效时，`Wait` 返回不代表原生工作已经结束。Controller 销毁可能提交 inactive 动作并暂时返回 `ErrInUse`，需等待后重试。
- AgentServer 只允许在启动前配置；活动阶段的自定义注册和再次启动返回 `ErrInUse`，添加 sink 返回 0。未 detach 的服务关闭后不支持重启：启动和自定义注册返回 `ErrClosed`，添加 sink 返回 0，`Release` 后再次 `Init` 也不会恢复服务，但仍可 `Release`，重复关闭不再调用原生接口。自定义识别与动作共用名称，重名注册返回错误并保留已有注册。生命周期操作需由调用方串行协调。

完整使用边界见 [并发与回调](README_zh.md#并发与回调)。

#### TaskJob

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 查询方法 | `GetDetail` |
| 设置方法 | `OverridePipeline` |
| 新增方法 | `Error() error` |

**错误处理增强**：当任务提交过程中发生错误（如 JSON 序列化失败）时，`PostTask` 会返回终态失败的 `TaskJob` 和对应的 error。此时：
- `Status()` 返回 `StatusFailure`
- `Error()` 返回具体的错误信息
- `Wait()` 会跳过等待直接返回
- `GetDetail()` 和 `OverridePipeline()` 会返回保存的错误

#### Controller

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 构造函数 | `NewAdbController`, `NewPlayCoverController`, `NewWin32Controller`, `NewLinuxController`, `NewMacOSController`, `NewAndroidNativeController`, `NewReplayController`, `NewRecordController`, `NewGamepadController`, `NewCustomController`, `NewBlankController`, `NewCarouselImageController` |
| 设置方法 | `SetScreenshot`（改用 Option 模式）, `SetMouseLockFollow` |
| 查询方法 | `GetShellOutput`, `CacheImage`, `CacheImageInto`, `GetUUID`, `GetResolution`, `GetInfo` |

**移除方法**：`SetScreenshotTargetLongSide`, `SetScreenshotTargetShortSide`, `SetScreenshotUseRawSize`
**移除构造函数**：`NewCarouselImageController` 已移除。若仅需空操作控制器，请使用 `NewBlankController()`；若需基于录制数据回放，请使用 `NewReplayController(recordingPath)`，录制入口为 `NewRecordController(inner, recordingPath)`。
**新增**：`SetScreenshot(opts ...ScreenshotOption) error` 与配套选项函数；新增 `WithScreenshotResizeMethod(...)` / `ScreenshotResizeMethod*` 常量，以及 `SetMouseLockFollow(enabled bool) error`
**截图选项组合与校验**：`SetScreenshot` 现在应用全部可组合的选项，不再只应用最后一个。长边、短边、Expand 目标在同次调用中互斥，原始尺寸 `true` 与尺寸目标互斥，同一设置重复指定也返回错误；尺寸须为正数，插值只允许 0 到 4。参数错误在任何原生修改之前返回；原生 setter 失败时保留此前成功的修改并跳过后续设置。分次调用仍保留原始尺寸模式下的目标与插值，关闭该模式后恢复缩放。示例及完整规则见 [Controller.SetScreenshot](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Controller.SetScreenshot)。
**新增控制器构造函数**：`NewLinuxController(configJson string)`、`NewMacOSController(...)`、`NewAndroidNativeController(...)`、`NewReplayController(...)`、`NewRecordController(...)`
**接口变更**：
- `CustomController` 接口新增 `RelativeMove(dx, dy int32) bool`、`Shell(cmd string, timeout int64) (string, bool)`、`GetInfo() (string, bool)` 必须实现方法。已有实现若无需支持，可返回 no-op 成功值
- `NewWlRootsController` 已移除，改用 `NewLinuxController(configJson string)`；通过 JSON 配置选择截图与输入方式，Wlr 输入可用 `use_win32_vk_code` 将按键视为 Win32 VK 键码。
**Win32 InputMethod 命名对齐**：
- `InputSendMessageWithCursorPosAndBlockInput` → `InputSendMessageWithWindowPos`
- `InputPostMessageWithCursorPosAndBlockInput` → `InputPostMessageWithWindowPos`
**截图缓存类型变更**：`Controller.CacheImageInto` 入参与返回值由 `*image.NRGBA` 调整为 `*image.RGBA`

#### Tasker

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 构造函数 | `NewTasker` |
| 查询方法 | `GetLatestNode`, `GetNodeDetail`, `GetTaskDetail`, `GetWaitFreezesDetail` |
| 设置方法 | `BindResource`, `BindController`, `ClearCache` |

**补充说明**：`TaskDetail` 不再预取完整 `NodeDetail` 列表，现改为返回懒加载的 `Nodes []NodeRef`；可通过 `NodeRef.GetDetail()` 或 `Tasker.GetNodeDetail(nodeId)` 按需获取节点详情。
**新增 WaitFreezes 查询**：`Tasker.GetWaitFreezesDetail(wfId int64) (*WaitFreezesDetail, error)` 可根据回调中的 `wf_id` 查询阶段、耗时、识别 ID 列表和 ROI。
**详情查询错误语义**：`GetRecognitionDetail`、`GetActionDetail`、`GetWaitFreezesDetail` 在无对应详情时返回非 nil 的 error，不再返回 `(nil, nil)`。
**任务详情修复**：`GetTaskDetail` 在没有记录节点时仍返回原生任务的 `Entry` 和 `Status`，不再返回空入口和 `StatusInvalid`。

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
`CustomActionFunc` 与 `CustomRecognitionFunc` 可将普通函数直接适配为对应 Runner，并传给 `Resource.RegisterCustomAction` / `Resource.RegisterCustomRecognition`。

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
- `Init()` 与 `Release()` 现为幂等操作：重复初始化或在未初始化状态下释放都会直接返回 `nil`；原导出的 `ErrAlreadyInitialized`、`ErrNotInitialized` 已移除。
- `Init()` 过程中若某个原生库加载失败，会自动释放此前已成功加载的库，避免残留半初始化状态。
- `LibraryLoadError` 现在会稳定包含库名与尝试加载的完整路径，便于排查动态库装载问题。

#### Toolkit

| 变更类型 | 受影响的方法 |
|---------|-------------|
| 设置方法 | `ConfigInitOption` |
| 查询方法 | `FindAdbDevices`, `FindDesktopWindows` |

**新增 macOS 权限接口**：`MacOSCheckPermission`、`MacOSRequestPermission`、`MacOSRevealPermissionSettings`，用于查询/申请权限并跳转系统设置。

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
- `Pipeline.UnmarshalJSON` 接受 `anchor` 字符串或字符串数组，并将每个锚点解析为包含它的节点名（pipeline 的 map key）；单独解码 `Node` 只接受对象形式，即使其 `Name` 已设置。对象形式保留目标节点名，不检查目标是否存在：
  - `{"A":"CurrentNode"}` 表示锚点指向目标节点
  - `{"A":""}` 表示显式清除锚点
- `Node.Anchor` 为 nil 时编码省略 `anchor`，覆盖已有原生节点时继承其配置；非 nil 空 map 编码为 `"anchor": {}`，清空该节点的 anchor 配置。`Pipeline` 解码 `"anchor": []` / `"anchor": {}` 时保留非 nil 空 map。清空配置不会移除已登记的运行时锚点，运行时清除仍使用对象中的空目标字符串。
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
- `InlineSubRecognition`：v2 内联子识别（含 `sub_name` 与嵌套的 `recognition: {type, param}`），与 C++ `InlineSubRecognition` 对应。
- `Ref(nodeName string) SubRecognitionItem`：按节点名引用。
- `Inline(rec *Recognition, name ...string) SubRecognitionItem`：内联识别，`name` 可选（Or 常省略）。

**受影响的方法与字段**：
- `RecAnd(items ...SubRecognitionItem)`、`RecOr(anyOf ...SubRecognitionItem)`
- `AndRecognitionParam.AllOf`、`OrRecognitionParam.AnyOf`
- `Ref` / `Inline` 为子识别项构造的推荐写法（原 `AndItem`、`SubRecognitionRef`/`SubRecognitionInline` 已移除）。

### Pipeline v2 协议对齐

类型化的 pipeline 模型与构造器统一编码为 v2 的嵌套 `type` / `param` 对象格式。解码同时接受 v1 扁平识别/动作字段，以及省略 `param` 的识别/动作对象，并规范化为 v2 模型。此次同步 MaaFramework v5.14.2 的协议行为：

- `next` / `on_error` 可解码单个节点值、节点名字符串，以及混合字符串和对象的数组；重编码为节点对象数组。成功解码会替换已有 `Node` / `Pipeline` 状态，失败则保留原状态；`Pipeline` 根据 map key 设置每个节点的 `Name`。
- `Target` 的二维坐标 `[x, y]` 规范化为 `[x, y, 1, 1]`；`false`、坐标数量或类型不合法的 target 会被拒绝。
- `Click`、`LongPress`、`Swipe` 和 `MultiSwipe` 补齐可区分继承与显式零值的 `pressure`；`ShellParam.ShellTimeout` 使用 `*time.Duration` 配置，JSON 编码为毫秒，支持显式 `0` 及 `-time.Millisecond` 无限等待。
- `TouchMoveParam` 与 `KeyUpParam` 对齐共享的 `auto_up` JSON 字段，但该字段仅在 `TouchDown` / `KeyDown` 执行时生效。
- `DirectHitParam` 支持 `ROI` 与 `ROIOffset`；`NeuralNetworkDetectParam` 支持 `Threshold`；新增 `TemplateMatchMethodSQDIFF_NORMED`（值 `1`）。
- And/Or 的内联识别使用嵌套的 `recognition` 字段：`{ "sub_name": "...", "recognition": { "type": "...", "param": { ... } } }`。
- 未知 Action/Recognition 类型的参数可通过 [`RawActionParam`](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#RawActionParam) / [`RawRecognitionParam`](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#RawRecognitionParam) 保留原始 JSON，用于读取和回写；这只是 Go 侧的透传能力，不代表 MaaFramework 原生库支持该未知类型。
- `tools/api-check` 可根据同一发行版的 pipeline schema 检查 v2 类型、解码分支和字段覆盖，CI 可据此及时发现协议漂移。

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

#### 调试控制器迁移（CarouselImageController → Blank / Replay）

```go
// 旧 API
ctrl, err := maa.NewCarouselImageController("./images")

// 新 API：仅需空操作 / 生命周期测试
ctrl, err := maa.NewBlankController()

// 新 API：需要基于录制文件回放截图与操作
ctrl, err := maa.NewReplayController("./MaaRecording.jsonl")
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
// 新 API：提交失败直接返回 error；wrapper 预检失败时不会提交到原生层
taskJob, err := tasker.PostTask("entry", invalidOverride)
if err != nil {
    // 处理任务提交错误（如 JSON 序列化失败）；taskJob 为终态失败的 Job
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

- `Tasker.GetRecognitionDetail(recId int64) (*RecognitionDetail, error)`
- `Tasker.GetActionDetail(actionId int64) (*ActionDetail, error)`
- `Tasker.GetWaitFreezesDetail(wfId int64) (*WaitFreezesDetail, error)` 与 `WaitFreezesDetail`
- `CustomActionFunc`、`CustomRecognitionFunc`，用于将普通函数适配为 `CustomActionRunner` / `CustomRecognitionRunner`，可直接传给 `Resource.RegisterCustomAction` / `Resource.RegisterCustomRecognition`
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
- `NewMacOSController(windowID uint32, screencapMethod macos.ScreencapMethod, inputMethod macos.InputMethod) (*Controller, error)`，以及 `controller/macos` 子包中的 `ScreencapMethod` / `InputMethod` 枚举
- `NewAndroidNativeController(configJson string) (*Controller, error)`
- `NewReplayController(recordingPath string) (*Controller, error)`
- `NewRecordController(inner *Controller, recordingPath string) (*Controller, error)`
- OCR 颜色过滤：`OCRParam.ColorFilter` 字段 & `WithOCRColorFilter` 选项函数，指定 ColorMatch 节点名对图像进行颜色二值化后再送入 OCR 识别（适配 [MaaFramework#1145](https://github.com/MaaXYZ/MaaFramework/pull/1145)）
- Controller inactive：`Controller.PostInactive() (*Job, error)` 与 `CustomController.Inactive() bool`，用于在任务结束后恢复窗口/输入状态（适配 [MaaFramework#1155](https://github.com/MaaXYZ/MaaFramework/pull/1155)；Win32 控制器会恢复窗口与解除输入阻塞，其他控制器为 no-op）
- Screencap Action：新增 `ActionTypeScreencap` / `ActScreencap(ScreencapParam)`，支持在流水线动作中保存当前截图（适配 [MaaFramework#1165](https://github.com/MaaXYZ/MaaFramework/pull/1165)）
- Win32 截图方式：`ScreencapMethod` 新增 `ScreencapAll`、`ScreencapForeground`、`ScreencapBackground`，并支持对应字符串解析/序列化
- `Controller.SetMouseLockFollow(enabled bool) error`
- `ScreenshotResizeMethod` / `WithScreenshotResizeMethod(method)`，用于指定截图缩放插值方式
- Controller info：新增 `Controller.GetInfo() (string, error)`，以 JSON 格式获取控制器结构化信息（类型、构造参数、当前状态等）（适配 [MaaFramework#1167](https://github.com/MaaXYZ/MaaFramework/pull/1167)）
- `CustomController` 接口新增 `GetInfo() (string, bool)` 方法，自定义控制器可提供额外信息（适配 [MaaFramework#1167](https://github.com/MaaXYZ/MaaFramework/pull/1167)）
- `ControllerActionDetail` 新增 `Info map[string]any` 字段，控制器动作事件回调中包含控制器信息（适配 [MaaFramework#1167](https://github.com/MaaXYZ/MaaFramework/pull/1167)）
- WlRoots Controller：新增 NewWlRootsController(wlrSocketPath string) (*Controller, error)，支持通过 Wayland socket 创建 WlRoots 控制器（适配 [MaaFramework#1131](https://github.com/MaaXYZ/MaaFramework/pull/1131)）
- Controller relative move：新增 `Controller.PostRelativeMove(dx, dy int32) (*Job, error)`，支持提交相对光标移动事件（适配 [MaaFramework#1189](https://github.com/MaaXYZ/MaaFramework/pull/1189)）
- `CustomController.RelativeMove(dx, dy int32) bool` 与 `CustomController.Shell(cmd string, timeout int64) (string, bool)`，补齐自定义控制器的相对移动与 shell 能力
- `MacOSPermission`、`MacOSCheckPermission`、`MacOSRequestPermission`、`MacOSRevealPermissionSettings`，用于检查或申请 macOS Screen Recording / Accessibility 权限

## Fixed

- `Context.GetTasker`、`Tasker.GetController`、`Tasker.GetResource` 现在会缓存返回的 Go wrapper，避免重复调用后旧 wrapper 的方法调用崩溃 ([#41](https://github.com/MaaXYZ/maa-framework-go/issues/41))

## Performance

- ImageBuffer RGBA 路径优化：`Set` 对 `*image.RGBA` 走直通转换路径，减少高频图像写入时的额外转换与分配开销
