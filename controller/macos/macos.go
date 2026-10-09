// Package macos defines the screencap and input method flags for macOS
// controllers.
package macos

// ScreencapMethod defines the macOS screencap method.
//
// Select ONE method only.
type ScreencapMethod uint64

// InputMethod defines the macOS input method.
//
// Select ONE method only.
type InputMethod uint64

const (
	// ScreencapNone selects no macOS screencap method.
	ScreencapNone ScreencapMethod = 0
	// ScreencapScreenCaptureKit selects the ScreenCaptureKit method:
	// modern macOS screencap using ScreenCaptureKit. Fast, high
	// compatibility, requires Screen Recording permission, supports
	// background capture, macOS 14.0+.
	ScreencapScreenCaptureKit ScreencapMethod = 1

	// InputNone selects no macOS input method.
	InputNone InputMethod = 0
	// InputGlobalEvent selects the GlobalEvent input method: injects into
	// the global HID event stream via CGEventPost(kCGHIDEventTap),
	// dispatched by the OS to the front window. High compatibility,
	// requires Accessibility permission, no background support.
	InputGlobalEvent InputMethod = 1
	// InputPostToPid selects the PostToPid input method: directly sends to
	// the target process using CGEventPostToPid. Medium compatibility,
	// requires Accessibility permission, supports background input.
	InputPostToPid InputMethod = 1 << 1
)
