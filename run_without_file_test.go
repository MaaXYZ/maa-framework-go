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
	image, err := ctx.Screencap()
	if err != nil {
		panic("failed to screencap:" + err.Error())
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
		panic("failed to marshal task param:" + err.Error())
	}

	_, err = ctx.RunRecognition(image, "MyColorMatching", string(taskParamStr))
	if err != nil {
		panic("failed to run recognition:" + err.Error())
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
	ctrlJob := ctrl.PostConnect()
	ctrlJob.Wait()

	res := NewResource(nil)
	defer res.Destroy()

	inst := New(nil)
	defer inst.Destroy()
	inst.BindResource(res)
	inst.BindController(ctrl)

	myAct := NewMyAct()
	defer myAct.Destroy()
	inst.RegisterCustomAction("MyAct", myAct)

	taskParam := map[string]interface{}{
		"MyTask": map[string]interface{}{
			"action":              "Custom",
			"custom_action":       "MyAct",
			"custom_action_param": "abcdefg",
		},
	}
	taskParamStr, err := json.Marshal(taskParam)
	require.NoError(t, err)

	taskJob := inst.PostTask("MyTask", string(taskParamStr))
	got := taskJob.Wait()
	require.True(t, got)
}
