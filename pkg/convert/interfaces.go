package convert

import (
	"reflect"
)

//go:generate go run $GOPATH/src/github.com/directxman12/envmap/cmd/basicimpl/main.go -p=Node -o=../generate/basic/types.go $GOFILE

var (
	// it's illegal for array lengths to be negative,
	// so this is fine, since Go doesn't have tagged unions.
	AutoLength = -1
)

// AsRawAST knows how to convert itself back into some go AST form
// +basicimpl:skip
type AsRawAST interface {
	// ToRaw converts this back into an appropriate Go AST object.
	// The object in question is dependent on the given implementer.
	// The object *may* (but is not guaranteed to) partially or wholly
	// return existing AST nodes
	ToRawNode() interface{}
}

// TypeDeclaration represents a declaration of a type in the AST.
type TypeDeclaration interface {
	Doced
	Name() Ident
	// IsAlias determines whether or not this declaration is an alias (i.e.
	// defined as `type name = ident`), or or actually defines a distinct type
	// (`type name spec`).
	IsAlias() bool
	// Type returns the actual underlying type.
	Type() TypeDefinition
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
	Doced
	Name() Ident
	Type() TypeDefinition
	Tag()  reflect.StructTag
}

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
