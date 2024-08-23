package main

import (
	"fmt"
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/MaaXYZ/maa-framework-go/buffer"
	"github.com/MaaXYZ/maa-framework-go/toolkit"
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

	myAct := NewAct()
	defer myAct.Destroy()
	inst.RegisterCustomAction("MyAct", myAct)

	inst.PostTask("Startup", "{}")
}

type MyAct struct {
	maa.CustomActionHandler
}

func NewAct() maa.CustomAction {
	return &MyAct{
		CustomActionHandler: maa.NewCustomActionHandler(),
	}
}

func (*MyAct) Run(ctx maa.SyncContext, taskName, ActionParam string, curBox buffer.Rect, curRecDetail string) bool {
	return true
}

func (*MyAct) Stop() {
}
