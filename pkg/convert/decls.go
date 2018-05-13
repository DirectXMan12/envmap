package convert

import (
	"go/ast"
	"go/token"
)

// extractCommentGroup extracts the actual contents of a
// comment group.  It will safely deal with nil comment groups.
// TODO: switch to just using cg.Text()?
func extractCommentGroup(cg *ast.CommentGroup) []string {
	if cg == nil {
		return nil
	}
	res := make([]string, len(cg.List))
	for i, commentFull := range cg.List {
		comment := commentFull.Text
		switch comment[1] {
		case '/':
			//-style comment
			comment = comment[2:]
			if len(comment) > 0 && comment[0] == ' ' {
				comment = comment[1:]
			}
		case '*':
			/*-style comment */
			comment = comment[2:len(comment)-2]
		}
		res[i] = comment
	}
	return res
}

type typeDeclaration struct {
	decl *ast.GenDecl
	spec *ast.TypeSpec
}

func (d *typeDeclaration) IsAlias() bool {
	return d.spec.Assign != token.NoPos
}

// Doc returns all the comments associated with this type
func (d *typeDeclaration) Doc() []string {
	// TODO: figure out how to include line comments?
	// TODO: separate declaration docs from spec docs?
	return append(extractCommentGroup(d.decl.Doc), extractCommentGroup(d.spec.Doc)...)
}

// Name returns the name of the type
func (d *typeDeclaration) Name() Ident {
	return unqualifiedIdent(d.spec.Name.Name)
}

// Type returns the actual underlying type.
func (d *typeDeclaration) Type() TypeDefinition {
	return exprToTypeDefinition(d.spec.Type)
}

type valueDeclaration struct {
	decl *ast.GenDecl
	spec *ast.ValueSpec
	name *ast.Ident
	value ast.Expr
}

func (d *valueDeclaration) IsConst() bool {
	return d.decl.Tok == token.CONST
}

func (d *valueDeclaration) Name() Ident {
	return unqualifiedIdent(d.name.Name)
}

func (d *valueDeclaration) Type() TypeDefinition {
	return exprToTypeDefinition(d.spec.Type)
}

func (d *valueDeclaration) Doc() []string {
	// TODO: figure out how to include line comments?
	// TODO: separate declaration docs from spec docs?
	return append(extractCommentGroup(d.decl.Doc), extractCommentGroup(d.spec.Doc)...)
}
 
func (d *valueDeclaration) Value() ast.Expr {
	return d.value
}

type funcDeclaration struct {
	decl *ast.FuncDecl
}

func (d *funcDeclaration) Receiver() (Ident, TypeDefinition) {
	if d.decl.Recv == nil {
		return nil, nil
	}

	names := d.decl.Recv.List[0].Names
	typeDef := exprToTypeDefinition(d.decl.Recv.List[0].Type)

	if names == nil {
		return Anonymous, typeDef
	}

	return unqualifiedIdent(names[0].Name), typeDef
}

func (d *funcDeclaration) Name() Ident {
	return unqualifiedIdent(d.decl.Name.Name)
}

func (d *funcDeclaration) Type() FuncTypeDefinition {
	return &funcTypeDefinition{
		typ: d.decl.Type,
	}
}

func (d *funcDeclaration) Body() *ast.BlockStmt {
	return d.decl.Body
}

func (d *funcDeclaration) Doc() []string {
	return extractCommentGroup(d.decl.Doc)
}
