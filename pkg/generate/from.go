package generate

import (
	"fmt"
	"strings"

	"go/ast"
	"go/token"

	"github.com/directxman12/envmap/pkg/convert"
)

// DeclSorter defines how to sort imports and other declarations.  It can also
// be used insert blank newlines by inserting a nil declaration.
type DeclSorter func([]convert.Import, []convert.TypeDeclaration, []convert.FuncDeclaration, []convert.ValueDeclaration) ([]convert.Import, []convert.Declaration)

// TODO: would it be better just to preserve insertion order?

// DefaultDeclSorter inserts lines after every func, struct, and
// interface declaration, and after the group of value declarations.
// It groups methods together with the defining struct.
func DefaultDeclSorter(imports []convert.Import, typeDecls []convert.TypeDeclaration, funcDecls []convert.FuncDeclaration, valDecls []convert.ValueDeclaration) ([]convert.Import, []convert.Declaration) {
	// TODO: sort
	resDecls := make([]convert.Declaration, 0, len(valDecls)+5+len(typeDecls)*3+len(funcDecls)*3)

	resDecls = append(resDecls, nil, nil)
	for _, valDecl := range valDecls {
		resDecls = append(resDecls, valDecl)
	}
	resDecls = append(resDecls, nil, nil, nil) // TODO: why do we need three newlines?

	for _, typeDecl := range typeDecls {
		resDecls = append(resDecls, typeDecl, nil, nil)
	}
	for _, funcDecl := range funcDecls {
		resDecls = append(resDecls, funcDecl, nil, nil)
	}

	return imports, resDecls
}

func NewASTBuilder() *ASTBuilder {
	fileSet := token.NewFileSet()
	return &ASTBuilder{
		fileSet: fileSet,
		nextPosAbs: fileSet.Base(),
		OffsetIncrement: 1,
		DeclSorter: DefaultDeclSorter,
	}
}

// ASTBuilder constructs go AST nodes from convert interfaces.
// When being called from FromAST, it will also generate appropriate
// fake positions and a matching FileSet to use for
// appropriate formatting.
type ASTBuilder struct {
	fileSet *token.FileSet

	lines []int
	nextPosAbs int

	// OffsetIncrement determines the minimum space between tokens.
	// Increase this is you need to add extra positioning after generation.
	OffsetIncrement int

	// DeclSorter sorts declarations (and imports).
	DeclSorter DeclSorter
}

func (b *ASTBuilder) FileSet() *token.FileSet {
	return b.fileSet
}

func (b *ASTBuilder) nextPos() token.Pos {
	// TODO: actually deal with files, etc
	res := token.Pos(b.nextPosAbs)
	b.nextPosAbs += b.OffsetIncrement
	return res
}

func (b *ASTBuilder) newCommentGroup(strs ...string) *ast.CommentGroup {
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
			Slash: b.nextPos(),
		}
	}
	return &ast.CommentGroup{
		List: comments,
	}
}

// maybeCommentGroup checks if the given object implements Doced, and
// extracts docs into a comment group if it does.
func (b *ASTBuilder) maybeCommentGroup(obj interface{}) *ast.CommentGroup {
	asDoced, hasDocs := obj.(convert.Doced)
	if !hasDocs {
		return nil
	}

	return b.newCommentGroup(asDoced.Doc()...)
}

// The b.FromXXX methods convert *any* implementation of one of the
// convert types to a Go AST node.  These functions are guaranteed not
// to re-use existing AST nodes.

// TODO: evaluate everywhere we can have docs

func (b *ASTBuilder) line() {
	b.lines = append(b.lines, b.nextPosAbs-b.fileSet.Base())
	b.nextPosAbs++
}

func (b *ASTBuilder) FromAST(a convert.AST) *ast.File {
	// reset lines before generating any positions
	b.lines = nil

	typeDecls := a.Types()
	valDecls := a.Values()
	funcDecls := a.Funcs()

	sortedImports, sortedDecls := b.DeclSorter(a.Imports(), typeDecls, funcDecls, valDecls)

	res := &ast.File{
		Doc: b.maybeCommentGroup(a),
		Name: b.FromIdent(a.PackageName()),
		Imports: make([]*ast.ImportSpec, len(sortedImports)),
		Decls: make([]ast.Decl, 0, len(typeDecls)+len(valDecls)+len(funcDecls)),
	}

	for i, imp := range sortedImports {
		res.Imports[i] = b.FromImport(imp)
	}

	for _, decl := range sortedDecls {
		if decl == nil {
			b.line()
			continue
		}

		switch typedDecl := decl.(type) {
		case convert.ValueDeclaration:
			res.Decls = append(res.Decls, b.FromValueDeclaration(typedDecl))
		case convert.FuncDeclaration:
			res.Decls = append(res.Decls, b.FromFuncDeclaration(typedDecl))
		case convert.TypeDeclaration:
			res.Decls = append(res.Decls, b.FromTypeDeclaration(typedDecl))
		}
	}


	// construct the file after all positions have been generated
	lastBase := b.fileSet.Base()
	fileSize := b.nextPosAbs - lastBase
	file := b.fileSet.AddFile(fmt.Sprintf("package_%s.go", a.PackageName().Name()), lastBase, fileSize)
	file.SetLines(b.lines)
	b.nextPosAbs = b.fileSet.Base()

	return res
}

func (b *ASTBuilder) FromImport(i convert.Import) *ast.ImportSpec {
	return &ast.ImportSpec{
		Doc: b.maybeCommentGroup(i),
		Name: b.FromIdent(i.Name()),
		Path: &ast.BasicLit{
			Kind: token.STRING,
			Value: i.Path(),
		},
	}
}

func (b *ASTBuilder) FromValueDeclaration(d convert.ValueDeclaration) ast.Decl {
	tok := token.VAR
	if d.IsConst() {
		tok = token.CONST
	}
	var vals []ast.Expr
	if d.Value() != nil {
		vals = []ast.Expr{d.Value()}
	}
	spec := &ast.ValueSpec{
		Names: []*ast.Ident{b.FromIdent(d.Name())},
		Type: b.FromTypeDefinition(d.Type()),
		Values: vals,
	}
	return &ast.GenDecl{
		Doc: b.maybeCommentGroup(d),
		// always put token later, so that we get docs before the keyword
		TokPos: b.nextPos(),
		Tok: tok,
		Specs: []ast.Spec{spec},
	}
}

func (b *ASTBuilder) FromFuncDeclaration(d convert.FuncDeclaration) ast.Decl {
	var receiver *ast.FieldList
	if recvName, recvType := d.Receiver(); recvType != nil {
		var recvIdent []*ast.Ident
		if recvName != nil {
			recvIdent = []*ast.Ident{b.FromIdent(recvName)}
		}
		receiver = &ast.FieldList{
			List: []*ast.Field{
				{
					Names: recvIdent,
					Type: b.FromTypeDefinition(recvType),
				},
			},
		}
	}
	res := &ast.FuncDecl{
		Doc: b.maybeCommentGroup(d),
		Name: b.FromIdent(d.Name()),
		Type: b.FromFuncTypeDefinition(d.Type()),
		Recv: receiver,
		Body: d.Body(),
	}
	return res
}

func (b *ASTBuilder) FromTypeDeclaration(d convert.TypeDeclaration) ast.Decl {
	doc := b.maybeCommentGroup(d) // generate doc first, to get the right position
	typeTokPos := b.nextPos()
	spec := &ast.TypeSpec{
		Name: b.FromIdent(d.Name()),
		Type: b.FromTypeDefinition(d.Type()),
	}
	if d.IsAlias() {
		// any position will do
		spec.Assign = b.nextPos()
	}
	res := &ast.GenDecl{
		Tok: token.TYPE,
		// always put token later, so that we get docs before the keyword
		TokPos: typeTokPos,
		Specs: []ast.Spec{spec},
		Doc: doc,
	}
	return res
}

func (b *ASTBuilder) FromTypeDefinition(d convert.TypeDefinition) ast.Expr {
	switch typed := d.(type) {
	case convert.StructTypeDefinition:
		return b.FromStructTypeDefinition(typed)
	case convert.InterfaceTypeDefinition:
		return b.FromInterfaceTypeDefinition(typed)
	case convert.FuncTypeDefinition:
		return b.FromFuncTypeDefinition(typed)
	case convert.MapTypeDefinition:
		return b.FromMapTypeDefinition(typed)
	case convert.ChanTypeDefinition:
		return b.FromChanTypeDefinition(typed)
	case convert.PointerTypeDefinition:
		return b.FromPointerTypeDefinition(typed)
	case convert.SplatTypeDefinition:
		return b.FromSplatTypeDefinition(typed)
	case convert.ArrayTypeDefinition:
		return b.FromArrayTypeDefinition(typed)
	case convert.QualifiedIdent:
		return b.FromQualifiedIdent(typed)
	case convert.Ident:
		// NB: this *must* be after qualified ident, for similar reasons to array/splat
		return b.FromIdent(typed)
	default:
		// TODO: return error instead of panic
		panic(fmt.Sprintf("unknown/invalid expression type %T", d))
	}
}

func (b *ASTBuilder) newFieldList(fields []convert.Field) *ast.FieldList {
	rawFields := make([]*ast.Field, len(fields))
	for i, field := range fields {
		rawFields[i] = b.FromField(field)
	}
	return &ast.FieldList{
		List: rawFields,
	}
}

func (b *ASTBuilder) FromField(f convert.Field) *ast.Field {
	res := &ast.Field{
		Doc: b.maybeCommentGroup(f),
		Type: b.FromTypeDefinition(f.Type()),
	}
	name := f.Name()
	if name != nil {
		res.Names = []*ast.Ident{b.FromIdent(f.Name())}
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

func (b *ASTBuilder) FromStructTypeDefinition(d convert.StructTypeDefinition) ast.Expr {
	return &ast.StructType{
		Struct: b.nextPos(),
		Fields: b.newFieldList(d.Fields()),
	}
}

func (b *ASTBuilder) FromInterfaceTypeDefinition(d convert.InterfaceTypeDefinition) ast.Expr {
	return &ast.InterfaceType{
		Methods: b.newFieldList(d.Methods()),
	}

}

func (b *ASTBuilder) FromFuncTypeDefinition(d convert.FuncTypeDefinition) *ast.FuncType {
	return &ast.FuncType{
		Params: b.newFieldList(d.Params()),
		Results: b.newFieldList(d.Results()),
	}
}

func (b *ASTBuilder) FromMapTypeDefinition(d convert.MapTypeDefinition) ast.Expr {
	return &ast.MapType{
		Key: b.FromTypeDefinition(d.KeyType()),
		Value: b.FromTypeDefinition(d.ValueType()),
	}
}

func (b *ASTBuilder) FromArrayTypeDefinition(d convert.ArrayTypeDefinition) ast.Expr {
	res := &ast.ArrayType{
		Elt: b.FromTypeDefinition(d.ElemType()),
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

func (b *ASTBuilder) FromChanTypeDefinition(d convert.ChanTypeDefinition) ast.Expr {
	res := &ast.ChanType{
		Value: b.FromTypeDefinition(d.ValueType()),
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

func (b *ASTBuilder) FromPointerTypeDefinition(d convert.PointerTypeDefinition) ast.Expr {
	return &ast.StarExpr{
		X: b.FromTypeDefinition(d.ReferentType()),
	}
}

func (b *ASTBuilder) FromSplatTypeDefinition(d convert.SplatTypeDefinition) ast.Expr {
	return &ast.Ellipsis{
		Elt: b.FromTypeDefinition(d.ElemType()),
	}
}

func (b *ASTBuilder) FromIdent(i convert.Ident) *ast.Ident {
	if i == nil {
		return nil
	}
	return &ast.Ident{Name: i.Name()}
}

func (b *ASTBuilder) FromQualifiedIdent(i convert.QualifiedIdent) ast.Expr {
	return &ast.SelectorExpr{
		X: &ast.Ident{Name: i.PackageName()},
		Sel: &ast.Ident{Name: i.Name()},
	}
}
