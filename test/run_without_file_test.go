package test

import (
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v3"
	"github.com/stretchr/testify/require"
)

func TestRunWithoutFile(t *testing.T) {
	testingPath := "./data_set/PipelineSmoking/Screenshot"

	ctrl, err := maa.NewCarouselImageController(testingPath)
	require.NoError(t, err)
	require.NotNil(t, ctrl)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)

	res, err := maa.NewResource()
	require.NoError(t, err)
	require.NotNil(t, res)
	defer res.Destroy()

	tasker, err := maa.NewTasker()
	require.NoError(t, err)
	require.NotNil(t, tasker)
	defer tasker.Destroy()
	err = tasker.BindResource(res)
	require.NoError(t, err)
	err = tasker.BindController(ctrl)
	require.NoError(t, err)

	err = res.RegisterCustomAction("MyAct", &MyAct{t})
	require.NoError(t, err)

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
	img, err := ctrl.CacheImage()
	require.NoError(a.t, err)
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
