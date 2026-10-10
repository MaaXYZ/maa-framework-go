# v3 → v4.0.0 迁移指南

[English](../../en/migration/from-v3.md) | 简体中文

本指南覆盖 `v3.6.0-beta.5` → `v4.0.0` 的迁移。

已经使用 v4 的项目请按起点选择 [beta.18 → beta.19](from-v4.0.0-beta.18.md) 或 [beta.19 → v4.0.0](from-v4.0.0-beta.19.md)。变更详单见 [CHANGELOG](../../../CHANGELOG.md)，本指南按实际修改顺序组织。

## 1. 更新模块路径和原生库

需要 Go 1.25 或更新版本。把主包及控制器子包的 import 路径从 `/v3` 改为 `/v4`：

```go
import (
	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/MaaXYZ/maa-framework-go/v4/controller/adb"
)
```

更新 Go 依赖：

```sh
go get github.com/MaaXYZ/maa-framework-go/v4@v4.0.0
go mod tidy
```

更新全部 import 后，检查 `go.mod` 是否还保留不再使用的 `/v3` 依赖。

同时更新 MaaFramework 动态库。本绑定跟踪最新 MaaFramework（包括预发布版），不保证兼容较早版本；`MaaFramework`、`MaaToolkit`、`MaaAgentServer`、`MaaAgentClient` 四个库应来自同一兼容版本。缺库或缺符号时，检查 `LibraryLoadError` / `SymbolLookupError` 中的库路径与符号名，再更换整套库。

## 2. 区分调用错误与异步执行结果

构造函数现在返回 `(*T, error)`。此前返回 `bool` 或 `(T, bool)` 的设置、查询、绑定、注册等接口大多改为 `error` 或 `(T, error)`，不要继续用布尔判断检查结果。

| 原写法涉及的接口 | v4 处理方式 |
| --- | --- |
| `NewTasker`、`NewResource`、各控制器构造函数 | 检查第二个返回值，成功后才使用对象 |
| `BindResource`、`BindController`、`OverridePipeline`、自定义动作/识别注册、全局配置 | 检查返回的 `error` |
| 设备/窗口发现、节点/任务详情、资源 hash 和列表、截图读取 | 接收结果与 `error`；`Pipeline.GetNode` 仍返回 `(*Node, bool)` |
| `Context.RunTask` / `RunRecognition` / `RunAction` 及 Direct 版本 | 接收详情与 `error` |
| Controller / Resource / Tasker 的 `Post*` | 接收 Job 与提交 `error`，再等待并检查执行状态 |

例如，连接控制器时分别处理两个阶段：

```go
func connect(ctrl *maa.Controller) error {
	job, err := ctrl.PostConnect()
	if err != nil {
		return fmt.Errorf("submit connection: %w", err)
	}
	if !job.Wait().Success() {
		return fmt.Errorf("connection failed: status=%v, error=%v", job.Status(), job.Error())
	}
	return nil
}
```

`err == nil` 只表示提交成功。`Job.Error()` / `TaskJob.Error()` 保留，用于提交或生命周期错误；异步任务失败可能没有 Go `error`，仍要检查 `Wait().Success()`。原来的 `PostTask(...).Wait()`、`PostStop().Wait()` 等链式调用需要拆开；具体签名表与停止示例见 [beta.19 → v4.0.0 指南](from-v4.0.0-beta.19.md#1-修改-post-调用)。

覆盖参数仍可使用 JSON 字符串、`[]byte` 或可编码的 Go 值；nil 表示 `{}`。现在编码失败会返回错误并阻止执行，不再静默退回 `{}`。检查自定义 JSON codec 和包含不可编码字段的覆盖对象。

## 3. 调整对象清理和回调代码

先保存构造函数返回的所有者，并在成功初始化后立即安排覆盖提前失败路径的清理。退出前停止新提交，等待所有已提交的 Job 和回调结束；随后按 **AgentClient → Tasker → Resource / Controller → Release** 的顺序清理，并处理 `Destroy()` 的 `error`。

以下辅助函数接收所有者指针；它适合从一个集中清理的 defer 中调用，以覆盖部分构造成功的路径。调用前须收束异步工作，不能把销毁错误当成已清理成功：

```go
func closeOwned(client *maa.AgentClient, tasker *maa.Tasker, res *maa.Resource, ctrl *maa.Controller) error {
	var err error
	if client != nil {
		err = errors.Join(err, client.Destroy())
	}
	if tasker != nil {
		err = errors.Join(err, tasker.Destroy())
	}
	if res != nil {
		err = errors.Join(err, res.Destroy())
	}
	if ctrl != nil {
		err = errors.Join(err, ctrl.Destroy())
	}
	return errors.Join(err, maa.Release())
}
```

如果项目把 `Destroy` 存入 `[]func()` 或要求类型实现 `interface{ Destroy() }`，将其改为返回 `error` 的函数或接口，并在调用处处理错误。

移除对 getter 和事件回调返回的借用对象的 `Destroy`；它们由原所有者清理。不要把回调收到的 `Context` 或 `Clone` 保存到回调之外，跨 goroutine 使用时也必须在本次回调返回前完成。配置和注册变更安排在实例及关联 Tasker 空闲时，避免在回调中替换、移除或清空注册。

完整的所有权、并发与回调契约见 [包文档源码](../../../doc.go)，各类 `Destroy` 的可重试错误见对应 API 注释。

## 4. 改写 Pipeline 构造代码

### 类型命名和构造函数

| v3 API | v4 API |
| --- | --- |
| `NodeAction` / `NodeActionType` / `NodeActionParam` | `Action` / `ActionType` / `ActionParam` |
| `NodeRecognition` / `NodeRecognitionType` / `NodeRecognitionParam` | `Recognition` / `RecognitionType` / `RecognitionParam` |
| 带 `Node` 前缀的动作、识别参数和枚举 | 去掉前缀，如 `NodeClickParam` → `ClickParam` |
| `NodeNextItem` / `NodeMultiSwipeItem` / `NodeWaitFreezes` | `NextItem` / `MultiSwipeItem` / `WaitFreezesParam` |
| `NewNode(name, WithRecognition(...), WithAction(...), ...)` | `NewNode(name).SetRecognition(...).SetAction(...)` |
| `Act*` / `Rec*` 的参数配置 `With*` 函数 | 对应参数结构；保留的简便构造函数按其签名调用 |
| `NewMultiSwipeItem`、`WaitFreezes` 构造函数及相关选项 | 直接创建 `MultiSwipeItem`、`WaitFreezesParam` |

例如，改为字段初始化与链式节点设置：

```go
func buildPipeline() *maa.Pipeline {
	duration := 500 * time.Millisecond
	node := maa.NewNode("Press").
		SetRecognition(maa.RecTemplateMatch(maa.TemplateMatchParam{
			Template: maa.StringList{"button.png"},
		})).
		SetAction(maa.ActLongPress(maa.LongPressParam{
			Target:   maa.NewTargetBool(true),
			Duration: &duration,
		})).
		SetNext([]maa.NextItem{{Name: "Done"}}).
		SetPreDelay(100 * time.Millisecond)
	return maa.NewPipeline().AddNode(node).
		AddNode(maa.NewNode("Done").SetAction(maa.ActDoNothing()))
}
```

Go 构造器编码为嵌套的 Pipeline v2 `type` / `param` JSON。现有 v1 扁平 JSON 可被读取并规范化为 v2，无需仅为升级批量重写全部资源文件。不要依赖 JSON 字节完全一致；已知参数的未建模字段不会在解码后保留。未知动作/识别类型的参数可用 `RawActionParam` / `RawRecognitionParam` 保存，原生库仍须支持该类型。

### 时间、可选值与列表

- `LongPressParam.Duration`、`LongPressKeyParam.Duration` 为 `*time.Duration`；Swipe / MultiSwipe 的持续时间、结束停留及 MultiSwipe 起始时间，以及 `WaitFreezesParam` 时间字段使用 `time.Duration` 或其切片。JSON 仍以整数毫秒表示。节点的 `RateLimit`、`Timeout` 等字段仍是 `*int64` 毫秒值，也可继续通过接收 `time.Duration` 的 `Set*` 方法设置。
- `ShellParam.ShellTimeout` 为 `*time.Duration`。可选指针的 nil 表示不覆盖已有值或默认配置；指向零的指针显式写出零，不能把 nil 当作“恢复内置默认值”。
- 模板、OCR 期望文本及神经网络标签列表改用 `StringList`。神经网络 `Expected` 改用 `ClassSelectors`，数字类别使用 `ClassIndex(n)`，标签使用 `ClassLabel("name")`，如 `maa.ClassSelectors{maa.ClassIndex(0), maa.ClassLabel("button")}`。
- 检查原来靠零 `Target` / `Rect` 覆盖已有 ROI 或坐标的代码，以及非 nil 空列表的使用。具体字段的省略、清空和显式零规则见 [beta.19 → v4.0.0 指南](from-v4.0.0-beta.19.md#3-检查-json-省略与清空语义)，不要为所有字段套用一种规则。

### 锚点、next 与组合识别

- `Node.Anchor` 从 `[]string` 改为 `map[string]string`。用 `AddAnchor("name")` 指向本节点，用 `SetAnchorTarget("name", "TargetNode")` 指向指定节点；清空全部节点锚点配置使用 `SetAnchor(map[string]string{})`，从配置移除指定锚点使用 `RemoveAnchor("name")`。`ClearAnchor("name")` 设置空目标，安排节点执行后清除该运行时锚点。
- `Context.OverrideNext` / `Resource.OverrideNext` 改收 `[]NextItem`。用 `NextItem{Name: "name", JumpBack: true}` 或 `Anchor: true` 表示属性，不再传 `[]string`。
- And / Or 改用 `SubRecognitionItem`；使用 `RecAnd(...)` / `RecOr(...)` 的变参、`Ref("NodeName")` 和 `Inline(recognition)`。需要 box 索引时用 `Recognition.SetBoxIndex`，不再使用 `NodeAndRecognitionItem`、`AndItem` 或旧配置选项。
- `Context.WaitFreezes` 的参数改为固定的 `*WaitFreezesParam`，结果改为 `error`；调用细节见 [API 注释](../../../context.go)。

## 5. 更新详情读取和自定义扩展

| v3 用法 | v4 用法 |
| --- | --- |
| `TaskDetail.NodeDetails []*NodeDetail` | 遍历 `TaskDetail.Nodes []NodeRef`，按需调用 `NodeRef.GetDetail()` 并处理错误 |
| `TaskJob.Job` | 直接调用 `TaskJob` 的 `Wait`、状态、`Error` 等方法 |
| 自定义动作/识别参数的 `TaskDetail` | 读取 `TaskID`，在回调内按需用 `ctx.GetTasker().GetTaskDetail(arg.TaskID)` 查询 |
| `CustomAction` / `CustomRecognition` 别名 | 使用 `CustomActionRunner` / `CustomRecognitionRunner`；普通函数可用 `CustomActionFunc` / `CustomRecognitionFunc` |
| 导出的 `*EventSinkAdapter` | 实现相应 EventSink 接口，或使用 `On*` 注册方法 |
| `Context.GetNodeData` / `Resource.OverriderImage` | `Context.GetNode` / `Resource.OverrideImage` |
| `RecognitionResults.Best` 切片 | `*RecognitionResult`，访问前检查 nil |
| `ShellActionResult.Timeout` / `NodeNextListDetail.NextList` | `ShellTimeout` / `List []NextItem`，JSON 名称同步为 `shell_timeout` / `list` |
| `NeuralNetworkClassifyResult.Raw` / `Probs` | 删除对这两个已移除字段的访问 |
| 返回图像的具体类型 `*image.NRGBA` | `*image.RGBA`；需要具体类型时更新类型断言，通常使用 `image.Image` |

`CustomController` 需要额外实现 `RelativeMove(dx, dy int32) bool`、`Shell(cmd string, timeout int64) (string, bool)`、`Inactive() bool`、`GetInfo() (string, bool)`。实现必须支持原生线程并发调用；键盘/触摸抬起处理也可能发生在销毁阶段。

自定义识别未命中且没有诊断信息时返回 `nil, false`；返回非 nil 结果及 false 会保留其 Box / Detail。识别是否命中由布尔值决定。

## 6. 调整控制器、Agent 和初始化配置

- `Resource.UseCoreml` 已移除，使用 MaaFramework v5.14.3 或更新版本时，将调用改为 `Resource.UseWebgpu(maa.InferenceDeviceAuto)`。在 v5.14.3 的 MaaDeps 分发库中，Auto 按 CUDA、DirectML、WebGPU 的优先级选择可用提供程序；没有可用项或所选项初始化失败时回退 CPU。设备选择与旧原生库的行为见[推理提供程序迁移步骤](from-v4.0.0-beta.19.md#5-更换推理提供程序入口)。
- 截图的三个旧 `SetScreenshotTarget*` / `SetScreenshotUseRawSize` 方法改为 `SetScreenshot(WithScreenshot*...) error`。目标尺寸、插值和关闭 raw-size 可合并设置；使用前检查目标互斥和重复选项。示例见 [beta.19 → v4.0.0 指南](from-v4.0.0-beta.19.md#4-调整截图选项)。
- Win32 `InputSendMessageWithCursorPosAndBlockInput` / `InputPostMessageWithCursorPosAndBlockInput` 改为 `InputSendMessageWithWindowPos` / `InputPostMessageWithWindowPos`，字符串配置也要同步。`InterenceDevice` / `InterenceDeviceAuto` 拼写修正为 `InferenceDevice` / `InferenceDeviceAuto`。
- `CarouselImageController` / `NewCarouselImageController` 已移除。按用途选择自定义控制器或录制/回放控制器。v3 到 v4.0.0 的 Linux 新入口是 `NewLinuxController(configJson)`；无需经过已在中间版本新增后移除的 `NewWlRootsController`。
- Agent 构造统一为 `NewAgentClient(opts ...AgentClientOption) (*AgentClient, error)`。用 `WithIdentifier(id)` 替代位置参数，用 `WithTcpPort(port)` 替代 `NewAgentClientTcp`；以 `client.Identifier()` 的实际返回值配对 `AgentServerStartUp`，并处理两处错误。
- `InitConfig` 不再导出，使用提供的 `With*` 初始化选项。`Init` 不再隐式设置日志目录、日志级别等全局配置，需要的选项应显式传入；初始化成功后的重复 `Init` 是空操作，后来传入的选项不会应用。删除对已移除的 `ErrAlreadyInitialized` / `ErrNotInitialized` 的判断。

Agent Server 的启动、Join、关闭、Detach 应按 [API 注释](../../../agent_server.go) 安排；需要卸载库的程序应保持服务线程未 detach。Windows 使用非空 `WithLibDir` 会影响进程 DLL 搜索配置，失败回滚与 `Release` 不恢复此设置，见 [WithLibDir](../../../maa.go)。

## 7. 验证迁移

- [ ] 本地与 CI 使用 Go 1.25 或更新版本，更新依赖后运行 `go mod tidy`。
- [ ] 所有包路径已换成 `/v4`，构造、绑定、配置、查询和提交错误均已处理。
- [ ] 对连接、资源加载与最终任务分别检查异步执行结果。
- [ ] 用真实资源比较关键节点的 JSON 与执行行为，检查省略字段、显式零、空列表和锚点。
- [ ] 检查回调内详情查询及 `Context` 使用，退出时等待工作结束并确认 `Destroy` / `Release` 成功。
- [ ] 执行项目的 `go build ./...`、`go vet ./...` 和相关测试，验证正常完成、执行失败及提前退出路径。
