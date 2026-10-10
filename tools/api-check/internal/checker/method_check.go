package checker

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strings"
)

const (
	methodGroupAdbScreencap   = "adb.screencap"
	methodGroupAdbInput       = "adb.input"
	methodGroupWin32Screencap = "win32.screencap"
	methodGroupWin32Input     = "win32.input"
)

type methodGroupSpec struct {
	group   string
	cPrefix string
}

var methodGroupSpecs = []methodGroupSpec{
	{group: methodGroupAdbScreencap, cPrefix: "MaaAdbScreencapMethod_"},
	{group: methodGroupAdbInput, cPrefix: "MaaAdbInputMethod_"},
	{group: methodGroupWin32Screencap, cPrefix: "MaaWin32ScreencapMethod_"},
	{group: methodGroupWin32Input, cPrefix: "MaaWin32InputMethod_"},
}

type normalizedCMethod struct {
	name  string
	value uint64
}

type unknownIdentifierError struct {
	name string
}

// Error reports the identifier that the C constant evaluator could not resolve.
func (e *unknownIdentifierError) Error() string {
	return "unknown identifier: " + e.name
}

func checkControllerMethodCoverage(maaDefHeaderPath string, adbControllerPath string, win32ControllerPath string) ([]issue, error) {
	cGroups, cIssues, err := parseCControllerMethodGroups(maaDefHeaderPath)
	if err != nil {
		return nil, fmt.Errorf("parse C controller methods: %w", err)
	}

	adbGroups, adbIssues, err := parseGoControllerMethodGroups(adbControllerPath, "adb")
	if err != nil {
		return nil, fmt.Errorf("parse Go adb controller methods: %w", err)
	}
	win32Groups, win32Issues, err := parseGoControllerMethodGroups(win32ControllerPath, "win32")
	if err != nil {
		return nil, fmt.Errorf("parse Go win32 controller methods: %w", err)
	}

	goGroups := map[string]map[string]uint64{
		methodGroupAdbScreencap:   {},
		methodGroupAdbInput:       {},
		methodGroupWin32Screencap: {},
		methodGroupWin32Input:     {},
	}
	mergeMethodGroups(goGroups, adbGroups)
	mergeMethodGroups(goGroups, win32Groups)

	issues := make([]issue, 0, len(cIssues)+len(adbIssues)+len(win32Issues))
	issues = append(issues, cIssues...)
	issues = append(issues, adbIssues...)
	issues = append(issues, win32Issues...)

	for _, spec := range methodGroupSpecs {
		issues = append(issues, compareMethodGroupValues(spec.group, cGroups[spec.group], goGroups[spec.group])...)
	}

	return issues, nil
}

// parseCControllerMethodGroups extracts the adb/win32 method macros from
// MaaDef.h. The macros are evaluated with the shared C constant evaluator, so
// method values may depend on helper macros and enums anywhere in the header
// while width and unsigned semantics are preserved.
func parseCControllerMethodGroups(headerPath string) (map[string]map[string]uint64, []issue, error) {
	content, err := os.ReadFile(headerPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", headerPath, err)
	}

	env, err := evaluateCConstSources([]cConstSource{{path: headerPath, content: string(content)}})
	if err != nil {
		return nil, nil, fmt.Errorf("parse C constants: %w", err)
	}

	valuesByGroup := map[string]map[string]uint64{
		methodGroupAdbScreencap:   {},
		methodGroupAdbInput:       {},
		methodGroupWin32Screencap: {},
		methodGroupWin32Input:     {},
	}
	issues := make([]issue, 0)

	for _, spec := range methodGroupSpecs {
		rawDefs := map[string]string{}
		for name := range env.decls {
			logicalName, ok := strings.CutPrefix(name, spec.cPrefix)
			if !ok || logicalName == "" {
				continue
			}
			rawDefs[logicalName] = name
		}

		for _, logicalName := range sortedStringKeys(rawDefs) {
			fullName := rawDefs[logicalName]
			if failure := env.failures[fullName]; failure != "" {
				issues = append(issues, issue{
					section: sectionControllerMethod,
					message: fmt.Sprintf("[%s] failed to evaluate C method value: %s expr=%s (%s)", spec.group, logicalName, env.decls[fullName].expr, failure),
				})
				continue
			}
			rawValue, ok := env.values[fullName]
			if !ok {
				issues = append(issues, issue{
					section: sectionControllerMethod,
					message: fmt.Sprintf("[%s] failed to evaluate C method value: %s expr=%s (not evaluated)", spec.group, logicalName, env.decls[fullName].expr),
				})
				continue
			}
			normalized, err := normalizeCNumericValue(numericValue{kind: numericUnsigned, bits: 64, val: rawValue})
			if err != nil {
				issues = append(issues, issue{
					section: sectionControllerMethod,
					message: fmt.Sprintf("[%s] failed to evaluate C method value: %s expr=%s (%s)", spec.group, logicalName, env.decls[fullName].expr, err),
				})
				continue
			}
			value, err := uint64Value(numericValue{kind: numericUnsigned, bits: 64, val: normalized})
			if err != nil {
				issues = append(issues, issue{
					section: sectionControllerMethod,
					message: fmt.Sprintf("[%s] failed to evaluate C method value: %s expr=%s (%s)", spec.group, logicalName, env.decls[fullName].expr, err),
				})
				continue
			}
			valuesByGroup[spec.group][logicalName] = value
		}
	}

	return valuesByGroup, issues, nil
}

// parseGoControllerMethodGroups evaluates the exported Screencap*/Input*
// constants of one controller file with go/types, keeping Go conversions,
// widths, iota, and dependency semantics intact.
func parseGoControllerMethodGroups(goPath string, controller string) (map[string]map[string]uint64, []issue, error) {
	screencapGroup, inputGroup, err := goMethodGroups(controller)
	if err != nil {
		return nil, nil, err
	}

	evaluation, err := evaluateGoConstFile(goPath)
	if err != nil {
		return nil, nil, err
	}

	out := map[string]map[string]uint64{
		screencapGroup: {},
		inputGroup:     {},
	}
	issues := make([]issue, 0)

	names := make([]string, 0, len(evaluation.values)+len(evaluation.failures))
	for name := range evaluation.values {
		names = append(names, name)
	}
	for name := range evaluation.failures {
		if _, ok := evaluation.values[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)

	for _, name := range names {
		group, logicalName, matched := classifyGoMethodConst(name, screencapGroup, inputGroup)
		if !matched {
			continue
		}
		if failure, failed := evaluation.failures[name]; failed {
			issues = append(issues, issue{
				section: sectionControllerMethod,
				message: fmt.Sprintf("[%s] failed to evaluate Go method value: %s expr=%s (%s)", group, logicalName, failure.expr, failure.err),
			})
			continue
		}
		decl := evaluation.values[name]
		value, err := uint64Value(numericValue{kind: decl.kind, bits: decl.bits, val: decl.value})
		if err != nil {
			issues = append(issues, issue{
				section: sectionControllerMethod,
				message: fmt.Sprintf("[%s] failed to evaluate Go method value: %s expr=%s (%s)", group, logicalName, decl.expr, err),
			})
			continue
		}
		out[group][logicalName] = value
	}

	return out, issues, nil
}

func goMethodGroups(controller string) (string, string, error) {
	switch controller {
	case "adb":
		return methodGroupAdbScreencap, methodGroupAdbInput, nil
	case "win32":
		return methodGroupWin32Screencap, methodGroupWin32Input, nil
	default:
		return "", "", fmt.Errorf("unknown controller: %s", controller)
	}
}

func classifyGoMethodConst(name string, screencapGroup string, inputGroup string) (string, string, bool) {
	if strings.HasPrefix(name, "Screencap") {
		logicalName := strings.TrimPrefix(name, "Screencap")
		if logicalName == "" {
			return "", "", false
		}
		return screencapGroup, logicalName, true
	}
	if strings.HasPrefix(name, "Input") {
		logicalName := strings.TrimPrefix(name, "Input")
		if logicalName == "" {
			return "", "", false
		}
		return inputGroup, logicalName, true
	}
	return "", "", false
}

func compareMethodGroupValues(group string, cValues map[string]uint64, goValues map[string]uint64) []issue {
	if cValues == nil {
		cValues = map[string]uint64{}
	}
	if goValues == nil {
		goValues = map[string]uint64{}
	}

	normalizedC, ambiguous := normalizeCMethodMap(cValues)
	issues := make([]issue, 0)
	for _, key := range sortedStringKeysForSlices(ambiguous) {
		names := append([]string{}, ambiguous[key]...)
		sort.Strings(names)
		parts := make([]string, 0, len(names))
		for _, name := range names {
			parts = append(parts, fmt.Sprintf("%s(c=%d)", name, cValues[name]))
		}
		issues = append(issues, issue{
			section: sectionControllerMethod,
			message: fmt.Sprintf("[%s] C method name collision after underscore normalization: %s => %s", group, key, strings.Join(parts, ", ")),
		})
	}

	matchedNormalizedC := map[string]struct{}{}
	for _, goName := range sortedUint64Keys(goValues) {
		goValue := goValues[goName]
		cMethod, ok := normalizedC[goName]
		if !ok {
			if _, hasAmbiguous := ambiguous[goName]; hasAmbiguous {
				continue
			}
			issues = append(issues, issue{
				section: sectionControllerMethod,
				message: fmt.Sprintf("[%s] Go method not found in C: %s (go=%d)", group, goName, goValue),
			})
			continue
		}
		matchedNormalizedC[goName] = struct{}{}
		if cMethod.value != goValue {
			issues = append(issues, issue{
				section: sectionControllerMethod,
				message: fmt.Sprintf("[%s] method value mismatch: %s (go=%d c=%d)", group, goName, goValue, cMethod.value),
			})
		}
	}
	for _, key := range sortedNormalizedCKeys(normalizedC) {
		if _, ok := matchedNormalizedC[key]; ok {
			continue
		}
		cMethod := normalizedC[key]
		issues = append(issues, issue{
			section: sectionControllerMethod,
			message: fmt.Sprintf("[%s] C method not found in Go: %s (c=%d)", group, cMethod.name, cMethod.value),
		})
	}

	return issues
}

func normalizeCMethodMap(cValues map[string]uint64) (map[string]normalizedCMethod, map[string][]string) {
	normalized := make(map[string]normalizedCMethod, len(cValues))
	ambiguous := map[string][]string{}

	for _, cName := range sortedUint64Keys(cValues) {
		normalizedKey := normalizeCMethodNameForMatch(cName)
		if names, exists := ambiguous[normalizedKey]; exists {
			ambiguous[normalizedKey] = append(names, cName)
			continue
		}
		prev, exists := normalized[normalizedKey]
		if !exists {
			normalized[normalizedKey] = normalizedCMethod{
				name:  cName,
				value: cValues[cName],
			}
			continue
		}

		ambiguous[normalizedKey] = []string{prev.name, cName}
		delete(normalized, normalizedKey)
	}

	return normalized, ambiguous
}

func normalizeCMethodNameForMatch(name string) string {
	return strings.ReplaceAll(name, "_", "")
}

func mergeMethodGroups(dst map[string]map[string]uint64, src map[string]map[string]uint64) {
	for group, methods := range src {
		if _, ok := dst[group]; !ok {
			dst[group] = map[string]uint64{}
		}
		for name, value := range methods {
			dst[group][name] = value
		}
	}
}

func parseCDefineExprs(content string) map[string]string {
	out := map[string]string{}
	for _, decl := range parseCDefineDecls(content) {
		out[decl.name] = decl.expr
	}
	return out
}

// parseCDefineDecls extracts every object-like macro declaration in source
// order, retaining duplicates for the constant evaluator's conflict checks.
func parseCDefineDecls(content string) []cConstDecl {
	lines := strings.Split(content, "\n")
	out := make([]cConstDecl, 0)

	var current strings.Builder
	flush := func() {
		line := normalizeSpaces(current.String())
		current.Reset()
		if line == "" {
			return
		}
		name, expr, ok := parseCDefineLine(line)
		if !ok {
			return
		}
		out = append(out, cConstDecl{name: name, expr: expr})
	}

	for _, rawLine := range lines {
		line := strings.TrimSpace(strings.TrimRight(rawLine, "\r"))
		if line == "" {
			continue
		}
		hasContinuation := strings.HasSuffix(line, "\\")
		line = strings.TrimSpace(strings.TrimSuffix(line, "\\"))
		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(line)
		if !hasContinuation {
			flush()
		}
	}
	if current.Len() > 0 {
		flush()
	}
	return out
}

func parseCDefineLine(line string) (string, string, bool) {
	if !strings.HasPrefix(line, "#define ") {
		return "", "", false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(line, "#define"))
	if rest == "" {
		return "", "", false
	}

	fields := strings.Fields(rest)
	if len(fields) < 2 {
		return "", "", false
	}
	name := fields[0]
	if strings.Contains(name, "(") {
		return "", "", false
	}

	expr := strings.TrimSpace(rest[len(name):])
	if expr == "" {
		return "", "", false
	}
	return name, normalizeSpaces(expr), true
}

func formatGoExpr(expr ast.Expr) string {
	if expr == nil {
		return ""
	}
	var builder strings.Builder
	if err := printer.Fprint(&builder, token.NewFileSet(), expr); err != nil {
		return ""
	}
	return normalizeSpaces(builder.String())
}

func sortedUint64Keys(m map[string]uint64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedStringKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedStringKeysForSlices(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedNormalizedCKeys(m map[string]normalizedCMethod) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

var cMethodIntSuffixRe = regexp.MustCompile(`(?i)\b(0x[0-9a-f]+|[0-9]+)(?:ull|llu|ul|lu|ll|u|l)\b`)
