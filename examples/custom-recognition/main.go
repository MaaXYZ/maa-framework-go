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

	if err := res.RegisterCustomRecognition("MyRec", &MyRec{}); err != nil {
		return fmt.Errorf("register custom recognition: %w", err)
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

// MyRec demonstrates a custom recognition registered with a resource.
type MyRec struct{}

// Run implements maa.CustomRecognitionRunner and reports callback failures without exiting.
func (r *MyRec) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	result, err := r.run(ctx, arg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "custom recognition:", err)
		return nil, false
	}
	return result, true
}

func (r *MyRec) run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, error) {
	// DirectHit demonstrates Context operations without requiring OCR model files.
	detail, err := ctx.RunRecognition("MyDirectHit", arg.Img, directHitOverride(maa.Rect{100, 100, 200, 300}))
	if err != nil {
		return nil, fmt.Errorf("run recognition: %w", err)
	}
	if !detail.Hit {
		return nil, errors.New("DirectHit recognition did not match")
	}

	if err := ctx.OverridePipeline(directHitOverride(maa.Rect{1, 1, 114, 514})); err != nil {
		return nil, fmt.Errorf("override pipeline: %w", err)
	}
	newContext := ctx.Clone()
	if newContext == nil {
		return nil, errors.New("clone context failed")
	}
	if err := newContext.OverridePipeline(directHitOverride(maa.Rect{100, 200, 300, 400})); err != nil {
		return nil, fmt.Errorf("override cloned pipeline: %w", err)
	}
	// RunTask accepts a pipeline override, not an image. The clone already has its override.
	taskDetail, err := newContext.RunTask("MyDirectHit")
	if err != nil {
		return nil, fmt.Errorf("run cloned task: %w", err)
	}
	if !taskDetail.Status.Success() {
		return nil, errors.New("cloned task failed")
	}

	// Getter handles are borrowed from this callback and must not be destroyed.
	tasker := ctx.GetTasker()
	if tasker == nil {
		return nil, errors.New("get tasker failed")
	}
	ctrl := tasker.GetController()
	if ctrl == nil {
		return nil, errors.New("get controller failed")
	}
	clickJob, err := ctrl.PostClick(10, 20)
	if err != nil {
		return nil, fmt.Errorf("post click: %w", err)
	}
	if !clickJob.Wait().Success() {
		return nil, errors.New("click failed")
	}

	if err := ctx.OverrideNext(arg.CurrentTaskName, []maa.NextItem{
		{Name: "TaskA"},
		{Name: "TaskB"},
	}); err != nil {
		return nil, fmt.Errorf("override next: %w", err)
	}

	return &maa.CustomRecognitionResult{
		Box:    maa.Rect{0, 0, 100, 100},
		Detail: "Hello World!",
	}, nil
}

func directHitOverride(roi maa.Rect) map[string]any {
	return map[string]any{
		"MyDirectHit": map[string]any{
			"recognition": map[string]any{
				"type": "DirectHit",
				"param": map[string]any{
					"roi": roi,
				},
			},
		},
	}
}
