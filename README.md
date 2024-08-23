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

### Quirk start

See [quirk-start](examples/quick-start) for details.

Here is a basic example to get you started:

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

### Custom Recognizer

See [custom-recognizer](examples/custom-recognizer) for details.

Here is a basic example to implement your custom recognizer:

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
		Box:    buffer.Rect{0, 0, 100, 100},
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

## Documentation

Currently, there is no detailed documentation available. Please refer to the source code and compare it with the interfaces in the original MaaFramework project to understand how to use the bindings. We are actively working on adding more comments and documentation to the source code.

## Contributing

We welcome contributions to the MaaFramework Go binding. If you find a bug or have a feature request, please open an issue on the GitHub repository. If you want to contribute code, feel free to fork the repository and submit a pull request.

## License

This project is licensed under the LGPL-3.0 License. See the [LICENSE](https://github.com/MaaXYZ/maa-framework-go/blob/main/LICENSE.md) file for details.

## Discussion

QQ Group: 595990173