package convert

import (
	"go/ast"
	"go/token"
)

// AST wraps Go AST File declarations,
// making them easily traversable in
// familiar forms.
// 
// Note that the interfaces provided by AST "lose" some
// structural form from the original AST by breaking up grouped
// declarations into individual ones. This makes it more similar
// to reflection, but means code looking for groups of declarations
// is harder to write.  Docs from the overall declaration are merged
// with docs from individual ones.
type astImpl struct {
	file *ast.File
}

func FromRaw(raw *ast.File) AST {
	return &astImpl{
		file: raw,
	}
}

// TODO: do iota generations and var lists with skipped types have values/types in the AST?

// TODO: make this more like a visitor to avoid extra allocations

// Types fetches all types defined in this AST
func (a *astImpl) Types() []TypeDeclaration {
	var res []TypeDeclaration

	for _, decl := range a.file.Decls {
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

func (a *astImpl) Funcs() []FuncDeclaration {
	var res []FuncDeclaration

	for _, decl := range a.file.Decls {
		// skip non-type declarations
		funcDecl, isFuncDecl := decl.(*ast.FuncDecl)
		if !isFuncDecl {
			continue
		}
		res = append(res, &funcDeclaration{
			decl: funcDecl,
		})
	}
	return res
}

func (a *astImpl) Values() []ValueDeclaration {
	var res []ValueDeclaration

	for _, decl := range a.file.Decls {
		// skip non-type declarations
		genDecl, isGenDecl := decl.(*ast.GenDecl)
		if !isGenDecl {
			continue
		}
		if genDecl.Tok != token.VAR && genDecl.Tok != token.CONST {
			continue
		}

		for _, specRaw := range genDecl.Specs {
			spec := specRaw.(*ast.ValueSpec)

			for i, name := range spec.Names {
				var val ast.Expr
				if spec.Values != nil {
					val = spec.Values[i]
				}
				res = append(res, &valueDeclaration{
					decl: genDecl,
					spec: spec, 
					name: name,
					value: val,
				})
			}
		}
	}
	return res
}

func (a *astImpl) PackageName() Ident {
	return NewIdent(a.file.Name.Name)
}

func (a *astImpl) Doc() []string {
	return extractCommentGroup(a.file.Doc)
}

func (a *astImpl) Imports() []Import {
	res := make([]Import, len(a.file.Imports))
	for i, spec := range a.file.Imports {
		res[i] = &importSpec{
			spec: spec,
		}
	}
	return res
}

type importSpec struct {
	spec *ast.ImportSpec
}

func (s *importSpec) Doc() []string {
	return extractCommentGroup(s.spec.Doc)
}

func (s *importSpec) Name() Ident {
	if s.spec.Name == nil {
		return nil
	}
	return NewIdent(s.spec.Name.Name)
}

func (s *importSpec) Path() string {
	return s.spec.Path.Value
}
