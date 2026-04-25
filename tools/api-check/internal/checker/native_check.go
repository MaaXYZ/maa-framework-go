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

			entryRegistrations, entryIssues := parseGoEntryRegistrations(parsedFile, fset, file, module)
			issues = append(issues, entryIssues...)
			for _, registration := range entryRegistrations {
				result[module][registration.name] = struct{}{}
				if sig, ok := varSigs[registration.funcVar]; ok {
					goSigs[module][registration.name] = sig
				}
				if declLoc, ok := varDeclLocs[registration.funcVar]; ok {
					goDeclLocs[module][registration.name] = declLoc
				}
				goRegisterLocs[module][registration.name] = goRegistrationLoc{
					file: registration.file,
					line: registration.line,
				}
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

			for _, value := range valueSpec.Values {
				entriesLit, ok := value.(*ast.CompositeLit)
				if !ok || !isEntrySliceLiteral(entriesLit.Type) {
					continue
				}

				for _, entryExpr := range entriesLit.Elts {
					registration, ok := parseGoEntryRegistration(entryExpr, fset, sourceFile)
					if !ok {
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

	funcVar := extractRegisterFuncVarName(entryLit.Elts[0])
	nameLit, ok := entryLit.Elts[1].(*ast.BasicLit)
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
