package maa

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

type MyAct struct {
	CustomActionHandler
}

func (act MyAct) Run(ctx SyncContext, taskName, ActionParam string, curBox Rect, curRecDetail string) bool {
	image, ok := ctx.Screencap()
	defer image.Destroy()
	if !ok {
		panic("failed to screencap")
	}

	taskParam := map[string]interface{}{
		"MyColorMatching": map[string]interface{}{
			"recognition": "ColorMatch",
			"lower":       []int{100, 100, 100},
			"upper":       []int{255, 255, 255},
		},
	}
	taskParamStr, err := json.Marshal(taskParam)
	if err != nil {
		panic(err)
	}

	_, ok = ctx.RunRecognition(image, "MyColorMatching", string(taskParamStr))
	if !ok {
		panic("failed to run recognition")
	}

	return true
}

func (act MyAct) Stop() {
	// do nothing
}

func NewMyAct() MyAct {
	return MyAct{
		CustomActionHandler: NewCustomActionHandler(),
	}
}

func TestRunWithoutFile(t *testing.T) {
	testingPath := "./TestingDataSet/PipelineSmoking/Screenshot"
	resultPath := "./TestingDataSet/debug"

	ctrl := NewDbgController(testingPath, resultPath, DbgControllerTypeCarouselImage, "{}", nil)
	defer ctrl.Destroy()
	ctrlId := ctrl.PostConnect()
	ctrl.Wait(ctrlId)

	res := NewResource(nil)
	defer res.Destroy()

	inst := New(nil)
	defer inst.Destroy()
	inst.BindResource(res)
	inst.BindController(ctrl)

	inst.RegisterCustomAction("MyAct", NewMyAct())

	taskParam := map[string]interface{}{
		"MyTask": map[string]interface{}{
			"action":              "Custom",
			"custom_action":       "MyAct",
			"custom_action_param": "abcdefg",
		},
	}
	taskParamStr, err := json.Marshal(taskParam)
	require.NoError(t, err)

	taskId := inst.PostTask("MyTask", string(taskParamStr))
	status := inst.WaitTask(taskId)

	require.Equal(t, status, Success)
}
