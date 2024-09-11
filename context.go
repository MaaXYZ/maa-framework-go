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

func (ctx *Context) handleOverride(override ...any) string {
	if len(override) == 0 {
		return "{}"
	}
	if str, ok := override[0].(string); ok {
		return str
	}
	str, err := toJSON(override[0])
	if err != nil {
		str = "{}"
	}
	return str
}

func (ctx *Context) runPipeline(entry, override string) *TaskDetail {
	cEntry := C.CString(entry)
	defer C.free(unsafe.Pointer(cEntry))
	cOverride := C.CString(override)
	defer C.free(unsafe.Pointer(cOverride))

	taskId := int64(C.MaaContextRunPipeline(ctx.handle, cEntry, cOverride))
	tasker := ctx.GetTasker()
	return tasker.getTaskDetail(taskId)
}

// RunPipeline runs a pipeline and return it detail.
// `override` is an optional parameter. If provided, it should be a single value
// that can be a JSON string or any data type that can be marshaled to JSON.
// If multiple values are provided, only the first one will be used.
func (ctx *Context) RunPipeline(entry string, override ...any) *TaskDetail {
	return ctx.runPipeline(entry, ctx.handleOverride(override...))
}

func (ctx *Context) runRecognition(entry, override string, img image.Image) *RecognitionDetail {
	cEntry := C.CString(entry)
	defer C.free(unsafe.Pointer(cEntry))
	cOverride := C.CString(override)
	defer C.free(unsafe.Pointer(cOverride))
	imgBuf := buffer.NewImageBuffer()
	_ = imgBuf.SetRawData(img)
	defer imgBuf.Destroy()

	recId := int64(C.MaaContextRunRecognition(ctx.handle, cEntry, cOverride, (*C.MaaImageBuffer)(imgBuf.Handle())))
	tasker := ctx.GetTasker()
	return tasker.getRecognitionDetail(recId)
}

// RunRecognition run a recognition and return it detail.
// `override` is an optional parameter. If provided, it should be a single value
// that can be a JSON string or any data type that can be marshaled to JSON.
// If multiple values are provided, only the first one will be used.
func (ctx *Context) RunRecognition(entry string, img image.Image, override ...any) *RecognitionDetail {
	return ctx.runRecognition(entry, ctx.handleOverride(override...), img)
}

func (ctx *Context) runAction(entry, override string, box Rect, recognitionDetail string) *NodeDetail {
	cEntry := C.CString(entry)
	defer C.free(unsafe.Pointer(cEntry))
	cOverride := C.CString(override)
	defer C.free(unsafe.Pointer(cOverride))
	rectBuf := newRectBuffer()
	rectBuf.Set(box)
	defer rectBuf.Destroy()
	cRecognitionDetail := C.CString(recognitionDetail)
	defer C.free(unsafe.Pointer(cRecognitionDetail))

	nodeId := int64(C.MaaContextRunAction(ctx.handle, cEntry, cOverride, (*C.MaaRect)(rectBuf.Handle()), cRecognitionDetail))
	tasker := ctx.GetTasker()
	return tasker.getNodeDetail(nodeId)
}

// RunAction run an action and return it detail.
// `override` is an optional parameter. If provided, it should be a single value
// that can be a JSON string or any data type that can be marshaled to JSON.
// If multiple values are provided, only the first one will be used.
func (ctx *Context) RunAction(entry string, box Rect, recognitionDetail string, override ...any) *NodeDetail {
	return ctx.runAction(entry, ctx.handleOverride(override...), box, recognitionDetail)
}

func (ctx *Context) overridePipeline(override string) bool {
	cPipelineOverride := C.CString(override)
	defer C.free(unsafe.Pointer(cPipelineOverride))

	got := C.MaaContextOverridePipeline(ctx.handle, cPipelineOverride)
	return got != 0
}

// OverridePipeline overrides pipeline.
// The `override` parameter can be a JSON string or any data type that can be marshaled to JSON.
func (ctx *Context) OverridePipeline(override any) bool {
	if str, ok := override.(string); ok {
		return ctx.overridePipeline(str)
	}
	str, err := toJSON(override)
	if err != nil {
		return false
	}
	return ctx.overridePipeline(str)
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
