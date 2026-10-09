# 常见问题

[English](../../en/guides/faq.md) | 简体中文

本页处理库加载和清理问题。库文件名和部署方法见[运行时库加载](library-loading.md)。API 契约统一维护在 godoc 中；可访问下方 API 链接，或在项目中使用 `go doc` 查看已安装绑定版本的注释。

## Init 无法加载库

遇到 `LibraryLoadError` 时，检查 `LibraryName`、`LibraryPath` 和底层 `Err`：

1. 确认尝试加载的目录包含[全部四个库文件](library-loading.md#库文件)。显式指定目录时，传入库文件所在目录，而不是其父目录。
2. 确认发布包与程序的操作系统和架构一致。
3. 如果报错中的文件存在却仍无法加载，重新完整解压发布包，检查是否遗漏依赖。阅读底层系统错误，确认依赖名称或其他原因。
4. 使用环境变量时，在启动进程前设置。通过 IDE 或服务启动时，检查其工作目录和继承的环境，或改用绝对路径的 `WithLibDir`。

加载契约见 [Init](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Init) 和 [WithLibDir](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#WithLibDir)。

## Init 找不到必需的符号

遇到 `SymbolLookupError` 时，记录 `LibraryPath`、`SymbolName`、`Requirement` 和底层 `Err`。用同一个兼容发布包中的文件替换全部四个库；只替换报错所指的库可能仍会混用版本。检查环境变量或系统搜索路径中是否存在旧副本，使用显式库目录缩小实际加载文件的范围。

头部徽章表示测试基线；选择版本时参照[版本策略](../../../README_zh.md#要求)和 [Init](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Init) 的兼容性要求。

使用 `errors.As` 可识别这两类错误，包括被包装的错误。下面的辅助函数输出诊断字段，并将初始化错误与清理重试错误一并返回：

```go
func initialize(libDir string) error {
	if err := maa.Init(maa.WithLibDir(libDir)); err != nil {
		var loadErr *maa.LibraryLoadError
		var symbolErr *maa.SymbolLookupError
		switch {
		case errors.As(err, &loadErr):
			fmt.Fprintf(os.Stderr, "library=%s path=%s: %v\n",
				loadErr.LibraryName, loadErr.LibraryPath, loadErr.Err)
		case errors.As(err, &symbolErr):
			fmt.Fprintf(os.Stderr, "library=%s path=%s symbol=%s: %v; %s\n",
				symbolErr.LibraryName, symbolErr.LibraryPath,
				symbolErr.SymbolName, symbolErr.Err, symbolErr.Requirement)
		}
		return errors.Join(err, maa.Release())
	}
	return nil
}
```

需导入 `errors`、`fmt`、`os` 和 `github.com/MaaXYZ/maa-framework-go/v4`。初始化成功后，由调用方管理对象清理和最终的 `Release`。

## 初始化或释放报告卸载错误

保留完整错误，包括清理错误。再次调用 `Release` 并检查结果，再重试 `Init`；部分卸载后不要继续调用原生函数。如果清理持续失败，保留错误用于报告问题，并在重试前重启进程。失败清理与保留句柄的契约见 [Init](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Init) 和 [Release](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Release)。

## Release 返回 ErrLibraryInUse

使用 `errors.Is(err, maa.ErrLibraryInUse)` 识别该错误，然后检查程序使用的对象和 Agent Server：

1. 停止提交新工作，等待活动调用、Job 和回调结束。
2. 先销毁拥有原生句柄的 `AgentClient` 和 `Tasker`，再销毁它们持有的资源或控制器。同时销毁其余拥有原生句柄的 `Resource`、`Controller` 和 `PortalHelper`。检查每个 `Destroy` 的结果，处理其报告的错误后再重试 `Release`。
3. 如果启动过 Agent Server 且未 detach，按 [AgentServerJoin](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#AgentServerJoin) 中的顺序关闭，再重试 `Release`。仅从 `Join` 返回并未完成服务端清理。
4. 如果调用过 `AgentServerDetach`，按下一节处理。

所有权和销毁规则见[包概览](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#pkg-overview)、[Tasker.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Tasker.Destroy)、[Resource.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Resource.Destroy)、[Controller.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Controller.Destroy)、[AgentClient.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#AgentClient.Destroy) 和 [PortalHelper.Destroy](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#PortalHelper.Destroy)。卸载保护见 [Release](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#Release)。

## AgentServerDetach 后 Release 一直被阻止

保留已加载的库直到进程退出。之后调用 `Join` 或 `ShutDown` 也不能使卸载库变得安全。需要释放库的程序应移除 `Detach` 并重启进程，再从 [AgentServerJoin](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#AgentServerJoin) 中选择关闭顺序。

仅为了让服务在后台运行，不需要调用 `Detach`。选择服务生命周期时，查阅 [AgentServerStartUp](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#AgentServerStartUp) 和 [AgentServerDetach](https://pkg.go.dev/github.com/MaaXYZ/maa-framework-go/v4#AgentServerDetach)。

## 报告未解决的问题

提供绑定版本、MaaFramework 版本、操作系统和架构、加载方式、完整错误以及最小复现。日志中应移除凭据等隐私信息。贡献说明见[贡献指南](../contributing.md)。
