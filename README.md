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
    <a href="https://github.com/MaaXYZ/MaaFramework/releases/tag/v5.14.0-beta.1">
      <img alt="maa framework" src="https://img.shields.io/badge/MaaFramework-v5.14.0--beta.1-blue">
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

## ✨ Features

- **Cross-platform Controllers** - ADB, Win32, Linux, macOS, PlayCover, and Android Native
- **Recording and Replay** - Capture controller operations to JSONL, then replay them for debugging and regression testing
- **Virtual Gamepad Controller** (Windows only) - Gamepad automation via ViGEm
- **Toolkit Utilities** - Find ADB devices and desktop windows; manage macOS automation permissions
- **Image Recognition** - Template matching, OCR, and feature detection
- **Custom Extensions** - Custom recognitions, actions, and controllers in pure Go
- **Agent Support** - Run custom recognition and action logic from an external process
- **Async Jobs and Events** - Poll job status and task details, or subscribe to resource, controller, and tasker events
- **Pipeline and Runtime APIs** - Declarative JSON task flows; run tasks, recognitions, and actions from a Context at runtime

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
	ctrl.PostConnect().Wait()
	tasker.BindController(ctrl)

	res, err := maa.NewResource()
	if err != nil {
		fmt.Println("Failed to create resource")
		os.Exit(1)
	}
	defer res.Destroy()
	res.PostBundle("./resource").Wait()
	tasker.BindResource(res)
	defer tasker.Destroy()
	if !tasker.Initialized() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	detail, err := tasker.PostTask("Startup").Wait().GetDetail()
	if err != nil {
		fmt.Println("Failed to get task detail:", err)
		os.Exit(1)
	}
	fmt.Println(detail)
}
```

### Native object lifetime

`NewTasker`, `NewResource`, and controller constructors return objects that own their native handles. `GetResource`, `GetController`, `Context.GetTasker`, and event callbacks return borrowed views; calling `Destroy` on one returns `ErrBorrowed`. Repeated `Destroy` calls on an owner are safe. After closing, methods that return an error report `ErrClosed`, and jobs expose it through `Error()`.

Keep a bound resource and controller alive until the tasker is destroyed. Closing either one while it is bound returns `ErrBound`. An `AgentClient` also keeps its bound resource and registered event sources alive until the client is destroyed. Destroying an owner from its callback returns `ErrInCallback`. A callback `Context`, including a clone, expires when the callback returns.

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
