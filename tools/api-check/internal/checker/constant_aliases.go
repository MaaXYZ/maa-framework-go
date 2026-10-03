package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// evaluatePublicConstants includes the native declarations referenced by
// public numeric aliases. It uses source declarations, so changing either a
// public alias or its native value remains visible to the inventory check.
func evaluatePublicConstants(repoRoot, rel string) (*goConstEvaluation, error) {
	var source strings.Builder
	source.WriteString("package constants\n")
	for _, file := range []string{"internal/native/framework.go", "internal/native/toolkit.go", rel} {
		path := filepath.Join(repoRoot, file)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		fset := token.NewFileSet()
		parsed, err := parser.ParseFile(fset, path, data, 0)
		if err != nil {
			return nil, err
		}
		for _, decl := range parsed.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST && gen.Tok != token.TYPE {
				continue
			}
			start, end := fset.Position(gen.Pos()).Offset, fset.Position(gen.End()).Offset
			text := string(data[start:end])
			if file == rel {
				type replacement struct {
					start, end int
					name       string
				}
				var replacements []replacement
				ast.Inspect(gen, func(node ast.Node) bool {
					selector, ok := node.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					pkg, ok := selector.X.(*ast.Ident)
					if ok && pkg.Name == "native" {
						replacements = append(replacements, replacement{fset.Position(selector.Pos()).Offset - start, fset.Position(selector.End()).Offset - start, selector.Sel.Name})
					}
					return true
				})
				sort.Slice(replacements, func(i, j int) bool { return replacements[i].start > replacements[j].start })
				for _, r := range replacements {
					text = text[:r.start] + r.name + text[r.end:]
				}
			}
			source.WriteString(text)
			source.WriteByte('\n')
		}
	}
	evaluation, err := evaluateGoConstSource(rel, []byte(source.String()))
	if err != nil {
		return nil, fmt.Errorf("evaluate public aliases in %s: %w", rel, err)
	}
	return evaluation, nil
}
