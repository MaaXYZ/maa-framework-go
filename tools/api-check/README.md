# API Check Tool

`tools/api-check` is a consistency checker for:

- `internal/native` registered C symbols (`purego.RegisterLibFunc`)
- exported C functions in header files
- `CustomController` interface vs `MaaCustomControllerCallbacks`

## Usage

Working directory:

- Run commands from repository root (`maa-framework-go`).
- If you run inside `tools/api-check`, use `go run .` and adjust relative paths accordingly.

Run with defaults:

```bash
go run ./tools/api-check
```

Run with explicit config file:

```bash
go run ./tools/api-check --config tools/api-check/config.yaml
```

Override header directory for one run:

```bash
go run ./tools/api-check --header-dir ../MaaFramework/include
```

Add blacklist entries from CLI:

```bash
go run ./tools/api-check --blacklist MaaDbgControllerCreate --blacklist MaaDbgControllerType
```

CI-style config file path:

```bash
go run ./tools/api-check --config tools/api-check/config.ci.yaml
```

## Config resolution

- If `--config` is provided, that file is required and will be loaded.
- If `--config` is not provided, tool tries `tools/api-check/config.yaml`.
- If no config file is found, tool uses built-in defaults.
- Priority is: `CLI flags > config file > defaults`.

Defaults:

- `header_dir: deps/include`
- `blacklist: []`

## Config template

Copy `tools/api-check/config.example.yaml` to `tools/api-check/config.yaml` and edit as needed.
