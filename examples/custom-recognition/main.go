package main

import (
	"fmt"
	"os"

	"github.com/MaaXYZ/maa-framework-go/v4"
)

func main() {
	maa.Init()
	if err := maa.ConfigInitOption("./", "{}"); err != nil {
		fmt.Println("Failed to init config:", err)
		os.Exit(1)
	}
	tasker, err := maa.NewTasker()
	if err != nil {
		fmt.Println("Failed to create tasker")
		os.Exit(1)
	}

	devices, err := maa.FindAdbDevices()
	if err != nil {
		fmt.Println("Failed to find adb devices:", err)
		os.Exit(1)
	}
	device := devices[0]
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
	connectJob, err := ctrl.PostConnect()
	if err != nil {
		fmt.Println("Failed to post connect:", err)
		os.Exit(1)
	}
	connectJob.Wait()
	tasker.BindController(ctrl)

	res, err := maa.NewResource()
	if err != nil {
		fmt.Println("Failed to create resource:", err)
		os.Exit(1)
	}
	defer res.Destroy()
	bundleJob, err := res.PostBundle("./resource")
	if err != nil {
		fmt.Println("Failed to post bundle:", err)
		os.Exit(1)
	}
	bundleJob.Wait()
	tasker.BindResource(res)
	defer tasker.Destroy()
	if !tasker.Initialized() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	if err := res.RegisterCustomRecognition("MyRec", &MyRec{}); err != nil {
		fmt.Println("Failed to register custom recognition:", err)
		os.Exit(1)
	}

	taskJob, err := tasker.PostTask("Startup")
	if err != nil {
		fmt.Println("Failed to post task:", err)
		os.Exit(1)
	}
	detail, err := taskJob.Wait().GetDetail()
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

	clickJob, err := ctx.GetTasker().GetController().PostClick(10, 20)
	if err != nil {
		fmt.Println("Failed to post click:", err)
		os.Exit(1)
	}
	clickJob.Wait()

	ctx.OverrideNext(arg.CurrentTaskName, []maa.NextItem{
		{Name: "TaskA"},
		{Name: "TaskB"},
	})

	return &maa.CustomRecognitionResult{
		Box:    maa.Rect{0, 0, 100, 100},
		Detail: "Hello World!",
	}, true
}
