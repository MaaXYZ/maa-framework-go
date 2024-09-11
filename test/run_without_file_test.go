package test

import (
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/stretchr/testify/require"
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

	tasker := maa.NewTasker(nil)
	defer tasker.Destroy()
	tasker.BindResource(res)
	tasker.BindController(ctrl)

	res.RegisterCustomAction("MyAct", &MyAct{})

	taskParam := map[string]interface{}{
		"MyTask": map[string]interface{}{
			"action":              "Custom",
			"custom_action":       "MyAct",
			"custom_action_param": "abcdefg",
		},
	}

	got := tasker.PostPipeline("MyTask", taskParam).Wait()
	require.True(t, got)
}

type MyAct struct{}

func (a *MyAct) Run(ctx *maa.Context, _ int64, _, _ string, _ maa.Rect, _ string) bool {
	tasker := ctx.GetTasker()
	ctrl := tasker.GetController()
	img, _ := ctrl.CacheImage()

	override := maa.J{
		"MyColorMatching": maa.J{
			"recognition": "ColorMatch",
			"lower":       []int{100, 100, 100},
			"upper":       []int{255, 255, 255},
		},
	}

	_ = ctx.RunRecognition("MyColorMatching", img, override)

	return true
}
