<p align="center">
  <img alt="LOGO" src="https://cdn.jsdelivr.net/gh/MaaAssistantArknights/design@main/logo/maa-logo_512x512.png" width="256" height="256" />
</p>

# MaaFramework Golang 绑定

<p>
	<a href="https://github.com/MaaXYZ/maa-framework-go/blob/main/LICENSE.md">
		<img alt="license" src="https://img.shields.io/github/license/MaaXYZ/maa-framework-go">
	</a>
	<a href="https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go">
		<img alt="go reference" src="https://pkg.go.dev/badge/github.com/MaaXYZ/maa-framework-go">
	</a>
</p>

[English](README.md) | 简体中文

这是 [MaaFramework](https://github.com/MaaXYZ/MaaFramework) 的Go语言绑定，为Go开发者提供了一种简单而有效的方式，在他们的Go应用程序中使用MaaFramework的功能。

## 安装

要安装MaaFramework Go绑定，请在终端中运行以下命令：

```shell
go get github.com/MaaXYZ/maa-framework-go
```

## 使用

要在您的Go项目中使用MaaFramework，请像导入其他Go包一样导入此包：

```go
import "github.com/MaaXYZ/maa-framework-go"
```

然后，您可以使用MaaFramework提供的功能。有关详细用法，请参阅仓库中提供的示例和文档。

## 文档

目前没有太多详细的文档。请参阅源代码，并与MaaFramework项目中的接口进行比较，以了解如何使用这些绑定。我们正在积极添加更多注释和文档到源代码中。

以下是一些可能对您有帮助的MaaFramework文档：

- [快速开始](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/zh_cn/1.1-%E5%BF%AB%E9%80%9F%E5%BC%80%E5%A7%8B.md)
- [任务流水线协议](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/zh_cn/3.1-%E4%BB%BB%E5%8A%A1%E6%B5%81%E6%B0%B4%E7%BA%BF%E5%8D%8F%E8%AE%AE.md)

## 平台特定说明

### Windows

在Windows上，MaaFramework的默认位置是 `C:\maa`。请确保MaaFramework安装在此目录，以便绑定能够开箱即用。

如果您需要指定自定义安装路径，请参阅 [自定义环境](#自定义环境) 部分。

### Linux 和 macOS

在Linux和macOS上，您需要创建一个名为 `maa.pc` 的 `pkg-config` 文件。此文件应正确指向MaaFramework头文件和库的位置。将此文件放在 `pkg-config` 可以找到的目录中（例如，`/usr/lib/pkgconfig`）。

一个示例 `maa.pc` 文件可能如下所示：

```
prefix=/path/to/maafw
exec_prefix=${prefix}
libdir=${exec_prefix}/lib
includedir=${prefix}/include

Name: MaaFramework
Description: MaaFramework library
Version: 1.0
Libs: -L${libdir} -lMaaFramework -lMaaToolkit
Cflags: -I${includedir}
```

如果您需要指定自定义环境，请参阅 [自定义环境](#自定义环境) 部分。

## 自定义环境

如果您需要为MaaFramework指定自定义安装路径，可以使用 `-tags customenv` 构建标记禁用默认位置。然后，设置必要的环境变量 `CGO_CFLAGS` 和 `CGO_LDFLAGS`。

```shell
go build -tags customenv
```

设置环境变量如下：

```shell
export CGO_CFLAGS="-I[path to maafw include directory]"
export CGO_LDFLAGS="-L[path to maafw lib directory] -lMaaFramework -lMaaToolkit"
```

将 `[path to maafw include directory]` 替换为MaaFramework包含目录的实际路径，将 `[path to maafw lib directory]` 替换为MaaFramework库目录的实际路径。

## 示例

### 快速开始

有关详细信息，请参阅 [quick-start](examples/quick-start)。

以下是一个基本示例，帮助您快速入门：

```go
package main

import (
	"fmt"
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/MaaXYZ/maa-framework-go/toolkit"
	"os"
)

func main() {
	toolkit.InitOption("./", "{}")
	inst := maa.New(nil)
	defer inst.Destroy()

	devices := toolkit.AdbDevices()
	device := devices[0]
	ctrl := maa.NewAdbController(
		device.AdbPath,
		device.Address,
		device.ControllerType,
		device.Config,
		"path/to/MaaAgentBinary",
		nil,
	)
	defer ctrl.Destroy()
	ctrl.PostConnect().Wait()
	inst.BindController(ctrl)

	res := maa.NewResource(nil)
	defer res.Destroy()
	res.PostPath("./resource").Wait()
	inst.BindResource(res)
	if inst.Inited() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	inst.PostTask("Startup", "{}")
}

```

### 自定义识别器

有关详细信息，请参阅 [custom-recognizer](examples/custom-recognizer)。

以下是一个实现自定义识别器的基本示例：

```go
package main

import (
	"fmt"
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/MaaXYZ/maa-framework-go/buffer"
	"github.com/MaaXYZ/maa-framework-go/toolkit"
	"image"
	"os"
)

func main() {
	toolkit.InitOption("./", "{}")
	inst := maa.New(nil)
	defer inst.Destroy()

	devices := toolkit.AdbDevices()
	device := devices[0]
	ctrl := maa.NewAdbController(
		device.AdbPath,
		device.Address,
		device.ControllerType,
		device.Config,
		"path/to/MaaAgentBinary",
		nil,
	)
	defer ctrl.Destroy()
	ctrl.PostConnect().Wait()
	inst.BindController(ctrl)

	res := maa.NewResource(nil)
	defer res.Destroy()
	res.PostPath("./resource").Wait()
	inst.BindResource(res)
	if inst.Inited() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	myRec := NewMyRec()
	defer myRec.Destroy()
	inst.RegisterCustomRecognizer("MyRec", myRec)

	inst.PostTask("Startup", "{}")
}

type MyRec struct {
	maa.CustomRecognizerHandler
}

func NewMyRec() maa.CustomRecognizer {
	return &MyRec{
		CustomRecognizerHandler: maa.NewCustomRecognizerHandler(),
	}
}

func (m MyRec) Analyze(syncCtx maa.SyncContext, img image.Image, taskName, RecognitionParam string) (maa.AnalyzeResult, bool) {
	return maa.AnalyzeResult{
		Box:    maa.Rect{0, 0, 100, 100},
		Detail: "Hello World!",
	}, true
}

```

### 自定义动作

有关详细信息，请参阅 [custom-action](examples/custom-action)。

以下是一个实现自定义动作的基本示例：

```go
package main

import (
	"fmt"
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/MaaXYZ/maa-framework-go/buffer"
	"github.com/MaaXYZ/maa-framework-go/toolkit"
	"os"
)

func main() {
	toolkit.InitOption("./", "{}")
	inst := maa.New(nil)
	defer inst.Destroy()

	devices := toolkit.AdbDevices()
	device := devices[0]
	ctrl := maa.NewAdbController(
		device.AdbPath,
		device.Address,
		device.ControllerType,
		device.Config,
		"path/to/MaaAgentBinary",
		nil,
	)
	defer ctrl.Destroy()
	ctrl.PostConnect().Wait()
	inst.BindController(ctrl)

	res := maa.NewResource(nil)
	defer res.Destroy()
	res.PostPath("./resource").Wait()
	inst.BindResource(res)
	if inst.Inited() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	myAct := NewAct()
	defer myAct.Destroy()
	inst.RegisterCustomAction("MyAct", myAct)

	inst.PostTask("Startup", "{}")
}

type MyAct struct {
	maa.CustomActionHandler
}

func NewAct() maa.CustomAction {
	return &MyAct{
		CustomActionHandler: maa.NewCustomActionHandler(),
	}
}

func (*MyAct) Run(ctx maa.SyncContext, taskName, ActionParam string, curBox buffer.Rect, curRecDetail string) bool {
	return true
}

func (*MyAct) Stop() {
}

```

## 贡献

我们欢迎对MaaFramework Go绑定的贡献。如果您发现了bug或有功能请求，请在GitHub仓库上打开一个issue。如果您想贡献代码，欢迎fork仓库并提交pull request。

## 许可证

本项目使用 LGPL-3.0 许可证。详细信息请参阅 [LICENSE](https://github.com/MaaXYZ/maa-framework-go/blob/main/LICENSE.md) 文件。

## 讨论

QQ 群: 595990173