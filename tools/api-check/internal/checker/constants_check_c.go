package checker

import (
	"errors"
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"regexp"
	"strconv"
	"strings"
)

// cConstDecl is one raw C integer constant declaration: an object-like macro
// or an enum member.
type cConstDecl struct {
	name string
	expr string
}

// cConstSource is one header file handed to the C constant evaluator.
type cConstSource struct {
	path    string
	content string
}

// cConstEnv holds every macro and enum constant found in the configured
// headers, evaluated with exact integer arithmetic.
type cConstEnv struct {
	decls       map[string]cConstDecl
	values      map[string]constant.Value
	typedValues map[string]numericValue
	failures    map[string]string
	types       map[string]cTypeAlias
}

// evaluateCConstSources parses object-like #define macros and C-style enum
// members from the given headers, then evaluates them as one dependency graph.
//
// Header helpers are resolved globally, so a macro may reference another macro
// or enum member declared earlier, later, or in a different configured header.
// Expressions that stay unresolved (unknown identifier, dependency cycle,
// division by zero, unsupported call) are reported per name through failures.
func evaluateCConstSources(sources []cConstSource) (*cConstEnv, error) {
	env := &cConstEnv{
		decls:       map[string]cConstDecl{},
		values:      map[string]constant.Value{},
		typedValues: map[string]numericValue{},
		failures:    map[string]string{},
	}
	rawTypes := map[string]string{}

	for _, source := range sources {
		stripped := removeCComments(source.content)

		for name, expr := range parseCDefineExprs(stripped) {
			if prior, exists := env.decls[name]; exists {
				if normalizeSpaces(prior.expr) != normalizeSpaces(expr) {
					return nil, fmt.Errorf("conflicting C constant %s in %s", name, source.path)
				}
				continue
			}
			env.decls[name] = cConstDecl{name: name, expr: expr}
		}

		enumDecls, err := parseCEnumDecls(stripped)
		if err != nil {
			return nil, fmt.Errorf("parse C enums in %s: %w", source.path, err)
		}
		for _, decl := range enumDecls {
			if prior, exists := env.decls[decl.name]; exists {
				if normalizeSpaces(prior.expr) != normalizeSpaces(decl.expr) {
					return nil, fmt.Errorf("conflicting C constant %s in %s", decl.name, source.path)
				}
				continue
			}
			env.decls[decl.name] = decl
		}

		mergeCTypedefs(rawTypes, parseCTypedefs(stripped))
	}

	env.types = resolveCTypedefs(rawTypes)
	env.evaluate()
	return env, nil
}

func (env *cConstEnv) evaluate() {
	pending := make(map[string]cConstDecl, len(env.decls))
	for name, decl := range env.decls {
		pending[name] = decl
	}

	for len(pending) > 0 {
		progress := false
		for name, decl := range pending {
			expr, err := parseCConstExpr(decl.expr)
			if err != nil {
				env.failures[name] = err.Error()
				delete(pending, name)
				progress = true
				continue
			}
			value, err := evalCIntExpr(expr, env.typedValues, env.types)
			if err != nil {
				var unknownErr *unknownIdentifierError
				if errors.As(err, &unknownErr) {
					continue
				}
				env.failures[name] = err.Error()
				delete(pending, name)
				progress = true
				continue
			}
			env.values[name] = value.val
			env.typedValues[name] = value
			delete(pending, name)
			progress = true
		}
		if progress {
			continue
		}

		for name, decl := range pending {
			expr, parseErr := parseCConstExpr(decl.expr)
			if parseErr != nil {
				env.failures[name] = parseErr.Error()
				continue
			}
			if _, err := evalCIntExpr(expr, env.typedValues, env.types); err != nil {
				env.failures[name] = err.Error()
			}
		}
		break
	}
}

// parseCConstExpr converts a C integer constant expression into a Go
// expression. Only the shapes used by MaaDef.h and MaaToolkitDef.h are
// supported; anything else fails parsing or evaluation explicitly.
func parseCConstExpr(expr string) (ast.Expr, error) {
	normalized := cMethodIntSuffixRe.ReplaceAllStringFunc(expr, func(literal string) string {
		match := cMethodIntSuffixRe.FindStringSubmatch(literal)
		suffix := strings.ToLower(literal[len(match[1]):])
		return "__maa_literal_" + suffix + "(" + match[1] + ")"
	})
	node, err := parser.ParseExpr(normalized)
	if err != nil {
		return nil, fmt.Errorf("parse C expression %q: %w", expr, err)
	}
	return node, nil
}

// cInteger literals and operations use a fixed 32-bit int and 64-bit long
// long. Explicit long suffixes are rejected because long is platform dependent.
func cInteger(kind numericKind, bits int, value constant.Value) (numericValue, error) {
	result := numericValue{kind: kind, bits: bits, val: value}
	normalized, err := normalizeCNumericValue(result)
	if err != nil {
		return numericValue{}, err
	}
	result.val = normalized
	return result, nil
}

func cInt(value int64) numericValue {
	return numericValue{kind: numericSigned, bits: 32, val: constant.MakeInt64(value)}
}

func cLiteral(node *ast.BasicLit, suffix string) (numericValue, error) {
	if node.Kind != token.INT && node.Kind != token.CHAR {
		return numericValue{}, fmt.Errorf("unsupported literal kind: %s", node.Kind)
	}
	value := constant.MakeFromLiteral(node.Value, node.Kind, 0)
	if value.Kind() != constant.Int {
		return numericValue{}, fmt.Errorf("unsupported integer literal: %s", node.Value)
	}
	if node.Kind == token.CHAR {
		return cInteger(numericSigned, 32, value)
	}
	var candidates []cTypeAlias
	switch suffix {
	case "":
		candidates = []cTypeAlias{{numericSigned, 32}, {numericSigned, 64}}
		if strings.HasPrefix(node.Value, "0") && len(node.Value) > 1 {
			candidates = []cTypeAlias{{numericSigned, 32}, {numericUnsigned, 32}, {numericSigned, 64}, {numericUnsigned, 64}}
		}
	case "u":
		candidates = []cTypeAlias{{numericUnsigned, 32}, {numericUnsigned, 64}}
	case "ll":
		candidates = []cTypeAlias{{numericSigned, 64}}
		if strings.HasPrefix(node.Value, "0") && len(node.Value) > 1 {
			candidates = append(candidates, cTypeAlias{numericUnsigned, 64})
		}
	case "ull", "llu":
		candidates = []cTypeAlias{{numericUnsigned, 64}}
	case "l", "ul", "lu":
		return numericValue{}, fmt.Errorf("platform-dependent C long literal is unsupported: %s%s", node.Value, suffix)
	default:
		return numericValue{}, fmt.Errorf("unsupported C literal suffix: %s", suffix)
	}
	for _, shape := range candidates {
		result := shape.numericValue(value)
		if normalized, err := normalizeNumericValue(result); err == nil {
			result.val = normalized
			return result, nil
		}
	}
	return numericValue{}, fmt.Errorf("integer literal is outside supported C ranges: %s%s", node.Value, suffix)
}

// cPromote performs integer promotion before unary and binary operations.
func cPromote(value numericValue) numericValue {
	if value.bits < 32 {
		value.kind, value.bits = numericSigned, 32
	}
	return value
}

func cCommonOperands(left, right numericValue) (numericValue, numericValue, error) {
	left, right = cPromote(left), cPromote(right)
	kind, bits := left.kind, left.bits
	if left.kind == right.kind {
		bits = max(left.bits, right.bits)
	} else {
		signed, unsigned := left, right
		if signed.kind != numericSigned {
			signed, unsigned = right, left
		}
		if unsigned.bits >= signed.bits {
			kind, bits = numericUnsigned, unsigned.bits
		} else {
			kind, bits = numericSigned, signed.bits
		}
	}
	left, err := cInteger(kind, bits, left.val)
	if err != nil {
		return numericValue{}, numericValue{}, err
	}
	right, err = cInteger(kind, bits, right.val)
	return left, right, err
}

func evalCIntExpr(expr ast.Expr, env map[string]numericValue, typeEnv map[string]cTypeAlias) (numericValue, error) {
	switch node := expr.(type) {
	case *ast.ParenExpr:
		return evalCIntExpr(node.X, env, typeEnv)
	case *ast.BasicLit:
		return cLiteral(node, "")
	case *ast.Ident:
		if value, ok := env[node.Name]; ok {
			return value, nil
		}
		return numericValue{}, &unknownIdentifierError{name: node.Name}
	case *ast.UnaryExpr:
		operand, err := evalCIntExpr(node.X, env, typeEnv)
		if err != nil {
			return numericValue{}, err
		}
		operand = cPromote(operand)
		switch node.Op {
		case token.ADD:
			return operand, nil
		case token.SUB:
			return cInteger(operand.kind, operand.bits, constant.UnaryOp(token.SUB, operand.val, 0))
		case token.XOR, token.TILDE:
			return cInteger(operand.kind, operand.bits, constant.UnaryOp(token.XOR, operand.val, 0))
		case token.NOT:
			if constant.Sign(operand.val) == 0 {
				return cInt(1), nil
			}
			return cInt(0), nil
		default:
			return numericValue{}, fmt.Errorf("unsupported unary operator: %s", node.Op)
		}
	case *ast.BinaryExpr:
		left, err := evalCIntExpr(node.X, env, typeEnv)
		if err != nil {
			return numericValue{}, err
		}
		if node.Op == token.LAND && constant.Sign(left.val) == 0 {
			return cInt(0), nil
		}
		if node.Op == token.LOR && constant.Sign(left.val) != 0 {
			return cInt(1), nil
		}
		right, err := evalCIntExpr(node.Y, env, typeEnv)
		if err != nil {
			return numericValue{}, err
		}
		if node.Op == token.SHL || node.Op == token.SHR {
			left = cPromote(left)
			shift, ok := constant.Uint64Val(right.val)
			if !ok || shift >= uint64(left.bits) {
				return numericValue{}, fmt.Errorf("undefined C shift count: %s for %d-bit value", right.val.ExactString(), left.bits)
			}
			if left.kind == numericSigned && constant.Sign(left.val) < 0 {
				return numericValue{}, fmt.Errorf("unsupported C shift of negative signed value")
			}
			return cInteger(left.kind, left.bits, constant.Shift(left.val, node.Op, uint(shift)))
		}
		if node.Op == token.LAND || node.Op == token.LOR {
			if constant.Sign(right.val) != 0 {
				return cInt(1), nil
			}
			return cInt(0), nil
		}
		left, right, err = cCommonOperands(left, right)
		if err != nil {
			return numericValue{}, err
		}
		switch node.Op {
		case token.ADD, token.SUB, token.MUL, token.AND, token.OR, token.XOR:
			return cInteger(left.kind, left.bits, constant.BinaryOp(left.val, node.Op, right.val))
		case token.QUO, token.REM:
			if constant.Sign(right.val) == 0 {
				operation := "division"
				if node.Op == token.REM {
					operation = "modulo"
				}
				return numericValue{}, fmt.Errorf("%s by zero", operation)
			}
			op := node.Op
			if op == token.QUO {
				op = token.QUO_ASSIGN
			}
			return cInteger(left.kind, left.bits, constant.BinaryOp(left.val, op, right.val))
		case token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ:
			if constant.Compare(left.val, node.Op, right.val) {
				return cInt(1), nil
			}
			return cInt(0), nil
		default:
			return numericValue{}, fmt.Errorf("unsupported binary operator: %s", node.Op)
		}
	case *ast.CallExpr:
		return evalCConstCast(node, env, typeEnv)
	default:
		return numericValue{}, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// evalCConstCast handles explicit casts and the literal suffix markers inserted
// by parseCConstExpr. Integer conversions retain their width for later operators.
func evalCConstCast(node *ast.CallExpr, env map[string]numericValue, typeEnv map[string]cTypeAlias) (numericValue, error) {
	fun := node.Fun
	if paren, ok := fun.(*ast.ParenExpr); ok {
		fun = paren.X
	}
	ident, ok := fun.(*ast.Ident)
	if !ok || len(node.Args) != 1 {
		return numericValue{}, fmt.Errorf("unsupported call expression: %s", formatGoExpr(node))
	}
	if suffix, ok := strings.CutPrefix(ident.Name, "__maa_literal_"); ok {
		literal, ok := node.Args[0].(*ast.BasicLit)
		if !ok {
			return numericValue{}, fmt.Errorf("invalid C literal marker: %s", formatGoExpr(node))
		}
		return cLiteral(literal, suffix)
	}
	target, ok := typeEnv[ident.Name]
	if !ok {
		return numericValue{}, fmt.Errorf("unsupported call expression: %s", ident.Name)
	}
	value, err := evalCIntExpr(node.Args[0], env, typeEnv)
	if err != nil {
		return numericValue{}, err
	}
	return cInteger(target.kind, target.bits, value.val)
}

var cEnumKeywordRe = regexp.MustCompile(`\benum\b`)
var cConstTypedefRe = regexp.MustCompile(`(?m)\btypedef\s+([A-Za-z_][A-Za-z0-9_]*)\s+([A-Za-z_][A-Za-z0-9_]*)\s*;`)
var cConstIdentifierRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// parseCEnumDecls extracts C-style enum members. Members without an explicit
// value inherit the previous member value.
func parseCEnumDecls(content string) ([]cConstDecl, error) {
	decls := make([]cConstDecl, 0)
	searchFrom := 0
	for {
		loc := cEnumKeywordRe.FindStringIndex(content[searchFrom:])
		if loc == nil {
			break
		}
		enumStart := searchFrom + loc[0]
		rest := content[enumStart:]
		bodyStart := strings.IndexByte(rest, '{')
		semi := strings.IndexByte(rest, ';')
		if bodyStart < 0 {
			break
		}
		if semi >= 0 && semi < bodyStart {
			// Forward declaration such as "enum Foo;"; skip it.
			searchFrom = enumStart + semi + 1
			continue
		}
		openPos := enumStart + bodyStart
		closePos := matchCBrace(content, openPos)
		if closePos < 0 {
			return nil, errors.New("unterminated enum body")
		}

		previous := ""
		for _, entry := range splitCEnumEntries(content, openPos+1, closePos) {
			name, expr := splitCEnumEntry(entry)
			if name == "" {
				return nil, fmt.Errorf("unsupported enum member: %q", entry)
			}
			if expr == "" {
				if strings.Contains(entry, "=") {
					return nil, fmt.Errorf("missing enum initializer: %s", name)
				}
				if previous == "" {
					expr = "0"
				} else {
					expr = "(" + previous + " + 1)"
				}
			}
			decls = append(decls, cConstDecl{name: name, expr: expr})
			previous = name
		}
		searchFrom = closePos + 1
	}
	return decls, nil
}

func matchCBrace(content string, openPos int) int {
	depth := 0
	for i := openPos; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func splitCEnumEntries(content string, start int, end int) []string {
	entries := make([]string, 0)
	depth := 0
	segmentStart := start
	flush := func(segmentEnd int) {
		trimmed := strings.TrimSpace(content[segmentStart:segmentEnd])
		if trimmed == "" {
			return
		}
		entries = append(entries, trimmed)
	}
	for i := start; i < end; i++ {
		switch content[i] {
		case '(', '[':
			depth++
		case ')', ']':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				flush(i)
				segmentStart = i + 1
			}
		}
	}
	flush(end)
	return entries
}

func splitCEnumEntry(text string) (string, string) {
	depth := 0
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '(', '[':
			depth++
		case ')', ']':
			if depth > 0 {
				depth--
			}
		case '=':
			if depth == 0 {
				name := strings.TrimSpace(text[:i])
				if !cConstIdentifierRe.MatchString(name) {
					return "", ""
				}
				return name, strings.TrimSpace(text[i+1:])
			}
		}
	}
	name := strings.TrimSpace(text)
	if !cConstIdentifierRe.MatchString(name) {
		return "", ""
	}
	return name, ""
}

func parseCTypedefs(content string) map[string]string {
	aliases := map[string]string{}
	for _, match := range cConstTypedefRe.FindAllStringSubmatch(content, -1) {
		name := match[2]
		base := match[1]
		if name == base {
			continue
		}
		if _, exists := aliases[name]; !exists {
			aliases[name] = base
		}
	}
	return aliases
}

func mergeCTypedefs(dst map[string]string, src map[string]string) {
	for name, base := range src {
		if _, exists := dst[name]; !exists {
			dst[name] = base
		}
	}
}

// resolveCTypedefs turns typedef aliases into numeric widths and signedness.
func resolveCTypedefs(aliases map[string]string) map[string]cTypeAlias {
	base := cBaseIntegerTypes()
	resolved := make(map[string]cTypeAlias, len(base)+len(aliases))
	for name, shape := range base {
		resolved[name] = shape
	}

	for changed := true; changed; {
		changed = false
		for name, target := range aliases {
			if _, done := resolved[name]; done {
				continue
			}
			if alias, ok := resolved[target]; ok {
				resolved[name] = alias
				changed = true
				continue
			}
			if alias, ok := base[target]; ok {
				resolved[name] = alias
				changed = true
			}
		}
	}
	return resolved
}

// cTypeAlias is the numeric shape of one C typedef.
type cTypeAlias struct {
	kind numericKind
	bits int
}

func (alias cTypeAlias) numericValue(value constant.Value) numericValue {
	return numericValue{kind: alias.kind, bits: alias.bits, val: value}
}

func cBaseIntegerTypes() map[string]cTypeAlias {
	wordBits := strconv.IntSize
	return map[string]cTypeAlias{
		"uint8_t":   {kind: numericUnsigned, bits: 8},
		"uint16_t":  {kind: numericUnsigned, bits: 16},
		"uint32_t":  {kind: numericUnsigned, bits: 32},
		"uint64_t":  {kind: numericUnsigned, bits: 64},
		"int8_t":    {kind: numericSigned, bits: 8},
		"int16_t":   {kind: numericSigned, bits: 16},
		"int32_t":   {kind: numericSigned, bits: 32},
		"int64_t":   {kind: numericSigned, bits: 64},
		"size_t":    {kind: numericUnsigned, bits: wordBits},
		"uintptr_t": {kind: numericUnsigned, bits: wordBits},
		"intptr_t":  {kind: numericSigned, bits: wordBits},
		"ssize_t":   {kind: numericSigned, bits: wordBits},
		"char":      {kind: numericSigned, bits: 8},
		"short":     {kind: numericSigned, bits: 16},
		"int":       {kind: numericSigned, bits: 32},
		"unsigned":  {kind: numericUnsigned, bits: 32},
	}
}
