package maa

/*
#include <stdlib.h>
#include <MaaFramework/MaaAPI.h>
*/
import "C"
import (
	"github.com/MaaXYZ/maa-framework-go/buffer"
	"image"
	"unsafe"
)

type Context struct {
	handle *C.MaaContext
}

func (ctx *Context) RunPipeline(entry, pipelineOverride string) int64 {
	cEntry := C.CString(entry)
	defer C.free(unsafe.Pointer(cEntry))
	cPipelineOverride := C.CString(pipelineOverride)
	defer C.free(unsafe.Pointer(cPipelineOverride))

	return int64(C.MaaContextRunPipeline(ctx.handle, cEntry, cPipelineOverride))
}

func (ctx *Context) RunRecognition(entry, pipelineOverride string, img image.Image) int64 {
	cEntry := C.CString(entry)
	defer C.free(unsafe.Pointer(cEntry))
	cPipelineOverride := C.CString(pipelineOverride)
	defer C.free(unsafe.Pointer(cPipelineOverride))
	imgBuf := buffer.NewImageBuffer()
	_ = imgBuf.SetRawData(img)
	defer imgBuf.Destroy()

	return int64(C.MaaContextRunRecognition(ctx.handle, cEntry, cPipelineOverride, (*C.MaaImageBuffer)(imgBuf.Handle())))
}

func (ctx *Context) RunAction(entry, pipelineOverride string, box Rect, recognitionDetail string) int64 {
	cEntry := C.CString(entry)
	defer C.free(unsafe.Pointer(cEntry))
	cPipelineOverride := C.CString(pipelineOverride)
	defer C.free(unsafe.Pointer(cPipelineOverride))
	rectBuf := newRectBuffer()
	rectBuf.Set(box)
	defer rectBuf.Destroy()
	cRecognitionDetail := C.CString(recognitionDetail)
	defer C.free(unsafe.Pointer(cRecognitionDetail))

	return int64(C.MaaContextRunAction(ctx.handle, cEntry, cPipelineOverride, (*C.MaaRect)(rectBuf.Handle()), cRecognitionDetail))
}

func (ctx *Context) OverridePipeline(pipelineOverride string) bool {
	cPipelineOverride := C.CString(pipelineOverride)
	defer C.free(unsafe.Pointer(cPipelineOverride))

	got := C.MaaContextOverridePipeline(ctx.handle, cPipelineOverride)
	return got != 0
}

func (ctx *Context) GetTaskId() int64 {
	return int64(C.MaaContextGetTaskId(ctx.handle))
}

func (ctx *Context) GetTasker() *Tasker {
	handle := C.MaaContextGetTasker(ctx.handle)
	return &Tasker{handle: handle}
}

func (ctx *Context) Clone() *Context {
	handle := C.MaaContextClone(ctx.handle)
	return &Context{handle: handle}
}
