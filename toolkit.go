package maa

/*
#include <stdlib.h>
#include <MaaToolkit/MaaToolkitAPI.h>

extern uint8_t _MaaCustomRecognitionCallbackAgent(
	MaaContext* ctx,
	int64_t task_id,
	const char* current_task_name,
	const char* custom_recognizer_name,
	const char* custom_recognition_param,
	const MaaImageBuffer* image,
	const MaaRect* roi,
	void* recognizer_arg,
	MaaRect* out_box,
	MaaStringBuffer* out_detail);

extern uint8_t _MaaCustomActionCallbackAgent(
	MaaContext* ctx,
	int64_t task_id,
	const char* current_task_name,
	const char* custom_action_name,
	const char* custom_action_param,
	int64_t rec_id,
	const MaaRect* box ,
	void* actionArg);

extern void _MaaNotificationCallbackAgent(const char* message, const char* details_json, void* callback_arg);
*/
import "C"
import (
	"github.com/MaaXYZ/maa-framework-go/internal/notification"
	"unsafe"
)

// AdbDevice represents a single ADB device with various properties about its information.
type AdbDevice struct {
	Name            string
	AdbPath         string
	Address         string
	ScreencapMethod AdbScreencapMethod
	InputMethod     AdbInputMethod
	Config          string
}

// DesktopWindow represents a single desktop window with various properties about its information.
type DesktopWindow struct {
	Handle     unsafe.Pointer
	ClassName  string
	WindowName string
}

type Toolkit struct{}

// NewToolkit creates a new toolkit instance.
func NewToolkit() *Toolkit {
	return &Toolkit{}
}

// ConfigInitOption inits the toolkit config option.
func (t *Toolkit) ConfigInitOption(userPath, defaultJson string) bool {
	cUserPath := C.CString(userPath)
	defer C.free(unsafe.Pointer(cUserPath))
	cDefaultJson := C.CString(defaultJson)
	defer C.free(unsafe.Pointer(cDefaultJson))

	return C.MaaToolkitConfigInitOption(cUserPath, cDefaultJson) != 0
}

// FindAdbDevices finds adb devices.
func (t *Toolkit) FindAdbDevices(specifiedAdb ...string) []*AdbDevice {
	listHandle := C.MaaToolkitAdbDeviceListCreate()
	defer C.MaaToolkitAdbDeviceListDestroy(listHandle)
	var got C.uint8_t
	if len(specifiedAdb) > 0 {
		cAdbPath := C.CString(specifiedAdb[0])
		defer C.free(unsafe.Pointer(cAdbPath))
		got = C.MaaToolkitAdbDeviceFindSpecified(cAdbPath, listHandle)
	} else {
		got = C.MaaToolkitAdbDeviceFind(listHandle)
	}
	if got == 0 {
		return nil
	}

	size := uint64(C.MaaToolkitAdbDeviceListSize(listHandle))
	list := make([]*AdbDevice, size)
	for i := uint64(0); i < size; i++ {
		deviceHandle := C.MaaToolkitAdbDeviceListAt(listHandle, C.uint64_t(i))
		name := C.GoString(C.MaaToolkitAdbDeviceGetName(deviceHandle))
		adbPath := C.GoString(C.MaaToolkitAdbDeviceGetAdbPath(deviceHandle))
		address := C.GoString(C.MaaToolkitAdbDeviceGetAddress(deviceHandle))
		screencapMethod := AdbScreencapMethod(C.MaaToolkitAdbDeviceGetScreencapMethods(deviceHandle))
		inputMethod := AdbInputMethod(C.MaaToolkitAdbDeviceGetInputMethods(deviceHandle))
		config := C.GoString(C.MaaToolkitAdbDeviceGetConfig(deviceHandle))
		list[i] = &AdbDevice{
			Name:            name,
			AdbPath:         adbPath,
			Address:         address,
			ScreencapMethod: screencapMethod,
			InputMethod:     inputMethod,
			Config:          config,
		}
	}
	return list
}

// FindDesktopWindows finds desktop windows.
func (t *Toolkit) FindDesktopWindows() []*DesktopWindow {
	listHandle := C.MaaToolkitDesktopWindowListCreate()
	defer C.MaaToolkitDesktopWindowListDestroy(listHandle)
	got := C.MaaToolkitDesktopWindowFindAll(listHandle)
	if got == 0 {
		return nil
	}

	size := uint64(C.MaaToolkitDesktopWindowListSize(listHandle))
	list := make([]*DesktopWindow, size)
	for i := uint64(0); i < size; i++ {
		windowHandle := C.MaaToolkitDesktopWindowListAt(listHandle, C.uint64_t(i))
		handle := unsafe.Pointer(C.MaaToolkitDesktopWindowGetHandle(windowHandle))
		className := C.GoString(C.MaaToolkitDesktopWindowGetClassName(windowHandle))
		windowName := C.GoString(C.MaaToolkitDesktopWindowGetWindowName(windowHandle))
		list[i] = &DesktopWindow{
			Handle:     handle,
			ClassName:  className,
			WindowName: windowName,
		}
	}
	return list
}

type piStoreValue struct {
	CustomRecognizersCallbackID map[string]uint64
	CustomActionsCallbackID     map[string]uint64
}

var piStore = make(map[uint64]piStoreValue)

// RegisterPICustomRecognition registers a custom recognizer.
func (t *Toolkit) RegisterPICustomRecognition(instId uint64, name string, recognition CustomRecognition) {
	id := registerCustomRecognition(recognition)
	if _, ok := piStore[instId]; !ok {
		piStore[instId] = piStoreValue{
			CustomRecognizersCallbackID: make(map[string]uint64),
			CustomActionsCallbackID:     make(map[string]uint64),
		}
	}
	piStore[instId].CustomRecognizersCallbackID[name] = id

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	C.MaaToolkitProjectInterfaceRegisterCustomRecognition(
		C.uint64_t(instId),
		cName,
		C.MaaCustomRecognitionCallback(C._MaaCustomRecognitionCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

// RegisterPICustomAction registers a custom action.
func (t *Toolkit) RegisterPICustomAction(instId uint64, name string, action CustomAction) {
	id := registerCustomAction(action)
	if _, ok := piStore[instId]; !ok {
		piStore[instId] = piStoreValue{
			CustomRecognizersCallbackID: make(map[string]uint64),
			CustomActionsCallbackID:     make(map[string]uint64),
		}
	}
	piStore[instId].CustomActionsCallbackID[name] = id

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	C.MaaToolkitProjectInterfaceRegisterCustomAction(
		C.uint64_t(instId),
		cName,
		C.MaaCustomActionCallback(C._MaaCustomActionCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

// ClearPICustom unregisters all custom recognitions and actions for a given instance.
func (t *Toolkit) ClearPICustom(instId uint64) {
	value := piStore[instId]
	for _, id := range value.CustomRecognizersCallbackID {
		unregisterCustomRecognition(id)
	}
	for _, id := range value.CustomActionsCallbackID {
		unregisterCustomAction(id)
	}
}

// RunCli runs the PI CLI.
func (t *Toolkit) RunCli(instId uint64, resourcePath, userPath string, directly bool, callback func(msg, detailsJson string)) bool {
	cResourcePath := C.CString(resourcePath)
	defer C.free(unsafe.Pointer(cResourcePath))
	cUserPath := C.CString(userPath)
	defer C.free(unsafe.Pointer(cUserPath))
	var cDirectly uint8
	if directly {
		cDirectly = 1
	}
	id := notification.RegisterCallback(callback)
	got := C.MaaToolkitProjectInterfaceRunCli(
		C.uint64_t(instId),
		cResourcePath,
		cUserPath,
		C.uint8_t(cDirectly),
		C.MaaNotificationCallback(C._MaaNotificationCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	return got != 0
}
