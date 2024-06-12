package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include "controller_custom.h"

extern void _MaaAPICallbackAgent(_GoString_ msg, _GoString_ detailsJson, MaaTransparentArg callbackArg);

extern uint8_t _ConnectAgent(MaaTransparentArg handleArg);
extern uint8_t _RequestUUIDAgent(MaaTransparentArg handle_arg, MaaStringBufferHandle buffer);
extern uint8_t _RequestResolutionAgent(MaaTransparentArg handle_arg, int32_t* width, int32_t* height);
extern uint8_t _StartAppAgent(_GoString_ intent, MaaTransparentArg handle_arg);
extern uint8_t _StopAppAgent(_GoString_ intent, MaaTransparentArg handle_arg);
extern uint8_t _ScreencapAgent(MaaTransparentArg handle_arg, MaaImageBufferHandle buffer);
extern uint8_t _ClickAgent(int32_t x, int32_t y, MaaTransparentArg handle_arg);
extern uint8_t _SwipeAgent(
			int32_t x1,
			int32_t y1,
			int32_t x2,
			int32_t y2,
			int32_t duration,
			MaaTransparentArg handle_arg);
extern uint8_t _TouchDownAgent(
			int32_t contact,
            int32_t x,
            int32_t y,
            int32_t pressure,
            MaaTransparentArg handle_arg);
extern uint8_t _TouchMoveAgent(
			int32_t contact,
            int32_t x,
            int32_t y,
            int32_t pressure,
            MaaTransparentArg handle_arg);
extern uint8_t _TouchUpAgent(int32_t contact, MaaTransparentArg handle_arg);
extern uint8_t _PressKey(int32_t keycode, MaaTransparentArg handle_arg);
extern uint8_t _InputText(_GoString_ text, MaaTransparentArg handle_arg);
*/
import "C"
import (
	"unsafe"
)

type CustomController struct {
	handle C.MaaCustomControllerHandle

	connect           func(handleArg interface{}) bool
	requestUUID       func(handleArg interface{}) (string, bool)
	requestResolution func(handleArg interface{}) (width, height int32, ok bool)
	startApp          func(intent string, handleArg interface{}) bool
	stopApp           func(intent string, handleArg interface{}) bool
	screencap         func(handle interface{}) (ImageBuffer, bool)
	click             func(x, y int32, handleArg interface{}) bool
	swipe             func(x1, y1, x2, y2, duration int32, handleArg interface{}) bool
	touchDown         func(contact, x, y, pressure int32, handleArg interface{}) bool
	touchMove         func(contact, x, y, pressure int32, handleArg interface{}) bool
	touchUp           func(contact int32, handleArg interface{}) bool
	pressKey          func(keycode int32, handleArg interface{}) bool
	inputText         func(text string, handleArg interface{}) bool
}

func (c *CustomController) Set(
	connect func(handleArg interface{}) bool,
	requestUUID func(handleArg interface{}) (string, bool),
	requestResolution func(handleArg interface{}) (width, height int32, ok bool),
	startApp func(intent string, handleArg interface{}) bool,
	stopApp func(intent string, handleArg interface{}) bool,
	screencap func(handle interface{}) (ImageBuffer, bool),
	click func(x, y int32, handleArg interface{}) bool,
	swipe func(x1, y1, x2, y2, duration int32, handleArg interface{}) bool,
	touchDown func(contact, x, y, pressure int32, handleArg interface{}) bool,
	touchMove func(contact, x, y, pressure int32, handleArg interface{}) bool,
	touchUp func(contact int32, handleArg interface{}) bool,
	pressKey func(keycode int32, handleArg interface{}) bool,
	inputText func(text string, handleArg interface{}) bool,
) {
	c.connect = connect
	c.requestUUID = requestUUID
	c.requestResolution = requestResolution
	c.startApp = startApp
	c.stopApp = stopApp
	c.screencap = screencap
	c.click = click
	c.swipe = swipe
	c.touchDown = touchDown
	c.touchMove = touchMove
	c.touchUp = touchUp
	c.pressKey = pressKey
	c.inputText = inputText
	c.handle = C.MaaCustomControllerHandleCreate(
		C.ConnectCallback(C._ConnectAgent),
		C.RequestUUIDCallback(C._RequestUUIDAgent),
		C.RequestResolutionCallback(C._RequestResolutionAgent),
		C.StartAppCallback(C._StartAppAgent),
		C.StopAppCallback(C._StopAppAgent),
		C.ScreencapCallback(C._ScreencapAgent),
		C.ClickCallback(C._ClickAgent),
		C.SwipeCallback(C._SwipeAgent),
		C.TouchDownCallback(C._TouchDownAgent),
		C.TouchMoveCallback(C._TouchMoveAgent),
		C.TouchUpCallback(C._TouchUpAgent),
		C.PressKeyCallback(C._PressKey),
		C.InputTextCallback(C._InputText),
	)
}

func (c *CustomController) Handle() unsafe.Pointer {
	return unsafe.Pointer(c.handle)
}

func (c *CustomController) Destroy() {
	C.MaaCustomControllerHandleDestroy(c.handle)
}

type customControllerAgent struct {
	ctrl *CustomController
	arg  interface{}
}

//export _ConnectAgent
func _ConnectAgent(handleArg unsafe.Pointer) C.uint8_t {
	agent := (*customControllerAgent)(handleArg)
	ctrl := agent.ctrl
	ok := ctrl.connect(agent.arg)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _RequestUUIDAgent
func _RequestUUIDAgent(handleArg unsafe.Pointer, buffer C.MaaStringBufferHandle) C.uint8_t {
	agent := (*customControllerAgent)(handleArg)
	ctrl := agent.ctrl
	uuid, ok := ctrl.requestUUID(agent.arg)
	if ok {
		uuidStringBuffer := &stringBuffer{handle: buffer}
		uuidStringBuffer.Set(uuid)
	}
	return C.uint8_t(0)
}

//export _RequestResolutionAgent
func _RequestResolutionAgent(handleArg unsafe.Pointer, width, height *C.int32_t) C.uint8_t {
	agent := (*customControllerAgent)(handleArg)
	ctrl := agent.ctrl
	w, h, ok := ctrl.requestResolution(agent.arg)
	if ok {
		*width = C.int32_t(w)
		*height = C.int32_t(h)
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _StartAppAgent
func _StartAppAgent(intent string, handleArg unsafe.Pointer) C.uint8_t {
	agent := (*customControllerAgent)(handleArg)
	ctrl := agent.ctrl
	ok := ctrl.startApp(intent, agent.arg)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _StopAppAgent
func _StopAppAgent(intent string, handleArg unsafe.Pointer) C.uint8_t {
	agent := (*customControllerAgent)(handleArg)
	ctrl := agent.ctrl
	ok := ctrl.stopApp(intent, agent.arg)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _ScreencapAgent
func _ScreencapAgent(handleArg unsafe.Pointer, buffer C.MaaImageBufferHandle) C.uint8_t {
	agent := (*customControllerAgent)(handleArg)
	ctrl := agent.ctrl
	img, ok := ctrl.screencap(agent.arg)
	defer img.Destroy()
	if ok {
		imgBuffer := &imageBuffer{handle: buffer}
		imgBuffer.SetRawData(img.GetRawData(), img.GetWidth(), img.GetHeight(), img.GetType())
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _ClickAgent
func _ClickAgent(x, y C.int32_t, handleArg unsafe.Pointer) C.uint8_t {
	agent := (*customControllerAgent)(handleArg)
	ctrl := agent.ctrl
	ok := ctrl.click(int32(x), int32(y), agent.arg)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _SwipeAgent
func _SwipeAgent(x1, y1, x2, y2, duration C.int32_t, handleArg unsafe.Pointer) C.uint8_t {
	agent := (*customControllerAgent)(handleArg)
	ctrl := agent.ctrl
	ok := ctrl.swipe(int32(x1), int32(y1), int32(x2), int32(y2), int32(duration), agent.arg)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _TouchDownAgent
func _TouchDownAgent(contact, x, y, pressure C.int32_t, handleArg unsafe.Pointer) C.uint8_t {
	agent := (*customControllerAgent)(handleArg)
	ctrl := agent.ctrl
	ok := ctrl.touchDown(int32(contact), int32(x), int32(y), int32(pressure), agent.arg)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _TouchMoveAgent
func _TouchMoveAgent(contact, x, y, pressure C.int32_t, handleArg unsafe.Pointer) C.uint8_t {
	agent := (*customControllerAgent)(handleArg)
	ctrl := agent.ctrl
	ok := ctrl.touchMove(int32(contact), int32(x), int32(y), int32(pressure), agent.arg)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _TouchUpAgent
func _TouchUpAgent(contact C.int32_t, handleArg unsafe.Pointer) C.uint8_t {
	agent := (*customControllerAgent)(handleArg)
	ctrl := agent.ctrl
	ok := ctrl.touchUp(int32(contact), agent.arg)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _PressKey
func _PressKey(key C.int32_t, handleArg unsafe.Pointer) C.uint8_t {
	agent := (*customControllerAgent)(handleArg)
	ctrl := agent.ctrl
	ok := ctrl.pressKey(int32(key), agent.arg)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

//export _InputText
func _InputText(text string, handleArg unsafe.Pointer) C.uint8_t {
	agent := (*customControllerAgent)(handleArg)
	ctrl := agent.ctrl
	ok := ctrl.inputText(text, agent.arg)
	if ok {
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}

func NewCustomController(
	customCtrl *CustomController,
	handleArg interface{},
	callback func(msg, detailsJson string, callbackArg interface{}),
	callbackArg interface{},
) Controller {
	ctrlAgent := &customControllerAgent{ctrl: customCtrl, arg: handleArg}
	cbAgent := &callbackAgent{callback: callback, arg: callbackArg}
	handle := C.MaaCustomControllerCreate(
		customCtrl.handle,
		C.MaaTransparentArg(unsafe.Pointer(ctrlAgent)),
		C.MaaAPICallback(C._MaaAPICallbackAgent),
		C.MaaTransparentArg(unsafe.Pointer(cbAgent)),
	)
	return &controller{handle: handle}
}
