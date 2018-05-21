// package builder declares builder helpers for building implementations
// of the convert interfaces.
package builder

import (
	"fmt"
	"reflect"

	"go/ast"

	"github.com/directxman12/envmap/pkg/convert"
)

// builtPtr represents a concrete pointer to another type
type builtPtr struct { referent convert.TypeDefinition }
func (p *builtPtr) ReferentType() convert.TypeDefinition { return p.referent }
func PointerTo(referent convert.TypeDefinition) convert.PointerTypeDefinition {
	return &builtPtr{referent: referent}
}

// builtSplat represents a concrete splat of another type
type builtSplat struct { elemType convert.TypeDefinition }
func (b *builtSplat) ElemType() convert.TypeDefinition { return b.elemType }
func (b *builtSplat) IsSplat() struct{} { return struct{}{} }
func SplatOf(elemType convert.TypeDefinition) convert.SplatTypeDefinition {
	return &builtSplat{elemType: elemType}
}

// builtArray represents a concrete array or slice of another type
type builtArray struct {
	elemType convert.TypeDefinition
	length *int
}
func (b *builtArray) ElemType() convert.TypeDefinition { return b.elemType }
func (b *builtArray) Length() *int { return b.length }
func SliceOf(elemType convert.TypeDefinition) convert.ArrayTypeDefinition {
	return &builtArray{elemType: elemType}
}
func ArrayOf(elemType convert.TypeDefinition, length int) convert.ArrayTypeDefinition {
	return &builtArray{elemType: elemType, length: &length}
}

// builtMap represents a concrete map type
type builtMap struct { key, value convert.TypeDefinition }
func (b *builtMap) KeyType() convert.TypeDefinition { return b.key }
func (b *builtMap) ValueType() convert.TypeDefinition { return b.value }
func MapOf(key, value convert.TypeDefinition) convert.MapTypeDefinition {
	return &builtMap{key: key, value: value}
}

// builtChan represents a concrete channel type
type builtChan struct {
	elem convert.TypeDefinition
	recv, send bool
}
func (b *builtChan) ValueType() convert.TypeDefinition { return b.elem }
func (b *builtChan) Directions() (receive bool, send bool) { return b.recv, b.send }
func ChanOf(value convert.TypeDefinition) convert.ChanTypeDefinition {
	return &builtChan{
		elem: value,
		recv: true,
		send: true,
	}
}
func SendChanOf(value convert.TypeDefinition) convert.ChanTypeDefinition {
	return &builtChan{
		elem: value,
		send: true,
	}
}
func ReceiveChanOf(value convert.TypeDefinition) convert.ChanTypeDefinition {
	return &builtChan{
		elem: value,
		recv: true,
	}
}

// builtDoc represents some concrete docs
type builtDoc struct {
	doc []string
}
func (d *builtDoc) Doc() []string { return d.doc }

// builtField represents a concrete field
type builtField struct {
	name string
	typ convert.TypeDefinition
	tag reflect.StructTag
}
func (f *builtField) Name() convert.Ident {
	if f.name == "" { return nil }
	return convert.NewIdent(f.name)
}
func (f *builtField) Type() convert.TypeDefinition { return f.typ }
func (f *builtField) Tag() reflect.StructTag { return f.tag }

// builtImport represents a concrete imported package
// TODO: expose constructing this manually?
type builtImport struct {
	alias, path string
}
func (i *builtImport) Name() convert.Ident {
	if (i.alias != "") { return convert.NewIdent(i.alias) }
	return nil
}
func (i *builtImport) Path() string { return i.path }

// TypeDeclarationBuilder builds a concrete type declaration
type TypeDeclarationBuilder struct {
	builtDoc
	name string
	isAlias bool
	typ convert.TypeDefinition
}
func (d *TypeDeclarationBuilder) Name() convert.Ident { return convert.NewIdent(d.name) }
func (d *TypeDeclarationBuilder) IsAlias() bool { return d.isAlias }
func (d *TypeDeclarationBuilder) Type() convert.TypeDefinition { return d.typ }
func (d *TypeDeclarationBuilder) WithDoc(lines ...string) *TypeDeclarationBuilder {
	d.doc = lines
	return d
}
func Alias(name string, typ convert.TypeDefinition) *TypeDeclarationBuilder {
	return &TypeDeclarationBuilder{
		name: name,
		typ: typ,
		isAlias: true,
	}
}
func Type(name string, typ convert.TypeDefinition) *TypeDeclarationBuilder {
	return &TypeDeclarationBuilder{
		name: name,
		typ: typ,
	}
}

// FuncDeclBuilder build a function or method declarations
type FuncDeclBuilder struct {
	builtDoc
	name string
	typ convert.FuncTypeDefinition
	body *ast.BlockStmt

	receiverName string
	receiverType convert.Ident
	ptrReceiver bool
}
func (d *FuncDeclBuilder) Name() convert.Ident { return convert.NewIdent(d.name) }
func (d *FuncDeclBuilder) Type() convert.FuncTypeDefinition { return d.typ }
func (d *FuncDeclBuilder) Body() *ast.BlockStmt { return d.body }
func (d *FuncDeclBuilder) Receiver() (convert.Ident, convert.TypeDefinition) {
	name := convert.NewIdent(d.receiverName)
	var typ convert.TypeDefinition = d.receiverType
	if d.ptrReceiver {
		typ = &builtPtr{typ}
	}

	return name, typ
}

func (d *FuncDeclBuilder) WithDoc(lines ...string) *FuncDeclBuilder {
	d.doc = lines
	return d
}
func (d *FuncDeclBuilder) WithBody(body *ast.BlockStmt) *FuncDeclBuilder {
	d.body = body
	return d
}
func (d *FuncDeclBuilder) AsMethodFor(id, typeName string) *FuncDeclBuilder {
	d.receiverName = id
	d.receiverType = convert.NewIdent(typeName)
	return d
}
func (d *FuncDeclBuilder) AsMethodForPointer(id, typeName string) *FuncDeclBuilder {
	d.receiverName = id
	d.receiverType = convert.NewIdent(typeName)
	d.ptrReceiver = true
	return d
}

// ValueDeclBuilder builds a variable or constant declaration
type ValueDeclBuilder struct {
	builtDoc
	isConst bool
	name string
	typ convert.TypeDefinition
	val ast.Expr
}
func (d *ValueDeclBuilder) IsConst() bool { return d.isConst }
func (d *ValueDeclBuilder) Name() convert.Ident {
	if d.name == "" { return nil }
	return convert.NewIdent(d.name)
}
func (d *ValueDeclBuilder) Type() convert.TypeDefinition { return d.typ }
func (d *ValueDeclBuilder) Value() ast.Expr { return d.val }
func (d *ValueDeclBuilder) WithDoc(lines ...string) *ValueDeclBuilder {
	d.doc = lines
	return d
}
func Var(name string, typ convert.TypeDefinition, val ast.Expr) *ValueDeclBuilder {
	return &ValueDeclBuilder{
		name: name,
		typ: typ,
		val: val,
	}
}
func Const(name string, typ convert.TypeDefinition, val ast.Expr) *ValueDeclBuilder {
	return &ValueDeclBuilder{
		isConst: true,
		name: name,
		typ: typ,
		val: val,
	}
}

// PackageBuilder builds a package (convert.AST)
type PackageBuilder struct {
	name string

	types []convert.TypeDeclaration
	funcs []convert.FuncDeclaration
	vals  []convert.ValueDeclaration
	imports []convert.Import
}

func (b *PackageBuilder) PackageName() convert.Ident { return convert.NewIdent(b.name) }
func (b *PackageBuilder) Types() []convert.TypeDeclaration { return b.types }
func (b *PackageBuilder) Funcs() []convert.FuncDeclaration { return b.funcs }
func (b *PackageBuilder) Values() []convert.ValueDeclaration { return b.vals }
func (b *PackageBuilder) Imports() []convert.Import { return b.imports }

func Package(name string) *PackageBuilder {
	return &PackageBuilder{
		name: name,
	}
}
func (b *PackageBuilder) Import(path string) *PackageBuilder {
	b.imports = append(b.imports, &builtImport{path: path})
	return b
}
func (b *PackageBuilder) ImportAs(name, path string) *PackageBuilder {
	b.imports = append(b.imports, &builtImport{alias: name, path: path})
	return b
}
func (b *PackageBuilder) Declare(decl convert.Declaration) *PackageBuilder {
	switch typedDecl := decl.(type) {
	case convert.TypeDeclaration:
		b.types = append(b.types, typedDecl)
	case convert.FuncDeclaration:
		b.funcs = append(b.funcs, typedDecl)
	case convert.ValueDeclaration:
		b.vals = append(b.vals, typedDecl)
	default:
		// TODO: figure out a better way to communicate this
		panic(fmt.Sprintf("unknown declaration type %T", decl))
	}

	return b
}

// FuncTypeBuilder builds a function type definition
type FuncTypeBuilder struct {
	params []convert.Field
	results []convert.Field
}
func (b *FuncTypeBuilder) Params() []convert.Field { return b.params }
func (b *FuncTypeBuilder) Results() []convert.Field { return b.results }

func Function() *FuncTypeBuilder { return &FuncTypeBuilder{} }

func (b *FuncTypeBuilder) Param(name string, typ convert.TypeDefinition) *FuncTypeBuilder {
	b.params = append(b.params, &builtField{name: name, typ: typ})
	return b
}
func (b *FuncTypeBuilder) Return(name string, typ convert.TypeDefinition) *FuncTypeBuilder {
	b.results = append(b.results, &builtField{name: name, typ: typ})
	return b
}
func (b *FuncTypeBuilder) DeclaredAs(name string) *FuncDeclBuilder {
	return &FuncDeclBuilder{
		name: name,
		typ: b,
	}
}

// StructTypeBuilder builds a struct type definition
type StructTypeBuilder struct { fields []convert.Field }
func Struct() *StructTypeBuilder { return &StructTypeBuilder{} }
func (b *StructTypeBuilder) Fields() []convert.Field { return b.fields }

func (b *StructTypeBuilder) Field(name string, typ convert.TypeDefinition, tag string) *StructTypeBuilder {
	// TODO: field-level doc?
	b.fields = append(b.fields, &builtField{
		name: name,
		typ: typ,
		tag: reflect.StructTag(tag),
	})
	return b
}

type InterfaceTypeBuilder struct { methods []convert.Field }
func Interface() *InterfaceTypeBuilder { return &InterfaceTypeBuilder{} }
func (b *InterfaceTypeBuilder) Methods() []convert.Field { return b.methods }

func (b *InterfaceTypeBuilder) Method(name string, typ convert.FuncTypeDefinition) *InterfaceTypeBuilder {
	// TODO: method-level doc?
	b.methods = append(b.methods, &builtField{
		name: name,
		typ: typ,
	})
	return b
}

// TODO: support for iota and groupings
