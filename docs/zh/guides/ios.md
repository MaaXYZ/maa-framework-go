# iOS 设备发现与控制

[English](../../en/guides/ios.md) | 简体中文

iOS 是被控设备类型，不是运行 Go 程序的宿主操作系统。`controller/ios` 包提供截图标志和连接方式，可在本绑定支持的宿主平台上使用。

四个原生库请使用同一套 MaaFramework v5.16.0 或更新的兼容发行版。所有宿主平台的 `Init` 都会加载 iOS 符号，即使只使用其他控制器。旧库缺少符号时会返回 `SymbolLookupError` 并回滚初始化。[版本策略](../../../README_zh.md#要求) 仍跟随最新已发布版本（含预发布版）；README 徽章记录的是上次绑定发布时已测试的基线。

## 发现设备

`maa.Init` 成功后，调用 `maa.FindIOSDevices()` 获取默认 USB/Wi-Fi 快照。需要选择连接方式、超时或取消时：

```go
result, err := maa.FindIOSDevicesWithOptions(
	ios.ConnectionAuto,
	`{"timeout_ms":3000}`,
	func() bool { return ctx.Err() != nil },
)
if result != nil {
	fmt.Println("discovery warnings:", result.Warnings)
}
if err != nil {
	return err
}
for _, device := range result.Devices {
	fmt.Println(device.Name, device.UDID, device.State, device.Connections)
}
```

将 `github.com/MaaXYZ/maa-framework-go/v4` 以 `maa` 导入，同时导入其 `controller/ios` 包。这里 `ctx` 是 `context.Context`；无需取消时传入 `nil`。原生代码可能从工作线程调用回调，因此它必须支持并发调用、及时返回，不得 panic 或重入 MaaFramework。回调仅保留到同步发现调用返回为止。

`Warnings` 和每个设备的 `Connections` 保留原生 JSON 数组，包括绑定尚不认识的字段。原生发现失败或取消时，结果仍可能包含警告，请先读取结果再处理错误。只有发现成功、但读取后续列表项失败时才会返回部分 `Devices`。返回的全部字符串都是 Go 拥有的副本，原生发现存储销毁后仍然有效。

发现过程不会配对设备、安装 DDI 或启动媒体。发现记录不保证设备可连接；`State == "connected"` 仅表示 xdb 存在活动连接或会话。原生选项包括 `timeout_ms`（0–60000，默认 3000）和 `server_port`（默认取 `XDB_PORT` 或 17423）。USB 枚举另有一秒预算。

## 创建并连接控制器

选择非空 UDID，调用 `maa.NewIOSController(udid, ios.ScreencapDefault, ios.ConnectionAuto, "{}")`。截图方式是 64 位标志（`ScreencapStream`、`ScreencapScreenshotService`）；连接方式可选 `ConnectionAuto`、`ConnectionUSB`、`ConnectionWiFi`。

创建返回一个由调用方管理、尚未连接的控制器。调用 `PostConnect`，检查错误并等待 Job 成功后再绑定 Tasker。所有 Job 和持有它的 Tasker 结束后再销毁控制器。设备凭据和 DDI 配置由 xdb 管理；连接及视频配置见[原生控制器契约](https://github.com/MaaXYZ/MaaFramework/blob/v5.16.0/include/MaaFramework/Instance/MaaController.h)。

