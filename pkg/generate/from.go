package generate

import (
	"fmt"
	"strings"

	"go/ast"
	"go/token"

	"github.com/directxman12/envmap/pkg/convert"
)

func newCommentGroup(strs ...string) *ast.CommentGroup {
	comments := make([]*ast.Comment, len(strs))
	for i, rawComment := range strs {
		if strings.Contains(rawComment, "\n") {
			// multiline
			rawComment = "/*"+rawComment+"*/"
		} else {
			// single line
			rawComment = "// "+rawComment
		}

		comments[i] = &ast.Comment{
			Text: rawComment,
			// make sure decl docs appear before the `type` keyword
			Slash: token.Pos(1), 
		}
	}
	return &ast.CommentGroup{
		List: comments,
	}
}

func declIdent(i convert.Ident) *ast.Ident {
	return &ast.Ident{
		Name: i.Name(),
	}
}

// The FromXXX methods convert *any* implementation of one of the
// convert types to a Go AST node, similarly to ToRawNode.  These
// do not attempt to call ToRawNode -- callers should do that manually
// if they want to take advantage of that optimization.  These functions
// are guaranteed not to re-use AST nodes.

func FromTypeDeclaration(d convert.TypeDeclaration) ast.Decl {
	spec := &ast.TypeSpec{
		// TODO: manually check type to make sure it's unqualified?
		Name: declIdent(d.Name()),
		Type: FromTypeDefinition(d.Type()),
	}
	if d.IsAlias() {
		// any position will do
		spec.Assign = token.Pos(1)
	}
	return &ast.GenDecl{
		Tok: token.TYPE,
		// always put token later, so that we get docs before the keyword
		TokPos: token.Pos(2),
		Specs: []ast.Spec{spec},
		Doc: newCommentGroup(d.Doc()...),
	}

}

func FromTypeDefinition(d convert.TypeDefinition) ast.Expr {
	switch typed := d.(type) {
	case convert.StructTypeDefinition:
		return FromStructTypeDefinition(typed)
	case convert.InterfaceTypeDefinition:
		return FromInterfaceTypeDefinition(typed)
	case convert.FuncTypeDefinition:
		return FromFuncTypeDefinition(typed)
	case convert.MapTypeDefinition:
		return FromMapTypeDefinition(typed)
	case convert.ChanTypeDefinition:
		return FromChanTypeDefinition(typed)
	case convert.PointerTypeDefinition:
		return FromPointerTypeDefinition(typed)
	case convert.SplatTypeDefinition:
		return FromSplatTypeDefinition(typed)
	case convert.ArrayTypeDefinition:
		return FromArrayTypeDefinition(typed)
	case convert.QualifiedIdent:
		return FromQualifiedIdent(typed)
	case convert.Ident:
		// NB: this *must* be after qualified ident, for similar reasons to array/splat
		return FromIdent(typed)
	default:
		// TODO: return error instead of panic
		panic(fmt.Sprintf("unknown/invalid expression type %T", d))
	}
}

func newFieldList(fields []convert.Field) *ast.FieldList {
	rawFields := make([]*ast.Field, len(fields))
	for i, field := range fields {
		rawFields[i] = FromField(field)
	}
	return &ast.FieldList{
		List: rawFields,
	}
}

func FromField(f convert.Field) *ast.Field {
	res := &ast.Field{
		Doc: newCommentGroup(f.Doc()...),
		Type: FromTypeDefinition(f.Type()),
	}
	name := f.Name()
	if name != convert.Anonymous {
		res.Names = []*ast.Ident{declIdent(f.Name())}
	}
	tag := f.Tag()
	if tag != "" {
		res.Tag = &ast.BasicLit{
			Kind: token.STRING,
			Value: string(f.Tag()),
		}
	}
	return res
}

func FromStructTypeDefinition(d convert.StructTypeDefinition) ast.Expr {
	return &ast.StructType{
		Fields: newFieldList(d.Fields()),
	}
}

func FromInterfaceTypeDefinition(d convert.InterfaceTypeDefinition) ast.Expr {
	return &ast.InterfaceType{
		Methods: newFieldList(d.Methods()),
	}

}

func FromFuncTypeDefinition(d convert.FuncTypeDefinition) ast.Expr {
	return &ast.FuncType{
		Params: newFieldList(d.Params()),
		Results: newFieldList(d.Results()),
	}
}

func FromMapTypeDefinition(d convert.MapTypeDefinition) ast.Expr {
	return &ast.MapType{
		Key: FromTypeDefinition(d.KeyType()),
		Value: FromTypeDefinition(d.ValueType()),
	}
}

func FromArrayTypeDefinition(d convert.ArrayTypeDefinition) ast.Expr {
	res := &ast.ArrayType{
		Elt: FromTypeDefinition(d.ElemType()),
	}
	maybeLen := d.Length()
	if maybeLen == nil {
		// it's just a slice
		return res
	}

	if *maybeLen == convert.AutoLength {
		// auto length is `[...]T`
		res.Len = &ast.Ellipsis{}
		return res
	}
	
	res.Len = &ast.BasicLit{
		Kind: token.INT,
		Value: fmt.Sprintf("%d", *maybeLen),
	}

	return res
}

func FromChanTypeDefinition(d convert.ChanTypeDefinition) ast.Expr {
	res := &ast.ChanType{
		Value: FromTypeDefinition(d.ValueType()),
	}
	recv, send := d.Directions()
	if recv {
		res.Dir |= ast.RECV
	}
	if send {
		res.Dir |= ast.SEND
	}
	return res
}

func FromPointerTypeDefinition(d convert.PointerTypeDefinition) ast.Expr {
	return &ast.StarExpr{
		X: FromTypeDefinition(d.ReferentType()),
	}
}

func FromSplatTypeDefinition(d convert.SplatTypeDefinition) ast.Expr {
	return &ast.Ellipsis{
		Elt: FromTypeDefinition(d.ElemType()),
	}
}

func FromIdent(i convert.Ident) *ast.Ident {
	// TODO: we already have this code in ToRawNode,
	// can we find a good way not to duplicate it?
	if i == convert.Anonymous {
		return nil
	}
	return &ast.Ident{Name: i.Name()}
}

func FromQualifiedIdent(i convert.QualifiedIdent) ast.Expr {
	return &ast.SelectorExpr{
		X: &ast.Ident{Name: i.PackageName()},
		Sel: &ast.Ident{Name: i.Name()},
	}
}
