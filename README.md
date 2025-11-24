<!-- markdownlint-disable MD033 MD041 -->
<p align="center">
  <img alt="LOGO" src="https://cdn.jsdelivr.net/gh/MaaAssistantArknights/design@main/logo/maa-logo_512x512.png" width="256" height="256" />
</p>

# MaaFramework Golang Binding

<p>
    <a href="https://github.com/MaaXYZ/maa-framework-go/blob/main/LICENSE.md">
        <img alt="license" src="https://img.shields.io/github/license/MaaXYZ/maa-framework-go">
    </a>
    <a href="https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v3">
        <img alt="go reference" src="https://pkg.go.dev/badge/github.com/MaaXYZ/maa-framework-go">
    </a>
    <a href="https://github.com/MaaXYZ/MaaFramework/releases/tag/v5.0.5">
        <img alt="maa framework" src="https://img.shields.io/badge/MaaFramework-v5.0.5-blue">
    </a>
</p>

English | [简体中文](README_zh.md)

This is the Go binding for [MaaFramework](https://github.com/MaaXYZ/MaaFramework), providing Go developers with a simple and effective way to use MaaFramework's features within their Go applications.

> No Cgo required!

## Installation

To install the MaaFramework Go binding, run the following command in your terminal:

```shell
go get github.com/MaaXYZ/maa-framework-go/v3
```

In addition, please download the [Release package](https://github.com/MaaXYZ/MaaFramework/releases) for MaaFramework to get the necessary dynamic library files.

## Usage

To use MaaFramework in your Go project, import the package as you would with any other Go package:

```go
import "github.com/MaaXYZ/maa-framework-go/v3"
```

Then, you can use the functionalities provided by MaaFramework. For detailed usage, refer to the [documentation](#documentation) provided in the repository.

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

## Contributing

We welcome contributions to the MaaFramework Go binding. If you find a bug or have a feature request, please open an issue on the GitHub repository. If you want to contribute code, feel free to fork the repository and submit a pull request.

## License

This project is licensed under the LGPL-3.0 License. See the [LICENSE](https://github.com/MaaXYZ/maa-framework-go/blob/main/LICENSE.md) file for details.

## Discussion

QQ Group: 595990173
