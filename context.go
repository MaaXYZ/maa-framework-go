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

// RunPipeline runs a pipeline and return it detail.
func (ctx *Context) RunPipeline(entry, pipelineOverride string) *TaskDetail {
	cEntry := C.CString(entry)
	defer C.free(unsafe.Pointer(cEntry))
	cPipelineOverride := C.CString(pipelineOverride)
	defer C.free(unsafe.Pointer(cPipelineOverride))

	taskId := int64(C.MaaContextRunPipeline(ctx.handle, cEntry, cPipelineOverride))
	tasker := ctx.GetTasker()
	return tasker.getTaskDetail(taskId)
}

// RunRecognition run a recognition and return it detail.
func (ctx *Context) RunRecognition(entry, pipelineOverride string, img image.Image) *RecognitionDetail {
	cEntry := C.CString(entry)
	defer C.free(unsafe.Pointer(cEntry))
	cPipelineOverride := C.CString(pipelineOverride)
	defer C.free(unsafe.Pointer(cPipelineOverride))
	imgBuf := buffer.NewImageBuffer()
	_ = imgBuf.SetRawData(img)
	defer imgBuf.Destroy()

	recId := int64(C.MaaContextRunRecognition(ctx.handle, cEntry, cPipelineOverride, (*C.MaaImageBuffer)(imgBuf.Handle())))
	tasker := ctx.GetTasker()
	return tasker.getRecognitionDetail(recId)
}

// RunAction run an action and return it detail.
func (ctx *Context) RunAction(entry, pipelineOverride string, box Rect, recognitionDetail string) *NodeDetail {
	cEntry := C.CString(entry)
	defer C.free(unsafe.Pointer(cEntry))
	cPipelineOverride := C.CString(pipelineOverride)
	defer C.free(unsafe.Pointer(cPipelineOverride))
	rectBuf := newRectBuffer()
	rectBuf.Set(box)
	defer rectBuf.Destroy()
	cRecognitionDetail := C.CString(recognitionDetail)
	defer C.free(unsafe.Pointer(cRecognitionDetail))

	nodeId := int64(C.MaaContextRunAction(ctx.handle, cEntry, cPipelineOverride, (*C.MaaRect)(rectBuf.Handle()), cRecognitionDetail))
	tasker := ctx.GetTasker()
	return tasker.getNodeDetail(nodeId)
}

// OverridePipeline overrides pipeline.
func (ctx *Context) OverridePipeline(pipelineOverride string) bool {
	cPipelineOverride := C.CString(pipelineOverride)
	defer C.free(unsafe.Pointer(cPipelineOverride))

	got := C.MaaContextOverridePipeline(ctx.handle, cPipelineOverride)
	return got != 0
}

// OverrideNext overrides the next list of task by name.
func (ctx *Context) OverrideNext(name string, nextList []string) bool {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	list := buffer.NewStringListBuffer()
	defer list.Destroy()
	size := len(nextList)
	items := make([]*buffer.StringBuffer, size)
	for i := 0; i < size; i++ {
		items[i] = buffer.NewStringBuffer()
		items[i].Set(nextList[i])
		list.Append(items[i])
	}
	defer func() {
		for _, item := range items {
			item.Destroy()
		}
	}()
	got := C.MaaContextOverrideNext(ctx.handle, cName, (*C.MaaStringListBuffer)(list.Handle()))
	return got != 0
}

// GetTaskId returns current task id.
func (ctx *Context) GetTaskId() int64 {
	return int64(C.MaaContextGetTaskId(ctx.handle))
}

// GetTasker return current Tasker.
func (ctx *Context) GetTasker() *Tasker {
	handle := C.MaaContextGetTasker(ctx.handle)
	return &Tasker{handle: handle}
}

// Clone clones current Context.
func (ctx *Context) Clone() *Context {
	handle := C.MaaContextClone(ctx.handle)
	return &Context{handle: handle}
}
