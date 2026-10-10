# Changelog

## 变更总览（v3.6.0-beta.5 → v4.0.0，发布准备草稿）

本总览及下一节以提交 [`fb84de6`](https://github.com/MaaXYZ/maa-framework-go/tree/fb84de66f3db7514b15d0063d515cd1272d558ff) 为固定核验终点。`v4.0.0` 尚未打 tag；正式版确定后需复核终点差异，再冻结这份里程碑记录。

### 破坏性变更与升级入口

- Go 模块由 `/v3` 升为 `/v4`；构造、配置、查询与运行接口采用 Go `error`，异步 `Post*` 也直接报告提交错误。升级时分别检查调用错误与 Job 的执行状态。
- 原生对象明确所有者与借用视图，`Destroy` 返回错误并保护仍被持有或使用的句柄；回调的 Context 仅在该次回调内有效。
- Pipeline 构造器移除 `Node` 类型前缀与大量配置 option，改用参数结构及节点链式设置。时间、可选指针、类别选择、锚点与列表编码均有需要迁移的变化。
- v3 用户见 [v3 → v4 迁移指南](docs/zh/migration/from-v3.md)；v4 用户按起点选择 [beta.18 → beta.19](docs/zh/migration/from-v4.0.0-beta.18.md) 或 [beta.19 → 发布准备快照](docs/zh/migration/from-v4.0.0-beta.19.md)。

### 错误处理模型

- 构造失败、参数编码失败、原生提交失败及详情不可用通过 `error` 报告；`Job.Error` 保留提交和生命周期诊断，异步执行结果仍由状态表示。
- 初始化提供库加载与符号缺失诊断，失败时尝试清理，卸载失败可重试。详单见下方「模块路径与错误返回」、beta.19 节及发布准备节的破坏性变更与修复。

### 所有权、生命周期与回调

- 按对象依赖收束工作并销毁，再卸载动态库；借用对象、绑定关系、活动调用、Job 与回调均参与销毁保护。
- 后续修复加强 Job 并发等待、停止后的空闲确认、回调清理与注册串行化。行为契约见固定快照的 [包文档源码](https://github.com/MaaXYZ/maa-framework-go/blob/fb84de66f3db7514b15d0063d515cd1272d558ff/doc.go)，区间详单见 beta.19 与发布准备节。

### Pipeline 配置与详情

- 类型化构造器生成嵌套 Pipeline v2 JSON，读取兼容 v1 与协议简写；未知类型可保存原始参数，原生支持仍由所加载库决定。
- 配置补齐 OCR 颜色过滤、Screencap、压力、自动抬起、Shell 超时、DirectHit ROI、检测阈值和标签类别选择，明确省略、显式零与空集合的区别。
- 节点详情改为按需查询，补齐画面稳定等待的详情与事件，并修复任务、图像、结果及参数编解码。详单分别见三个区间的 Pipeline、详情、新增与修复条目。

### 控制器与平台

- 增加 macOS、Android Native、Linux、录制与回放控制器及权限、gamescope、portal 工具；Linux 的终点入口为 `NewLinuxController`，中途新增的 `NewWlRootsController` 已移除。
- 扩展相对移动、窗口恢复、控制器信息与 Win32 输入能力；截图设置支持组合、验证与覆盖参考宽高。Android 增加动态库加载分支。
- 图像使用 RGBA，提供可复用截图内存与像素转换快速路径。详单见早期区间的控制器与性能、beta.19 的平台条目及发布准备节。

### 命名对齐与开发工具

- 修正节点、图像覆盖、推理设备与 Win32 枚举命名，任务和识别结果结构对齐原生 API；详单见早期区间的「详情、回调与命名」。
- API 检查工具覆盖原生符号与签名、回调 ABI、常量、事件和可选 Pipeline schema，使用同一原生版本的输入核验绑定；详单见早期区间的开发工具及发布准备节。

## v4.0.0（发布准备草稿）

本节记录 `v4.0.0-beta.19` → `fb84de6` 的净变化，完整差异见 [固定提交对比](https://github.com/MaaXYZ/maa-framework-go/compare/v4.0.0-beta.19...fb84de66f3db7514b15d0063d515cd1272d558ff)。这不是已发布正式版记录。升级操作见 [迁移指南](docs/zh/migration/from-v4.0.0-beta.19.md)；API 契约链接固定快照源码，正式 tag 确定后再切换到对应 godoc。

### 破坏性变更

- Controller 的全部 `Post*` 与 Resource 的 `PostBundle`、`PostOcrModel`、`PostPipeline`、`PostImage` 由 `*Job` 改为 `(*Job, error)`；Tasker 的 `PostTask`、`PostRecognition`、`PostAction`、`PostStop` 由 `*TaskJob` 改为 `(*TaskJob, error)`。提交失败同时返回非 nil 的终态失败 Job 和 `error`，原生返回无效提交 ID 也立即报告错误。
- `Tasker.PostTask` 及 `Context.RunTask`、`RunRecognition`、`RunAction`、`WaitFreezes` 不再把参数编码失败静默替换为 `{}`，改为返回错误并跳过原生执行。
- `LongPressParam.Duration`、`LongPressKeyParam.Duration` 由 `time.Duration` 改为 `*time.Duration`；神经网络分类与检测的 `Expected` 由 `[]int` 改为 `ClassSelectors`。
- TemplateMatch / FeatureMatch 的 `Template`、OCR 的 `Expected`、神经网络分类/检测的 `Labels` 由 `[]string` 改为 `StringList`。这些字段及神经网络 `Expected`、TemplateMatch `Threshold`、Swipe `End` 的非 nil 空列表现在写出 `[]`；`Node.Anchor` 的非 nil 空 map 写出 `{}`。
- 动作/识别的值类型 offset 与 `WaitFreezesParam.TargetOffset` 的零 `Rect` 现在省略，不再写出零数组；`Rect` 解码限定为 2 或 4 个整数，点简写扩展为 1×1 框。`Target` 编解码拒绝 false、null 与无效形状。
- `Node.UnmarshalJSON` 替换整个节点，未出现字段及 `Name` 重置；`CustomActionParam.CustomActionParam` 中的数字解码为 `json.Number`，不再使用 `float64`。
- `SetScreenshot` 从仅应用最后一项改为组合设置，拒绝重复/冲突选项、非正目标尺寸和无效插值方法。
- `GetRecognitionDetail`、`GetActionDetail`、`GetWaitFreezesDetail` 无可用详情时由 `(nil, nil)` 改为返回错误。
- Agent Server 增加生命周期与配置状态检查：运行、已 Join 或已 detach 时不允许重新启动或配置；未 detach 时关闭后进入永久关闭态，不能通过 `Release` / `Init` 重启。具体边界见 [Agent Server API](https://github.com/MaaXYZ/maa-framework-go/blob/fb84de66f3db7514b15d0063d515cd1272d558ff/agent_server.go)。
- `Resource.UseCoreml` 已移除。上游在 MaaFramework v5.14.3 弃用 `MaaInferenceExecutionProvider_CoreML`，MaaDeps 不再分发 CoreML 提供程序，Go 侧不再暴露该入口；需要 GPU 推理时改用 `Resource.UseWebgpu`。`MaaDef.h` 仍保留数值 3 供 C++ 调用方编译，因此该枚举的后续取值未重新编号。

### 新增

- `RawActionParam`、`RawRecognitionParam` 保留未知类型的参数 JSON；已知类型的参数错误仍报告失败，不回退到 raw。
- `ClassSelector`、`ClassSelectors`、`ClassIndex`、`ClassLabel`，支持整数或标签及混合类别列表；`StringList` 支持单字符串输入。
- Click / LongPress / Swipe / MultiSwipe 的 `Pressure *int`；TouchDown / TouchMove / KeyDown / KeyUp 的 `AutoUp *bool`；`ShellParam.ShellTimeout *time.Duration`；`DirectHitParam.ROI` / `ROIOffset`；`NeuralNetworkDetectParam.Threshold`。
- `TemplateMatchMethodSQDIFF_NORMED`、`FeatureMatchMethodSURF`；SURF 需要原生库支持 OpenCV xfeatures2d。
- `Pipeline.UnmarshalJSON` 从节点名映射读取 Pipeline、补上 `Node.Name`、跳过 `$` 元数据并规范化锚点简写；Node 与参数解码兼容 v1 扁平格式及协议的单项/列表简写，编码统一为规范 v2。
- `WithScreenshotTargetExpand` 按比例覆盖参考宽高；`Controller.SetBackgroundManagedKeys` 配置 Win32 后台托管键；`ControllerFeatureNoScalingTouchPoints` 禁用自定义控制器的触摸点自动缩放。
- `EventNodeWaitFreezes`、`NodeWaitFreezesDetail`、可选 `ContextWaitFreezesEventSink` 与 `OnNodeWaitFreezesInContext`；`ResourceLoadingDetail.Type` 区分加载类型。
- API 检查工具扩展回调 ABI、结构布局与 trampoline、更多常量组、事件分发及可选 Pipeline v2 类型/字段检查；新增 `--pipeline-schema` 与带理由的 exclusions，强化配置和输入验证。
- `Resource.UseWebgpu(deviceID InferenceDevice)` 选择 WebGPU 执行提供程序，`deviceID` 为 WebGPU 设备 id，`InferenceDeviceAuto` 交由框架选择（设备 0）。该提供程序要求 MaaFramework v5.14.3 或更高；更低版本没有对应枚举值，设置后加载模型时不会选出该提供程序。

### 修复

- Job / TaskJob 的等待与状态查询增加并发保护，多个等待者共享完成结果；停止使 Job ID 无效后，销毁仍需确认原生工作空闲，避免提前释放。契约见 [Job](https://github.com/MaaXYZ/maa-framework-go/blob/fb84de66f3db7514b15d0063d515cd1272d558ff/job.go) 与各对象的 Destroy 注释。
- 事件与自定义控制器回调关联所有者，销毁等待已接纳回调结束并阻止后续用户回调；自定义控制器回调保留至原生销毁完成，使 KeyUp / TouchUp 可在清理阶段释放输入。
- sink 与 Resource runner 注册操作串行化；原生注册失败时清理 Go 回调记录。Resource / AgentServer 拒绝 nil runner；Resource 注销未知名称成为成功空操作，外部注册的名称保留并返回错误。
- 图像、矩形等缓冲区创建和写入失败传播到调用方；空图像返回错误，空 buffer 读取返回真正的 nil 图像。
- Action / Recognition 缺少 `param` 时读取对象内平铺参数；And / Or 内联识别输出正确的 `recognition` 嵌套对象；补齐 `key_code`、OCR `text` 兼容及简写解码，修复复用参数对象时的旧值残留。Swipe 动作结果解析失败时不再部分修改接收值。
- 无节点任务保留 Entry / Status，节点详情跳过 ID 为 0 的子详情；Shell 输出及自定义识别 detail 保留内嵌 NUL。
- AgentClient 正的不足 1ms 通信超时向上取整为 1ms；Windows 空库目录初始化保留当前 DLL 搜索配置；Win32 截图方式输出对齐 `DXGI_DesktopDup` / `DXGI_DesktopDup_Window`，解析兼容旧名称。

## [v4.0.0-beta.19](https://github.com/MaaXYZ/maa-framework-go/compare/v4.0.0-beta.18...v4.0.0-beta.19)

本节记录 v4.0.0-beta.18 → v4.0.0-beta.19 的净变化。升级操作见 [迁移指南](docs/zh/migration/from-v4.0.0-beta.18.md)。

### 破坏性变更

- `Tasker.Destroy()`、`Resource.Destroy()`、`Controller.Destroy()`、`AgentClient.Destroy()` 由无返回值改为返回 `error`。
- 原生对象引入所有者与借用视图区分，并增加销毁保护：借用对象、仍被绑定或持有的对象、存在活动调用或未完成 Job 的对象、回调中的对象不再直接释放，分别通过 `ErrBorrowed`、`ErrBound`、`ErrInUse`、`ErrInCallback` 报告。销毁后的访问增加 `ErrClosed` 检查。具体边界见 [Tasker.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.19#Tasker.Destroy)、[Resource.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.19#Resource.Destroy)、[Controller.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.19#Controller.Destroy)、[AgentClient.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.19#AgentClient.Destroy)。
- 回调收到的 `Context` 及其 `Clone` 在回调返回后失效，后续操作受到生命周期检查；使用边界见 [Context](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.19#Context)。
- Tasker 有待执行或运行中的任务时，`BindResource`、`BindController` 改为返回 `ErrTaskerRunning`。
- `Release` 在原生对象仍存活、Agent Server 尚未关闭或已 detach 时改为返回 `ErrLibraryInUse`；detach 后的卸载限制见 [Release](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.19#Release)。
- 移除 `NewWlRootsController` 及对应原生绑定，新增 `NewLinuxController(configJson string) (*Controller, error)`，通过 JSON 配置选择 Linux 截图与输入方式，包括 WlRoots、PipeWire 和 libei。

### 新增

- `Job.Error() error`，提供提交或生命周期错误查询；`TaskJob.Error()` 扩展为报告所属对象的生命周期错误。接口说明见 [Job](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.19#Job)。
- `SymbolLookupError`，提供缺失原生符号的库名、库路径、符号名、版本匹配说明和底层错误。
- `GamescopeInstance` 与 `FindGamescopeInstances() ([]*GamescopeInstance, error)`，提供运行中实例的显示编号、PipeWire 节点 ID 和 EIS socket 路径。
- `PortalHelper` 与 `NewPortalHelper() (*PortalHelper, error)`，支持打开 xdg-desktop-portal ScreenCast 流；提供 `OpenStream`、`Persist` / `SetPersist`、`PipeWireFD`、`PipeWireNodeID`、`RestoreToken` / `SetRestoreToken` 和 `Destroy`。接口说明见 [PortalHelper](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.19#PortalHelper)。
- Win32 `InputAnchoredTouch` 输入方式，支持不移动光标或窗口的触摸点击与滑动，补充对应的字符串转换与解析。
- Android 动态库加载支持：四个 MaaFramework 库的名称选择增加 `GOOS=android` 分支，加载对应的 `.so` 文件。
- `ClickKeyActionResult.AutoUp`、`TouchActionResult.AutoUp`，补充原生动作详情中的 JSON `auto_up` 字段。

### 修复

- 原生对象的重复或并发销毁只执行一次清理，避免重复释放原生句柄。
- Tasker 的绑定与提交操作串行化，避免任务提交与重新绑定交错；AgentClient 保留绑定资源和已注册 sink 对象的句柄引用。
- `Init`、`Release` 互相串行化，`IsInited` 增加同步保护；并发边界见 [Init](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.19#Init)。
- 动态库逐库预检所需符号后再注册，缺失符号与注册 panic 转为初始化错误；失败时清理当前库及此前加载的库，并保留清理失败的句柄，修正部分初始化或卸载状态影响后续初始化的问题。
- 非空库目录先解析为绝对路径，修复 `WithLibDir(".")` 在 Unix 上退化为默认搜索路径的问题；macOS 显式指定的库文件不存在时返回加载错误，并将符号查找限定到目标库。
- 自定义识别未命中时保留非空结果中的 Box 与 Detail，支持返回未命中诊断信息；结果语义见 [CustomRecognitionRunner](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.19#CustomRecognitionRunner)。
- `InlineSubRecognition` 解码兼容包含 `recognition` 嵌套对象的子识别 JSON。
- 动作详情为 JSON 字面值 `null` 时返回空动作结果，避免继续按具体动作类型解码。
- API 检查工具将 C `int` 与 Go `int32` 对齐，修正原生函数签名检查误报。

## v3.6.0-beta.5 → v4.0.0-beta.18 的变更

本节记录两个 tag 之间的净变化，完整差异见 [v3.6.0-beta.5…v4.0.0-beta.18](https://github.com/MaaXYZ/maa-framework-go/compare/v3.6.0-beta.5...v4.0.0-beta.18)。

### 破坏性变更

#### 模块路径与错误返回

- Go 模块路径由 `github.com/MaaXYZ/maa-framework-go/v3` 改为 `github.com/MaaXYZ/maa-framework-go/v4`。
- `NewTasker`、`NewResource`、`NewAdbController`、`NewPlayCoverController`、`NewWin32Controller`、`NewGamepadController`、`NewCustomController`、`NewBlankController` 的返回值由 `*T` 改为 `(*T, error)`。
- Agent 客户端构造入口合并为 `NewAgentClient(opts ...AgentClientOption) (*AgentClient, error)`，原标识符参数和 `NewAgentClientTcp` 入口由 `WithIdentifier`、`WithTcpPort` 选项替代。
- 下列设置、查询与运行接口改用 Go 的 `error` 返回失败原因：

| 组件 | 返回值变化与涉及接口 |
| --- | --- |
| AgentClient | `BindResource`、`RegisterResourceSink`、`RegisterControllerSink`、`RegisterTaskerSink`、`Connect`、`Disconnect`、`SetTimeout`：`bool` → `error`；`Identifier`、`GetCustomRecognitionList`、`GetCustomActionList`：`(T, bool)` → `(T, error)` |
| AgentServer | `AgentServerRegisterCustomRecognition`、`AgentServerRegisterCustomAction`、`AgentServerStartUp`：`bool` → `error` |
| Context | `RunTask`、`RunRecognition`、`RunAction`、`RunRecognitionDirect`、`RunActionDirect`：`*Detail` → `(*Detail, error)`；`OverridePipeline`、`OverrideNext`、`OverrideImage`、`SetAnchor`、`ClearHitCount`：`bool` → `error`；`GetNodeJSON`、`GetAnchor`、`GetHitCount`：`(T, bool)` → `(T, error)` |
| Controller | `GetShellOutput`、`GetUUID`：`(string, bool)` → `(string, error)`；`GetResolution` 的 `ok bool` 改为 `err error`；`CacheImage`：`image.Image` → `(image.Image, error)` |
| Resource | `UseCPU`、`UseDirectml`、`UseCoreml`、`UseAutoExecutionProvider`、`RegisterCustomRecognition`、`UnregisterCustomRecognition`、`ClearCustomRecognition`、`RegisterCustomAction`、`UnregisterCustomAction`、`ClearCustomAction`、`OverridePipeline`、`OverrideNext`、`Clear`：`bool` → `error`；`GetNodeJSON`、`GetHash`、`GetNodeList`、`GetCustomRecognitionList`、`GetCustomActionList`、`GetDefaultRecognitionParam`、`GetDefaultActionParam`：`(T, bool)` → `(T, error)` |
| Tasker / TaskJob | `BindResource`、`BindController`、`ClearCache`、`TaskJob.OverridePipeline`：`bool` → `error`；`GetLatestNode`、`TaskJob.GetDetail`：`*Detail` → `(*Detail, error)` |
| 全局配置 / Toolkit | `SetLogDir`、`SetSaveDraw`、`SetStdoutLevel`、`SetDebugMode`、`SetSaveOnError`、`SetDrawQuality`、`SetRecoImageCacheLimit`、`LoadPlugin`、`ConfigInitOption`：`bool` → `error`；`FindAdbDevices`、`FindDesktopWindows`：列表返回值 → `(列表, error)` |

#### 初始化与控制器配置

- `InitConfig` 改为未导出的配置类型，`InitOption` 不再支持包外自行定义配置函数。`Init` 只应用显式传入的日志、调试与插件选项，不再隐式设置日志目录、日志级别等全局选项；重复 `Init` 和未初始化时的 `Release` 改为返回 `nil`，移除 `ErrAlreadyInitialized`、`ErrNotInitialized`。接口说明见 [Init](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.18#Init) 与 [InitOption](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.18#InitOption)。
- 移除 `Controller.SetScreenshotTargetLongSide`、`SetScreenshotTargetShortSide`、`SetScreenshotUseRawSize`，统一为 `SetScreenshot(opts ...ScreenshotOption) error`，由对应的 `WithScreenshotTargetLongSide`、`WithScreenshotTargetShortSide`、`WithScreenshotUseRawSize` 配置；新增 `ScreenshotResizeMethod` 和 `WithScreenshotResizeMethod`，支持最近邻、线性、三次、区域和 Lanczos4 插值。beta.18 的选项应用规则见 [ScreenshotOption](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.18#ScreenshotOption)。
- `CustomController` 接口新增必须实现的 `RelativeMove(dx, dy int32) bool`、`Shell(cmd string, timeout int64) (string, bool)`、`Inactive() bool`、`GetInfo() (string, bool)`。
- 移除 `CarouselImageController` 与 `NewCarouselImageController`。
- 图像读取返回的具体类型由 `*image.NRGBA` 改为 `*image.RGBA`，影响 `Controller.CacheImage`、识别详情中的 `Raw` / `Draws` 等图像。

#### Pipeline 配置 API

| 原 API | 新 API |
| --- | --- |
| `NodeAction` / `NodeActionType` / `NodeActionParam` | `Action` / `ActionType` / `ActionParam` |
| `NodeRecognition` / `NodeRecognitionType` / `NodeRecognitionParam` | `Recognition` / `RecognitionType` / `RecognitionParam` |
| 带 `Node` 前缀的动作参数、识别参数、排序/方法/检测器类型及枚举常量 | 去掉 `Node` 前缀的对应名称，例如 `NodeClickParam` → `ClickParam`、`NodeOCROrderBy` → `OCROrderBy` |
| `NodeMultiSwipeItem` / `NodeNextItem` | `MultiSwipeItem` / `NextItem` |
| `NodeWaitFreezes` | `WaitFreezesParam` |

- `NewNode(name, opts...)` 改为 `NewNode(name)`，移除 `NodeOption` 及节点配置的 `WithRecognition`、`WithAction`、`WithNext` 等选项函数。
- 动作构造函数 `ActClick`、`ActLongPress`、`ActSwipe`、`ActTouchDown`、`ActTouchMove`、`ActLongPressKey`、`ActScroll`、`ActCommand`、`ActCustom` 改为接收对应参数结构的值；`ActTouchUp` 改为接收 `contact int`。移除对应的配置 option 类型与 `With*` 函数，以及 `NewMultiSwipeItem`。
- 识别构造函数 `RecTemplateMatch`、`RecFeatureMatch`、`RecColorMatch`、`RecOCR`、`RecNeuralNetworkClassify`、`RecNeuralNetworkDetect`、`RecCustom` 改为接收对应参数结构的值，移除对应的配置 option 类型与 `With*` 函数。
- `LongPressParam.Duration`、`LongPressKeyParam.Duration`、`SwipeParam.Duration` / `EndHold`、`MultiSwipeItem.Duration` / `EndHold` / `Starting`，以及 `WaitFreezesParam.Time` / `RateLimit` / `Timeout` 从整数毫秒改为 `time.Duration` 或其切片；JSON 编码仍使用整数毫秒。
- 移除 `WaitFreezes` 构造函数、`WaitFreezesOption` 及 `WithWaitFreezes*`。`Context.WaitFreezes` 的可选 `...any` 参数改为固定的 `*WaitFreezesParam`，返回值由 `bool` 改为 `error`。
- `Node.Anchor` 字段及 `Node.SetAnchor` 的参数由 `[]string` 改为 `map[string]string`；`AddAnchor` 改为建立锚点到当前节点的映射，并新增 `SetAnchorTarget`、`ClearAnchor`。
- `Context.OverrideNext`、`Resource.OverrideNext` 的参数由 `[]string` 改为 `[]NextItem`；新增 `NextItem.FormatName()`，将跳回和锚点属性编码为原生接口使用的节点名称。
- And/Or 子识别统一为 `SubRecognitionItem`：`RecAnd`、`RecOr` 改为变参形式，`AndRecognitionParam.AllOf`、`OrRecognitionParam.AnyOf` 均为 `[]SubRecognitionItem`；新增 `Ref`、`Inline` 和 `InlineSubRecognition` 表达节点引用或内联识别。移除 `NodeAndRecognitionItem`、`AndItem`、`AndRecognitionOption`、`WithAndRecognitionBoxIndex`，新增 `Recognition.SetBoxIndex` 设置 box 索引。数据格式见 [SubRecognitionItem](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.18#SubRecognitionItem)。

#### 详情、回调与命名

- `TaskDetail.NodeDetails []*NodeDetail` 改为 `Nodes []NodeRef`，从立即展开全部节点详情改为按需查询；`NodeRef` 提供 `ID()` 与 `GetDetail()`。
- `TaskJob` 不再嵌入导出的 `*Job`，移除可直接访问的 `Job` 字段。
- `CustomActionArg.TaskDetail`、`CustomRecognitionArg.TaskDetail` 改为 `TaskID int64`；移除 `CustomAction`、`CustomRecognition` 别名。
- `ControllerEventSinkAdapter`、`ResourceEventSinkAdapter`、`TaskerEventSinkAdapter`、`ContextEventSinkAdapter` 改为未导出的内部类型。
- 以下名称和结果结构与原生 API 对齐：

| 原名称 / 字段 | 新名称 / 字段 |
| --- | --- |
| `Context.GetNodeData` | `Context.GetNode` |
| `Resource.OverriderImage` | `Resource.OverrideImage`，并由 `bool` 返回值改为 `error` |
| `InterenceDevice` / `InterenceDeviceAuto` | `InferenceDevice` / `InferenceDeviceAuto`，设备常量改为 `InferenceDevice` 类型 |
| Win32 `InputSendMessageWithCursorPosAndBlockInput` / `InputPostMessageWithCursorPosAndBlockInput` | `InputSendMessageWithWindowPos` / `InputPostMessageWithWindowPos`，对应字符串名称同步调整 |
| `RecognitionResults.Best []*RecognitionResult` | `Best *RecognitionResult` |
| `ShellActionResult.Timeout` / JSON `timeout` | `ShellTimeout` / JSON `shell_timeout` |
| `NodeNextListDetail.NextList` / JSON `next_list` | `List []NextItem` / JSON `list` |
| `NeuralNetworkClassifyResult.Raw` / `Probs` | 移除这两个字段 |

### 新增

#### 任务、节点与识别

- 公开 `Tasker.GetRecognitionDetail`、`GetActionDetail`、`GetNodeDetail`、`GetTaskDetail`，增加 `GetWaitFreezesDetail` 与 `WaitFreezesDetail`，可查询画面稳定等待的阶段、耗时、关联识别 ID 和 ROI。详情可用性与错误语义见 [Tasker](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.18#Tasker)。
- `Resource.GetNode`，以及 `Pipeline.GetNode`、`HasNode`、`RemoveNode`、`Clear`、`Len`，补充节点读取和管理能力。
- `CustomActionFunc`、`CustomRecognitionFunc`，将普通函数适配为自定义动作或识别 runner。
- 通用排序类型 `OrderBy`；OCR、神经网络分类与检测新增 `Expected` 排序常量。
- `OCRParam.ColorFilter`，通过引用 ColorMatch 节点配置支持 OCR 前的颜色二值化（[MaaFramework#1145](https://github.com/MaaXYZ/MaaFramework/pull/1145)）。
- Screencap 动作：`ActionTypeScreencap`、`ActScreencap`、`ScreencapParam`，以及 `ScreencapActionResult`、`ActionResult.AsScreencap`，支持在流水线动作中保存截图（[MaaFramework#1165](https://github.com/MaaXYZ/MaaFramework/pull/1165)）。

#### 控制器与平台

- `NewWlRootsController(wlrSocketPath string, useWin32VkCode bool) (*Controller, error)`，支持通过 Wayland socket 创建 WlRoots 控制器，并可使用 Win32 VK 键码（[MaaFramework#1131](https://github.com/MaaXYZ/MaaFramework/pull/1131)）。
- `NewMacOSController(windowID uint32, screencapMethod macos.ScreencapMethod, inputMethod macos.InputMethod) (*Controller, error)`，以及 `controller/macos` 包中的截图与输入方式枚举。
- `NewAndroidNativeController(configJson string) (*Controller, error)`，支持 Android 原生截图与输入。
- `NewRecordController(inner *Controller, recordingPath string) (*Controller, error)`、`NewReplayController(recordingPath string) (*Controller, error)`，支持控制器操作录制与回放。
- `Controller.PostRelativeMove(dx, dy int32) *Job`，支持提交相对光标移动（[MaaFramework#1189](https://github.com/MaaXYZ/MaaFramework/pull/1189)）；平台支持范围见 [PostRelativeMove](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.18#Controller.PostRelativeMove)。
- `Controller.PostInactive() *Job`，支持恢复 Win32 窗口与输入状态（[MaaFramework#1155](https://github.com/MaaXYZ/MaaFramework/pull/1155)）。
- `Controller.GetInfo() (string, error)` 与 `ControllerActionDetail.Info map[string]any`，提供控制器结构化信息（[MaaFramework#1167](https://github.com/MaaXYZ/MaaFramework/pull/1167)）。
- `Controller.SetMouseLockFollow(enabled bool) error`，支持 Win32 消息输入方式的鼠标锁定跟随；Win32 新增 `ScreencapAll`、`ScreencapForeground`、`ScreencapBackground` 截图方式与 `InputInterception` 输入方式。
- `MacOSPermission`、`MacOSCheckPermission`、`MacOSRequestPermission`、`MacOSRevealPermissionSettings`，支持检查或申请屏幕录制、辅助功能权限。

#### 全局配置与开发工具

- 可配置的 `JSONEncoder`、`JSONDecoder`，以及 `SetJSONEncoder`、`SetJSONDecoder`、`GetJSONEncoder`、`GetJSONDecoder`、`ResetJSONCodec`；初始化选项增加 `WithJSONEncoder`、`WithJSONDecoder`。
- `LibraryLoadError`，提供动态库加载失败时的库名、路径与底层错误。
- 参数校验错误 `ErrInvalidAgentClient`、`ErrInvalidResource`、`ErrInvalidController`、`ErrInvalidTasker`、`ErrInvalidTimeout`；全局配置与插件错误 `ErrEmptyLogDir`、`ErrSetLogDir`、`ErrSetSaveDraw`、`ErrSetStdoutLevel`、`ErrSetDebugMode`、`ErrSetSaveOnError`、`ErrSetDrawQuality`、`ErrSetRecoImageCacheLimit`、`ErrLoadPlugin`。
- `tools/api-check`，检查原生符号覆盖与签名、自定义控制器回调接口、ADB / Win32 控制方式枚举的一致性，并接入 CI。

### 修复

- `Context.GetNodeJSON` 改用 `MaaContextGetNodeData`，`Resource.OverrideNext` 改用 `MaaResourceOverrideNext`，修正调用了另一类句柄接口的问题。
- `Context.GetTasker`、`Tasker.GetController`、`Tasker.GetResource` 缓存返回的 Go 包装对象，修复重复获取后旧包装对象调用崩溃的问题（[#41](https://github.com/MaaXYZ/maa-framework-go/issues/41)）。
- Resource 自定义动作/识别的注册、替换、注销、清空在原生操作失败时保留已有 Go 注册记录，并清理失败的新回调；AgentServer 注册失败时也清理新回调。
- Controller sink 的移除和清空同步释放对应的 Go 回调记录。
- 自定义识别返回空结果时避免解引用空指针；自定义动作在 `reco_id == 0` 时跳过识别详情查询。
- 自定义识别结果的 `detail` 同时支持 JSON 字符串与对象；识别结果解析返回未知算法和 JSON 解码错误，并规范处理空内容与 `null`。
- `Tasker.GetNodeDetail` 在缺少识别或动作详情时保留节点详情。
- Context 运行方法与 `Tasker.PostTask` 的覆盖参数统一处理普通 nil、带类型的 nil 和序列化失败，回退为 `{}`；Context 增加原始 `[]byte` 覆盖参数透传。
- `Tasker.PostRecognition`、`PostAction` 在参数或识别详情序列化失败时保留错误并生成失败的 `TaskJob`，通过新增的 `TaskJob.Error()` 提供错误查询。接口说明见 [TaskJob](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4@v4.0.0-beta.18#TaskJob)。
- `NewNode` 初始化 `Attach`，`SetAttach` 浅复制传入的 map；`ActMultiSwipe` 深复制各项的切片，`RecColorMatch` 深复制颜色上下界的内层切片，减少调用方后续修改对配置的影响。
- 图像写入正确处理非零原点和带额外行间距的子图像；零宽或零高图像改为清空 buffer，避免索引空像素切片。
- 初始化加载库或应用选项失败时尝试卸载已加载的库；`Release` 按加载逆序卸载，收集卸载错误，并在卸载成功后清空绑定函数。

### 性能

- 新增 `Controller.CacheImageInto(dst *image.RGBA) (*image.RGBA, error)`，尺寸相同时复用图像内存。
- 图像读写改用直接像素转换，针对 RGBA / NRGBA、连续内存与不透明图像增加快速路径，减少通用转换与分配开销。
- 内部事件消息即时解析减少字符串复制；识别详情改用字节切片解析，减少结果解码时的重复转换。
