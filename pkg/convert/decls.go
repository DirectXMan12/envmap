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

// Declarations to support:
// - [ ] import (gen)
// - [ ] constant (gen)
// - [x] type (gen)
// - [ ] variable (gen)
// - [ ] bad?
// - [ ] declstmt?
// - [ ] func

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

func (d *typeDeclaration) ToRawNode() interface{} {
	// we can't just use the underlying decl,
	// because it might have been shared, so
	// construct a new one (losing the grouping)
	return &ast.GenDecl{
		Doc: d.decl.Doc,
		Tok: token.TYPE,
		Specs: []ast.Spec{d.spec},
	}
}
