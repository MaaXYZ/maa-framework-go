package main

import (
	"fmt"
	"os"

	"github.com/MaaXYZ/maa-framework-go/v4"
)

func main() {
	maa.Init()

	socketID := os.Args[1]

	if err := maa.AgentServerRegisterCustomAction("TestAgentServer", NewAgentServerAction()); err != nil {
		fmt.Println(err)
		return
	}

	if err := maa.AgentServerStartUp(socketID); err != nil {
		fmt.Println(err)
		return
	}

	maa.AgentServerJoin()

	maa.AgentServerShutDown()
}

type AgentServerAction struct{}

// Run implements maa.CustomAction.
func (a *AgentServerAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {

	fmt.Println("Agent server custom action is running")

	return true
}

func NewAgentServerAction() maa.CustomActionRunner {
	return &AgentServerAction{}
}
