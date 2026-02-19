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

func checkCustomControllerConsistency(headerPath string) ([]issue, error) {
	goSigs, err := parseCustomControllerGo()
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

func parseCustomControllerGo() (map[string]methodSig, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "custom_controller.go", nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse custom_controller.go: %w", err)
	}

	result := map[string]methodSig{}
	typeDefs := collectGoTypeDefs(file)
	found := false

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name.Name != "CustomController" {
				continue
			}
			iface, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				return nil, fmt.Errorf("CustomController is not an interface")
			}
			for _, m := range iface.Methods.List {
				if len(m.Names) != 1 {
					continue
				}
				fnType, ok := m.Type.(*ast.FuncType)
				if !ok {
					continue
				}
				name := m.Names[0].Name
				params := parseGoFieldTypesCanonical(fnType.Params, typeDefs)
				returns := parseGoFieldTypesCanonical(fnType.Results, typeDefs)
				result[name] = methodSig{params: params, returns: returns}
			}
			found = true
		}
	}

	if !found {
		return nil, fmt.Errorf("CustomController interface not found")
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
	content := string(data)
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
	lineRe := regexp.MustCompile(`(?m)^\s*([A-Za-z0-9_]+)\s*\(\s*\*\s*([a-z_]+)\s*\)\s*\(([^)]*)\)\s*;`)
	matches := lineRe.FindAllStringSubmatch(block, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no callback field found in callbacks struct")
	}

	result := make(map[string]methodSig, len(matches))
	for _, m := range matches {
		if len(m) < 4 {
			continue
		}
		retType := strings.TrimSpace(m[1])
		cName := strings.TrimSpace(m[2])
		paramsRaw := strings.TrimSpace(m[3])

		goName := callbackNameToGoMethod(cName)
		params := parseControllerParams(paramsRaw, aliases)
		returns := deriveControllerReturns(retType, params, aliases)
		goParams := deriveControllerParams(params)

		result[goName] = methodSig{params: goParams, returns: returns}
	}

	return result, nil
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
