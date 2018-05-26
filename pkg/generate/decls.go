package generate

import (
	"go/ast"
	"go/token"
)

//go:generate go run $GOPATH/src/github.com/directxman12/envmap/cmd/buildergen/main.go -- $GOFILE

// TODO: with-comments?

//+envmap:with-docs=decl
//+envmap:render=decl,ast.Decl
type TypesDeclarationBuilder struct {
	decl *ast.GenDecl
}
func Types() *TypesDeclarationBuilder {
	decl := &ast.GenDecl{
		Tok: token.TYPE,
		Specs: []ast.Spec{},
	}
	return &TypesDeclarationBuilder{decl: decl}
}
func (b *TypesDeclarationBuilder) With(s TypeDeclaration) *TypesDeclarationBuilder {
	b.decl.Specs = append(b.decl.Specs, s.RenderTypeSpec())
	return b
}

// Type can also be used as a declaration, in which case it's
// equivalent to Types().With(Type(...))
//+envmap:with-docs=spec
//+envmap:render=spec,*ast.TypeSpec
type TypeDeclarationBuilder struct {
	spec *ast.TypeSpec
}

func Type(name string, d TypeDefinition) *TypeDeclarationBuilder {
	return &TypeDeclarationBuilder{
		&ast.TypeSpec{
			Name: ast.NewIdent(name),
			Type: d.RenderExpr(),
		},
	}
}
func (b *TypeDeclarationBuilder) AsAlias() *TypeDeclarationBuilder {
	b.spec.Assign = token.Pos(1)
	return b
}


func (b *TypeDeclarationBuilder) RenderDecl() ast.Decl {
	return Types().With(b).RenderDecl()
}

//+envmap:with-docs=decl
//+envmap:render=decl,*ast.FuncDecl
//+envmap:render=decl,ast.Decl
type FunctionDeclarationBuilder struct {
	decl *ast.FuncDecl
}

func (b *FunctionDeclarationBuilder) AsMethodFor(alias string, d TypeDefinition) *FunctionDeclarationBuilder {
	b.decl.Recv = &ast.FieldList{
		List: []*ast.Field{
			{
				Names: []*ast.Ident{ast.NewIdent(alias)},
				Type: d.RenderExpr(),
			},
		},
	}
	return b
}

func (b *FunctionDeclarationBuilder) WithBody(s BlockStatement) *FunctionDeclarationBuilder {
	b.decl.Body = s.RenderBlockStmt()
	return b
}

//+envmap:with-docs=decl
//+envmap:render=decl,ast.Decl
type ValuesDeclarationBuilder struct {
	decl *ast.GenDecl
}
func Vars() *ValuesDeclarationBuilder {
	decl := &ast.GenDecl{
		Tok: token.VAR,
		Specs: []ast.Spec{},
	}
	return &ValuesDeclarationBuilder{decl: decl}
}
func Consts() *ValuesDeclarationBuilder {
	decl := &ast.GenDecl{
		Tok: token.CONST,
		Specs: []ast.Spec{},
	}
	return &ValuesDeclarationBuilder{decl: decl}
}
func (b *ValuesDeclarationBuilder) With(s ValueDeclaration) *ValuesDeclarationBuilder {
	b.decl.Specs = append(b.decl.Specs, s.RenderValueSpec())
	return b
}

//+envmap:with-docs=spec
//+envmap:render=spec,*ast.ValueSpec
type ValueDeclarationBuilder struct {
	spec *ast.ValueSpec
}

func Value(name string, d TypeDefinition, v Expr) *ValueDeclarationBuilder {
	return Values(d).Set(name, v)
}
func Values(d TypeDefinition) *ValueDeclarationBuilder {
	var typ ast.Expr
	if d != nil {
		typ = d.RenderExpr()
	}
	return &ValueDeclarationBuilder{
		spec: &ast.ValueSpec{
			Type: typ,
		},
	}
}

func (b *ValueDeclarationBuilder) Set(name string, v Expr) *ValueDeclarationBuilder {
	b.spec.Names = append(b.spec.Names, ast.NewIdent(name))
	if v != nil {
		b.spec.Values = append(b.spec.Values, v.RenderExpr())
	}
	return b
}

//+envmap:with-docs=spec
type VarDeclarationBuilder struct{ *ValueDeclarationBuilder }
func Var(name string, d TypeDefinition, v Expr) *VarDeclarationBuilder {
	return &VarDeclarationBuilder{Value(name, d, v)}
}
func (b *VarDeclarationBuilder) RenderDecl() ast.Decl {
	return Vars().With(b.ValueDeclarationBuilder).RenderDecl()
}
//+envmap:with-docs=spec
type ConstDeclarationBuilder struct{ *ValueDeclarationBuilder }
func Const(name string, d TypeDefinition, v Expr) *ConstDeclarationBuilder {
	return &ConstDeclarationBuilder{Value(name, d, v)}
}
func (b *ConstDeclarationBuilder) RenderDecl() ast.Decl {
	return Consts().With(b.ValueDeclarationBuilder).RenderDecl()
}
