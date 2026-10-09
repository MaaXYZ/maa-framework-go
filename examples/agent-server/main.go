package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/MaaXYZ/maa-framework-go/v4"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (err error) {
	if len(os.Args) != 2 || os.Args[1] == "" {
		return fmt.Errorf("usage: %s <agent-client-identifier>", os.Args[0])
	}
	identifier := os.Args[1]

	if err := maa.Init(); err != nil {
		return fmt.Errorf("init MAA: %w", err)
	}
	defer func() {
		if releaseErr := maa.Release(); releaseErr != nil {
			err = errors.Join(err, fmt.Errorf("release MAA: %w", releaseErr))
		}
	}()
	// Close the server's sockets even if registration or startup fails.
	defer maa.AgentServerShutDown()

	if err := maa.AgentServerRegisterCustomAction("TestAgentServer", NewAgentServerAction()); err != nil {
		return fmt.Errorf("register agent server custom action: %w", err)
	}
	if err := maa.AgentServerStartUp(identifier); err != nil {
		return fmt.Errorf("start agent server: %w", err)
	}

	// The client disconnects after its task finishes, allowing Join to return.
	maa.AgentServerJoin()
	return nil
}

// AgentServerAction demonstrates an action executed by the Agent server.
type AgentServerAction struct{}

// Run implements maa.CustomActionRunner.
func (a *AgentServerAction) Run(_ *maa.Context, _ *maa.CustomActionArg) bool {
	fmt.Println("Agent server custom action is running")
	return true
}

// NewAgentServerAction creates the custom action registered by this server.
func NewAgentServerAction() maa.CustomActionRunner {
	return &AgentServerAction{}
}
