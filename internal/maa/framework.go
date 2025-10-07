package maa

import (
	"fmt"
	"path/filepath"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

var maaFramework uintptr

var (
	MaaVersion func() string
)

type MaaNotificationCallback func(message, detailsJson *byte, notifyTransArg uintptr) uintptr

type MaaTaskerOption int32

var (
	MaaTaskerCreate               func(notify MaaNotificationCallback, notifyTransArg uintptr) uintptr
	MaaTaskerDestroy              func(tasker uintptr)
	MaaTaskerAddSink              func(tasker uintptr, notify MaaNotificationCallback, notifyTransArg uintptr) bool
	MaaTaskerRemoveSink           func(tasker uintptr, notify MaaNotificationCallback) bool
	MaaTaskerClearSinks           func(tasker uintptr) bool
	MaaTaskerSetOption            func(tasker uintptr, key MaaTaskerOption, value unsafe.Pointer, valSize uint64) bool
	MaaTaskerBindResource         func(tasker uintptr, res uintptr) bool
	MaaTaskerBindController       func(tasker uintptr, ctrl uintptr) bool
	MaaTaskerInited               func(tasker uintptr) bool
	MaaTaskerPostTask             func(tasker uintptr, entry, pipelineOverride string) int64
	MaaTaskerStatus               func(tasker uintptr, id int64) int32
	MaaTaskerWait                 func(tasker uintptr, id int64) int32
	MaaTaskerRunning              func(tasker uintptr) bool
	MaaTaskerPostStop             func(tasker uintptr) int64
	MaaTaskerStopping             func(tasker uintptr) bool
	MaaTaskerGetResource          func(tasker uintptr) uintptr
	MaaTaskerGetController        func(tasker uintptr) uintptr
	MaaTaskerClearCache           func(tasker uintptr) bool
	MaaTaskerGetRecognitionDetail func(tasker uintptr, recoId int64, nodeName uintptr, algorithm uintptr, hit *bool, box uintptr, detailJson uintptr, raw uintptr, draws uintptr) bool
	MaaTaskerGetNodeDetail        func(tasker uintptr, nodeId int64, nodeName uintptr, recoId *int64, completed *bool) bool
	MaaTaskerGetTaskDetail        func(tasker uintptr, taskId int64, entry uintptr, nodeIdList uintptr, nodeIdListSize *uint64, status *int32) bool
	MaaTaskerGetLatestNode        func(tasker uintptr, taskName string, latestId *int64) bool
)

type MaaCustomRecognitionCallback func(context uintptr, taskId int64, currentTaskName, customRecognitionName, customRecognitionParam *byte, image, roi uintptr, transArg uintptr, outBox, outDetail uintptr) uint64

type MaaCustomActionCallback func(context uintptr, taskId int64, currentTaskName, customActionName, customActionParam *byte, recoId int64, box uintptr, transArg uintptr) uint64

type MaaInferenceDevice int32

const (
	MaaInferenceDevice_CPU  MaaInferenceDevice = -2
	MaaInferenceDevice_Auto MaaInferenceDevice = -1
	MaaInferenceDevice_0    MaaInferenceDevice = 0
	MaaInferenceDevice_1    MaaInferenceDevice = 1
	// and more gpu id or flag...
)

type MaaInferenceExecutionProvider int32

const (

	// I don't recommend setting up MaaResOption_InferenceDevice in this case,
	// because you don't know which EP will be used on different user devices.
	MaaInferenceExecutionProvider_Auto = 0

	// MaaResOption_InferenceDevice will not work.
	MaaInferenceExecutionProvider_CPU = 1

	// MaaResOption_InferenceDevice will be used to set adapter id,
	// It's from Win32 API `EnumAdapters1`.
	MaaInferenceExecutionProvider_DirectML = 2

	// MaaResOption_InferenceDevice will be used to set coreml_flag,
	// Reference to
	// https://github.com/microsoft/onnxruntime/blob/main/include/onnxruntime/core/providers/coreml/coreml_provider_factory.h
	// But you need to pay attention to the onnxruntime version we use, the latest flag may not be supported.
	MaaInferenceExecutionProvider_CoreML = 3

	// MaaResOption_InferenceDevice will be used to set NVIDIA GPU ID
	// TODO!
	MaaInferenceExecutionProvider_CUDA = 4
)

type MaaResOption int32

const (
	MaaResOption_Invalid MaaResOption = 0

	/// Use the specified inference device.
	/// Please set this option before loading the model.
	///
	/// value: MaaInferenceDevice, eg: 0; val_size: sizeof(MaaInferenceDevice)
	/// default value is MaaInferenceDevice_Auto
	MaaResOption_InferenceDevice MaaResOption = 1

	/// Use the specified inference execution provider
	/// Please set this option before loading the model.
	///
	/// value: MaaInferenceExecutionProvider, eg: 0; val_size: sizeof(MaaInferenceExecutionProvider)
	/// default value is MaaInferenceExecutionProvider_Auto
	MaaResOption_InferenceExecutionProvider MaaResOption = 2
)

var (
	MaaResourceCreate                      func(notify MaaNotificationCallback, notifyTransArg uintptr) uintptr
	MaaResourceDestroy                     func(res uintptr)
	MaaResourceAddSink                     func(res uintptr, notify MaaNotificationCallback, notifyTransArg uintptr) bool
	MaaResourceRemoveSink                  func(res uintptr, notify MaaNotificationCallback) bool
	MaaResourceClearSinks                  func(res uintptr) bool
	MaaResourceRegisterCustomRecognition   func(res uintptr, name string, recognition MaaCustomRecognitionCallback, transArg uintptr) bool
	MaaResourceUnregisterCustomRecognition func(res uintptr, name string) bool
	MaaResourceClearCustomRecognition      func(res uintptr) bool
	MaaResourceRegisterCustomAction        func(res uintptr, name string, action MaaCustomActionCallback, transArg uintptr) bool
	MaaResourceUnregisterCustomAction      func(res uintptr, name string) bool
	MaaResourceClearCustomAction           func(res uintptr) bool
	MaaResourcePostBundle                  func(res uintptr, path string) int64
	MaaResourceOverridePipeline            func(res uintptr, pipelineOverride string) bool
	MaaResourceOverrideNext                func(res uintptr, nodeName string, nextList uintptr) bool
	MaaResourceGetNodeData                 func(res uintptr, nodeName string, buffer uintptr) bool
	MaaResourceClear                       func(res uintptr) bool
	MaaResourceStatus                      func(res uintptr, id int64) int32
	MaaResourceWait                        func(res uintptr, id int64) int32
	MaaResourceLoaded                      func(res uintptr) bool
	MaaResourceSetOption                   func(res uintptr, key MaaResOption, value unsafe.Pointer, valSize uint64) bool
	MaaResourceGetHash                     func(res uintptr, buffer uintptr) bool
	MaaResourceGetNodeList                 func(res uintptr, buffer uintptr) bool
)

// MaaAdbScreencapMethod
//
// Use bitwise OR to set the method you need,
// MaaFramework will test their speed and use the fastest one.
type MaaAdbScreencapMethod uint64

const (
	MaaAdbScreencapMethod_None                MaaAdbScreencapMethod = 0
	MaaAdbScreencapMethod_EncodeToFileAndPull MaaAdbScreencapMethod = 1
	MaaAdbScreencapMethod_Encode              MaaAdbScreencapMethod = 1 << 1
	MaaAdbScreencapMethod_RawWithGzip         MaaAdbScreencapMethod = 1 << 2
	MaaAdbScreencapMethod_RawByNetcat         MaaAdbScreencapMethod = 1 << 3
	MaaAdbScreencapMethod_MinicapDirect       MaaAdbScreencapMethod = 1 << 4
	MaaAdbScreencapMethod_MinicapStream       MaaAdbScreencapMethod = 1 << 5
	MaaAdbScreencapMethod_EmulatorExtras      MaaAdbScreencapMethod = 1 << 6

	MaaAdbScreencapMethod_All     = ^MaaAdbScreencapMethod_None
	MaaAdbScreencapMethod_Default = MaaAdbScreencapMethod_All & (^MaaAdbScreencapMethod_MinicapDirect) & (^MaaAdbScreencapMethod_MinicapStream)
)

// MaaAdbInputMethod
//
// Use bitwise OR to set the method you need,
// MaaFramework will select the available ones according to priority.
// The priority is: EmulatorExtras > Maatouch > MinitouchAndAdbKey > AdbShell
type MaaAdbInputMethod uint64

const (
	MaaAdbInputMethod_None               MaaAdbInputMethod = 0
	MaaAdbInputMethod_AdbShell           MaaAdbInputMethod = 1
	MaaAdbInputMethod_MinitouchAndAdbKey MaaAdbInputMethod = 1 << 1
	MaaAdbInputMethod_Maatouch           MaaAdbInputMethod = 1 << 2
	MaaAdbInputMethod_EmulatorExtras     MaaAdbInputMethod = 1 << 3

	MaaAdbInputMethod_All     = ^MaaAdbInputMethod_None
	MaaAdbInputMethod_Default = MaaAdbInputMethod_All & (^MaaAdbInputMethod_EmulatorExtras)
)

// MaaWin32ScreencapMethod
//
// No bitwise OR, just set it.
type MaaWin32ScreencapMethod uint64

const (
	MaaWin32ScreencapMethod_None           MaaWin32ScreencapMethod = 0
	MaaWin32ScreencapMethod_GDI            MaaWin32ScreencapMethod = 1
	MaaWin32ScreencapMethod_FramePool      MaaWin32ScreencapMethod = 1 << 1
	MaaWin32ScreencapMethod_DXGIDesktopDup MaaWin32ScreencapMethod = 1 << 2
)

// MaaWin32InputMethod
//
// No bitwise OR, just set it.
type MaaWin32InputMethod uint64

const (
	MaaWin32InputMethod_None        MaaWin32ScreencapMethod = 0
	MaaWin32InputMethod_Seize       MaaWin32ScreencapMethod = 1
	MaaWin32InputMethod_SendMessage MaaWin32ScreencapMethod = 1 << 1
)

// DbgControllerType
//
// No bitwise OR, just set it.
type MaaDbgControllerType uint64

const (
	MaaDbgControllerType_None            MaaDbgControllerType = 0
	MaaDbgControllerType_CarouselImage   MaaDbgControllerType = 1
	MaaDbgControllerType_ReplayRecording MaaDbgControllerType = 1 << 1
)

type MaaCtrlOption int32

const (
	MaaCtrlOption_Invalid MaaCtrlOption = 0

	// MaaCtrlOptionScreenshotTargetLongSide specifies that only the long side can be set, and the short side
	// is automatically scaled according to the aspect ratio.
	MaaCtrlOption_ScreenshotTargetLongSide MaaCtrlOption = 1

	// MaaCtrlOptionScreenshotTargetShortSide specifies that only the short side can be set, and the long side
	// is automatically scaled according to the aspect ratio.
	MaaCtrlOption_ScreenshotTargetShortSide MaaCtrlOption = 2

	// MaaCtrlOptionScreenshotUseRawSize specifies that the screenshot uses the raw size without scaling.
	// Note that this option may cause incorrect coordinates on user devices with different resolutions if scaling is not performed.
	MaaCtrlOption_ScreenshotUseRawSize MaaCtrlOption = 3
)

var (
	MaaAdbControllerCreate      func(adbPath, address string, screencapMethods MaaAdbScreencapMethod, inputMethods MaaAdbInputMethod, config, agentPath string, notify MaaNotificationCallback, notifyTransArg uintptr) uintptr
	MaaWin32ControllerCreate    func(hWnd unsafe.Pointer, screencapMethods MaaWin32ScreencapMethod, inputMethods MaaWin32InputMethod, notify MaaNotificationCallback, notifyTransArg uintptr) uintptr
	MaaCustomControllerCreate   func(controller uintptr, controllerArg uintptr, notify MaaNotificationCallback, notifyTransArg uintptr) uintptr
	MaaDbgControllerCreate      func(readPath, writePath string, dbgCtrlType MaaDbgControllerType, config string, notify MaaNotificationCallback, notifyTransArg uintptr) uintptr
	MaaControllerDestroy        func(ctrl uintptr)
	MaaControllerAddSink        func(ctrl uintptr, notify MaaNotificationCallback, notifyTransArg uintptr) bool
	MaaControllerRemoveSink     func(ctrl uintptr, notify MaaNotificationCallback) bool
	MaaControllerClearSinks     func(ctrl uintptr) bool
	MaaControllerSetOption      func(ctrl uintptr, key MaaCtrlOption, value unsafe.Pointer, valSize uint64) bool
	MaaControllerPostConnection func(ctrl uintptr) int64
	MaaControllerPostClick      func(ctrl uintptr, x, y int32) int64
	MaaControllerPostSwipe      func(ctrl uintptr, x1, y1, x2, y2, duration int32) int64
	MaaControllerPostClickKey   func(ctrl uintptr, keycode int32) int64
	MaaControllerPostInputText  func(ctrl uintptr, text string) int64
	MaaControllerPostStartApp   func(ctrl uintptr, intent string) int64
	MaaControllerPostStopApp    func(ctrl uintptr, intent string) int64
	// for adb controller, contact means finger id (0 for first finger, 1 for second finger, etc)
	// for win32 controller, contact means mouse button id (0 for left, 1 for right, 2 for middle)
	MaaControllerPostTouchDown func(ctrl uintptr, contact, x, y, pressure int32) int64
	MaaControllerPostTouchMove func(ctrl uintptr, contact, x, y, pressure int32) int64
	// for adb controller, contact means finger id (0 for first finger, 1 for second finger, etc)
	// for win32 controller, contact means mouse button id (0 for left, 1 for right, 2 for middle)
	MaaControllerPostTouchUp   func(ctrl uintptr, contact int32) int64
	MaaControllerPostKeyDown   func(ctrl uintptr, keycode int32) int64
	MaaControllerPostKeyUp     func(ctrl uintptr, keycode int32) int64
	MaaControllerPostScreencap func(ctrl uintptr) int64
	MaaControllerStatus        func(ctrl uintptr, id int64) int32
	MaaControllerWait          func(ctrl uintptr, id int64) int32
	MaaControllerConnected     func(ctrl uintptr) bool
	MaaControllerCachedImage   func(ctrl uintptr, buffer uintptr) bool
	MaaControllerGetUuid       func(ctrl uintptr, buffer uintptr) bool
)

type MaaCustomControllerCallbacks struct {
	Connect     uintptr
	RequestUUID uintptr
	StartApp    uintptr
	StopApp     uintptr
	Screencap   uintptr
	Click       uintptr
	Swipe       uintptr
	TouchDown   uintptr
	TouchMove   uintptr
	TouchUp     uintptr
	ClickKey    uintptr
	InputText   uintptr
	KeyDown     uintptr
	KeyUp       uintptr
}

type (
	ConnectCallback     func(transArg uintptr) bool
	RequestUUIDCallback func(transArg uintptr, buffer uintptr) bool
	StartAppCallback    func(intent string, transArg uintptr) bool
	StopAppCallback     func(intent string, transArg uintptr) bool
	ScreencapCallback   func(transArg uintptr, buffer uintptr) bool
	ClickCallback       func(x, y int32, transArg uintptr) bool
	SwipeCallback       func(x1, y1, x2, y2, duration int32, transArg uintptr) bool
	TouchDownCallback   func(contact, x, y, pressure int32, transArg uintptr) bool
	TouchMoveCallback   func(contact, x, y, pressure int32, transArg uintptr) bool
	TouchUpCallback     func(contact int32, transArg uintptr) bool
	ClickKeyCallback    func(keycode int32, transArg uintptr) bool
	InputTextCallback   func(text string, transArg uintptr) bool
	KeyDownCallback     func(keycode int32, transArg uintptr) bool
	KeyUpCallback       func(keycode int32, transArg uintptr) bool
)

func MaaCustomControllerCallbacksCreate(
	connect ConnectCallback,
	requestUUID RequestUUIDCallback,
	startApp StartAppCallback,
	stopApp StopAppCallback,
	screencap ScreencapCallback,
	click ClickCallback,
	swipe SwipeCallback,
	touchDown TouchDownCallback,
	touchMove TouchMoveCallback,
	touchUp TouchUpCallback,
	clickKey ClickKeyCallback,
	inputText InputTextCallback,
	keyDown KeyDownCallback,
	keyUp KeyUpCallback,
) uintptr {
	callbacks := &MaaCustomControllerCallbacks{
		Connect:     purego.NewCallback(connect),
		RequestUUID: purego.NewCallback(requestUUID),
		StartApp:    purego.NewCallback(startApp),
		StopApp:     purego.NewCallback(stopApp),
		Screencap:   purego.NewCallback(screencap),
		Click:       purego.NewCallback(click),
		Swipe:       purego.NewCallback(swipe),
		TouchDown:   purego.NewCallback(touchDown),
		TouchMove:   purego.NewCallback(touchMove),
		TouchUp:     purego.NewCallback(touchUp),
		ClickKey:    purego.NewCallback(clickKey),
		InputText:   purego.NewCallback(inputText),
		KeyDown:     purego.NewCallback(keyDown),
		KeyUp:       purego.NewCallback(keyUp),
	}
	return uintptr(unsafe.Pointer(callbacks))
}

var (
	MaaContextRunTask          func(context uintptr, entry, pipelineOverride string) int64
	MaaContextRunRecognition   func(context uintptr, entry, pipelineOverride string, image uintptr) int64
	MaaContextRunAction        func(context uintptr, entry, pipelineOverride string, box uintptr, recoDetail string) int64
	MaaContextOverridePipeline func(context uintptr, pipelineOverride string) bool
	MaaContextOverrideNext     func(context uintptr, nodeName string, nextList uintptr) bool
	MaaContextGetNodeData      func(context uintptr, nodeName string, buffer uintptr) bool
	MaaContextGetTaskId        func(context uintptr) int64
	MaaContextGetTasker        func(context uintptr) uintptr
	MaaContextClone            func(context uintptr) uintptr
)

var (
	MaaStringBufferCreate  func() uintptr
	MaaStringBufferDestroy func(handle uintptr)
	MaaStringBufferIsEmpty func(handle uintptr) bool
	MaaStringBufferClear   func(handle uintptr) bool
	MaaStringBufferGet     func(handle uintptr) string
	MaaStringBufferSize    func(handle uintptr) uint64
	MaaStringBufferSet     func(handle uintptr, str string) bool
	MaaStringBufferSetEx   func(handle uintptr, str string, size uint64) bool

	MaaStringListBufferCreate  func() uintptr
	MaaStringListBufferDestroy func(handle uintptr)
	MaaStringListBufferIsEmpty func(handle uintptr) bool
	MaaStringListBufferSize    func(handle uintptr) uint64
	MaaStringListBufferAt      func(handle uintptr, index uint64) uintptr
	MaaStringListBufferAppend  func(handle uintptr, value uintptr) bool
	MaaStringListBufferRemove  func(handle uintptr, index uint64) bool
	MaaStringListBufferClear   func(handle uintptr) bool

	MaaImageBufferCreate     func() uintptr
	MaaImageBufferDestroy    func(handle uintptr)
	MaaImageBufferIsEmpty    func(handle uintptr) bool
	MaaImageBufferClear      func(handle uintptr) bool
	MaaImageBufferGetRawData func(handle uintptr) unsafe.Pointer
	MaaImageBufferWidth      func(handle uintptr) int32
	MaaImageBufferHeight     func(handle uintptr) int32
	MaaImageBufferChannels   func(handle uintptr) int32
	MaaImageBufferType       func(handle uintptr) int32
	MaaImageBufferSetRawData func(handle uintptr, data unsafe.Pointer, width, height, imageType int32) bool

	MaaImageListBufferCreate  func() uintptr
	MaaImageListBufferDestroy func(handle uintptr)
	MaaImageListBufferIsEmpty func(handle uintptr) bool
	MaaImageListBufferSize    func(handle uintptr) uint64
	MaaImageListBufferAt      func(handle uintptr, index uint64) uintptr
	MaaImageListBufferAppend  func(handle uintptr, value uintptr) bool
	MaaImageListBufferRemove  func(handle uintptr, index uint64) bool
	MaaImageListBufferClear   func(handle uintptr) bool

	MaaRectCreate  func() uintptr
	MaaRectDestroy func(handle uintptr)
	MaaRectGetX    func(handle uintptr) int32
	MaaRectGetY    func(handle uintptr) int32
	MaaRectGetW    func(handle uintptr) int32
	MaaRectGetH    func(handle uintptr) int32
	MaaRectSet     func(handle uintptr, x, y, w, h int32) bool
)

type MaaGlobalOption int32

const (
	MaaGlobalOption_Invalid MaaGlobalOption = 0

	// MaaGlobalOption_LogDir Log dir
	//
	// value: string, eg: "C:\\Users\\Administrator\\Desktop\\log"; val_size: string length
	MaaGlobalOption_LogDir MaaGlobalOption = 1

	// MaaGlobalOption_SaveDraw Whether to save draw
	//
	// value: bool, eg: true; val_size: sizeof(bool)
	MaaGlobalOption_SaveDraw MaaGlobalOption = 2

	// MaaGlobalOption_StdoutLevel The level of log output to stdout
	//
	// value: MaaLoggingLevel, val_size: sizeof(MaaLoggingLevel)
	// default value is MaaLoggingLevel_Error
	MaaGlobalOption_StdoutLevel MaaGlobalOption = 4

	// MaaGlobalOption_DebugMode Whether to debug
	//
	// value: bool, eg: true; val_size: sizeof(bool)
	MaaGlobalOption_DebugMode MaaGlobalOption = 6
)

var MaaSetGlobalOption func(key MaaGlobalOption, value unsafe.Pointer, valSize uint64) bool

func initFramework(libDir string) error {
	libName := getMaaFrameworkLibrary()
	libPath := filepath.Join(libDir, libName)

	handle, err := openLibrary(libPath)
	if err != nil {
		return err
	}

	maaFramework = handle

	registerFramework()

	return nil
}

func getMaaFrameworkLibrary() string {
	switch runtime.GOOS {
	case "darwin":
		return "libMaaFramework.dylib"
	case "linux":
		return "libMaaFramework.so"
	case "windows":
		return "MaaFramework.dll"
	default:
		panic(fmt.Errorf("GOOS=%s is not supported", runtime.GOOS))
	}
}

func registerFramework() {
	purego.RegisterLibFunc(&MaaVersion, maaFramework, "MaaVersion")
	// Tasker
	purego.RegisterLibFunc(&MaaTaskerCreate, maaFramework, "MaaTaskerCreate")
	purego.RegisterLibFunc(&MaaTaskerDestroy, maaFramework, "MaaTaskerDestroy")
	purego.RegisterLibFunc(&MaaTaskerAddSink, maaFramework, "MaaTaskerAddSink")
	purego.RegisterLibFunc(&MaaTaskerRemoveSink, maaFramework, "MaaTaskerRemoveSink")
	purego.RegisterLibFunc(&MaaTaskerClearSinks, maaFramework, "MaaTaskerClearSinks")
	purego.RegisterLibFunc(&MaaTaskerSetOption, maaFramework, "MaaTaskerSetOption")
	purego.RegisterLibFunc(&MaaTaskerBindResource, maaFramework, "MaaTaskerBindResource")
	purego.RegisterLibFunc(&MaaTaskerBindController, maaFramework, "MaaTaskerBindController")
	purego.RegisterLibFunc(&MaaTaskerInited, maaFramework, "MaaTaskerInited")
	purego.RegisterLibFunc(&MaaTaskerPostTask, maaFramework, "MaaTaskerPostTask")
	purego.RegisterLibFunc(&MaaTaskerStopping, maaFramework, "MaaTaskerStopping")
	purego.RegisterLibFunc(&MaaTaskerStatus, maaFramework, "MaaTaskerStatus")
	purego.RegisterLibFunc(&MaaTaskerWait, maaFramework, "MaaTaskerWait")
	purego.RegisterLibFunc(&MaaTaskerRunning, maaFramework, "MaaTaskerRunning")
	purego.RegisterLibFunc(&MaaTaskerPostStop, maaFramework, "MaaTaskerPostStop")
	purego.RegisterLibFunc(&MaaTaskerGetResource, maaFramework, "MaaTaskerGetResource")
	purego.RegisterLibFunc(&MaaTaskerGetController, maaFramework, "MaaTaskerGetController")
	purego.RegisterLibFunc(&MaaTaskerClearCache, maaFramework, "MaaTaskerClearCache")
	purego.RegisterLibFunc(&MaaTaskerGetRecognitionDetail, maaFramework, "MaaTaskerGetRecognitionDetail")
	purego.RegisterLibFunc(&MaaTaskerGetNodeDetail, maaFramework, "MaaTaskerGetNodeDetail")
	purego.RegisterLibFunc(&MaaTaskerGetTaskDetail, maaFramework, "MaaTaskerGetTaskDetail")
	purego.RegisterLibFunc(&MaaTaskerGetLatestNode, maaFramework, "MaaTaskerGetLatestNode")
	// Resource
	purego.RegisterLibFunc(&MaaResourceCreate, maaFramework, "MaaResourceCreate")
	purego.RegisterLibFunc(&MaaResourceDestroy, maaFramework, "MaaResourceDestroy")
	purego.RegisterLibFunc(&MaaResourceAddSink, maaFramework, "MaaResourceAddSink")
	purego.RegisterLibFunc(&MaaResourceRemoveSink, maaFramework, "MaaResourceRemoveSink")
	purego.RegisterLibFunc(&MaaResourceClearSinks, maaFramework, "MaaResourceClearSinks")
	purego.RegisterLibFunc(&MaaResourceRegisterCustomRecognition, maaFramework, "MaaResourceRegisterCustomRecognition")
	purego.RegisterLibFunc(&MaaResourceUnregisterCustomRecognition, maaFramework, "MaaResourceUnregisterCustomRecognition")
	purego.RegisterLibFunc(&MaaResourceClearCustomRecognition, maaFramework, "MaaResourceClearCustomRecognition")
	purego.RegisterLibFunc(&MaaResourceRegisterCustomAction, maaFramework, "MaaResourceRegisterCustomAction")
	purego.RegisterLibFunc(&MaaResourceUnregisterCustomAction, maaFramework, "MaaResourceUnregisterCustomAction")
	purego.RegisterLibFunc(&MaaResourceClearCustomAction, maaFramework, "MaaResourceClearCustomAction")
	purego.RegisterLibFunc(&MaaResourcePostBundle, maaFramework, "MaaResourcePostBundle")
	purego.RegisterLibFunc(&MaaResourceOverridePipeline, maaFramework, "MaaResourceOverridePipeline")
	purego.RegisterLibFunc(&MaaResourceOverrideNext, maaFramework, "MaaResourceOverrideNext")
	purego.RegisterLibFunc(&MaaResourceGetNodeData, maaFramework, "MaaResourceGetNodeData")
	purego.RegisterLibFunc(&MaaResourceClear, maaFramework, "MaaResourceClear")
	purego.RegisterLibFunc(&MaaResourceStatus, maaFramework, "MaaResourceStatus")
	purego.RegisterLibFunc(&MaaResourceWait, maaFramework, "MaaResourceWait")
	purego.RegisterLibFunc(&MaaResourceLoaded, maaFramework, "MaaResourceLoaded")
	purego.RegisterLibFunc(&MaaResourceSetOption, maaFramework, "MaaResourceSetOption")
	purego.RegisterLibFunc(&MaaResourceGetHash, maaFramework, "MaaResourceGetHash")
	purego.RegisterLibFunc(&MaaResourceGetNodeList, maaFramework, "MaaResourceGetNodeList")
	// Controller
	purego.RegisterLibFunc(&MaaAdbControllerCreate, maaFramework, "MaaAdbControllerCreate")
	purego.RegisterLibFunc(&MaaWin32ControllerCreate, maaFramework, "MaaWin32ControllerCreate")
	purego.RegisterLibFunc(&MaaCustomControllerCreate, maaFramework, "MaaCustomControllerCreate")
	purego.RegisterLibFunc(&MaaDbgControllerCreate, maaFramework, "MaaDbgControllerCreate")
	purego.RegisterLibFunc(&MaaControllerDestroy, maaFramework, "MaaControllerDestroy")
	purego.RegisterLibFunc(&MaaControllerAddSink, maaFramework, "MaaControllerAddSink")
	purego.RegisterLibFunc(&MaaControllerRemoveSink, maaFramework, "MaaControllerRemoveSink")
	purego.RegisterLibFunc(&MaaControllerClearSinks, maaFramework, "MaaControllerClearSinks")
	purego.RegisterLibFunc(&MaaControllerSetOption, maaFramework, "MaaControllerSetOption")
	purego.RegisterLibFunc(&MaaControllerPostConnection, maaFramework, "MaaControllerPostConnection")
	purego.RegisterLibFunc(&MaaControllerPostClick, maaFramework, "MaaControllerPostClick")
	purego.RegisterLibFunc(&MaaControllerPostSwipe, maaFramework, "MaaControllerPostSwipe")
	purego.RegisterLibFunc(&MaaControllerPostClickKey, maaFramework, "MaaControllerPostClickKey")
	purego.RegisterLibFunc(&MaaControllerPostInputText, maaFramework, "MaaControllerPostInputText")
	purego.RegisterLibFunc(&MaaControllerPostStartApp, maaFramework, "MaaControllerPostStartApp")
	purego.RegisterLibFunc(&MaaControllerPostStopApp, maaFramework, "MaaControllerPostStopApp")
	purego.RegisterLibFunc(&MaaControllerPostTouchDown, maaFramework, "MaaControllerPostTouchDown")
	purego.RegisterLibFunc(&MaaControllerPostTouchMove, maaFramework, "MaaControllerPostTouchMove")
	purego.RegisterLibFunc(&MaaControllerPostTouchUp, maaFramework, "MaaControllerPostTouchUp")
	purego.RegisterLibFunc(&MaaControllerPostKeyDown, maaFramework, "MaaControllerPostKeyDown")
	purego.RegisterLibFunc(&MaaControllerPostKeyUp, maaFramework, "MaaControllerPostKeyUp")
	purego.RegisterLibFunc(&MaaControllerPostScreencap, maaFramework, "MaaControllerPostScreencap")
	purego.RegisterLibFunc(&MaaControllerStatus, maaFramework, "MaaControllerStatus")
	purego.RegisterLibFunc(&MaaControllerWait, maaFramework, "MaaControllerWait")
	purego.RegisterLibFunc(&MaaControllerConnected, maaFramework, "MaaControllerConnected")
	purego.RegisterLibFunc(&MaaControllerCachedImage, maaFramework, "MaaControllerCachedImage")
	purego.RegisterLibFunc(&MaaControllerGetUuid, maaFramework, "MaaControllerGetUuid")
	// Context
	purego.RegisterLibFunc(&MaaContextRunTask, maaFramework, "MaaContextRunTask")
	purego.RegisterLibFunc(&MaaContextRunRecognition, maaFramework, "MaaContextRunRecognition")
	purego.RegisterLibFunc(&MaaContextRunAction, maaFramework, "MaaContextRunAction")
	purego.RegisterLibFunc(&MaaContextOverridePipeline, maaFramework, "MaaContextOverridePipeline")
	purego.RegisterLibFunc(&MaaContextOverrideNext, maaFramework, "MaaContextOverrideNext")
	purego.RegisterLibFunc(&MaaContextGetNodeData, maaFramework, "MaaContextGetNodeData")
	purego.RegisterLibFunc(&MaaContextGetTaskId, maaFramework, "MaaContextGetTaskId")
	purego.RegisterLibFunc(&MaaContextGetTasker, maaFramework, "MaaContextGetTasker")
	purego.RegisterLibFunc(&MaaContextClone, maaFramework, "MaaContextClone")
	// Buffer
	purego.RegisterLibFunc(&MaaStringBufferCreate, maaFramework, "MaaStringBufferCreate")
	purego.RegisterLibFunc(&MaaStringBufferDestroy, maaFramework, "MaaStringBufferDestroy")
	purego.RegisterLibFunc(&MaaStringBufferIsEmpty, maaFramework, "MaaStringBufferIsEmpty")
	purego.RegisterLibFunc(&MaaStringBufferClear, maaFramework, "MaaStringBufferClear")
	purego.RegisterLibFunc(&MaaStringBufferGet, maaFramework, "MaaStringBufferGet")
	purego.RegisterLibFunc(&MaaStringBufferSize, maaFramework, "MaaStringBufferSize")
	purego.RegisterLibFunc(&MaaStringBufferSet, maaFramework, "MaaStringBufferSet")
	purego.RegisterLibFunc(&MaaStringBufferSetEx, maaFramework, "MaaStringBufferSetEx")
	purego.RegisterLibFunc(&MaaStringListBufferCreate, maaFramework, "MaaStringListBufferCreate")
	purego.RegisterLibFunc(&MaaStringListBufferDestroy, maaFramework, "MaaStringListBufferDestroy")
	purego.RegisterLibFunc(&MaaStringListBufferIsEmpty, maaFramework, "MaaStringListBufferIsEmpty")
	purego.RegisterLibFunc(&MaaStringListBufferSize, maaFramework, "MaaStringListBufferSize")
	purego.RegisterLibFunc(&MaaStringListBufferAt, maaFramework, "MaaStringListBufferAt")
	purego.RegisterLibFunc(&MaaStringListBufferAppend, maaFramework, "MaaStringListBufferAppend")
	purego.RegisterLibFunc(&MaaStringListBufferRemove, maaFramework, "MaaStringListBufferRemove")
	purego.RegisterLibFunc(&MaaStringListBufferClear, maaFramework, "MaaStringListBufferClear")
	purego.RegisterLibFunc(&MaaImageBufferCreate, maaFramework, "MaaImageBufferCreate")
	purego.RegisterLibFunc(&MaaImageBufferDestroy, maaFramework, "MaaImageBufferDestroy")
	purego.RegisterLibFunc(&MaaImageBufferIsEmpty, maaFramework, "MaaImageBufferIsEmpty")
	purego.RegisterLibFunc(&MaaImageBufferClear, maaFramework, "MaaImageBufferClear")
	purego.RegisterLibFunc(&MaaImageBufferGetRawData, maaFramework, "MaaImageBufferGetRawData")
	purego.RegisterLibFunc(&MaaImageBufferWidth, maaFramework, "MaaImageBufferWidth")
	purego.RegisterLibFunc(&MaaImageBufferHeight, maaFramework, "MaaImageBufferHeight")
	purego.RegisterLibFunc(&MaaImageBufferChannels, maaFramework, "MaaImageBufferChannels")
	purego.RegisterLibFunc(&MaaImageBufferType, maaFramework, "MaaImageBufferType")
	purego.RegisterLibFunc(&MaaImageBufferSetRawData, maaFramework, "MaaImageBufferSetRawData")
	purego.RegisterLibFunc(&MaaImageListBufferCreate, maaFramework, "MaaImageListBufferCreate")
	purego.RegisterLibFunc(&MaaImageListBufferDestroy, maaFramework, "MaaImageListBufferDestroy")
	purego.RegisterLibFunc(&MaaImageListBufferIsEmpty, maaFramework, "MaaImageListBufferIsEmpty")
	purego.RegisterLibFunc(&MaaImageListBufferSize, maaFramework, "MaaImageListBufferSize")
	purego.RegisterLibFunc(&MaaImageListBufferAt, maaFramework, "MaaImageListBufferAt")
	purego.RegisterLibFunc(&MaaImageListBufferAppend, maaFramework, "MaaImageListBufferAppend")
	purego.RegisterLibFunc(&MaaImageListBufferRemove, maaFramework, "MaaImageListBufferRemove")
	purego.RegisterLibFunc(&MaaImageListBufferClear, maaFramework, "MaaImageListBufferClear")
	purego.RegisterLibFunc(&MaaRectCreate, maaFramework, "MaaRectCreate")
	purego.RegisterLibFunc(&MaaRectDestroy, maaFramework, "MaaRectDestroy")
	purego.RegisterLibFunc(&MaaRectGetX, maaFramework, "MaaRectGetX")
	purego.RegisterLibFunc(&MaaRectGetY, maaFramework, "MaaRectGetY")
	purego.RegisterLibFunc(&MaaRectGetW, maaFramework, "MaaRectGetW")
	purego.RegisterLibFunc(&MaaRectGetH, maaFramework, "MaaRectGetH")
	purego.RegisterLibFunc(&MaaRectSet, maaFramework, "MaaRectSet")
	// Option
	purego.RegisterLibFunc(&MaaSetGlobalOption, maaFramework, "MaaSetGlobalOption")
}
