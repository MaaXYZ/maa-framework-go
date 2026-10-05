// Package win32 defines the screencap and input method flags for Win32
// controllers, with name parsing helpers.
package win32

import (
	"fmt"
	"strconv"
	"strings"
)

// ScreencapMethod defines the Win32 screencap method flags.
//
// Use bitwise OR to combine methods; MaaFramework tests them and uses the
// fastest available one.
type ScreencapMethod uint64

// InputMethod defines the Win32 input method.
//
// Select ONE method only; do not combine with bitwise OR.
type InputMethod uint64

const (
	// ScreencapNone selects no Win32 screencap method.
	ScreencapNone ScreencapMethod = 0
	// ScreencapGDI selects the GDI method: fast, medium compatibility.
	ScreencapGDI ScreencapMethod = 1
	// ScreencapFramePool selects the FramePool method: very fast, medium
	// compatibility. Requires Windows 10 1903+ and supports background
	// capture and pseudo-minimize.
	ScreencapFramePool ScreencapMethod = 1 << 1
	// ScreencapDXGIDesktopDup selects the DXGI_DesktopDup method: very
	// fast, low compatibility. Desktop duplication (full screen).
	ScreencapDXGIDesktopDup ScreencapMethod = 1 << 2
	// ScreencapDXGIDesktopDupWindow selects the DXGI_DesktopDup_Window
	// method: very fast, low compatibility. Desktop duplication then crop.
	ScreencapDXGIDesktopDupWindow ScreencapMethod = 1 << 3
	// ScreencapPrintWindow selects the PrintWindow method: medium speed,
	// medium compatibility. Supports background capture and
	// pseudo-minimize.
	ScreencapPrintWindow ScreencapMethod = 1 << 4
	// ScreencapScreenDC selects the ScreenDC method: fast, high
	// compatibility.
	ScreencapScreenDC ScreencapMethod = 1 << 5

	// ScreencapAll selects every Win32 screencap method.
	ScreencapAll ScreencapMethod = ^ScreencapNone
	// ScreencapForeground selects the predefined foreground combination:
	// DXGI_DesktopDup_Window | ScreenDC.
	ScreencapForeground ScreencapMethod = ScreencapDXGIDesktopDupWindow | ScreencapScreenDC
	// ScreencapBackground selects the predefined background combination:
	// FramePool | PrintWindow.
	ScreencapBackground ScreencapMethod = ScreencapFramePool | ScreencapPrintWindow

	// InputNone selects no Win32 input method.
	InputNone InputMethod = 0
	// InputSeize selects the Seize input method.
	InputSeize InputMethod = 1
	// InputSendMessage selects the SendMessage input method: medium
	// compatibility, supports background input.
	InputSendMessage InputMethod = 1 << 1
	// InputPostMessage selects the PostMessage input method: medium
	// compatibility, supports background input.
	InputPostMessage InputMethod = 1 << 2
	// InputLegacyEvent selects the LegacyEvent input method: low
	// compatibility, seizes the mouse, no background support.
	InputLegacyEvent InputMethod = 1 << 3
	// InputPostThreadMessage selects the PostThreadMessage input method.
	//
	// Deprecated: upstream deprecated this method and no longer implements
	// it. A controller created with it still connects, but every input
	// action fails.
	InputPostThreadMessage InputMethod = 1 << 4
	// InputSendMessageWithCursorPos selects the SendMessageWithCursorPos
	// input method: medium compatibility, supports background input.
	// Briefly moves the cursor to the target position, then restores it.
	InputSendMessageWithCursorPos InputMethod = 1 << 5
	// InputPostMessageWithCursorPos selects the PostMessageWithCursorPos
	// input method: medium compatibility, supports background input.
	// Briefly moves the cursor to the target position, then restores it.
	InputPostMessageWithCursorPos InputMethod = 1 << 6
	// InputSendMessageWithWindowPos selects the SendMessageWithWindowPos
	// input method: medium compatibility, supports background input.
	// Briefly moves the window to align the target with the cursor, then
	// restores it; the cursor is not moved.
	InputSendMessageWithWindowPos InputMethod = 1 << 7
	// InputPostMessageWithWindowPos selects the PostMessageWithWindowPos
	// input method: medium compatibility, supports background input.
	// Briefly moves the window to align the target with the cursor, then
	// restores it; the cursor is not moved.
	InputPostMessageWithWindowPos InputMethod = 1 << 8
	// InputInterception selects the Interception input method:
	// driver-level input injection via the Interception driver.
	InputInterception InputMethod = 1 << 9
	// InputAnchoredTouch injects touch points without moving the cursor or target window.
	// It supports clicks and swipes, but not scrolling or keyboard input.
	InputAnchoredTouch InputMethod = 1 << 10
)

const (
	screencapNoneStr                 = ""
	screencapGDIStr                  = "GDI"
	screencapFramePoolStr            = "FramePool"
	screencapDXGIDesktopDupStr       = "DXGI_DesktopDup"
	screencapDXGIDesktopDupWindowStr = "DXGI_DesktopDup_Window"
	screencapPrintWindowStr          = "PrintWindow"
	screencapScreenDCStr             = "ScreenDC"

	screencapAllStr        = "All"
	screencapForegroundStr = "Foreground"
	screencapBackgroundStr = "Background"

	// Legacy Go-only spellings without underscores, emitted by earlier v4
	// betas. ParseScreencapMethod still accepts them.
	screencapDXGIDesktopDupLegacyStr       = "DXGIDesktopDup"
	screencapDXGIDesktopDupWindowLegacyStr = "DXGIDesktopDupWindow"

	inputNoneStr                     = ""
	inputSeizeStr                    = "Seize"
	inputSendMessageStr              = "SendMessage"
	inputPostMessageStr              = "PostMessage"
	inputLegacyEventStr              = "LegacyEvent"
	inputPostThreadMessageStr        = "PostThreadMessage"
	inputSendMessageWithCursorPosStr = "SendMessageWithCursorPos"
	inputPostMessageWithCursorPosStr = "PostMessageWithCursorPos"
	inputSendMessageWithWindowPosStr = "SendMessageWithWindowPos"
	inputPostMessageWithWindowPosStr = "PostMessageWithWindowPos"
	inputInterceptionStr             = "Interception"
	inputAnchoredTouchStr            = "AnchoredTouch"
)

// String returns the screencap method's canonical upstream spelling (for
// example "DXGI_DesktopDup_Window"), or its decimal value for unnamed
// combinations.
func (m ScreencapMethod) String() string {
	switch m {
	case ScreencapNone:
		return screencapNoneStr
	case ScreencapGDI:
		return screencapGDIStr
	case ScreencapFramePool:
		return screencapFramePoolStr
	case ScreencapDXGIDesktopDup:
		return screencapDXGIDesktopDupStr
	case ScreencapDXGIDesktopDupWindow:
		return screencapDXGIDesktopDupWindowStr
	case ScreencapPrintWindow:
		return screencapPrintWindowStr
	case ScreencapScreenDC:
		return screencapScreenDCStr
	case ScreencapAll:
		return screencapAllStr
	case ScreencapForeground:
		return screencapForegroundStr
	case ScreencapBackground:
		return screencapBackgroundStr
	}
	return strconv.FormatUint(uint64(m), 10)
}

// String returns the input method's canonical name, or its decimal value
// for unnamed combinations.
func (m InputMethod) String() string {
	switch m {
	case InputNone:
		return inputNoneStr
	case InputSeize:
		return inputSeizeStr
	case InputSendMessage:
		return inputSendMessageStr
	case InputPostMessage:
		return inputPostMessageStr
	case InputLegacyEvent:
		return inputLegacyEventStr
	case InputPostThreadMessage:
		return inputPostThreadMessageStr
	case InputSendMessageWithCursorPos:
		return inputSendMessageWithCursorPosStr
	case InputPostMessageWithCursorPos:
		return inputPostMessageWithCursorPosStr
	case InputSendMessageWithWindowPos:
		return inputSendMessageWithWindowPosStr
	case InputPostMessageWithWindowPos:
		return inputPostMessageWithWindowPosStr
	case InputInterception:
		return inputInterceptionStr
	case InputAnchoredTouch:
		return inputAnchoredTouchStr
	}
	return strconv.FormatUint(uint64(m), 10)
}

// ParseScreencapMethod parses a screencap method name, matched
// case-insensitively with surrounding whitespace ignored. It accepts both
// the canonical upstream spellings and this binding's legacy
// underscore-free DXGI spellings ("DXGIDesktopDup",
// "DXGIDesktopDupWindow"). Unrecognized names fall back to a decimal
// numeric value; input matching neither returns an error.
func ParseScreencapMethod(s string) (ScreencapMethod, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.EqualFold(screencapNoneStr, s):
		return ScreencapNone, nil
	case strings.EqualFold(screencapGDIStr, s):
		return ScreencapGDI, nil
	case strings.EqualFold(screencapFramePoolStr, s):
		return ScreencapFramePool, nil
	case strings.EqualFold(screencapDXGIDesktopDupStr, s), strings.EqualFold(screencapDXGIDesktopDupLegacyStr, s):
		return ScreencapDXGIDesktopDup, nil
	case strings.EqualFold(screencapDXGIDesktopDupWindowStr, s), strings.EqualFold(screencapDXGIDesktopDupWindowLegacyStr, s):
		return ScreencapDXGIDesktopDupWindow, nil
	case strings.EqualFold(screencapPrintWindowStr, s):
		return ScreencapPrintWindow, nil
	case strings.EqualFold(screencapScreenDCStr, s):
		return ScreencapScreenDC, nil
	case strings.EqualFold(screencapAllStr, s):
		return ScreencapAll, nil
	case strings.EqualFold(screencapForegroundStr, s):
		return ScreencapForeground, nil
	case strings.EqualFold(screencapBackgroundStr, s):
		return ScreencapBackground, nil
	default:
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return ScreencapNone, fmt.Errorf("invalid screencap method: %s", s)
		}
		return ScreencapMethod(i), nil
	}
}

// ParseInputMethod parses an input method name, matched case-insensitively
// with surrounding whitespace ignored. Unrecognized names fall back to a
// decimal numeric value; input matching neither returns an error.
func ParseInputMethod(s string) (InputMethod, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.EqualFold(inputNoneStr, s):
		return InputNone, nil
	case strings.EqualFold(inputSeizeStr, s):
		return InputSeize, nil
	case strings.EqualFold(inputSendMessageStr, s):
		return InputSendMessage, nil
	case strings.EqualFold(inputPostMessageStr, s):
		return InputPostMessage, nil
	case strings.EqualFold(inputLegacyEventStr, s):
		return InputLegacyEvent, nil
	case strings.EqualFold(inputPostThreadMessageStr, s):
		return InputPostThreadMessage, nil
	case strings.EqualFold(inputSendMessageWithCursorPosStr, s):
		return InputSendMessageWithCursorPos, nil
	case strings.EqualFold(inputPostMessageWithCursorPosStr, s):
		return InputPostMessageWithCursorPos, nil
	case strings.EqualFold(inputSendMessageWithWindowPosStr, s):
		return InputSendMessageWithWindowPos, nil
	case strings.EqualFold(inputPostMessageWithWindowPosStr, s):
		return InputPostMessageWithWindowPos, nil
	case strings.EqualFold(inputInterceptionStr, s):
		return InputInterception, nil
	case strings.EqualFold(inputAnchoredTouchStr, s):
		return InputAnchoredTouch, nil
	default:
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return InputNone, fmt.Errorf("invalid input method: %s", s)
		}
		return InputMethod(i), nil
	}
}
