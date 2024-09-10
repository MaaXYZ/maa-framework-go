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
</p>

English | [简体中文](README_zh.md)


This is the Go binding for [MaaFramework](https://github.com/MaaXYZ/MaaFramework), providing Go developers with a simple and effective way to use MaaFramework's features within their Go applications.

## Installation

To install the MaaFramework Go binding, run the following command in your terminal:

```shell
go get github.com/MaaXYZ/maa-framework-go
```

## Usage

To use MaaFramework in your Go project, import the package as you would with any other Go package:

```go
import "github.com/MaaXYZ/maa-framework-go"
```

Then, you can use the functionalities provided by MaaFramework. For detailed usage, refer to the examples and documentation provided in the repository.

## Documentation

Currently, there is not much detailed documentation available. Please refer to the source code and compare it with the interfaces in the original MaaFramework project to understand how to use the bindings. We are actively working on adding more comments and documentation to the source code.

Here are some documents from the maa framework that might help you:

- [QuickStarted](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/1.1-QuickStarted.md)
- [PipelineProtocol](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/en_us/3.1-PipelineProtocol.md)

## Platform-Specific Notes

### Windows

On Windows, the default location for MaaFramework is `C:\maa`. Ensure that MaaFramework is installed in this directory for the binding to work out of the box.

If you need to specify a custom installation path, refer to the [Custom Environment](#custom-environment) section.

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

If you need to specify a custom environment, refer to the [Custom Environment](#custom-environment) section.

## Custom Environment

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
	toolkit.ConfigInitOption("./", "{}")
	tasker := maa.New(nil)
	defer tasker.Destroy()

	deviceFinder := toolkit.NewAdbDeviceFinder()
	deviceFinder.Find()
	device := deviceFinder.List()[0]
	ctrl := maa.NewAdbController(
		device.GetAdbPath(),
		device.GetAddress(),
		device.GetScreencapMethod(),
		device.GetInputMethod(),
		device.GetConfig(),
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
	if tasker.Inited() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	tasker.PostPipeline("Startup", "{}")
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
	"github.com/MaaXYZ/maa-framework-go/toolkit"
	"image"
	"os"
)

func main() {
	toolkit.ConfigInitOption("./", "{}")
	tasker := maa.New(nil)
	defer tasker.Destroy()

	deviceFinder := toolkit.NewAdbDeviceFinder()
	deviceFinder.Find()
	device := deviceFinder.List()[0]
	ctrl := maa.NewAdbController(
		device.GetAdbPath(),
		device.GetAddress(),
		device.GetScreencapMethod(),
		device.GetInputMethod(),
		device.GetConfig(),
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
	if tasker.Inited() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	res.RegisterCustomRecognizer("MyRec", &MyRec{})

	tasker.PostPipeline("Startup", "{}")
}

type MyRec struct{}

func (r *MyRec) Run(_ *maa.Context, _ int64, _, _ string, _ image.Image) (maa.CustomRecognizerResult, bool) {
	return maa.CustomRecognizerResult{
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
	"github.com/MaaXYZ/maa-framework-go/toolkit"
	"os"
)

func main() {
	toolkit.ConfigInitOption("./", "{}")
	tasker := maa.New(nil)
	defer tasker.Destroy()

	deviceFinder := toolkit.NewAdbDeviceFinder()
	deviceFinder.Find()
	device := deviceFinder.List()[0]
	ctrl := maa.NewAdbController(
		device.GetAdbPath(),
		device.GetAddress(),
		device.GetScreencapMethod(),
		device.GetInputMethod(),
		device.GetConfig(),
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
	if tasker.Inited() {
		fmt.Println("Failed to init MAA.")
		os.Exit(1)
	}

	res.RegisterCustomAction("MyAct", &MyAct{})

	tasker.PostPipeline("Startup", "{}")
}

type MyAct struct{}

func (a *MyAct) Run(_ *maa.Context, _ int64, _, _ string, _ maa.Rect, _ string) bool {
	return true
}

```

## Contributing

We welcome contributions to the MaaFramework Go binding. If you find a bug or have a feature request, please open an issue on the GitHub repository. If you want to contribute code, feel free to fork the repository and submit a pull request.

## License

This project is licensed under the LGPL-3.0 License. See the [LICENSE](https://github.com/MaaXYZ/maa-framework-go/blob/main/LICENSE.md) file for details.

## Discussion

QQ Group: 595990173