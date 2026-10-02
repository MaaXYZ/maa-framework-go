package checker

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
)

type pipelineGo struct {
	fset    *token.FileSet
	types   map[string]ast.Expr
	aliases map[string]bool
	funcs   map[string]*ast.FuncDecl
	methods map[string]map[string]*ast.FuncDecl
	enums   map[string]map[string]string
}

func readPipelineGo(repoRoot string) (*pipelineGo, error) {
	g := &pipelineGo{fset: token.NewFileSet(), types: map[string]ast.Expr{}, aliases: map[string]bool{}, funcs: map[string]*ast.FuncDecl{}, methods: map[string]map[string]*ast.FuncDecl{}, enums: map[string]map[string]string{"action": {}, "recognition": {}}}
	entries, err := os.ReadDir(repoRoot)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		f, err := parser.ParseFile(g.fset, filepath.Join(repoRoot, entry.Name()), nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parse pipeline Go source: %w", err)
		}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				var constantType string
				var previous []ast.Expr
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						g.types[s.Name.Name] = s.Type
						g.aliases[s.Name.Name] = s.Assign.IsValid()
					case *ast.ValueSpec:
						if d.Tok != token.CONST {
							continue
						}
						if s.Type != nil {
							constantType = pipelineTypeName(s.Type)
						} else if len(s.Values) != 0 {
							constantType = ""
						}
						values := s.Values
						if len(values) == 0 {
							values = previous
						} else {
							previous = values
						}
						kind := ""
						if constantType == "ActionType" {
							kind = "action"
						}
						if constantType == "RecognitionType" {
							kind = "recognition"
						}
						if kind == "" {
							continue
						}
						if len(values) != len(s.Names) {
							return nil, fmt.Errorf("%s: unsupported %s constant declaration", g.fset.Position(s.Pos()), constantType)
						}
						for i, name := range s.Names {
							lit, ok := values[i].(*ast.BasicLit)
							if !ok || lit.Kind != token.STRING {
								return nil, fmt.Errorf("%s: %s constant must be a string literal", g.fset.Position(name.Pos()), constantType)
							}
							value, err := strconv.Unquote(lit.Value)
							if err != nil {
								return nil, err
							}
							g.enums[kind][name.Name] = value
						}
					}
				}
			case *ast.FuncDecl:
				if d.Recv == nil {
					g.funcs[d.Name.Name] = d
					continue
				}
				receiver := pipelineTypeName(d.Recv.List[0].Type)
				if g.methods[receiver] == nil {
					g.methods[receiver] = map[string]*ast.FuncDecl{}
				}
				g.methods[receiver][d.Name.Name] = d
			}
		}
	}
	return g, nil
}

func pipelineTypeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return pipelineTypeName(e.X)
	case *ast.ParenExpr:
		return pipelineTypeName(e.X)
	}
	return ""
}

// decoderParams follows only the UnmarshalJSON method and helpers it calls.
// A default raw fallback never establishes coverage for a typed enum case.
func (g *pipelineGo) decoderParams(kind string) (map[string]string, error) {
	typeName := "Action"
	if kind == "recognition" {
		typeName = "Recognition"
	}
	start := g.methods[typeName]["UnmarshalJSON"]
	if start == nil {
		return nil, fmt.Errorf("Go %s: missing UnmarshalJSON decoder", typeName)
	}
	out := map[string]string{}
	visited := map[*ast.FuncDecl]bool{}
	var visit func(*ast.FuncDecl) error
	visit = func(fn *ast.FuncDecl) error {
		if visited[fn] {
			return nil
		}
		visited[fn] = true
		paramVariables := map[string]bool{}
		receiverName := ""
		if fn.Recv != nil && pipelineTypeName(fn.Recv.List[0].Type) == typeName && len(fn.Recv.List[0].Names) != 0 {
			receiverName = fn.Recv.List[0].Names[0].Name
		}
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}
			if spec, ok := node.(*ast.ValueSpec); ok && pipelineTypeName(spec.Type) == typeName+"Param" {
				for _, name := range spec.Names {
					paramVariables[name.Name] = true
				}
			}
			return true
		})
		paramTarget := func(expr ast.Expr) bool {
			if name, ok := expr.(*ast.Ident); ok {
				return paramVariables[name.Name]
			}
			if field, ok := expr.(*ast.SelectorExpr); ok && field.Sel.Name == "Param" {
				name, ok := field.X.(*ast.Ident)
				return ok && receiverName != "" && name.Name == receiverName
			}
			return false
		}
		var problem error
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			if problem != nil {
				return false
			}
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}
			if call, ok := node.(*ast.CallExpr); ok {
				if name, ok := call.Fun.(*ast.Ident); ok && g.funcs[name.Name] != nil {
					problem = visit(g.funcs[name.Name])
				}
			}
			clause, ok := node.(*ast.CaseClause)
			if !ok {
				return true
			}
			var values []string
			for _, label := range clause.List {
				if name, ok := label.(*ast.Ident); ok {
					if value, exists := g.enums[kind][name.Name]; exists {
						values = append(values, value)
					}
				}
			}
			if len(values) == 0 {
				return true
			}
			candidates := map[string]bool{}
			addLiteral := func(expr ast.Expr) {
				unary, ok := expr.(*ast.UnaryExpr)
				if !ok || unary.Op != token.AND {
					return
				}
				literal, ok := unary.X.(*ast.CompositeLit)
				if !ok {
					return
				}
				name := pipelineTypeName(literal.Type)
				if name != "" && g.types[name] != nil {
					candidates[name] = true
				}
			}
			for _, stmt := range clause.Body {
				ast.Inspect(stmt, func(n ast.Node) bool {
					switch node := n.(type) {
					case *ast.FuncLit:
						return false
					case *ast.ReturnStmt:
						for _, result := range node.Results {
							addLiteral(result)
						}
					case *ast.AssignStmt:
						for i, lhs := range node.Lhs {
							if i < len(node.Rhs) && paramTarget(lhs) {
								addLiteral(node.Rhs[i])
							}
						}
					}
					return true
				})
			}
			if len(candidates) != 1 {
				problem = fmt.Errorf("%s: unsupported typed %s decoder case (expected one parameter literal assigned to Param or returned)", g.fset.Position(clause.Pos()), kind)
				return false
			}
			var param string
			for name := range candidates {
				param = name
			}
			for _, value := range values {
				if prior := out[value]; prior != "" && prior != param {
					problem = fmt.Errorf("Go %s.%s: conflicting decoder parameters %s and %s", kind, value, prior, param)
					return false
				}
				out[value] = param
			}
			return true
		})
		return problem
	}
	if err := visit(start); err != nil {
		return nil, err
	}
	return out, nil
}

type pipelineFieldSet map[string]string

type pipelineLocalType struct {
	expr  ast.Expr
	alias bool
}

// A new defined type (the usual local NoMethod DTO) strips methods; an alias
// preserves them. Promoted codecs need method-set analysis outside this scope.
func (g *pipelineGo) hasCodec(name string, local map[string]pipelineLocalType) bool {
	seen := map[string]bool{}
	for name != "" && !seen[name] {
		seen[name] = true
		if info, ok := local[name]; ok {
			if !info.alias {
				return false
			}
			name = pipelineTypeName(info.expr)
			continue
		}
		if g.methods[name]["MarshalJSON"] != nil || g.methods[name]["UnmarshalJSON"] != nil {
			return true
		}
		if !g.aliases[name] {
			return false
		}
		name = pipelineTypeName(g.types[name])
	}
	return false
}

func (g *pipelineGo) fields(name string) (pipelineFieldSet, error) {
	if g.aliases[name] && g.hasCodec(name, nil) {
		return nil, fmt.Errorf("Go %s: unsupported alias inheriting a custom JSON codec", name)
	}
	defaults, err := g.rawFields(g.types[name], nil, map[string]bool{name: true})
	if err != nil {
		return nil, fmt.Errorf("Go %s: %w", name, err)
	}
	encoded, decoded := defaults, defaults
	encodingSource, encodingTarget := "JSON", "default JSON encoding"
	decodingSource, decodingTarget := "JSON", "default JSON decoding"
	for _, codec := range []string{"MarshalJSON", "UnmarshalJSON"} {
		method := g.methods[name][codec]
		if method == nil {
			continue
		}
		fields, err := g.codecFields(name, codec, method)
		if err != nil {
			return nil, err
		}
		if codec == "MarshalJSON" {
			encoded = fields
			encodingSource, encodingTarget = "MarshalJSON", "MarshalJSON wire DTO"
		} else {
			decoded = fields
			decodingSource, decodingTarget = "UnmarshalJSON", "UnmarshalJSON wire DTO"
		}
	}
	// A codec replaces only its own direction. The other direction still uses
	// the default struct fields unless it has a codec of its own.
	for _, field := range sortedPipelineKeys(encoded) {
		if _, ok := decoded[field]; !ok {
			return nil, fmt.Errorf("Go %s: %s field %s absent from %s", name, encodingSource, field, decodingTarget)
		}
	}
	for _, field := range sortedPipelineKeys(decoded) {
		if _, ok := encoded[field]; !ok {
			return nil, fmt.Errorf("Go %s: %s field %s absent from %s", name, decodingSource, field, encodingTarget)
		}
	}
	return encoded, nil
}

func (g *pipelineGo) rawFields(expr ast.Expr, local map[string]pipelineLocalType, stack map[string]bool) (pipelineFieldSet, error) {
	out := pipelineFieldSet{}
	add := func(name, location string) error {
		if prior, exists := out[name]; exists {
			return fmt.Errorf("unsupported conflicting JSON field %s at %s and %s", name, prior, location)
		}
		out[name] = location
		return nil
	}
	if expr == nil {
		return nil, fmt.Errorf("missing struct definition")
	}
	if name := pipelineTypeName(expr); name != "" {
		target := local[name].expr
		if target == nil {
			target = g.types[name]
		}
		if target == nil {
			return nil, fmt.Errorf("unresolved embedded type %s", name)
		}
		if stack[name] {
			return nil, fmt.Errorf("embedded type cycle at %s", name)
		}
		stack[name] = true
		fields, err := g.rawFields(target, local, stack)
		delete(stack, name)
		return fields, err
	}
	st, ok := expr.(*ast.StructType)
	if !ok {
		return nil, fmt.Errorf("unsupported wire type %T (expected struct)", expr)
	}
	for _, field := range st.Fields.List {
		if len(field.Names) == 0 && g.hasCodec(pipelineTypeName(field.Type), local) {
			return nil, fmt.Errorf("unsupported anonymous embedding promoting a JSON codec at %s", g.fset.Position(field.Pos()))
		}
		if len(field.Names) != 0 {
			exported := false
			for _, name := range field.Names {
				exported = exported || name.IsExported()
			}
			if !exported {
				continue
			}
		}
		tag := ""
		if field.Tag != nil {
			value, err := strconv.Unquote(field.Tag.Value)
			if err != nil {
				return nil, err
			}
			tag = strings.Split(reflect.StructTag(value).Get("json"), ",")[0]
		}
		if tag == "-" {
			continue
		}
		loc := g.fset.Position(field.Pos()).String()
		if len(field.Names) == 0 && tag == "" {
			fields, err := g.rawFields(field.Type, local, stack)
			if err != nil {
				return nil, err
			}
			for f, position := range fields {
				if err := add(f, position); err != nil {
					return nil, err
				}
			}
			continue
		}
		if tag != "" {
			if len(field.Names) > 1 {
				return nil, fmt.Errorf("unsupported conflicting JSON tag %s on grouped fields at %s", tag, loc)
			}
			if err := add(tag, loc); err != nil {
				return nil, err
			}
			continue
		}
		for _, name := range field.Names {
			if name.IsExported() {
				if err := add(name.Name, loc); err != nil {
					return nil, err
				}
			}
		}
	}
	return out, nil
}

// codecFields traces the actual argument passed to the JSON helper. Merely
// declaring an unrelated tagged struct inside a codec does not add coverage.
func (g *pipelineGo) codecFields(receiver, codec string, fn *ast.FuncDecl) (pipelineFieldSet, error) {
	localTypes := map[string]pipelineLocalType{}
	variables := map[string]ast.Expr{}
	reassigned := map[string]bool{}
	if len(fn.Recv.List[0].Names) != 0 {
		variables[fn.Recv.List[0].Names[0].Name] = ast.NewIdent(receiver)
	}
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncLit:
			return false
		case *ast.TypeSpec:
			localTypes[n.Name.Name] = pipelineLocalType{expr: n.Type, alias: n.Assign.IsValid()}
		case *ast.ValueSpec:
			for i, name := range n.Names {
				if variables[name.Name] != nil {
					reassigned[name.Name] = true
				}
				if n.Type != nil {
					variables[name.Name] = n.Type
				} else if i < len(n.Values) {
					variables[name.Name] = n.Values[i]
				}
			}
		case *ast.AssignStmt:
			for i, lhs := range n.Lhs {
				if name, ok := lhs.(*ast.Ident); ok && i < len(n.Rhs) {
					if variables[name.Name] != nil {
						reassigned[name.Name] = true
					}
					if n.Tok == token.DEFINE {
						variables[name.Name] = n.Rhs[i]
					}
				}
			}
		}
		return true
	})
	var resolve func(ast.Expr, map[string]bool) (ast.Expr, error)
	resolve = func(expr ast.Expr, seen map[string]bool) (ast.Expr, error) {
		switch e := expr.(type) {
		case *ast.UnaryExpr:
			if e.Op == token.AND {
				return resolve(e.X, seen)
			}
		case *ast.CompositeLit:
			return e.Type, nil
		case *ast.Ident:
			if reassigned[e.Name] {
				return nil, fmt.Errorf("reassigned or shadowed wire variable %s", e.Name)
			}
			if value := variables[e.Name]; value != nil {
				if seen[e.Name] {
					return nil, fmt.Errorf("wire variable cycle at %s", e.Name)
				}
				seen[e.Name] = true
				return resolve(value, seen)
			}
			if g.types[e.Name] != nil || localTypes[e.Name].expr != nil {
				return e, nil
			}
		case *ast.CallExpr:
			if name, ok := e.Fun.(*ast.Ident); ok && (g.types[name.Name] != nil || localTypes[name.Name].expr != nil) {
				return name, nil
			}
		case *ast.StarExpr:
			return resolve(e.X, seen)
		case *ast.StructType:
			return e, nil
		}
		return nil, fmt.Errorf("unsupported wire expression %T", expr)
	}
	var wire pipelineFieldSet
	var problem error
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if problem != nil {
			return false
		}
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		var call *ast.CallExpr
		if codec == "MarshalJSON" {
			ret, ok := node.(*ast.ReturnStmt)
			if !ok {
				return true
			}
			if len(ret.Results) == 1 {
				call, _ = ret.Results[0].(*ast.CallExpr)
			}
			if call == nil {
				problem = fmt.Errorf("unsupported MarshalJSON return shape (expected a direct JSON helper call)")
				return false
			}
		} else {
			var ok bool
			call, ok = node.(*ast.CallExpr)
			if !ok {
				return true
			}
		}
		name := ""
		switch f := call.Fun.(type) {
		case *ast.Ident:
			name = f.Name
		case *ast.SelectorExpr:
			if pkg, ok := f.X.(*ast.Ident); ok && pkg.Name == "json" {
				name = "json." + f.Sel.Name
			}
		}
		index := 0
		if codec == "MarshalJSON" && name != "marshalJSON" && name != "json.Marshal" {
			problem = fmt.Errorf("unsupported MarshalJSON return shape (expected a direct JSON helper call)")
			return false
		}
		if codec == "UnmarshalJSON" {
			if name != "unmarshalJSON" && name != "json.Unmarshal" {
				return true
			}
			index = 1
		}
		if len(call.Args) <= index {
			problem = fmt.Errorf("missing JSON helper argument")
			return false
		}
		expr, err := resolve(call.Args[index], map[string]bool{})
		if err != nil {
			problem = err
			return false
		}
		if wireType := pipelineTypeName(expr); wireType != "" && g.hasCodec(wireType, localTypes) {
			problem = fmt.Errorf("unsupported wire type %s with a custom JSON codec", wireType)
			return false
		}
		fields, err := g.rawFields(expr, localTypes, map[string]bool{})
		if err != nil {
			problem = err
			return false
		}
		if wire != nil && !samePipelineFields(wire, fields) {
			problem = fmt.Errorf("multiple JSON wire DTOs with different fields")
			return false
		}
		wire = fields
		// For marshaling only the returned helper's DTO contributes coverage.
		// Unrelated helper calls and calls in unused closures are irrelevant.
		return codec != "MarshalJSON"
	})
	if problem == nil && wire == nil {
		problem = fmt.Errorf("no supported JSON helper call")
	}
	if problem != nil {
		return nil, fmt.Errorf("%s: unsupported %s.%s codec: %w", g.fset.Position(fn.Pos()), receiver, codec, problem)
	}
	return wire, nil
}

func samePipelineFields(a, b pipelineFieldSet) bool {
	if len(a) != len(b) {
		return false
	}
	for key := range a {
		if _, ok := b[key]; !ok {
			return false
		}
	}
	return true
}
