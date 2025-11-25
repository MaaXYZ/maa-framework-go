package test

import (
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v3"
	"github.com/stretchr/testify/require"
)

func TestRunWithoutFile(t *testing.T) {
	testingPath := "./data_set/PipelineSmoking/Screenshot"

	ctrl := maa.NewCarouselImageController(testingPath)
	require.NotNil(t, ctrl)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res := maa.NewResource()
	require.NotNil(t, res)
	defer res.Destroy()

	tasker := maa.NewTasker()
	require.NotNil(t, tasker)
	defer tasker.Destroy()
	isResBound := tasker.BindResource(res)
	require.True(t, isResBound)
	isCtrlBound := tasker.BindController(ctrl)
	require.True(t, isCtrlBound)

	ok := res.RegisterCustomAction("MyAct", &MyAct{t})
	require.True(t, ok)

	taskParam := maa.J{
		"MyTask": maa.J{
			"action":              "Custom",
			"custom_action":       "MyAct",
			"custom_action_param": "abcdefg",
		},
	}

	got := tasker.PostTask("MyTask", taskParam).Wait().Success()
	require.True(t, got)
}

type MyAct struct {
	t *testing.T
}

func (a *MyAct) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	tasker := ctx.GetTasker()
	require.NotNil(a.t, tasker)
	ctrl := tasker.GetController()
	require.NotNil(a.t, ctrl)
	img := ctrl.CacheImage()
	require.NotNil(a.t, img)

	override := maa.J{
		"MyColorMatching": maa.J{
			"recognition": "ColorMatch",
			"lower":       []int{100, 100, 100},
			"upper":       []int{255, 255, 255},
		},
	}
	detail := ctx.RunRecognition("MyColorMatching", img, override)
	require.NotNil(a.t, detail)

	return true
}
