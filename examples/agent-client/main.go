package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/MaaXYZ/maa-framework-go/v4"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() (err error) {
	if err := maa.Init(); err != nil {
		return fmt.Errorf("init MAA: %w", err)
	}
	defer func() {
		if releaseErr := maa.Release(); releaseErr != nil {
			err = errors.Join(err, fmt.Errorf("release MAA: %w", releaseErr))
		}
	}()

	ctrl, err := maa.NewBlankController()
	if err != nil {
		return fmt.Errorf("create blank controller: %w", err)
	}
	defer func() {
		if destroyErr := ctrl.Destroy(); destroyErr != nil {
			err = errors.Join(err, fmt.Errorf("destroy blank controller: %w", destroyErr))
		}
	}()

	connectJob, err := ctrl.PostConnect()
	if err != nil {
		return fmt.Errorf("post connect: %w", err)
	}
	if !connectJob.Wait().Success() {
		return errors.New("blank controller connection failed")
	}

	// The task supplies its pipeline as an override, so no resource bundle is needed.
	res, err := maa.NewResource()
	if err != nil {
		return fmt.Errorf("create resource: %w", err)
	}
	defer func() {
		if destroyErr := res.Destroy(); destroyErr != nil {
			err = errors.Join(err, fmt.Errorf("destroy resource: %w", destroyErr))
		}
	}()

	client, err := maa.NewAgentClient(maa.WithTcpPort(7788))
	if err != nil {
		return fmt.Errorf("create agent client: %w", err)
	}
	defer func() {
		if disconnectErr := client.Disconnect(); disconnectErr != nil {
			err = errors.Join(err, fmt.Errorf("disconnect agent client: %w", disconnectErr))
		}
		if destroyErr := client.Destroy(); destroyErr != nil {
			err = errors.Join(err, fmt.Errorf("destroy agent client: %w", destroyErr))
		}
	}()

	if err := client.BindResource(res); err != nil {
		return fmt.Errorf("bind agent client resource: %w", err)
	}
	// Bound each communication wait, including the initial handshake.
	if err := client.SetTimeout(30 * time.Second); err != nil {
		return fmt.Errorf("set agent client timeout: %w", err)
	}
	identifier, err := client.Identifier()
	if err != nil {
		return fmt.Errorf("get agent client identifier: %w", err)
	}
	fmt.Printf("Start examples/agent-server with identifier %s\n", identifier)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("connect agent client: %w", err)
	}

	// Create the tasker last so cleanup finishes tasks before disconnecting the client.
	tasker, err := maa.NewTasker()
	if err != nil {
		return fmt.Errorf("create tasker: %w", err)
	}
	defer func() {
		// Wait can return just before the native task runner becomes idle.
		for tasker.Running() {
			time.Sleep(time.Millisecond)
		}
		if destroyErr := tasker.Destroy(); destroyErr != nil {
			err = errors.Join(err, fmt.Errorf("destroy tasker: %w", destroyErr))
		}
	}()

	if err := tasker.BindController(ctrl); err != nil {
		return fmt.Errorf("bind controller: %w", err)
	}
	if err := tasker.BindResource(res); err != nil {
		return fmt.Errorf("bind resource: %w", err)
	}
	if !tasker.Initialized() {
		return errors.New("tasker initialization check failed")
	}

	taskJob, err := tasker.PostTask("Test", map[string]any{
		"Test": map[string]any{
			"action": map[string]any{
				"type": "Custom",
				"param": map[string]any{
					"custom_action": "TestAgentServer",
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("post task: %w", err)
	}
	if !taskJob.Wait().Success() {
		return errors.New("Test task failed")
	}
	detail, err := taskJob.GetDetail()
	if err != nil {
		return fmt.Errorf("get task detail: %w", err)
	}
	fmt.Println(detail)
	return nil
}
