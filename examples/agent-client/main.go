package main

import (
	"fmt"
	"os"

	"github.com/MaaXYZ/maa-framework-go/v3"
)

func main() {
	maa.Init()

	tasker := maa.NewTasker()
	defer tasker.Destroy()

	res, err := maa.NewResource()
	if err != nil {
		fmt.Println("Failed to create resource")
		os.Exit(1)
	}
	defer res.Destroy()

	if !tasker.BindResource(res) {
		fmt.Println("Failed to bind resource to MAA Tasker")
		os.Exit(1)
	}

	ctrl, err := maa.NewBlankController()
	if err != nil {
		fmt.Println("Failed to create blank controller")
		os.Exit(1)
	}
	defer ctrl.Destroy()

	ctrl.PostConnect().Wait()

	if !tasker.BindController(ctrl) {
		fmt.Println("Failed to bind controller to MAA Tasker")
		os.Exit(1)
	}

	if !tasker.Initialized() {
		fmt.Println("Failed to init MAA Tasker")
		os.Exit(1)
	}

	client, err := maa.NewAgentClient(maa.WithTcpPort(7788))
	if err != nil {
		fmt.Println("Failed to create agent client")
		os.Exit(1)
	}
	defer client.Destroy()

	client.BindResource(res)

	client.Connect()

	tasker.PostTask("Test", map[string]any{
		"Test": map[string]any{
			"action":        "Custom",
			"custom_action": "TestAgentServer",
		},
	}).Wait()

	client.Disconnect()

}
