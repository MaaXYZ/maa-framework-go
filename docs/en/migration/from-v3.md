# v3 → v4.0.0 migration guide (release preparation draft)

English | [简体中文](../../zh/migration/from-v3.md)

This guide starts from `v3.6.0-beta.5` and uses commit [`2c24945`](https://github.com/MaaXYZ/maa-framework-go/tree/2c24945e7a9e23993f13743cc4962a2bd985641a) as its verified endpoint. `v4.0.0` has not been tagged. This document describes migration to that fixed snapshot; differences from the final release must be checked once its endpoint is determined.

Projects already using v4 should choose either [beta.18 → beta.19](from-v4.0.0-beta.18.md) or [beta.19 → release preparation snapshot](from-v4.0.0-beta.19.md) based on their starting version. See [CHANGELOG](../../../CHANGELOG.md) for the detailed changes. This guide follows the order of changes needed in application code.

## 1. Update module paths and native libraries

Go 1.25 or later is required. Change the import paths for the main package and controller subpackages from `/v3` to `/v4`:

```go
import (
	maa "github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/MaaXYZ/maa-framework-go/v4/controller/adb"
)
```

Before the final tag is created, specify the commit in your project to try the fixed snapshot covered by this guide:

```sh
go get github.com/MaaXYZ/maa-framework-go/v4@2c24945e7a9e23993f13743cc4962a2bd985641a
go mod tidy
```

After updating all imports, check `go.mod` for unused `/v3` dependencies. Do not use an installation command with the unpublished `@v4.0.0` tag.

Also update the MaaFramework dynamic libraries. This binding tracks the latest MaaFramework, including prereleases, and does not guarantee compatibility with earlier versions. All four libraries—`MaaFramework`, `MaaToolkit`, `MaaAgentServer`, and `MaaAgentClient`—should come from the same compatible release. If a library or symbol is missing, check the library paths and symbol names in `LibraryLoadError` / `SymbolLookupError`, then replace the full set of libraries.

## 2. Separate call errors from asynchronous execution results

Constructors now return `(*T, error)`. Most setters, queries, binding methods, registration methods, and similar APIs that previously returned `bool` or `(T, bool)` now return `error` or `(T, error)`. Update boolean checks accordingly.

| APIs used by previous code | Handling in v4 |
| --- | --- |
| `NewTasker`, `NewResource`, and controller constructors | Check the second return value and use the object only after success |
| `BindResource`, `BindController`, `OverridePipeline`, custom action/recognition registration, and global configuration | Check the returned `error` |
| Device/window discovery, node/task details, resource hashes and lists, and screenshot reads | Receive the result and `error`; `Pipeline.GetNode` still returns `(*Node, bool)` |
| `Context.RunTask` / `RunRecognition` / `RunAction` and their Direct variants | Receive the details and `error` |
| Controller / Resource / Tasker `Post*` methods | Receive the Job and submission `error`, then wait and check execution status |

For example, handle the two stages of connecting a controller separately:

```go
func connect(ctrl *maa.Controller) error {
	job, err := ctrl.PostConnect()
	if err != nil {
		return fmt.Errorf("submit connection: %w", err)
	}
	if !job.Wait().Success() {
		return fmt.Errorf("connection failed: status=%v, error=%v", job.Status(), job.Error())
	}
	return nil
}
```

`err == nil` means only that submission succeeded. `Job.Error()` / `TaskJob.Error()` remain available for submission or lifecycle errors. An asynchronous task may fail without a Go `error`, so also check `Wait().Success()`. Split chained calls such as `PostTask(...).Wait()` and `PostStop().Wait()`. See the [beta.19 → v4.0.0 guide](from-v4.0.0-beta.19.md#1-update-post-calls) for the signature table and a stopping example.

Override parameters still accept JSON strings, `[]byte`, or encodable Go values; nil means `{}`. Encoding failures now return an error and prevent execution instead of silently falling back to `{}`. Check custom JSON codecs and override objects with fields that cannot be encoded.

## 3. Update object cleanup and callback code

Keep the owners returned by constructors, and arrange cleanup immediately after successful initialization so that early failure paths are covered. Before exiting, stop new submissions and wait for all submitted Jobs and callbacks to finish. Then clean up in the order **AgentClient → Tasker → Resource / Controller → Release**, handling errors from `Destroy()`.

The following helper accepts owner pointers. It can be called from a single deferred cleanup function to cover paths where only some constructors succeeded. Finish asynchronous work before calling it; a destruction error does not mean cleanup succeeded:

```go
func closeOwned(client *maa.AgentClient, tasker *maa.Tasker, res *maa.Resource, ctrl *maa.Controller) error {
	var err error
	if client != nil {
		err = errors.Join(err, client.Destroy())
	}
	if tasker != nil {
		err = errors.Join(err, tasker.Destroy())
	}
	if res != nil {
		err = errors.Join(err, res.Destroy())
	}
	if ctrl != nil {
		err = errors.Join(err, ctrl.Destroy())
	}
	return errors.Join(err, maa.Release())
}
```

If your project stores `Destroy` in `[]func()` or requires a type to implement `interface{ Destroy() }`, change the function or interface to return `error` and handle it at the call site.

Remove `Destroy` calls on borrowed objects returned by getters and event callbacks; their original owners clean them up. Do not keep a callback's `Context` or its `Clone` beyond the callback. Use across goroutines must also finish before that callback returns. Make configuration and registration changes when the instance and associated Tasker are idle, and avoid replacing, removing, or clearing registrations within callbacks.

See the fixed snapshot's [package documentation source](https://github.com/MaaXYZ/maa-framework-go/blob/2c24945e7a9e23993f13743cc4962a2bd985641a/doc.go) for the full ownership, concurrency, and callback contracts. The relevant API comments list retryable errors for each `Destroy` method.

## 4. Rewrite Pipeline construction code

### Type names and constructors

| v3 API | v4 API |
| --- | --- |
| `NodeAction` / `NodeActionType` / `NodeActionParam` | `Action` / `ActionType` / `ActionParam` |
| `NodeRecognition` / `NodeRecognitionType` / `NodeRecognitionParam` | `Recognition` / `RecognitionType` / `RecognitionParam` |
| Action and recognition parameters and enums prefixed with `Node` | Remove the prefix, such as `NodeClickParam` → `ClickParam` |
| `NodeNextItem` / `NodeMultiSwipeItem` / `NodeWaitFreezes` | `NextItem` / `MultiSwipeItem` / `WaitFreezesParam` |
| `NewNode(name, WithRecognition(...), WithAction(...), ...)` | `NewNode(name).SetRecognition(...).SetAction(...)` |
| `With*` parameter options for `Act*` / `Rec*` | Use the corresponding parameter structs; call retained convenience constructors according to their signatures |
| `NewMultiSwipeItem`, the `WaitFreezes` constructor, and related options | Create `MultiSwipeItem` and `WaitFreezesParam` directly |

For example, use field initialization and chained node setters:

```go
func buildPipeline() *maa.Pipeline {
	duration := 500 * time.Millisecond
	node := maa.NewNode("Press").
		SetRecognition(maa.RecTemplateMatch(maa.TemplateMatchParam{
			Template: maa.StringList{"button.png"},
		})).
		SetAction(maa.ActLongPress(maa.LongPressParam{
			Target:   maa.NewTargetBool(true),
			Duration: &duration,
		})).
		SetNext([]maa.NextItem{{Name: "Done"}}).
		SetPreDelay(100 * time.Millisecond)
	return maa.NewPipeline().AddNode(node).
		AddNode(maa.NewNode("Done").SetAction(maa.ActDoNothing()))
}
```

Go builders encode to nested Pipeline v2 `type` / `param` JSON. Existing flat v1 JSON can be read and normalized to v2, so upgrading alone does not require rewriting all resource files. Do not rely on byte-for-byte JSON equality. Unmodeled fields in known parameters are not preserved after decoding. `RawActionParam` / `RawRecognitionParam` can preserve parameters for unknown action/recognition types, but the native library must still support those types.

### Time values, optional values, and lists

- `LongPressParam.Duration` and `LongPressKeyParam.Duration` are `*time.Duration`. Swipe / MultiSwipe durations, end holds, MultiSwipe start times, and `WaitFreezesParam` time fields use `time.Duration` or slices of it. JSON still represents these values as integer milliseconds. Node fields such as `RateLimit` and `Timeout` remain `*int64` millisecond values; they can still be set through `Set*` methods accepting `time.Duration`.
- `ShellParam.ShellTimeout` is `*time.Duration`. A nil optional pointer leaves the existing value or default configuration unchanged. A pointer to zero explicitly writes zero; nil does not mean "restore the built-in default".
- Template lists, OCR expected text lists, and neural network label lists now use `StringList`. Neural network `Expected` now uses `ClassSelectors`: use `ClassIndex(n)` for numeric classes and `ClassLabel("name")` for labels, for example `maa.ClassSelectors{maa.ClassIndex(0), maa.ClassLabel("button")}`.
- Check code that previously used a zero `Target` / `Rect` to override an existing ROI or coordinates, and code using non-nil empty lists. See the [beta.19 → v4.0.0 guide](from-v4.0.0-beta.19.md#3-check-json-omission-and-clearing) for field-specific omission, clearing, and explicit-zero rules. Do not apply a single rule to all fields.

### Anchors, next, and composite recognition

- `Node.Anchor` changes from `[]string` to `map[string]string`. Use `AddAnchor("name")` to point to the current node and `SetAnchorTarget("name", "TargetNode")` to point to a specified node. Use `SetAnchor(map[string]string{})` to clear all node anchor configuration and `RemoveAnchor("name")` to remove a specific anchor from the configuration. `ClearAnchor("name")` sets an empty target, arranging for that runtime anchor to be cleared after the node executes.
- `Context.OverrideNext` / `Resource.OverrideNext` now takes `[]NextItem`. Use `NextItem{Name: "name", JumpBack: true}` or `Anchor: true` for attributes instead of passing `[]string`.
- And / Or now uses `SubRecognitionItem`. Use the variadic arguments of `RecAnd(...)` / `RecOr(...)`, `Ref("NodeName")`, and `Inline(recognition)`. Use `Recognition.SetBoxIndex` when a box index is needed; remove uses of `NodeAndRecognitionItem`, `AndItem`, and the old configuration options.
- `Context.WaitFreezes` now takes a fixed `*WaitFreezesParam` argument and returns `error`. See the [API comments](https://github.com/MaaXYZ/maa-framework-go/blob/2c24945e7a9e23993f13743cc4962a2bd985641a/context.go) for call details.

## 5. Update detail access and custom extensions

| v3 usage | v4 usage |
| --- | --- |
| `TaskDetail.NodeDetails []*NodeDetail` | Iterate over `TaskDetail.Nodes []NodeRef`, call `NodeRef.GetDetail()` as needed, and handle errors |
| `TaskJob.Job` | Call `Wait`, status, `Error`, and other methods directly on `TaskJob` |
| `TaskDetail` in custom action/recognition arguments | Read `TaskID` and query `ctx.GetTasker().GetTaskDetail(arg.TaskID)` as needed within the callback |
| `CustomAction` / `CustomRecognition` aliases | Use `CustomActionRunner` / `CustomRecognitionRunner`; use `CustomActionFunc` / `CustomRecognitionFunc` for ordinary functions |
| Exported `*EventSinkAdapter` types | Implement the corresponding EventSink interface or use `On*` registration methods |
| `Context.GetNodeData` / `Resource.OverriderImage` | `Context.GetNode` / `Resource.OverrideImage` |
| `RecognitionResults.Best` slice | `*RecognitionResult`; check for nil before accessing it |
| `ShellActionResult.Timeout` / `NodeNextListDetail.NextList` | `ShellTimeout` / `List []NextItem`; JSON names also change to `shell_timeout` / `list` |
| `NeuralNetworkClassifyResult.Raw` / `Probs` | Remove access to these two removed fields |
| Concrete image return type `*image.NRGBA` | `*image.RGBA`; update type assertions when a concrete type is needed, or use `image.Image` in most cases |

`CustomController` must additionally implement `RelativeMove(dx, dy int32) bool`, `Shell(cmd string, timeout int64) (string, bool)`, `Inactive() bool`, and `GetInfo() (string, bool)`. Implementations must support concurrent calls from native threads. Keyboard/touch release handling may also occur during destruction.

Return `nil, false` when custom recognition misses and has no diagnostic information. Returning a non-nil result with false preserves its Box / Detail. The boolean determines whether recognition matches.

## 6. Update controllers, Agent, and initialization options

- `Resource.UseCoreml` has been removed. Replace calls with `Resource.UseWebgpu(maa.InferenceDeviceAuto)` using MaaFramework v5.14.3 or later. The Auto provider now tries CUDA, DirectML, then WebGPU before falling back to CPU. See the [provider migration step](from-v4.0.0-beta.19.md#5-switch-the-inference-provider-entry-point) for device selection and older-library behavior.
- The three previous screenshot methods, `SetScreenshotTarget*` / `SetScreenshotUseRawSize`, are replaced by `SetScreenshot(WithScreenshot*...) error`. Target size, interpolation, and disabling raw size can be configured together. Check mutually exclusive targets and duplicate options before use. See the [beta.19 → v4.0.0 guide](from-v4.0.0-beta.19.md#4-update-screenshot-options) for examples.
- Win32 `InputSendMessageWithCursorPosAndBlockInput` / `InputPostMessageWithCursorPosAndBlockInput` become `InputSendMessageWithWindowPos` / `InputPostMessageWithWindowPos`; update string configuration as well. The spellings `InterenceDevice` / `InterenceDeviceAuto` are corrected to `InferenceDevice` / `InferenceDeviceAuto`.
- `CarouselImageController` / `NewCarouselImageController` have been removed. Choose a custom controller or a recording/playback controller based on your use case. The new Linux entry point from v3 to this snapshot is `NewLinuxController(configJson)`; there is no need to use `NewWlRootsController`, which was added and then removed in intermediate versions.
- Agent construction is unified as `NewAgentClient(opts ...AgentClientOption) (*AgentClient, error)`. Replace the positional argument with `WithIdentifier(id)` and replace `NewAgentClientTcp` with `WithTcpPort(port)`. Pair the actual value returned by `client.Identifier()` with `AgentServerStartUp`, and handle errors from both construction and startup.
- `InitConfig` is no longer exported; use the provided `With*` initialization options. `Init` no longer implicitly sets global configuration such as the log directory or log level, so pass required options explicitly. Repeated `Init` calls after successful initialization are no-ops, and options passed later are not applied. Remove checks for the removed `ErrAlreadyInitialized` / `ErrNotInitialized` errors.

Arrange Agent Server startup, Join, shutdown, and Detach according to the [API comments](https://github.com/MaaXYZ/maa-framework-go/blob/2c24945e7a9e23993f13743cc4962a2bd985641a/agent_server.go). Programs that need to unload the libraries should keep the server thread undetached. On Windows, a nonempty `WithLibDir` affects the process DLL search configuration; failure rollback and `Release` do not restore this setting. See [WithLibDir](https://github.com/MaaXYZ/maa-framework-go/blob/2c24945e7a9e23993f13743cc4962a2bd985641a/maa.go).

## 7. Verify the migration

- [ ] Use Go 1.25 or later locally and in CI, and run `go mod tidy` after updating the dependency.
- [ ] All package paths use `/v4`, and construction, binding, configuration, query, and submission errors are handled.
- [ ] Asynchronous execution results are checked separately for connection, resource loading, and the final task.
- [ ] Compare JSON and execution behavior for key nodes using real resources, checking omitted fields, explicit zeros, empty lists, and anchors.
- [ ] Check detail queries and `Context` use within callbacks. Wait for work to finish on exit and confirm that `Destroy` / `Release` succeeds.
- [ ] Run your project's `go build ./...`, `go vet ./...`, and relevant tests, covering normal completion, execution failure, and early exit paths.
- [ ] Once the final release tag is determined, use the updated guide to check additional changes from this snapshot to the final release.
