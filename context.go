package maa

import (
	"encoding/json"
	"errors"
	"image"
	"time"

	"github.com/MaaXYZ/maa-framework-go/v3/internal/buffer"
	"github.com/MaaXYZ/maa-framework-go/v3/internal/native"
)

// Context provides the runtime context for custom actions/recognitions
// and exposes task, recognition, action, and pipeline operations.
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

func (ctx *Context) runTask(entry, override string) (*TaskDetail, error) {
	taskId := native.MaaContextRunTask(ctx.handle, entry, override)
	if taskId == 0 {
		return nil, errors.New("failed to run task")
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
func (ctx *Context) RunTask(entry string, override ...any) (*TaskDetail, error) {
	return ctx.runTask(entry, ctx.handleOverride(override...))
}

func (ctx *Context) runRecognition(
	entry, override string,
	img image.Image,
) (*RecognitionDetail, error) {
	imgBuf := buffer.NewImageBuffer()
	imgBuf.Set(img)
	defer imgBuf.Destroy()

	recId := native.MaaContextRunRecognition(ctx.handle, entry, override, imgBuf.Handle())
	if recId == 0 {
		return nil, errors.New("failed to run recognition")
	}
	tasker := ctx.GetTasker()
	recognitionDetail, err := tasker.getRecognitionDetail(recId)
	return recognitionDetail, err
}

// RunRecognition runs a recognition and returns its detail.
// It accepts an entry string and an optional override parameter which can be
// a JSON string or any data type that can be marshaled to JSON.
// If multiple overrides are provided, only the first one will be used.
//
// Example 1:
//
//	ctx.RunRecognition("Task", img, `{"Task":{"recognition":"OCR","expected":"Hello"}}`)
//
// Example 2:
//
//	ctx.RunRecognition("Task", img, map[string]interface{}{
//	    "Task": map[string]interface{}{
//	        "recognition": "OCR",
//	        "expected": "Hello",
//		}
//	})
func (ctx *Context) RunRecognition(
	entry string,
	img image.Image,
	override ...any,
) (*RecognitionDetail, error) {
	return ctx.runRecognition(entry, ctx.handleOverride(override...), img)
}

func (ctx *Context) runAction(
	entry, override string,
	box Rect,
	recognitionDetail string,
) (*ActionDetail, error) {
	rectBuf := buffer.NewRectBuffer()
	rectBuf.Set(box)
	defer rectBuf.Destroy()

	actId := native.MaaContextRunAction(
		ctx.handle,
		entry,
		override,
		rectBuf.Handle(),
		recognitionDetail,
	)
	if actId == 0 {
		return nil, errors.New("failed to run action")
	}
	tasker := ctx.GetTasker()
	actionDetail, err := tasker.getActionDetail(actId)
	return actionDetail, err
}

// RunAction runs an action and returns its detail.
// It accepts an entry string and an optional override parameter which can be
// a JSON string or any data type that can be marshaled to JSON.
// If multiple overrides are provided, only the first one will be used.
//
// Example 1:
//
//	ctx.RunAction("Task", box, recognitionDetail, `{"Task":{"action":"Click","target":[100, 200, 100, 100]}}`)
//
// Example 2:
//
//	ctx.RunAction("Task", box, recognitionDetail, map[string]interface{}{
//	    "Task": map[string]interface{}{
//	        "action": "Click",
//	        "target": []int{100, 200, 100, 100},
//		}
//	})
func (ctx *Context) RunAction(
	entry string,
	box Rect,
	recognitionDetail string,
	override ...any,
) (*ActionDetail, error) {
	return ctx.runAction(
		entry,
		ctx.handleOverride(override...),
		box,
		recognitionDetail,
	)
}

// RunRecognitionDirect runs recognition directly with type and parameters, without requiring a pipeline entry.
// It accepts a recognition type string (e.g., "OCR", "TemplateMatch"), recognition parameters as JSON,
// and an image to recognize.
func (ctx *Context) RunRecognitionDirect(
	recoType NodeRecognitionType,
	recoParam NodeRecognitionParam,
	img image.Image,
) (*RecognitionDetail, error) {
	imgBuf := buffer.NewImageBuffer()
	imgBuf.Set(img)
	defer imgBuf.Destroy()

	recParamJSON, err := json.Marshal(recoParam)
	if err != nil {
		return nil, err
	}

	recId := native.MaaContextRunRecognitionDirect(
		ctx.handle,
		string(recoType),
		string(recParamJSON),
		imgBuf.Handle(),
	)
	if recId == 0 {
		return nil, errors.New("failed to run recognition direct")
	}
	tasker := ctx.GetTasker()
	recognitionDetail, err := tasker.getRecognitionDetail(recId)
	return recognitionDetail, err
}

// RunActionDirect runs action directly with type and parameters, without requiring a pipeline entry.
// It accepts an action type string (e.g., "Click", "Swipe"), action parameters as JSON,
// a box for the action position, and recognition details.
func (ctx *Context) RunActionDirect(
	actionType NodeActionType,
	actionParam NodeActionParam,
	box Rect,
	recoDetail *RecognitionDetail,
) (*ActionDetail, error) {
	rectBuf := buffer.NewRectBuffer()
	rectBuf.Set(box)
	defer rectBuf.Destroy()

	actParamJSON, err := json.Marshal(actionParam)
	if err != nil {
		return nil, err
	}
	recoDetailJSON, err := json.Marshal(recoDetail)
	if err != nil {
		return nil, err
	}

	actId := native.MaaContextRunActionDirect(
		ctx.handle,
		string(actionType),
		string(actParamJSON),
		rectBuf.Handle(),
		string(recoDetailJSON),
	)
	if actId == 0 {
		return nil, errors.New("failed to run action direct")
	}
	tasker := ctx.GetTasker()
	actionDetail, err := tasker.getActionDetail(actId)
	return actionDetail, err
}

func (ctx *Context) overridePipeline(override string) error {
	if !native.MaaContextOverridePipeline(ctx.handle, override) {
		return errors.New("failed to override pipeline")
	}
	return nil
}

// OverridePipeline overrides pipeline.
// The `override` parameter can be a JSON string or any data type that can be marshaled to JSON.
func (ctx *Context) OverridePipeline(override any) error {
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
			return err
		}
		return ctx.overridePipeline(string(jsonBytes))
	}
}

// OverrideNext overrides the next list of task by name.
func (ctx *Context) OverrideNext(name string, nextList []string) error {
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
	if !native.MaaContextOverrideNext(ctx.handle, name, list.Handle()) {
		return errors.New("failed to override next")
	}
	return nil
}

func (ctx *Context) OverrideImage(imageName string, image image.Image) error {
	img := buffer.NewImageBuffer()
	defer img.Destroy()
	img.Set(image)
	if !native.MaaContextOverrideImage(ctx.handle, imageName, img.Handle()) {
		return errors.New("failed to override image")
	}
	return nil
}

// GetNodeJSON gets the node JSON by name.
func (ctx *Context) GetNodeJSON(name string) (string, error) {
	buf := buffer.NewStringBuffer()
	defer buf.Destroy()
	ok := native.MaaContextGetNodeData(ctx.handle, name, buf.Handle())
	if !ok {
		return "", errors.New("failed to get node JSON")
	}
	return buf.Get(), nil
}

func (ctx *Context) GetNodeData(name string) (*Node, error) {
	raw, err := ctx.GetNodeJSON(name)
	if err != nil {
		return nil, err
	}

	var node Node
	err = json.Unmarshal([]byte(raw), &node)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// GetTaskJob returns current task job.
func (ctx *Context) GetTaskJob() *TaskJob {
	tasker := ctx.GetTasker()
	taskId := native.MaaContextGetTaskId(ctx.handle)
	return newTaskJob(
		taskId,
		tasker.status,
		tasker.wait,
		tasker.getTaskDetail,
		tasker.overridePipeline,
		nil,
	)
}

// GetTasker returns the current Tasker.
func (ctx *Context) GetTasker() *Tasker {
	handle := native.MaaContextGetTasker(ctx.handle)
	return &Tasker{handle: handle}
}

// WaitFreezes waits until the screen stabilizes (no significant changes).
// duration: The duration that the screen must remain stable.
// box: The recognition hit box, used when target is "Self" to calculate the ROI. If nil, uses entire screen.
// waitFreezesParam: Additional wait_freezes parameters. Can be a JSON string or any data type that can be marshaled to JSON.
// Returns true if the screen stabilized within the timeout, false otherwise.
func (ctx *Context) WaitFreezes(duration time.Duration, box *Rect, waitFreezesParam ...any) bool {
	var boxHandle uintptr
	if box != nil {
		rectBuf := buffer.NewRectBuffer()
		rectBuf.Set(*box)
		defer rectBuf.Destroy()
		boxHandle = rectBuf.Handle()
	}

	return native.MaaContextWaitFreezes(ctx.handle, uint64(duration.Milliseconds()), boxHandle, ctx.handleOverride(waitFreezesParam...))
}

// Clone clones current Context.
func (ctx *Context) Clone() *Context {
	handle := native.MaaContextClone(ctx.handle)
	return &Context{handle: handle}
}

// SetAnchor sets an anchor by name.
func (ctx *Context) SetAnchor(anchorName, nodeName string) error {
	if !native.MaaContextSetAnchor(ctx.handle, anchorName, nodeName) {
		return errors.New("failed to set anchor")
	}
	return nil
}

// GetAnchor gets an anchor by name.
func (ctx *Context) GetAnchor(anchorName string) (string, error) {
	buf := buffer.NewStringBuffer()
	defer buf.Destroy()
	ok := native.MaaContextGetAnchor(ctx.handle, anchorName, buf.Handle())
	if !ok {
		return "", errors.New("failed to get anchor")
	}
	return buf.Get(), nil
}

// GetHitCount gets the hit count of a node by name.
func (ctx *Context) GetHitCount(nodeName string) (uint64, error) {
	var count uint64
	ok := native.MaaContextGetHitCount(ctx.handle, nodeName, &count)
	if !ok {
		return 0, errors.New("failed to get hit count")
	}
	return count, nil
}

// ClearHitCount clears the hit count of a node by name.
func (ctx *Context) ClearHitCount(nodeName string) error {
	if !native.MaaContextClearHitCount(ctx.handle, nodeName) {
		return errors.New("failed to clear hit count")
	}
	return nil
}
