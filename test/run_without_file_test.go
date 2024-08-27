package test

import (
	"encoding/json"
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/MaaXYZ/maa-framework-go/buffer"
	"github.com/stretchr/testify/require"
	"testing"
)

type MyAct struct {
	maa.CustomActionHandler
}

func (act *MyAct) Run(ctx maa.SyncContext, taskName, ActionParam string, curBox buffer.Rect, curRecDetail string) bool {
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

func (act *MyAct) Stop() {
	// do nothing
}

func NewMyAct() maa.CustomAction {
	return &MyAct{
		CustomActionHandler: maa.NewCustomActionHandler(),
	}
}

func TestRunWithoutFile(t *testing.T) {
	testingPath := "./data_set/PipelineSmoking/Screenshot"
	resultPath := "./data_set/debug"

	ctrl := maa.NewDbgController(testingPath, resultPath, maa.DbgControllerTypeCarouselImage, "{}", nil)
	defer ctrl.Destroy()
	ctrlJob := ctrl.PostConnect()
	ctrlJob.Wait()

	res := maa.NewResource(nil)
	defer res.Destroy()

	inst := maa.New(nil)
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
