package convert

import (
	"reflect"
	"go/ast"
)

//go:generate go run $GOPATH/src/github.com/directxman12/envmap/cmd/basicimpl/main.go -p=Node -o=../generate/basic/types.go $GOFILE

var (
	// it's illegal for array lengths to be negative,
	// so this is fine, since Go doesn't have tagged unions.
	AutoLength = -1
)

// NB: any method which returns something from the "go/ast" package
// may have its signature changed in the future.

// Declarations

// Declaration represents a go declaration.
// It can be a:
// - TypeDeclaration
// - ValueDeclaration
// - FuncDeclaration
type Declaration interface{}

// TypeDeclaration represents a declaration of a type in the AST.
type TypeDeclaration interface {
	Name() Ident
	// IsAlias determines whether or not this declaration is an alias (i.e.
	// defined as `type name = ident`), or or actually defines a distinct type
	// (`type name spec`).
	IsAlias() bool
	// Type returns the actual underlying type.
	Type() TypeDefinition
}

// We skip import declaratations because we get those elsewhere

// ValueDeclaration represents a const or var declaration
type ValueDeclaration interface {
	IsConst() bool
	Name() Ident
	Type() TypeDefinition
	Value() ast.Expr // TODO: deal with this
}

type FuncDeclaration interface {
	// Receiver returns the receiver for this function, if it's a method.
	// A nil typedefinition indicates no receiver (a method)
	Receiver() (Ident, TypeDefinition)
	Name() Ident
	Type() FuncTypeDefinition
	Body() *ast.BlockStmt // TODO: deal with this
}

// Type Definitions

// TypeDefinition represents some type in Go.
// It may be a:
// - StructTypeDefinition
// - InterfaceTypeDefinition
// - FuncTypeDefinition
// - MapTypeDefinition
// - ArrayTypeDefinition (slices or arrays)
// - ChanTypeDefinition
// - PointerTypeDefinition
// - SplatTypeDefinition
// - Ident
// - QualifiedIdent
// +basicimpl:skip
type TypeDefinition interface{}

type StructTypeDefinition interface {
	Fields() []Field
}
type InterfaceTypeDefinition interface {
	Methods() []Field
}
type FuncTypeDefinition interface {
	Params() []Field
	Results() []Field
}
type MapTypeDefinition interface {
	KeyType() TypeDefinition
	ValueType() TypeDefinition
}

type ArrayTypeDefinition interface {
	ElemType() TypeDefinition
	// Length is the length of the array.
	// A nil length length represents a slice,
	// and a length of AutoLength represents `[...]T`.
	// TODO: fix this so that we can represent constant expressions here
	Length() *int
}
type SplatTypeDefinition interface {
	ElemType() TypeDefinition
	// IsSplat indicates that this is a splat, instead
	// of a normal slice.
	IsSplat() struct{}
}
type ChanTypeDefinition interface {
	ValueType() TypeDefinition
	Directions() (receive bool, send bool)
}
type PointerTypeDefinition interface {
	ReferentType() TypeDefinition
}

// Other types

type Import interface {
	Name() Ident
	Path() string
}

type AST interface {
	PackageName() Ident
	Types() []TypeDeclaration
	Funcs() []FuncDeclaration
	Values() []ValueDeclaration
	Imports() []Import
}

// +basicimpl:skip
type TypeIdent interface {
	LocateType() TypeDefinition
	Ident
}

// Ident is a bare identifier
// +basicimpl:skip
type Ident interface {
	Name() string
}

// QualifiedIdent is an identifier qualified by a package name
// +basicimpl:skip
type QualifiedIdent interface {
	Ident
	PackageName() string
}
// TODO: capture underlying object as well for convinience?

// +basicimpl:skip
type Doced interface {
	Doc() []string
}

type Field interface {
	Name() Ident
	Type() TypeDefinition
	Tag()  reflect.StructTag
}
