package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import "unsafe"

type GlobalOption int32

// GlobalOption
const (
	GlobalOptionInvalid GlobalOption = iota

	// GlobalOptionLogDir Log dir
	//
	// value: string, eg: "C:\\Users\\Administrator\\Desktop\\log"; val_size: string length
	GlobalOptionLogDir

	// GlobalOptionSaveDraw Whether to save draw
	//
	// value: bool, eg: true; val_size: sizeof(bool)
	GlobalOptionSaveDraw

	// GlobalOptionRecording Dump all screenshots and actions
	//
	// Recording will evaluate to true if any of this or MaaCtrlOptionEnum::MaaCtrlOption_Recording
	// is true. value: bool, eg: true; val_size: sizeof(bool)
	GlobalOptionRecording

	// GlobalOptionStdoutLevel The level of log output to stdout
	//
	// value: MaaLoggingLevel, val_size: sizeof(MaaLoggingLevel)
	// default value is MaaLoggingLevel_Error
	GlobalOptionStdoutLevel

	// GlobalOptionShowHitDraw Whether to show hit draw
	//
	// value: bool, eg: true; val_size: sizeof(bool)
	GlobalOptionShowHitDraw

	// GlobalOptionDebugMode Whether to debug
	//
	// value: bool, eg: true; val_size: sizeof(bool)
	GlobalOptionDebugMode
)

// SetLogDir sets the log directory.
func SetLogDir(path string) bool {
	if path == "" {
		return false
	}
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	return C.MaaSetGlobalOption(C.int32_t(GlobalOptionLogDir), C.MaaOptionValue(cPath), C.uint64_t(len(path))) != 0
}

// SetSaveDraw sets whether to save draw.
func SetSaveDraw(enabled bool) bool {
	var cEnabled uint8
	if enabled {
		cEnabled = 1
	}
	return C.MaaSetGlobalOption(C.int32_t(GlobalOptionSaveDraw), C.MaaOptionValue(unsafe.Pointer(&cEnabled)), C.uint64_t(unsafe.Sizeof(cEnabled))) != 0
}

// SetRecording sets whether to dump all screenshots and actions.
func SetRecording(enabled bool) bool {
	var cEnabled uint8
	if enabled {
		cEnabled = 1
	}
	return C.MaaSetGlobalOption(C.int32_t(GlobalOptionRecording), C.MaaOptionValue(unsafe.Pointer(&cEnabled)), C.uint64_t(unsafe.Sizeof(cEnabled))) != 0
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
	return C.MaaSetGlobalOption(C.int32_t(GlobalOptionStdoutLevel), C.MaaOptionValue(unsafe.Pointer(&level)), C.uint64_t(unsafe.Sizeof(level))) != 0
}

// SetShowHitDraw sets whether to show hit draw.
func SetShowHitDraw(enabled bool) bool {
	var cEnabled uint8
	if enabled {
		cEnabled = 1
	}
	return C.MaaSetGlobalOption(C.int32_t(GlobalOptionShowHitDraw), C.MaaOptionValue(unsafe.Pointer(&cEnabled)), C.uint64_t(unsafe.Sizeof(cEnabled))) != 0
}

// SetDebugMode sets whether to enable debug mode.
func SetDebugMode(enabled bool) bool {
	var cEnabled uint8
	if enabled {
		cEnabled = 1
	}
	return C.MaaSetGlobalOption(C.int32_t(GlobalOptionDebugMode), C.MaaOptionValue(unsafe.Pointer(&cEnabled)), C.uint64_t(unsafe.Sizeof(cEnabled))) != 0
}
