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
    <a href="https://github.com/MaaXYZ/MaaFramework/releases/tag/v5.8.0-beta.1">
      <img alt="maa framework" src="https://img.shields.io/badge/MaaFramework-v5.8.0--beta.1-blue">
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

- **ADB 控制器** - 通过 ADB 实现 Android 设备自动化
- **Win32 控制器** - Windows 桌面应用自动化
- **PlayCover 控制器** - 在 macOS 上控制通过 PlayCover 运行的 iOS 应用
- **虚拟手柄控制器** - 通过 ViGEm 模拟手柄输入（仅 Windows）
- **图像识别** - 模板匹配、OCR、特征检测等
- **自定义识别** - 实现自定义图像识别算法
- **自定义动作** - 定义你自己的自动化逻辑
- **Agent 支持** - 支持从外部进程挂载自定义识别和动作
- **流水线驱动** - 基于 JSON 配置的声明式任务流

## 📦 安装

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
| Linux    | arm64      | `MAA-linux-aarch64-*.zip` |
| macOS    | amd64       | `MAA-macos-x86_64-*.zip` |
| macOS    | arm64      | `MAA-macos-aarch64-*.zip` |

## ⚙️ 运行时要求

使用 maa-framework-go 构建的程序需要 MaaFramework 动态库才能运行。你有以下几种方式：

1. **通过 `Init()` 选项** - 在代码中指定库文件路径：

   ```go
   maa.Init(maa.WithLibDir("path/to/MaaFramework/bin"))
   ```

2. **工作目录** - 将 MaaFramework 库文件放在程序的工作目录中

3. **环境变量** - 将库文件路径添加到 `PATH`（Windows）或 `LD_LIBRARY_PATH`（Linux/macOS）

4. **系统库路径** - 将库文件安装到系统库目录

## 🚀 快速开始

```go
package main

import (
	"fmt"
	"os"

	"github.com/MaaXYZ/maa-framework-go/v4"
)

func main() {
	maa.Init()
	if err := maa.ConfigInitOption("./", "{}"); err != nil {
		fmt.Println("Failed to init config:", err)
		os.Exit(1)
	}
	tasker, err := maa.NewTasker()
	if err != nil {
		fmt.Println("Failed to create tasker")
		os.Exit(1)
	}
	defer tasker.Destroy()

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
- [Go 包文档](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v3)

## 🤝 贡献

欢迎贡献！你可以：

- 通过 Issue 报告 Bug
- 提出功能建议或改进意见
- 提交 Pull Request

## 📄 许可证

本项目采用 [LGPL-3.0 许可证](LICENSE.md)。

## 💬 社区

- **QQ 群**: 595990173
- **GitHub Discussions**: [MaaFramework Discussions](https://github.com/MaaXYZ/MaaFramework/discussions)
