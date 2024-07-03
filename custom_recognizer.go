package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include "custom_recognizer.h"

extern uint8_t _AnalyzeAgent(
			MaaSyncContextHandle sync_context,
            const MaaImageBufferHandle image,
            MaaStringView task_name,
            MaaStringView custom_recognition_param,
            MaaTransparentArg recognizer_arg,
           	MaaRectHandle out_box,
			MaaStringBufferHandle out_detail);
*/
import "C"
import "unsafe"

// CustomRecognizerImpl defines an interface for custom recognizer.
// Implementers of this interface must embed a CustomRecognizerHandler struct
// and provide an implementation for the Analyze method.
type CustomRecognizerImpl interface {
	Analyze(syncCtx SyncContext, image ImageBuffer, taskName, RecognitionParam string) (AnalyzeResult, bool)

	Handle() unsafe.Pointer
	Destroy()
}

type AnalyzeResult struct {
	Box    RectBuffer
	Detail string
}

type CustomRecognizerHandler struct {
	handle C.MaaCustomRecognizerHandle
}

func NewCustomRecognizerHandler() CustomRecognizerHandler {
	return CustomRecognizerHandler{handle: C.MaaCustomRecognizerHandleCreate(C.AnalyzeCallback(C._AnalyzeAgent))}
}

func (r CustomRecognizerHandler) Handle() unsafe.Pointer {
	return unsafe.Pointer(r.handle)
}

func (r CustomRecognizerHandler) Destroy() {
	C.MaaCustomRecognizerHandleDestroy(r.handle)
}

//export _AnalyzeAgent
func _AnalyzeAgent(
	ctx C.MaaSyncContextHandle,
	image C.MaaImageBufferHandle,
	taskName, customRecognitionParam C.MaaStringView,
	recognizerArg C.MaaTransparentArg,
	outBox C.MaaRectHandle,
	outDetail C.MaaStringBufferHandle,
) C.uint8_t {
	rec := *(*CustomRecognizerImpl)(unsafe.Pointer(recognizerArg))

	ret, ok := rec.Analyze(
		SyncContext{handle: ctx},
		&imageBuffer{handle: image},
		C.GoString(taskName),
		C.GoString(customRecognitionParam),
	)
	if ok {
		box := ret.Box
		outBoxRect := &rectBuffer{handle: outBox}
		outBoxRect.Set(box.GetX(), box.GetY(), box.GetW(), box.GetH())
		outDetailString := &stringBuffer{handle: outDetail}
		outDetailString.Set(ret.Detail)
		return C.uint8_t(1)
	}
	return C.uint8_t(0)
}
