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

// CustomRecognizer defines an interface for custom recognizer.
// Implementers of this interface must embed a Recognizer struct
// and provide an implementation for the Analyze method.
type CustomRecognizer interface {
	Analyze(syncCtx SyncContext, image ImageBuffer, taskName, RecognitionParam string) (AnalyzeResult, bool)

	Handle() unsafe.Pointer
	Destroy()
}

type AnalyzeResult struct {
	Box    RectBuffer
	Detail string
}

type Recognizer struct {
	handle C.MaaCustomRecognizerHandle
}

func NewRecognizer() Recognizer {
	return Recognizer{handle: C.MaaCustomRecognizerHandleCreate(C.AnalyzeCallback(C._AnalyzeAgent))}
}

func (r Recognizer) Handle() unsafe.Pointer {
	return unsafe.Pointer(r.handle)
}

func (r Recognizer) Destroy() {
	C.MaaCustomRecognizerHandleDestroy(r.handle)
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
	if recognizerArg == nil {
		return C.uint8_t(0)
	}

	rec := *(*CustomRecognizer)(recognizerArg)

	ret, ok := rec.Analyze(
		SyncContext(syncCtx),
		&imageBuffer{handle: image},
		taskName,
		customRecognitionParam,
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
