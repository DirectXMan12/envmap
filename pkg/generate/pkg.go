package generate

import (
	"go/ast"
)
//go:generate go run $GOPATH/src/github.com/directxman12/envmap/cmd/buildergen/main.go -- $GOFILE

//+envmap:with-docs=file
//+envmap:render=file,*ast.File
type PackageBuilder struct {
	file *ast.File
	b *Builder
}

func Package(name string) *PackageBuilder {
	return &PackageBuilder{
		file: &ast.File{
			Name: ast.NewIdent(name),
		},
		// TODO: allow setting positioner
	}

}
func (p *PackageBuilder) Declare(decl Declaration) *PackageBuilder {
	p.file.Decls = append(p.file.Decls, decl.RenderDecl())
	return p
}
