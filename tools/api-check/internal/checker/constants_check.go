package checker

import (
	"fmt"
	"go/constant"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const maaToolkitDefHeaderRel = "MaaToolkit/MaaToolkitDef.h"

// constantIssueSection is the report section used for constant inventories.
const constantIssueSection = sectionConstants

// constantCHeaderRels are the headers whose macros and enums feed every
// constant inventory family.
var constantCHeaderRels = []string{maaDefHeaderRel, maaToolkitDefHeaderRel}

type numericKind uint8

const (
	numericUntyped numericKind = iota
	numericSigned
	numericUnsigned
)

// numericValue is an exact integer constant together with the signedness and
// width of its source declaration.
type numericValue struct {
	kind numericKind
	bits int
	val  constant.Value
}

// normalizeNumericValue applies the declared signedness and width, rejecting
// values outside the declared range. C's unsigned conversion rules are
// applied separately by normalizeCNumericValue.
func normalizeNumericValue(value numericValue) (constant.Value, error) {
	if value.val == nil {
		return nil, fmt.Errorf("missing constant value")
	}
	integer := constant.ToInt(value.val)
	if integer.Kind() != constant.Int {
		return nil, fmt.Errorf("value is not an integer constant: %s", value.val.ExactString())
	}

	switch value.kind {
	case numericUntyped:
		return integer, nil
	case numericSigned:
		if value.bits <= 0 {
			return integer, nil
		}
		min := constant.Shift(constant.MakeInt64(-1), token.SHL, uint(value.bits-1))
		max := constant.BinaryOp(
			constant.Shift(constant.MakeInt64(1), token.SHL, uint(value.bits-1)),
			token.SUB,
			constant.MakeInt64(1),
		)
		if constant.Compare(integer, token.LSS, min) || constant.Compare(integer, token.GTR, max) {
			return nil, fmt.Errorf("value %s does not fit signed %d-bit range", integer.ExactString(), value.bits)
		}
		return integer, nil
	case numericUnsigned:
		if value.bits <= 0 {
			return integer, nil
		}
		mask := unsignedBitMask(value.bits)
		if constant.Sign(integer) < 0 {
			return nil, fmt.Errorf("value %s does not fit unsigned %d-bit range", integer.ExactString(), value.bits)
		}
		if constant.Compare(integer, token.GTR, mask) {
			return nil, fmt.Errorf("value %s does not fit unsigned %d-bit range", integer.ExactString(), value.bits)
		}
		return integer, nil
	default:
		return nil, fmt.Errorf("unsupported numeric kind")
	}
}

// normalizeCNumericValue applies C's unsigned conversion rule before checking
// the range. Both positive and negative integers convert modulo 2^bits.
func normalizeCNumericValue(value numericValue) (constant.Value, error) {
	if value.val != nil && value.kind == numericUnsigned && value.bits > 0 {
		integer := constant.ToInt(value.val)
		if integer.Kind() == constant.Int {
			value.val = constant.BinaryOp(integer, token.AND, unsignedBitMask(value.bits))
		}
	}
	return normalizeNumericValue(value)
}

func unsignedBitMask(bits int) constant.Value {
	shifted := constant.Shift(constant.MakeInt64(1), token.SHL, uint(bits))
	return constant.BinaryOp(shifted, token.SUB, constant.MakeInt64(1))
}

// uint64Value normalizes a numeric value and requires it to fit uint64. Go
// method constants use this so unsigned wraps such as ^uint32(0) are widened
// instead of truncated, while untyped negative values are rejected.
func uint64Value(value numericValue) (uint64, error) {
	normalized, err := normalizeNumericValue(value)
	if err != nil {
		return 0, err
	}
	unsigned, ok := constant.Uint64Val(normalized)
	if !ok {
		return 0, fmt.Errorf("value %s does not fit uint64", normalized.ExactString())
	}
	return unsigned, nil
}

// constantFamilySpec describes one cross-language constant inventory.
type constantFamilySpec struct {
	// name is the report label, for example "macos.screencap".
	name string
	// goFile is the repository-root-relative Go source declaring the family.
	goFile string
	// goPrefix is stripped from exported Go constant names.
	goPrefix string
	// cPrefix is stripped from C macro and enum names.
	cPrefix string
	// publicAliases resolves references to the internal native constants.
	publicAliases bool
	// kind and bits describe the C typedef underlying type used to normalize C
	// values.
	kind numericKind
	bits int
	// foldCase matches names case-insensitively after removing underscores.
	// It is used by the gamepad inventories, whose C names are upper case
	// (LEFT_THUMB) while the Go names are CamelCase (LeftThumb).
	foldCase bool
	// aliases maps a canonical C name to the canonical Go name when the two
	// sides deliberately use different names. Aliases are explicit and local to
	// one family; unknown names are still reported.
	aliases map[string]string
}

// constantFamilySpecs lists every constant inventory compared by
// checkConstantCoverage. Gamepad DS4 aliases (CROSS/CIRCLE/SQUARE/TRIANGLE,
// L1/R1/L3/R3, OPTIONS/SHARE) are declared constants on both sides and are
// matched by name like any other member; equal values never collapse names, so
// unknown additions or removals stay visible.
var constantFamilySpecs = []constantFamilySpec{
	{name: "gamepad.type", goFile: "controller.go", goPrefix: "GamepadType", cPrefix: "MaaGamepadType_", kind: numericUnsigned, bits: 64, publicAliases: true},
	{name: "macos.permission", goFile: "toolkit.go", goPrefix: "MacOSPermission", cPrefix: "MaaMacOSPermission", kind: numericSigned, bits: 32, publicAliases: true},
	{
		name:     "macos.screencap",
		goFile:   "controller/macos/macos.go",
		goPrefix: "Screencap",
		cPrefix:  "MaaMacOSScreencapMethod_",
		kind:     numericUnsigned,
		bits:     64,
	},
	{
		name:     "macos.input",
		goFile:   "controller/macos/macos.go",
		goPrefix: "Input",
		cPrefix:  "MaaMacOSInputMethod_",
		kind:     numericUnsigned,
		bits:     64,
	},
	{
		name:     "native.gamepad_type",
		goFile:   "internal/native/framework.go",
		goPrefix: "MaaGamepadType_",
		cPrefix:  "MaaGamepadType_",
		kind:     numericUnsigned,
		bits:     64,
	},
	{
		name:     "gamepad.button",
		goFile:   "controller/gamepad/gamepad.go",
		goPrefix: "Button",
		cPrefix:  "MaaGamepadButton_",
		kind:     numericUnsigned,
		bits:     64,
		foldCase: true,
	},
	{
		name:     "gamepad.touch",
		goFile:   "controller/gamepad/gamepad.go",
		goPrefix: "Touch",
		cPrefix:  "MaaGamepadTouch_",
		kind:     numericUnsigned,
		bits:     64,
		foldCase: true,
	},
	{
		name:     "controller.feature",
		goFile:   "custom_controller.go",
		goPrefix: "ControllerFeature",
		cPrefix:  "MaaControllerFeature_",
		kind:     numericUnsigned,
		bits:     64,
	},
	{
		name:     "status",
		goFile:   "status.go",
		goPrefix: "Status",
		cPrefix:  "MaaStatus_",
		kind:     numericSigned,
		bits:     32,
		aliases: map[string]string{
			"Succeeded": "Success",
			"Failed":    "Failure",
		},
	},
	{
		name:     "logging_level",
		goFile:   "maa.go",
		goPrefix: "LoggingLevel",
		cPrefix:  "MaaLoggingLevel_",
		kind:     numericSigned,
		bits:     32,
	},
	{
		name:     "native.global_option",
		goFile:   "internal/native/framework.go",
		goPrefix: "MaaGlobalOption_",
		cPrefix:  "MaaGlobalOption_",
		kind:     numericSigned,
		bits:     32,
	},
	{
		name:     "native.res_option",
		goFile:   "internal/native/framework.go",
		goPrefix: "MaaResOption_",
		cPrefix:  "MaaResOption_",
		kind:     numericSigned,
		bits:     32,
	},
	{
		name:     "native.ctrl_option",
		goFile:   "internal/native/framework.go",
		goPrefix: "MaaCtrlOption_",
		cPrefix:  "MaaCtrlOption_",
		kind:     numericSigned,
		bits:     32,
	},
	{
		name:     "native.tasker_option",
		goFile:   "internal/native/framework.go",
		goPrefix: "MaaTaskerOption_",
		cPrefix:  "MaaTaskerOption_",
		kind:     numericSigned,
		bits:     32,
	},
	{
		name:     "native.inference_device",
		goFile:   "internal/native/framework.go",
		goPrefix: "MaaInferenceDevice_",
		cPrefix:  "MaaInferenceDevice_",
		kind:     numericSigned,
		bits:     32,
	},
	{
		name:     "native.execution_provider",
		goFile:   "internal/native/framework.go",
		goPrefix: "MaaInferenceExecutionProvider_",
		cPrefix:  "MaaInferenceExecutionProvider_",
		kind:     numericSigned,
		bits:     32,
	},
	{
		name:     "native.macos_permission",
		goFile:   "internal/native/toolkit.go",
		goPrefix: "MaaMacOSPermission",
		cPrefix:  "MaaMacOSPermission",
		kind:     numericSigned,
		bits:     32,
	},
}

// checkConstantCoverage compares the C constant inventories in the configured
// headers with the Go constants that expose them. Every family reports C
// constants missing in Go, Go constants missing in C, and value mismatches.
func checkConstantCoverage(repoRoot string, headerDir string) ([]issue, error) {
	sources := make([]cConstSource, 0, len(constantCHeaderRels))
	for _, rel := range constantCHeaderRels {
		path := resolveFromRepoRoot(repoRoot, filepath.Join(headerDir, rel))
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read C header %s: %w", path, err)
		}
		sources = append(sources, cConstSource{path: path, content: string(data)})
	}

	cEnv, err := evaluateCConstSources(sources)
	if err != nil {
		return nil, fmt.Errorf("evaluate C constants: %w", err)
	}

	goEvaluations := map[string]*goConstEvaluation{}
	goErrors := map[string]error{}
	loadGo := func(rel string, publicAliases bool) (*goConstEvaluation, error) {
		if loadErr, ok := goErrors[rel]; ok {
			return nil, loadErr
		}
		if evaluation, ok := goEvaluations[rel]; ok {
			return evaluation, nil
		}
		var evaluation *goConstEvaluation
		var loadErr error
		if publicAliases {
			evaluation, loadErr = evaluatePublicConstants(repoRoot, rel)
		} else {
			evaluation, loadErr = evaluateGoConstFile(resolveFromRepoRoot(repoRoot, rel))
		}
		if loadErr != nil {
			goErrors[rel] = loadErr
			return nil, loadErr
		}
		goEvaluations[rel] = evaluation
		return evaluation, nil
	}

	issues := make([]issue, 0)
	for _, spec := range constantFamilySpecs {
		goEvaluation, err := loadGo(spec.goFile, spec.publicAliases)
		if err != nil {
			return nil, fmt.Errorf("evaluate %s constants: %w", spec.name, err)
		}
		issues = append(issues, compareConstantFamily(spec, cEnv, goEvaluation)...)
	}
	return issues, nil
}

type constantFamilyEntry struct {
	logical string
	value   numericValue
	failed  string
}

func compareConstantFamily(spec constantFamilySpec, cEnv *cConstEnv, goEvaluation *goConstEvaluation) []issue {
	issues := make([]issue, 0)

	cEntries := map[string]constantFamilyEntry{}
	for _, name := range sortedConstantNames(cEnv.decls) {
		decl := cEnv.decls[name]
		logical, ok := strings.CutPrefix(name, spec.cPrefix)
		if !ok || logical == "" {
			continue
		}
		entry := constantFamilyEntry{logical: logical}
		switch {
		case cEnv.failures[name] != "":
			entry.failed = cEnv.failures[name]
			issues = append(issues, issue{
				section: constantIssueSection,
				message: fmt.Sprintf("[%s] failed to evaluate C constant: %s expr=%s (%s)", spec.name, logical, decl.expr, entry.failed),
			})
		default:
			value, ok := cEnv.values[name]
			if !ok {
				entry.failed = "not evaluated"
				issues = append(issues, issue{
					section: constantIssueSection,
					message: fmt.Sprintf("[%s] failed to evaluate C constant: %s expr=%s (not evaluated)", spec.name, logical, decl.expr),
				})
				break
			}
			normalized, err := normalizeCNumericValue(numericValue{kind: spec.kind, bits: spec.bits, val: value})
			if err != nil {
				entry.failed = err.Error()
				issues = append(issues, issue{
					section: constantIssueSection,
					message: fmt.Sprintf("[%s] invalid C constant value: %s (%s)", spec.name, logical, err),
				})
				break
			}
			entry.value = numericValue{kind: spec.kind, bits: spec.bits, val: normalized}
		}
		cEntries[logical] = entry
	}

	goEntries := map[string]constantFamilyEntry{}
	for _, name := range sortedConstantNames(goEvaluation.values) {
		decl := goEvaluation.values[name]
		logical, ok := strings.CutPrefix(name, spec.goPrefix)
		if !ok || logical == "" {
			continue
		}
		goEntries[logical] = constantFamilyEntry{
			logical: logical,
			value:   numericValue{kind: decl.kind, bits: decl.bits, val: decl.value},
		}
	}
	for _, name := range sortedConstantNames(goEvaluation.failures) {
		failure := goEvaluation.failures[name]
		logical, ok := strings.CutPrefix(name, spec.goPrefix)
		if !ok || logical == "" {
			continue
		}
		if _, exists := goEntries[logical]; exists {
			continue
		}
		goEntries[logical] = constantFamilyEntry{logical: logical, failed: failure.err}
		issues = append(issues, issue{
			section: constantIssueSection,
			message: fmt.Sprintf("[%s] failed to evaluate Go constant: %s expr=%s (%s)", spec.name, logical, failure.expr, failure.err),
		})
	}
	if len(cEntries) == 0 {
		issues = append(issues, issue{section: constantIssueSection, message: fmt.Sprintf("[%s] no C constants found", spec.name)})
	}
	if len(goEntries) == 0 {
		issues = append(issues, issue{section: constantIssueSection, message: fmt.Sprintf("[%s] no Go constants found", spec.name)})
	}

	colliding := map[string]bool{}
	cByKey := canonicalizeConstantEntries(spec, "C", cEntries, colliding, &issues)
	goByKey := canonicalizeConstantEntries(spec, "Go", goEntries, colliding, &issues)

	matchedGo := map[string]bool{}
	for _, cKey := range sortedConstantFamilyKeys(cByKey) {
		cEntry := cByKey[cKey]
		goKey := cKey
		if alias, ok := spec.aliases[cKey]; ok {
			goKey = alias
		}
		if colliding[cKey] || colliding[goKey] {
			continue
		}
		goEntry, ok := goByKey[goKey]
		if !ok {
			issues = append(issues, issue{
				section: constantIssueSection,
				message: fmt.Sprintf("[%s] C constant not found in Go: %s (c=%s)", spec.name, cEntry.logical, constantEntryValueText(cEntry)),
			})
			continue
		}
		matchedGo[goKey] = true
		if cEntry.failed != "" || goEntry.failed != "" {
			continue
		}
		cValue, err := normalizeNumericValue(cEntry.value)
		if err != nil {
			issues = append(issues, issue{
				section: constantIssueSection,
				message: fmt.Sprintf("[%s] invalid C constant value: %s (%s)", spec.name, cEntry.logical, err),
			})
			continue
		}
		goValue, err := normalizeNumericValue(goEntry.value)
		if err != nil {
			issues = append(issues, issue{
				section: constantIssueSection,
				message: fmt.Sprintf("[%s] invalid Go constant value: %s (%s)", spec.name, goEntry.logical, err),
			})
			continue
		}
		if !constant.Compare(cValue, token.EQL, goValue) {
			matchName := cEntry.logical
			if goEntry.logical != cEntry.logical {
				matchName = cEntry.logical + " -> " + goEntry.logical
			}
			issues = append(issues, issue{
				section: constantIssueSection,
				message: fmt.Sprintf("[%s] constant value mismatch: %s (go=%s c=%s)", spec.name, matchName, goValue.ExactString(), cValue.ExactString()),
			})
		}
	}
	for _, goKey := range sortedConstantFamilyKeys(goByKey) {
		if matchedGo[goKey] || colliding[goKey] {
			continue
		}
		goEntry := goByKey[goKey]
		issues = append(issues, issue{
			section: constantIssueSection,
			message: fmt.Sprintf("[%s] Go constant not found in C: %s (go=%s)", spec.name, goEntry.logical, constantEntryValueText(goEntry)),
		})
	}

	return issues
}

func canonicalizeConstantEntries(
	spec constantFamilySpec,
	side string,
	entries map[string]constantFamilyEntry,
	colliding map[string]bool,
	issues *[]issue,
) map[string]constantFamilyEntry {
	grouped := map[string][]constantFamilyEntry{}
	for _, logical := range sortedConstantNames(entries) {
		entry := entries[logical]
		key := canonicalConstantName(logical, spec.foldCase)
		grouped[key] = append(grouped[key], entry)
	}

	byKey := map[string]constantFamilyEntry{}
	for _, key := range sortedConstantFamilyGroupKeys(grouped) {
		group := grouped[key]
		if len(group) == 1 {
			byKey[key] = group[0]
			continue
		}
		colliding[key] = true
		parts := make([]string, 0, len(group))
		for _, entry := range group {
			parts = append(parts, fmt.Sprintf("%s(%s=%s)", entry.logical, side, constantEntryValueText(entry)))
		}
		*issues = append(*issues, issue{
			section: constantIssueSection,
			message: fmt.Sprintf("[%s] %s constant name collision after normalization: %s => %s", spec.name, side, key, strings.Join(parts, ", ")),
		})
	}
	return byKey
}

func canonicalConstantName(name string, foldCase bool) string {
	var builder strings.Builder
	for _, r := range name {
		if r == '_' {
			continue
		}
		if foldCase {
			r = unicode.ToLower(r)
		}
		builder.WriteRune(r)
	}
	return builder.String()
}

func constantEntryValueText(entry constantFamilyEntry) string {
	if entry.failed != "" {
		return "?"
	}
	normalized, err := normalizeNumericValue(entry.value)
	if err != nil {
		if entry.value.val != nil {
			return entry.value.val.ExactString()
		}
		return "?"
	}
	return normalized.ExactString()
}

func sortedConstantNames[T any](entries map[string]T) []string {
	names := make([]string, 0, len(entries))
	for name := range entries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func sortedConstantFamilyKeys(entries map[string]constantFamilyEntry) []string {
	return sortedConstantNames(entries)
}

func sortedConstantFamilyGroupKeys(entries map[string][]constantFamilyEntry) []string {
	return sortedConstantNames(entries)
}
