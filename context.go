package maa

import (
	"image"

	"github.com/MaaXYZ/maa-framework-go/v2/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v2/internal/maa"
)

type Context struct {
	handle uintptr
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

func (ctx *Context) runTask(entry, override string) *TaskDetail {
	taskId := maa.MaaContextRunTask(ctx.handle, entry, override)
	if taskId == 0 {
		return nil
	}
	tasker := ctx.GetTasker()
	return tasker.getTaskDetail(taskId)
}

// RunTask runs a task and returns its detail.
// It accepts an entry string and an optional override parameter which can be
// a JSON string or any data type that can be marshaled to JSON.
// If multiple overrides are provided, only the first one will be used.
//
// Example 1:
//
//	ctx.RunTask("Task", `{"Task":{"action":"Click","target":[100, 200, 100, 100]}}`)
//
// Example 2:
//
//	ctx.RunTask("Task", map[string]interface{}{
//	    "Task": map[string]interface{}{
//	        "action": "Click",
//	        "target": []int{100, 200, 100, 100},
//		}
//	})
func (ctx *Context) RunTask(entry string, override ...any) *TaskDetail {
	return ctx.runTask(entry, ctx.handleOverride(override...))
}

func (ctx *Context) runRecognition(entry, override string, img image.Image) *RecognitionDetail {
	imgBuf := buffer.NewImageBuffer()
	imgBuf.Set(img)
	defer imgBuf.Destroy()

	recId := maa.MaaContextRunRecognition(ctx.handle, entry, override, uintptr(imgBuf.Handle()))
	if recId == 0 {
		return nil
	}
	tasker := ctx.GetTasker()
	return tasker.getRecognitionDetail(recId)
}

// RunRecognition run a recognition and return its detail.
// It accepts an entry string and an optional override parameter which can be
// a JSON string or any data type that can be marshaled to JSON.
// If multiple overrides are provided, only the first one will be used.
//
// Example 1:
//
//	ctx.RunRecognition("Task", `{"Task":{"recognition":"OCR","expected":"Hello"}}`)
//
// Example 2:
//
//	ctx.RunRecognition("Task", map[string]interface{}{
//	    "Task": map[string]interface{}{
//	        "recognition": "OCR",
//	        "expected": "Hello",
//		}
//	})
func (ctx *Context) RunRecognition(entry string, img image.Image, override ...any) *RecognitionDetail {
	return ctx.runRecognition(entry, ctx.handleOverride(override...), img)
}

func (ctx *Context) runAction(entry, override string, box Rect, recognitionDetail string) *NodeDetail {
	rectBuf := buffer.NewRectBuffer()
	rectBuf.Set(box)
	defer rectBuf.Destroy()

	nodeId := maa.MaaContextRunAction(ctx.handle, entry, override, uintptr(rectBuf.Handle()), recognitionDetail)
	if nodeId == 0 {
		return nil
	}
	tasker := ctx.GetTasker()
	return tasker.getNodeDetail(nodeId)
}

// RunAction run an action and return its detail.
// It accepts an entry string and an optional override parameter which can be
// a JSON string or any data type that can be marshaled to JSON.
// If multiple overrides are provided, only the first one will be used.
//
// Example 1:
//
//	ctx.RunAction("Task", `{"Task":{"action":"Click","target":[100, 200, 100, 100]}}`)
//
// Example 2:
//
//	ctx.RunAction("Task", map[string]interface{}{
//	    "Task": map[string]interface{}{
//	        "action": "Click",
//	        "target": []int{100, 200, 100, 100},
//		}
//	})
func (ctx *Context) RunAction(entry string, box Rect, recognitionDetail string, override ...any) *NodeDetail {
	return ctx.runAction(entry, ctx.handleOverride(override...), box, recognitionDetail)
}

func (ctx *Context) overridePipeline(override string) bool {
	return maa.MaaContextOverridePipeline(ctx.handle, override)
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
	return maa.MaaContextOverrideNext(ctx.handle, name, uintptr(list.Handle()))
}

// GetTaskJob returns current task job.
func (ctx *Context) GetTaskJob() *TaskJob {
	tasker := ctx.GetTasker()
	taskId := maa.MaaContextGetTaskId(ctx.handle)
	return NewTaskJob(taskId, tasker.status, tasker.wait, tasker.getTaskDetail)
}

// GetTasker return current Tasker.
func (ctx *Context) GetTasker() *Tasker {
	handle := maa.MaaContextGetTasker(ctx.handle)
	return &Tasker{handle: handle}
}

// Clone clones current Context.
func (ctx *Context) Clone() *Context {
	handle := maa.MaaContextClone(ctx.handle)
	return &Context{handle: handle}
}
