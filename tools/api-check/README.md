# API Check Tool

`tools/api-check` is a consistency checker for:

- `internal/native` registered C symbols declared in `[]Entry` tables and bound by `purego.RegisterLibFunc`
- exported C functions in header files
- `CustomController` interface vs `MaaCustomControllerCallbacks`
- controller method enums/constants in `controller/adb` and `controller/win32` vs `MaaDef.h`
- pipeline v2 type and field-name coverage against a release schema (opt in)

It checks both symbol coverage and function signatures.

## What is checked

- Native API coverage:
  - header function exists but Go is not registering it
  - Go registers function not found in headers
  - `Entry` mismatch: `ptrToFunc` target name != symbol string
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
- Pipeline v2 coverage, when `pipeline_schema` is configured:
  - compare `ActionEnum` and `RecognitionEnum` independently with Go type constants
  - require a typed parameter decoder case for each known type; raw unknown-type fallback does not count as typed support
  - compare JSON field names in both directions for `Node`, action/recognition parameters, `SwipeListItem`, `WaitFreezes`, and `NodeAttr`
  - inspect JSON tags, anonymous embedding, type aliases, and the actual struct passed to supported custom JSON codecs, including duration wire fields
  - reject missing, invalid, or unsupported schema/code shapes instead of silently skipping them
  - report exact schema-oriented paths, Go parameter types, and source locations for extra Go fields

Pipeline coverage is limited to the v2 object format (`{"type": "Click", "param": {...}}`). It does not check v1 flat parameters, JSON value validation, default inheritance, omission behavior, units, scalar/list normalization, or runtime behavior. And/Or parameter field names are checked, but the upstream schema does not completely describe inline sub-recognition semantics; those require JSON and native round-trip tests.

Schema metadata (`jsonComments`, `jsonCode`, `jsonDocument`, `jsonKeywords`), deprecated node fields, the v2 default-field helper branches, and the `CustomActionSchema`/`CustomRecognitionSchema` extension hooks are outside this inventory. Intrinsic Custom parameter fields are checked; arbitrary custom payload contents stay open. Other external schema references are rejected. The checker reads source using Go AST and requires no native libraries.

Supported custom codecs pass a struct, a defined type without methods, or a traced local variable to `marshalJSON`/`unmarshalJSON` or `json.Marshal`/`json.Unmarshal`. `MarshalJSON` must directly return the supported JSON helper call; unrelated calls and calls inside closures do not establish coverage. Returning encoded bytes through variables, wire types with their own or inherited custom codecs, anonymous embedding promoting custom codecs, and conflicting JSON field names are rejected as unsupported shapes. Local `type NoMethod Param` DTOs remain supported because they strip methods; `type NoMethod = Param` aliases preserve methods. Other shapes require extending the checker explicitly.

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

Enable pipeline v2 coverage for one run:

```bash
go -C tools/api-check run . --pipeline-schema deps/tools/pipeline.schema.json
```

Use the pipeline schema, C headers, and test dynamic libraries from the same MaaFramework release. A configured schema is required input: a missing or unreadable schema fails the run.

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
- `pipeline_schema: ""` (pipeline checking disabled)
- `pipeline_exclusions: {}`

All header and schema paths are resolved relative to the detected repository root, including paths provided in a config file.

`pipeline_exclusions` accepts exact reported paths with nonempty reasons, for example:

```yaml
pipeline_schema: deps/tools/pipeline.schema.json
pipeline_exclusions:
  action.KeyDown.param.key_code: "Legacy alias; the Go wrapper emits the canonical key field."
```

An exclusion suppresses only the matching difference. Wildcards are not supported; stale exclusions are reported as failures. Exclusions require pipeline checking to be enabled and are listed with their reasons in the report.

Exit status is `0` for a consistent implementation, `1` for reported differences or stale exclusions, and `2` for invalid configuration or inputs, including unsupported extraction shapes.

## Config template

Copy `config.example.yaml` to `config.yaml` in `tools/api-check` and edit as needed.
