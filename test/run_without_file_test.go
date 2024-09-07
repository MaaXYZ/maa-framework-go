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

	tasker := maa.New(nil)
	defer tasker.Destroy()
	tasker.BindResource(res)
	tasker.BindController(ctrl)

	res.RegisterCustomAction("MyAct", myAct)

	taskParam := map[string]interface{}{
		"MyTask": map[string]interface{}{
			"action":              "Custom",
			"custom_action":       "MyAct",
			"custom_action_param": "abcdefg",
		},
	}
	taskParamStr, err := json.Marshal(taskParam)
	require.NoError(t, err)

	got := tasker.PostPipeline("MyTask", string(taskParamStr)).Wait()
	require.True(t, got)
}

func myAct(ctx *maa.Context, taskId int64, actionName, customActionParam string, box maa.Rect, recognitionDetail string) bool {
	tasker := ctx.GetTasker()
	ctrl := tasker.GetController()
	img, _ := ctrl.CacheImage()

	ppOverride := map[string]interface{}{
		"MyColorMatching": map[string]interface{}{
			"recognition": "ColorMatch",
			"lower":       []int{100, 100, 100},
			"upper":       []int{255, 255, 255},
		},
	}
	ppOverrideStr, err := json.Marshal(ppOverride)
	if err != nil {
		log.Println("failed to marshal task param:" + err.Error())
		return false
	}

	recId := ctx.RunRecognition("MyColorMatching", string(ppOverrideStr), img)
	_, _ = tasker.GetRecognitionDetail(recId)

	return true
}
