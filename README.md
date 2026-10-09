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

> **No Cgo Required!** Pure Go implementation using [purego](https://github.com/ebitengine/purego). Android builds are the exception; see the [Android build guide](docs/en/guides/android.md).

## Features

- **Cross-platform Controllers** - ADB, Win32, Linux, macOS, PlayCover, and Android Native
- **Recording and Replay** - Capture controller operations to JSONL, then replay them for debugging and regression testing
- **Virtual Gamepad Controller** (Windows only) - Gamepad automation via ViGEm
- **Toolkit Utilities** - Find ADB devices and desktop windows; manage macOS automation permissions
- **Image Recognition** - Template matching, OCR, and feature detection
- **Custom Extensions** - Custom recognitions, actions, and controllers in pure Go
- **Agent Support** - Run custom recognition and action logic from an external process
- **Async Jobs and Events** - Poll job status and task details, or subscribe to resource, controller, and tasker events
- **Pipeline v2 model APIs** - Typed builders emit nested v2 JSON

## Requirements

- **Go 1.24 or later**
- **MaaFramework** - Tracks the latest MaaFramework release, including prereleases; compatibility with older releases is not guaranteed. The header badge records the MaaFramework release that passed unit tests before this binding's latest release; it is a tested baseline, not a version requirement.

## Installation

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

### 3. Load Runtime Libraries

Programs built with maa-framework-go require MaaFramework dynamic libraries at runtime. Two common loading methods are shown below; see the [runtime library loading guide](docs/en/guides/library-loading.md) for additional methods using environment variables or system library directories, and for the library file list.

1. **Library Directory** - Point `WithLibDir` at the directory containing the libraries:

   ```go
   if err := maa.Init(maa.WithLibDir("path/to/MaaFramework/bin")); err != nil {
       return err
   }
   ```

2. **Working Directory** - Place the libraries and their dependencies in the program's working directory and select it explicitly:

   ```go
   if err := maa.Init(maa.WithLibDir(".")); err != nil {
       return err
   }
   ```

## Quick Start

The core steps are shown below. See the [complete quick-start example](examples/quick-start/main.go) for runtime initialization, device discovery, object creation, and cleanup. Set the agent binary directory in `NewAdbController` and run from `examples/quick-start` so `./resource` resolves to the included bundle.

```go
// Initialize MAA, find an ADB device, and create ctrl (see the full example).
connectJob, err := ctrl.PostConnect()
if err != nil {
	return fmt.Errorf("post connect: %w", err)
}
if !connectJob.Wait().Success() {
	return errors.New("ADB controller connection failed")
}

// Create res before loading the resource bundle.
bundleJob, err := res.PostBundle("./resource")
if err != nil {
	return fmt.Errorf("post resource bundle: %w", err)
}
if !bundleJob.Wait().Success() {
	return errors.New("resource bundle loading failed")
}

// Create tasker after the controller is connected and the resource is loaded.
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
```

## Examples

For more examples, see the [examples](examples) directory:

- [quick-start](examples/quick-start) - Basic usage
- [custom-action](examples/custom-action) - Custom action
- [custom-recognition](examples/custom-recognition) - Custom recognition
- [agent-client](examples/agent-client) - Agent client
- [agent-server](examples/agent-server) - Agent server

## Documentation

- [Runtime library loading](docs/en/guides/library-loading.md)
- [Android build guide](docs/en/guides/android.md)
- [Frequently asked questions](docs/en/guides/faq.md)
- Migration guides: [v3 to v4](docs/en/migration/from-v3.md), [beta.18 to beta.19](docs/en/migration/from-v4.0.0-beta.18.md), [beta.19 to v4](docs/en/migration/from-v4.0.0-beta.19.md). The v4.0.0 guides are release preparation drafts based on a fixed commit; the final tag has not been created.
- [Changelog](CHANGELOG.md)
- [MaaFramework Quick Start](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/1.1-QuickStarted.md)
- [Pipeline Protocol](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/3.1-PipelineProtocol.md)
- [Integration Guide](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/2.1-Integration.md)
- [Go Package Documentation](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4)

## Contributing

Bug reports, feature suggestions, and pull requests are welcome. See the [contribution guide](CONTRIBUTING.md) for development setup, checks, and pull request guidelines.

## License

This project is licensed under the [LGPL-3.0 License](LICENSE.md).

## Community

- **QQ Group**: 595990173
- **GitHub Discussions**: [MaaFramework Discussions](https://github.com/MaaXYZ/MaaFramework/discussions)
