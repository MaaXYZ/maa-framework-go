package maa

import (
	"image"
	"sync"
	"sync/atomic"

	"github.com/MaaXYZ/maa-framework-go/v4/internal/buffer"
)

var (
	customRecognitionRunnerCallbackID          uint64
	customRecognitionRunnerCallbackAgents      = make(map[uint64]CustomRecognitionRunner)
	customRecognitionRunnerCallbackAgentsMutex sync.RWMutex
)

func registerCustomRecognition(recognizer CustomRecognitionRunner) uint64 {
	id := atomic.AddUint64(&customRecognitionRunnerCallbackID, 1)

	customRecognitionRunnerCallbackAgentsMutex.Lock()
	customRecognitionRunnerCallbackAgents[id] = recognizer
	customRecognitionRunnerCallbackAgentsMutex.Unlock()

	return id
}

func unregisterCustomRecognition(id uint64) bool {
	customRecognitionRunnerCallbackAgentsMutex.Lock()
	defer customRecognitionRunnerCallbackAgentsMutex.Unlock()

	if _, ok := customRecognitionRunnerCallbackAgents[id]; !ok {
		return false
	}
	delete(customRecognitionRunnerCallbackAgents, id)
	return true
}

// CustomRecognitionArg is the argument passed to CustomRecognitionRunner.Run
// when the framework executes a node with a Custom recognition.
type CustomRecognitionArg struct {
	TaskID int64 // Task ID. Task details can be retrieved via Tasker.GetTaskDetail.
	// CurrentTaskName is the name of the pipeline node currently executing.
	CurrentTaskName string
	// CustomRecognitionName is the registered name of this custom recognizer.
	CustomRecognitionName string
	// CustomRecognitionParam is the node's custom_recognition_param,
	// serialized as a JSON string.
	CustomRecognitionParam string
	// Img is the current frame to recognize, decoded into a copy that stays
	// valid after Run returns.
	Img image.Image
	// Roi is the region of interest the recognizer should search, resolved
	// from the node's roi.
	Roi Rect
}

// CustomRecognitionResult contains the box and detail returned by a custom recognizer.
// The result can include diagnostic information even when recognition does not match.
type CustomRecognitionResult struct {
	Box    Rect   `json:"box"`
	Detail string `json:"detail"`
}

// CustomRecognitionRunner performs recognition for a registered custom recognizer.
type CustomRecognitionRunner interface {
	// Run returns a result and whether recognition matched. A non-nil result is
	// passed to MaaFramework even when matched is false. A nil result is always
	// treated as no match.
	Run(ctx *Context, arg *CustomRecognitionArg) (*CustomRecognitionResult, bool)
}

// CustomRecognitionFunc is an adapter to allow use of ordinary functions as CustomRecognitionRunner.
// If f is a function with the appropriate signature, CustomRecognitionFunc(f) is a
// CustomRecognitionRunner that calls f.
type CustomRecognitionFunc func(ctx *Context, arg *CustomRecognitionArg) (*CustomRecognitionResult, bool)

// Run calls f(ctx, arg).
func (f CustomRecognitionFunc) Run(ctx *Context, arg *CustomRecognitionArg) (*CustomRecognitionResult, bool) {
	return f(ctx, arg)
}

func _MaaCustomRecognitionCallbackAgent(
	context uintptr,
	taskId int64,
	currentTaskName, customRecognitionName, customRecognitionParam *byte,
	image, roi uintptr,
	transArg uintptr,
	outBox, outDetail uintptr,
) uintptr {
	// Here, we are simply passing the uint64 value as a pointer
	// and will not actually dereference this pointer.
	id := uint64(transArg)

	customRecognitionRunnerCallbackAgentsMutex.RLock()
	recognition, exists := customRecognitionRunnerCallbackAgents[id]
	customRecognitionRunnerCallbackAgentsMutex.RUnlock()

	if !exists || recognition == nil {
		return 0
	}

	imgBuffer := buffer.NewImageBufferByHandle(image)
	imgImg := imgBuffer.Get()

	ctx := newCallbackContext(context)
	if ctx == nil {
		return 0
	}
	defer ctx.invalidate()
	ret, ok := recognition.Run(
		ctx,
		&CustomRecognitionArg{
			TaskID:                 taskId,
			CurrentTaskName:        cStringToString(currentTaskName),
			CustomRecognitionName:  cStringToString(customRecognitionName),
			CustomRecognitionParam: cStringToString(customRecognitionParam),
			Img:                    imgImg,
			Roi:                    buffer.NewRectBufferByHandle(roi).Get(),
		},
	)
	if ret == nil {
		return 0
	}

	box := ret.Box
	outBoxRect := buffer.NewRectBufferByHandle(outBox)
	outBoxRect.Set(box)
	outDetailString := buffer.NewStringBufferByHandle(outDetail)
	outDetailString.Set(ret.Detail)
	if ok {
		return 1
	}
	return 0
}
