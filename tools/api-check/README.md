# API Check Tool

`tools/api-check` is a consistency checker for:

- `internal/native` registered C symbols (`purego.RegisterLibFunc`)
- exported C functions in header files
- `CustomController` interface vs `MaaCustomControllerCallbacks`
- controller method enums/constants in `controller/adb` and `controller/win32` vs `MaaDef.h`

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
- Controller method coverage:
  - compare `adb/win32` `ScreencapMethod` and `InputMethod` names against C macros in `MaaDef.h`
  - C-side method names are matched after removing `_` (for example `DXGI_DesktopDup` matches `DXGIDesktopDup`)
  - report C method missing in Go
  - report Go method missing in C
  - report value mismatch for same method name

## Usage

Working directory:

- The tool is repo-root aware and can be run from repository root or any subdirectory.
- You can still run it from `tools/api-check` if preferred.

Run with defaults:

```bash
go -C tools/api-check run .
```

Run with explicit config file:

```bash
go -C tools/api-check run . --config config.yaml
```

Override header directory for one run:

```bash
go -C tools/api-check run . --header-dir deps/include
```

Add blacklist entries from CLI:

```bash
go -C tools/api-check run . --blacklist MaaDbgControllerCreate --blacklist MaaDbgControllerType
```

CI-style config file path:

```bash
go -C tools/api-check run . --config config.ci.yaml
```

Equivalent from `tools/api-check`:

```bash
go run . --config config.ci.yaml
```

## Config resolution

- If `--config` is provided, that file is required and will be loaded.
- If `--config` is not provided, the tool tries `config.yaml` in the current working directory, then `tools/api-check/config.yaml` under detected repo root.
- If no config file is found, the tool uses built-in defaults.
- Priority is: `CLI flags > config file > defaults`.

Defaults:

- `header_dir: deps/include`
- `blacklist: []`

## Config template

Copy `config.example.yaml` to `config.yaml` in `tools/api-check` and edit as needed.
