package main

import (
	"fmt"
	"os"

	"github.com/MaaXYZ/maa-framework-go/v3"
)

func main() {
	maa.Init()
	maa.ConfigInitOption("./", "{}")
	tasker := maa.NewTasker()
	defer tasker.Destroy()

	device := maa.FindAdbDevices()[0]
	ctrl, err := maa.NewAdbController(
		device.AdbPath,
		device.Address,
		device.ScreencapMethod,
		device.InputMethod,
		device.Config,
		"path/to/MaaAgentBinary",
	)
	if err != nil {
		fmt.Println("Failed to create ADB controller")
		os.Exit(1)
	}
	defer ctrl.Destroy()
	ctrl.PostConnect().Wait()
	tasker.BindController(ctrl)

	res, err := maa.NewResource()
	if err != nil {
		fmt.Println("Failed to create resource:", err)
		os.Exit(1)
	}
	defer res.Destroy()
	res.PostBundle("./resource").Wait()
	tasker.BindResource(res)
	if !tasker.Initialized() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	if err := res.RegisterCustomRecognition("MyRec", &MyRec{}); err != nil {
		fmt.Println("Failed to register custom recognition:", err)
		os.Exit(1)
	}

	detail, err := tasker.PostTask("Startup").Wait().GetDetail()
	if err != nil {
		fmt.Println("Failed to get task detail:", err)
		os.Exit(1)
	}
	fmt.Println(detail)
}

type MyRec struct{}

func (r *MyRec) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	ctx.RunRecognition("MyCustomOCR", arg.Img, map[string]any{
		"MyCustomOCR": map[string]any{
			"roi": []int{100, 100, 200, 300},
		},
	})

	ctx.OverridePipeline(map[string]any{
		"MyCustomOCR": map[string]any{
			"roi": []int{1, 1, 114, 514},
		},
	})

	newContext := ctx.Clone()
	newContext.OverridePipeline(map[string]any{
		"MyCustomOCR": map[string]any{
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
