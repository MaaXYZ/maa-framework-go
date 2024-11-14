package maa

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	MaaVersion func() string
)

type MaaNotificationCallback func(message, detailsJson *byte, notifyTransArg unsafe.Pointer) uintptr

type MaaTaskerOption int32

var (
	MaaTaskerCreate               func(notify MaaNotificationCallback, notifyTransArg unsafe.Pointer) uintptr
	MaaTaskerDestroy              func(tasker uintptr)
	MaaTaskerSetOption            func(tasker uintptr, key MaaTaskerOption, value unsafe.Pointer, valSize uint64) bool
	MaaTaskerBindResource         func(tasker uintptr, res uintptr) bool
	MaaTaskerBindController       func(tasker uintptr, ctrl uintptr) bool
	MaaTaskerInited               func(tasker uintptr) bool
	MaaTaskerPostPipeline         func(tasker uintptr, entry, pipelineOverride string) int64
	MaaTaskerStatus               func(tasker uintptr, id int64) int32
	MaaTaskerWait                 func(tasker uintptr, id int64) int32
	MaaTaskerRunning              func(tasker uintptr) bool
	MaaTaskerPostStop             func(tasker uintptr) bool
	MaaTaskerGetResource          func(tasker uintptr) uintptr
	MaaTaskerGetController        func(tasker uintptr) uintptr
	MaaTaskerClearCache           func(tasker uintptr) bool
	MaaTaskerGetRecognitionDetail func(tasker uintptr, recoId int64, name uintptr, algorithm uintptr, hit *bool, box uintptr, detailJson uintptr, raw uintptr, draws uintptr) bool
	MaaTaskerGetNodeDetail        func(tasker uintptr, nodeId int64, name uintptr, recoId *int64, completed *bool) bool
	MaaTaskerGetTaskDetail        func(tasker uintptr, taskId int64, entry uintptr, nodeIdList uintptr, nodeIdListSize *uint64, status *int32) bool
	MaaTaskerGetLatestNode        func(tasker uintptr, taskName string, latestId *int64) bool
)

type MaaCustomRecognitionCallback func(context uintptr, taskId int64, currentTaskName, customRecognitionName, customRecognitionParam *byte, image, roi uintptr, transArg unsafe.Pointer, outBox, outDetail uintptr) uint64

type MaaCustomActionCallback func(context uintptr, taskId int64, currentTaskName, customActionName, customActionParam *byte, recoId int64, box uintptr, transArg unsafe.Pointer) uint64

type MaaResOption int32

const (
	MaaResOption_Invalid MaaResOption = 0

	/// Use the specified inference device.
	/// Please set this option before loading the model.
	///
	/// value: MaaInferenceDevice, eg: 0; val_size: sizeof(MaaInferenceDevice)
	/// default value is MaaInferenceDevice_Auto
	MaaResOption_InterfaceDevice MaaResOption = 1
)

var (
	MaaResourceCreate                      func(notify MaaNotificationCallback, notifyTransArg unsafe.Pointer) uintptr
	MaaResourceDestroy                     func(res uintptr)
	MaaResourceRegisterCustomRecognition   func(res uintptr, name string, recognition MaaCustomRecognitionCallback, transArg unsafe.Pointer) bool
	MaaResourceUnregisterCustomRecognition func(res uintptr, name string) bool
	MaaResourceClearCustomRecognition      func(res uintptr) bool
	MaaResourceRegisterCustomAction        func(res uintptr, name string, action MaaCustomActionCallback, transArg unsafe.Pointer) bool
	MaaResourceUnregisterCustomAction      func(res uintptr, name string) bool
	MaaResourceClearCustomAction           func(res uintptr) bool
	MaaResourcePostPath                    func(res uintptr, path string) int64
	MaaResourceClear                       func(res uintptr) bool
	MaaResourceStatus                      func(res uintptr, id int64) int32
	MaaResourceWait                        func(res uintptr, id int64) int32
	MaaResourceLoaded                      func(res uintptr) bool
	MaaResourceSetOption                   func(res uintptr, key MaaResOption, value unsafe.Pointer, valSize uint64) bool
	MaaResourceGetHash                     func(res uintptr, buffer uintptr) bool
	MaaResourceGetTaskList                 func(res uintptr, buffer uintptr) bool
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

	/// MaaCtrlOptionRecording indicates that all screenshots and actions should be dumped.
	// Recording will evaluate to true if either this or MaaGlobalOptionEnum::MaaGlobalOption_Recording is true.
	MaaCtrlOption_Recording MaaCtrlOption = 5
)

var (
	MaaAdbControllerCreate      func(adbPath, address string, screencapMethods MaaAdbScreencapMethod, inputMethods MaaAdbInputMethod, config, agentPath string, notify MaaNotificationCallback, notifyTransArg unsafe.Pointer) uintptr
	MaaWin32ControllerCreate    func(hWnd unsafe.Pointer, screencapMethods MaaWin32ScreencapMethod, inputMethods MaaWin32InputMethod, notify MaaNotificationCallback, notifyTransArg unsafe.Pointer) uintptr
	MaaCustomControllerCreate   func(controller uintptr, controllerArg unsafe.Pointer, notify MaaNotificationCallback, notifyTransArg unsafe.Pointer) uintptr
	MaaDbgControllerCreate      func(readPath, writePath string, dbgCtrlType MaaDbgControllerType, config string, notify MaaNotificationCallback, notifyTransArg unsafe.Pointer) uintptr
	MaaControllerDestroy        func(ctrl uintptr)
	MaaControllerSetOption      func(ctrl uintptr, key MaaCtrlOption, value unsafe.Pointer, valSize uint64) bool
	MaaControllerPostConnection func(ctrl uintptr) int64
	MaaControllerPostClick      func(ctrl uintptr, x, y int32) int64
	MaaControllerPostSwipe      func(ctrl uintptr, x1, y1, x2, y2, duration int32) int64
	MaaControllerPostPressKey   func(ctrl uintptr, keycode int32) int64
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
	MaaControllerPostScreencap func(ctrl uintptr) int64
	MaaControllerStatus        func(ctrl uintptr, id int64) int32
	MaaControllerWait          func(ctrl uintptr, id int64) int32
	MaaControllerConnected     func(ctrl uintptr) bool
	MaaControllerCachedImage   func(ctrl uintptr, buffer uintptr) bool
	MaaControllerGetUuid       func(ctrl uintptr, buffer uintptr) bool
)

func init() {
	maaFramework, err := openLibrary(getMaaFrameworkLibrary())
	if err != nil {
		panic(err)
	}

	purego.RegisterLibFunc(&MaaVersion, maaFramework, "MaaVersion")
	// Tasker
	purego.RegisterLibFunc(&MaaTaskerCreate, maaFramework, "MaaTaskerCreate")
	purego.RegisterLibFunc(&MaaTaskerDestroy, maaFramework, "MaaTaskerDestroy")
	purego.RegisterLibFunc(&MaaTaskerSetOption, maaFramework, "MaaTaskerSetOption")
	purego.RegisterLibFunc(&MaaTaskerBindResource, maaFramework, "MaaTaskerBindResource")
	purego.RegisterLibFunc(&MaaTaskerBindController, maaFramework, "MaaTaskerBindController")
	purego.RegisterLibFunc(&MaaTaskerInited, maaFramework, "MaaTaskerInited")
	purego.RegisterLibFunc(&MaaTaskerPostPipeline, maaFramework, "MaaTaskerPostPipeline")
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
	purego.RegisterLibFunc(&MaaResourceRegisterCustomRecognition, maaFramework, "MaaResourceRegisterCustomRecognition")
	purego.RegisterLibFunc(&MaaResourceUnregisterCustomRecognition, maaFramework, "MaaResourceUnregisterCustomRecognition")
	purego.RegisterLibFunc(&MaaResourceClearCustomRecognition, maaFramework, "MaaResourceClearCustomRecognition")
	purego.RegisterLibFunc(&MaaResourceRegisterCustomAction, maaFramework, "MaaResourceRegisterCustomAction")
	purego.RegisterLibFunc(&MaaResourceUnregisterCustomAction, maaFramework, "MaaResourceUnregisterCustomAction")
	purego.RegisterLibFunc(&MaaResourceClearCustomAction, maaFramework, "MaaResourceClearCustomAction")
	purego.RegisterLibFunc(&MaaResourcePostPath, maaFramework, "MaaResourcePostPath")
	purego.RegisterLibFunc(&MaaResourceClear, maaFramework, "MaaResourceClear")
	purego.RegisterLibFunc(&MaaResourceStatus, maaFramework, "MaaResourceStatus")
	purego.RegisterLibFunc(&MaaResourceWait, maaFramework, "MaaResourceWait")
	purego.RegisterLibFunc(&MaaResourceLoaded, maaFramework, "MaaResourceLoaded")
	purego.RegisterLibFunc(&MaaResourceSetOption, maaFramework, "MaaResourceSetOption")
	purego.RegisterLibFunc(&MaaResourceGetHash, maaFramework, "MaaResourceGetHash")
	purego.RegisterLibFunc(&MaaResourceGetTaskList, maaFramework, "MaaResourceGetTaskList")
	// Controller
	purego.RegisterLibFunc(&MaaAdbControllerCreate, maaFramework, "MaaAdbControllerCreate")
	purego.RegisterLibFunc(&MaaWin32ControllerCreate, maaFramework, "MaaWin32ControllerCreate")
	purego.RegisterLibFunc(&MaaCustomControllerCreate, maaFramework, "MaaCustomControllerCreate")
	purego.RegisterLibFunc(&MaaDbgControllerCreate, maaFramework, "MaaDbgControllerCreate")
	purego.RegisterLibFunc(&MaaControllerDestroy, maaFramework, "MaaControllerDestroy")
	purego.RegisterLibFunc(&MaaControllerSetOption, maaFramework, "MaaControllerSetOption")
	purego.RegisterLibFunc(&MaaControllerPostConnection, maaFramework, "MaaControllerPostConnection")
	purego.RegisterLibFunc(&MaaControllerPostClick, maaFramework, "MaaControllerPostClick")
	purego.RegisterLibFunc(&MaaControllerPostSwipe, maaFramework, "MaaControllerPostSwipe")
	purego.RegisterLibFunc(&MaaControllerPostPressKey, maaFramework, "MaaControllerPostPressKey")
	purego.RegisterLibFunc(&MaaControllerPostInputText, maaFramework, "MaaControllerPostInputText")
	purego.RegisterLibFunc(&MaaControllerPostStartApp, maaFramework, "MaaControllerPostStartApp")
	purego.RegisterLibFunc(&MaaControllerPostStopApp, maaFramework, "MaaControllerPostStopApp")
	purego.RegisterLibFunc(&MaaControllerPostTouchDown, maaFramework, "MaaControllerPostTouchDown")
	purego.RegisterLibFunc(&MaaControllerPostTouchMove, maaFramework, "MaaControllerPostTouchMove")
	purego.RegisterLibFunc(&MaaControllerPostTouchUp, maaFramework, "MaaControllerPostTouchUp")
	purego.RegisterLibFunc(&MaaControllerPostScreencap, maaFramework, "MaaControllerPostScreencap")
	purego.RegisterLibFunc(&MaaControllerStatus, maaFramework, "MaaControllerStatus")
	purego.RegisterLibFunc(&MaaControllerWait, maaFramework, "MaaControllerWait")
	purego.RegisterLibFunc(&MaaControllerConnected, maaFramework, "MaaControllerConnected")
	purego.RegisterLibFunc(&MaaControllerCachedImage, maaFramework, "MaaControllerCachedImage")
	purego.RegisterLibFunc(&MaaControllerGetUuid, maaFramework, "MaaControllerGetUuid")

}
