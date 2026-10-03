package checker

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type goVarSigMap map[string]methodSig

func parseGoVarFuncSignatures(parsedFile *ast.File) goVarSigMap {
	sigs, _ := parseGoVarFuncSignaturesWithLoc(parsedFile, nil, "")
	return sigs
}

type goVarDeclLoc struct {
	file string
	line int
}

func parseGoVarFuncSignaturesWithLoc(parsedFile *ast.File, fset *token.FileSet, sourceFile string) (goVarSigMap, map[string]goVarDeclLoc) {
	return collectGoFuncSignatures([]*ast.File{parsedFile}, fset, sourceFile)
}

// collectGoFuncSignatures resolves the package declarations before examining
// registrations, so declaration order and file boundaries do not hide types.
func collectGoFuncSignatures(files []*ast.File, fset *token.FileSet, sourceFile string) (goVarSigMap, map[string]goVarDeclLoc) {
	types := map[string]ast.Expr{}
	variables := map[string]ast.Expr{}
	functions := map[string]*ast.FuncType{}
	locs := map[string]goVarDeclLoc{}
	for _, file := range files {
		for name, expr := range collectGoTypeDefs(file) {
			types[name] = expr
		}
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil {
				functions[fn.Name.Name] = fn.Type
			}
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				vs := spec.(*ast.ValueSpec)
				for i, name := range vs.Names {
					expr := vs.Type
					if expr == nil && i < len(vs.Values) {
						expr = vs.Values[i]
					}
					variables[name.Name] = expr
					if fset != nil {
						pos := fset.Position(name.Pos())
						fileName := sourceFile
						if pos.Filename != "" {
							fileName = filepath.Clean(pos.Filename)
						}
						locs[name.Name] = goVarDeclLoc{file: fileName, line: pos.Line}
					}
				}
			}
		}
	}
	var resolve func(ast.Expr, map[string]bool) *ast.FuncType
	resolve = func(expr ast.Expr, seen map[string]bool) *ast.FuncType {
		switch e := expr.(type) {
		case *ast.FuncType:
			return e
		case *ast.FuncLit:
			return e.Type
		case *ast.ParenExpr:
			return resolve(e.X, seen)
		case *ast.CallExpr:
			// Only a named function type conversion can establish this type.
			if name, ok := e.Fun.(*ast.Ident); ok && types[name.Name] != nil {
				return resolve(types[name.Name], seen)
			}
		case *ast.Ident:
			if seen[e.Name] {
				return nil
			}
			seen[e.Name] = true
			defer delete(seen, e.Name)
			if target := types[e.Name]; target != nil {
				return resolve(target, seen)
			}
			if fn := functions[e.Name]; fn != nil {
				return fn
			}
			return resolve(variables[e.Name], seen)
		}
		return nil
	}
	out := make(goVarSigMap)
	for name, expr := range variables {
		if fn := resolve(expr, map[string]bool{}); fn != nil {
			out[name] = methodSig{params: parseGoFieldTypesCanonical(fn.Params, types), returns: parseGoFieldTypesCanonical(fn.Results, types)}
		}
	}
	return out, locs
}

func collectGoTypeDefs(parsedFile *ast.File) map[string]ast.Expr {
	out := make(map[string]ast.Expr)
	for _, decl := range parsedFile.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name == nil || ts.Name.Name == "" || ts.Type == nil {
				continue
			}
			out[ts.Name.Name] = ts.Type
		}
	}
	return out
}

func parseGoFieldTypesCanonical(fields *ast.FieldList, typeDefs map[string]ast.Expr) []string {
	if fields == nil || len(fields.List) == 0 {
		return []string{}
	}
	out := make([]string, 0)
	for _, f := range fields.List {
		t := normalizeGoTypeExprCanonical(f.Type, typeDefs, map[string]struct{}{})
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

func normalizeGoTypeExprCanonical(expr ast.Expr, typeDefs map[string]ast.Expr, seen map[string]struct{}) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return normalizeGoIdentType(t.Name, typeDefs, seen)
	case *ast.SelectorExpr:
		if x, ok := t.X.(*ast.Ident); ok {
			if x.Name == "unsafe" && t.Sel != nil && t.Sel.Name == "Pointer" {
				return "ptr"
			}
			if x.Name == "image" && t.Sel != nil && t.Sel.Name == "Image" {
				return "image"
			}
		}
		return "named:" + exprToString(t)
	case *ast.StarExpr:
		return "ptr"
	case *ast.InterfaceType:
		return "interface"
	default:
		return fmt.Sprintf("<unsupported:%T>", expr)
	}
}

func normalizeGoIdentType(name string, typeDefs map[string]ast.Expr, seen map[string]struct{}) string {
	switch name {
	case "bool":
		return "bool"
	case "string":
		return "cstring"
	case "int32":
		return "int32"
	case "int64":
		return "int64"
	case "uint8", "byte":
		return "uint8"
	case "uint16":
		return "uint16"
	case "uint32":
		return "uint32"
	case "uint64":
		return "uint64"
	case "uintptr":
		return "ptr"
	}
	if strings.HasSuffix(name, "Callback") {
		return "callback:" + name
	}
	target, ok := typeDefs[name]
	if !ok {
		return "named:" + name
	}
	if _, exists := seen[name]; exists {
		return "named:" + name
	}
	seen[name] = struct{}{}
	defer delete(seen, name)
	if _, ok := target.(*ast.FuncType); ok {
		return "callback:" + name
	}
	return normalizeGoTypeExprCanonical(target, typeDefs, seen)
}

func parseCTypedefAliases(headerDir string) (map[string]string, error) {
	aliases := make(map[string]string)
	err := filepath.WalkDir(headerDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(path) != ".h" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := removeCComments(string(data))
		matches := cTypedefRe.FindAllStringSubmatch(content, -1)
		for _, m := range matches {
			if len(m) != 3 {
				continue
			}
			rhs := normalizeSpaces(strings.TrimSpace(m[1]))
			lhs := strings.TrimSpace(m[2])
			if strings.Contains(rhs, "(*") {
				continue
			}
			if rhs == "" || lhs == "" {
				continue
			}
			aliases[lhs] = rhs
		}
		return nil
	})
	return aliases, err
}

func parseHeaderFunctionSignatures(headerDir string, aliases map[string]string) (map[string]map[string]methodSig, error) {
	result := map[string]map[string]methodSig{
		"framework":    {},
		"toolkit":      {},
		"agent_server": {},
		"agent_client": {},
	}
	err := filepath.WalkDir(headerDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || filepath.Ext(path) != ".h" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := removeCPreprocessorLines(removeCComments(string(data)))
		stmts := strings.Split(content, ";")
		for _, raw := range stmts {
			stmt := normalizeSpaces(strings.TrimSpace(raw))
			if stmt == "" || !strings.Contains(stmt, "_API") {
				continue
			}
			if strings.Contains(stmt, "MAA_DEPRECATED") {
				continue
			}
			m := cAPIMacroInStmtRe.FindStringSubmatch(stmt)
			if len(m) != 3 {
				if cAPINameRe.MatchString(stmt) {
					return fmt.Errorf("%s: unsupported exported C declaration: %s", path, stmt)
				}
				continue
			}
			module := moduleFromAPIMacro(strings.TrimSpace(m[1]))
			decl := normalizeSpaces(strings.TrimSpace(m[2]))
			if module == "" || decl == "" {
				continue
			}
			sig, name, ok := parseCFunctionDecl(decl, aliases)
			if !ok {
				return fmt.Errorf("%s: unsupported exported C declaration: %s", path, decl)
			}
			if prior, exists := result[module][name]; exists && (!sameStringSlice(prior.params, sig.params) || !sameStringSlice(prior.returns, sig.returns)) {
				return fmt.Errorf("%s: conflicting C declarations for %s", path, name)
			}
			result[module][name] = sig
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func parseCFunctionDecl(decl string, aliases map[string]string) (methodSig, string, bool) {
	matches := cFuncDeclRe.FindStringSubmatch(decl)
	if len(matches) != 4 {
		return methodSig{}, "", false
	}
	retRaw := strings.TrimSpace(matches[1])
	name := strings.TrimSpace(matches[2])
	paramsRaw := strings.TrimSpace(matches[3])
	if name == "" {
		return methodSig{}, "", false
	}
	params := parseCParamTypesCanonical(paramsRaw, aliases)
	returns := []string{}
	ret := normalizeCTypeCanonical(retRaw, aliases)
	if ret != "void" {
		returns = []string{ret}
	}
	return methodSig{params: params, returns: returns}, name, true
}

func moduleFromAPIMacro(m string) string {
	switch m {
	case "FRAMEWORK":
		return "framework"
	case "TOOLKIT":
		return "toolkit"
	case "AGENT_SERVER":
		return "agent_server"
	case "AGENT_CLIENT":
		return "agent_client"
	default:
		return ""
	}
}

func parseCParamTypesCanonical(raw string, aliases map[string]string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "void" {
		return []string{}
	}
	parts := splitCSV(raw)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		pt := normalizeCTypeCanonical(stripCParamName(p), aliases)
		if pt == "" {
			continue
		}
		out = append(out, pt)
	}
	return out
}

func normalizeCTypeCanonical(raw string, aliases map[string]string) string {
	typ := normalizeSpaces(strings.TrimSpace(raw))
	if typ == "" {
		return ""
	}
	typ = strings.ReplaceAll(typ, "MAA_CALL", "")
	typ = normalizeSpaces(typ)
	origBase, origPtrCount := splitCBaseAndPtr(typ)
	if origBase == "MaaBool" {
		if origPtrCount > 0 {
			return "ptr"
		}
		return "bool"
	}
	converged := false
	for i := 0; i < 16; i++ {
		base, ptrCount := splitCBaseAndPtr(typ)
		alias, ok := aliases[base]
		if !ok {
			converged = true
			break
		}
		newTyp := normalizeSpaces(alias + strings.Repeat("*", ptrCount))
		if newTyp == typ {
			converged = true
			break
		}
		typ = newTyp
	}
	if !converged {
		return "<unsupported:c-alias:" + origBase + ">"
	}
	base, ptrCount := splitCBaseAndPtr(typ)
	if base == "char" && ptrCount == 1 {
		return "cstring"
	}
	if ptrCount > 0 {
		return "ptr"
	}
	if strings.HasSuffix(base, "Callback") {
		return "callback:" + base
	}
	switch base {
	case "void":
		return "void"
	case "bool":
		return "bool"
	case "int", "int32_t":
		return "int32"
	case "int64_t":
		return "int64"
	case "uint8_t":
		return "uint8"
	case "uint16_t":
		return "uint16"
	case "uint32_t":
		return "uint32"
	case "uint64_t", "size_t":
		return "uint64"
	case "char":
		return "uint8"
	}
	if strings.HasPrefix(base, "struct ") {
		return "ptr"
	}
	return "named:" + base
}

func stripCParamName(raw string) string {
	s := normalizeSpaces(strings.TrimSpace(raw))
	if s == "" || s == "void" {
		return s
	}
	s = strings.ReplaceAll(s, "[]", "")
	if strings.Contains(s, "(*") {
		return s
	}
	matches := cParamNameRe.FindStringSubmatch(s)
	if len(matches) == 2 {
		return strings.TrimSpace(matches[1])
	}
	return s
}

func splitCBaseAndPtr(raw string) (string, int) {
	s := strings.ReplaceAll(normalizeSpaces(raw), "const ", "")
	s = strings.ReplaceAll(s, "volatile ", "")
	s = normalizeSpaces(strings.TrimSpace(s))
	ptr := strings.Count(s, "*")
	s = strings.ReplaceAll(s, "*", " ")
	s = normalizeSpaces(strings.TrimSpace(s))
	return s, ptr
}

func splitCSV(raw string) []string {
	out := make([]string, 0)
	current := strings.Builder{}
	depth := 0
	for _, ch := range raw {
		switch ch {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				part := strings.TrimSpace(current.String())
				if part != "" {
					out = append(out, part)
				}
				current.Reset()
				continue
			}
		}
		current.WriteRune(ch)
	}
	if tail := strings.TrimSpace(current.String()); tail != "" {
		out = append(out, tail)
	}
	return out
}

func normalizeSpaces(s string) string {
	return spacesRe.ReplaceAllString(strings.TrimSpace(s), " ")
}

// removeCComments preserves token boundaries and source positions. Quoted
// strings and character literals can contain comment markers verbatim.
func removeCComments(s string) string {
	out := []byte(s)
	var quote byte
	for i := 0; i < len(s); i++ {
		if quote != 0 {
			if s[i] == '\\' {
				i++
				continue
			}
			if s[i] == quote {
				quote = 0
			}
			continue
		}
		if s[i] == '"' || s[i] == '\'' {
			quote = s[i]
			continue
		}
		if i+1 >= len(s) || s[i] != '/' {
			continue
		}
		end := i
		switch s[i+1] {
		case '/':
			end = i + 2
			for end < len(s) && s[end] != '\n' {
				end++
			}
		case '*':
			end = i + 2
			for end+1 < len(s) && s[end:end+2] != "*/" {
				end++
			}
			if end+1 < len(s) {
				end += 2
			} else {
				end = len(s)
			}
		default:
			continue
		}
		for j := i; j < end; j++ {
			if out[j] != '\n' && out[j] != '\r' {
				out[j] = ' '
			}
		}
		i = end - 1
	}
	return string(out)
}

func removeCPreprocessorLines(s string) string {
	lines := strings.Split(s, "\n")
	continued := false
	for i, line := range lines {
		if continued || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continued = strings.HasSuffix(strings.TrimSpace(line), "\\")
			lines[i] = strings.Repeat(" ", len(line))
		}
	}
	return strings.Join(lines, "\n")
}

var (
	cTypedefRe        = regexp.MustCompile(`(?s)\btypedef\s+([^;]+?)\s+([A-Za-z_][A-Za-z0-9_]*)\s*;`)
	cAPINameRe        = regexp.MustCompile(`\bMAA_(FRAMEWORK|TOOLKIT|AGENT_SERVER|AGENT_CLIENT)_API\b`)
	cAPIMacroInStmtRe = regexp.MustCompile(`MAA_(FRAMEWORK|TOOLKIT|AGENT_SERVER|AGENT_CLIENT)_API\s+(.+)`)
	cFuncDeclRe       = regexp.MustCompile(`^(.+?)\b(Maa[A-Za-z0-9_]+)\s*\((.*)\)$`)
	cParamNameRe      = regexp.MustCompile(`^(.+?)(?:\s+[A-Za-z_][A-Za-z0-9_]*)$`)
	spacesRe          = regexp.MustCompile(`\s+`)
	cBlockCommentRe   = regexp.MustCompile(`(?s)/\*.*?\*/`)
	cLineCommentRe    = regexp.MustCompile(`(?m)//.*$`)
)
