package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
)

func checkCustomControllerConsistency(headerPath string) ([]issue, error) {
	goSigs, err := parseCustomControllerGo()
	if err != nil {
		return nil, err
	}

	headerSigs, err := parseCustomControllerHeader(headerPath)
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
				params := parseGoFieldTypes(fnType.Params)
				returns := parseGoFieldTypes(fnType.Results)
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

func parseGoFieldTypes(fields *ast.FieldList) []string {
	if fields == nil || len(fields.List) == 0 {
		return []string{}
	}
	out := make([]string, 0)
	for _, f := range fields.List {
		t := normalizeGoTypeExpr(f.Type)
		count := 1
		if len(f.Names) > 0 {
			count = len(f.Names)
		}
		for i := 0; i < count; i++ {
			out = append(out, t)
		}
	}
	return out
}

func normalizeGoTypeExpr(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "bool":
			return "bool"
		case "string":
			return "string"
		case "int32":
			return "int32"
		case "ControllerFeature":
			return "controller_feature"
		default:
			return t.Name
		}
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok && x.Name == "image" && t.Sel.Name == "Image" {
			return "image"
		}
		return fmt.Sprintf("%s.%s", exprToString(t.X), t.Sel.Name)
	default:
		return exprToString(expr)
	}
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

func parseCustomControllerHeader(headerPath string) (map[string]methodSig, error) {
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
		paramsRaw := removeCComments(strings.TrimSpace(m[3]))

		goName := callbackNameToGoMethod(cName)
		params := parseCParams(paramsRaw)
		returns := deriveGoReturnsFromC(retType, params)
		goParams := deriveGoParamsFromC(params)

		result[goName] = methodSig{params: goParams, returns: returns}
	}

	return result, nil
}

func removeCComments(s string) string {
	re := regexp.MustCompile(`(?s)/\*.*?\*/`)
	return re.ReplaceAllString(s, "")
}

func parseCParams(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "void" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := normalizeCParamType(strings.TrimSpace(p))
		if t == "" {
			continue
		}
		out = append(out, t)
	}
	return out
}

func normalizeCParamType(param string) string {
	switch {
	case strings.Contains(param, "void* trans_arg"):
		return "trans_arg"
	case strings.Contains(param, "const char*"):
		return "string"
	case strings.Contains(param, "int32_t"):
		return "int32"
	case strings.Contains(param, "MaaStringBuffer*"):
		return "string_buffer"
	case strings.Contains(param, "MaaImageBuffer*"):
		return "image_buffer"
	default:
		return "unknown"
	}
}

func deriveGoParamsFromC(cParams []string) []string {
	out := make([]string, 0, len(cParams))
	for _, p := range cParams {
		switch p {
		case "trans_arg", "string_buffer", "image_buffer":
			continue
		case "string", "int32":
			out = append(out, p)
		default:
			out = append(out, p)
		}
	}
	return out
}

func deriveGoReturnsFromC(cReturn string, cParams []string) []string {
	switch cReturn {
	case "MaaControllerFeature":
		return []string{"controller_feature"}
	case "MaaBool":
		hasStringBuffer := false
		hasImageBuffer := false
		for _, p := range cParams {
			if p == "string_buffer" {
				hasStringBuffer = true
			}
			if p == "image_buffer" {
				hasImageBuffer = true
			}
		}
		switch {
		case hasStringBuffer:
			return []string{"string", "bool"}
		case hasImageBuffer:
			return []string{"image", "bool"}
		default:
			return []string{"bool"}
		}
	default:
		return []string{"unknown"}
	}
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
