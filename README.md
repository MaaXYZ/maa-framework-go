<!-- markdownlint-disable MD033 MD041 -->
<p align="center">
  <img alt="LOGO" src="https://cdn.jsdelivr.net/gh/MaaAssistantArknights/design@main/logo/maa-logo_512x512.png" width="256" height="256" />
</p>

<h1 align="center">MaaFramework Go Binding</h1>

<div align="center">
  <div>
    <a href="https://github.com/MaaXYZ/maa-framework-go/blob/main/LICENSE.md">
      <img alt="license" src="https://img.shields.io/github/license/MaaXYZ/maa-framework-go">
    </a>
    <a href="https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4">
      <img alt="go reference" src="https://pkg.go.dev/badge/github.com/MaaXYZ/maa-framework-go/v4.svg">
    </a>
    <a href="https://goreportcard.com/report/github.com/MaaXYZ/maa-framework-go/v4">
      <img alt="go report" src="https://goreportcard.com/badge/github.com/MaaXYZ/maa-framework-go/v4">
    </a>
  </div>
  <div>
    <a href="https://github.com/MaaXYZ/MaaFramework/releases/tag/v5.14.2">
      <img alt="maa framework" src="https://img.shields.io/badge/MaaFramework-v5.14.2-blue">
    </a>
    <a href="https://deepwiki.com/MaaXYZ/maa-framework-go">
      <img src="https://deepwiki.com/badge.svg" alt="Ask DeepWiki">
    </a>
  </div>
</div>

<br />

<p align="center">
  English | <a href="README_zh.md">简体中文</a>
</p>

Go binding for [MaaFramework](https://github.com/MaaXYZ/MaaFramework), a cross-platform automation testing framework based on image recognition.

> **🚀 No Cgo Required!** Pure Go implementation using [purego](https://github.com/ebitengine/purego).
>
> Android is the exception: purego loads libraries through cgo there, so Android builds need
> `CGO_ENABLED=1` and an NDK clang as `CC`, e.g.
> `GOOS=android GOARCH=arm64 CGO_ENABLED=1 CC=<ndk>/toolchains/llvm/prebuilt/<host>/bin/clang CGO_CFLAGS=--target=aarch64-linux-android24 CGO_LDFLAGS=--target=aarch64-linux-android24 go build`

## ✨ Features

- **Cross-platform Controllers** - ADB, Win32, Linux, macOS, PlayCover, and Android Native
- **Recording and Replay** - Capture controller operations to JSONL, then replay them for debugging and regression testing
- **Virtual Gamepad Controller** (Windows only) - Gamepad automation via ViGEm
- **Toolkit Utilities** - Find ADB devices and desktop windows; manage macOS automation permissions
- **Image Recognition** - Template matching, OCR, and feature detection
- **Custom Extensions** - Custom recognitions, actions, and controllers in pure Go
- **Agent Support** - Run custom recognition and action logic from an external process
- **Async Jobs and Events** - Poll job status and task details, or subscribe to resource, controller, and tasker events
- **Pipeline v2 model and runtime APIs** - Typed `Pipeline`, `Node`, `Action`, and `Recognition` builders emit nested v2 JSON and support runtime execution from a `Context`. Unknown parameters can be retained as raw JSON with [`RawActionParam`](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#RawActionParam) and [`RawRecognitionParam`](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#RawRecognitionParam); this fallback does not add native support for unknown types.

## 📦 Installation

Requires Go 1.24 or later.

### 1. Install Go Package

```shell
go get github.com/MaaXYZ/maa-framework-go/v4
```

### 2. Download MaaFramework

Download the [MaaFramework Release](https://github.com/MaaXYZ/MaaFramework/releases) for your platform and extract it.

| Platform | Architecture | Download                    |
| -------- | ------------ | --------------------------- |
| Windows  | amd64        | `MAA-win-x86_64-*.zip`      |
| Windows  | arm64        | `MAA-win-aarch64-*.zip`     |
| Linux    | amd64        | `MAA-linux-x86_64-*.zip`    |
| Linux    | arm64        | `MAA-linux-aarch64-*.zip`   |
| macOS    | amd64        | `MAA-macos-x86_64-*.zip`    |
| macOS    | arm64        | `MAA-macos-aarch64-*.zip`   |
| Android  | amd64        | `MAA-android-x86_64-*.zip`  |
| Android  | arm64        | `MAA-android-aarch64-*.zip` |

## ⚙️ Runtime Requirements

Programs built with maa-framework-go require MaaFramework dynamic libraries at runtime. Provide them in one of these ways:

1. **Via `Init()` Option** - Specify library path programmatically:

   ```go
   maa.Init(maa.WithLibDir("path/to/MaaFramework/bin"))
   ```

2. **Working Directory** - Place MaaFramework libraries in your program's working directory

3. **Environment Variables** - Add library path to `PATH` (Windows), `LD_LIBRARY_PATH` (Linux), or `DYLD_LIBRARY_PATH` (macOS)

4. **System Library Path** - Install libraries to system library directories

### MaaFramework Compatibility

This binding tracks the latest MaaFramework release, including prereleases; compatibility with older releases is not guaranteed. A missing library or symbol generally means the installed MaaFramework is older than the release this binding targets.

`Init` requires all four libraries from one compatible release: `MaaFramework`, `MaaToolkit`, `MaaAgentServer`, and `MaaAgentClient`. Missing libraries or symbols make `Init` fail with a diagnostic error instead of leaving partial state behind.

`Release` unloads the libraries only after all native objects are destroyed and the Agent Server is shut down; otherwise it returns `ErrLibraryInUse`.

After `AgentServerDetach`, `Release` remains blocked for the rest of the process: the native API cannot confirm that the detached service thread has exited, even after `AgentServerJoin` or `AgentServerShutDown`. Keep the service thread attached if you need to release the libraries.

Calls to `Init` and `Release` are serialized, but other MAA operations must not run concurrently with either. If unloading fails, `IsInited` becomes false; retry `Release` to finish cleanup before calling `Init` again.

## 🚀 Quick Start

```go
package main

import (
	"fmt"
	"os"

	"github.com/MaaXYZ/maa-framework-go/v4"
)

func main() {
	if err := maa.Init(); err != nil {
		fmt.Println("Failed to init MAA:", err)
		os.Exit(1)
	}
	if err := maa.ConfigInitOption("./", "{}"); err != nil {
		fmt.Println("Failed to init config:", err)
		os.Exit(1)
	}
	tasker, err := maa.NewTasker()
	if err != nil {
		fmt.Println("Failed to create tasker")
		os.Exit(1)
	}

	devices, err := maa.FindAdbDevices()
	if err != nil {
		fmt.Println("Failed to find adb devices:", err)
		os.Exit(1)
	}
	if len(devices) == 0 {
		fmt.Println("No ADB devices found. Connect a device or start an emulator.")
		os.Exit(1)
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
		fmt.Println("Failed to create ADB controller")
		os.Exit(1)
	}
	defer ctrl.Destroy()
	connectJob, err := ctrl.PostConnect()
	if err != nil {
		fmt.Println("Failed to connect controller:", err)
		os.Exit(1)
	}
	connectJob.Wait()
	tasker.BindController(ctrl)

	res, err := maa.NewResource()
	if err != nil {
		fmt.Println("Failed to create resource")
		os.Exit(1)
	}
	defer res.Destroy()
	bundleJob, err := res.PostBundle("./resource")
	if err != nil {
		fmt.Println("Failed to post resource bundle:", err)
		os.Exit(1)
	}
	bundleJob.Wait()
	tasker.BindResource(res)
	defer tasker.Destroy()
	if !tasker.Initialized() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	taskJob, err := tasker.PostTask("Startup")
	if err != nil {
		fmt.Println("Failed to post task:", err)
		os.Exit(1)
	}
	detail, err := taskJob.Wait().GetDetail()
	if err != nil {
		fmt.Println("Failed to get task detail:", err)
		os.Exit(1)
	}
	fmt.Println(detail)
}
```

### Native object lifetime

`NewTasker`, `NewResource`, and controller constructors return objects that own their native handles. `GetResource`, `GetController`, `Context.GetTasker`, and event callbacks return borrowed views; calling `Destroy` on one returns `ErrBorrowed`. Repeated successful `Destroy` calls on an owner are safe. `Destroy` returns `ErrInUse` if a call or asynchronous job is active, even if the returned Job was discarded; retry after it finishes. Call `Wait` before destroying the owner if you need the job's outcome. After closing, methods that return an error report `ErrClosed`, and jobs expose it through `Error()`.

Keep every resource and controller bound to a tasker alive until the tasker is destroyed, including earlier bindings after rebinding. Closing one before then returns `ErrBound`. Rebinding a running tasker returns `ErrTaskerRunning`. An `AgentClient` also keeps its bound resource and registered event sources alive until the client is destroyed. Destroying an owner from its callback returns `ErrInCallback`. A callback `Context`, including a clone, expires when the callback returns.

### Concurrency and callbacks

`Job` and `TaskJob` support concurrent `Wait`, status queries and predicates, and `Error`. Multiple waiters share the completed result, while status queries remain available during a wait. Do not copy these objects. `Tasker.GetResource` and `GetController` may run alongside binding changes; each returns a borrowed view of the binding observed during that call.

Other operations require caller coordination: serialize options, resource and pipeline changes, calls through a `Context`, and AgentClient/AgentServer lifecycle operations. Mutable configuration objects such as `Pipeline` and `Node` also require caller synchronization. Handle lifetime checks do not make all native operations thread-safe.

Change event sinks and custom recognition/action registrations only when the instance and all associated taskers are idle. Stop new submissions and let their work and callbacks finish before adding, replacing, removing, or clearing registrations. Configuration transactions on the same instance are serialized, including through borrowed views, but must not overlap native execution. Do not change registrations from a callback. `AddSink` and `AddContextSink` return 0 on registration failure; removing an already-removed sink is a no-op. `Resource.UnregisterCustomRecognition` and `Resource.UnregisterCustomAction` treat unregistered names as a no-op and return nil.

`Tasker.OnNodeWaitFreezesInContext` subscribes to `Node.WaitFreezes` with `NodeWaitFreezesDetail`. Custom context sinks can optionally implement `ContextWaitFreezesEventSink`. For Win32 background input, call `Controller.SetBackgroundManagedKeys` before connecting; an empty slice clears the key list.

Configure AgentServer before `AgentServerStartUp`. During service execution or after join or detach, custom registration and startup return `ErrInUse`, and adding a sink returns 0. Without prior detach, `AgentServerShutDown` permanently closes the server, even when called before startup: later startup and custom registration return `ErrClosed`, and adding a sink returns 0. The native communication context cannot be reset, so `Release` followed by `Init` does not restore startup or configuration. After an attached server shuts down, `Release` is allowed and repeated shutdown calls are no-ops. Configuration and startup are serialized; callers must serialize startup, shutdown, join, and detach with each other. After detach, shutdown cannot establish idleness and `Release` remains unavailable.

Callbacks execute synchronously on native calling threads and may overlap. Synchronize shared state in your handlers, and never wait for work that needs the current callback to return. Do not call AgentServer lifecycle methods from its callbacks. A successful `Destroy` returns after native cleanup and prevents subsequent user callbacks; active callbacks cause `ErrInCallback`. Custom controller `KeyUp` and `TouchUp` callbacks may run during native destruction, before `Destroy` returns.

After `PostStop`, an earlier job's `Wait` can return because its ID was invalidated while native work is still executing. It is not proof that all callbacks have ended. `Destroy` retains this uncertainty until the worker is idle. For a controller it may post an inactive action, invoke the custom `Inactive` handler, and return `ErrInUse` until the action completes. Retry destruction after active work and callbacks finish.

## 📖 Examples

For more examples, see the [examples](examples) directory:

- [quick-start](examples/quick-start) - Basic usage
- [custom-action](examples/custom-action) - Custom action
- [custom-recognition](examples/custom-recognition) - Custom recognition
- [agent-client](examples/agent-client) - Agent client
- [agent-server](examples/agent-server) - Agent server

## 📚 Documentation

- [MaaFramework Quick Start](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/1.1-QuickStarted.md)
- [Pipeline Protocol](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/3.1-PipelineProtocol.md)
- [Integration Guide](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/2.1-Integration.md)
- [Go Package Documentation](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4)

## 🤝 Contributing

Bug reports, feature suggestions, and pull requests are welcome.

## 📄 License

This project is licensed under the [LGPL-3.0 License](LICENSE.md).

## 💬 Community

- **QQ Group**: 595990173
- **GitHub Discussions**: [MaaFramework Discussions](https://github.com/MaaXYZ/MaaFramework/discussions)
