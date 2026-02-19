package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strconv"
)

func checkNativeAPICoverage(headerDir string, nativeFiles map[string][]string, blacklist map[string]struct{}) ([]issue, error) {
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

	issues := make([]issue, 0, len(registerIssues))
	issues = append(issues, registerIssues...)
	for _, module := range moduleOrder {
		goSet := goRegistered[module]
		headerSet := sigKeys(headerSigs[module])

		headerOnly := setDiff(headerSet, goSet, blacklist)
		goOnly := setDiff(goSet, headerSet, blacklist)

		for _, fn := range headerOnly {
			issues = append(issues, issue{
				section: sectionNativeAPI,
				message: fmt.Sprintf("[%s] header has function but Go is not registering it: %s", module, fn),
			})
		}
		for _, fn := range goOnly {
			loc := formatGoLocation(goDeclLocs[module][fn], goRegisterLocs[module][fn])
			issues = append(issues, issue{
				section: sectionNativeAPI,
				message: fmt.Sprintf("[%s] Go registers function not found in headers: %s%s", module, fn, loc),
			})
		}
		for fn := range goSet {
			if _, ignored := blacklist[fn]; ignored {
				continue
			}
			goSig, ok1 := goSigs[module][fn]
			cSig, ok2 := headerSigs[module][fn]
			if !ok1 || !ok2 {
				continue
			}
			if unsupportedType, found := findUnsupportedType(goSig.params); found {
				loc := formatGoLocation(goDeclLocs[module][fn], goRegisterLocs[module][fn])
				issues = append(issues, issue{
					section: sectionNativeAPI,
					message: fmt.Sprintf("[%s] %s has unsupported Go param type expression: %s (normalized=%v)%s", module, fn, unsupportedType, goSig.params, loc),
				})
				continue
			}
			if unsupportedType, found := findUnsupportedType(goSig.returns); found {
				loc := formatGoLocation(goDeclLocs[module][fn], goRegisterLocs[module][fn])
				issues = append(issues, issue{
					section: sectionNativeAPI,
					message: fmt.Sprintf("[%s] %s has unsupported Go return type expression: %s (normalized=%v)%s", module, fn, unsupportedType, goSig.returns, loc),
				})
				continue
			}
			if !sameStringSlice(goSig.params, cSig.params) || !sameStringSlice(goSig.returns, cSig.returns) {
				locLine := formatLocationLine(goDeclLocs[module][fn], goRegisterLocs[module][fn])
				issues = append(issues, issue{
					section: sectionNativeAPI,
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

	return issues, nil
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
		for _, file := range files {
			parsedFile, err := parser.ParseFile(fset, file, nil, 0)
			if err != nil {
				return nil, nil, nil, nil, nil, fmt.Errorf("parse %s: %w", file, err)
			}
			varSigs, varDeclLocs := parseGoVarFuncSignaturesWithLoc(parsedFile, fset, file)

			ast.Inspect(parsedFile, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				switch fun := call.Fun.(type) {
				case *ast.Ident:
					if fun.Name != "RegisterLibFunc" {
						return true
					}
				case *ast.SelectorExpr:
					if fun.Sel == nil || fun.Sel.Name != "RegisterLibFunc" {
						return true
					}
				default:
					return true
				}

				if len(call.Args) < 3 {
					return true
				}

				funcVar := extractRegisterFuncVarName(call.Args[0])
				lit, ok := call.Args[2].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}

				name, err := strconv.Unquote(lit.Value)
				if err != nil || name == "" {
					return true
				}
				pos := fset.Position(call.Pos())
				if funcVar != "" && funcVar != name {
					issues = append(issues, issue{
						section: sectionNativeAPI,
						message: fmt.Sprintf("[%s] RegisterLibFunc argument mismatch in %s:%d: var=%s symbol=%s", module, filepath.Clean(file), pos.Line, funcVar, name),
					})
				}
				result[module][name] = struct{}{}
				if sig, ok := varSigs[funcVar]; ok {
					goSigs[module][name] = sig
				}
				if declLoc, ok := varDeclLocs[funcVar]; ok {
					goDeclLocs[module][name] = declLoc
				}
				goRegisterLocs[module][name] = goRegistrationLoc{
					file: filepath.Clean(file),
					line: pos.Line,
				}
				return true
			})
		}
	}

	return result, goSigs, goRegisterLocs, goDeclLocs, issues, nil
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
