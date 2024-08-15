package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import (
	"errors"
	"github.com/MaaXYZ/maa-framework-go/buffer"
	"image"
	"unsafe"
)

// Version returns the version of the maa framework.
func Version() string {
	return C.GoString(C.MaaVersion())
}

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

	// GlobalOptionDebugMessage Whether to callback debug message
	//
	// value: bool, eg: true; val_size: sizeof(bool)
	GlobalOptionDebugMessage
)

// SetLogDir sets the log directory.
func SetLogDir(path string) bool {
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

// SetDebugMessage sets whether to callback debug message.
func SetDebugMessage(enabled bool) bool {
	var cEnabled uint8
	if enabled {
		cEnabled = 1
	}
	return C.MaaSetGlobalOption(C.int32_t(GlobalOptionDebugMessage), C.MaaOptionValue(unsafe.Pointer(&cEnabled)), C.uint64_t(unsafe.Sizeof(cEnabled))) != 0
}

type RecognitionDetail struct {
	ID         int64
	Name       string
	Hit        bool
	DetailJson string
	Raw        image.Image
	Draws      []image.Image
}

// QueryRecognitionDetail queries recognition detail.
func QueryRecognitionDetail(recId int64) (RecognitionDetail, error) {
	name := buffer.NewStringBuffer()
	var hit uint8
	hitBox := buffer.NewRectBuffer()
	detailJson := buffer.NewStringBuffer()
	raw := buffer.NewImageBuffer()
	draws := buffer.NewImageListBuffer()
	defer func() {
		name.Destroy()
		detailJson.Destroy()
	}()
	got := C.MaaQueryRecognitionDetail(
		C.int64_t(recId),
		C.MaaStringBufferHandle(name.Handle()),
		(*C.uint8_t)(unsafe.Pointer(&hit)),
		C.MaaRectHandle(hitBox.Handle()),
		C.MaaStringBufferHandle(detailJson.Handle()),
		C.MaaImageBufferHandle(raw.Handle()),
		C.MaaImageListBufferHandle(draws.Handle()),
	) != 0

	if !got {
		return RecognitionDetail{}, errors.New("failed to query recognition detail")
	}

	rawImg, err := raw.GetByRawData()
	if err != nil {
		return RecognitionDetail{}, err
	}

	DrawImages, err := draws.GetAll()
	if err != nil {
		return RecognitionDetail{}, err
	}

	return RecognitionDetail{
		ID:         recId,
		Name:       name.Get(),
		Hit:        hit != 0,
		DetailJson: detailJson.Get(),
		Raw:        rawImg,
		Draws:      DrawImages,
	}, nil
}

type NodeDetail struct {
	ID           int64
	Name         string
	Recognition  RecognitionDetail
	RunCompleted bool
}

// QueryNodeDetail queries running detail.
func QueryNodeDetail(nodeId int64) (NodeDetail, bool) {
	name := buffer.NewStringBuffer()
	defer name.Destroy()
	var recId int64
	var runCompleted uint8
	got := C.MaaQueryNodeDetail(
		C.int64_t(nodeId),
		C.MaaStringBufferHandle(name.Handle()),
		(*C.int64_t)(unsafe.Pointer(&recId)),
		(*C.uint8_t)(unsafe.Pointer(&runCompleted)),
	) != 0

	recognitionDetail, err := QueryRecognitionDetail(recId)
	if err != nil {

	}

	return NodeDetail{
		ID:           nodeId,
		Name:         name.Get(),
		Recognition:  recognitionDetail,
		RunCompleted: runCompleted != 0,
	}, got
}

type TaskDetail struct {
	ID          int64
	Entry       string
	NodeDetails []NodeDetail
}

// QueryTaskDetail queries task detail.
func QueryTaskDetail(taskId int64) (TaskDetail, bool) {
	entry := buffer.NewStringBuffer()
	defer entry.Destroy()
	var size uint64
	got := C.MaaQueryTaskDetail(C.int64_t(taskId), nil, nil, (*C.uint64_t)(unsafe.Pointer(&size))) != 0
	if !got {
		return TaskDetail{}, got
	}
	nodeIdList := make([]int64, size)
	got = C.MaaQueryTaskDetail(
		C.int64_t(taskId),
		C.MaaStringBufferHandle(entry.Handle()),
		(*C.int64_t)(unsafe.Pointer(&nodeIdList[0])),
		(*C.uint64_t)(unsafe.Pointer(&size)),
	) != 0

	nodeDetails := make([]NodeDetail, size)
	for i, nodeId := range nodeIdList {
		nodeDetail, ok := QueryNodeDetail(nodeId)
		if !ok {

		}
		nodeDetails[i] = nodeDetail
	}

	return TaskDetail{
		ID:          taskId,
		Entry:       entry.Get(),
		NodeDetails: nodeDetails,
	}, got
}
