package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Native module ownership comes from the Library.entries tables consumed by
// initialization, rather than from arbitrary unconsumed Entry literals.
func libraryEntryTables(files []*ast.File) (map[string]map[string]bool, bool, error) {
	out := map[string]map[string]bool{}
	found := false
	handles := map[string]string{"maaFramework": "framework", "maaToolkit": "toolkit", "maaAgentServer": "agent_server", "maaAgentClient": "agent_client"}
	for _, file := range files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				vs := spec.(*ast.ValueSpec)
				for i, name := range vs.Names {
					if name.Name != "libraries" {
						continue
					}
					if found {
						return nil, true, fmt.Errorf("duplicate libraries declaration")
					}
					found = true
					if i >= len(vs.Values) {
						return nil, true, fmt.Errorf("unsupported libraries declaration")
					}
					lit, ok := vs.Values[i].(*ast.CompositeLit)
					if !ok {
						return nil, true, fmt.Errorf("unsupported libraries initializer")
					}
					for _, item := range lit.Elts {
						entry, ok := item.(*ast.CompositeLit)
						if !ok {
							return nil, true, fmt.Errorf("unsupported Library entry")
						}
						module, table := "", ""
						for _, field := range entry.Elts {
							kv, ok := field.(*ast.KeyValueExpr)
							if !ok {
								return nil, true, fmt.Errorf("Library entries must use named fields")
							}
							key, ok := kv.Key.(*ast.Ident)
							if !ok {
								return nil, true, fmt.Errorf("unsupported Library field")
							}
							switch key.Name {
							case "handle":
								module = handles[extractRegisterFuncVarName(kv.Value)]
							case "entries":
								if ident, ok := kv.Value.(*ast.Ident); ok {
									table = ident.Name
								} else {
									return nil, true, fmt.Errorf("Library.entries must reference a named Entry table")
								}
							}
						}
						if module == "" || table == "" {
							return nil, true, fmt.Errorf("Library entry has unknown module handle or missing entries")
						}
						if out[module] != nil {
							return nil, true, fmt.Errorf("duplicate Library module %s", module)
						}
						out[module] = map[string]bool{table: true}
					}
				}
			}
		}
	}
	if found {
		for _, module := range moduleOrder {
			if out[module] == nil {
				return nil, true, fmt.Errorf("libraries is missing module %s", module)
			}
		}
	}
	return out, found, nil
}

func discoverNativeFiles(repoRoot string) (map[string][]string, error) {
	dir := filepath.Join(repoRoot, "internal", "native")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no native source files under %s", dir)
	}
	out := map[string][]string{}
	for _, module := range moduleOrder {
		out[module] = files
	}
	return out, nil
}

func checkNativeLibraryEntries(repoRoot string) error {
	path := filepath.Join(repoRoot, "internal", "native", "native.go")
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return err
	}
	_, found, err := libraryEntryTables([]*ast.File{f})
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if !found {
		return fmt.Errorf("%s: libraries declaration not found", path)
	}
	return nil
}
