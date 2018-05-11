package convert

import (
	"go/ast"
	"go/token"
)

// AST wraps Go AST File declarations,
// making them easily traversable in
// familiar forms.
type AST struct {
	decls []ast.Decl
}

func FromRaw(raw *ast.File) *AST {
	return &AST{
		decls: raw.Decls,
	}
}

// TODO: make this more like a visitor to avoid extra allocations

// Types fetches all types defined in this AST
func (a *AST) Types() []TypeDeclaration {
	var res []TypeDeclaration

	for _, decl := range a.decls {
		// skip non-type declarations
		genDecl, isGenDecl := decl.(*ast.GenDecl)
		if !isGenDecl {
			continue
		}
		if genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			res = append(res, &typeDeclaration{
				decl: genDecl,
				spec: spec.(*ast.TypeSpec),
			})
		}
	}
	return res
}

