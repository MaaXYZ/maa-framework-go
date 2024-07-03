<!-- markdownlint-disable MD033 MD041 -->
<p align="center">
  <img alt="LOGO" src="https://cdn.jsdelivr.net/gh/MaaAssistantArknights/design@main/logo/maa-logo_512x512.png" width="256" height="256" />
</p>

# MaaFramework Golang Binding

This is the Go binding for MaaFramework, providing Go developers with a simple and effective way to use MaaFramework's features within their Go applications. Currently, the Go binding is quite rudimentary and closely mirrors the C interface. Future updates will include significant revisions to improve usability and functionality.

## Installation

To install the MaaFramework Go binding, run the following command in your terminal:

```shell
go get github.com/MaaXYZ/maa-framework-go
```

## Platform-Specific Notes

### Windows

On Windows, the default location for MaaFramework is `C:\maa`. Ensure that MaaFramework is installed in this directory for the binding to work out of the box.

### Linux and macOS

On Linux and macOS, you will need to create a `pkg-config` file named `maa.pc`. This file should correctly point to the locations of the MaaFramework headers and libraries. Place this file in a directory where `pkg-config` can find it (e.g., `/usr/lib/pkgconfig`).

A sample `maa.pc` file might look like this:

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

## Custom Installation Path
If you need to specify a custom installation path for MaaFramework, you can disable the default location using the `-tags customenv` build tag. Then, set the necessary environment variables `CGO_CFLAGS` and `CGO_LDFLAGS`.

```shell
go build -tags customenv
```

Set the environment variables as follows:

```shell
export CGO_CFLAGS="-I[path to maafw include directory]"
export CGO_LDFLAGS="-L[path to maafw lib directory] -lMaaFramework -lMaaToolkit"
```
Replace `[path to maafw include directory]` with the actual path to the MaaFramework include directory and `[path to maafw lib directory]` with the actual path to the MaaFramework library directory.

## Usage

To use MaaFramework in your Go project, import the package as you would with any other Go package:

```go
import "github.com/MaaXYZ/maa-framework-go"
```

Then, you can use the functionalities provided by MaaFramework. For detailed usage, refer to the examples and documentation provided in the repository.


## Examples

Here is a basic example to get you started:

```go
package main

import (
	"fmt"
	"github.com/MaaXYZ/maa-framework-go"
	"github.com/MaaXYZ/maa-framework-go/toolkit"
)

func main() {
	toolkit.InitOption("./", "{}")

	res := maa.NewResource(nil)
	defer res.Destroy()
	resId := res.PostPath("sample/resource")
	res.Wait(resId)

	devices := toolkit.AdbDevices()
	if len(devices) == 0 {
		fmt.Println("No Adb device found.")
		return
	}

	device := devices[0]
	ctrl := maa.NewAdbController(
		device.AdbPath,
		device.Address,
		device.ControllerType,
		device.Config,
		"sample/MaaAgentBinary",
		nil,
	)
	defer ctrl.Destroy()
	ctrlId := ctrl.PostConnect()
	ctrl.Wait(ctrlId)

	inst := maa.New(nil)
	defer inst.Destroy()

	inst.BindResource(res)
	inst.BindController(ctrl)

	inst.RegisterCustomRecognizer("MyRec", NewMyRec())
	inst.RegisterCustomAction("MyAct", NewMyAct())

	if !inst.Inited() {
		panic("Failed to init Maa Instance.")
	}

	taskId := inst.PostTask("TaskA", "{}")
	inst.WaitTask(taskId)
}

type MyRec struct {
	maa.CustomRecognizerHandler
}

func NewMyRec() MyRec {
	return MyRec{
		CustomRecognizerHandler: maa.NewCustomRecognizerHandler(),
	}
}

func (MyRec) Analyze(
	syncCtx maa.SyncContext,
	image maa.ImageBuffer,
	taskName string,
	customRecognitionParam string,
) (maa.AnalyzeResult, bool) {
	outBox := maa.NewRect()
	outBox.Set(0, 0, 100, 100)
	return maa.AnalyzeResult{
		Box:    outBox,
		Detail: "Hello world.",
	}, true
}

type MyAct struct {
	maa.CustomActionHandler
}

func NewMyAct() MyAct {
	return MyAct{
		CustomActionHandler: maa.NewCustomActionHandler(),
	}
}

func (MyAct) Run(
	ctx maa.SyncContext,
	taskName string,
	customActionParam string,
	curBox maa.RectBuffer,
	curRecDetail string,
) bool {
	return true
}

func (MyAct) Stop() {
}
```

## Documentation

Currently, there is no detailed documentation available. Please refer to the source code and compare it with the interfaces in the original MaaFramework project to understand how to use the bindings. We are actively working on adding more comments and documentation to the source code.

## Contributing

We welcome contributions to the MaaFramework Go binding. If you find a bug or have a feature request, please open an issue on the GitHub repository. If you want to contribute code, feel free to fork the repository and submit a pull request.

## License

This project is licensed under the LGPL-3.0 License. See the [LICENSE](https://github.com/MaaXYZ/maa-framework-go/blob/main/LICENSE.md) file for details.

## Discussion

QQ Group: 595990173