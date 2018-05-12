package convert

import (
	"reflect"
)

// AsRawAST knows how to convert itself back into some go AST form
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

// Ident represents some sort of identifier
// (or potentially the lack thereof)
// TODO: capture underlying object as well for convinience?
// TODO: better method names, or get rid of the interface entirely
type Ident interface {
	Qualifier() string
	Unqualified() string
}

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
	IsSlice() bool
	AutoLength() bool
	Length() int
	ElemType() TypeDefinition
}
type ChanTypeDefinition interface {
	ValueType() TypeDefinition
	Directions() (receive bool, send bool)
}
type PointerTypeDefinition interface {
	ReferentType() TypeDefinition
}
type SplatTypeDefinition interface {
	ElemType() TypeDefinition
}