package checker

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"runtime"
)

// goConstDecl is one constant declaration successfully evaluated by the
// synthetic-package type checker.
type goConstDecl struct {
	name  string
	expr  string
	value constant.Value
	kind  numericKind
	bits  int
}

// goConstEvalFailure describes one constant declaration the evaluator rejected.
type goConstEvalFailure struct {
	name string
	expr string
	err  string
}

// goConstEvaluation is the evaluation result for every constant declaration in
// one Go source file.
type goConstEvaluation struct {
	values   map[string]goConstDecl
	failures map[string]goConstEvalFailure
}

// evaluateGoConstFile evaluates all constants declared in a Go source file.
func evaluateGoConstFile(path string) (*goConstEvaluation, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return evaluateGoConstSource(path, src)
}

// evaluateGoConstSource evaluates all constants declared in Go source.
//
// It parses the file, copies only its type and const declarations into a
// synthetic package with no imports, and type-checks that package with
// go/types. Evaluating on top of the type checker preserves Go semantics that
// a plain expression walker cannot: declared widths and signedness,
// conversions such as uint64(^uint32(0)), iota and inherited expressions,
// constant aliases, and forward dependencies between declarations.
//
// Constants that depend on anything outside the synthetic package (imported
// symbols, function calls, runtime values) fail type checking and are returned
// in failures instead of being approximated.
func evaluateGoConstSource(path string, src []byte) (*goConstEvaluation, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	var synthetic bytes.Buffer
	synthetic.WriteString("package synthetic\n\n")
	included := 0
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || (gen.Tok != token.CONST && gen.Tok != token.TYPE) {
			continue
		}
		start := fset.Position(gen.Pos()).Offset
		end := fset.Position(gen.End()).Offset
		if start < 0 || end > len(src) || start >= end {
			continue
		}
		synthetic.Write(src[start:end])
		synthetic.WriteString("\n\n")
		included++
	}

	eval := &goConstEvaluation{
		values:   map[string]goConstDecl{},
		failures: map[string]goConstEvalFailure{},
	}
	if included == 0 {
		return eval, nil
	}

	synFset := token.NewFileSet()
	synFile, err := parser.ParseFile(synFset, "synthetic_constants.go", synthetic.Bytes(), 0)
	if err != nil {
		return nil, fmt.Errorf("parse synthetic constants for %s: %w", path, err)
	}

	sizes := types.SizesFor("gc", runtime.GOARCH)
	if sizes == nil {
		sizes = &types.StdSizes{WordSize: 8, MaxAlign: 8}
	}

	info := &types.Info{Defs: map[*ast.Ident]types.Object{}}
	typeErrs := make([]types.Error, 0)
	conf := types.Config{
		Sizes: sizes,
		Error: func(err error) {
			var typeErr types.Error
			if errors.As(err, &typeErr) {
				typeErrs = append(typeErrs, typeErr)
			}
		},
	}
	// Errors are collected through the Error callback; the returned error only
	// summarizes them and is intentionally ignored.
	_, _ = conf.Check("synthetic", synFset, []*ast.File{synFile}, info)

	specs := collectSyntheticConstSpecs(synFset, synFile)
	attributeGoConstErrors(synFset, specs, typeErrs, eval)
	collectGoConstValues(specs, info, sizes, eval)

	return eval, nil
}

// syntheticConstSpec is one ValueSpec of the synthetic const declarations.
type syntheticConstSpec struct {
	names  []string
	idents []*ast.Ident
	exprs  []string
	start  int
	end    int
}

func collectSyntheticConstSpecs(fset *token.FileSet, file *ast.File) []*syntheticConstSpec {
	specs := make([]*syntheticConstSpec, 0)
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}
		for _, spec := range gen.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			syntheticSpec := &syntheticConstSpec{
				start: fset.Position(valueSpec.Pos()).Offset,
				end:   fset.Position(valueSpec.End()).Offset,
			}
			for i, ident := range valueSpec.Names {
				if ident == nil {
					continue
				}
				syntheticSpec.names = append(syntheticSpec.names, ident.Name)
				syntheticSpec.idents = append(syntheticSpec.idents, ident)
				syntheticSpec.exprs = append(syntheticSpec.exprs, formatGoConstValueExpr(valueSpec, i))
			}
			specs = append(specs, syntheticSpec)
		}
	}
	return specs
}

func formatGoConstValueExpr(spec *ast.ValueSpec, index int) string {
	if len(spec.Values) == 0 {
		return "<inherited>"
	}
	if index >= len(spec.Values) {
		index = len(spec.Values) - 1
	}
	return formatGoExpr(spec.Values[index])
}

// attributeGoConstErrors records a type-check error against the const spec that
// contains it. Errors outside a const spec (for example in an unrelated type
// declaration) are ignored here and handled later through invalid constant
// types.
func attributeGoConstErrors(fset *token.FileSet, specs []*syntheticConstSpec, typeErrs []types.Error, eval *goConstEvaluation) {
	for _, typeErr := range typeErrs {
		if !typeErr.Pos.IsValid() {
			continue
		}
		offset := fset.Position(typeErr.Pos).Offset
		for _, spec := range specs {
			if offset < spec.start || offset >= spec.end {
				continue
			}
			for i, name := range spec.names {
				if name == "_" {
					continue
				}
				if _, exists := eval.failures[name]; exists {
					continue
				}
				eval.failures[name] = goConstEvalFailure{
					name: name,
					expr: spec.exprs[i],
					err:  typeErr.Msg,
				}
			}
			break
		}
	}
}

func collectGoConstValues(specs []*syntheticConstSpec, info *types.Info, sizes types.Sizes, eval *goConstEvaluation) {
	for _, spec := range specs {
		for i, ident := range spec.idents {
			name := spec.names[i]
			if name == "_" {
				continue
			}
			if _, failed := eval.failures[name]; failed {
				continue
			}
			failure := func(message string) {
				eval.failures[name] = goConstEvalFailure{
					name: name,
					expr: spec.exprs[i],
					err:  message,
				}
			}
			object, ok := info.Defs[ident]
			if !ok {
				failure("type checker did not produce a constant")
				continue
			}
			constantObject, ok := object.(*types.Const)
			if !ok {
				failure("declaration is not a constant")
				continue
			}
			value, err := numericValueForGoConst(constantObject, sizes)
			if err != nil {
				failure(err.Error())
				continue
			}
			eval.values[name] = goConstDecl{
				name:  name,
				expr:  spec.exprs[i],
				value: value.val,
				kind:  value.kind,
				bits:  value.bits,
			}
		}
	}
}

// numericValueForGoConst converts a type-checked Go constant into a numeric
// value that keeps the declared signedness and width.
func numericValueForGoConst(constantObject *types.Const, sizes types.Sizes) (numericValue, error) {
	basic, ok := constantObject.Type().Underlying().(*types.Basic)
	if !ok {
		return numericValue{}, fmt.Errorf("unsupported constant type: %s", constantObject.Type())
	}
	if basic.Kind() == types.Invalid {
		return numericValue{}, fmt.Errorf("type checking failed")
	}
	value := constant.ToInt(constantObject.Val())
	if basic.Info()&types.IsUntyped != 0 {
		if value.Kind() != constant.Int {
			return numericValue{}, fmt.Errorf("not an integer constant: %s", constantObject.Val().ExactString())
		}
		return numericValue{kind: numericUntyped, val: value}, nil
	}
	if basic.Info()&types.IsInteger == 0 {
		return numericValue{}, fmt.Errorf("unsupported constant type: %s", constantObject.Type())
	}
	if value.Kind() != constant.Int {
		return numericValue{}, fmt.Errorf("not an integer constant: %s", constantObject.Val().ExactString())
	}
	bits := int(sizes.Sizeof(basic)) * 8
	if basic.Info()&types.IsUnsigned != 0 {
		return numericValue{kind: numericUnsigned, bits: bits, val: value}, nil
	}
	return numericValue{kind: numericSigned, bits: bits, val: value}, nil
}
