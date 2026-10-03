package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func checkEventCoverage(repoRoot, headerDir string) ([]issue, error) {
	path := filepath.Join(headerDir, "MaaFramework", "MaaMsg.h")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	messages := map[string]bool{}
	for name, raw := range parseCDefineExprs(removeCComments(string(data))) {
		if !strings.HasPrefix(name, "MaaMsg_") {
			continue
		}
		raw = strings.TrimSpace(raw)
		if strings.HasPrefix(raw, "(") && strings.HasSuffix(raw, ")") {
			raw = strings.TrimSpace(raw[1 : len(raw)-1])
		}
		text, err := strconv.Unquote(raw)
		if err != nil {
			return nil, fmt.Errorf("%s: unsupported event message %s", path, name)
		}
		if messages[text] {
			return nil, fmt.Errorf("%s: duplicate event message %s", path, text)
		}
		messages[text] = true
	}
	fset := token.NewFileSet()
	source := filepath.Join(repoRoot, "event.go")
	f, err := parser.ParseFile(fset, source, nil, 0)
	if err != nil {
		return nil, err
	}
	events := map[string]string{}
	dispatch := map[string]bool{}
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			vs := spec.(*ast.ValueSpec)
			for i, name := range vs.Names {
				if !strings.HasPrefix(name.Name, "Event") || strings.HasPrefix(name.Name, "EventStatus") {
					continue
				}
				if i >= len(vs.Values) {
					return nil, fmt.Errorf("unsupported event constant %s", name.Name)
				}
				expr := vs.Values[i]
				if call, ok := expr.(*ast.CallExpr); ok {
					callee, ok := call.Fun.(*ast.Ident)
					if !ok || callee.Name != "Event" || len(call.Args) != 1 {
						return nil, fmt.Errorf("unsupported event constant %s", name.Name)
					}
					expr = call.Args[0]
				}
				lit, ok := expr.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return nil, fmt.Errorf("unsupported event constant %s", name.Name)
				}
				value, err := strconv.Unquote(lit.Value)
				if err != nil {
					return nil, err
				}
				events[name.Name] = value
			}
		}
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "handleRaw" {
			continue
		}
		eventName := ""
		for _, stmt := range fn.Body.List {
			assign, ok := stmt.(*ast.AssignStmt)
			if !ok || len(assign.Lhs) != 2 || len(assign.Rhs) != 1 {
				continue
			}
			call, ok := assign.Rhs[0].(*ast.CallExpr)
			if !ok || len(call.Args) != 1 {
				continue
			}
			callee, ok := call.Fun.(*ast.Ident)
			if !ok || callee.Name != "parseEvent" {
				continue
			}
			msg, ok := call.Args[0].(*ast.Ident)
			if !ok || msg.Name != "msg" {
				continue
			}
			if name, ok := assign.Lhs[0].(*ast.Ident); ok {
				eventName = name.Name
			}
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			if _, closure := node.(*ast.FuncLit); closure {
				return false
			}
			stmt, ok := node.(*ast.SwitchStmt)
			if !ok {
				return true
			}
			call, ok := stmt.Tag.(*ast.CallExpr)
			if !ok || len(call.Args) != 1 {
				return false
			}
			callee, ok := call.Fun.(*ast.Ident)
			if !ok || callee.Name != "Event" {
				return false
			}
			value, ok := call.Args[0].(*ast.Ident)
			if !ok || eventName == "" || value.Name != eventName {
				return false
			}
			for _, body := range stmt.Body.List {
				clause := body.(*ast.CaseClause)
				for _, label := range clause.List {
					if name, ok := label.(*ast.Ident); ok && events[name.Name] != "" {
						// Empty cases do not implement dispatch.
						if len(clause.Body) > 0 {
							dispatch[name.Name] = true
						}
					}
				}
			}
			return false
		})
	}
	supported := map[string]bool{}
	issues := []issue{}
	for _, name := range sortedPipelineKeys(events) {
		event := events[name]
		for _, status := range []string{"Starting", "Succeeded", "Failed"} {
			msg := event + "." + status
			supported[msg] = true
			if !messages[msg] {
				issues = append(issues, issue{section: sectionEvents, message: "Go event message absent from C: " + msg})
			}
		}
		if !dispatch[name] {
			issues = append(issues, issue{section: sectionEvents, message: "Go event missing dispatch: " + name})
		}
	}
	for _, msg := range sortedPipelineKeys(messages) {
		if !supported[msg] {
			issues = append(issues, issue{section: sectionEvents, message: "C event message missing in Go: " + msg})
		}
	}
	return issues, nil
}
