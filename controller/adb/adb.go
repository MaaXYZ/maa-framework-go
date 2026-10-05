// Package adb defines the screencap and input method flags for ADB
// controllers, with name parsing helpers.
package adb

import (
	"fmt"
	"strconv"
	"strings"
)

// ScreencapMethod defines the ADB screencap method flags.
//
// Use bitwise OR to set the method you need,
// MaaFramework will test their speed and use the fastest one.
type ScreencapMethod uint64

// InputMethod defines the ADB input method flags.
//
// Use bitwise OR to set the method you need,
// MaaFramework will select the available ones according to priority.
// The priority is: EmulatorExtras > Maatouch > MinitouchAndAdbKey > AdbShell
type InputMethod uint64

const (
	// ScreencapNone selects no ADB screencap method.
	ScreencapNone ScreencapMethod = 0
	// ScreencapEncodeToFileAndPull selects the EncodeToFileAndPull method:
	// slow, high compatibility, lossless encoding.
	ScreencapEncodeToFileAndPull ScreencapMethod = 1
	// ScreencapEncode selects the Encode method:
	// slow, high compatibility, lossless encoding.
	ScreencapEncode ScreencapMethod = 1 << 1
	// ScreencapRawWithGzip selects the RawWithGzip method:
	// medium speed, high compatibility, lossless encoding.
	ScreencapRawWithGzip ScreencapMethod = 1 << 2
	// ScreencapRawByNetcat selects the RawByNetcat method:
	// fast, low compatibility, lossless encoding.
	ScreencapRawByNetcat ScreencapMethod = 1 << 3
	// ScreencapMinicapDirect selects the MinicapDirect method: fast, low
	// compatibility. It uses lossy JPEG encoding, which may significantly
	// reduce template matching accuracy; not recommended.
	ScreencapMinicapDirect ScreencapMethod = 1 << 4
	// ScreencapMinicapStream selects the MinicapStream method: very fast,
	// low compatibility. It uses lossy JPEG encoding, which may
	// significantly reduce template matching accuracy; not recommended.
	ScreencapMinicapStream ScreencapMethod = 1 << 5
	// ScreencapEmulatorExtras selects the EmulatorExtras method: very fast,
	// low compatibility, lossless encoding. Emulators only: MuMu 12,
	// LDPlayer 9.
	ScreencapEmulatorExtras ScreencapMethod = 1 << 6
	// ScreencapAll selects every ADB screencap method.
	ScreencapAll = ^ScreencapNone
	// ScreencapDefault selects every ADB screencap method except
	// RawByNetcat, MinicapDirect, and MinicapStream.
	ScreencapDefault = ScreencapAll &
		(^ScreencapRawByNetcat) &
		(^ScreencapMinicapDirect) &
		(^ScreencapMinicapStream)

	// InputNone selects no ADB input method.
	InputNone InputMethod = 0
	// InputAdbShell selects the AdbShell method: slow, high compatibility.
	InputAdbShell InputMethod = 1
	// InputMinitouchAndAdbKey selects the MinitouchAndAdbKey method: fast,
	// medium compatibility. Key input still uses AdbShell.
	InputMinitouchAndAdbKey InputMethod = 1 << 1
	// InputMaatouch selects the Maatouch method: fast, medium compatibility.
	InputMaatouch InputMethod = 1 << 2
	// InputEmulatorExtras selects the EmulatorExtras method: fast, low
	// compatibility. Emulators only: MuMu 12.
	InputEmulatorExtras InputMethod = 1 << 3
	// InputAll selects every ADB input method.
	InputAll = ^InputNone
	// InputDefault selects every ADB input method except EmulatorExtras.
	InputDefault = InputAll & (^InputEmulatorExtras)
)

const (
	screencapNoneStr                = ""
	screencapEncodeToFileAndPullStr = "EncodeToFileAndPull"
	screencapEncodeStr              = "Encode"
	screencapRawWithGzipStr         = "RawWithGzip"
	screencapRawByNetcatStr         = "RawByNetcat"
	screencapMinicapDirectStr       = "MinicapDirect"
	screencapMinicapStreamStr       = "MinicapStream"
	screencapEmulatorExtrasStr      = "EmulatorExtras"
	screencapAllStr                 = "All"
	screencapDefaultStr             = "Default"

	inputNoneStr               = ""
	inputAdbShellStr           = "AdbShell"
	inputMinitouchAndAdbKeyStr = "MinitouchAndAdbKey"
	inputMaatouchStr           = "Maatouch"
	inputEmulatorExtrasStr     = "EmulatorExtras"
	inputAllStr                = "All"
	inputDefaultStr            = "Default"
)

// String returns the screencap method's canonical name, or its decimal
// value for unnamed combinations.
func (m ScreencapMethod) String() string {
	switch m {
	case ScreencapNone:
		return screencapNoneStr
	case ScreencapEncodeToFileAndPull:
		return screencapEncodeToFileAndPullStr
	case ScreencapEncode:
		return screencapEncodeStr
	case ScreencapRawWithGzip:
		return screencapRawWithGzipStr
	case ScreencapRawByNetcat:
		return screencapRawByNetcatStr
	case ScreencapMinicapDirect:
		return screencapMinicapDirectStr
	case ScreencapMinicapStream:
		return screencapMinicapStreamStr
	case ScreencapEmulatorExtras:
		return screencapEmulatorExtrasStr
	case ScreencapAll:
		return screencapAllStr
	case ScreencapDefault:
		return screencapDefaultStr
	}
	return strconv.FormatUint(uint64(m), 10)
}

// String returns the input method's canonical name, or its decimal value
// for unnamed combinations.
func (m InputMethod) String() string {
	switch m {
	case InputNone:
		return inputNoneStr
	case InputAdbShell:
		return inputAdbShellStr
	case InputMinitouchAndAdbKey:
		return inputMinitouchAndAdbKeyStr
	case InputMaatouch:
		return inputMaatouchStr
	case InputEmulatorExtras:
		return inputEmulatorExtrasStr
	case InputAll:
		return inputAllStr
	case InputDefault:
		return inputDefaultStr
	}
	return strconv.FormatUint(uint64(m), 10)
}

// ParseScreencapMethod parses a screencap method name, matched
// case-insensitively with surrounding whitespace ignored. Unrecognized
// names fall back to a decimal numeric value; input matching neither
// returns an error.
func ParseScreencapMethod(s string) (ScreencapMethod, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.EqualFold(screencapNoneStr, s):
		return ScreencapNone, nil
	case strings.EqualFold(screencapEncodeToFileAndPullStr, s):
		return ScreencapEncodeToFileAndPull, nil
	case strings.EqualFold(screencapEncodeStr, s):
		return ScreencapEncode, nil
	case strings.EqualFold(screencapRawWithGzipStr, s):
		return ScreencapRawWithGzip, nil
	case strings.EqualFold(screencapRawByNetcatStr, s):
		return ScreencapRawByNetcat, nil
	case strings.EqualFold(screencapMinicapDirectStr, s):
		return ScreencapMinicapDirect, nil
	case strings.EqualFold(screencapMinicapStreamStr, s):
		return ScreencapMinicapStream, nil
	case strings.EqualFold(screencapEmulatorExtrasStr, s):
		return ScreencapEmulatorExtras, nil
	case strings.EqualFold(screencapAllStr, s):
		return ScreencapAll, nil
	case strings.EqualFold(screencapDefaultStr, s):
		return ScreencapDefault, nil
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
	case strings.EqualFold(inputAdbShellStr, s):
		return InputAdbShell, nil
	case strings.EqualFold(inputMinitouchAndAdbKeyStr, s):
		return InputMinitouchAndAdbKey, nil
	case strings.EqualFold(inputMaatouchStr, s):
		return InputMaatouch, nil
	case strings.EqualFold(inputEmulatorExtrasStr, s):
		return InputEmulatorExtras, nil
	case strings.EqualFold(inputAllStr, s):
		return InputAll, nil
	case strings.EqualFold(inputDefaultStr, s):
		return InputDefault, nil
	default:
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return InputNone, fmt.Errorf("invalid input method: %s", s)
		}
		return InputMethod(i), nil
	}
}
