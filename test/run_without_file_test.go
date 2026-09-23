package test

import (
	"fmt"
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/stretchr/testify/require"
)

func TestRunWithoutFile(t *testing.T) {
	ctrl, err := maa.NewBlankController()
	require.NoError(t, err)
	require.NotNil(t, ctrl)
	defer ctrl.Destroy()
	isConnected := ctrl.PostConnect().Wait().Success()
	require.True(t, isConnected)
	require.True(t, ctrl.PostScreencap().Wait().Success())

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

	resultsCh := make(chan error, 1)
	err = res.RegisterCustomAction("MyAct", &MyAct{resultsCh})
	require.NoError(t, err)

	pipeline := maa.NewPipeline()
	myTaskNode := maa.NewNode("MyTask").
		SetAction(maa.ActCustom(maa.CustomActionParam{
			CustomAction:      "MyAct",
			CustomActionParam: "abcdefg",
		}))
	pipeline.AddNode(myTaskNode)

	got := tasker.PostTask("MyTask", pipeline).Wait().Success()
	select {
	case callbackErr := <-resultsCh:
		require.NoError(t, callbackErr)
	default:
		t.Fatal("custom action callback was not called")
	}
	require.True(t, got)
}

type MyAct struct {
	results chan error
}

func (a *MyAct) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	err := a.run(ctx)
	a.results <- err
	return err == nil
}

func (a *MyAct) run(ctx *maa.Context) error {
	tasker := ctx.GetTasker()
	if tasker == nil {
		return fmt.Errorf("context has no tasker")
	}
	ctrl := tasker.GetController()
	if ctrl == nil {
		return fmt.Errorf("tasker has no controller")
	}
	img, err := ctrl.CacheImage()
	if err != nil {
		return fmt.Errorf("get cached image: %w", err)
	}
	if img == nil {
		return fmt.Errorf("cached image is nil")
	}

	pipeline := maa.NewPipeline()
	myColorMatchingNode := maa.NewNode("MyColorMatching").
		SetRecognition(maa.RecColorMatch(maa.ColorMatchParam{
			Lower: [][]int{{100, 100, 100}},
			Upper: [][]int{{255, 255, 255}},
		}))
	pipeline.AddNode(myColorMatchingNode)

	detail, err := ctx.RunRecognition("MyColorMatching", img, pipeline)
	if err != nil {
		return fmt.Errorf("run recognition: %w", err)
	}
	if detail == nil {
		return fmt.Errorf("recognition detail is nil")
	}

	return nil
}
