package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// callbackABIMatches applies the purego trampoline rules separately from the
// ordinary function rules. C strings are byte pointers in callbacks; a C void
// or MaaBool return needs a uintptr result in the Go trampoline. On the
// supported 64-bit targets, Windows requires exactly one pointer-sized,
// non-floating-point Go result, even when its type matches the C result.
func callbackABIMatches(goSig, cSig methodSig) bool {
	if len(goSig.params) != len(cSig.params) {
		return false
	}
	for i, c := range cSig.params {
		g := goSig.params[i]
		if g == "cstring" || g == "image" || g == "interface" || strings.HasPrefix(g, "callback:") || strings.HasPrefix(g, "<unsupported:") {
			return false
		}
		if c == "cstring" && g == "ptr" {
			continue
		}
		if c != g {
			return false
		}
	}
	if len(goSig.returns) != 1 {
		return false
	}
	switch goSig.returns[0] {
	case "ptr", "int64", "uint64":
		// The supported Windows, Linux and macOS architectures are 64-bit.
	default:
		return false
	}
	if sameStringSlice(goSig.returns, cSig.returns) {
		return true
	}
	return len(goSig.returns) == 1 && goSig.returns[0] == "ptr" && (len(cSig.returns) == 0 || len(cSig.returns) == 1 && (cSig.returns[0] == "bool" || cSig.returns[0] == "cstring"))
}

func parseCNamedCallbacks(headerDir string, aliases map[string]string) (map[string]methodSig, error) {
	out := map[string]methodSig{}
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
		for _, raw := range strings.Split(removeCPreprocessorLines(removeCComments(string(data))), ";") {
			at := strings.LastIndex(raw, "typedef ")
			if at < 0 {
				continue
			}
			decl := strings.TrimSpace(raw[at+len("typedef "):])
			if !strings.Contains(decl, "Callback") || !strings.Contains(decl, "(") {
				continue
			}
			ret, name, params, ok := parseControllerCallbackField(decl)
			if !ok {
				return fmt.Errorf("%s: unsupported callback typedef: %s", path, decl)
			}
			if !strings.HasSuffix(name, "Callback") {
				continue
			}
			sig := methodSig{params: parseCParamTypesCanonical(params, aliases)}
			if ret := normalizeCTypeCanonical(ret, aliases); ret != "void" {
				sig.returns = []string{ret}
			}
			if prior, ok := out[name]; ok && (!sameStringSlice(prior.params, sig.params) || !sameStringSlice(prior.returns, sig.returns)) {
				return fmt.Errorf("conflicting C callback typedef %s", name)
			}
			out[name] = sig
		}
		return nil
	})
	return out, err
}

func checkCallbackABICoverage(repoRoot, headerDir string, nativeFiles map[string][]string) ([]issue, error) {
	aliases, err := parseCTypedefAliases(headerDir)
	if err != nil {
		return nil, err
	}
	cCallbacks, err := parseCNamedCallbacks(headerDir, aliases)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	types := map[string]ast.Expr{}
	for _, module := range moduleOrder {
		for _, path := range nativeFiles[module] {
			f, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return nil, err
			}
			for name, t := range collectGoTypeDefs(f) {
				types[name] = t
			}
		}
	}
	issues := []issue{}
	for _, name := range sortedPipelineKeys(types) {
		fn, ok := types[name].(*ast.FuncType)
		if !ok || !strings.HasSuffix(name, "Callback") {
			continue
		}
		c, ok := cCallbacks[name]
		if !ok {
			issues = append(issues, issue{section: sectionController, message: "Go callback typedef absent from C: " + name})
			continue
		}
		g := methodSig{params: parseGoFieldTypesCanonical(fn.Params, types), returns: parseGoFieldTypesCanonical(fn.Results, types)}
		if !callbackABIMatches(g, c) {
			issues = append(issues, issue{section: sectionController, message: fmt.Sprintf("callback ABI mismatch for %s: go=%v -> %v c=%v -> %v", name, g.params, g.returns, c.params, c.returns)})
		}
		// The named callback declaration and the actual root trampoline must agree.
		path := map[string]string{"MaaEventCallback": "event.go", "MaaCustomActionCallback": "custom_action.go", "MaaCustomRecognitionCallback": "custom_recognition.go"}[name]
		if path != "" {
			source, err := parser.ParseFile(fset, filepath.Join(repoRoot, path), nil, 0)
			if err != nil {
				return nil, err
			}
			rootTypes := collectGoTypeDefs(source)
			found := false
			for _, decl := range source.Decls {
				if agent, ok := decl.(*ast.FuncDecl); ok && agent.Name.Name == "_"+name+"Agent" {
					found = true
					sig := methodSig{params: parseGoFieldTypesCanonical(agent.Type.Params, rootTypes), returns: parseGoFieldTypesCanonical(agent.Type.Results, rootTypes)}
					if !callbackABIMatches(sig, c) {
						issues = append(issues, issue{section: sectionController, message: fmt.Sprintf("callback trampoline ABI mismatch for %s at %s", name, fset.Position(agent.Pos()))})
					}
				}
			}
			if !found {
				issues = append(issues, issue{section: sectionController, message: "missing callback trampoline for " + name})
			}
		}
	}
	usedCallbacks, err := exportedCallbackTypes(headerDir, cCallbacks)
	if err != nil {
		return nil, err
	}
	for _, name := range sortedPipelineKeys(usedCallbacks) {
		if _, ok := types[name].(*ast.FuncType); !ok {
			issues = append(issues, issue{section: sectionController, message: "C callback typedef missing in Go: " + name})
		}
	}
	layoutIssues, err := checkControllerCallbackLayout(filepath.Join(headerDir, controllerHeaderRel), filepath.Join(repoRoot, customControllerRel), aliases)
	return append(issues, layoutIssues...), err
}

func readControllerCallbackFields(path string) ([]controllerCallbackField, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	text := removeCComments(string(data))
	start := strings.Index(text, "struct MaaCustomControllerCallbacks")
	if start < 0 {
		return nil, fmt.Errorf("%s: missing callbacks struct", path)
	}
	open := strings.Index(text[start:], "{")
	if open < 0 {
		return nil, fmt.Errorf("missing callbacks open brace")
	}
	open += start
	close := strings.Index(text[open:], "};")
	if close < 0 {
		return nil, fmt.Errorf("missing callbacks close brace")
	}
	return parseControllerCallbackFields(text[open+1 : open+close])
}

func checkControllerCallbackLayout(headerPath, sourcePath string, aliases map[string]string) ([]issue, error) {
	callbacks, err := readControllerCallbackFields(headerPath)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, sourcePath, nil, 0)
	if err != nil {
		return nil, err
	}
	types := collectGoTypeDefs(f)
	st, ok := types["MaaCustomControllerCallbacks"].(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("%s: MaaCustomControllerCallbacks struct not found", sourcePath)
	}
	expected := make([]string, len(callbacks))
	for i, cb := range callbacks {
		expected[i] = callbackNameToGoMethod(cb.name)
	}
	actual := []string{}
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 {
			return nil, fmt.Errorf("unsupported embedded callbacks layout at %s", fset.Position(field.Pos()))
		}
		if normalizeGoTypeExprCanonical(field.Type, types, map[string]struct{}{}) != "ptr" {
			return nil, fmt.Errorf("callbacks field must have pointer-sized type at %s", fset.Position(field.Pos()))
		}
		for _, name := range field.Names {
			actual = append(actual, name.Name)
		}
	}
	issues := []issue{}
	if !sameStringSlice(actual, expected) {
		issues = append(issues, issue{section: sectionController, message: fmt.Sprintf("callback struct layout mismatch: go=%v c=%v", actual, expected)})
	}
	funcs := map[string]*ast.FuncDecl{}
	bindings := map[string]string{}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		funcs[fn.Name.Name] = fn
		if fn.Name.Name != "init" {
			continue
		}
		var problem error
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			if _, closure := node.(*ast.FuncLit); closure {
				return false
			}
			assignment, ok := node.(*ast.AssignStmt)
			if !ok || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
				return true
			}
			field, ok := assignment.Lhs[0].(*ast.SelectorExpr)
			if !ok {
				return true
			}
			owner, ok := field.X.(*ast.Ident)
			if !ok || owner.Name != "customControllerCallbacksHandle" {
				return true
			}
			call, ok := assignment.Rhs[0].(*ast.CallExpr)
			if !ok || len(call.Args) != 1 {
				problem = fmt.Errorf("unsupported callback binding at %s", fset.Position(assignment.Pos()))
				return false
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || selector.Sel.Name != "NewCallback" {
				problem = fmt.Errorf("unsupported callback binding at %s", fset.Position(assignment.Pos()))
				return false
			}
			pkg, ok := selector.X.(*ast.Ident)
			if !ok || pkg.Name != "purego" {
				problem = fmt.Errorf("unsupported callback binding package")
				return false
			}
			target, ok := call.Args[0].(*ast.Ident)
			if !ok {
				problem = fmt.Errorf("unsupported callback target")
				return false
			}
			if bindings[field.Sel.Name] != "" {
				problem = fmt.Errorf("duplicate callback binding %s", field.Sel.Name)
				return false
			}
			bindings[field.Sel.Name] = target.Name
			return true
		})
		if problem != nil {
			return nil, problem
		}
	}
	used := map[string]bool{}
	for _, cb := range callbacks {
		name := callbackNameToGoMethod(cb.name)
		target := bindings[name]
		used[name] = true
		expectedTarget := "_" + name + "Agent"
		switch name {
		case "ClickKey", "InputText", "KeyDown", "KeyUp":
			expectedTarget = "_" + name
		}
		if target != "" && target != expectedTarget {
			issues = append(issues, issue{section: sectionController, message: fmt.Sprintf("callback binding target mismatch for %s: got %s, want %s", name, target, expectedTarget)})
		}
		fn := funcs[target]
		if fn == nil {
			issues = append(issues, issue{section: sectionController, message: "missing callback binding or target for " + name})
			continue
		}
		c := methodSig{params: parseCParamTypesCanonical(cb.paramsRaw, aliases)}
		if ret := normalizeCTypeCanonical(cb.retType, aliases); ret != "void" {
			c.returns = []string{ret}
		}
		g := methodSig{params: parseGoFieldTypesCanonical(fn.Type.Params, types), returns: parseGoFieldTypesCanonical(fn.Type.Results, types)}
		if !callbackABIMatches(g, c) {
			issues = append(issues, issue{section: sectionController, message: fmt.Sprintf("callback binding ABI mismatch for %s (%s) at %s: go=%v -> %v c=%v -> %v", name, target, fset.Position(fn.Pos()), g.params, g.returns, c.params, c.returns)})
		}
	}
	for _, name := range sortedPipelineKeys(bindings) {
		if !used[name] {
			issues = append(issues, issue{section: sectionController, message: "Go callback binding absent from C: " + name})
		}
	}
	return issues, nil
}

// Only callback typedefs used by exported functions need a Go binding. Retired
// typedefs left in the headers do not expand the supported public API.
func exportedCallbackTypes(headerDir string, callbacks map[string]methodSig) (map[string]bool, error) {
	used := map[string]bool{}
	err := filepath.WalkDir(headerDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".h" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, stmt := range strings.Split(removeCPreprocessorLines(removeCComments(string(data))), ";") {
			match := cAPIMacroInStmtRe.FindStringSubmatch(stmt)
			if match == nil {
				continue
			}
			for _, word := range strings.FieldsFunc(match[2], func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' }) {
				if _, ok := callbacks[word]; ok {
					used[word] = true
				}
			}
		}
		return nil
	})
	return used, err
}
