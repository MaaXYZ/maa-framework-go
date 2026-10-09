# v4.0.0-beta.18 → v4.0.0-beta.19 迁移指南

[English](../../en/migration/from-v4.0.0-beta.18.md) | 简体中文

本指南只覆盖 **v4.0.0-beta.18 → v4.0.0-beta.19** 的迁移。继续升级时，请阅读 [beta.19 → v4.0.0 指南](from-v4.0.0-beta.19.md)。

本版本的核心变化是**明确了原生句柄的所有权与生命周期**，并让 `Init` / `Release` 在并发与异常路径下安全。多数代码无需改动即可编译，但以下几类写法在运行时的行为会改变，请逐项检查。

## 升级前检查清单

- [ ] MaaFramework 升级到 **v5.13.0 或更高**（推荐 v5.14.0，`auto_up` 需要 v5.14.0）
- [ ] 不再使用 `NewWlRootsController`，改用 `NewLinuxController`
- [ ] 检查 `Destroy` 的返回值，尤其是 `defer` 的顺序
- [ ] 回调里拿到的 `Context` 不要在回调返回后继续使用（包括传给 goroutine）
- [ ] 不要对 `GetResource` / `GetController` / `Context.GetTasker` 等返回的对象调用 `Destroy`
- [ ] 使用 `Release` 的程序检查其返回值
- [ ] 自定义识别在未命中时若返回了非 nil 结果，确认这是预期行为

## MaaFramework 版本要求

beta.19 绑定了 MaaFramework v5.13.0 新增的导出符号（`MaaLinuxControllerCreate`、`MaaToolkitGamescopeInstance*`、`MaaToolkitPortalHelper*`），这些符号在所有平台的动态库中都会导出，因此**任何平台**都需要 MaaFramework ≥ v5.13.0。

`Init` 现在会逐库在注册前解析全部符号，缺失时返回 `*SymbolLookupError`（此前会在调用到缺失函数时才出问题）：

```go
if err := maa.Init(maa.WithLibDir(libDir)); err != nil {
	var symErr *maa.SymbolLookupError
	if errors.As(err, &symErr) {
		// symErr.LibraryName / LibraryPath / SymbolName 指明是哪个库缺了哪个符号，
		// 通常是 MaaFramework 版本过旧
	}
	return err
}
```

## 破坏性变更

### API 变更概览

| 变更类型 | 旧 API | 新 API |
|---------|--------|--------|
| 返回值 | `(*Tasker).Destroy()` | `(*Tasker).Destroy() error` |
| 返回值 | `(*Resource).Destroy()` | `(*Resource).Destroy() error` |
| 返回值 | `(*Controller).Destroy()` | `(*Controller).Destroy() error` |
| 返回值 | `(*AgentClient).Destroy()` | `(*AgentClient).Destroy() error` |
| 移除 | `NewWlRootsController(wlrSocketPath string, useWin32VkCode bool)` | `NewLinuxController(configJson string)` |
| 行为 | `Release() error` 仅在卸载失败时报错 | 仍有存活对象或 Agent Server 未关闭时返回 `ErrLibraryInUse` |

#### Destroy 返回 error

直接调用 `x.Destroy()` 或 `defer x.Destroy()` 仍能编译，但**错误会被静默丢弃**。以下写法会编译失败，需要改写：

- 把 `Destroy` 当作 `func()` 值使用，例如 `cleanups = append(cleanups, tasker.Destroy)`
- 自定义接口 `interface{ Destroy() }` 期望这些类型实现它

`Destroy` 可能返回的错误：

| 错误 | 触发条件 | 处理方式 |
|------|---------|---------|
| `ErrBorrowed` | 对借用视图（getter / 回调得到的对象）调用 `Destroy` | 去掉该调用，由所有者负责销毁 |
| `ErrBound` | Resource / Controller 仍被 Tasker 绑定，或被 AgentClient 持有；Tasker 仍注册为 AgentClient 的 sink | 按依赖顺序销毁：AgentClient → Tasker → Resource / Controller |
| `ErrInCallback` | 在对象自身的回调中销毁它 | 回调返回后再销毁 |
| `ErrInUse` | 还有方法调用或异步 Job 在执行，**即使 Job 已被丢弃** | 等待其结束（`Wait` / `PostStop().Wait()`）后重试 |

对所有者重复成功调用 `Destroy` 是安全的。

#### 销毁顺序：AgentClient → Tasker → Resource / Controller

旧版本中先销毁被绑定的对象也会释放原生对象；新版本会返回 `ErrBound`，对象**不会被释放**，并进一步导致 `Release` 返回 `ErrLibraryInUse`。

```go
// 迁移前：defer 按 LIFO 执行，res 会先于 tasker 销毁
tasker, _ := maa.NewTasker()
defer tasker.Destroy()
res, _ := maa.NewResource()
defer res.Destroy() // 执行时 tasker 仍绑定 res → ErrBound，被静默忽略
tasker.BindResource(res)

// 迁移后：先创建并 defer 被绑定的对象，最后 defer tasker
ctrl, _ := maa.NewAdbController(...)
defer ctrl.Destroy()
res, _ := maa.NewResource()
defer res.Destroy()
tasker, _ := maa.NewTasker()
defer tasker.Destroy() // 最先执行，解除绑定后 res / ctrl 才能销毁
tasker.BindResource(res)
tasker.BindController(ctrl)
```

使用 AgentClient 时同理：AgentClient 会持有其绑定的 Resource 以及注册为 sink 的 Resource / Controller / Tasker，需要最先销毁它。

**补充说明**：重新 `BindResource` / `BindController` 后，之前绑定过的对象仍由 Tasker 持有，直到 Tasker 销毁前对其调用 `Destroy` 都会返回 `ErrBound`。

#### 任务仍在运行时销毁 Tasker

Tasker 有未结束的 Job 时 `Destroy` 返回 `ErrInUse`。退出前请先停止并等待：

```go
tasker.PostStop().Wait()
if err := tasker.Destroy(); err != nil {
	log.Printf("destroy tasker: %v", err)
}
```

#### 借用视图不能 Destroy

构造函数（`NewTasker`、`NewResource`、各 `NewXxxController`）返回的是**所有者**；以下来源返回的是**借用视图**，对其调用 `Destroy` 返回 `ErrBorrowed`：

- `Tasker.GetResource()`、`Tasker.GetController()`
- `Context.GetTasker()`
- 事件回调中传入的对象

旧版本中对它们调用 `Destroy` 会直接释放原生句柄，可能导致重复释放或悬空句柄；新版本中该调用被拒绝。借用视图也不能作为参数传给 `BindResource` / `BindController`，借用的 Tasker 也不能再绑定其他对象，二者都返回 `ErrBorrowed`。

#### 回调 Context 在回调返回后失效

自定义识别、自定义动作、事件回调收到的 `*Context`（包括通过 `Clone` 得到的副本）只在本次回调期间有效。回调返回后：

- 返回 error 的方法（`RunTask`、`OverridePipeline` 等）返回 `ErrClosed`
- `GetTasker()`、`Clone()` 返回 nil
- `GetTaskJob()` 返回一个失败的 Job，其 `Error()` 为 `ErrClosed`

```go
// 迁移前：回调返回后继续使用 ctx
func (a *MyAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	go func() {
		ctx.RunTask("Next") // 回调已返回 → ErrClosed
	}()
	return true
}

// 迁移后：在回调内完成对 ctx 的使用
func (a *MyAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	_, err := ctx.RunTask("Next")
	return err == nil
}
```

#### 对象关闭后的调用

对象销毁后，返回 error 的方法会返回 `ErrClosed`；已提交的 Job 通过新增的 `(*Job).Error()` 暴露该错误。需要 Job 结果时，请在销毁所有者前调用 `Wait`。

#### Tasker 运行中重新绑定

Tasker 有待执行或运行中的任务时调用 `BindResource` / `BindController` 返回 `ErrTaskerRunning`。

#### Init / Release

- `Init` 与 `Release` 互相串行；**其他 MAA 函数不得与二者并发执行**。`IsInited` 可并发调用。
- `Release` 在以下情况返回 `ErrLibraryInUse`：
  - 仍有存活的原生对象：未 Destroy 或 Destroy 失败的 Tasker / Resource / Controller / AgentClient，以及未 Destroy 的 `PortalHelper`（其 `Destroy` 没有返回值）
  - Agent Server 尚未 `AgentServerShutDown`
  - 调用过 `AgentServerDetach`：此后**整个进程生命周期内** `Release` 都会被拒绝，即使之后调用了 `AgentServerJoin` / `AgentServerShutDown`。需要释放库的程序请不要 detach 服务线程
- 卸载失败时 `IsInited` 变为 false，可再次调用 `Release` 重试清理，然后再 `Init`。

```go
// Agent Server 进程
maa.AgentServerStartUp(socketID)
maa.AgentServerJoin()     // 阻塞到 AgentClient 调用 Disconnect 发出关闭请求、服务线程退出为止；它本身不会请求关闭
maa.AgentServerShutDown() // Join 返回后仍需调用，Release 才会放行
if err := maa.Release(); err != nil {
	log.Printf("release: %v", err) // ErrLibraryInUse 说明还有对象没有销毁
}
```

#### 移除 NewWlRootsController

MaaFramework 已将 `MaaWlRootsControllerCreate` 标记为废弃，Go 绑定移除了对应构造函数，改用 `NewLinuxController`，通过 JSON 配置选择截图与输入方式：

```go
// 旧
ctrl, err := maa.NewWlRootsController("/run/user/1000/wayland-0", true)

// 新（1 = MaaLinuxScreencapMethod_Wlr / MaaLinuxInputMethod_Wlr）
ctrl, err := maa.NewLinuxController(`{
	"screencap_method": 1,
	"input_method": 1,
	"wlr_socket_path": "/run/user/1000/wayland-0",
	"use_win32_vk_code": true
}`)
```

完整配置项（PipeWire、UInput、Libei 等）见 MaaFramework `MaaLinuxControllerCreate` 文档。该控制器仅在 Linux 可用。

## 行为变更

- **自定义识别未命中时保留结果**：`CustomRecognitionRunner.Run` 返回 `(result, false)` 且 `result` 非 nil 时，`Box` 与 `Detail` 现在会传给 MaaFramework，出现在识别详情中（此前被丢弃）。识别是否命中仍只由第二个返回值决定；返回 nil 始终视为未命中。若旧代码在未命中时随手返回了占位结果，请改为返回 `nil, false`。
- **动作详情为 `null` 时视为空**：`ActionDetail` 解析时 `"null"` 与 `""`、`"{}"` 一样得到 nil 结果，不再报解析错误。

## 新增

| 类别 | 新增 API | 说明 |
|------|---------|------|
| Controller | `NewLinuxController(configJson string) (*Controller, error)` | Linux 原生应用控制器，支持 Wlr / PipeWire 截图与 Wlr / UInput / Libei 输入 |
| Toolkit | `FindGamescopeInstances() ([]*GamescopeInstance, error)`、`GamescopeInstance` | 发现 gamescope 实例（显示号、EIS socket、PipeWire 节点），用于配置 Linux 控制器 |
| Toolkit | `NewPortalHelper() (*PortalHelper, error)`、`PortalHelper` | 通过 xdg-desktop-portal 获取 PipeWire 屏幕流，支持持久化 restore token；仅桌面 Linux 可用，其他平台（包括 Android）创建会失败 |
| Win32 | `win32.InputAnchoredTouch` | 注入触摸点而不移动光标或目标窗口；支持点击与滑动，不支持滚动和键盘输入 |
| Job | `(*Job).Error() error` | 返回 Job 无法提交或对象已关闭等原因 |
| ActionResult | `ClickKeyActionResult.AutoUp`、`TouchActionResult.AutoUp` | 对应 MaaFramework 动作详情中的 `auto_up` 字段；原生自动抬起功能需要 MaaFramework ≥ v5.14.0，更低版本该字段始终为 false |
| 错误 | `ErrClosed`、`ErrBorrowed`、`ErrBound`、`ErrInCallback`、`ErrInUse`、`ErrTaskerRunning`、`ErrLibraryInUse` | 见上文 |
| 错误 | `SymbolLookupError` | `Init` 找不到所需导出符号时返回 |
| 平台 | Android | `GOOS=android` 时加载 `lib*.so`；需要 `CGO_ENABLED=1` 并使用 Android NDK 工具链编译 |

## 修复

- 明确原生句柄的所有权，修复借用对象被重复释放、对象销毁后仍被回调访问等问题（#46）
- `Init` / `Release` 在并发调用和加载失败路径下安全，失败后可重试（#47）
- 自定义识别未命中时保留识别详情（#42）
