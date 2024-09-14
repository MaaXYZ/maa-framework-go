package main

import (
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/MaaXYZ/maa-framework-go/toolkit"
)

func main() {
	pi := toolkit.NewProjectInterface()
	pi.RegisterCustomAction(0, "MyAct", &MyAct{})
	pi.RunCli(0, "./resource", "./", false, nil)
}

type MyAct struct{}

func (m MyAct) Run(ctx *maa.Context, _ *maa.TaskDetail, currentTaskName, _, _ string, _ *maa.RecognitionDetail, _ maa.Rect) bool {
	ctx.OverrideNext(currentTaskName, []string{"TaskA", "TaskB"})

	img, _ := ctx.GetTasker().GetController().CacheImage()
	ctx.GetTasker().GetController().PostClick(100, 100).Wait()

	ctx.RunRecognition("Cat", img, maa.J{
		"recognition": "OCR",
		"expected":    "cat",
	})
	return true
}
