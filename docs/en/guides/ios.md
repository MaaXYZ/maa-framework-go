# iOS device discovery and control

English | [简体中文](../../zh/guides/ios.md)

iOS is the controlled device type, not the host operating system. The `controller/ios` package provides screenshot flags and connection choices on the binding's supported hosts.

Use all four native libraries from MaaFramework v5.16.0 or a newer compatible release. The iOS symbols are mandatory during `Init` on every host, even when using other controllers. An older library without these symbols causes `SymbolLookupError` and initialization rollback. The [version policy](../../../README.md#requirements) still tracks the latest published release, including prereleases; the README badge records the last binding release's tested baseline.

## Discover devices

After `maa.Init`, call `maa.FindIOSDevices()` for a default USB/Wi-Fi snapshot. To choose a connection, timeout, or cancellation:

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

Import `github.com/MaaXYZ/maa-framework-go/v4` as `maa` and its `controller/ios` package. Here `ctx` is a `context.Context`; pass `nil` instead of a callback when cancellation is unnecessary. Native code may invoke it from a worker thread, so it must be safe for concurrent calls, return promptly, and not panic or reenter MaaFramework. It is retained only until the synchronous discovery call returns.

`Warnings` and each device's `Connections` preserve native JSON arrays, including fields unknown to the binding. A native failure or cancellation may still return warnings; inspect the result before handling the error. Partial `Devices` are returned only if discovery succeeds but reading a later list entry fails. All returned strings are owned Go copies that survive destruction of native discovery storage.

Discovery does not pair devices, install a DDI, or start media. A record does not prove connectivity; `State == "connected"` only means xdb has an active connection/session. Native options include `timeout_ms` (0–60000, default 3000) and `server_port` (from `XDB_PORT` or 17423 by default). USB enumeration has a separate one-second budget.

## Create and connect a controller

Choose a nonempty UDID and call `maa.NewIOSController(udid, ios.ScreencapDefault, ios.ConnectionAuto, "{}")`. Screenshot methods are 64-bit flags (`ScreencapStream`, `ScreencapScreenshotService`); connections are `ConnectionAuto`, `ConnectionUSB`, and `ConnectionWiFi`.

Creation returns an owned, unconnected controller. Call `PostConnect`, check its error, and wait for its job to succeed before binding it to a tasker. Destroy the controller after its jobs and owning taskers finish. xdb manages credentials and DDI setup; see the [native controller contract](https://github.com/MaaXYZ/MaaFramework/blob/v5.16.0/include/MaaFramework/Instance/MaaController.h) for connection and video configuration.

