package convert

import (
	"go/ast"
)

type anonIdent struct{}
func (i anonIdent) Name() string { return "" }

type unqualifiedIdent string
func (i unqualifiedIdent) Name() string { return string(i) }
func (i unqualifiedIdent) ToRawNode() interface{} {
	return &ast.Ident{Name: string(i)}
}

type qualifiedIdent struct {
	packageName string
	Ident
}

func (i qualifiedIdent) PackageName() string {
	return i.packageName
}
func (i qualifiedIdent) ToRawNode() interface{} {
	return &ast.SelectorExpr{
		X: &ast.Ident{Name: i.packageName},
		Sel: &ast.Ident{Name: i.Name()},
	}
}

var (
	Anonymous = anonIdent{}
)

func NewIdent(name string) Ident {
	return unqualifiedIdent(name)
}

func NewQualifiedIdent(pkgName string, i Ident) QualifiedIdent {
	return qualifiedIdent{
		packageName: pkgName,
		Ident: i,
	}
}
