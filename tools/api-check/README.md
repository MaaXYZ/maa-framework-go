# API Check Tool

`tools/api-check` is a consistency checker for:

- `internal/native` registered C symbols in the `[]Entry` tables consumed by `Library.entries`
- exported C functions in header files
- `CustomController` interface vs `MaaCustomControllerCallbacks`
- controller method enums/constants in `controller/adb`, `controller/win32`, and `controller/macos` vs `MaaDef.h`
- callback typedefs, custom-controller struct layout, callback bindings, and trampoline signatures
- gamepad codes/types, controller features, status/log levels, native options, inference constants, and macOS permissions
- event message constants and dispatch vs `MaaMsg.h`
- pipeline v2 type and field-name coverage against a release schema (opt in)

It checks both symbol coverage and function signatures.

## What is checked

- Native API coverage:
  - header function exists but Go is not registering it
  - Go registers function not found in headers
  - `Entry` mismatch: `ptrToFunc` target name != symbol string
  - discover all non-test native Go files, including declarations in separate files
  - resolve named function types, aliases, inferred function literals, and function-type conversions
  - reject missing signatures, malformed entries, duplicate registrations, and missing library tables
  - report C/Go symbol counts for each module; empty inventories fail extraction
- Native API signature consistency:
  - compare Go var function signature vs C exported function signature
  - compare params/returns with strict arity/order
  - C `char*` returns accept Go `string` or raw pointers, allowing callers to read length-delimited buffer content
  - C types are normalized with typedef expansion (for example `MaaTaskId -> MaaId -> int64_t`)
- CustomController consistency:
  - method existence on both sides
  - method signature consistency using the same canonical type rules, including local embedded interfaces
  - validate context/output adapter argument types, counts, and positions before removing them
  - C string callback arguments/results use Go byte pointers; Go `string` is rejected by purego callbacks
  - require exactly one pointer-sized, non-floating-point Go callback result on the supported 64-bit targets, including C `void` and `MaaBool` callbacks
  - compare callback struct field order and pointer-sized storage, bindings, and trampoline ABI
  - compare native callback typedefs and their root trampolines; require Go declarations for callbacks used by exported C functions
- Controller method coverage:
  - compare `adb/win32` `ScreencapMethod` and `InputMethod` names against C macros in `MaaDef.h`
  - C-side method names are matched after removing `_` (for example `DXGI_DesktopDup` matches `DXGIDesktopDup`)
  - report C method missing in Go
  - report Go method missing in C
  - report value mismatch for same method name
- Constant coverage:
  - compare the 17 additional constant families listed above in both directions, keeping aliases distinct
  - evaluate Go constants with `go/types`, preserving conversions, widths, dependencies, and `iota`
  - evaluate C macros/enums with integer promotions, unsigned wrapping, casts, and integer division
  - reject unknown dependencies, ambiguous names, and empty families; C `long` literal suffixes are unsupported because their width differs by platform
  - evaluate public gamepad and macOS permission aliases against their native declarations
- Event coverage:
  - compare every Starting/Succeeded/Failed message in both directions
  - require a nonempty dispatch case on the parsed incoming event in `handleRaw`
- Pipeline v2 coverage, when `pipeline_schema` is configured:
  - compare `ActionEnum` and `RecognitionEnum` independently with Go type constants
  - require a typed parameter decoder case for each known type, selected by the decoded Type and flowing into the receiver's Param
  - reject writes to decoded envelopes or their Type fields after decoding, hidden envelope aliases or escapes, and reassigned or shadowed helper discriminants
  - ignore unused helpers and reject ambiguous receiver writes, shadowing, helper cycles, and fixed discriminants; raw fallback does not count as typed support
  - compare JSON field names in both directions for `Node`, Action/Recognition envelopes, parameters, `SwipeListItem`, `WaitFreezes`, `NodeAttr`, and optional `SubRecognitionInline`
  - inspect JSON tags, anonymous embedding, type aliases, and the actual struct passed to supported custom JSON codecs, including duration wire fields
  - reject missing, invalid, or unsupported schema/code shapes instead of silently skipping them
  - report exact schema-oriented paths, Go parameter types, and source locations for extra Go fields

Pipeline coverage is limited to the v2 object format (`{"type": "Click", "param": {...}}`). It does not check v1 flat parameters, JSON value validation, default inheritance, omission behavior, units, scalar/list normalization, or runtime behavior. And/Or parameter field names and the available inline sub-recognition envelope are checked; value semantics still require JSON and native round-trip tests.

Schema metadata (`jsonComments`, `jsonCode`, `jsonDocument`, `jsonKeywords`), deprecated node fields, the v2 default-field helper branches, and the `CustomActionSchema`/`CustomRecognitionSchema` extension hooks are outside this inventory. Intrinsic Custom parameter fields are checked; arbitrary custom payload contents stay open. Other external schema references are rejected. The checker reads source using Go AST and requires no native libraries. It supports the repository's declaration shapes; it does not preprocess arbitrary C conditional branches or prove general Go control flow.

Supported custom codecs pass a struct, a defined type without methods, or a traced local variable to `marshalJSON`/`unmarshalJSON` or `json.Marshal`/`json.Unmarshal`. `MarshalJSON` must directly return the supported JSON helper call; unrelated calls and calls inside closures do not establish coverage. Returning encoded bytes through variables, wire types with their own or inherited custom codecs, anonymous embedding promoting custom codecs, and conflicting JSON field names are rejected as unsupported shapes. Local `type NoMethod Param` DTOs remain supported because they strip methods; `type NoMethod = Param` aliases preserve methods. Other shapes require extending the checker explicitly.

## Preventing false passes

Checker changes need both a valid baseline that passes and isolated, compilable mutations that must fail. Test the complete extraction or CLI path, rather than only a comparison helper: overwrite a decoded discriminator, replace its envelope, reassign a helper parameter, or change a callback return type while keeping the C and Go signatures identical. Include nearby aliases, shadowing, and closure writes when they can hide the same fault. A mutation must produce a reported difference or an unsupported-shape error, never `PASS`.

Use independent runtime evidence for ABI rules. The Windows-only checker test registers callbacks with the root module's purego dependency and compares the observed registration results with the checker's decisions. It needs no MaaFramework libraries. Cross-compilation alone cannot expose callback registration panics; the existing Windows CI executes this test.

Keep unsupported data flow conservative: reject shapes the extractor cannot establish, and expand the supported subset only with passing baselines and failing mutations. Source inventory checks do not prove arbitrary runtime behavior, so retain native round-trip tests for value semantics. Statement coverage measures exercised code, not the number of broken bindings detected; review surviving mutations against the documented scope before accepting a checker change.

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
go -C tools/api-check run . --blacklist MaaDbgControllerCreate
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
- `native_exclusions: {}`
- `blacklist: []` (legacy symbol exclusions)
- `pipeline_schema: ""` (pipeline checking disabled)
- `pipeline_exclusions: {}`

YAML decoding is strict: unknown fields, duplicate keys, wrong scalar types, and multiple documents fail configuration. Positional command-line arguments are rejected. Each run uses an independent flag set.

All header and schema paths are resolved relative to the detected repository root, including paths provided in a config file.

`native_exclusions` accepts exact native function names with nonempty reasons:

```yaml
native_exclusions:
  MaaDbgControllerCreate: "BlankController provides the Go no-op controller."
```

Only active symbol differences are suppressed; stale exclusions fail. Malformed registrations and callback/constant/event checks cannot be hidden by native exclusions. Legacy `blacklist` and repeatable `--blacklist` remain supported and also reject stale entries.

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
