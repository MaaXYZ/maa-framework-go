package main

import (
	"fmt"
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/MaaXYZ/maa-framework-go/toolkit"
	"image"
	"os"
)

func main() {
	toolkit.InitOption("./", "{}")
	inst := maa.New(nil)
	defer inst.Destroy()

	devices := toolkit.AdbDevices()
	device := devices[0]
	ctrl := maa.NewAdbController(
		device.AdbPath,
		device.Address,
		device.ControllerType,
		device.Config,
		"path/to/MaaAgentBinary",
		nil,
	)
	defer ctrl.Destroy()
	ctrl.PostConnect().Wait()
	inst.BindController(ctrl)

	res := maa.NewResource(nil)
	defer res.Destroy()
	res.PostPath("./resource").Wait()
	inst.BindResource(res)
	if inst.Inited() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	myRec := NewMyRec()
	defer myRec.Destroy()
	inst.RegisterCustomRecognizer("MyRec", myRec)

	inst.PostTask("Startup", "{}")
}

type MyRec struct {
	maa.CustomRecognizerHandler
}

func NewMyRec() maa.CustomRecognizer {
	return &MyRec{
		CustomRecognizerHandler: maa.NewCustomRecognizerHandler(),
	}
}

func (m MyRec) Analyze(syncCtx maa.SyncContext, img image.Image, taskName, RecognitionParam string) (maa.AnalyzeResult, bool) {
	return maa.AnalyzeResult{
		Box:    maa.Rect{0, 0, 100, 100},
		Detail: "Hello World!",
	}, true
}
