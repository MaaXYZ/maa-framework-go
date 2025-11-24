<!-- markdownlint-disable MD033 MD041 -->
<p align="center">
  <img alt="LOGO" src="https://cdn.jsdelivr.net/gh/MaaAssistantArknights/design@main/logo/maa-logo_512x512.png" width="256" height="256" />
</p>

# MaaFramework Golang 绑定

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

[English](README.md) | 简体中文

这是 [MaaFramework](https://github.com/MaaXYZ/MaaFramework) 的Go语言绑定，为Go开发者提供了一种简单而有效的方式，在他们的Go应用程序中使用MaaFramework的功能。

> 无需 Cgo！

## 安装

要安装MaaFramework Go绑定，请在终端中运行以下命令：

```shell
go get github.com/MaaXYZ/maa-framework-go/v3
```

此外，请下载MaaFramework的[Release 包](https://github.com/MaaXYZ/MaaFramework/releases)，以获取必要的动态库文件。

## 使用

要在您的Go项目中使用MaaFramework，请像导入其他Go包一样导入此包：

```go
import "github.com/MaaXYZ/maa-framework-go/v3"
```

然后，您可以使用MaaFramework提供的功能。有关详细用法，请参阅仓库中提供的 [文档](#文档)。

> 注意: 使用 maa-framework-go 构建的程序依赖于 MaaFramework 的动态库运行。请确保以下条件之一满足：
>
> 1. 程序的工作目录包含 MaaFramework 的动态库。
> 2. 设置了指向动态库的环境变量（如 LD_LIBRARY_PATH 或 PATH）。
>
> 否则，程序可能无法正确运行。

## 文档

目前没有太多详细的文档。请参阅源代码，并与MaaFramework项目中的接口进行比较，以了解如何使用这些绑定。我们正在积极添加更多注释和文档到源代码中。

以下是一些可能对您有帮助的MaaFramework文档：

- [快速开始](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/zh_cn/1.1-%E5%BF%AB%E9%80%9F%E5%BC%80%E5%A7%8B.md)
- [任务流水线协议](https://github.com/MaaXYZ/MaaFramework/blob/main/docs/zh_cn/3.1-%E4%BB%BB%E5%8A%A1%E6%B5%81%E6%B0%B4%E7%BA%BF%E5%8D%8F%E8%AE%AE.md)

## 贡献

我们欢迎对MaaFramework Go绑定的贡献。如果您发现了bug或有功能请求，请在GitHub仓库上打开一个issue。如果您想贡献代码，欢迎fork仓库并提交pull request。

## 许可证

本项目使用 LGPL-3.0 许可证。详细信息请参阅 [LICENSE](https://github.com/MaaXYZ/maa-framework-go/blob/main/LICENSE.md) 文件。

## 讨论

QQ 群: 595990173
