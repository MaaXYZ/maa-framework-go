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

	pipeline := maa.NewPipeline()
	myTaskNode := maa.NewNode("MyTask",
		maa.WithAction(maa.ActCustom("MyAct",
			maa.WithCustomActionParam("abcdefg"),
		)),
	)
	pipeline.AddNode(myTaskNode)

	got := tasker.PostTask("MyTask", pipeline).Wait().Success()
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

	pipeline := maa.NewPipeline()
	myColorMatchingNode := maa.NewNode("MyColorMatching",
		maa.WithRecognition(maa.RecColorMatch(
			[][]int{{100, 100, 100}},
			[][]int{{255, 255, 255}},
		)),
	)
	pipeline.AddNode(myColorMatchingNode)

	detail, err := ctx.RunRecognition("MyColorMatching", img, pipeline)
	require.NoError(a.t, err)
	require.NotNil(a.t, detail)

	return true
}
