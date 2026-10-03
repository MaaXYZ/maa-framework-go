package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func checkCustomControllerConsistency(headerPath string, customControllerPath string) ([]issue, error) {
	goSigs, err := parseCustomControllerGo(customControllerPath)
	if err != nil {
		return nil, err
	}

	headerRoot := filepath.Clean(filepath.Join(filepath.Dir(headerPath), "..", ".."))
	aliases, err := parseCTypedefAliases(headerRoot)
	if err != nil {
		return nil, fmt.Errorf("parse C typedef aliases: %w", err)
	}

	headerSigs, err := parseCustomControllerHeader(headerPath, aliases)
	if err != nil {
		return nil, err
	}

	issues := make([]issue, 0)

	for goName := range goSigs {
		if _, ok := headerSigs[goName]; !ok {
			issues = append(issues, issue{
				section: sectionController,
				message: fmt.Sprintf("Go interface method not found in C callbacks: %s", goName),
			})
		}
	}
	for cName := range headerSigs {
		if _, ok := goSigs[cName]; !ok {
			issues = append(issues, issue{
				section: sectionController,
				message: fmt.Sprintf("C callback missing in Go interface: %s", cName),
			})
		}
	}

	for name, cSig := range headerSigs {
		goSig, ok := goSigs[name]
		if !ok {
			continue
		}
		if unsupportedType, found := findUnsupportedType(goSig.params); found {
			issues = append(issues, issue{
				section: sectionController,
				message: fmt.Sprintf("%s has unsupported Go param type expression: %s (normalized=%v)", name, unsupportedType, goSig.params),
			})
		} else if !sameStringSlice(goSig.params, cSig.params) {
			issues = append(issues, issue{
				section: sectionController,
				message: fmt.Sprintf("%s param mismatch: go=%v c=%v", name, goSig.params, cSig.params),
			})
		}
		if unsupportedType, found := findUnsupportedType(goSig.returns); found {
			issues = append(issues, issue{
				section: sectionController,
				message: fmt.Sprintf("%s has unsupported Go return type expression: %s (normalized=%v)", name, unsupportedType, goSig.returns),
			})
		} else if !sameStringSlice(goSig.returns, cSig.returns) {
			issues = append(issues, issue{
				section: sectionController,
				message: fmt.Sprintf("%s return mismatch: go=%v c=%v", name, goSig.returns, cSig.returns),
			})
		}
	}

	return issues, nil
}

func parseCustomControllerGo(customControllerPath string) (map[string]methodSig, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, customControllerPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", customControllerPath, err)
	}

	typeDefs := collectGoTypeDefs(file)
	result := map[string]methodSig{}
	stack := map[string]bool{}
	var expand func(string) error
	expand = func(name string) error {
		if stack[name] {
			return fmt.Errorf("CustomController embedded interface cycle at %s", name)
		}
		stack[name] = true
		defer delete(stack, name)
		expr := typeDefs[name]
		if target, ok := expr.(*ast.Ident); ok {
			return expand(target.Name)
		}
		iface, ok := expr.(*ast.InterfaceType)
		if !ok {
			return fmt.Errorf("CustomController: unresolved or unsupported embedded interface %s", name)
		}
		for _, field := range iface.Methods.List {
			if len(field.Names) == 0 {
				target, ok := field.Type.(*ast.Ident)
				if !ok {
					return fmt.Errorf("CustomController: unsupported embedded interface at %s", fset.Position(field.Pos()))
				}
				if err := expand(target.Name); err != nil {
					return err
				}
				continue
			}
			fn, ok := field.Type.(*ast.FuncType)
			if !ok || len(field.Names) != 1 {
				return fmt.Errorf("CustomController: unsupported method at %s", fset.Position(field.Pos()))
			}
			sig := methodSig{params: parseGoFieldTypesCanonical(fn.Params, typeDefs), returns: parseGoFieldTypesCanonical(fn.Results, typeDefs)}
			method := field.Names[0].Name
			if prior, exists := result[method]; exists && (!sameStringSlice(prior.params, sig.params) || !sameStringSlice(prior.returns, sig.returns)) {
				return fmt.Errorf("CustomController: conflicting embedded method %s", method)
			}
			result[method] = sig
		}
		return nil
	}
	if typeDefs["CustomController"] == nil {
		return nil, fmt.Errorf("CustomController interface not found")
	}
	if err := expand("CustomController"); err != nil {
		return nil, err
	}
	return result, nil
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	default:
		return fmt.Sprintf("<unsupported:%T>", expr)
	}
}

func parseCustomControllerHeader(headerPath string, aliases map[string]string) (map[string]methodSig, error) {
	data, err := os.ReadFile(headerPath)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", headerPath, err)
	}
	content := removeCComments(string(data))
	start := strings.Index(content, "struct MaaCustomControllerCallbacks")
	if start < 0 {
		return nil, fmt.Errorf("MaaCustomControllerCallbacks struct not found")
	}
	open := strings.Index(content[start:], "{")
	if open < 0 {
		return nil, fmt.Errorf("callbacks struct open brace not found")
	}
	open += start
	close := strings.Index(content[open:], "};")
	if close < 0 {
		return nil, fmt.Errorf("callbacks struct close not found")
	}
	close += open

	block := content[open+1 : close]
	callbacks, err := parseControllerCallbackFields(block)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", headerPath, err)
	}
	if len(callbacks) == 0 {
		return nil, fmt.Errorf("no callback field found in callbacks struct")
	}

	result := make(map[string]methodSig, len(callbacks))
	for _, cb := range callbacks {
		goName := callbackNameToGoMethod(cb.name)
		params := parseControllerParams(cb.paramsRaw, aliases)
		if err := validateControllerAdapter(cb, params); err != nil {
			return nil, fmt.Errorf("%s: %w", headerPath, err)
		}
		returns := deriveControllerReturns(cb.retType, params, aliases)
		goParams := deriveControllerParams(params)

		if _, exists := result[goName]; exists {
			return nil, fmt.Errorf("duplicate controller callback %s", goName)
		}
		result[goName] = methodSig{params: goParams, returns: returns}
	}

	return result, nil
}

type controllerCallbackField struct {
	retType   string
	name      string
	paramsRaw string
}

func parseControllerCallbackFields(block string) ([]controllerCallbackField, error) {
	cleaned := removeCComments(block)
	stmts := splitControllerFieldStatements(cleaned)
	out := make([]controllerCallbackField, 0, len(stmts))
	for _, stmt := range stmts {
		retType, name, paramsRaw, ok := parseControllerCallbackField(stmt)
		if !ok {
			return nil, fmt.Errorf("unsupported controller callback field: %s", stmt)
		}
		out = append(out, controllerCallbackField{
			retType:   retType,
			name:      name,
			paramsRaw: paramsRaw,
		})
	}
	return out, nil
}

func splitControllerFieldStatements(block string) []string {
	parts := make([]string, 0)
	var current strings.Builder
	depth := 0
	for _, ch := range block {
		switch ch {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ';':
			if depth == 0 {
				stmt := normalizeSpaces(current.String())
				if stmt != "" {
					parts = append(parts, stmt)
				}
				current.Reset()
				continue
			}
		}
		current.WriteRune(ch)
	}
	if tail := normalizeSpaces(current.String()); tail != "" {
		parts = append(parts, tail)
	}
	return parts
}

func parseControllerCallbackField(stmt string) (retType string, name string, paramsRaw string, ok bool) {
	s := normalizeSpaces(stmt)
	if s == "" {
		return "", "", "", false
	}

	pointer := callbackPointerRe.FindStringIndex(s)
	if pointer == nil {
		return "", "", "", false
	}
	retType = strings.TrimSpace(s[:pointer[0]])
	if retType == "" {
		return "", "", "", false
	}

	rest := strings.TrimSpace(s[pointer[1]:])
	closeName := strings.IndexByte(rest, ')')
	if closeName < 0 {
		return "", "", "", false
	}
	name = strings.TrimSpace(rest[:closeName])
	if !cIdentifierRe.MatchString(name) {
		return "", "", "", false
	}

	afterName := strings.TrimSpace(rest[closeName+1:])
	if afterName == "" || afterName[0] != '(' {
		return "", "", "", false
	}
	closeParams := findMatchingParen(afterName, 0)
	if closeParams <= 0 {
		return "", "", "", false
	}
	paramsRaw = strings.TrimSpace(afterName[1:closeParams])
	if trailing := strings.TrimSpace(afterName[closeParams+1:]); trailing != "" {
		return "", "", "", false
	}
	return retType, name, paramsRaw, true
}

func findMatchingParen(s string, start int) int {
	if start < 0 || start >= len(s) || s[start] != '(' {
		return -1
	}
	depth := 0
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

type controllerParam struct {
	name      string
	raw       string
	canonical string
}

func parseControllerParams(raw string, aliases map[string]string) []controllerParam {
	cleaned := removeCComments(raw)
	if strings.TrimSpace(cleaned) == "" || strings.TrimSpace(cleaned) == "void" {
		return []controllerParam{}
	}
	parts := splitCSV(cleaned)
	out := make([]controllerParam, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		out = append(out, controllerParam{
			name:      extractCParamName(p),
			raw:       p,
			canonical: normalizeCTypeCanonical(stripCParamName(p), aliases),
		})
	}
	return out
}

func extractCParamName(raw string) string {
	s := normalizeSpaces(strings.TrimSpace(raw))
	if s == "" {
		return ""
	}
	if strings.Contains(s, "(*") {
		return ""
	}
	m := regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)$`).FindStringSubmatch(s)
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

var callbackPointerRe = regexp.MustCompile(`\(\s*(?:MAA_CALL\s*)?\*\s*`)
var cIdentifierRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Validate the raw ABI before removing the context and converting one output
// buffer into a Go return value. Only the documented adapters are supported.
func validateControllerAdapter(cb controllerCallbackField, params []controllerParam) error {
	output := ""
	switch cb.name {
	case "request_uuid", "shell", "get_info":
		output = "MaaStringBuffer"
	case "screencap":
		output = "MaaImageBuffer"
	}
	contextIndex := len(params) - 1
	if output != "" {
		contextIndex--
	}
	contexts, buffers := 0, 0
	for i, p := range params {
		base, depth := splitCBaseAndPtr(stripCParamName(p.raw))
		if p.name == "trans_arg" {
			contexts++
			if base != "void" || depth != 1 || i != contextIndex {
				return fmt.Errorf("%s: unsupported trans_arg ABI (expected void* at parameter %d)", cb.name, contextIndex+1)
			}
		}
		if base == "MaaStringBuffer" || base == "MaaImageBuffer" {
			buffers++
			if base != output || depth != 1 || i != len(params)-1 {
				return fmt.Errorf("%s: unsupported output buffer ABI", cb.name)
			}
		}
	}
	if contexts != 1 {
		return fmt.Errorf("%s: expected exactly one trans_arg", cb.name)
	}
	expectedBuffers := 0
	if output != "" {
		expectedBuffers = 1
	}
	if buffers != expectedBuffers {
		return fmt.Errorf("%s: expected %d output buffer, got %d", cb.name, expectedBuffers, buffers)
	}
	if output != "" && strings.TrimSpace(cb.retType) != "MaaBool" {
		return fmt.Errorf("%s: output adapter requires MaaBool return", cb.name)
	}
	return nil
}

func deriveControllerParams(params []controllerParam) []string {
	out := make([]string, 0, len(params))
	for _, p := range params {
		if p.name == "trans_arg" {
			continue
		}
		if strings.Contains(p.raw, "MaaStringBuffer") || strings.Contains(p.raw, "MaaImageBuffer") {
			continue
		}
		out = append(out, p.canonical)
	}
	return out
}

func deriveControllerReturns(cReturn string, params []controllerParam, aliases map[string]string) []string {
	ret := normalizeCTypeCanonical(cReturn, aliases)
	if ret == "void" {
		return []string{}
	}
	if ret == "bool" {
		hasStringBuffer := false
		hasImageBuffer := false
		for _, p := range params {
			if strings.Contains(p.raw, "MaaStringBuffer") {
				hasStringBuffer = true
			}
			if strings.Contains(p.raw, "MaaImageBuffer") {
				hasImageBuffer = true
			}
		}
		switch {
		case hasStringBuffer:
			return []string{"cstring", "bool"}
		case hasImageBuffer:
			return []string{"image", "bool"}
		default:
			return []string{"bool"}
		}
	}
	return []string{ret}
}

func callbackNameToGoMethod(name string) string {
	switch name {
	case "request_uuid":
		return "RequestUUID"
	case "get_features":
		return "GetFeature"
	default:
		parts := strings.Split(name, "_")
		for i := range parts {
			if parts[i] == "" {
				continue
			}
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
		return strings.Join(parts, "")
	}
}

func sameStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func findUnsupportedType(types []string) (string, bool) {
	for _, t := range types {
		if strings.HasPrefix(t, "<unsupported:") {
			return t, true
		}
	}
	return "", false
}
