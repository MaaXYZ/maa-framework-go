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
  <a href="README.md">English</a> | 简体中文
</p>

[MaaFramework](https://github.com/MaaXYZ/MaaFramework) 的 Go 语言绑定。MaaFramework 是一个基于图像识别的跨平台自动化测试框架。

> **无需 Cgo！** 基于 [purego](https://github.com/ebitengine/purego) 的纯 Go 实现。Android 构建是例外，详见 [Android 构建指南](docs/zh/guides/android.md)。

## 特性

- **跨平台控制器** - ADB、Win32、Linux、macOS、PlayCover 与 Android Native
- **录制与回放** - 将控制器操作记录为 JSONL，并用于调试与回归测试
- **虚拟手柄控制器**（仅 Windows）- 通过 ViGEm 实现手柄自动化
- **Toolkit 工具能力** - 发现 ADB 设备和桌面窗口，管理 macOS 自动化权限
- **图像识别** - 模板匹配、OCR、特征检测
- **自定义扩展** - 以纯 Go 编写自定义识别、动作和控制器
- **Agent 支持** - 从外部进程执行自定义识别与动作逻辑
- **异步任务与事件** - 轮询 Job 状态与任务详情，或订阅 Resource、Controller、Tasker 事件
- **Pipeline v2 模型 API** - 类型化构造器生成嵌套 v2 JSON

## 要求

- **Go 1.24 及以上**
- **MaaFramework** - 本绑定跟踪最新的 MaaFramework 版本（包括预发布版），不保证兼容更早的版本。头部徽章记录的是本绑定发布前通过单元测试的 MaaFramework 版本，是验证基线，不是版本要求。

## 安装

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

### 3. 加载运行时库

使用 maa-framework-go 构建的程序需要 MaaFramework 动态库才能运行。这里列出两种常用的加载方式；环境变量、系统库目录等其他方式及库文件清单见[运行时库加载指南](docs/zh/guides/library-loading.md)。

1. **库目录** - 通过 `WithLibDir` 指定包含库文件的目录：

   ```go
   if err := maa.Init(maa.WithLibDir("path/to/MaaFramework/bin")); err != nil {
       return err
   }
   ```

2. **工作目录** - 将库及其依赖放入程序的工作目录，并显式选择该目录：

   ```go
   if err := maa.Init(maa.WithLibDir(".")); err != nil {
       return err
   }
   ```

## 快速开始

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

## 示例

更多示例请查看 [examples](examples) 目录：

- [quick-start](examples/quick-start) - 基础使用
- [custom-action](examples/custom-action) - 自定义动作
- [custom-recognition](examples/custom-recognition) - 自定义识别
- [agent-client](examples/agent-client) - Agent 客户端
- [agent-server](examples/agent-server) - Agent 服务端

## 文档

- [运行时库加载](docs/zh/guides/library-loading.md)
- [Android 构建指南](docs/zh/guides/android.md)
- [常见问题](docs/zh/guides/faq.md)
- 迁移指南：[v3 → v4](docs/zh/migration/from-v3.md)、[beta.18 → beta.19](docs/zh/migration/from-v4.0.0-beta.18.md)、[beta.19 → v4](docs/zh/migration/from-v4.0.0-beta.19.md)。v4.0.0 指南是基于固定提交的发布准备草稿，正式版 tag 尚未创建。
- [变更记录](CHANGELOG.md)
- [MaaFramework 快速开始](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/zh_cn/1.1-%E5%BF%AB%E9%80%9F%E5%BC%80%E5%A7%8B.md)
- [任务流水线协议](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/zh_cn/3.1-%E4%BB%BB%E5%8A%A1%E6%B5%81%E6%B0%B4%E7%BA%BF%E5%8D%8F%E8%AE%AE.md)
- [集成文档](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/zh_cn/2.1-%E9%9B%86%E6%88%90%E6%96%87%E6%A1%A3.md)
- [Go 包文档](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4)

## 贡献

欢迎提交问题报告、功能建议和 Pull Request。开发环境、检查命令和 Pull Request 要求见[贡献指南](docs/zh/contributing.md)。

## 许可证

本项目采用 [LGPL-3.0 许可证](LICENSE.md)。

## 社区

- **QQ 群**: 595990173
- **GitHub Discussions**: [MaaFramework Discussions](https://github.com/MaaXYZ/MaaFramework/discussions)
