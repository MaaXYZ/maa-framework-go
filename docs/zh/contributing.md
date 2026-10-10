# 贡献指南

[English](../../CONTRIBUTING.md) | 简体中文

欢迎提交问题报告、功能建议和 Pull Request。

## 开始贡献

- 通过 [Issue](https://github.com/MaaXYZ/maa-framework-go/issues) 报告问题或提出修改建议。报告问题时，请提供 Go 版本、操作系统与架构、Go 绑定和 MaaFramework 版本、复现步骤及相关日志。
- 对于较大的 API 或行为变更，先在 Issue 中讨论预期范围，再开始实现。
- Fork 仓库，为修改创建分支，并让每个 Pull Request 聚焦于一个目的。

## 构建与测试

使用 Go 1.25 及以上版本。以下命令均在仓库根目录执行。

初始化测试资源子模块：

```shell
git submodule update --init --recursive
```

下载适合操作系统与架构的 [MaaFramework release](https://github.com/MaaXYZ/MaaFramework/releases)，并解压到 `deps/`。头文件、schema 和动态库必须来自同一个 release：

- 头文件：`deps/include/`
- 流水线 schema：`deps/tools/pipeline.schema.json`
- 动态库及其依赖：`deps/bin/`。测试会加载 MaaFramework、MaaToolkit、MaaAgentServer 和 MaaAgentClient。

启动原生测试前，将 `deps/bin/` 的绝对路径加入加载器搜索路径：

| 平台 | 环境设置 |
| --- | --- |
| Windows（PowerShell） | `$env:PATH = "$((Resolve-Path deps/bin).Path);$env:PATH"` |
| Linux | `export LD_LIBRARY_PATH="$PWD/deps/bin${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"` |
| macOS | `export DYLD_LIBRARY_PATH="$PWD/deps/bin${DYLD_LIBRARY_PATH:+:$DYLD_LIBRARY_PATH}"` |

为集成测试准备 OCR 和分类模型。在 POSIX shell 中执行：

```shell
mkdir -p test/data_set/PipelineSmoking/resource/model/ocr
cp -r test/data_set/MaaCommonAssets/OCR/ppocr_v4/zh_cn/* test/data_set/PipelineSmoking/resource/model/ocr/
mkdir -p test/data_set/PipelineSmoking/resource/model/classify
cp test/classifier.onnx test/data_set/PipelineSmoking/resource/model/classify/
```

在 Windows 上，使用当前 shell 的等价命令，将相同文件复制到这些目标目录。

执行检查：

```shell
go build ./...
go vet ./...
go test -v ./...
go -C tools/api-check test ./...
go -C tools/api-check run . --config config.ci.yaml
```

[测试工作流](../../.github/workflows/test.yml) 使用稳定版 Go 工具链和已发布的 MaaFramework release，包括预发布版本。根模块测试在 Windows、Linux amd64 和 macOS arm64 上以 `CGO_ENABLED=0` 运行；本地复现 CI 时请设置相同环境变量。

[API 检查器](../../tools/api-check/README.md) 读取 Go 源码、原生头文件和流水线 schema，无需加载动态库。它检查 API 一致性以及流水线类型和字段名覆盖；默认值、继承、单位和运行时行为需要单独测试。

## 代码与文档

- 使用 `gofmt` 格式化修改过的 Go 文件，并遵循标准 Go 命名与导入约定。
- 为导出 API 编写注释，确保公开包装与底层 MaaFramework API 一致。
- 将平台相关代码放在对应的 controller 包或 `internal/native/` 中。
- 使用 Go 的 `testing` 包和现有 `testify/require` 断言补充有针对性的回归测试。单元测试放在代码旁；回放或流水线集成测试放在 `test/` 中。
- 保持 [README.md](../../README.md) 与 [README_zh.md](../../README_zh.md) 内容一致。章节、示例、徽章和链接的修改应在同一组变更中同步。其他双语文档也应保持对应。

## 提交与分支

使用 [Conventional Commits](https://www.conventionalcommits.org/) 格式和英文提交标题，例如 `fix: handle empty action results` 或 `docs: explain library loading`。可以添加 scope：`type(scope): description`。

类型从 `feat`、`fix`、`docs`、`test`、`refactor`、`perf`、`build`、`ci`、`chore`、`revert` 中选择。分支采用 `<type>/<kebab-case-description>` 格式，以修改的主要目的确定前缀，例如 `test/agent-server-callback-test` 或 `fix/action-result-parser`。

## Pull Request

完成相关检查后，向 `main` 提交 [Pull Request](https://github.com/MaaXYZ/maa-framework-go/pulls)。修改尚未完成时使用草稿状态。

在描述中提供：

- 问题、修改后的行为和关联 Issue。
- 受影响的平台，以及 MaaFramework 版本或 API 兼容性影响。
- 验证命令与结果，包括未能执行的检查及原因。
- 有助于展示可观察行为变化的日志或截图。

处理评审反馈后，重新运行后续修改影响的检查，再请求新一轮评审。
