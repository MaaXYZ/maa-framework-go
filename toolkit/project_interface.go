package toolkit

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include <MaaToolkit/MaaToolkitAPI.h>

extern void _MaaNotificationCallbackAgent(const char* message, const char* details_json, void* callback_arg);

extern uint8_t _MaaCustomRecognizerCallbackAgent(
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
*/
import "C"
import (
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/MaaXYZ/maa-framework-go/internal/notification"
	"unsafe"
)

var piSore = make(map[uint64]map[string][]string)

type ProjectInterface struct {
}

// NewProjectInterface creates a new ProjectInterface instance.
func NewProjectInterface() *ProjectInterface {
	return &ProjectInterface{}
}

// RegisterCustomRecognizer registers a custom recognizer.
func (i *ProjectInterface) RegisterCustomRecognizer(instId uint64, name string, recognizer maa.CustomRecognizer) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	id := maa.RegisterCustomRecognizer(name, recognizer)
	if _, ok := piSore[instId]; !ok {
		piSore[instId] = make(map[string][]string)
	}
	piSore[instId]["recognizer"] = append(piSore[instId]["recognizer"], name)
	C.MaaToolkitProjectInterfaceRegisterCustomRecognition(
		C.uint64_t(instId),
		cName,
		C.MaaCustomRecognizerCallback(C._MaaCustomRecognizerCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

// RegisterCustomAction registers a custom action.
func (i *ProjectInterface) RegisterCustomAction(instId uint64, name string, action maa.CustomAction) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	id := maa.RegisterCustomAction(name, action)
	if _, ok := piSore[instId]; !ok {
		piSore[instId] = make(map[string][]string)
	}
	piSore[instId]["action"] = append(piSore[instId]["action"], name)
	C.MaaToolkitProjectInterfaceRegisterCustomAction(
		C.uint64_t(instId),
		cName,
		C.MaaCustomActionCallback(C._MaaCustomActionCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
}

// ClearCustom unregisters all custom recognizers and actions for a given instance.
func (i *ProjectInterface) ClearCustom(instId uint64) {
	if _, ok := piSore[instId]["recognizer"]; ok {
		for _, name := range piSore[instId]["recognizer"] {
			maa.UnregisterCustomRecognizer(name)
		}
	}
	if _, ok := piSore[instId]["action"]; ok {
		for _, name := range piSore[instId]["action"] {
			maa.UnregisterCustomAction(name)
		}
	}
}

// RunCli runs the PI CLI.
func (i *ProjectInterface) RunCli(instId uint64, resourcePath, userPath string, directly bool, callback func(msg, detailsJson string)) bool {
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
		cDirectly,
		C.MaaNotificationCallback(C._MaaNotificationCallbackAgent),
		// Here, we are simply passing the uint64 value as a pointer
		// and will not actually dereference this pointer.
		unsafe.Pointer(uintptr(id)),
	)
	return got != 0
}
