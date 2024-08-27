package test

import (
	"encoding/json"
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestRunWithoutFile(t *testing.T) {
	testingPath := "./data_set/PipelineSmoking/Screenshot"
	resultPath := "./data_set/debug"

	ctrl := maa.NewDbgController(testingPath, resultPath, maa.DbgControllerTypeCarouselImage, "{}", nil)
	defer ctrl.Destroy()
	ctrl.PostConnect().Wait()

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

	got := inst.PostTask("MyTask", string(taskParamStr)).Wait()
	require.True(t, got)
}

type MyAct struct {
	maa.CustomActionHandler
}

func NewMyAct() maa.CustomAction {
	return &MyAct{
		CustomActionHandler: maa.NewCustomActionHandler(),
	}
}

func (act *MyAct) Run(ctx maa.SyncContext, _, _ string, _ maa.Rect, _ string) bool {
	image, err := ctx.Screencap()
	if err != nil {
		log.Println("failed to screencap:" + err.Error())
		return false
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
		log.Println("failed to marshal task param:" + err.Error())
		return false
	}

	_, err = ctx.RunRecognition(image, "MyColorMatching", string(taskParamStr))
	if err != nil {
		log.Println("failed to run recognition:" + err.Error())
		return false
	}

	return true
}

func (act *MyAct) Stop() {
	// do nothing
}
