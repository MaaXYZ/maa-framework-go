package main

import (
	"fmt"
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/MaaXYZ/maa-framework-go/toolkit"
	"os"
)

func main() {
	toolkit.ConfigInitOption("./", "{}")
	tasker := maa.New(nil)
	defer tasker.Destroy()

	deviceFinder := toolkit.NewAdbDeviceFinder()
	deviceFinder.Find()
	device := deviceFinder.List()[0]
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

	res.RegisterCustomAction("MyAct", myAct)

	tasker.PostPipeline("Startup", "{}")
}

func myAct(_ *maa.Context, _ int64, _, _ string, _ maa.Rect, _ string) bool {
	return true
}
