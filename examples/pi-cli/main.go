package main

import (
	"github.com/MaaXYZ/maa-framework-go/v2"
)

func main() {
	toolkit := maa.NewToolkit()
	toolkit.RegisterPICustomAction(0, "MyAct", &MyAct{})
	toolkit.RunCli(0, "./resource", "./", false, nil)
}

type MyAct struct{}

func (m MyAct) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	ctx.OverrideNext(arg.CurrentTaskName, []string{"TaskA", "TaskB"})

	img := ctx.GetTasker().GetController().CacheImage()
	ctx.GetTasker().GetController().PostClick(100, 100).Wait()

	ctx.RunRecognition("Cat", img, maa.J{
		"recognition": "OCR",
		"expected":    "cat",
	})
	return true
}
