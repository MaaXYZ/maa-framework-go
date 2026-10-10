# v4.0.0-beta.19 → v4.0.0 migration guide

English | [简体中文](../../zh/migration/from-v4.0.0-beta.19.md)

This guide covers the net changes from `v4.0.0-beta.19` to `v4.0.0`.

If you use v3, follow the [v3 → v4 guide](from-v3.md). If you use beta.18, first follow the [beta.19 guide](from-v4.0.0-beta.18.md) for ownership changes. See the [CHANGELOG](../../../CHANGELOG.md) for the complete list.

v4.0.0 raises the minimum Go version from 1.24 to 1.25. Update your local and CI toolchains, then update the Go dependency:

```sh
go get github.com/MaaXYZ/maa-framework-go/v4@v4.0.0
```

Use a compatible set of MaaFramework shared libraries. Do not mix versions across the four libraries.

## 1. Update Post calls

All asynchronous submission methods now return submission errors directly. Update call sites, wrapper functions, interfaces, mocks, and stored method values:

| Type | Methods | Return value change |
| --- | --- | --- |
| Controller | `PostConnect`, `PostClick`, `PostClickV2`, `PostSwipe`, `PostSwipeV2`, `PostClickKey`, `PostInputText`, `PostStartApp`, `PostStopApp`, `PostTouchDown`, `PostTouchMove`, `PostTouchUp`, `PostRelativeMove`, `PostKeyDown`, `PostKeyUp`, `PostScreencap`, `PostScroll`, `PostInactive`, `PostShell` | `*Job` → `(*Job, error)` |
| Resource | `PostBundle`, `PostOcrModel`, `PostPipeline`, `PostImage` | `*Job` → `(*Job, error)` |
| Tasker | `PostTask`, `PostRecognition`, `PostAction`, `PostStop` | `*TaskJob` → `(*TaskJob, error)` |

Split `tasker.PostTask("Entry").Wait()` into submission and execution checks:

```go
func runTask(tasker *maa.Tasker, entry string) error {
	job, err := tasker.PostTask(entry)
	if err != nil {
		return fmt.Errorf("submit task: %w", err)
	}
	if !job.Wait().Success() {
		return fmt.Errorf("task failed: status=%v, error=%v", job.Status(), job.Error())
	}
	return nil
}
```

`Job.Error()` / `TaskJob.Error()` remain available. A failed submission returns both a non-nil Job in a terminal failed state and an `error`; you can still query the submission error if you retained only the Job. An asynchronous execution failure does not necessarily provide a Go `error`, so `job.Error() == nil` cannot replace a success status check.

Split the call when stopping a task as well:

```go
func stopTask(tasker *maa.Tasker) error {
	job, err := tasker.PostStop()
	if err != nil {
		return fmt.Errorf("submit stop: %w", err)
	}
	if !job.Wait().Success() {
		return fmt.Errorf("stop failed: status=%v, error=%v", job.Status(), job.Error())
	}
	return nil
}
```

Stopping may invalidate earlier Job IDs. An old Job's `Wait` returning does not by itself prove that all work has ended. Keep owner references, handle `ErrInUse` returned by `Destroy`, and retry after work has settled. Destroy the Resource / Controller after successfully destroying the Tasker. See [Controller.Destroy](../../../controller.go) and the [package documentation source](../../../doc.go) for the precise boundaries.

Parameter encoding failures in `Tasker.PostTask` and `Context.RunTask` / `RunRecognition` / `RunAction` / `WaitFreezes` now return an error and skip native execution, instead of falling back to `{}`. Check code and custom JSON codecs that relied on the silent fallback. Image inputs should also be non-nil and nonempty; failures to create or write image / rectangle buffers propagate to public calls.

## 2. Update Pipeline parameter types

| Field | beta.19 | v4.0.0 |
| --- | --- | --- |
| `LongPressParam.Duration`, `LongPressKeyParam.Duration` | `time.Duration` | `*time.Duration` |
| Neural network classification / detection `Expected` | `[]int` | `ClassSelectors` |
| TemplateMatch / FeatureMatch `Template`, OCR `Expected`, neural network classification / detection `Labels` | `[]string` | `StringList` |

Use a pointer to a variable for long press durations and `ClassIndex` for numeric classes. For example:

```go
func buildRecognitions() (*maa.Action, *maa.Recognition) {
	duration := 500 * time.Millisecond
	press := maa.ActLongPress(maa.LongPressParam{Duration: &duration})
	classify := maa.RecNeuralNetworkClassify(maa.NeuralNetworkClassifyParam{
		Model:  "model.onnx",
		Labels: maa.StringList{"button", "other"},
		Expected: maa.ClassSelectors{
			maa.ClassIndex(0),
			maa.ClassLabel("button"),
		},
	})
	return press, classify
}
```

`StringList` accepts a single string or an array of strings and encodes as an array. Most `[]string` literals assigned directly to these fields still compile; using `StringList` explicitly makes the new type clear. Class selections can mix numeric indices and labels, with numeric indices limited to the int32 range.

This interval also adds the following optional parameters. Set the fields directly when needed:

- Click / LongPress / Swipe / MultiSwipe `Pressure *int`.
- TouchDown / TouchMove / KeyDown / KeyUp `AutoUp *bool`; the feature takes effect in TouchDown / KeyDown, while the other two fields support protocol round trips.
- `ShellParam.ShellTimeout *time.Duration`, encoded as integer milliseconds in JSON.
- `DirectHitParam.ROI` and `ROIOffset *Rect`, for configuring the direct hit box.
- `NeuralNetworkDetectParam.Threshold []float64`, for configuring detection confidence thresholds.

## 3. Check JSON omission and clearing

Nil, zero values, and empty collections now have more distinct encoding behavior. In particular, check code that overrides an existing node with a zero-value object:

| Field | Omission / inheritance | Explicit override |
| --- | --- | --- |
| Optional pointers such as long press Duration, ShellTimeout, Pressure, AutoUp | nil omits the field | A pointer to zero or false emits that value |
| String lists, neural network Expected, TemplateMatch Threshold | nil omits the field | A non-nil empty list emits `[]`; whether it can execute still depends on native parameter requirements |
| `Node.Anchor` | nil omits the field | A non-nil empty map emits `{}`, clearing the node's anchor configuration; this does not clear runtime anchors already established |
| Value-type action / recognition offsets, `WaitFreezesParam.TargetOffset` | A zero `Rect` omits the field | Use raw parameters or raw JSON to explicitly reset the offset to zero |
| `DirectHitParam.ROIOffset` | nil omits the field | `&maa.Rect{}` emits `[0,0,0,0]` |
| Neural network detection `Threshold` | Both nil and an empty list omit the field | `[]float64{0}` emits a zero threshold |

Omitting a field lets MaaFramework retain the existing node value or use configured defaults; it does not guarantee a reset to the built-in default. For example, if the existing long press duration is 500 ms, `Duration: nil` does not change it back to 1000 ms.

Other value-type offsets cannot be explicitly reset to zero with `Rect{}`. You can use raw parameters for a known action:

```go
func clearClickOffset() *maa.Action {
	return &maa.Action{
		Type:  maa.ActionTypeClick,
		Param: maa.RawActionParam(`{"target_offset":[0,0,0,0]}`),
	}
}
```

`Target` accepts only true, a node name string, or 2 or 4 integers; false / null now return an error. An unspecified field can still use its zero value and be omitted, but do not put a zero `Target` in a target list that must be encoded or use `NewTargetBool(false)`. `Rect` decoding accepts only 2 or 4 integers, normalizing `[x,y]` to `[x,y,1,1]`; null, other lengths, and non-integer items return errors.

When decoding into a reused Node, note that the new `Node.UnmarshalJSON` replaces the entire node. Missing fields and `Name` are reset instead of merging into the old object. `Pipeline.UnmarshalJSON` likewise replaces the Pipeline and fills `Name` from each node key. For runtime incremental overrides, use `OverridePipeline`; do not treat Go decoding as a merge operation.

The following input forms are normalized: v1 flat recognition/action, flat parameters in v2 objects without param, single next/on_error entries, single templates / text / classes, single Swipe targets and timing values, key lists, Command args, ColorMatch color rows, OCR replace, and bare millisecond values for WaitFreezes. Anchor strings / string arrays resolve to the containing node only when decoding a complete Pipeline; direct Node decoding still requires a map. Output consistently uses canonical v2 forms; unmodeled fields in known parameters are not retained.

And / Or inline recognitions now encode as `{"sub_name":"...","recognition":{"type":"...","param":{...}}}`. Decoding remains compatible with the old flat form. Update expected values if tests or persistence code depend on the old JSON shape.

Numbers decoded in `CustomActionParam.CustomActionParam` now use `json.Number`. Replace `value.(float64)` with conversions from `json.Number` and handle `Int64` / `Float64` errors. This change does not apply to all `any` fields.

`RawActionParam` / `RawRecognitionParam` can also preserve parameter JSON for unknown types. They do not add native support for unknown algorithms or provide result parsing for unknown types; invalid parameters for known types still return errors.

## 4. Update screenshot options

In beta.19, `SetScreenshot` applied only the last option. Target dimensions, interpolation, and disabling raw size can now be combined. All parameters are validated before the options are applied in sequence. For example:

```go
func configureScreenshot(ctrl *maa.Controller) error {
	return ctrl.SetScreenshot(
		maa.WithScreenshotTargetLongSide(1280),
		maa.WithScreenshotResizeMethod(maa.ScreenshotResizeMethodLinear),
		maa.WithScreenshotUseRawSize(false),
	)
}
```

Remove repeated settings, avoid mixing long-side, short-side, and expand targets, and do not combine a target with `WithScreenshotUseRawSize(true)` in one call. Target dimensions must be positive and interpolation enum values must be valid. A Go parameter validation failure leaves native settings unchanged; if a native setter fails partway through, earlier changes are not rolled back.

The new `WithScreenshotTargetExpand(width, height)` scales proportionally up or down to cover both reference dimensions, without cropping or stretching. If you persist or compare Win32 names using `ScreencapMethod.String()`, update them to `DXGI_DesktopDup` / `DXGI_DesktopDup_Window`; parsing still accepts the old spellings.

## 5. Switch the inference provider entry point

`Resource.UseCoreml` has been removed. Upstream deprecated `MaaInferenceExecutionProvider_CoreML` in MaaFramework v5.14.3, and MaaDeps no longer ships the CoreML provider. Code that called it should use `Resource.UseWebgpu`, which takes a WebGPU device id (`InferenceDeviceAuto` lets the framework choose, which selects device 0):

```go
// beta.19
err := res.UseCoreml(maa.InferenceDeviceAuto)

// now
err := res.UseWebgpu(maa.InferenceDeviceAuto)
```

`UseWebgpu` needs a native library at MaaFramework v5.14.3 or later. An earlier version does not recognize the value: the setter still succeeds, but loading a model logs an invalid inference execution provider and falls back to CPU. That matches the existing behavior of `UseDirectml`, `UseCPU`, and the other provider setters, so confirm which provider is actually selected at runtime.

With the MaaFramework v5.14.3 libraries distributed by MaaDeps, `UseAutoExecutionProvider` selects an available provider in CUDA, DirectML, then WebGPU priority order. It falls back to CPU if none is available or the selected provider fails to initialize. MaaDeps no longer ships CoreML; check code that depends on the automatically selected provider.

## 6. Check registrations, Agent Server, and detail queries

- Configure event sinks and custom runners while the instance and its associated Tasker are idle. The new serialization and callback protections do not replace caller coordination during native execution; see the [package documentation source](../../../doc.go) for the precise contract. Do not pass a nil runner, including a typed nil function. Unregistering a nonexistent Resource runner name now succeeds; names registered outside the Go wrapper are preserved and return an error.
- Configure Agent Server before `AgentServerStartUp`. Restarting or configuring it is no longer allowed after startup, Join, or detach. Without detach, `AgentServerShutDown` enters a permanently closed state; `Release` / `Init` cannot recreate the service. Restart the process if you need a new service. See the [API comments](../../../agent_server.go) for how to arrange Join / ShutDown.
- `GetRecognitionDetail`, `GetActionDetail`, and `GetWaitFreezesDetail` now return errors when no detail is available. Handle `error` first instead of checking only `detail == nil`. `NodeDetail.Recognition` / `Action` can still be nil.
- Custom controllers support `ControllerFeatureNoScalingTouchPoints`. To disable automatic scaling of touch points, set it in the bitmask returned by `GetFeature`. Native destruction may call `KeyUp` / `TouchUp`; do not tear down the state needed by your implementation before `Destroy` succeeds.
- To subscribe to screen stabilization wait events, use `OnNodeWaitFreezesInContext` or the optional `ContextWaitFreezesEventSink`; existing `ContextEventSink` implementations do not need additional methods.
- On Windows, initialization with the default / empty library directory now preserves the existing DLL search configuration. Settings from a nonempty `WithLibDir` are still not restored after failure or `Release`; programs that depend on this process setting must coordinate it themselves.

If you use custom configuration for the API checker, check paths relative to the repository root and remove obsolete blacklist / exclusions entries. With MaaFramework v5.14.3 headers, add `MaaInferenceExecutionProvider_CoreML` to `constant_exclusions` with a nonempty reason for its intentional omission from Go; see the [CI configuration](../../../tools/api-check/config.ci.yaml). These exclusions use exact C constant names, suppress only missing Go constants, and fail when stale. When enabling Pipeline schema checks, use the schema from the same MaaFramework release as the headers and native libraries. See the [tool documentation](../../../tools/api-check/README.md).

## 7. Verify the migration

- [ ] Use Go 1.25 or later locally and in CI, and run `go mod tidy` after updating the dependency.
- [ ] Update all `Post*` signatures and check submission errors separately from asynchronous status.
- [ ] Check pointers, class selection types, zero offsets, empty collections, and Node decoding replacement behavior.
- [ ] Update code and tests that depend on the inline recognition JSON shape or `float64` type assertions.
- [ ] Verify paths for parameter encoding failures, empty images, unavailable details, and conflicting screenshot options.
- [ ] Replace `UseCoreml` calls with `UseWebgpu`, and confirm the native library version and the provider actually selected at runtime.
- [ ] Verify stop and exit paths, handle `Destroy` / `Release` errors, and do not attempt to restart Agent Server after shutdown.
- [ ] Run your project's `go build ./...`, `go vet ./...`, and relevant tests.
