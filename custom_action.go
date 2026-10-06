package maa

import (
	"errors"
	"sync"
	"sync/atomic"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/buffer"
)

var (
	customActionRunnerCallbackID          uint64
	customActionRunnerCallbackAgents      = make(map[uint64]CustomActionRunner)
	customActionRunnerCallbackAgentsMutex sync.RWMutex
)

func registerCustomAction(action CustomActionRunner) (uint64, error) {
	if isNilCustomActionRunner(action) {
		return 0, errors.New("custom action runner is nil")
	}

	id := atomic.AddUint64(&customActionRunnerCallbackID, 1)

	customActionRunnerCallbackAgentsMutex.Lock()
	customActionRunnerCallbackAgents[id] = action
	customActionRunnerCallbackAgentsMutex.Unlock()

	return id, nil
}

func isNilCustomActionRunner(action CustomActionRunner) bool {
	if action == nil {
		return true
	}
	// A typed-nil CustomActionFunc is a non-nil interface whose callback
	// would panic; upstream rejects null callbacks at registration.
	if f, ok := action.(CustomActionFunc); ok {
		return f == nil
	}
	return false
}

func unregisterCustomAction(id uint64) bool {
	customActionRunnerCallbackAgentsMutex.Lock()
	defer customActionRunnerCallbackAgentsMutex.Unlock()

	if _, ok := customActionRunnerCallbackAgents[id]; !ok {
		return false
	}
	delete(customActionRunnerCallbackAgents, id)
	return true
}

// CustomActionArg is the argument passed to CustomActionRunner.Run when the
// framework executes a node with a Custom action.
type CustomActionArg struct {
	TaskID int64 // Task ID. Task details can be retrieved via Tasker.GetTaskDetail.
	// CurrentTaskName is the name of the node currently executing. Entries
	// run via Context.RunActionDirect use a synthesized name
	// ("action/<type>/<uuid>") instead of a pipeline node name.
	CurrentTaskName string
	// CustomActionName is the registered name of this custom action.
	CustomActionName string
	// CustomActionParam is the node's custom_action_param, serialized as a
	// JSON string.
	CustomActionParam string
	// RecognitionDetail may be nil when the custom action runs on an action-only
	// node (e.g. invoked via Context.RunAction), where reco_id is invalid.
	RecognitionDetail *RecognitionDetail
	// Box is the action's resolved target rect, the position the action
	// should act on. It equals the preceding recognition's hit box only when
	// the action target is Self (the default); with other targets it is the
	// resolved target rect instead, and it may be empty when that rect is
	// empty.
	Box Rect
}

// CustomActionRunner performs the action for nodes registered under a name
// via Resource.RegisterCustomAction.
type CustomActionRunner interface {
	// Run reports whether the action succeeded. Returning false marks the
	// node's action as failed; for a pipeline node the task then continues
	// from the node's on_error list, while action-only runs (e.g. via
	// Context.RunAction) only report the failure in their ActionDetail.
	Run(ctx *Context, arg *CustomActionArg) bool
}

// CustomActionFunc is an adapter to allow use of ordinary functions as CustomActionRunner.
// If f is a function with the appropriate signature, CustomActionFunc(f) is a
// CustomActionRunner that calls f.
type CustomActionFunc func(ctx *Context, arg *CustomActionArg) bool

// Run calls f(ctx, arg).
func (f CustomActionFunc) Run(ctx *Context, arg *CustomActionArg) bool {
	return f(ctx, arg)
}

func _MaaCustomActionCallbackAgent(
	context uintptr,
	taskId int64,
	currentTaskName, customActionName, customActionParam *byte,
	recoId int64,
	box uintptr,
	transArg uintptr,
) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(transArg)

	customActionRunnerCallbackAgentsMutex.RLock()
	action, exists := customActionRunnerCallbackAgents[id]
	customActionRunnerCallbackAgentsMutex.RUnlock()

	if !exists || action == nil {
		return 0
	}

	ctx := newCallbackContext(context)
	if ctx == nil {
		return 0
	}
	defer ctx.invalidate()
	tasker := ctx.GetTasker()
	if tasker == nil {
		return 0
	}
	// Skip GetRecognitionDetail for invalid recoId to avoid a spurious framework error log.
	var recognitionDetail *RecognitionDetail
	if recoId != 0 {
		var err error
		recognitionDetail, err = tasker.GetRecognitionDetail(recoId)
		if err != nil {
			return 0
		}
	}
	curBoxRectBuffer := buffer.NewRectBufferByHandle(box)

	ok := action.Run(
		ctx,
		&CustomActionArg{
			TaskID:            taskId,
			CurrentTaskName:   cStringToString(currentTaskName),
			CustomActionName:  cStringToString(customActionName),
			CustomActionParam: cStringToString(customActionParam),
			RecognitionDetail: recognitionDetail,
			Box:               curBoxRectBuffer.Get(),
		},
	)
	if ok {
		return 1
	}
	return 0
}
