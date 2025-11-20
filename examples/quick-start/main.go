package main

import (
	"fmt"
	"os"

	"github.com/MaaXYZ/maa-framework-go/v2"
)

func main() {
	maa.ConfigInitOption("./", "{}")
	tasker := maa.NewTasker()
	defer tasker.Destroy()

	device := maa.FindAdbDevices()[0]
	ctrl := maa.NewAdbController(
		device.AdbPath,
		device.Address,
		device.ScreencapMethod,
		device.InputMethod,
		device.Config,
		"path/to/MaaAgentBinary",
	)
	defer ctrl.Destroy()
	ctrl.PostConnect().Wait()
	tasker.BindController(ctrl)

	res := maa.NewResource()
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
