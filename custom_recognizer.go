package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
#include "custom_recognizer.h"

extern uint8_t _AnalyzeAgent(MaaSyncContextHandle sync_context,
            const MaaImageBufferHandle image,
            _GoString_ task_name,
            _GoString_ custom_recognition_param,
            MaaTransparentArg recognizer_arg,
           	MaaRectHandle out_box,
			MaaStringBufferHandle out_detail);
*/
import "C"
import "unsafe"

type AnalyzeResult struct {
	Box    RectBuffer
	Detail string
}

type CustomRecognizer struct {
	handle C.MaaCustomRecognizerHandle

	analyze func(
		syncCtx SyncContext,
		image ImageBuffer,
		taskName, customRecognitionParam string,
		recognizerArg interface{},
	) (AnalyzeResult, bool)
}

func (r *CustomRecognizer) Set(
	analyze func(
		syncCtx SyncContext,
		image ImageBuffer,
		taskName, customRecognitionParam string,
		recognizerArg interface{},
	) (AnalyzeResult, bool),
) {
	r.analyze = analyze
	r.handle = C.MaaCustomRecognizerHandleCreate(C.AnalyzeCallback(C._AnalyzeAgent))
}

func (r *CustomRecognizer) Handle() unsafe.Pointer {
	return unsafe.Pointer(r.handle)
}

func (r *CustomRecognizer) Destroy() {
	C.MaaCustomRecognizerHandleDestroy(r.handle)
}

type customRecognizerAgent struct {
	rec *CustomRecognizer
	arg interface{}
}

//export _AnalyzeAgent
func _AnalyzeAgent(
	syncCtx C.MaaSyncContextHandle,
	image C.MaaImageBufferHandle,
	taskName, customRecognitionParam string,
	recognizerArg unsafe.Pointer,
	outBox C.MaaRectHandle,
	outDetail C.MaaStringBufferHandle,
) C.uint8_t {
	agent := (*customRecognizerAgent)(recognizerArg)
	rec := agent.rec
	arg := agent.arg

	ret, ok := rec.analyze(
		SyncContext(syncCtx),
		&imageBuffer{handle: image},
		taskName,
		customRecognitionParam,
		arg,
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
