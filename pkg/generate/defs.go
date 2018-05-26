package generate

import (
	"go/ast"
	"go/token"
)

//go:generate go run $GOPATH/src/github.com/directxman12/envmap/cmd/buildergen/main.go -- $GOFILE

func AutoSize() Expr {
	return &expr{&ast.Ellipsis{}}
}

func SliceOf(d TypeDefinition) TypeDefinition {
	return typeDefinition{&ast.ArrayType{Elt: d.RenderExpr()}}
}

func SplatOf(d TypeDefinition) TypeDefinition {
	return typeDefinition{&ast.Ellipsis{Elt: d.RenderExpr()}}
}

func ArrayOf(d TypeDefinition, size Expr) TypeDefinition {
	return typeDefinition{&ast.ArrayType{Elt: d.RenderExpr(), Len: size.RenderExpr()}}
}

func PointerTo(d TypeDefinition) TypeDefinition {
	return typeDefinition{&ast.StarExpr{X: d.RenderExpr()}}
}

func MapOf(k TypeDefinition, v TypeDefinition) TypeDefinition {
	return typeDefinition{&ast.MapType{Key: k.RenderExpr(), Value: v.RenderExpr()}}
}

func ChanOf(d TypeDefinition, send, recv bool) TypeDefinition {
	var dir ast.ChanDir
	if send {
		dir |= ast.SEND
	}
	if recv {
		dir |= ast.RECV
	}
	return typeDefinition{
		&ast.ChanType{
			Dir: dir,
			Value: d.RenderExpr(),
		},
	}
}

//+envmap:render=iface,ast.Expr
type InterfaceDefinitionBuilder struct {
	iface *ast.InterfaceType
}

func Interface() *InterfaceDefinitionBuilder {
	return &InterfaceDefinitionBuilder{
		iface: &ast.InterfaceType{
			Methods: &ast.FieldList{},
		},
	}
}

func (b *InterfaceDefinitionBuilder) With(m FieldGroup) *InterfaceDefinitionBuilder {
	b.iface.Methods.List = append(b.iface.Methods.List, m.RenderField())
	return b
}

//+envmap:render=strct,ast.Expr
type StructDefinitionBuilder struct {
	strct *ast.StructType
}

func Struct() *StructDefinitionBuilder {
	return &StructDefinitionBuilder{
		strct: &ast.StructType{
			Fields: &ast.FieldList{},
		},
	}
}

func (b *StructDefinitionBuilder) With(m FieldGroup) *StructDefinitionBuilder {
	b.strct.Fields.List = append(b.strct.Fields.List, m.RenderField())
	return b
}

//+envmap:render=field,*ast.Field
//+envmap:with-docs=field
type FieldGroupBuilder struct {
	field *ast.Field
}

func Fields(d TypeDefinition, names... string) *FieldGroupBuilder {
	return &FieldGroupBuilder{
		field: &ast.Field{
			Names: nameIdents(names),
			Type: d.RenderExpr(),
		},
	}
}

func Field(name string, d TypeDefinition) *FieldGroupBuilder {
	return &FieldGroupBuilder{
		field: &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(name)},
			Type: d.RenderExpr(),
		},
	}
}

func Method(name string, d TypeDefinition) *FieldGroupBuilder {
	return &FieldGroupBuilder{
		field: &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(name)},
			Type: d.RenderExpr(),
		},
	}
}

func Embed(d TypeDefinition) *FieldGroupBuilder {
	return &FieldGroupBuilder{
		field: &ast.Field{Type: d.RenderExpr()},
	}
}

func (b *FieldGroupBuilder) WithTag(tag string) *FieldGroupBuilder {
	b.field.Tag = &ast.BasicLit{
		Kind: token.STRING,
		Value: "`"+tag+"`",
	}
	return b
}
// TODO: comment

//+envmap:render=fun,*ast.FuncType
//+envmap:render=fun,ast.Expr
type FunctionDefinitionBuilder struct {
	fun *ast.FuncType
}

func Function() *FunctionDefinitionBuilder {
	return &FunctionDefinitionBuilder{
		fun: &ast.FuncType{
			Params: &ast.FieldList{},
			Results: &ast.FieldList{},
		},
	}
}

func (b *FunctionDefinitionBuilder) WithParams(d TypeDefinition, names ...string) *FunctionDefinitionBuilder {
	b.fun.Params.List = append(b.fun.Params.List, &ast.Field{
		Names: nameIdents(names),
		Type: d.RenderExpr(),
	})
	return b
}

func (b *FunctionDefinitionBuilder) WithResults(d TypeDefinition, names ...string) *FunctionDefinitionBuilder {
	b.fun.Results.List = append(b.fun.Results.List, &ast.Field{
		Names: nameIdents(names),
		Type: d.RenderExpr(),
	})
	return b
}

func (b *FunctionDefinitionBuilder) WithParam(name string, d TypeDefinition) *FunctionDefinitionBuilder {
	var nameIdents []*ast.Ident
	if name != "" {
		nameIdents = []*ast.Ident{ast.NewIdent(name)}
	}
	b.fun.Params.List = append(b.fun.Params.List, &ast.Field{
		Names: nameIdents,
		Type: d.RenderExpr(),
	})
	return b
}

func (b *FunctionDefinitionBuilder) WithResult(name string, d TypeDefinition) *FunctionDefinitionBuilder {
	var nameIdents []*ast.Ident
	if name != "" {
		nameIdents = []*ast.Ident{ast.NewIdent(name)}
	}
	b.fun.Results.List = append(b.fun.Results.List, &ast.Field{
		Names: nameIdents,
		Type: d.RenderExpr(),
	})
	return b
}

func (b *FunctionDefinitionBuilder) WithBody(body BlockStatement) Expr {
	return &expr{
		&ast.FuncLit{
			Type: b.RenderFuncType(),
			Body: body.RenderBlockStmt(),
		},
	}
}

func (b *FunctionDefinitionBuilder) DeclaredAs(name string) *FunctionDeclarationBuilder {
	return &FunctionDeclarationBuilder{
		decl: &ast.FuncDecl{
			Name: ast.NewIdent(name),
			Type: b.RenderFuncType(),
		},
	}
}
