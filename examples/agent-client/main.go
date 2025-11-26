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

	res := maa.NewResource()
	defer res.Destroy()

	if !tasker.BindResource(res) {
		fmt.Println("Failed to bind resource to MAA Tasker")
		os.Exit(1)
	}

	ctrl := maa.NewBlankController()
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

	client := maa.NewAgentClient("7788")
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
