<!-- markdownlint-disable MD033 MD041 -->
<p align="center">
  <img alt="LOGO" src="https://cdn.jsdelivr.net/gh/MaaAssistantArknights/design@main/logo/maa-logo_512x512.png" width="256" height="256" />
</p>

<h1 align="center">MaaFramework Go 绑定</h1>

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
  <a href="README.md">English</a> | 简体中文
</p>

[MaaFramework](https://github.com/MaaXYZ/MaaFramework) 的 Go 语言绑定。MaaFramework 是一个基于图像识别的跨平台自动化测试框架。

> **🚀 无需 Cgo！** 基于 [purego](https://github.com/ebitengine/purego) 的纯 Go 实现。

## ✨ 特性

- **跨平台控制器** - ADB、Win32、Linux、macOS、PlayCover 与 Android Native
- **录制与回放** - 将控制器操作记录为 JSONL，并用于调试与回归测试
- **虚拟手柄控制器**（仅 Windows）- 通过 ViGEm 实现手柄自动化
- **Toolkit 工具能力** - 发现 ADB 设备和桌面窗口，管理 macOS 自动化权限
- **图像识别** - 模板匹配、OCR、特征检测
- **自定义扩展** - 以纯 Go 编写自定义识别、动作和控制器
- **Agent 支持** - 从外部进程执行自定义识别与动作逻辑
- **异步任务与事件** - 轮询 Job 状态与任务详情，或订阅 Resource、Controller、Tasker 事件
- **流水线与运行时 API** - 基于 JSON 的声明式任务流；可在 Context 中动态执行任务、识别和动作

## 📦 安装

需要 Go 1.24 及以上。

### 1. 安装 Go 包

```shell
go get github.com/MaaXYZ/maa-framework-go/v4
```

### 2. 下载 MaaFramework

根据你的平台下载 [MaaFramework Release](https://github.com/MaaXYZ/MaaFramework/releases) 并解压。

| 平台 | 架构 | 下载 |
|------|------|------|
| Windows  | amd64       | `MAA-win-x86_64-*.zip` |
| Windows  | arm64       | `MAA-win-aarch64-*.zip` |
| Linux    | amd64       | `MAA-linux-x86_64-*.zip` |
| Linux    | arm64       | `MAA-linux-aarch64-*.zip` |
| macOS    | amd64       | `MAA-macos-x86_64-*.zip` |
| macOS    | arm64       | `MAA-macos-aarch64-*.zip` |
| Android  | amd64       | `MAA-android-x86_64-*.zip` |
| Android  | arm64       | `MAA-android-aarch64-*.zip` |

## ⚙️ 运行时要求

使用 maa-framework-go 构建的程序需要 MaaFramework 动态库才能运行。可通过以下任一方式提供：

1. **通过 `Init()` 选项** - 在代码中指定库文件路径：

   ```go
   maa.Init(maa.WithLibDir("path/to/MaaFramework/bin"))
   ```

2. **工作目录** - 将 MaaFramework 库文件放在程序的工作目录中

3. **环境变量** - 将库文件路径添加到 `PATH`（Windows）、`LD_LIBRARY_PATH`（Linux）或 `DYLD_LIBRARY_PATH`（macOS）

4. **系统库路径** - 将库文件安装到系统库目录

### MaaFramework 兼容性

本绑定跟踪最新的 MaaFramework 版本（包括预发布版），不保证兼容更早的版本。缺少库或符号通常意味着已安装的 MaaFramework 早于本绑定所跟踪的版本。

`Init` 需要来自同一兼容版本的四个库：`MaaFramework`、`MaaToolkit`、`MaaAgentServer` 和 `MaaAgentClient`。缺少库或符号时，`Init` 会返回可诊断的错误，而不会遗留部分初始化状态。

仅当所有原生对象均已销毁且 Agent Server 已关闭后，`Release` 才会卸载库；否则返回 `ErrLibraryInUse`。

## 🚀 快速开始

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

### 原生对象生命周期

`NewTasker`、`NewResource` 以及控制器构造函数返回的对象拥有其原生句柄。`GetResource`、`GetController`、`Context.GetTasker` 以及事件回调返回的是借用视图；对其调用 `Destroy` 会返回 `ErrBorrowed`。对所有者重复成功调用 `Destroy` 是安全的。如果仍有调用或异步 Job 执行，`Destroy` 会返回 `ErrInUse`，即使返回的 Job 已被丢弃；请在其结束后重试。如果需要取得 Job 的结果，请在销毁所有者之前调用 `Wait`。对象关闭后，会返回错误的方法将报告 `ErrClosed`，Job 则通过 `Error()` 暴露该错误。

在 tasker 销毁之前，请保持所有曾绑定的 Resource 和 Controller 存活，包括重新绑定后替换下来的对象。提前关闭会返回 `ErrBound`。运行中的 tasker 重新绑定会返回 `ErrTaskerRunning`。`AgentClient` 同样会保持其绑定的 Resource 与已注册的事件源存活，直到该 client 被销毁。在回调中销毁其所属对象会返回 `ErrInCallback`。回调的 `Context`（包括克隆）会在回调返回时失效。

## 📖 示例

更多示例请查看 [examples](examples) 目录：

- [quick-start](examples/quick-start) - 基础使用
- [custom-action](examples/custom-action) - 自定义动作
- [custom-recognition](examples/custom-recognition) - 自定义识别
- [agent-client](examples/agent-client) - Agent 客户端
- [agent-server](examples/agent-server) - Agent 服务端

## 📚 文档

- [MaaFramework 快速开始](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/zh_cn/1.1-%E5%BF%AB%E9%80%9F%E5%BC%80%E5%A7%8B.md)
- [任务流水线协议](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/zh_cn/3.1-%E4%BB%BB%E5%8A%A1%E6%B5%81%E6%B0%B4%E7%BA%BF%E5%8D%8F%E8%AE%AE.md)
- [集成文档](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/zh_cn/2.1-%E9%9B%86%E6%88%90%E6%96%87%E6%A1%A3.md)
- [Go 包文档](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4)

## 🤝 贡献

欢迎提交 Issue 与 Pull Request。

## 📄 许可证

本项目采用 [LGPL-3.0 许可证](LICENSE.md)。

## 💬 社区

- **QQ 群**: 595990173
- **GitHub Discussions**: [MaaFramework Discussions](https://github.com/MaaXYZ/MaaFramework/discussions)
