package checker

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strconv"
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

type uintExprDef struct {
	expr      ast.Expr
	raw       string
	iotaValue uint64
}

type normalizedCMethod struct {
	name  string
	value uint64
}

type unknownIdentifierError struct {
	name string
}

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

func parseCControllerMethodGroups(headerPath string) (map[string]map[string]uint64, []issue, error) {
	data, err := os.ReadFile(headerPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", headerPath, err)
	}

	defines := parseCDefineExprs(removeCComments(string(data)))
	rawByGroup := map[string]map[string]string{
		methodGroupAdbScreencap:   {},
		methodGroupAdbInput:       {},
		methodGroupWin32Screencap: {},
		methodGroupWin32Input:     {},
	}
	for macroName, expr := range defines {
		for _, spec := range methodGroupSpecs {
			if !strings.HasPrefix(macroName, spec.cPrefix) {
				continue
			}
			logicalName := strings.TrimPrefix(macroName, spec.cPrefix)
			if logicalName == "" {
				continue
			}
			rawByGroup[spec.group][logicalName] = expr
			break
		}
	}

	valuesByGroup := map[string]map[string]uint64{
		methodGroupAdbScreencap:   {},
		methodGroupAdbInput:       {},
		methodGroupWin32Screencap: {},
		methodGroupWin32Input:     {},
	}
	issues := make([]issue, 0)

	for _, spec := range methodGroupSpecs {
		rawDefs := rawByGroup[spec.group]
		if len(rawDefs) == 0 {
			continue
		}

		replacements := make(map[string]string, len(rawDefs))
		for logicalName := range rawDefs {
			replacements[spec.cPrefix+logicalName] = logicalName
		}

		exprDefs := map[string]uintExprDef{}
		parseFailures := map[string]string{}

		for logicalName, rawExpr := range rawDefs {
			normalized := normalizeCMacroExpr(rawExpr, replacements)
			parsedExpr, parseErr := parser.ParseExpr(normalized)
			if parseErr != nil {
				parseFailures[logicalName] = parseErr.Error()
				continue
			}
			exprDefs[logicalName] = uintExprDef{
				expr: parsedExpr,
				raw:  normalized,
			}
		}

		groupValues, evalFailures := evaluateUintExprDefs(exprDefs)
		valuesByGroup[spec.group] = groupValues

		for _, logicalName := range sortedStringKeys(parseFailures) {
			rawExpr := normalizeCMacroExpr(rawDefs[logicalName], replacements)
			issues = append(issues, issue{
				section: sectionControllerMethod,
				message: fmt.Sprintf("[%s] failed to evaluate C method value: %s expr=%s (%s)", spec.group, logicalName, rawExpr, parseFailures[logicalName]),
			})
		}
		for _, logicalName := range sortedStringKeys(evalFailures) {
			rawExpr := exprDefs[logicalName].raw
			issues = append(issues, issue{
				section: sectionControllerMethod,
				message: fmt.Sprintf("[%s] failed to evaluate C method value: %s expr=%s (%s)", spec.group, logicalName, rawExpr, evalFailures[logicalName]),
			})
		}
	}

	return valuesByGroup, issues, nil
}

func parseGoControllerMethodGroups(goPath string, controller string) (map[string]map[string]uint64, []issue, error) {
	screencapGroup, inputGroup, err := goMethodGroups(controller)
	if err != nil {
		return nil, nil, err
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, goPath, nil, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", goPath, err)
	}

	type goMethodMeta struct {
		group       string
		logicalName string
	}

	defs := map[string]uintExprDef{}
	metas := map[string]goMethodMeta{}

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}

		iotaValue := uint64(0)
		var prevValues []ast.Expr
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				iotaValue++
				continue
			}

			values := vs.Values
			if len(values) > 0 {
				prevValues = values
			} else {
				values = prevValues
			}
			if len(values) == 0 {
				iotaValue++
				continue
			}

			for i, ident := range vs.Names {
				if ident == nil || ident.Name == "" || !ident.IsExported() {
					continue
				}
				group, logicalName, matched := classifyGoMethodConst(ident.Name, screencapGroup, inputGroup)
				if !matched {
					continue
				}
				exprIdx := i
				if exprIdx >= len(values) {
					exprIdx = len(values) - 1
				}
				expr := values[exprIdx]
				defs[ident.Name] = uintExprDef{
					expr:      expr,
					raw:       formatGoExpr(expr),
					iotaValue: iotaValue,
				}
				metas[ident.Name] = goMethodMeta{
					group:       group,
					logicalName: logicalName,
				}
			}

			iotaValue++
		}
	}

	values, failures := evaluateUintExprDefs(defs)
	out := map[string]map[string]uint64{
		screencapGroup: {},
		inputGroup:     {},
	}

	for name, value := range values {
		meta, ok := metas[name]
		if !ok {
			continue
		}
		out[meta.group][meta.logicalName] = value
	}

	issues := make([]issue, 0, len(failures))
	for _, name := range sortedStringKeys(failures) {
		meta, ok := metas[name]
		if !ok {
			continue
		}
		issues = append(issues, issue{
			section: sectionControllerMethod,
			message: fmt.Sprintf("[%s] failed to evaluate Go method value: %s expr=%s (%s)", meta.group, meta.logicalName, defs[name].raw, failures[name]),
		})
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
	lines := strings.Split(content, "\n")
	out := map[string]string{}

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
		out[name] = expr
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

func normalizeCMacroExpr(expr string, replacements map[string]string) string {
	noSuffix := cMethodIntSuffixRe.ReplaceAllString(expr, "$1")
	return cMethodIdentRe.ReplaceAllStringFunc(noSuffix, func(token string) string {
		if replacement, ok := replacements[token]; ok {
			return replacement
		}
		return token
	})
}

func evaluateUintExprDefs(defs map[string]uintExprDef) (map[string]uint64, map[string]string) {
	values := map[string]uint64{}
	failures := map[string]string{}
	pending := make(map[string]uintExprDef, len(defs))
	for name, def := range defs {
		pending[name] = def
	}

	for {
		if len(pending) == 0 {
			break
		}
		progress := false
		for name, def := range pending {
			value, err := evalUintExpr(def.expr, values, def.iotaValue)
			if err == nil {
				values[name] = value
				delete(pending, name)
				progress = true
				continue
			}
			var unknownErr *unknownIdentifierError
			if errors.As(err, &unknownErr) {
				continue
			}
			failures[name] = err.Error()
			delete(pending, name)
			progress = true
		}
		if progress {
			continue
		}

		for name, def := range pending {
			_, err := evalUintExpr(def.expr, values, def.iotaValue)
			if err != nil {
				failures[name] = err.Error()
			}
		}
		break
	}

	return values, failures
}

func evalUintExpr(expr ast.Expr, env map[string]uint64, iotaValue uint64) (uint64, error) {
	switch node := expr.(type) {
	case *ast.ParenExpr:
		return evalUintExpr(node.X, env, iotaValue)
	case *ast.BasicLit:
		if node.Kind != token.INT {
			return 0, fmt.Errorf("unsupported literal kind: %s", node.Kind.String())
		}
		value, err := strconv.ParseUint(node.Value, 0, 64)
		if err != nil {
			return 0, fmt.Errorf("parse int literal %q: %w", node.Value, err)
		}
		return value, nil
	case *ast.Ident:
		if node.Name == "iota" {
			return iotaValue, nil
		}
		value, ok := env[node.Name]
		if !ok {
			return 0, &unknownIdentifierError{name: node.Name}
		}
		return value, nil
	case *ast.UnaryExpr:
		value, err := evalUintExpr(node.X, env, iotaValue)
		if err != nil {
			return 0, err
		}
		switch node.Op {
		case token.ADD:
			return value, nil
		case token.SUB:
			return ^value + 1, nil
		case token.XOR, token.TILDE:
			return ^value, nil
		default:
			return 0, fmt.Errorf("unsupported unary operator: %s", node.Op.String())
		}
	case *ast.BinaryExpr:
		left, err := evalUintExpr(node.X, env, iotaValue)
		if err != nil {
			return 0, err
		}
		right, err := evalUintExpr(node.Y, env, iotaValue)
		if err != nil {
			return 0, err
		}
		switch node.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		case token.REM:
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			return left % right, nil
		case token.SHL:
			return left << right, nil
		case token.SHR:
			return left >> right, nil
		case token.AND:
			return left & right, nil
		case token.OR:
			return left | right, nil
		case token.XOR:
			return left ^ right, nil
		default:
			return 0, fmt.Errorf("unsupported binary operator: %s", node.Op.String())
		}
	case *ast.CallExpr:
		if len(node.Args) != 1 {
			return 0, fmt.Errorf("unsupported call expression")
		}
		return evalUintExpr(node.Args[0], env, iotaValue)
	default:
		return 0, fmt.Errorf("unsupported expression type: %T", expr)
	}
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

var (
	cMethodIntSuffixRe = regexp.MustCompile(`(?i)\b(0x[0-9a-f]+|[0-9]+)(?:ull|llu|ul|lu|ll|u|l)\b`)
	cMethodIdentRe     = regexp.MustCompile(`\b[A-Za-z_][A-Za-z0-9_]*\b`)
)
