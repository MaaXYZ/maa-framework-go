# v4.0.0-beta.19 → v4.0.0 迁移指南

[English](../../en/migration/from-v4.0.0-beta.19.md) | 简体中文

本文覆盖 `v4.0.0-beta.19` → `v4.0.0` 的净变化。

v3 用户请使用 [v3 → v4 指南](from-v3.md)，beta.18 用户先按 [beta.19 指南](from-v4.0.0-beta.18.md) 处理所有权变化。完整条目见 [CHANGELOG](../../../CHANGELOG.md)。

v4.0.0 将最低 Go 版本由 1.24 提升到 1.25。先更新本地与 CI 工具链，再更新 Go 依赖：

```sh
go get github.com/MaaXYZ/maa-framework-go/v4@v4.0.0
```

同时使用一套与绑定兼容的 MaaFramework 动态库；不要把四个库混用不同版本。

## 1. 修改 Post 调用

所有异步提交方法现在直接返回提交错误。更新调用点，以及封装函数、接口、mock 和保存的方法值：

| 类型 | 方法 | 返回值变化 |
| --- | --- | --- |
| Controller | `PostConnect`、`PostClick`、`PostClickV2`、`PostSwipe`、`PostSwipeV2`、`PostClickKey`、`PostInputText`、`PostStartApp`、`PostStopApp`、`PostTouchDown`、`PostTouchMove`、`PostTouchUp`、`PostRelativeMove`、`PostKeyDown`、`PostKeyUp`、`PostScreencap`、`PostScroll`、`PostInactive`、`PostShell` | `*Job` → `(*Job, error)` |
| Resource | `PostBundle`、`PostOcrModel`、`PostPipeline`、`PostImage` | `*Job` → `(*Job, error)` |
| Tasker | `PostTask`、`PostRecognition`、`PostAction`、`PostStop` | `*TaskJob` → `(*TaskJob, error)` |

将 `tasker.PostTask("Entry").Wait()` 拆为提交检查与执行检查：

```go
func runTask(tasker *maa.Tasker, entry string) error {
	job, err := tasker.PostTask(entry)
	if err != nil {
		return fmt.Errorf("submit task: %w", err)
	}
	if !job.Wait().Success() {
		return fmt.Errorf("task failed: status=%v, error=%v", job.Status(), job.Error())
	}
	return nil
}
```

`Job.Error()` / `TaskJob.Error()` 仍然存在。提交失败同时返回一个非 nil、终态失败的 Job 和 `error`；即使只保存了 Job，也可查询提交错误。异步执行失败不一定提供 Go `error`，不能用 `job.Error() == nil` 代替成功状态检查。

停止任务时也要拆开调用：

```go
func stopTask(tasker *maa.Tasker) error {
	job, err := tasker.PostStop()
	if err != nil {
		return fmt.Errorf("submit stop: %w", err)
	}
	if !job.Wait().Success() {
		return fmt.Errorf("stop failed: status=%v, error=%v", job.Status(), job.Error())
	}
	return nil
}
```

停止可能使此前 Job ID 无效，旧 Job 的 `Wait` 返回不能单独证明所有工作已结束。保持所有者引用，处理 `Destroy` 返回的 `ErrInUse`，在工作收束后重试；成功销毁 Tasker 后再销毁其 Resource / Controller。具体边界见 [Controller.Destroy](../../../controller.go) 与 [包文档源码](../../../doc.go)。

`Tasker.PostTask`、`Context.RunTask` / `RunRecognition` / `RunAction` / `WaitFreezes` 的参数编码失败现在会返回错误并跳过原生执行，不再退回 `{}`。检查原来依赖静默回退的代码及自定义 JSON codec。图像输入也应提供非 nil、非空图像；图像/矩形缓冲区创建与写入失败会传播到公开调用。

## 2. 更新 Pipeline 参数类型

| 字段 | beta.19 | v4.0.0 |
| --- | --- | --- |
| `LongPressParam.Duration`、`LongPressKeyParam.Duration` | `time.Duration` | `*time.Duration` |
| 神经网络分类/检测的 `Expected` | `[]int` | `ClassSelectors` |
| TemplateMatch / FeatureMatch 的 `Template`、OCR 的 `Expected`、神经网络分类/检测的 `Labels` | `[]string` | `StringList` |

长按时长改用变量指针，整数类别改用 `ClassIndex`。例如：

```go
func buildRecognitions() (*maa.Action, *maa.Recognition) {
	duration := 500 * time.Millisecond
	press := maa.ActLongPress(maa.LongPressParam{Duration: &duration})
	classify := maa.RecNeuralNetworkClassify(maa.NeuralNetworkClassifyParam{
		Model:  "model.onnx",
		Labels: maa.StringList{"button", "other"},
		Expected: maa.ClassSelectors{
			maa.ClassIndex(0),
			maa.ClassLabel("button"),
		},
	})
	return press, classify
}
```

`StringList` 接受单字符串或字符串数组，编码为数组。多数直接赋给字段的 `[]string` 字面量仍可编译；显式使用 `StringList` 可清楚表达新的类型。类别选择支持数字与标签混合，数字限定在 int32 范围内。

本区间还补齐以下可选参数，需要时直接设置字段：

- Click / LongPress / Swipe / MultiSwipe 的 `Pressure *int`。
- TouchDown / TouchMove / KeyDown / KeyUp 的 `AutoUp *bool`；功能在 TouchDown / KeyDown 生效，另两个字段用于协议往返。
- `ShellParam.ShellTimeout *time.Duration`，JSON 单位为整数毫秒。
- `DirectHitParam.ROI`、`ROIOffset *Rect`，可配置直接命中的框。
- `NeuralNetworkDetectParam.Threshold []float64`，可配置检测置信阈值。

## 3. 检查 JSON 省略与清空语义

nil、零值与空集合现在有更明确的编码差异。尤其要检查以零值对象覆盖现有节点的代码：

| 字段 | 省略/继承 | 显式覆盖 |
| --- | --- | --- |
| 长按 Duration、ShellTimeout、Pressure、AutoUp 等可选指针 | nil 不写出字段 | 指向零或 false 的指针写出该值 |
| 字符串列表、神经网络 Expected、TemplateMatch Threshold | nil 不写出字段 | 非 nil 空列表写出 `[]`；是否能执行仍取决于原生参数要求 |
| `Node.Anchor` | nil 不写出字段 | 非 nil 空 map 写出 `{}`，清除节点的锚点配置，已建立的运行时锚点不受此操作清除 |
| 动作/识别的值类型 offset、`WaitFreezesParam.TargetOffset` | 零 `Rect` 不写出字段 | 需要显式清零时使用 raw 参数或原始 JSON |
| `DirectHitParam.ROIOffset` | nil 不写出字段 | `&maa.Rect{}` 写出 `[0,0,0,0]` |
| 神经网络检测的 `Threshold` | nil 和空列表都不写出字段 | `[]float64{0}` 写出零阈值 |

省略字段让 MaaFramework 沿用已有节点或配置默认值，不保证恢复内置默认值。例如，已有长按时长为 500ms 时，`Duration: nil` 不会把它改回 1000ms。

其他值类型 offset 无法通过 `Rect{}` 显式清零；可对已知动作使用 raw 参数：

```go
func clearClickOffset() *maa.Action {
	return &maa.Action{
		Type:  maa.ActionTypeClick,
		Param: maa.RawActionParam(`{"target_offset":[0,0,0,0]}`),
	}
}
```

`Target` 只接受 true、节点名字符串、2 或 4 个整数；false / null 现在报错。未指定的字段仍可留零值并被省略，但不要在必须编码的目标列表中放零 `Target`，也不要用 `NewTargetBool(false)`。`Rect` 解码只接受 2 或 4 个整数，`[x,y]` 规范化为 `[x,y,1,1]`；null、其他长度或非整数项报错。

直接复用 Node 解码时注意：新的 `Node.UnmarshalJSON` 替换整个节点，未出现的字段及 `Name` 重置，不再按旧对象合并。`Pipeline.UnmarshalJSON` 同样替换 Pipeline，并从节点 key 补上 `Name`。要做运行时增量覆盖，应使用 `OverridePipeline`，不要把 Go 解码当作合并操作。

以下输入会被规范化：v1 扁平 recognition/action、v2 缺少 param 时的平铺参数、单项 next/on_error、单模板/文本/类别、Swipe 单目标与时间、键列表、Command args、ColorMatch 颜色行、OCR replace、WaitFreezes 裸毫秒数。锚点字符串/字符串数组仅在整份 Pipeline 解码时解析到所在节点，直接 Node 解码仍要求 map。输出统一为规范 v2；已知参数中的未建模字段不会被保留。

And / Or 内联识别编码现在使用 `{"sub_name":"...","recognition":{"type":"...","param":{...}}}`；读取仍兼容旧平铺形式。若测试或持久化代码依赖旧 JSON 形状，应更新预期值。

`CustomActionParam.CustomActionParam` 解码后的数字改为 `json.Number`。将 `value.(float64)` 改为 `json.Number` 的转换并处理 `Int64` / `Float64` 错误；此变化不适用于所有 `any` 字段。

`RawActionParam` / `RawRecognitionParam` 还可保留未知类型的参数 JSON。它们不会让原生库支持未知算法，也不会提供未知类型的结果解析；已知类型的非法参数仍会报错。

## 4. 调整截图选项

beta.19 的 `SetScreenshot` 实际只应用最后一个选项。现在目标尺寸、插值与关闭 raw-size 可以组合，先验证所有参数，再依次应用。例如：

```go
func configureScreenshot(ctrl *maa.Controller) error {
	return ctrl.SetScreenshot(
		maa.WithScreenshotTargetLongSide(1280),
		maa.WithScreenshotResizeMethod(maa.ScreenshotResizeMethodLinear),
		maa.WithScreenshotUseRawSize(false),
	)
}
```

删除重复设置，避免把长边、短边、expand 目标混用，也不要在同次调用中把目标与 `WithScreenshotUseRawSize(true)` 组合。目标尺寸必须为正，插值枚举必须有效。Go 参数验证失败时不修改原生设置；原生 setter 中途失败时此前修改不会回滚。

新增的 `WithScreenshotTargetExpand(width, height)` 按比例放大或缩小至同时覆盖参考宽高，不裁剪、不拉伸。Win32 使用 `ScreencapMethod.String()` 持久化或比较名称时，更新为 `DXGI_DesktopDup` / `DXGI_DesktopDup_Window`；解析仍接受旧拼写。

## 5. 更换推理提供程序入口

`Resource.UseCoreml` 已移除：上游在 MaaFramework v5.14.3 弃用 `MaaInferenceExecutionProvider_CoreML`，MaaDeps 不再分发 CoreML 提供程序。调用该方法的代码改用 `Resource.UseWebgpu`，它接收 WebGPU 设备 id（`InferenceDeviceAuto` 交由框架选择，即设备 0）：

```go
// beta.19
err := res.UseCoreml(maa.InferenceDeviceAuto)

// 现在
err := res.UseWebgpu(maa.InferenceDeviceAuto)
```

`UseWebgpu` 需要 MaaFramework v5.14.3 或更高的原生库。更低版本不认识该取值，设置本身仍返回成功，但加载模型时会记录 invalid inference execution provider 并回退到 CPU；这与 `UseDirectml`、`UseCPU` 等入口的既有行为一致，所以请同时确认运行时实际选中的提供程序。

在 MaaFramework v5.14.3 的 MaaDeps 分发库中，`UseAutoExecutionProvider` 按 CUDA、DirectML、WebGPU 的优先级选择可用提供程序；没有可用项或所选项初始化失败时回退 CPU。MaaDeps 不再分发 CoreML；依赖自动选择结果的代码需要重新核对。

## 6. 检查注册、Agent Server 和详情查询

- 把事件 sink 和自定义 runner 配置安排在实例及关联 Tasker 空闲时。新增串行化与回调保护不替代原生执行期间的调用方协调，具体契约见 [包文档源码](../../../doc.go)。不要传 nil runner，包括带类型的 nil 函数。注销不存在的 Resource runner 名称现在成功返回；由 Go wrapper 外部注册的名称会保留并返回错误。
- 将 Agent Server 配置放在 `AgentServerStartUp` 之前。启动、已 Join 或已 detach 后不再允许重新启动或配置；未 detach 时，`AgentServerShutDown` 后进入永久关闭态，不能靠 `Release` / `Init` 重建服务，需要新服务时重启进程。Join / ShutDown 的安排见 [API 注释](../../../agent_server.go)。
- `GetRecognitionDetail`、`GetActionDetail`、`GetWaitFreezesDetail` 无可用详情时改为返回错误；调用方先处理 `error`，不再只判断 `detail == nil`。`NodeDetail.Recognition` / `Action` 仍可能为空。
- 自定义控制器支持 `ControllerFeatureNoScalingTouchPoints`，如需禁用触摸点自动缩放，可在 `GetFeature` 返回的 bitmask 中设置。原生销毁期间可能调用 `KeyUp` / `TouchUp`，不要在 `Destroy` 成功之前拆除实现所需状态。
- 订阅画面稳定等待事件可使用 `OnNodeWaitFreezesInContext` 或可选的 `ContextWaitFreezesEventSink`；已有 `ContextEventSink` 无需增加方法。
- Windows 默认/空库目录初始化现在保留已有 DLL 搜索配置。非空 `WithLibDir` 的设置仍不会随失败或 `Release` 恢复，依赖该进程设置的程序应自行协调。

如果使用 API 检查工具的自定义配置，核对仓库根目录相对路径，清理已失效的 blacklist / exclusions。使用 MaaFramework v5.14.3 头文件时，在 `constant_exclusions` 中为 `MaaInferenceExecutionProvider_CoreML` 添加非空理由，声明 Go 有意不再暴露该常量，参见 [CI 配置](../../../tools/api-check/config.ci.yaml)。这些排除使用精确 C 常量名，只抑制 Go 中缺失的常量报告，失效条目会使检查失败。启用 Pipeline schema 检查时，使用与头文件和原生库同一 MaaFramework release 的 schema，参见 [工具说明](../../../tools/api-check/README.md)。

## 7. 验证迁移

- [ ] 本地与 CI 使用 Go 1.25 或更新版本，更新依赖后运行 `go mod tidy`。
- [ ] 更新全部 `Post*` 签名，分别检查提交错误和异步状态。
- [ ] 检查指针、类别选择类型，以及零 offset、空集合、Node 解码替换行为。
- [ ] 更新依赖内联识别 JSON 形状或 `float64` 类型断言的代码和测试。
- [ ] 验证参数编码失败、图像为空、详情不可用及截图选项冲突等路径。
- [ ] 把 `UseCoreml` 调用改为 `UseWebgpu`，并确认原生库版本与运行时实际选中的提供程序。
- [ ] 验证停止与退出路径，处理 `Destroy` / `Release` 错误，不在 Agent Server 关闭后尝试重启。
- [ ] 运行项目的 `go build ./...`、`go vet ./...` 和相关测试。
