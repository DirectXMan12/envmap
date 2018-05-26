package traverse

import (
	"go/ast"
	"go/token"
)

type Traverser struct {
	File *ast.File

	prePackageComments []*ast.CommentGroup
	preDeclComments map[ast.Decl][]*ast.CommentGroup
}

func File(File *ast.File) *Traverser {
	return &Traverser{
		File: File,
	}
}

func (t *Traverser) PackageName() string {
	return t.File.Name.Name
}

/*func (t *Traverser) buildCommentAssociations() {
	prePackageComments := nil
	preDeclComments := make(map[ast.Decl][]*ast.CommentGroup)

	commentInd := 0
	comments := File.Comments
	res := make(map[ast.Decl][]*ast.CommentGroup)

	for ; comments[commentInd].Pos() < t.File.Package; commentInd++ {
		prePackageComments = append(prePackageComments, comments[commentInd])
	}


	lastEnd := t.File.Package
	for _, decl := range t.File.Decls {
		for comments[commentInd].Pos() < lastEnd {
			// skip comments inside a decl
			commentInd++
			continue
		}
		for comments[commentInd].Pos() < decl.Pos() {
			res[decl] = append(res[decl], comments[commentInd])
			commentInd++
		}
	}

	t.prePackageComments = prePackageComments
	t.preDeclComments = preDeclComments

	// working
	commentInd := 0
	comments := File.Comments
	res := make(map[ast.Node][]*ast.CommentGroup)

	for ; commentInd < len(comments) && comments[commentInd].Pos() < File.Package; commentInd++ {
		res[File] = append(res[File], comments[commentInd])
	}

	for _, decl := range File.Decls {
		for commentInd < len(comments) && comments[commentInd].Pos() < decl.Pos() {
			res[decl] = append(res[decl], comments[commentInd])
			commentInd++
		}
	}
	return &commentAssociations{assocs: res}
}*/

// Types returns all top-level types declared in this File.
func (t *Traverser) Types() Types {
	var res []TypeDeclaration
	for _, decl := range t.File.Decls {
		// type declaration groups are GenDecls of kind TYPE
		genDecl, isGenDecl := decl.(*ast.GenDecl)
		if !isGenDecl || genDecl.Tok != token.TYPE {
			continue
		}

		for _, specRaw := range genDecl.Specs {
			typeSpec := specRaw.(*ast.TypeSpec)
			res = append(res, TypeDeclaration{
				Decl: genDecl,
				Spec: typeSpec,
				traverser: t,
			})
		}
	}
	return res
}
