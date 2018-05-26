package traverse

import (
	"go/ast"
)

// TypeDeclaration represents an individual declared type.
type TypeDeclaration struct {
	Decl *ast.GenDecl
	Spec *ast.TypeSpec
	traverser *Traverser

	cachedTags map[string][]GenTag
}

func (t *TypeDeclaration) Tag(label string) []GenTag {
	// TODO: make this more efficient
	allTags := ExtractTags(t.AllDocs())
	return allTags[label]
}

// AllDocs returns all lines of doc comments associated with this
// type declaration (from both the decl and the spec).
func (t *TypeDeclaration) AllDocs() []string {
	var res []string
	res = append(res, ExtractCommentGroup(t.Decl.Doc)...)
	res = append(res, ExtractCommentGroup(t.Spec.Doc)...)
	return res
}

// Name returns the string form of this type's name.
func (t *TypeDeclaration) Name() string {
	return t.Spec.Name.Name
}

type Types []TypeDeclaration
func (t Types) WithTag(label string) Types {
	var res Types
	for _, decl := range t {
		if tags := decl.Tag(label); tags != nil {
			res = append(res, decl)
		}
	}
	return res
}

func AsQualifiedIdent(typeRaw ast.Expr) (pkgAlias string, typeName string, isIdent bool) {
	switch typed := typeRaw.(type) {
	case *ast.Ident:
		return "", typed.Name, true
	case *ast.SelectorExpr:
		if pkgNameAsIdent, isIdent := typed.X.(*ast.Ident); isIdent {
			return pkgNameAsIdent.Name, typed.Sel.Name, true
		}
	}
	return "", "", false
}
