# Contributing

English | [简体中文](docs/zh/contributing.md)

Bug reports, feature suggestions, and pull requests are welcome.

## Getting started

- Use [issues](https://github.com/MaaXYZ/maa-framework-go/issues) to report bugs or propose changes. For a bug, include the Go version, OS and architecture, binding and MaaFramework versions, reproduction steps, and relevant logs.
- For a larger API or behavior change, discuss the intended scope in an issue before implementing it.
- Fork the repository, create a branch for your change, and keep the pull request focused on one purpose.

## Build and test

Use Go 1.25 or newer. Run the following commands from the repository root.

Initialize the test-assets submodule:

```shell
git submodule update --init --recursive
```

Download a [MaaFramework release](https://github.com/MaaXYZ/MaaFramework/releases) for your OS and architecture and extract it into `deps/`. Use headers, schema, and libraries from the same release:

- Headers: `deps/include/`
- Pipeline schema: `deps/tools/pipeline.schema.json`
- Shared libraries and their dependencies: `deps/bin/`. Tests load MaaFramework, MaaToolkit, MaaAgentServer, and MaaAgentClient.

Before launching native tests, add the absolute `deps/bin/` path to the loader's search path:

| Platform | Environment setup |
| --- | --- |
| Windows (PowerShell) | `$env:PATH = "$((Resolve-Path deps/bin).Path);$env:PATH"` |
| Linux | `export LD_LIBRARY_PATH="$PWD/deps/bin${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"` |
| macOS | `export DYLD_LIBRARY_PATH="$PWD/deps/bin${DYLD_LIBRARY_PATH:+:$DYLD_LIBRARY_PATH}"` |

Prepare the OCR and classifier models for integration tests. In a POSIX shell:

```shell
mkdir -p test/data_set/PipelineSmoking/resource/model/ocr
cp -r test/data_set/MaaCommonAssets/OCR/ppocr_v4/zh_cn/* test/data_set/PipelineSmoking/resource/model/ocr/
mkdir -p test/data_set/PipelineSmoking/resource/model/classify
cp test/classifier.onnx test/data_set/PipelineSmoking/resource/model/classify/
```

On Windows, copy the same files to these destinations using your shell's equivalent commands.

Run the checks:

```shell
go build ./...
go vet ./...
go test -v ./...
go -C tools/api-check test ./...
go -C tools/api-check run . --config config.ci.yaml
```

The [test workflow](.github/workflows/test.yml) uses the stable Go toolchain and a published MaaFramework release, including prereleases. It runs root-module tests with `CGO_ENABLED=0` on Windows and Linux amd64 and macOS arm64; set the same environment variable when reproducing CI locally.

The [API checker](tools/api-check/README.md) reads Go source, native headers, and the pipeline schema without loading shared libraries. It checks API consistency and pipeline type and field-name coverage; defaults, inheritance, units, and runtime behavior need separate tests.

## Code and documentation

- Format changed Go files with `gofmt` and follow standard Go naming and import conventions.
- Document exported APIs and keep public wrappers consistent with the underlying MaaFramework API.
- Put platform-specific code in its controller package or `internal/native/`.
- Add focused regression tests with Go's `testing` package and the existing `testify/require` assertions. Keep unit tests beside the code; use `test/` for replay or pipeline integration tests.
- Keep [README.md](README.md) and [README_zh.md](README_zh.md) aligned. Mirror changes to sections, examples, badges, and links in the same change set. Keep other bilingual documents aligned as well.

## Commits and branches

Use [Conventional Commits](https://www.conventionalcommits.org/) with an English subject, such as `fix: handle empty action results` or `docs: explain library loading`. An optional scope is supported: `type(scope): description`.

Choose a type from `feat`, `fix`, `docs`, `test`, `refactor`, `perf`, `build`, `ci`, `chore`, or `revert`. Name branches `<type>/<kebab-case-description>` using the primary purpose of the change, for example `test/agent-server-callback-test` or `fix/action-result-parser`.

## Pull requests

After running the relevant checks, open a [pull request](https://github.com/MaaXYZ/maa-framework-go/pulls) against `main`. Use a draft while the change is still in progress.

In the description, include:

- The problem, resulting behavior, and related issues.
- Affected platforms and any MaaFramework version or API compatibility impact.
- Validation commands and results, including checks you could not run and why.
- Logs or screenshots when they help demonstrate an observable behavior change.

Address review feedback and rerun the checks affected by subsequent changes before requesting another review.
