package main

import (
	"fmt"
	"os"

	"github.com/MaaXYZ/maa-framework-go/v2"
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
	if !tasker.Initialized() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	detail := tasker.PostTask("Startup").Wait().GetDetail()
	fmt.Println(detail)
}
