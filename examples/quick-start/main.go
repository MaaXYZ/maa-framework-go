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

	tasker.PostPipeline("Startup", "{}")
}
