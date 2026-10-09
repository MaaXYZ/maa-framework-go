# v4.0.0-beta.18 → v4.0.0-beta.19 migration guide

English | [简体中文](../../zh/migration/from-v4.0.0-beta.18.md)

This guide covers only the migration from **v4.0.0-beta.18 to v4.0.0-beta.19**. For further upgrades, see the [beta.19 → v4.0.0 guide](from-v4.0.0-beta.19.md).

The main changes in this release **make native handle ownership and lifetimes explicit** and make `Init` / `Release` safe under concurrent calls and failure conditions. Most code will still compile, but the following patterns change behavior at runtime. Check each item before upgrading.

## Pre-upgrade checklist

- [ ] Upgrade MaaFramework to **v5.13.0 or later** (v5.14.0 is recommended; `auto_up` requires v5.14.0)
- [ ] Replace `NewWlRootsController` with `NewLinuxController`
- [ ] Check the return value of `Destroy`, especially the order of deferred calls
- [ ] Do not use a callback's `Context` after the callback returns, including passing it to a goroutine
- [ ] Do not call `Destroy` on objects returned by `GetResource`, `GetController`, `Context.GetTasker`, or similar getters
- [ ] Check the return value of `Release` if your program uses it
- [ ] If custom recognition returns a non-nil result on a miss, confirm that this is intentional

## MaaFramework version requirement

beta.19 binds symbols added in MaaFramework v5.13.0 (`MaaLinuxControllerCreate`, `MaaToolkitGamescopeInstance*`, and `MaaToolkitPortalHelper*`). These symbols are exported by the dynamic libraries on all platforms, so MaaFramework ≥ v5.13.0 is required on **every platform**.

`Init` now resolves each library's symbols before registering them and returns `*SymbolLookupError` if any are missing. Previously, a missing symbol only caused a problem when its function was called:

```go
if err := maa.Init(maa.WithLibDir(libDir)); err != nil {
	var symErr *maa.SymbolLookupError
	if errors.As(err, &symErr) {
		// symErr.LibraryName / LibraryPath / SymbolName identify the library
		// and missing symbol, usually because MaaFramework is too old.
	}
	return err
}
```

## Breaking changes

### API changes at a glance

| Change | Previous API | New API |
|--------|--------------|---------|
| Return value | `(*Tasker).Destroy()` | `(*Tasker).Destroy() error` |
| Return value | `(*Resource).Destroy()` | `(*Resource).Destroy() error` |
| Return value | `(*Controller).Destroy()` | `(*Controller).Destroy() error` |
| Return value | `(*AgentClient).Destroy()` | `(*AgentClient).Destroy() error` |
| Removed | `NewWlRootsController(wlrSocketPath string, useWin32VkCode bool)` | `NewLinuxController(configJson string)` |
| Behavior | `Release() error` only reports unload failures | Returns `ErrLibraryInUse` when objects are still alive or the Agent Server has not shut down |

#### Destroy returns an error

Direct calls to `x.Destroy()` and `defer x.Destroy()` still compile, but **silently discard errors**. The following patterns no longer compile and must be updated:

- Using `Destroy` as a `func()` value, such as `cleanups = append(cleanups, tasker.Destroy)`
- Requiring these types to implement a custom `interface{ Destroy() }`

`Destroy` can return the following errors:

| Error | Trigger | Handling |
|-------|---------|----------|
| `ErrBorrowed` | Calling `Destroy` on a borrowed view obtained from a getter or callback | Remove the call; the owner is responsible for destruction |
| `ErrBound` | A Resource / Controller is still bound to a Tasker or retained by an AgentClient; a Tasker is still registered as an AgentClient sink | Destroy in dependency order: AgentClient → Tasker → Resource / Controller |
| `ErrInCallback` | Destroying an object from its own callback | Destroy it after the callback returns |
| `ErrInUse` | A method call or asynchronous Job is still executing, **even if the Job has been discarded** | Wait for it to finish (`Wait` / `PostStop().Wait()`), then retry |

Repeated successful calls to `Destroy` on an owner are safe.

#### Destruction order: AgentClient → Tasker → Resource / Controller

Previously, destroying a bound object first still freed its native object. The new behavior returns `ErrBound` and **does not free the object**, which also causes `Release` to return `ErrLibraryInUse`.

```go
// Before: defer runs in LIFO order, so res is destroyed before tasker.
tasker, _ := maa.NewTasker()
defer tasker.Destroy()
res, _ := maa.NewResource()
defer res.Destroy() // tasker is still bound to res → ErrBound, silently ignored.
tasker.BindResource(res)

// After: create and defer bound objects first, then defer tasker last.
ctrl, _ := maa.NewAdbController(...)
defer ctrl.Destroy()
res, _ := maa.NewResource()
defer res.Destroy()
tasker, _ := maa.NewTasker()
defer tasker.Destroy() // Runs first; unbinds res / ctrl so they can be destroyed.
tasker.BindResource(res)
tasker.BindController(ctrl)
```

The same applies to AgentClient: it retains its bound Resource and any Resource / Controller / Tasker registered as a sink, so destroy it first.

**Additional note:** After another call to `BindResource` / `BindController`, previously bound objects are still retained by the Tasker. Calling `Destroy` on them returns `ErrBound` until the Tasker is destroyed.

#### Destroying a Tasker while tasks are running

`Destroy` returns `ErrInUse` when the Tasker has unfinished Jobs. Stop and wait before exiting:

```go
tasker.PostStop().Wait()
if err := tasker.Destroy(); err != nil {
	log.Printf("destroy tasker: %v", err)
}
```

#### Borrowed views cannot be destroyed

Constructors (`NewTasker`, `NewResource`, and the `NewXxxController` functions) return **owners**. The following sources return **borrowed views**, and calling `Destroy` on them returns `ErrBorrowed`:

- `Tasker.GetResource()`, `Tasker.GetController()`
- `Context.GetTasker()`
- Objects passed to event callbacks

Previously, calling `Destroy` on these objects directly freed the native handle, potentially causing a double free or dangling handle. The new behavior rejects that call. Borrowed views also cannot be passed to `BindResource` / `BindController`, and a borrowed Tasker cannot bind other objects. Both cases return `ErrBorrowed`.

#### A callback Context expires when the callback returns

The `*Context` received by custom recognition, custom actions, and event callbacks (including copies obtained through `Clone`) is valid only during that callback. After the callback returns:

- Methods that return errors (`RunTask`, `OverridePipeline`, etc.) return `ErrClosed`
- `GetTasker()` and `Clone()` return nil
- `GetTaskJob()` returns a failed Job whose `Error()` is `ErrClosed`

```go
// Before: use ctx after the callback returns.
func (a *MyAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	go func() {
		ctx.RunTask("Next") // The callback has returned → ErrClosed.
	}()
	return true
}

// After: finish using ctx within the callback.
func (a *MyAction) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	_, err := ctx.RunTask("Next")
	return err == nil
}
```

#### Calls after an object is closed

After an object is destroyed, methods that return errors return `ErrClosed`. Submitted Jobs expose this error through the new `(*Job).Error()` method. If you need a Job's result, call `Wait` before destroying its owner.

#### Rebinding a running Tasker

`BindResource` / `BindController` returns `ErrTaskerRunning` when the Tasker has pending or running tasks.

#### Init / Release

- `Init` and `Release` are serialized with each other; **other MAA functions must not run concurrently with either one**. `IsInited` may be called concurrently.
- `Release` returns `ErrLibraryInUse` in the following cases:
  - Native objects are still alive: a Tasker / Resource / Controller / AgentClient has not been destroyed or its `Destroy` call failed, or a `PortalHelper` has not been destroyed (its `Destroy` has no return value)
  - The Agent Server has not been shut down with `AgentServerShutDown`
  - `AgentServerDetach` has been called: after this, `Release` is rejected for **the rest of the process lifetime**, even if `AgentServerJoin` / `AgentServerShutDown` is later called. Programs that need to release the libraries must not detach the server thread
- If unloading fails, `IsInited` becomes false. Call `Release` again to retry cleanup, then call `Init` again.

```go
// Agent Server process.
maa.AgentServerStartUp(socketID)
maa.AgentServerJoin()     // Blocks until AgentClient calls Disconnect to request shutdown and the server thread exits; it does not request shutdown itself.
maa.AgentServerShutDown() // Still required after Join returns before Release is allowed.
if err := maa.Release(); err != nil {
	log.Printf("release: %v", err) // ErrLibraryInUse means some objects have not been destroyed.
}
```

#### NewWlRootsController removed

MaaFramework has deprecated `MaaWlRootsControllerCreate`. The Go binding removes the corresponding constructor in favor of `NewLinuxController`, which selects screenshot and input methods through JSON configuration:

```go
// Before.
ctrl, err := maa.NewWlRootsController("/run/user/1000/wayland-0", true)

// After (1 = MaaLinuxScreencapMethod_Wlr / MaaLinuxInputMethod_Wlr).
ctrl, err := maa.NewLinuxController(`{
	"screencap_method": 1,
	"input_method": 1,
	"wlr_socket_path": "/run/user/1000/wayland-0",
	"use_win32_vk_code": true
}`)
```

See MaaFramework's `MaaLinuxControllerCreate` documentation for all configuration options (PipeWire, UInput, Libei, etc.). This controller is available only on Linux.

## Behavior changes

- **Custom recognition results are preserved on a miss:** When `CustomRecognitionRunner.Run` returns `(result, false)` with a non-nil `result`, `Box` and `Detail` are now passed to MaaFramework and appear in recognition details; previously they were discarded. Whether recognition matches is still determined only by the second return value. Returning nil always counts as a miss. If old code returns a placeholder result on a miss, change it to `nil, false`.
- **A `null` action detail is treated as empty:** Parsing `ActionDetail` now returns nil for `"null"`, just as it does for `""` and `"{}"`, instead of reporting a parsing error.

## Added

| Category | New API | Description |
|----------|---------|-------------|
| Controller | `NewLinuxController(configJson string) (*Controller, error)` | Linux native application controller; supports Wlr / PipeWire screenshots and Wlr / UInput / Libei input |
| Toolkit | `FindGamescopeInstances() ([]*GamescopeInstance, error)`, `GamescopeInstance` | Discovers gamescope instances (display number, EIS socket, and PipeWire node) for Linux controller configuration |
| Toolkit | `NewPortalHelper() (*PortalHelper, error)`, `PortalHelper` | Obtains a PipeWire screen stream through xdg-desktop-portal, with persistent restore token support; available only on desktop Linux, and creation fails on other platforms (including Android) |
| Win32 | `win32.InputAnchoredTouch` | Injects touch points without moving the cursor or target window; supports clicks and swipes, but not scrolling or keyboard input |
| Job | `(*Job).Error() error` | Returns the reason a Job could not be submitted or an object was closed, among other errors |
| ActionResult | `ClickKeyActionResult.AutoUp`, `TouchActionResult.AutoUp` | Correspond to the `auto_up` field in MaaFramework action details; native automatic release requires MaaFramework ≥ v5.14.0, and the field is always false on earlier versions |
| Errors | `ErrClosed`, `ErrBorrowed`, `ErrBound`, `ErrInCallback`, `ErrInUse`, `ErrTaskerRunning`, `ErrLibraryInUse` | See above |
| Errors | `SymbolLookupError` | Returned when `Init` cannot find a required exported symbol |
| Platform | Android | Loads `lib*.so` when `GOOS=android`; requires `CGO_ENABLED=1` and compilation with the Android NDK toolchain |

## Fixed

- Made native handle ownership explicit, fixing double frees of borrowed objects and callbacks accessing destroyed objects (#46)
- Made `Init` / `Release` safe under concurrent calls and load failure conditions, allowing retries after failure (#47)
- Preserved recognition details when custom recognition misses (#42)
