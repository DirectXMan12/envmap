package convert

import (
	"go/ast"
)

type unqualifiedIdent string
func (i unqualifiedIdent) Name() string { return string(i) }

type qualifiedIdent struct {
	packageName string
	Ident
}

func (i qualifiedIdent) PackageName() string {
	return i.packageName
}

type typeIdent struct {
	Ident
	typDecl ast.Expr
}
func (i typeIdent) LocateType() TypeDefinition {
	// TODO: is this correct for things other than embeds?
	return exprToTypeDefinition(i.typDecl)
}

func NewIdent(name string) Ident {
	return unqualifiedIdent(name)
}

func NewQualifiedIdent(pkgName string, i Ident) QualifiedIdent {
	return qualifiedIdent{
		packageName: pkgName,
		Ident: i,
	}
}
