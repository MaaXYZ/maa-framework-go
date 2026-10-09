package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type nativeInventory struct {
	registered map[string]int
	exported   map[string]int
}

func checkNativeAPICoverage(headerDir string, nativeFiles map[string][]string, blacklist map[string]struct{}, inventories ...*nativeInventory) ([]issue, error) {
	goRegistered, goSigs, goRegisterLocs, goDeclLocs, registerIssues, err := parseGoRegistrations(nativeFiles)
	if err != nil {
		return nil, fmt.Errorf("parse Go registrations: %w", err)
	}

	aliases, err := parseCTypedefAliases(headerDir)
	if err != nil {
		return nil, fmt.Errorf("parse C typedef aliases: %w", err)
	}
	headerSigs, err := parseHeaderFunctionSignatures(headerDir, aliases)
	if err != nil {
		return nil, fmt.Errorf("parse C headers: %w", err)
	}

	for _, inventory := range inventories {
		inventory.registered = map[string]int{}
		inventory.exported = map[string]int{}
		for _, module := range moduleOrder {
			inventory.registered[module] = len(goRegistered[module])
			inventory.exported[module] = len(headerSigs[module])
		}
	}
	issues := make([]issue, 0, len(registerIssues))
	issues = append(issues, registerIssues...)
	for _, module := range moduleOrder {
		goSet := goRegistered[module]
		headerSet := sigKeys(headerSigs[module])

		headerOnly := setDiff(headerSet, goSet, nil)
		goOnly := setDiff(goSet, headerSet, nil)

		for _, fn := range headerOnly {
			issues = append(issues, issue{
				section: sectionNativeAPI,
				symbol:  fn,
				message: fmt.Sprintf("[%s] header has function but Go is not registering it: %s", module, fn),
			})
		}
		for _, fn := range goOnly {
			loc := formatGoLocation(goDeclLocs[module][fn], goRegisterLocs[module][fn])
			issues = append(issues, issue{
				section: sectionNativeAPI,
				symbol:  fn,
				message: fmt.Sprintf("[%s] Go registers function not found in headers: %s%s", module, fn, loc),
			})
		}
		for fn := range goSet {
			goSig, ok1 := goSigs[module][fn]
			cSig, ok2 := headerSigs[module][fn]
			if !ok1 || !ok2 {
				continue
			}
			if unsupportedType, found := findUnsupportedType(goSig.params); found {
				loc := formatGoLocation(goDeclLocs[module][fn], goRegisterLocs[module][fn])
				issues = append(issues, issue{
					section: sectionNativeAPI,
					symbol:  fn,
					message: fmt.Sprintf("[%s] %s has unsupported Go param type expression: %s (normalized=%v)%s", module, fn, unsupportedType, goSig.params, loc),
				})
				continue
			}
			if unsupportedType, found := findUnsupportedType(goSig.returns); found {
				loc := formatGoLocation(goDeclLocs[module][fn], goRegisterLocs[module][fn])
				issues = append(issues, issue{
					section: sectionNativeAPI,
					symbol:  fn,
					message: fmt.Sprintf("[%s] %s has unsupported Go return type expression: %s (normalized=%v)%s", module, fn, unsupportedType, goSig.returns, loc),
				})
				continue
			}
			if !sameStringSlice(goSig.params, cSig.params) || !nativeReturnTypesMatch(goSig.returns, cSig.returns) {
				locLine := formatLocationLine(goDeclLocs[module][fn], goRegisterLocs[module][fn])
				issues = append(issues, issue{
					section: sectionNativeAPI,
					symbol:  fn,
					message: fmt.Sprintf(
						"[%s] signature mismatch for %s\n"+
							"go params: %v\n"+
							"go returns: %v\n"+
							"c  params: %v\n"+
							"c  returns: %v%s",
						module,
						fn,
						goSig.params,
						goSig.returns,
						cSig.params,
						cSig.returns,
						locLine,
					),
				})
			}
		}
	}

	filtered := make([]issue, 0, len(issues))
	used := map[string]bool{}
	for _, it := range issues {
		if _, ignored := blacklist[it.symbol]; ignored && it.symbol != "" {
			used[it.symbol] = true
			continue
		}
		filtered = append(filtered, it)
	}
	for _, name := range sortedPipelineKeys(blacklist) {
		if !used[name] {
			filtered = append(filtered, issue{section: sectionNativeAPI, message: name + ": stale native exclusion (no current difference)"})
		}
	}
	return filtered, nil
}

// nativeReturnTypesMatch allows C char* returns to stay as raw Go pointers
// when the caller needs to read a byte count instead of a NUL-terminated string.
func nativeReturnTypesMatch(goTypes, cTypes []string) bool {
	if len(goTypes) != len(cTypes) {
		return false
	}
	for i, goType := range goTypes {
		if goType != cTypes[i] && !(goType == "ptr" && cTypes[i] == "cstring") {
			return false
		}
	}
	return true
}

func parseGoRegistrations(nativeFiles map[string][]string) (map[string]map[string]struct{}, map[string]map[string]methodSig, map[string]map[string]goRegistrationLoc, map[string]map[string]goVarDeclLoc, []issue, error) {
	result := map[string]map[string]struct{}{}
	goSigs := map[string]map[string]methodSig{}
	goRegisterLocs := map[string]map[string]goRegistrationLoc{}
	goDeclLocs := map[string]map[string]goVarDeclLoc{}
	issues := make([]issue, 0)
	fset := token.NewFileSet()

	for module, files := range nativeFiles {
		if _, ok := result[module]; !ok {
			result[module] = map[string]struct{}{}
		}
		if _, ok := goSigs[module]; !ok {
			goSigs[module] = map[string]methodSig{}
		}
		if _, ok := goRegisterLocs[module]; !ok {
			goRegisterLocs[module] = map[string]goRegistrationLoc{}
		}
		if _, ok := goDeclLocs[module]; !ok {
			goDeclLocs[module] = map[string]goVarDeclLoc{}
		}
		parsedFiles := make([]*ast.File, 0, len(files))
		for _, file := range files {
			parsedFile, err := parser.ParseFile(fset, file, nil, 0)
			if err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("parse %s: %w", file, err)
			}
			parsedFiles = append(parsedFiles, parsedFile)
		}
		ownership, live, err := libraryEntryTables(parsedFiles)
		if err != nil {
			return nil, nil, nil, nil, nil, err
		}
		var allowed map[string]bool
		if live {
			allowed = ownership[module]
		}
		if live {
			for table := range allowed {
				found := false
				for _, parsed := range parsedFiles {
					for _, decl := range parsed.Decls {
						if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
							for _, spec := range gen.Specs {
								vs := spec.(*ast.ValueSpec)
								for _, name := range vs.Names {
									if name.Name == table && len(vs.Values) > 0 {
										found = true
									}
								}
							}
						}
					}
				}
				if !found {
					return nil, nil, nil, nil, nil, fmt.Errorf("%s: missing Entry table %s", module, table)
				}
			}
		}
		varSigs, varDeclLocs := collectGoFuncSignatures(parsedFiles, fset, "")
		for i, parsedFile := range parsedFiles {
			file := files[i]
			entryRegistrations, entryIssues := parseGoEntryRegistrationsInTables(parsedFile, fset, file, module, allowed)
			for _, it := range entryIssues {
				if strings.Contains(it.message, "unsupported Entry") {
					return nil, nil, nil, nil, nil, fmt.Errorf("%s", it.message)
				}
			}
			issues = append(issues, entryIssues...)
			for _, registration := range entryRegistrations {
				if _, duplicate := result[module][registration.name]; duplicate {
					return nil, nil, nil, nil, nil, fmt.Errorf("%s:%d: duplicate registration for %s", file, registration.line, registration.name)
				}
				sig, ok := varSigs[registration.funcVar]
				if !ok {
					return nil, nil, nil, nil, nil, fmt.Errorf("%s:%d: cannot resolve Go function signature for registered target %q", file, registration.line, registration.funcVar)
				}
				result[module][registration.name] = struct{}{}
				goSigs[module][registration.name] = sig
				goDeclLocs[module][registration.name] = varDeclLocs[registration.funcVar]
				goRegisterLocs[module][registration.name] = goRegistrationLoc{file: registration.file, line: registration.line}
			}
		}
	}

	return result, goSigs, goRegisterLocs, goDeclLocs, issues, nil
}

type goEntryRegistration struct {
	funcVar string
	name    string
	file    string
	line    int
}

func parseGoEntryRegistrations(parsedFile *ast.File, fset *token.FileSet, sourceFile, module string) ([]goEntryRegistration, []issue) {
	return parseGoEntryRegistrationsInTables(parsedFile, fset, sourceFile, module, nil)
}

func parseGoEntryRegistrationsInTables(parsedFile *ast.File, fset *token.FileSet, sourceFile, module string, allowed map[string]bool) ([]goEntryRegistration, []issue) {
	registrations := make([]goEntryRegistration, 0)
	issues := make([]issue, 0)

	for _, decl := range parsedFile.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}

		for _, spec := range gen.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, value := range valueSpec.Values {
				if allowed != nil && (i >= len(valueSpec.Names) || !allowed[valueSpec.Names[i].Name]) {
					continue
				}
				entriesLit, ok := value.(*ast.CompositeLit)
				if !ok || !isEntrySliceLiteral(entriesLit.Type) {
					if allowed != nil {
						issues = append(issues, issue{section: sectionNativeAPI, message: fmt.Sprintf("[%s] unsupported Entry table at %s", module, fset.Position(value.Pos()))})
					}
					continue
				}

				for _, entryExpr := range entriesLit.Elts {
					registration, ok := parseGoEntryRegistration(entryExpr, fset, sourceFile)
					if !ok {
						issues = append(issues, issue{section: sectionNativeAPI, message: fmt.Sprintf("[%s] unsupported Entry in %s", module, fset.Position(entryExpr.Pos()))})
						continue
					}

					if registration.funcVar != registration.name {
						issues = append(issues, issue{
							section: sectionNativeAPI,
							message: fmt.Sprintf("[%s] Entry mismatch in %s:%d: var=%s symbol=%s", module, registration.file, registration.line, registration.funcVar, registration.name),
						})
					}

					registrations = append(registrations, registration)
				}
			}
		}
	}

	return registrations, issues
}

func isEntrySliceLiteral(expr ast.Expr) bool {
	arrayType, ok := expr.(*ast.ArrayType)
	if !ok {
		return false
	}

	ident, ok := arrayType.Elt.(*ast.Ident)
	return ok && ident.Name == "Entry"
}

func parseGoEntryRegistration(entryExpr ast.Expr, fset *token.FileSet, sourceFile string) (goEntryRegistration, bool) {
	entryLit, ok := entryExpr.(*ast.CompositeLit)
	if !ok || len(entryLit.Elts) < 2 {
		return goEntryRegistration{}, false
	}

	target, symbol := entryLit.Elts[0], entryLit.Elts[1]
	if _, keyed := target.(*ast.KeyValueExpr); keyed {
		target, symbol = nil, nil
		for _, elt := range entryLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				return goEntryRegistration{}, false
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				return goEntryRegistration{}, false
			}
			switch key.Name {
			case "ptrToFunc":
				target = kv.Value
			case "name":
				symbol = kv.Value
			default:
				return goEntryRegistration{}, false
			}
		}
	}
	funcVar := extractRegisterFuncVarName(target)
	if funcVar == "" {
		return goEntryRegistration{}, false
	}
	nameLit, ok := symbol.(*ast.BasicLit)
	if !ok || nameLit.Kind != token.STRING {
		return goEntryRegistration{}, false
	}

	name, err := strconv.Unquote(nameLit.Value)
	if err != nil || name == "" {
		return goEntryRegistration{}, false
	}

	pos := fset.Position(entryLit.Pos())
	file := filepath.Clean(sourceFile)
	if pos.Filename != "" {
		file = filepath.Clean(pos.Filename)
	}

	return goEntryRegistration{
		funcVar: funcVar,
		name:    name,
		file:    file,
		line:    pos.Line,
	}, true
}

func extractRegisterFuncVarName(arg ast.Expr) string {
	u, ok := arg.(*ast.UnaryExpr)
	if !ok || u.Op != token.AND {
		return ""
	}
	switch x := u.X.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		if x.Sel == nil {
			return ""
		}
		return x.Sel.Name
	default:
		return ""
	}
}

func setDiff(left, right map[string]struct{}, blacklist map[string]struct{}) []string {
	out := make([]string, 0)
	for k := range left {
		if _, ignored := blacklist[k]; ignored {
			continue
		}
		if _, ok := right[k]; !ok {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func sigKeys(m map[string]methodSig) map[string]struct{} {
	out := make(map[string]struct{}, len(m))
	for k := range m {
		out[k] = struct{}{}
	}
	return out
}

type goRegistrationLoc struct {
	file string
	line int
}

func formatRegisteredAt(loc goRegistrationLoc) string {
	if loc.file == "" || loc.line <= 0 {
		return ""
	}
	return fmt.Sprintf("registered at %s:%d", loc.file, loc.line)
}

func formatDeclAt(loc goVarDeclLoc) string {
	if loc.file == "" || loc.line <= 0 {
		return ""
	}
	return fmt.Sprintf("decl at %s:%d", loc.file, loc.line)
}

func formatGoLocation(decl goVarDeclLoc, reg goRegistrationLoc) string {
	plain := formatGoLocationPlain(decl, reg)
	if plain == "" {
		return ""
	}
	return fmt.Sprintf(" (%s)", plain)
}

func formatGoLocationPlain(decl goVarDeclLoc, reg goRegistrationLoc) string {
	declMsg := formatDeclAt(decl)
	regMsg := formatRegisteredAt(reg)
	if declMsg == "" && regMsg == "" {
		return ""
	}
	if declMsg != "" && regMsg != "" {
		return fmt.Sprintf("%s, %s", declMsg, regMsg)
	}
	if declMsg != "" {
		return declMsg
	}
	return regMsg
}

func formatLocationLine(decl goVarDeclLoc, reg goRegistrationLoc) string {
	plain := formatGoLocationPlain(decl, reg)
	if plain == "" {
		return ""
	}
	return "\nlocation: " + plain
}
