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

	devices, err := maa.FindAdbDevices()
	if err != nil {
		return fmt.Errorf("find ADB devices: %w", err)
	}
	if len(devices) == 0 {
		return errors.New("no ADB devices found; connect a device or start an emulator")
	}
	device := devices[0]
	ctrl, err := maa.NewAdbController(
		device.AdbPath,
		device.Address,
		device.ScreencapMethod,
		device.InputMethod,
		device.Config,
		"path/to/MaaAgentBinary",
	)
	if err != nil {
		return fmt.Errorf("create ADB controller: %w", err)
	}
	defer func() {
		if destroyErr := ctrl.Destroy(); destroyErr != nil {
			err = errors.Join(err, fmt.Errorf("destroy ADB controller: %w", destroyErr))
		}
	}()

	connectJob, err := ctrl.PostConnect()
	if err != nil {
		return fmt.Errorf("post connect: %w", err)
	}
	if !connectJob.Wait().Success() {
		return errors.New("ADB controller connection failed")
	}

	res, err := maa.NewResource()
	if err != nil {
		return fmt.Errorf("create resource: %w", err)
	}
	defer func() {
		if destroyErr := res.Destroy(); destroyErr != nil {
			err = errors.Join(err, fmt.Errorf("destroy resource: %w", destroyErr))
		}
	}()

	bundleJob, err := res.PostBundle("./resource")
	if err != nil {
		return fmt.Errorf("post resource bundle: %w", err)
	}
	if !bundleJob.Wait().Success() {
		return errors.New("resource bundle loading failed")
	}

	// Create the tasker last so deferred cleanup destroys it before its bindings.
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

	taskJob, err := tasker.PostTask("Startup")
	if err != nil {
		return fmt.Errorf("post task: %w", err)
	}
	if !taskJob.Wait().Success() {
		return errors.New("Startup task failed")
	}
	detail, err := taskJob.GetDetail()
	if err != nil {
		return fmt.Errorf("get task detail: %w", err)
	}
	fmt.Println(detail)
	return nil
}
