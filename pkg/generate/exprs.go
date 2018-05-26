package generate

import (
	"go/ast"
	"go/token"
)

//go:generate go run $GOPATH/src/github.com/directxman12/envmap/cmd/buildergen/main.go -- $GOFILE

func Ident(name string) Expr {
	return &expr{ast.NewIdent(name)}
}

func QualifiedIdent(pkg, name string) Expr {
	return &expr{&ast.SelectorExpr{X: ast.NewIdent(pkg), Sel: ast.NewIdent(name)}}
}

func ReceiveFromChan(ch Expr) Expr {
	return &expr{&ast.UnaryExpr{Op: token.ARROW, X: ch.RenderExpr()}}
}

func LiteralString(fullValue string) Expr {
	return &expr{&ast.BasicLit{Kind: token.STRING, Value: fullValue}}
}
func LiteralInt(fullValue string) Expr {
	return &expr{&ast.BasicLit{Kind: token.INT, Value: fullValue}}
}
func LiteralFloat(fullValue string) Expr {
	return &expr{&ast.BasicLit{Kind: token.FLOAT, Value: fullValue}}
}
func LiteralImaginary(fullValue string) Expr {
	return &expr{&ast.BasicLit{Kind: token.IMAG, Value: fullValue}}
}
func LiteralChar(fullValue string) Expr {
	return &expr{&ast.BasicLit{Kind: token.CHAR, Value: fullValue}}
}
// TODO: helpers for literals

// Func literals are created by calling `WithBody` on FunctionDefinitionBuilder

//+envmap:render=lit,ast.Expr
type CompositeLiteralBuilder struct {
	lit *ast.CompositeLit
}
func InitializerFor(typ TypeDefinition) *CompositeLiteralBuilder {
	return &CompositeLiteralBuilder{
		lit: &ast.CompositeLit{Type: typ.RenderExpr()},
	}
}
func (b *CompositeLiteralBuilder) WithKeyValue(k, v Expr) *CompositeLiteralBuilder {
	b.lit.Elts = append(b.lit.Elts, &ast.KeyValueExpr{
		Key: k.RenderExpr(),
		Value: v.RenderExpr(),
	})
	return b
}
func (b *CompositeLiteralBuilder) WithValue(v Expr) *CompositeLiteralBuilder {
	b.lit.Elts = append(b.lit.Elts, v.RenderExpr())
	return b
}
func (b *CompositeLiteralBuilder) RenderStmt() ast.Stmt {
	return &ast.ExprStmt{X: b.RenderExpr()}
}

// TODO: just make these part of the Expr interface and autogenerate them?

func InParens(e Expr) Expr {
	return &expr{&ast.ParenExpr{X: e.RenderExpr()}}
}

func FieldOf(e Expr, sel string) Expr {
	return &expr{&ast.SelectorExpr{X: e.RenderExpr(), Sel: ast.NewIdent(sel)}}
}

func AtIndex(e, ind Expr) Expr {
	return &expr{&ast.IndexExpr{X: e.RenderExpr(), Index: ind.RenderExpr()}}
}

func SlicedAs(e, start, end, max Expr) Expr {
	var startExpr, endExpr, maxExpr ast.Expr
	if start != nil {
		startExpr = start.RenderExpr()
	}
	if end != nil {
		endExpr = end.RenderExpr()
	}
	if max != nil {
		maxExpr = max.RenderExpr()
	}
	return &expr{
		&ast.SliceExpr{
			X: e.RenderExpr(),
			Low: startExpr,
			High: endExpr,
			Max: maxExpr,
			Slice3: maxExpr != nil,
		},
	}
}

func AssertType(e Expr, typ TypeDefinition) Expr {
	var typeExpr ast.Expr
	if typ != nil {
		typeExpr = typ.RenderExpr()
	}
	return &expr{&ast.TypeAssertExpr{X: e.RenderExpr(), Type: typeExpr}}
}

func CallFunc(f Expr, args ...Expr) CallExpr {
	// TODO: do we need to manually set the ellipsis position?
	res := RawCallExpr(ast.CallExpr{Fun: f.RenderExpr(), Args: exprs(args)})
	return &res
}

func ReferenceTo(e Expr) Expr {
	return &expr{&ast.UnaryExpr{Op: token.AND, X: e.RenderExpr()}}
}

func Dereference(e Expr) Expr {
	return &expr{&ast.StarExpr{X: e.RenderExpr()}}
}

func UnaryExpr(op string, e Expr) Expr {
	tok := token.ILLEGAL
	switch op {
	case "+":
		tok = token.ADD
	case "-":
		tok = token.SUB
	case "!":
		tok = token.NOT
	case "^":
		tok = token.XOR
	}
	return &expr{&ast.UnaryExpr{Op: tok, X: e.RenderExpr()}}
}

func BinaryExpr(lhs Expr, op string, rhs Expr) Expr {
	tok := token.ILLEGAL
	switch op {
	case "||":
		tok = token.LOR
	case "&&":
		tok = token.LAND
	case "==":
		tok = token.EQL
	case "!=":
		tok = token.NEQ
	case "<":
		tok = token.LSS
	case "<=":
		tok = token.LEQ
	case ">":
		tok = token.GTR
	case ">=":
		tok = token.GEQ
	case "+":
		tok = token.ADD
	case "-":
		tok = token.SUB
	case "|":
		tok = token.OR
	case "^":
		tok = token.XOR
	case "*":
		tok = token.MUL
	case "/":
		tok = token.QUO
	case "%":
		tok = token.REM
	case "<<":
		tok = token.SHL
	case ">>":
		tok = token.SHR
	case "&":
		tok = token.AND
	case "&^":
		tok = token.AND_NOT
	}
	return &expr{
		&ast.BinaryExpr{
			Op: tok,
			X: lhs.RenderExpr(),
			Y: rhs.RenderExpr(),
		},
	}
}
