<!-- markdownlint-disable MD033 MD041 -->
<p align="center">
  <img alt="LOGO" src="https://cdn.jsdelivr.net/gh/MaaAssistantArknights/design@main/logo/maa-logo_512x512.png" width="256" height="256" />
</p>

<h1 align="center">MaaFramework Go Binding</h1>

<p align="center">
    <a href="https://github.com/MaaXYZ/maa-framework-go/blob/main/LICENSE.md">
        <img alt="license" src="https://img.shields.io/github/license/MaaXYZ/maa-framework-go">
    </a>
    <a href="https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v3">
        <img alt="go reference" src="https://pkg.go.dev/badge/github.com/MaaXYZ/maa-framework-go">
    </a>
    <a href="https://github.com/MaaXYZ/MaaFramework/releases/tag/v5.2.6">
        <img alt="maa framework" src="https://img.shields.io/badge/MaaFramework-v5.2.6-blue">
    </a>
    <a href="https://goreportcard.com/report/github.com/MaaXYZ/maa-framework-go/v3">
        <img alt="go report" src="https://goreportcard.com/badge/github.com/MaaXYZ/maa-framework-go/v3">
    </a>
</p>

<p align="center">
    English | <a href="README_zh.md">简体中文</a>
</p>

Go binding for [MaaFramework](https://github.com/MaaXYZ/MaaFramework), a cross-platform automation testing framework based on image recognition.

> **🚀 No Cgo Required!** Pure Go implementation using [purego](https://github.com/ebitengine/purego).

## ✨ Features

- **ADB Controller** - Android device automation via ADB
- **Win32 Controller** - Windows desktop application automation
- **Image Recognition** - Template matching, OCR, feature detection and more
- **Custom Recognition** - Implement custom image recognition algorithms
- **Custom Actions** - Define your own automation logic
- **Agent Support** - Mount custom recognition and actions from external processes
- **Pipeline-based** - Declarative task flow with JSON configuration

## 📦 Installation

### 1. Install Go Package

```shell
go get github.com/MaaXYZ/maa-framework-go/v3
```

### 2. Download MaaFramework

Download the [MaaFramework Release](https://github.com/MaaXYZ/MaaFramework/releases) for your platform and extract it.

| Platform | Architecture | Download |
|----------|--------------|----------|
| Windows  | amd64       | `MAA-win-x86_64-*.zip` |
| Windows  | arm64       | `MAA-win-aarch64-*.zip` |
| Linux    | amd64       | `MAA-linux-x86_64-*.zip` |
| Linux    | arm64      | `MAA-linux-aarch64-*.zip` |
| macOS    | amd64       | `MAA-macos-x86_64-*.zip` |
| macOS    | arm64      | `MAA-macos-aarch64-*.zip` |

## ⚙️ Runtime Requirements

Programs built with maa-framework-go require MaaFramework dynamic libraries at runtime. You have several options:

1. **Via `Init()` Option** - Specify library path programmatically:

   ```go
   maa.Init(maa.WithLibDir("path/to/MaaFramework/bin"))
   ```

2. **Working Directory** - Place MaaFramework libraries in your program's working directory

3. **Environment Variables** - Add library path to `PATH` (Windows) or `LD_LIBRARY_PATH` (Linux/macOS)

4. **System Library Path** - Install libraries to system library directories

## 🚀 Quick Start

```go
package main

import (
    "fmt"
    "os"

    "github.com/MaaXYZ/maa-framework-go/v3"
)

func main() {
    // Initialize MaaFramework
    maa.Init()
    maa.ConfigInitOption("./", "{}")

    // Create tasker
    tasker := maa.NewTasker()
    defer tasker.Destroy()

    // Find and connect to ADB device
    devices := maa.FindAdbDevices()
    if len(devices) == 0 {
        fmt.Println("No ADB device found")
        os.Exit(1)
    }
    device := devices[0]

    ctrl := maa.NewAdbController(
        device.AdbPath,
        device.Address,
        device.ScreencapMethod,
        device.InputMethod,
        device.Config,
        "path/to/MaaAgentBinary",
    )
    defer ctrl.Destroy()
    ctrl.PostConnect().Wait()
    tasker.BindController(ctrl)

    // Load resource
    res := maa.NewResource()
    defer res.Destroy()
    res.PostBundle("./resource").Wait()
    tasker.BindResource(res)

    if !tasker.Initialized() {
        fmt.Println("Failed to initialize MAA")
        os.Exit(1)
    }

    // Run task
    detail := tasker.PostTask("Startup").Wait().GetDetail()
    fmt.Println(detail)
}
```

## 📖 Examples

For more examples, see the [examples](examples) directory:

- [quick-start](examples/quick-start) - Basic usage
- [custom-action](examples/custom-action) - Custom action implementation
- [custom-recognition](examples/custom-recognition) - Custom recognition implementation
- [agent-client](examples/agent-client) - Agent client
- [agent-server](examples/agent-server) - Agent server

## 📚 Documentation

- [MaaFramework Quick Start](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/1.1-QuickStarted.md)
- [Pipeline Protocol](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/3.1-PipelineProtocol.md)
- [Integration Guide](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/2.1-Integration.md)
- [Go Package Documentation](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v3)

## 🤝 Contributing

Contributions are welcome! Feel free to:

- Report bugs by opening issues
- Suggest features or improvements
- Submit pull requests

## 📄 License

This project is licensed under the [LGPL-3.0 License](LICENSE.md).

## 💬 Community

- **QQ Group**: 595990173
- **GitHub Discussions**: [MaaFramework Discussions](https://github.com/MaaXYZ/MaaFramework/discussions)
