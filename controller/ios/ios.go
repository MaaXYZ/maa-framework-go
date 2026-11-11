// Package ios defines capture flags and transport selection for iOS devices
// controlled through XDeviceBridge, independently of the host operating system.
package ios

// ScreencapMethod selects the permitted iOS capture backends with bitwise OR.
type ScreencapMethod uint64

const (
	// ScreencapNone permits no capture backend.
	ScreencapNone ScreencapMethod = 0
	// ScreencapStream permits decoded video frames without screenshot fallback.
	ScreencapStream ScreencapMethod = 1
	// ScreencapScreenshotService permits lossless screenshots without video.
	ScreencapScreenshotService ScreencapMethod = 1 << 1
	// ScreencapAll permits every backend, preferring video with screenshot fallback.
	ScreencapAll = ^ScreencapNone
	// ScreencapDefault selects every backend.
	ScreencapDefault = ScreencapAll
)

// ConnectionMethod selects a transport independently of wireless credentials.
type ConnectionMethod int32

const (
	// ConnectionAuto lets XDeviceBridge select the transport.
	ConnectionAuto ConnectionMethod = 0
	// ConnectionUSB selects USB transport.
	ConnectionUSB ConnectionMethod = 1
	// ConnectionWiFi selects Wi-Fi transport.
	ConnectionWiFi ConnectionMethod = 2
)
