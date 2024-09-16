package main

import (
	"fmt"
	"github.com/MaaXYZ/maa-framework-go"
	"os"
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
	res.PostPath("./resource").Wait()
	tasker.BindResource(res)
	if tasker.Inited() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	res.RegisterCustomAction("MyAct", &MyAct{})

	detail := tasker.PostPipeline("Startup").Wait().GetDetail()
	fmt.Println(detail)
}

type MyAct struct{}

func (a *MyAct) Run(_ *maa.Context, arg *maa.CustomActionArg) bool {
	return true
}
