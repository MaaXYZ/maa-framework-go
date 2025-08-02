package maa

import (
	"unsafe"

	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
)

func setGlobalOption(key maa.MaaGlobalOption, value unsafe.Pointer, valSize uintptr) bool {
	return maa.MaaSetGlobalOption(key, value, uint64(valSize))
}

// SetLogDir sets the log directory.
func SetLogDir(path string) bool {
	if path == "" {
		return false
	}
	return setGlobalOption(maa.MaaGlobalOption_LogDir, unsafe.Pointer(&[]byte(path)[0]), uintptr(len(path)))
}

// SetSaveDraw sets whether to save draw.
func SetSaveDraw(enabled bool) bool {
	return setGlobalOption(maa.MaaGlobalOption_SaveDraw, unsafe.Pointer(&enabled), unsafe.Sizeof(enabled))
}

// SetRecording sets whether to dump all screenshots and actions.
func SetRecording(enabled bool) bool {
	return setGlobalOption(maa.MaaGlobalOption_Recording, unsafe.Pointer(&enabled), unsafe.Sizeof(enabled))
}

type LoggingLevel int32

// LoggingLevel
const (
	LoggingLevelOff LoggingLevel = iota
	LoggingLevelFatal
	LoggingLevelError
	LoggingLevelWarn
	LoggingLevelInfo
	LoggingLevelDebug
	LoggingLevelTrace
	LoggingLevelAll
)

// SetStdoutLevel sets the level of log output to stdout.
func SetStdoutLevel(level LoggingLevel) bool {
	return setGlobalOption(maa.MaaGlobalOption_StdoutLevel, unsafe.Pointer(&level), unsafe.Sizeof(level))
}

// SetDebugMode sets whether to enable debug mode.
func SetDebugMode(enabled bool) bool {
	return setGlobalOption(maa.MaaGlobalOption_DebugMode, unsafe.Pointer(&enabled), unsafe.Sizeof(enabled))
}
