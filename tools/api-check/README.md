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
  - method existence on both sides
  - method signature consistency using the same canonical type rules

## Usage

Working directory:

- Recommended: run commands from `tools/api-check`.
- If you run from repository root, use `-C tools/api-check` with `go`.

Run with defaults:

```bash
go run .
```

Run with explicit config file:

```bash
go run . --config config.yaml
```

Override header directory for one run:

```bash
go run . --header-dir ../../deps/include
```

Add blacklist entries from CLI:

```bash
go run . --blacklist MaaDbgControllerCreate --blacklist MaaDbgControllerType
```

CI-style config file path:

```bash
go run . --config config.ci.yaml
```

Equivalent from repository root:

```bash
go -C tools/api-check run . --config config.ci.yaml
```

## Config resolution

- If `--config` is provided, that file is required and will be loaded.
- If `--config` is not provided, tool tries `config.yaml` in current working directory.
- If no config file is found, tool uses built-in defaults.
- Priority is: `CLI flags > config file > defaults`.

Defaults:

- `header_dir: ../../deps/include`
- `blacklist: []`

## Config template

Copy `config.example.yaml` to `config.yaml` in `tools/api-check` and edit as needed.
