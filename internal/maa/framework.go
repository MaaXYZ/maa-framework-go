package maa

import (
	"unsafe"

	"github.com/ebitengine/purego"
)

var (
	MaaVersion func() string
)

type MaaNotificationCallback func(message, detailsJson *byte, notifyTransArg unsafe.Pointer) uintptr

type MaaTaskerOption int32

var (
	MaaTaskerCreate               func(notify MaaNotificationCallback, notifyTransArg unsafe.Pointer) uintptr
	MaaTaskerDestroy              func(tasker uintptr)
	MaaTaskerSetOption            func(tasker uintptr, key MaaTaskerOption, value unsafe.Pointer, valSize uint64) bool
	MaaTaskerBindResource         func(tasker uintptr, res uintptr) bool
	MaaTaskerBindController       func(tasker uintptr, ctrl uintptr) bool
	MaaTaskerInited               func(tasker uintptr) bool
	MaaTaskerPostPipeline         func(tasker uintptr, entry, pipelineOverride string) int64
	MaaTaskerStatus               func(tasker uintptr, id int64) int32
	MaaTaskerWait                 func(tasker uintptr, id int64) int32
	MaaTaskerRunning              func(tasker uintptr) bool
	MaaTaskerPostStop             func(tasker uintptr) bool
	MaaTaskerGetResource          func(tasker uintptr) uintptr
	MaaTaskerGetController        func(tasker uintptr) uintptr
	MaaTaskerClearCache           func(tasker uintptr) bool
	MaaTaskerGetRecognitionDetail func(tasker uintptr, recoId int64, name uintptr, algorithm uintptr, hit *bool, box uintptr, detailJson uintptr, raw uintptr, draws uintptr) bool
	MaaTaskerGetNodeDetail        func(tasker uintptr, nodeId int64, name uintptr, recoId *int64, completed *bool) bool
	MaaTaskerGetTaskDetail        func(tasker uintptr, taskId int64, entry uintptr, nodeIdList uintptr, nodeIdListSize *uint64, status *int32) bool
	MaaTaskerGetLatestNode        func(tasker uintptr, taskName string, latestId *int64) bool
)

func init() {
	maaFramework, err := openLibrary(getMaaFrameworkLibrary())
	if err != nil {
		panic(err)
	}

	purego.RegisterLibFunc(&MaaVersion, maaFramework, "MaaVersion")
	// Tasker
	purego.RegisterLibFunc(&MaaTaskerCreate, maaFramework, "MaaTaskerCreate")
	purego.RegisterLibFunc(&MaaTaskerDestroy, maaFramework, "MaaTaskerDestroy")
	purego.RegisterLibFunc(&MaaTaskerSetOption, maaFramework, "MaaTaskerSetOption")
	purego.RegisterLibFunc(&MaaTaskerBindResource, maaFramework, "MaaTaskerBindResource")
	purego.RegisterLibFunc(&MaaTaskerBindController, maaFramework, "MaaTaskerBindController")
	purego.RegisterLibFunc(&MaaTaskerInited, maaFramework, "MaaTaskerInited")
	purego.RegisterLibFunc(&MaaTaskerPostPipeline, maaFramework, "MaaTaskerPostPipeline")
	purego.RegisterLibFunc(&MaaTaskerStatus, maaFramework, "MaaTaskerStatus")
	purego.RegisterLibFunc(&MaaTaskerWait, maaFramework, "MaaTaskerWait")
	purego.RegisterLibFunc(&MaaTaskerRunning, maaFramework, "MaaTaskerRunning")
	purego.RegisterLibFunc(&MaaTaskerPostStop, maaFramework, "MaaTaskerPostStop")
	purego.RegisterLibFunc(&MaaTaskerGetResource, maaFramework, "MaaTaskerGetResource")
	purego.RegisterLibFunc(&MaaTaskerGetController, maaFramework, "MaaTaskerGetController")
	purego.RegisterLibFunc(&MaaTaskerClearCache, maaFramework, "MaaTaskerClearCache")
	purego.RegisterLibFunc(&MaaTaskerGetRecognitionDetail, maaFramework, "MaaTaskerGetRecognitionDetail")
	purego.RegisterLibFunc(&MaaTaskerGetNodeDetail, maaFramework, "MaaTaskerGetNodeDetail")
	purego.RegisterLibFunc(&MaaTaskerGetTaskDetail, maaFramework, "MaaTaskerGetTaskDetail")
	purego.RegisterLibFunc(&MaaTaskerGetLatestNode, maaFramework, "MaaTaskerGetLatestNode")

}
