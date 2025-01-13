package main

import (
	"fmt"
	"os"

	"github.com/MaaXYZ/maa-framework-go"
)

func main() {
	toolkit := maa.NewToolkit()
	toolkit.ConfigInitOption("./", "{}")
	tasker := maa.NewTasker(nil)
	defer tasker.Destroy()

	device := toolkit.FindAdbDevices()[0]
	ctrl := maa.NewAdbController(
		device.AdbPath,
		device.Address,
		device.ScreencapMethod,
		device.InputMethod,
		device.Config,
		"path/to/MaaAgentBinary",
		nil,
	)
	defer ctrl.Destroy()
	ctrl.PostConnect().Wait()
	tasker.BindController(ctrl)

	res := maa.NewResource(nil)
	defer res.Destroy()
	res.PostBundle("./resource").Wait()
	tasker.BindResource(res)
	if tasker.Initialized() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	res.RegisterCustomRecognition("MyRec", &MyRec{})

	detail := tasker.PostTask("Startup").Wait().GetDetail()
	fmt.Println(detail)
}

type MyRec struct{}

func (r *MyRec) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	ctx.RunRecognition("MyCustomOCR", arg.Img, maa.J{
		"MyCustomOCR": maa.J{
			"roi": []int{100, 100, 200, 300},
		},
	})

	ctx.OverridePipeline(maa.J{
		"MyCustomOCR": maa.J{
			"roi": []int{1, 1, 114, 514},
		},
	})

	newContext := ctx.Clone()
	newContext.OverridePipeline(maa.J{
		"MyCustomOCR": maa.J{
			"roi": []int{100, 200, 300, 400},
		},
	})
	newContext.RunTask("MyCustomOCR", arg.Img)

	clickJob := ctx.GetTasker().GetController().PostClick(10, 20)
	clickJob.Wait()

	ctx.OverrideNext(arg.CurrentTaskName, []string{"TaskA", "TaskB"})

	return &maa.CustomRecognitionResult{
		Box:    maa.Rect{0, 0, 100, 100},
		Detail: "Hello World!",
	}, true
}
