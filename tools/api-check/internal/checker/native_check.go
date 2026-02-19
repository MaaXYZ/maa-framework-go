package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func checkNativeAPICoverage(headerDir string, blacklist map[string]struct{}) ([]issue, error) {
	goRegistered, registerIssues, err := parseGoRegistrations()
	if err != nil {
		return nil, fmt.Errorf("parse Go registrations: %w", err)
	}

	headerFuncs, err := parseHeaderFunctions(headerDir)
	if err != nil {
		return nil, fmt.Errorf("parse C headers: %w", err)
	}

	issues := make([]issue, 0, len(registerIssues))
	issues = append(issues, registerIssues...)
	for _, module := range moduleOrder {
		goSet := goRegistered[module]
		headerSet := headerFuncs[module]

		headerOnly := setDiff(headerSet, goSet, blacklist)
		goOnly := setDiff(goSet, headerSet, blacklist)

		for _, fn := range headerOnly {
			issues = append(issues, issue{
				section: sectionNativeAPI,
				message: fmt.Sprintf("[%s] header has function but Go is not registering it: %s", module, fn),
			})
		}
		for _, fn := range goOnly {
			issues = append(issues, issue{
				section: sectionNativeAPI,
				message: fmt.Sprintf("[%s] Go registers function not found in headers: %s", module, fn),
			})
		}
	}

	return issues, nil
}

func parseGoRegistrations() (map[string]map[string]struct{}, []issue, error) {
	result := map[string]map[string]struct{}{}
	issues := make([]issue, 0)
	fset := token.NewFileSet()

	for module, files := range nativeFilesByModule {
		if _, ok := result[module]; !ok {
			result[module] = map[string]struct{}{}
		}
		for _, file := range files {
			parsedFile, err := parser.ParseFile(fset, file, nil, 0)
			if err != nil {
				return nil, nil, fmt.Errorf("parse %s: %w", file, err)
			}

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
				if funcVar != "" && funcVar != name {
					pos := fset.Position(call.Pos())
					issues = append(issues, issue{
						section: sectionNativeAPI,
						message: fmt.Sprintf("[%s] RegisterLibFunc argument mismatch in %s:%d: var=%s symbol=%s", module, filepath.Clean(file), pos.Line, funcVar, name),
					})
				}
				result[module][name] = struct{}{}
				return true
			})
		}
	}

	return result, issues, nil
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

func parseHeaderFunctions(headerDir string) (map[string]map[string]struct{}, error) {
	result := map[string]map[string]struct{}{
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
		content := string(data)
		chunks := strings.Split(content, ";")
		for _, chunk := range chunks {
			stmt := strings.TrimSpace(chunk)
			if stmt == "" || !strings.Contains(stmt, "MAA_") {
				continue
			}
			if !strings.Contains(stmt, "_API") {
				continue
			}
			if strings.Contains(stmt, "MAA_DEPRECATED") {
				continue
			}

			module := moduleFromHeaderStmt(stmt)
			if module == "" {
				continue
			}
			name := extractFuncNameFromDecl(stmt)
			if name == "" {
				continue
			}
			result[module][name] = struct{}{}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func moduleFromHeaderStmt(stmt string) string {
	switch {
	case strings.Contains(stmt, "MAA_FRAMEWORK_API"):
		return "framework"
	case strings.Contains(stmt, "MAA_TOOLKIT_API"):
		return "toolkit"
	case strings.Contains(stmt, "MAA_AGENT_SERVER_API"):
		return "agent_server"
	case strings.Contains(stmt, "MAA_AGENT_CLIENT_API"):
		return "agent_client"
	default:
		return ""
	}
}

func extractFuncNameFromDecl(stmt string) string {
	re := regexp.MustCompile(`\b(Maa[A-Za-z0-9_]+)\s*\(`)
	matches := re.FindAllStringSubmatch(stmt, -1)
	if len(matches) == 0 {
		return ""
	}
	last := matches[len(matches)-1]
	if len(last) < 2 {
		return ""
	}
	return strings.TrimSpace(last[1])
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
