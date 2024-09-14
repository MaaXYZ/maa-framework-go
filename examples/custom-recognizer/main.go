package main

import (
	"fmt"
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/MaaXYZ/maa-framework-go/toolkit"
	"image"
	"os"
)

func main() {
	toolkit.ConfigInitOption("./", "{}")
	tasker := maa.NewTasker(nil)
	defer tasker.Destroy()

	deviceFinder := toolkit.NewAdbDeviceFinder()
	deviceFinder.Find()
	device := deviceFinder.Find()[0]
	ctrl := maa.NewAdbController(
		device.GetAdbPath(),
		device.GetAddress(),
		device.GetScreencapMethod(),
		device.GetInputMethod(),
		device.GetConfig(),
		"path/to/MaaAgentBinary",
		nil,
	)
	defer ctrl.Destroy()
	ctrl.PostConnect().Wait()
	tasker.BindController(ctrl)

	res := maa.NewResource(nil)
	defer res.Destroy()
	res.PostPath("./resource").Wait()
	tasker.BindResource(res)
	if tasker.Inited() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	res.RegisterCustomRecognizer("MyRec", &MyRec{})

	detail := tasker.PostPipeline("Startup").Wait().GetDetail()
	fmt.Println(detail)
}

type MyRec struct{}

func (r *MyRec) Run(ctx *maa.Context, _ *maa.TaskDetail, currentTaskName, _, _ string, img image.Image, _ maa.Rect) (maa.CustomRecognizerResult, bool) {
	ctx.RunRecognition("MyCustomOCR", img, maa.J{
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
	newContext.RunPipeline("MyCustomOCR", img)

	clickJob := ctx.GetTasker().GetController().PostClick(10, 20)
	clickJob.Wait()

	ctx.OverrideNext(currentTaskName, []string{"TaskA", "TaskB"})

	return maa.CustomRecognizerResult{
		Box:    maa.Rect{0, 0, 100, 100},
		Detail: "Hello World!",
	}, true
}
