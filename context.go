package maa

import (
	"encoding/json"
	"errors"
	"image"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
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

	if override[0] == nil {
		return "{}"
	}

	str, err := json.Marshal(override[0])
	if err != nil {
		str = []byte("{}")
	}
	return string(str)
}

func (ctx *Context) runTask(entry, override string) *TaskDetail {
	taskId := native.MaaContextRunTask(ctx.handle, entry, override)
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

	recId := native.MaaContextRunRecognition(ctx.handle, entry, override, imgBuf.Handle())
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

func (ctx *Context) runAction(entry, override string, box Rect, recognitionDetail string) *ActionDetail {
	rectBuf := buffer.NewRectBuffer()
	rectBuf.Set(box)
	defer rectBuf.Destroy()

	actId := native.MaaContextRunAction(ctx.handle, entry, override, rectBuf.Handle(), recognitionDetail)
	if actId == 0 {
		return nil
	}
	tasker := ctx.GetTasker()
	return tasker.getActionDetail(actId)
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
func (ctx *Context) RunAction(entry string, box Rect, recognitionDetail string, override ...any) *ActionDetail {
	return ctx.runAction(entry, ctx.handleOverride(override...), box, recognitionDetail)
}

func (ctx *Context) overridePipeline(override string) bool {
	return native.MaaContextOverridePipeline(ctx.handle, override)
}

// OverridePipeline overrides pipeline.
// The `override` parameter can be a JSON string or any data type that can be marshaled to JSON.
func (ctx *Context) OverridePipeline(override any) bool {
	switch v := override.(type) {
	case string:
		return ctx.overridePipeline(v)
	case []byte:
		return ctx.overridePipeline(string(v))
	default:
		if v == nil {
			return ctx.overridePipeline("{}")
		}

		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return false
		}
		return ctx.overridePipeline(string(jsonBytes))
	}
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
	return native.MaaContextOverrideNext(ctx.handle, name, list.Handle())
}

func (ctx *Context) OverrideImage(imageName string, image image.Image) bool {
	img := buffer.NewImageBuffer()
	defer img.Destroy()
	img.Set(image)
	return native.MaaContextOverrideImage(ctx.handle, imageName, img.Handle())
}

// GetNodeJSON gets the node JSON by name.
func (ctx *Context) GetNodeJSON(name string) (string, bool) {
	buf := buffer.NewStringBuffer()
	defer buf.Destroy()
	ok := native.MaaResourceGetNodeData(ctx.handle, name, buf.Handle())
	return buf.Get(), ok
}

func (ctx *Context) GetNodeData(name string) (*Node, error) {
	raw, ok := ctx.GetNodeJSON(name)
	if !ok {
		return nil, errors.New("node not found")
	}

	var node Node
	err := json.Unmarshal([]byte(raw), &node)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// GetTaskJob returns current task job.
func (ctx *Context) GetTaskJob() *TaskJob {
	tasker := ctx.GetTasker()
	taskId := native.MaaContextGetTaskId(ctx.handle)
	return newTaskJob(taskId, tasker.status, tasker.wait, tasker.getTaskDetail)
}

// GetTasker return current Tasker.
func (ctx *Context) GetTasker() *Tasker {
	handle := native.MaaContextGetTasker(ctx.handle)
	return &Tasker{handle: handle}
}

// Clone clones current Context.
func (ctx *Context) Clone() *Context {
	handle := native.MaaContextClone(ctx.handle)
	return &Context{handle: handle}
}

// SetAnchor sets an anchor by name.
func (ctx *Context) SetAnchor(anchorName, nodeName string) bool {
	return native.MaaContextSetAnchor(ctx.handle, anchorName, nodeName)
}

// GetAnchor gets an anchor by name.
func (ctx *Context) GetAnchor(anchorName string) (string, bool) {
	buf := buffer.NewStringBuffer()
	defer buf.Destroy()
	ok := native.MaaContextGetAnchor(ctx.handle, anchorName, buf.Handle())
	return buf.Get(), ok
}

// GetHitCount gets the hit count of a node by name.
func (ctx *Context) GetHitCount(nodeName string) (uint64, bool) {
	var count uint64
	ok := native.MaaContextGetHitCount(ctx.handle, nodeName, &count)
	return count, ok
}

// ClearHitCount clears the hit count of a node by name.
func (ctx *Context) ClearHitCount(nodeName string) bool {
	return native.MaaContextClearHitCount(ctx.handle, nodeName)
}
