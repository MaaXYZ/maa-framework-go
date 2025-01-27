<!-- markdownlint-disable MD033 MD041 -->
<p align="center">
  <img alt="LOGO" src="https://cdn.jsdelivr.net/gh/MaaAssistantArknights/design@main/logo/maa-logo_512x512.png" width="256" height="256" />
</p>

# MaaFramework Golang Binding

<p>
    <a href="https://github.com/MaaXYZ/maa-framework-go/blob/main/LICENSE.md">
        <img alt="license" src="https://img.shields.io/github/license/MaaXYZ/maa-framework-go">
    </a>
    <a href="https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go">
        <img alt="go reference" src="https://pkg.go.dev/badge/github.com/MaaXYZ/maa-framework-go">
    </a>
    <a href="https://github.com/MaaXYZ/MaaFramework/releases/tag/v2.4.2">
        <img alt="maa framework" src="https://img.shields.io/badge/MaaFramework-v2.4.2-blue">
    </a>
</p>

English | [简体中文](README_zh.md)

This is the Go binding for [MaaFramework](https://github.com/MaaXYZ/MaaFramework), providing Go developers with a simple and effective way to use MaaFramework's features within their Go applications.

> No Cgo required!

## Installation

To install the MaaFramework Go binding, run the following command in your terminal:

```shell
go get github.com/MaaXYZ/maa-framework-go
```

In addition, please download the [Release package](https://github.com/MaaXYZ/MaaFramework/releases) for MaaFramework to get the necessary dynamic library files.

## Usage

To use MaaFramework in your Go project, import the package as you would with any other Go package:

```go
import "github.com/MaaXYZ/maa-framework-go"
```

Then, you can use the functionalities provided by MaaFramework. For detailed usage, refer to the [documentation](#documentation) and [examples](#examples) provided in the repository.

> Note: Programs built with maa-framework-go rely on the dynamic libraries of MaaFramework. Please ensure one of the following conditions is met:
>
> 1. The program's working directory contains the MaaFramework dynamic libraries.
> 2. Environment variables (such as LD_LIBRARY_PATH or PATH) are set to include the path to the dynamic libraries.
>
> Otherwise, the program may not run correctly.

## Documentation

Currently, there is not much detailed documentation available. Please refer to the source code and compare it with the interfaces in the original MaaFramework project to understand how to use the bindings. We are actively working on adding more comments and documentation to the source code.

Here are some documents from the maa framework that might help you:

- [QuickStarted](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/1.1-QuickStarted.md)
- [PipelineProtocol](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/3.1-PipelineProtocol.md)

## Examples

- [Quirk Start](#quirk-start)
- [Custom Recognition](#custom-recognition)
- [Custom Action](#custom-action)
- [PI CLI](#pi-cli)

### Quirk start

See [quirk-start](examples/quick-start) for details.

Here is a basic example to get you started:

```go
package main

import (
    "fmt"
    "github.com/MaaXYZ/maa-framework-go"
    "os"
)

func main() {
    toolkit := maa.NewToolkit()
    toolkit.ConfigInitOption("./", "{}")
    tasker := maa.NewTasker(nil)
    defer tasker.Destroy()

    device := toolkit.FindAdbDevices()[0]
    ctrl := maa.NewAdbController(
        device.AdbPath,
        device.Address,
        device.ScreencapMethod,
        device.InputMethod,
        device.Config,
        "path/to/MaaAgentBinary",
        nil,
    )
    defer ctrl.Destroy()
    ctrl.PostConnect().Wait()
    tasker.BindController(ctrl)

    res := maa.NewResource(nil)
    defer res.Destroy()
    res.PostPath("./resource").Wait()
    tasker.BindResource(res)
    if tasker.Initialized() {
        fmt.Println("Failed to init MAA.")
        os.Exit(1)
    }

    detail := tasker.PostPipeline("Startup").Wait().GetDetail()
    fmt.Println(detail)
}

```

### Custom Recognition

See [custom-recognition](examples/custom-recognition) for details.

Here is a basic example to implement your custom recognition:

```go
package main

import (
    "fmt"
    "github.com/MaaXYZ/maa-framework-go"
    "os"
)

func main() {
    toolkit := maa.NewToolkit()
    toolkit.ConfigInitOption("./", "{}")
    tasker := maa.NewTasker(nil)
    defer tasker.Destroy()

    device := toolkit.FindAdbDevices()[0]
    ctrl := maa.NewAdbController(
        device.AdbPath,
        device.Address,
        device.ScreencapMethod,
        device.InputMethod,
        device.Config,
        "path/to/MaaAgentBinary",
        nil,
    )
    defer ctrl.Destroy()
    ctrl.PostConnect().Wait()
    tasker.BindController(ctrl)

    res := maa.NewResource(nil)
    defer res.Destroy()
    res.PostPath("./resource").Wait()
    tasker.BindResource(res)
    if tasker.Initialized() {
        fmt.Println("Failed to init MAA.")
        os.Exit(1)
    }

    res.RegisterCustomRecognition("MyRec", &MyRec{})

    detail := tasker.PostPipeline("Startup").Wait().GetDetail()
    fmt.Println(detail)
}

type MyRec struct{}

func (r *MyRec) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (maa.CustomRecognitionResult, bool) {
    ctx.RunRecognition("MyCustomOCR", arg.Img, maa.J{
        "MyCustomOCR": maa.J{
            "roi": []int{100, 100, 200, 300},
        },
    })

    ctx.OverridePipeline(maa.J{
        "MyCustomOCR": maa.J{
            "roi": []int{1, 1, 114, 514},
        },
    })

    newContext := ctx.Clone()
    newContext.OverridePipeline(maa.J{
        "MyCustomOCR": maa.J{
            "roi": []int{100, 200, 300, 400},
        },
    })
    newContext.RunPipeline("MyCustomOCR", arg.Img)

    clickJob := ctx.GetTasker().GetController().PostClick(10, 20)
    clickJob.Wait()

    ctx.OverrideNext(arg.CurrentTaskName, []string{"TaskA", "TaskB"})

    return maa.CustomRecognitionResult{
        Box:    maa.Rect{0, 0, 100, 100},
        Detail: "Hello World!",
    }, true
}

```

### Custom Action

See [custom-action](examples/custom-action) for details.

Here is a basic example to implement your custom action:

```go
package main

import (
    "fmt"
    "github.com/MaaXYZ/maa-framework-go"
    "os"
)

func main() {
    toolkit := maa.NewToolkit()
    toolkit.ConfigInitOption("./", "{}")
    tasker := maa.NewTasker(nil)
    defer tasker.Destroy()

    device := toolkit.FindAdbDevices()[0]
    ctrl := maa.NewAdbController(
        device.AdbPath,
        device.Address,
        device.ScreencapMethod,
        device.InputMethod,
        device.Config,
        "path/to/MaaAgentBinary",
        nil,
    )
    defer ctrl.Destroy()
    ctrl.PostConnect().Wait()
    tasker.BindController(ctrl)

    res := maa.NewResource(nil)
    defer res.Destroy()
    res.PostPath("./resource").Wait()
    tasker.BindResource(res)
    if tasker.Initialized() {
        fmt.Println("Failed to init MAA.")
        os.Exit(1)
    }

    res.RegisterCustomAction("MyAct", &MyAct{})

    detail := tasker.PostPipeline("Startup").Wait().GetDetail()
    fmt.Println(detail)
}

type MyAct struct{}

func (a *MyAct) Run(_ *maa.Context, _ *maa.CustomActionArg) bool {
    return true
}

```

### PI CLI

See [pi-cli](examples/pi-cli) for details.

Here is a basic example of using PI CLI:

```go
package main

import (
    "github.com/MaaXYZ/maa-framework-go"
)

func main() {
    toolkit := maa.NewToolkit()
    toolkit.RegisterPICustomAction(0, "MyAct", &MyAct{})
    toolkit.RunCli(0, "./resource", "./", false, nil)
}

type MyAct struct{}

func (m MyAct) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
    ctx.OverrideNext(arg.CurrentTaskName, []string{"TaskA", "TaskB"})

    img := ctx.GetTasker().GetController().CacheImage()
    ctx.GetTasker().GetController().PostClick(100, 100).Wait()

    ctx.RunRecognition("Cat", img, maa.J{
        "recognition": "OCR",
        "expected":    "cat",
    })
    return true
}

```

## Contributing

We welcome contributions to the MaaFramework Go binding. If you find a bug or have a feature request, please open an issue on the GitHub repository. If you want to contribute code, feel free to fork the repository and submit a pull request.

## License

This project is licensed under the LGPL-3.0 License. See the [LICENSE](https://github.com/MaaXYZ/maa-framework-go/blob/main/LICENSE.md) file for details.

## Discussion

QQ Group: 595990173
