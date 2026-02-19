# API Check Tool

`tools/api-check` is a consistency checker for:

- `internal/native` registered C symbols (`purego.RegisterLibFunc`)
- exported C functions in header files
- `CustomController` interface vs `MaaCustomControllerCallbacks`

It checks both symbol coverage and function signatures.

## What is checked

- Native API coverage:
  - header function exists but Go is not registering it
  - Go registers function not found in headers
  - `RegisterLibFunc` arg mismatch: first arg var name != third arg symbol string
- Native API signature consistency:
  - compare Go var function signature vs C exported function signature
  - compare params/returns with strict arity/order
  - C types are normalized with typedef expansion (for example `MaaTaskId -> MaaId -> int64_t`)
- CustomController consistency:
  - method existence in both sides
  - method signature consistency using the same canonical type rules

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
