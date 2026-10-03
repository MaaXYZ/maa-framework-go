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

func pipelineUnparen(expr ast.Expr) ast.Expr {
	for {
		p, ok := expr.(*ast.ParenExpr)
		if !ok {
			return expr
		}
		expr = p.X
	}
}

func pipelineCompositeField(lit *ast.CompositeLit, name string) ast.Expr {
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if key, ok := kv.Key.(*ast.Ident); ok && key.Name == name {
			return kv.Value
		}
	}
	return nil
}

func samePipelineExpr(a, b ast.Expr) bool {
	a, b = pipelineUnparen(a), pipelineUnparen(b)
	switch x := a.(type) {
	case *ast.Ident:
		y, ok := b.(*ast.Ident)
		return ok && x.Name == y.Name
	case *ast.SelectorExpr:
		y, ok := b.(*ast.SelectorExpr)
		return ok && x.Sel.Name == y.Sel.Name && samePipelineExpr(x.X, y.X)
	}
	return false
}

func pipelineDecodeCallName(call *ast.CallExpr) string {
	switch fun := call.Fun.(type) {
	case *ast.Ident:
		if fun.Name == "unmarshalJSON" {
			return fun.Name
		}
	case *ast.SelectorExpr:
		if pkg, ok := fun.X.(*ast.Ident); ok && pkg.Name == "json" && fun.Sel.Name == "Unmarshal" {
			return "json.Unmarshal"
		}
	}
	return ""
}

func pipelineIdentTarget(expr ast.Expr) string {
	expr = pipelineUnparen(expr)
	if unary, ok := expr.(*ast.UnaryExpr); ok && unary.Op == token.AND {
		expr = pipelineUnparen(unary.X)
	}
	id, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}
	return id.Name
}

// pipelineFieldJSONName follows encoding/json naming: an exact "-" tag ignores
// the field, while "-," names it "-" and ",option" keeps the Go field name.
func pipelineFieldJSONName(field *ast.Field) string {
	name := ""
	if field.Tag != nil {
		if value, err := strconv.Unquote(field.Tag.Value); err == nil {
			jsonTag := reflect.StructTag(value).Get("json")
			if jsonTag == "-" {
				return ""
			}
			name = jsonTag
			if idx := strings.IndexByte(name, ','); idx >= 0 {
				name = name[:idx]
			}
		}
	}
	if name != "" {
		return name
	}
	if len(field.Names) == 1 {
		return field.Names[0].Name
	}
	return ""
}

// pipelineDecoder resolves the typed parameter each enum value decodes to. It
// follows the actual UnmarshalJSON data flow: the switch discriminant must be
// the decoded Type, and a helper call only counts when its result is consumed
// by the assignment into receiver.Param. Ambiguous flow is an extraction error.
type pipelineDecoder struct {
	g          *pipelineGo
	kind       string
	typeName   string
	paramIface string
	enumType   string
	recvName   string
	decoded    []ast.Expr
	params     map[string]string
}

// pipelinePointerReceiver reports that a method decodes through a pointer
// receiver; a value receiver cannot update the caller's value.
func pipelinePointerReceiver(fn *ast.FuncDecl) bool {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return false
	}
	_, ok := pipelineUnparen(fn.Recv.List[0].Type).(*ast.StarExpr)
	return ok
}

func (g *pipelineGo) decoderParams(kind string) (map[string]string, error) {
	d := &pipelineDecoder{g: g, kind: kind, params: map[string]string{}}
	switch kind {
	case "action":
		d.typeName, d.paramIface, d.enumType = "Action", "ActionParam", "ActionType"
	case "recognition":
		d.typeName, d.paramIface, d.enumType = "Recognition", "RecognitionParam", "RecognitionType"
	default:
		return nil, fmt.Errorf("Go pipeline: unknown decoder kind %q", kind)
	}
	start := g.methods[d.typeName]["UnmarshalJSON"]
	if start == nil {
		return nil, fmt.Errorf("Go %s: missing UnmarshalJSON decoder", d.typeName)
	}
	if !pipelinePointerReceiver(start) {
		return nil, fmt.Errorf("Go %s: UnmarshalJSON must use a pointer receiver", d.typeName)
	}
	if start.Recv != nil && len(start.Recv.List) != 0 && len(start.Recv.List[0].Names) != 0 {
		d.recvName = start.Recv.List[0].Names[0].Name
	}
	if err := d.analyzeUnmarshal(start); err != nil {
		return nil, err
	}
	return d.params, nil
}

type pipelineVars struct {
	values      map[string]ast.Expr
	valuePos    map[string]token.Pos
	types       map[string]ast.Expr
	assigned    map[string][]ast.Expr
	assignedPos map[string][]token.Pos
	duplicate   map[string]bool
}

func (d *pipelineDecoder) collectVars(body ast.Node) *pipelineVars {
	vars := &pipelineVars{values: map[string]ast.Expr{}, valuePos: map[string]token.Pos{}, types: map[string]ast.Expr{}, assigned: map[string][]ast.Expr{}, assignedPos: map[string][]token.Pos{}, duplicate: map[string]bool{}}
	ast.Inspect(body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncLit:
			return false
		case *ast.ValueSpec:
			for i, name := range n.Names {
				if vars.values[name.Name] != nil || vars.types[name.Name] != nil {
					vars.duplicate[name.Name] = true
				}
				switch {
				case n.Type != nil:
					vars.types[name.Name] = n.Type
				case i < len(n.Values):
					vars.values[name.Name] = n.Values[i]
					vars.valuePos[name.Name] = n.Pos()
				case len(n.Values) == 1 && len(n.Names) > 1:
					vars.values[name.Name] = n.Values[0]
					vars.valuePos[name.Name] = n.Pos()
				}
			}
		case *ast.AssignStmt:
			for i, lhs := range n.Lhs {
				if i >= len(n.Rhs) {
					continue
				}
				id, ok := lhs.(*ast.Ident)
				if !ok {
					continue
				}
				if n.Tok == token.DEFINE {
					if vars.values[id.Name] != nil || vars.types[id.Name] != nil {
						vars.duplicate[id.Name] = true
					}
					vars.values[id.Name] = n.Rhs[i]
					vars.valuePos[id.Name] = n.Pos()
					continue
				}
				vars.assigned[id.Name] = append(vars.assigned[id.Name], n.Rhs[i])
				vars.assignedPos[id.Name] = append(vars.assignedPos[id.Name], n.Pos())
			}
		}
		return true
	})
	return vars
}

func (d *pipelineDecoder) analyzeUnmarshal(fn *ast.FuncDecl) error {
	vars := d.collectVars(fn.Body)
	selectors, err := d.envelopeSelectors(fn, vars)
	if err != nil {
		return err
	}
	d.decoded = append(d.decoded, selectors...)
	d.addReceiverTypeDiscriminant(fn)

	var switches []*ast.SwitchStmt
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		stmt, ok := node.(*ast.SwitchStmt)
		if !ok || stmt.Tag == nil || !d.isDecodedType(stmt.Tag, nil) {
			return true
		}
		switches = append(switches, stmt)
		return true
	})
	inSwitch := func(pos token.Pos) bool {
		for _, stmt := range switches {
			if stmt.Pos() <= pos && pos <= stmt.End() {
				return true
			}
		}
		return false
	}

	if err := d.checkReceiverUse(fn, vars); err != nil {
		return err
	}
	// Only writes that reach receiver.Param participate. Assignments inside a
	// decoder switch are case coverage; every other write must be the single
	// value the receiver ends up with.
	var topWrites, caseWrites []pipelineParamWrite
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, lhs := range assign.Lhs {
			if i >= len(assign.Rhs) {
				continue
			}
			var value ast.Expr
			switch {
			case d.isReceiverParam(lhs):
				value = assign.Rhs[i]
			case d.isReceiverTarget(lhs):
				lit, ok := pipelineUnparen(assign.Rhs[i]).(*ast.CompositeLit)
				if !ok || !d.isEnvelopeType(pipelineTypeName(lit.Type)) {
					continue
				}
				value = pipelineCompositeField(lit, "Param")
			default:
				continue
			}
			if value == nil {
				continue
			}
			write := pipelineParamWrite{expr: value, pos: assign.Pos()}
			if inSwitch(assign.Pos()) {
				caseWrites = append(caseWrites, write)
			} else {
				topWrites = append(topWrites, write)
			}
		}
		return true
	})
	if len(topWrites) > 1 {
		return fmt.Errorf("%s: ambiguous multiple %s decoder Param assignments", d.g.fset.Position(fn.Pos()), d.kind)
	}
	if len(topWrites) == 1 && len(caseWrites) != 0 {
		// The single remaining write must consume the variable the decoder
		// cases assigned; otherwise it overwrites their coverage.
		id, ok := pipelineUnparen(topWrites[0].expr).(*ast.Ident)
		if !ok || !d.caseAssignedVariable(vars, inSwitch, id.Name) {
			return fmt.Errorf("%s: ambiguous %s decoder cases overwritten by the receiver Param assignment", d.g.fset.Position(topWrites[0].pos), d.kind)
		}
	}
	paramFlow := map[string]bool{}
	for _, source := range topWrites {
		if id, ok := pipelineUnparen(source.expr).(*ast.Ident); ok {
			paramFlow[id.Name] = true
		}
	}
	for _, source := range topWrites {
		if err := d.resolveParamSource(source.expr, vars, inSwitch, source.pos); err != nil {
			return err
		}
	}
	for _, stmt := range switches {
		// A switch that never writes Param or a consumed parameter variable
		// adds no coverage; unrelated switches on the decoded type are allowed.
		if !d.switchContributes(stmt, paramFlow) {
			continue
		}
		if err := d.caseParams(stmt, vars, paramFlow, map[*ast.FuncDecl]bool{}); err != nil {
			return err
		}
	}
	return nil
}

// switchContributes reports that a decoder-case clause assigns Param or the
// parameter variable consumed by the receiver assignment.
func (d *pipelineDecoder) switchContributes(stmt *ast.SwitchStmt, paramFlow map[string]bool) bool {
	found := false
	for _, raw := range stmt.Body.List {
		clause, ok := raw.(*ast.CaseClause)
		if !ok || found {
			continue
		}
		hasEnum := false
		for _, label := range clause.List {
			if id, ok := label.(*ast.Ident); ok {
				if _, exists := d.g.enums[d.kind][id.Name]; exists {
					hasEnum = true
					break
				}
			}
		}
		if !hasEnum {
			continue
		}
		ast.Inspect(clause, func(node ast.Node) bool {
			if found {
				return false
			}
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}
			if _, ok := node.(*ast.SwitchStmt); ok {
				return false
			}
			assign, ok := node.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for i, lhs := range assign.Lhs {
				if i >= len(assign.Rhs) {
					continue
				}
				if d.isReceiverParam(lhs) || pipelineIdentIn(lhs, paramFlow) {
					found = true
					return false
				}
				if d.isReceiverTarget(lhs) {
					lit, ok := pipelineUnparen(assign.Rhs[i]).(*ast.CompositeLit)
					if ok && d.isEnvelopeType(pipelineTypeName(lit.Type)) && pipelineCompositeField(lit, "Param") != nil {
						found = true
						return false
					}
				}
			}
			return true
		})
	}
	return found
}

// caseAssignedVariable reports that every plain assignment of name happened
// inside a decoder switch, so its value comes from the enum cases.
func (d *pipelineDecoder) caseAssignedVariable(vars *pipelineVars, inSwitch func(token.Pos) bool, name string) bool {
	positions := vars.assignedPos[name]
	if len(positions) == 0 {
		return false
	}
	for _, at := range positions {
		if !inSwitch(at) {
			return false
		}
	}
	return true
}

// pipelineParamWrite is a value written into receiver.Param and where.
type pipelineParamWrite struct {
	expr ast.Expr
	pos  token.Pos
}

// isEnvelopeType reports that name is the envelope type or an alias of it.
func (d *pipelineDecoder) isEnvelopeType(name string) bool {
	seen := map[string]bool{}
	for name != "" && !seen[name] {
		if name == d.typeName {
			return true
		}
		seen[name] = true
		if !d.g.aliases[name] {
			return false
		}
		name = pipelineTypeName(d.g.types[name])
	}
	return false
}

// checkReceiverUse rejects receiver rebinding, aliasing, shadowing, method
// calls, and writes from inside closures: they escape the supported
// receiver.Param data flow and could hide a different final value.
func (d *pipelineDecoder) checkReceiverUse(fn *ast.FuncDecl, vars *pipelineVars) error {
	if d.recvName == "" {
		return nil
	}
	if vars.duplicate[d.recvName] || len(vars.assignedPos[d.recvName]) != 0 {
		return fmt.Errorf("%s: unsupported %s decoder receiver reassignment", d.g.fset.Position(fn.Pos()), d.kind)
	}
	recvType := &ast.SelectorExpr{X: ast.NewIdent(d.recvName), Sel: ast.NewIdent("Type")}
	var problem error
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if problem != nil {
			return false
		}
		switch n := node.(type) {
		case *ast.FuncLit:
			ast.Inspect(n, func(inner ast.Node) bool {
				assign, ok := inner.(*ast.AssignStmt)
				if !ok {
					return true
				}
				for _, lhs := range assign.Lhs {
					if d.isReceiverParam(lhs) || d.isReceiverTarget(lhs) || samePipelineExpr(lhs, recvType) {
						problem = fmt.Errorf("%s: unsupported %s decoder receiver write inside a closure", d.g.fset.Position(inner.Pos()), d.kind)
						return false
					}
				}
				return true
			})
			return false
		case *ast.AssignStmt:
			for i, lhs := range n.Lhs {
				if i >= len(n.Rhs) {
					continue
				}
				id, ok := lhs.(*ast.Ident)
				if ok && id.Name != d.recvName && pipelineReceiverRef(n.Rhs[i], d.recvName) {
					problem = fmt.Errorf("%s: unsupported %s decoder receiver alias %s", d.g.fset.Position(id.Pos()), d.kind, id.Name)
					return false
				}
			}
		case *ast.CallExpr:
			if sel, ok := n.Fun.(*ast.SelectorExpr); ok && pipelineReceiverRef(sel.X, d.recvName) {
				problem = fmt.Errorf("%s: unsupported %s decoder receiver method call %s", d.g.fset.Position(n.Pos()), d.kind, sel.Sel.Name)
				return false
			}
		}
		return true
	})
	return problem
}

// pipelineReceiverRef reports that expr is the receiver, a pointer to it, or a
// dereference of that pointer.
func pipelineReceiverRef(expr ast.Expr, name string) bool {
	for {
		expr = pipelineUnparen(expr)
		switch e := expr.(type) {
		case *ast.UnaryExpr:
			if e.Op != token.AND {
				return false
			}
			expr = e.X
		case *ast.StarExpr:
			expr = e.X
		default:
			id, ok := expr.(*ast.Ident)
			return ok && id.Name == name
		}
	}
}

// addReceiverTypeDiscriminant accepts receiver.Type as a decoded discriminant
// only when every assignment to it comes from a decoded envelope expression.
func (d *pipelineDecoder) addReceiverTypeDiscriminant(fn *ast.FuncDecl) {
	if d.recvName == "" {
		return
	}
	recvType := &ast.SelectorExpr{X: ast.NewIdent(d.recvName), Sel: ast.NewIdent("Type")}
	var writes []ast.Expr
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, lhs := range assign.Lhs {
			if i < len(assign.Rhs) && samePipelineExpr(lhs, recvType) {
				writes = append(writes, assign.Rhs[i])
			}
		}
		return true
	})
	if len(writes) == 0 {
		return
	}
	for _, expr := range writes {
		if !d.isDecodedType(expr, nil) {
			return
		}
	}
	d.decoded = append(d.decoded, recvType)
}

// pipelineEnvelope records a local envelope struct decoded from the method's
// JSON data argument and the decoded position of its Type field.
type pipelineEnvelope struct {
	field   string
	decoded token.Pos
}

// envelopeSelectors returns the decoded Type selectors of every raw envelope
// struct decoded from the method's JSON data argument. A selector only counts
// while the establishing decode is the last thing that can set it: a later
// write to the envelope or its Type field, a write from a closure, or an alias
// or pointer escape of the envelope is an extraction error.
func (d *pipelineDecoder) envelopeSelectors(fn *ast.FuncDecl, vars *pipelineVars) ([]ast.Expr, error) {
	dataParam := ""
	if fn.Type.Params != nil && len(fn.Type.Params.List) != 0 && len(fn.Type.Params.List[0].Names) != 0 {
		dataParam = fn.Type.Params.List[0].Names[0].Name
	}
	if dataParam == "" {
		return nil, nil
	}
	var order []string
	envelopes := map[string]pipelineEnvelope{}
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) < 2 || pipelineDecodeCallName(call) == "" {
			return true
		}
		if pipelineIdentTarget(call.Args[0]) != dataParam {
			return true
		}
		name := pipelineIdentTarget(call.Args[1])
		if name == "" || (vars.types[name] == nil && vars.values[name] == nil) {
			return true
		}
		st, ok := d.structType(name, vars)
		if !ok {
			return true
		}
		for _, field := range st.Fields.List {
			if len(field.Names) != 1 {
				continue
			}
			// The envelope tag itself is validated separately; accept the Go
			// field name too so a mutated tag still reaches that comparison.
			if pipelineFieldJSONName(field) != "type" && field.Names[0].Name != "Type" {
				continue
			}
			env, exists := envelopes[name]
			if !exists {
				envelopes[name] = pipelineEnvelope{field: field.Names[0].Name, decoded: call.Pos()}
				order = append(order, name)
			} else if call.Pos() < env.decoded {
				env.decoded = call.Pos()
				envelopes[name] = env
			}
		}
		return true
	})
	var out []ast.Expr
	for _, name := range order {
		env := envelopes[name]
		if err := d.checkEnvelopeUse(fn, dataParam, name, env.field, env.decoded); err != nil {
			return nil, err
		}
		out = append(out, &ast.SelectorExpr{X: ast.NewIdent(name), Sel: ast.NewIdent(env.field)})
	}
	return out, nil
}

// checkEnvelopeUse rejects writes to the decoded envelope or its Type field,
// writes from a closure, and aliases or pointer escapes that could rewrite the
// discriminant after the establishing decode.
func (d *pipelineDecoder) checkEnvelopeUse(fn *ast.FuncDecl, dataParam, name, field string, decoded token.Pos) error {
	if err := d.scanEnvelopeUse(fn.Body, dataParam, name, field, decoded, false); err != nil {
		return err
	}
	var problem error
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if problem != nil {
			return false
		}
		lit, ok := node.(*ast.FuncLit)
		if !ok {
			return true
		}
		problem = d.scanEnvelopeUse(lit.Body, dataParam, name, field, decoded, true)
		return false
	})
	return problem
}

// scanEnvelopeUse reports the first unsupported use of the decoded envelope in
// body. With closure set, writes are rejected regardless of their position
// because the invocation time of the closure cannot be proved.
func (d *pipelineDecoder) scanEnvelopeUse(body ast.Node, dataParam, name, field string, decoded token.Pos, closure bool) error {
	var problem error
	report := func(err error) {
		if problem == nil {
			problem = err
		}
	}
	checkWrite := func(lhs ast.Expr) {
		if target, ok := pipelineEnvelopeWriteTarget(lhs, name, field); ok {
			switch {
			case closure:
				report(fmt.Errorf("%s: unsupported %s decoder rewrite of decoded envelope %s inside a closure", d.g.fset.Position(lhs.Pos()), d.kind, target))
			case lhs.Pos() > decoded:
				report(fmt.Errorf("%s: unsupported %s decoder rewrite of decoded envelope %s after its JSON decode", d.g.fset.Position(lhs.Pos()), d.kind, target))
			}
		}
	}
	ast.Inspect(body, func(node ast.Node) bool {
		if problem != nil {
			return false
		}
		switch n := node.(type) {
		case *ast.FuncLit:
			// Nested closures are covered by the outer closure scan.
			return closure
		case *ast.AssignStmt:
			for i, lhs := range n.Lhs {
				checkWrite(lhs)
				if pipelineBlankIdent(lhs) || i >= len(n.Rhs) {
					continue
				}
				if target, ok := pipelineEnvelopeEscape(n.Rhs[i], name, field); ok {
					report(fmt.Errorf("%s: unsupported %s decoder alias of decoded envelope %s", d.g.fset.Position(lhs.Pos()), d.kind, target))
				}
			}
		case *ast.ValueSpec:
			blank := true
			for _, id := range n.Names {
				if id.Name != "_" {
					blank = false
					break
				}
			}
			if blank {
				return true
			}
			for _, value := range n.Values {
				if target, ok := pipelineEnvelopeEscape(value, name, field); ok {
					report(fmt.Errorf("%s: unsupported %s decoder alias of decoded envelope %s", d.g.fset.Position(value.Pos()), d.kind, target))
				}
			}
		case *ast.CallExpr:
			if pipelineEstablishingDecode(n, dataParam, name) {
				return true
			}
			for _, arg := range n.Args {
				if target, ok := pipelineEnvelopeEscape(arg, name, field); ok {
					report(fmt.Errorf("%s: unsupported %s decoder escape of decoded envelope %s", d.g.fset.Position(arg.Pos()), d.kind, target))
				}
			}
			if sel, ok := pipelineUnparen(n.Fun).(*ast.SelectorExpr); ok {
				if target, ok := pipelineEnvelopeWriteTarget(sel.X, name, field); ok {
					report(fmt.Errorf("%s: unsupported %s decoder method call on decoded envelope %s", d.g.fset.Position(n.Pos()), d.kind, target))
				}
			}
		case *ast.SendStmt:
			if target, ok := pipelineEnvelopeEscape(n.Value, name, field); ok {
				report(fmt.Errorf("%s: unsupported %s decoder escape of decoded envelope %s through a channel", d.g.fset.Position(n.Value.Pos()), d.kind, target))
			}
		case *ast.ReturnStmt:
			for _, result := range n.Results {
				if target, ok := pipelineEnvelopeEscape(result, name, field); ok {
					report(fmt.Errorf("%s: unsupported %s decoder escape of decoded envelope %s through a return", d.g.fset.Position(result.Pos()), d.kind, target))
				}
			}
		case *ast.RangeStmt:
			for _, lhs := range []ast.Expr{n.Key, n.Value} {
				checkWrite(lhs)
			}
			if target, ok := pipelineEnvelopeEscape(n.X, name, field); ok {
				report(fmt.Errorf("%s: unsupported %s decoder escape of decoded envelope %s through a range", d.g.fset.Position(n.X.Pos()), d.kind, target))
			}
		}
		return true
	})
	return problem
}

// pipelineEstablishingDecode reports that call is the supported JSON decode of
// the method's data argument into the envelope variable. That pointer escape
// establishes the discriminant instead of hiding a write.
func pipelineEstablishingDecode(call *ast.CallExpr, dataParam, name string) bool {
	return len(call.Args) >= 2 &&
		pipelineDecodeCallName(call) != "" &&
		pipelineIdentTarget(call.Args[0]) == dataParam &&
		pipelineIdentTarget(call.Args[1]) == name
}

// pipelineEnvelopeWriteTarget reports the decoded envelope reference written by
// lhs, if any: the envelope itself or its Type field, through pointers,
// dereferences, and parentheses.
func pipelineEnvelopeWriteTarget(lhs ast.Expr, name, field string) (string, bool) {
	lhs = pipelineReferenceBase(lhs)
	if sel, ok := lhs.(*ast.SelectorExpr); ok {
		if sel.Sel.Name == field && pipelineEnvelopeRef(sel.X, name) {
			return name + "." + field, true
		}
		return "", false
	}
	if pipelineEnvelopeRef(lhs, name) {
		return name, true
	}
	return "", false
}

// pipelineEnvelopeEscape reports the decoded envelope reference handed out by
// expr, if any: the envelope itself or an address of its Type field, including
// references nested in composite literals. Ordinary Type/Param reads are safe.
func pipelineEnvelopeEscape(expr ast.Expr, name, field string) (string, bool) {
	if pipelineEnvelopeRef(expr, name) {
		return name, true
	}
	switch e := pipelineUnparen(expr).(type) {
	case *ast.UnaryExpr:
		if e.Op == token.AND {
			if target, ok := pipelineEnvelopeWriteTarget(e.X, name, field); ok {
				return target, true
			}
			return pipelineEnvelopeEscape(e.X, name, field)
		}
	case *ast.CompositeLit:
		for _, element := range e.Elts {
			if kv, ok := element.(*ast.KeyValueExpr); ok {
				if target, ok := pipelineEnvelopeEscape(kv.Key, name, field); ok {
					return target, true
				}
				element = kv.Value
			}
			if target, ok := pipelineEnvelopeEscape(element, name, field); ok {
				return target, true
			}
		}
	}
	return "", false
}

// pipelineEnvelopeRef reports that expr is the decoded envelope or a pointer or
// dereference of it, ignoring parentheses.
func pipelineEnvelopeRef(expr ast.Expr, name string) bool {
	id, ok := pipelineReferenceBase(expr).(*ast.Ident)
	return ok && id.Name == name
}

// pipelineReferenceBase removes pointer, dereference, and parenthesis wrappers
// without following aliases or evaluating expressions.
func pipelineReferenceBase(expr ast.Expr) ast.Expr {
	for {
		expr = pipelineUnparen(expr)
		switch e := expr.(type) {
		case *ast.UnaryExpr:
			if e.Op != token.AND {
				return expr
			}
			expr = e.X
		case *ast.StarExpr:
			expr = e.X
		default:
			return expr
		}
	}
}

func pipelineBlankIdent(expr ast.Expr) bool {
	id, ok := pipelineUnparen(expr).(*ast.Ident)
	return ok && id.Name == "_"
}

func (d *pipelineDecoder) structType(name string, vars *pipelineVars) (*ast.StructType, bool) {
	typ := vars.types[name]
	if typ == nil {
		typ = vars.values[name]
	}
	seen := map[string]bool{}
	for typ != nil {
		typ = pipelineUnparen(typ)
		switch e := typ.(type) {
		case *ast.StructType:
			return e, true
		case *ast.CompositeLit:
			typ = e.Type
		case *ast.UnaryExpr:
			if e.Op != token.AND {
				return nil, false
			}
			typ = e.X
		case *ast.StarExpr:
			typ = e.X
		case *ast.Ident:
			if seen[e.Name] {
				return nil, false
			}
			seen[e.Name] = true
			typ = d.g.types[e.Name]
		default:
			return nil, false
		}
	}
	return nil, false
}

func (d *pipelineDecoder) isReceiverParam(expr ast.Expr) bool {
	sel, ok := pipelineUnparen(expr).(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Param" {
		return false
	}
	return d.recvName != "" && pipelineReceiverRef(sel.X, d.recvName)
}

func (d *pipelineDecoder) isReceiverTarget(expr ast.Expr) bool {
	if d.recvName == "" {
		return false
	}
	switch e := pipelineUnparen(expr).(type) {
	case *ast.Ident:
		return e.Name == d.recvName
	case *ast.StarExpr:
		id, ok := pipelineUnparen(e.X).(*ast.Ident)
		return ok && id.Name == d.recvName
	}
	return false
}

func (d *pipelineDecoder) isDecodedType(expr ast.Expr, allowed map[string]bool) bool {
	expr = pipelineUnparen(expr)
	if id, ok := expr.(*ast.Ident); ok && allowed[id.Name] {
		return true
	}
	for _, decoded := range d.decoded {
		if samePipelineExpr(expr, decoded) {
			return true
		}
	}
	return false
}

func (d *pipelineDecoder) record(value, param string) error {
	if prior := d.params[value]; prior != "" && prior != param {
		return fmt.Errorf("Go %s.%s: conflicting decoder parameters %s and %s", d.kind, value, prior, param)
	}
	d.params[value] = param
	return nil
}

// resolveParamSource verifies that the value assigned to receiver.Param comes
// from a decoder helper call, the variable filled by decoder cases, or a
// fallback literal that establishes no typed coverage. writePos is the
// receiver assignment; consuming a variable before it is assigned is rejected.
func (d *pipelineDecoder) resolveParamSource(expr ast.Expr, vars *pipelineVars, inSwitch func(token.Pos) bool, writePos token.Pos) error {
	return d.resolveParamSourceSeen(expr, vars, inSwitch, writePos, map[string]bool{})
}

func (d *pipelineDecoder) resolveParamSourceSeen(expr ast.Expr, vars *pipelineVars, inSwitch func(token.Pos) bool, writePos token.Pos, seen map[string]bool) error {
	expr = pipelineUnparen(expr)
	if unary, ok := expr.(*ast.UnaryExpr); ok && unary.Op == token.AND {
		expr = pipelineUnparen(unary.X)
	}
	switch e := expr.(type) {
	case *ast.Ident:
		if e.Name == "nil" {
			// An explicit nil Param is a fallback, not typed coverage.
			return nil
		}
		if vars.duplicate[e.Name] {
			return fmt.Errorf("%s: ambiguous shadowed %s decoder variable %s", d.g.fset.Position(e.Pos()), d.kind, e.Name)
		}
		if value := vars.values[e.Name]; value != nil {
			if len(vars.assigned[e.Name]) != 0 {
				return fmt.Errorf("%s: ambiguous %s decoder variable %s is assigned more than once", d.g.fset.Position(e.Pos()), d.kind, e.Name)
			}
			if at := vars.valuePos[e.Name]; at != 0 && writePos < at {
				return fmt.Errorf("%s: %s decoder variable %s is consumed before it is assigned", d.g.fset.Position(e.Pos()), d.kind, e.Name)
			}
			if seen[e.Name] {
				return fmt.Errorf("%s: cycle in %s decoder parameter %s", d.g.fset.Position(e.Pos()), d.kind, e.Name)
			}
			seen[e.Name] = true
			return d.resolveParamSourceSeen(value, vars, inSwitch, writePos, seen)
		}
		values := vars.assigned[e.Name]
		positions := vars.assignedPos[e.Name]
		if len(values) == 0 {
			return fmt.Errorf("%s: unsupported %s decoder parameter variable %s (never assigned)", d.g.fset.Position(e.Pos()), d.kind, e.Name)
		}
		nonCase := 0
		var call *ast.CallExpr
		last := token.NoPos
		for i, assigned := range values {
			if at := positions[i]; at > last {
				last = at
			}
			if inSwitch(assigned.Pos()) {
				continue
			}
			nonCase++
			if c, ok := pipelineUnparen(assigned).(*ast.CallExpr); ok {
				call = c
			}
		}
		if writePos < last {
			return fmt.Errorf("%s: %s decoder variable %s is consumed before it is assigned", d.g.fset.Position(e.Pos()), d.kind, e.Name)
		}
		switch {
		case nonCase == 0:
			// Decoder cases establish the coverage for this variable.
			return nil
		case nonCase == 1 && call != nil && len(values) == 1:
			params, err := d.helperParams(call, vars, map[*ast.FuncDecl]bool{}, nil)
			if err != nil {
				return err
			}
			for _, value := range sortedPipelineKeys(params) {
				if err := d.record(value, params[value]); err != nil {
					return err
				}
			}
			return nil
		default:
			return fmt.Errorf("%s: ambiguous %s decoder parameter variable %s has multiple assignments", d.g.fset.Position(e.Pos()), d.kind, e.Name)
		}
	case *ast.CompositeLit:
		return nil
	case *ast.CallExpr:
		params, err := d.helperParams(e, vars, map[*ast.FuncDecl]bool{}, nil)
		if err != nil {
			return err
		}
		for _, value := range sortedPipelineKeys(params) {
			if err := d.record(value, params[value]); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("%s: unsupported %s decoder parameter source %T", d.g.fset.Position(expr.Pos()), d.kind, expr)
}

func pipelineParamExpr(expr ast.Expr, vars *pipelineVars) bool {
	expr = pipelineUnparen(expr)
	switch e := expr.(type) {
	case *ast.CompositeLit, *ast.CallExpr:
		return true
	case *ast.UnaryExpr:
		if e.Op != token.AND {
			return false
		}
		_, ok := pipelineUnparen(e.X).(*ast.CompositeLit)
		return ok
	case *ast.Ident:
		return vars.values[e.Name] != nil
	}
	return false
}

func pipelineIdentIn(expr ast.Expr, names map[string]bool) bool {
	id, ok := pipelineUnparen(expr).(*ast.Ident)
	return ok && names[id.Name]
}

// caseParams maps enum values from a switch on the decoded Type to the typed
// parameters assigned to Param or returned by each case clause.
func (d *pipelineDecoder) caseParams(stmt *ast.SwitchStmt, vars *pipelineVars, paramFlow map[string]bool, stack map[*ast.FuncDecl]bool) error {
	for _, raw := range stmt.Body.List {
		clause, ok := raw.(*ast.CaseClause)
		if !ok {
			continue
		}
		var values []string
		for _, label := range clause.List {
			if id, ok := label.(*ast.Ident); ok {
				if value, exists := d.g.enums[d.kind][id.Name]; exists {
					values = append(values, value)
				}
			}
		}
		if len(values) == 0 {
			continue
		}
		candidates := map[string]bool{}
		var problem error
		record := func(expr ast.Expr) {
			if problem != nil || !pipelineParamExpr(expr, vars) {
				return
			}
			problem = d.collectCandidate(expr, candidates, vars, stack, nil)
		}
		ast.Inspect(clause, func(node ast.Node) bool {
			if problem != nil {
				return false
			}
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}
			if _, ok := node.(*ast.SwitchStmt); ok {
				return false
			}
			switch n := node.(type) {
			case *ast.AssignStmt:
				for i, lhs := range n.Lhs {
					if i >= len(n.Rhs) {
						continue
					}
					if d.isReceiverParam(lhs) || pipelineIdentIn(lhs, paramFlow) {
						if id, ok := pipelineUnparen(lhs).(*ast.Ident); ok && vars.duplicate[id.Name] {
							problem = fmt.Errorf("%s: ambiguous shadowed %s decoder variable %s", d.g.fset.Position(id.Pos()), d.kind, id.Name)
							return false
						}
						record(n.Rhs[i])
					}
					if lit, ok := pipelineUnparen(n.Rhs[i]).(*ast.CompositeLit); ok && d.isReceiverTarget(lhs) && d.isEnvelopeType(pipelineTypeName(lit.Type)) {
						if value := pipelineCompositeField(lit, "Param"); value != nil {
							record(value)
						}
					}
				}
			}
			return true
		})
		if problem != nil {
			return problem
		}
		if len(candidates) != 1 {
			return fmt.Errorf("%s: unsupported typed %s decoder case (expected one parameter literal assigned to Param or returned)", d.g.fset.Position(clause.Pos()), d.kind)
		}
		for param := range candidates {
			for _, value := range values {
				if err := d.record(value, param); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (d *pipelineDecoder) collectCandidate(expr ast.Expr, candidates map[string]bool, vars *pipelineVars, stack map[*ast.FuncDecl]bool, allowed map[string]bool) error {
	return d.collectCandidateSeen(expr, candidates, vars, stack, allowed, map[string]bool{})
}

func (d *pipelineDecoder) collectCandidateSeen(expr ast.Expr, candidates map[string]bool, vars *pipelineVars, stack map[*ast.FuncDecl]bool, allowed, seen map[string]bool) error {
	expr = pipelineUnparen(expr)
	if unary, ok := expr.(*ast.UnaryExpr); ok && unary.Op == token.AND {
		expr = pipelineUnparen(unary.X)
	}
	switch e := expr.(type) {
	case *ast.CompositeLit:
		name := pipelineTypeName(e.Type)
		if name == "" || d.g.types[name] == nil {
			return fmt.Errorf("%s: unsupported %s decoder parameter type", d.g.fset.Position(e.Pos()), d.kind)
		}
		candidates[name] = true
		return nil
	case *ast.Ident:
		if vars.duplicate[e.Name] {
			return fmt.Errorf("%s: ambiguous shadowed %s decoder variable %s", d.g.fset.Position(e.Pos()), d.kind, e.Name)
		}
		if value := vars.values[e.Name]; value != nil {
			if seen[e.Name] {
				return fmt.Errorf("%s: cycle in %s decoder parameter %s", d.g.fset.Position(e.Pos()), d.kind, e.Name)
			}
			seen[e.Name] = true
			return d.collectCandidateSeen(value, candidates, vars, stack, allowed, seen)
		}
		return fmt.Errorf("%s: unsupported %s decoder parameter expression %s", d.g.fset.Position(e.Pos()), d.kind, e.Name)
	case *ast.CallExpr:
		params, err := d.helperParams(e, vars, stack, allowed)
		if err != nil {
			return err
		}
		for _, param := range params {
			candidates[param] = true
		}
		return nil
	}
	return fmt.Errorf("%s: unsupported %s decoder parameter expression %T", d.g.fset.Position(expr.Pos()), d.kind, expr)
}

// helperParams extracts the enum coverage of a decoder helper. The helper must
// return the parameter interface, take the enum type, switch on that parameter,
// and assign each typed parameter to a returned variable.
func (d *pipelineDecoder) helperParams(call *ast.CallExpr, vars *pipelineVars, stack map[*ast.FuncDecl]bool, allowed map[string]bool) (map[string]string, error) {
	pos := d.g.fset.Position(call.Pos())
	id, ok := call.Fun.(*ast.Ident)
	if !ok {
		return nil, fmt.Errorf("%s: unsupported %s decoder helper call", pos, d.kind)
	}
	if vars.duplicate[id.Name] || vars.values[id.Name] != nil || vars.types[id.Name] != nil {
		return nil, fmt.Errorf("%s: unsupported %s decoder helper %s (shadowed by a local variable)", pos, d.kind, id.Name)
	}
	fn := d.g.funcs[id.Name]
	if fn == nil {
		return nil, fmt.Errorf("%s: unsupported %s decoder helper %s (missing declaration)", pos, d.kind, id.Name)
	}
	if stack[fn] {
		return nil, fmt.Errorf("%s: cyclic %s decoder helper %s", pos, d.kind, id.Name)
	}
	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 || pipelineTypeName(fn.Type.Results.List[0].Type) != d.paramIface {
		return nil, fmt.Errorf("%s: unsupported %s decoder helper %s (must return %s)", pos, d.kind, id.Name, d.paramIface)
	}
	type helperParam struct {
		name string
		typ  ast.Expr
	}
	var params []helperParam
	for _, field := range fn.Type.Params.List {
		if len(field.Names) == 0 {
			params = append(params, helperParam{"", field.Type})
			continue
		}
		for _, name := range field.Names {
			params = append(params, helperParam{name.Name, field.Type})
		}
	}
	typeIndex := -1
	for i, param := range params {
		if pipelineTypeName(param.typ) == d.enumType {
			typeIndex = i
			break
		}
	}
	if typeIndex < 0 || params[typeIndex].name == "" {
		return nil, fmt.Errorf("%s: unsupported %s decoder helper %s (missing %s parameter)", pos, d.kind, id.Name, d.enumType)
	}
	if typeIndex >= len(call.Args) || !d.isDecodedType(call.Args[typeIndex], allowed) {
		return nil, fmt.Errorf("%s: %s decoder helper %s is not called with the decoded %s", pos, d.kind, id.Name, d.enumType)
	}
	paramName := params[typeIndex].name
	if err := d.checkHelperParamUse(fn, paramName); err != nil {
		return nil, err
	}
	stack[fn] = true
	defer delete(stack, fn)

	bodyVars := d.collectVars(fn.Body)
	returned := map[string]bool{}
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		ret, ok := node.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, result := range ret.Results {
			if ident, ok := pipelineUnparen(result).(*ast.Ident); ok {
				returned[ident.Name] = true
			}
		}
		return true
	})
	var switches []*ast.SwitchStmt
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		stmt, ok := node.(*ast.SwitchStmt)
		if !ok || stmt.Tag == nil {
			return true
		}
		tag, ok := pipelineUnparen(stmt.Tag).(*ast.Ident)
		if !ok || tag.Name != paramName {
			return true
		}
		switches = append(switches, stmt)
		return true
	})
	if len(switches) == 0 {
		return nil, fmt.Errorf("%s: %s decoder helper %s does not switch on its %s parameter", pos, d.kind, id.Name, d.enumType)
	}
	inSwitch := func(at token.Pos) bool {
		for _, stmt := range switches {
			if stmt.Pos() <= at && at <= stmt.End() {
				return true
			}
		}
		return false
	}
	for _, stmt := range fn.Body.List {
		ret, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			continue
		}
		for _, sw := range switches {
			if ret.Pos() < sw.Pos() {
				return nil, fmt.Errorf("%s: unreachable %s decoder helper %s switch", pos, d.kind, id.Name)
			}
		}
	}
	inner := map[string]bool{paramName: true}
	for name := range allowed {
		inner[name] = true
	}
	out := map[string]string{}
	var problem error
	for _, stmt := range switches {
		if problem != nil {
			break
		}
		problem = d.helperCases(stmt, bodyVars, returned, stack, inner, out)
	}
	if problem != nil {
		return nil, problem
	}
	// A returned parameter assigned a typed value by the switch must not be
	// overwritten outside the switch, or the case mapping never reaches the
	// caller. Unrelated returned names (err, nil) are ignored.
	paramVars := map[string]bool{}
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		assign, ok := node.(*ast.AssignStmt)
		if !ok || !inSwitch(assign.Pos()) {
			return true
		}
		for i, lhs := range assign.Lhs {
			if i >= len(assign.Rhs) || !pipelineParamExpr(assign.Rhs[i], bodyVars) {
				continue
			}
			if name, ok := pipelineUnparen(lhs).(*ast.Ident); ok {
				paramVars[name.Name] = true
			}
		}
		return true
	})
	var overwrite error
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if overwrite != nil {
			return false
		}
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		assign, ok := node.(*ast.AssignStmt)
		if !ok || inSwitch(assign.Pos()) {
			return true
		}
		for _, lhs := range assign.Lhs {
			name, ok := pipelineUnparen(lhs).(*ast.Ident)
			if ok && paramVars[name.Name] {
				overwrite = fmt.Errorf("%s: %s decoder helper assigns case parameter %s outside its switch", d.g.fset.Position(assign.Pos()), d.kind, name.Name)
				return false
			}
		}
		return true
	})
	if overwrite != nil {
		return nil, overwrite
	}
	return out, nil
}

func (d *pipelineDecoder) helperCases(stmt *ast.SwitchStmt, vars *pipelineVars, returned map[string]bool, stack map[*ast.FuncDecl]bool, allowed map[string]bool, out map[string]string) error {
	for _, raw := range stmt.Body.List {
		clause, ok := raw.(*ast.CaseClause)
		if !ok {
			continue
		}
		var values []string
		for _, label := range clause.List {
			if id, ok := label.(*ast.Ident); ok {
				if value, exists := d.g.enums[d.kind][id.Name]; exists {
					values = append(values, value)
				}
			}
		}
		if len(values) == 0 {
			continue
		}
		candidates := map[string]bool{}
		var problem error
		record := func(expr ast.Expr) {
			if problem != nil || !pipelineParamExpr(expr, vars) {
				return
			}
			problem = d.collectCandidate(expr, candidates, vars, stack, allowed)
		}
		ast.Inspect(clause, func(node ast.Node) bool {
			if problem != nil {
				return false
			}
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}
			if _, ok := node.(*ast.SwitchStmt); ok {
				return false
			}
			switch n := node.(type) {
			case *ast.ReturnStmt:
				for _, result := range n.Results {
					record(result)
				}
			case *ast.AssignStmt:
				for i, lhs := range n.Lhs {
					if i >= len(n.Rhs) {
						continue
					}
					if pipelineIdentIn(lhs, returned) {
						if id, ok := pipelineUnparen(lhs).(*ast.Ident); ok && vars.duplicate[id.Name] {
							problem = fmt.Errorf("%s: ambiguous shadowed %s decoder variable %s", d.g.fset.Position(id.Pos()), d.kind, id.Name)
							return false
						}
						record(n.Rhs[i])
					}
				}
			}
			return true
		})
		if problem != nil {
			return problem
		}
		if len(candidates) != 1 {
			return fmt.Errorf("%s: unsupported typed %s decoder case (expected one parameter literal assigned to the returned Param)", d.g.fset.Position(clause.Pos()), d.kind)
		}
		for param := range candidates {
			for _, value := range values {
				if prior := out[value]; prior != "" && prior != param {
					return fmt.Errorf("Go %s.%s: conflicting decoder parameters %s and %s", d.kind, value, prior, param)
				}
				out[value] = param
			}
		}
	}
	return nil
}

// checkHelperParamUse rejects a decoder helper that reassigns or locally
// shadows its enum parameter, including writes from closures: the helper switch
// no longer proves that the caller's decoded discriminant drives the case
// mapping.
func (d *pipelineDecoder) checkHelperParamUse(fn *ast.FuncDecl, name string) error {
	if err := d.scanHelperParamWrites(fn.Body, fn.Name.Name, name, false); err != nil {
		return err
	}
	var problem error
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if problem != nil {
			return false
		}
		lit, ok := node.(*ast.FuncLit)
		if !ok {
			return true
		}
		problem = d.scanHelperParamWrites(lit.Body, fn.Name.Name, name, true)
		return false
	})
	return problem
}

// scanHelperParamWrites reports the first write to name in body. With closure
// set, the write is rejected regardless of its position because the invocation
// time of the closure cannot be proved.
func (d *pipelineDecoder) scanHelperParamWrites(body ast.Node, helper, name string, closure bool) error {
	var problem error
	report := func(pos token.Pos, action string) {
		if problem != nil {
			return
		}
		if closure {
			problem = fmt.Errorf("%s: unsupported %s decoder helper %s writes its %s parameter %s inside a closure", d.g.fset.Position(pos), d.kind, helper, d.enumType, name)
			return
		}
		problem = fmt.Errorf("%s: unsupported %s decoder helper %s %s its %s parameter %s", d.g.fset.Position(pos), d.kind, helper, action, d.enumType, name)
	}
	ast.Inspect(body, func(node ast.Node) bool {
		if problem != nil {
			return false
		}
		switch n := node.(type) {
		case *ast.FuncLit:
			// Nested closures are covered by the outer closure scan.
			return closure
		case *ast.AssignStmt:
			for _, lhs := range n.Lhs {
				if pipelineEnvelopeRef(lhs, name) {
					action := "reassigns"
					if n.Tok == token.DEFINE {
						action = "shadows"
					}
					report(lhs.Pos(), action)
					return false
				}
			}
		case *ast.ValueSpec:
			for _, id := range n.Names {
				if id.Name == name {
					report(id.Pos(), "shadows")
					return false
				}
			}
		case *ast.RangeStmt:
			for _, expr := range []ast.Expr{n.Key, n.Value} {
				if pipelineEnvelopeRef(expr, name) {
					action := "reassigns"
					if n.Tok == token.DEFINE {
						action = "shadows"
					}
					report(expr.Pos(), action)
					return false
				}
			}
		case *ast.UnaryExpr:
			if n.Op == token.AND && pipelineEnvelopeRef(n.X, name) {
				report(n.Pos(), "takes the address of")
				return false
			}
		case *ast.CallExpr:
			if sel, ok := pipelineUnparen(n.Fun).(*ast.SelectorExpr); ok && pipelineEnvelopeRef(sel.X, name) {
				report(n.Pos(), "calls a method on")
				return false
			}
		}
		return true
	})
	return problem
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
	if err := samePipelineDirections(name, encoded, decoded, encodingSource, encodingTarget, decodingSource, decodingTarget); err != nil {
		return nil, err
	}
	return encoded, nil
}

// envelopeFields extracts the wire fields of the pipeline envelope type. Unlike
// parameter codecs, the decoder's first JSON call is the envelope DTO; later
// calls decode the typed parameter and are handled by decoderParams.
func (g *pipelineGo) envelopeFields(name string) (pipelineFieldSet, error) {
	var encoded, decoded pipelineFieldSet
	encodingSource, encodingTarget := "JSON", "default JSON encoding"
	decodingSource, decodingTarget := "JSON", "default JSON decoding"
	if method := g.methods[name]["MarshalJSON"]; method != nil {
		fields, err := g.codecFields(name, "MarshalJSON", method)
		if err != nil {
			return nil, err
		}
		encoded = fields
		encodingSource, encodingTarget = "MarshalJSON", "MarshalJSON wire DTO"
	}
	if method := g.methods[name]["UnmarshalJSON"]; method != nil {
		fields, err := g.envelopeDecodeFields(name, method)
		if err != nil {
			return nil, err
		}
		decoded = fields
		decodingSource, decodingTarget = "UnmarshalJSON", "UnmarshalJSON wire DTO"
	}
	// Only directions without a codec fall back to the struct fields. This keeps
	// an embedded codec-bearing field from being rejected when both directions
	// have their own wire DTO.
	defaults := pipelineFieldSet(nil)
	for _, fields := range []*pipelineFieldSet{&encoded, &decoded} {
		if *fields != nil {
			continue
		}
		if defaults == nil {
			value, err := g.rawFields(g.types[name], nil, map[string]bool{name: true})
			if err != nil {
				return nil, fmt.Errorf("Go %s: %w", name, err)
			}
			defaults = value
		}
		*fields = defaults
	}
	if err := samePipelineDirections(name, encoded, decoded, encodingSource, encodingTarget, decodingSource, decodingTarget); err != nil {
		return nil, err
	}
	return encoded, nil
}

func samePipelineDirections(name string, encoded, decoded pipelineFieldSet, encodingSource, encodingTarget, decodingSource, decodingTarget string) error {
	for _, field := range sortedPipelineKeys(encoded) {
		if _, ok := decoded[field]; !ok {
			return fmt.Errorf("Go %s: %s field %s absent from %s", name, encodingSource, field, decodingTarget)
		}
	}
	for _, field := range sortedPipelineKeys(decoded) {
		if _, ok := encoded[field]; !ok {
			return fmt.Errorf("Go %s: %s field %s absent from %s", name, decodingSource, field, encodingTarget)
		}
	}
	return nil
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
		tag := ""
		if field.Tag != nil {
			value, err := strconv.Unquote(field.Tag.Value)
			if err != nil {
				return nil, err
			}
			jsonTag := reflect.StructTag(value).Get("json")
			// encoding/json ignores exactly "-"; "-," names the field "-".
			if jsonTag == "-" {
				continue
			}
			if idx := strings.IndexByte(jsonTag, ','); idx >= 0 {
				jsonTag = jsonTag[:idx]
			}
			tag = jsonTag
		}
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

type pipelineCodecScope struct {
	local      map[string]pipelineLocalType
	values     map[string]ast.Expr
	reassigned map[string]bool
}

// codecScope collects the codec's local types and variable bindings. Shadowed
// local types, local types replacing package types, and variables that reuse a
// type name are rejected instead of guessing which declaration a name refers
// to.
func (g *pipelineGo) codecScope(receiver string, fn *ast.FuncDecl) (*pipelineCodecScope, error) {
	scope := &pipelineCodecScope{local: map[string]pipelineLocalType{}, values: map[string]ast.Expr{}, reassigned: map[string]bool{}}
	duplicates := map[string]bool{}
	variables := map[string]bool{}
	if len(fn.Recv.List[0].Names) != 0 {
		scope.values[fn.Recv.List[0].Names[0].Name] = ast.NewIdent(receiver)
	}
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncLit:
			return false
		case *ast.TypeSpec:
			if _, exists := scope.local[n.Name.Name]; exists {
				duplicates[n.Name.Name] = true
			}
			scope.local[n.Name.Name] = pipelineLocalType{expr: n.Type, alias: n.Assign.IsValid()}
		case *ast.ValueSpec:
			for i, name := range n.Names {
				variables[name.Name] = true
				if scope.values[name.Name] != nil {
					scope.reassigned[name.Name] = true
				}
				switch {
				case n.Type != nil:
					scope.values[name.Name] = n.Type
				case i < len(n.Values):
					scope.values[name.Name] = n.Values[i]
				case len(n.Values) == 1 && len(n.Names) > 1:
					scope.values[name.Name] = n.Values[0]
				}
			}
		case *ast.AssignStmt:
			for i, lhs := range n.Lhs {
				if name, ok := lhs.(*ast.Ident); ok && i < len(n.Rhs) {
					if n.Tok == token.DEFINE {
						variables[name.Name] = true
					}
					if scope.values[name.Name] != nil {
						scope.reassigned[name.Name] = true
					}
					if n.Tok == token.DEFINE {
						scope.values[name.Name] = n.Rhs[i]
					}
				}
			}
		}
		return true
	})
	if len(duplicates) != 0 {
		return nil, fmt.Errorf("unsupported shadowed local type %s (same name declared in multiple scopes)", strings.Join(sortedPipelineKeys(duplicates), ", "))
	}
	for _, name := range sortedPipelineKeys(scope.local) {
		if g.types[name] != nil {
			return nil, fmt.Errorf("unsupported shadowed local type %s (shadows a package type)", name)
		}
		if variables[name] {
			return nil, fmt.Errorf("unsupported shadowed local type %s (reused as a variable)", name)
		}
	}
	for _, name := range sortedPipelineKeys(variables) {
		if scope.local[name].expr != nil || g.types[name] != nil {
			return nil, fmt.Errorf("unsupported shadowed codec variable %s (reuses a type name)", name)
		}
	}
	return scope, nil
}

func (g *pipelineGo) resolveWire(expr ast.Expr, scope *pipelineCodecScope) (ast.Expr, error) {
	var resolve func(ast.Expr, map[string]bool) (ast.Expr, error)
	resolve = func(expr ast.Expr, seen map[string]bool) (ast.Expr, error) {
		switch e := pipelineUnparen(expr).(type) {
		case *ast.UnaryExpr:
			if e.Op == token.AND {
				return resolve(e.X, seen)
			}
		case *ast.CompositeLit:
			return e.Type, nil
		case *ast.Ident:
			if scope.reassigned[e.Name] {
				return nil, fmt.Errorf("reassigned or shadowed wire variable %s", e.Name)
			}
			if value := scope.values[e.Name]; value != nil {
				if seen[e.Name] {
					return nil, fmt.Errorf("wire variable cycle at %s", e.Name)
				}
				seen[e.Name] = true
				return resolve(value, seen)
			}
			if g.types[e.Name] != nil || scope.local[e.Name].expr != nil {
				return e, nil
			}
		case *ast.CallExpr:
			if name, ok := e.Fun.(*ast.Ident); ok {
				if name.Name == "new" && len(e.Args) == 1 {
					return resolve(e.Args[0], seen)
				}
				if g.types[name.Name] != nil || scope.local[name.Name].expr != nil {
					return name, nil
				}
			}
		case *ast.StarExpr:
			return resolve(e.X, seen)
		case *ast.StructType:
			return e, nil
		}
		return nil, fmt.Errorf("unsupported wire expression %T", expr)
	}
	return resolve(expr, map[string]bool{})
}

// codecFields traces the actual argument passed to the JSON helper. Merely
// declaring an unrelated tagged struct inside a codec does not add coverage.
func (g *pipelineGo) codecFields(receiver, codec string, fn *ast.FuncDecl) (pipelineFieldSet, error) {
	scope, err := g.codecScope(receiver, fn)
	if err != nil {
		return nil, fmt.Errorf("%s: unsupported %s.%s codec: %w", g.fset.Position(fn.Pos()), receiver, codec, err)
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
		expr, err := g.resolveWire(call.Args[index], scope)
		if err != nil {
			problem = err
			return false
		}
		if wireType := pipelineTypeName(expr); wireType != "" && g.hasCodec(wireType, scope.local) {
			problem = fmt.Errorf("unsupported wire type %s with a custom JSON codec", wireType)
			return false
		}
		fields, err := g.rawFields(expr, scope.local, map[string]bool{})
		if err != nil {
			problem = err
			return false
		}
		if wire != nil && !samePipelineFields(wire, fields) {
			problem = fmt.Errorf("multiple JSON wire DTOs with different fields")
			return false
		}
		wire = fields
		// Every MarshalJSON return must use the same DTO; unrelated helper
		// calls and calls in unused closures are irrelevant.
		return true
	})
	if problem == nil && wire == nil {
		problem = fmt.Errorf("no supported JSON helper call")
	}
	if problem == nil && codec == "UnmarshalJSON" && !pipelinePointerReceiver(fn) {
		problem = fmt.Errorf("UnmarshalJSON must use a pointer receiver")
	}
	if problem == nil && codec == "UnmarshalJSON" && !codecWritesReceiver(fn) {
		problem = fmt.Errorf("UnmarshalJSON never assigns its receiver")
	}
	if problem != nil {
		return nil, fmt.Errorf("%s: unsupported %s.%s codec: %w", g.fset.Position(fn.Pos()), receiver, codec, problem)
	}
	return wire, nil
}

// codecWritesReceiver reports that a decoder assigns or decodes into its named
// receiver. A codec that only probes a temporary never decodes its value.
func codecWritesReceiver(fn *ast.FuncDecl) bool {
	if fn.Recv == nil || len(fn.Recv.List) == 0 || len(fn.Recv.List[0].Names) == 0 {
		return true
	}
	name := fn.Recv.List[0].Names[0].Name
	found := false
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if found {
			return false
		}
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}
		switch n := node.(type) {
		case *ast.AssignStmt:
			for _, lhs := range n.Lhs {
				if pipelineReceiverTarget(lhs, name) {
					found = true
					return false
				}
			}
		case *ast.CallExpr:
			for _, arg := range n.Args {
				if pipelineReceiverAddress(arg, name) {
					found = true
					return false
				}
			}
		}
		return true
	})
	return found
}

func pipelineReceiverTarget(expr ast.Expr, name string) bool {
	switch e := pipelineUnparen(expr).(type) {
	case *ast.Ident:
		return e.Name == name
	case *ast.StarExpr:
		id, ok := pipelineUnparen(e.X).(*ast.Ident)
		return ok && id.Name == name
	case *ast.SelectorExpr:
		return pipelineReceiverRef(e.X, name)
	}
	return false
}

func pipelineReceiverAddress(expr ast.Expr, name string) bool {
	unary, ok := pipelineUnparen(expr).(*ast.UnaryExpr)
	if !ok || unary.Op != token.AND {
		return false
	}
	return pipelineReceiverTarget(unary.X, name)
}

// envelopeDecodeFields extracts the wire DTOs decoded by the envelope's
// UnmarshalJSON. Every resolvable wire DTO must agree, and a DTO that owns a
// custom codec is rejected. Calls whose target is not a traceable wire value
// (for example an embedded field) are handled by the type codecs instead.
func (g *pipelineGo) envelopeDecodeFields(receiver string, fn *ast.FuncDecl) (pipelineFieldSet, error) {
	fail := func(err error) (pipelineFieldSet, error) {
		return nil, fmt.Errorf("%s: unsupported %s.UnmarshalJSON codec: %w", g.fset.Position(fn.Pos()), receiver, err)
	}
	scope, err := g.codecScope(receiver, fn)
	if err != nil {
		return fail(err)
	}
	if !pipelinePointerReceiver(fn) {
		return fail(fmt.Errorf("UnmarshalJSON must use a pointer receiver"))
	}
	if !codecWritesReceiver(fn) {
		return fail(fmt.Errorf("UnmarshalJSON never assigns its receiver"))
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
		call, ok := node.(*ast.CallExpr)
		if !ok || pipelineDecodeCallName(call) == "" || len(call.Args) < 2 {
			return true
		}
		expr, err := g.resolveWire(call.Args[1], scope)
		if err != nil {
			// Untraceable targets such as embedded fields are decoded by their
			// own codecs and are validated separately.
			return true
		}
		if wireType := pipelineTypeName(expr); wireType != "" && g.hasCodec(wireType, scope.local) {
			problem = fmt.Errorf("unsupported wire type %s with a custom JSON codec", wireType)
			return false
		}
		fields, err := g.rawFields(expr, scope.local, map[string]bool{})
		if err != nil {
			problem = err
			return false
		}
		if wire != nil && !samePipelineFields(wire, fields) {
			problem = fmt.Errorf("multiple JSON wire DTOs with different fields")
			return false
		}
		wire = fields
		return true
	})
	if problem != nil {
		return fail(problem)
	}
	if wire == nil {
		return fail(fmt.Errorf("no supported JSON helper call"))
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
