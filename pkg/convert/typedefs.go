package convert

import (
	"fmt"
	"reflect"

	"go/ast"
)

// fieldListToFields converts an ast.FieldList into a list Fields
func fieldListToFields(l *ast.FieldList) []Field {
	if l == nil {
		return nil
	}
	var res []Field
	for _, rawField := range l.List {
		if rawField.Names == nil {
			res = append(res, &field{
				name: nil,
				field: rawField,
			})
		}

		for _, name := range rawField.Names {
			res = append(res, &field{
				name: name,
				field: rawField,
			})
		}
	}
	return res
}

func checkBackingTypeDecl(obj *ast.Object) ast.Expr {
	if obj == nil {
		return nil
	}

	typeSpec, isTypeSpec := obj.Decl.(*ast.TypeSpec)
	if !isTypeSpec {
		return nil
	}
	return typeSpec.Type
}

// exprToTypeDefinition converts an expression into one of the
// type definition structs.
func exprToTypeDefinition(expr ast.Expr) TypeDefinition {
	switch typed := expr.(type) {
	case *ast.StructType:
		return &structTypeDefinition{
			typ: typed,
		}
	case *ast.InterfaceType:
		return &interfaceTypeDefinition{
			typ: typed,
		}
	case *ast.FuncType:
		return &funcTypeDefinition{
			typ: typed,
		}
	case *ast.MapType:
		return &mapTypeDefinition{
			typ: typed,
		}
	case *ast.ArrayType:
		return &arrayTypeDefinition{
			typ: typed,
		}
	case *ast.ChanType:
		return &chanTypeDefinition{
			typ: typed,
		}
	case *ast.Ident:
		id := unqualifiedIdent(typed.Name)
		typDecl := checkBackingTypeDecl(typed.Obj)
		if typDecl != nil {
			return typeIdent{
				Ident: id,
				typDecl: typDecl,
			}
		}
		return id
	case *ast.ParenExpr:
		// ParenExpr is just parens around a normal type
		return exprToTypeDefinition(typed.X)
	case *ast.SelectorExpr:
		// SelectorExpr is just a qualified name
		// TODO: allow qualified locatable idents?
		return qualifiedIdent{
			packageName: typed.X.(*ast.Ident).Name,
			Ident: unqualifiedIdent(typed.Sel.Name),
		}
	case *ast.StarExpr:
		// StarExpr is just a pointer to another type
		return &pointerTypeDefinition{
			typ: typed,
		}
	case *ast.Ellipsis:
		return &splatTypeDefinition{
			typ: typed,
		}
	default:
		// TODO: return error instead of panic
		panic(fmt.Sprintf("unknown/invalid expression type %T -- %#v", expr))
	}
}

// structTypeDefinition represents the type definition for a struct (fields, etc)
type structTypeDefinition struct {
	typ *ast.StructType
}
// TODO: what does incomplete mean in ast.StructType

func (d *structTypeDefinition) Fields() []Field {
	return fieldListToFields(d.typ.Fields)
}
func (d *structTypeDefinition) ToRawNode() interface{} {
	return d.typ
}

type interfaceTypeDefinition struct {
	typ *ast.InterfaceType
}

func (d *interfaceTypeDefinition) Methods() []Field {
	return fieldListToFields(d.typ.Methods)
}
func (d *interfaceTypeDefinition) ToRawNode() interface{} {
	return d.typ
}

type funcTypeDefinition struct {
	typ *ast.FuncType
}

func (d *funcTypeDefinition) Params() []Field {
	return fieldListToFields(d.typ.Params)
}

func (d *funcTypeDefinition) Results() []Field {
	return fieldListToFields(d.typ.Results)
}
func (d *funcTypeDefinition) ToRawNode() interface{} {
	return d.typ
}

type mapTypeDefinition struct {
	typ *ast.MapType
}

func (d *mapTypeDefinition) KeyType() TypeDefinition {
	return exprToTypeDefinition(d.typ.Key)
}

func (d *mapTypeDefinition) ValueType() TypeDefinition {
	return exprToTypeDefinition(d.typ.Value)
}
func (d *mapTypeDefinition) ToRawNode() interface{} {
	return d.typ
}

type arrayTypeDefinition struct {
	typ *ast.ArrayType
}

func (d *arrayTypeDefinition) IsSlice() bool {
	return d.typ.Len == nil
}

func (d *arrayTypeDefinition) AutoLength() bool {
	_, isEllipsis := d.typ.Len.(*ast.Ellipsis)
	return isEllipsis
}

func (d *arrayTypeDefinition) Length() *int {
	if d.typ.Len == nil {
		return nil
	}
	if _, isEllipsis := d.typ.Len.(*ast.Ellipsis); isEllipsis {
		return &AutoLength
	}
	panic("not implemented")
}

func (d *arrayTypeDefinition) ElemType() TypeDefinition {
	return exprToTypeDefinition(d.typ.Elt)
}
func (d *arrayTypeDefinition) ToRawNode() interface{} {
	return d.typ
}

type chanTypeDefinition struct {
	typ *ast.ChanType
}

func (d *chanTypeDefinition) ValueType() TypeDefinition {
	return exprToTypeDefinition(d.typ.Value)
}

func (d *chanTypeDefinition) Directions() (receive bool, send bool) {
	return d.typ.Dir & ast.SEND != 0, d.typ.Dir & ast.RECV != 0
}
func (d *chanTypeDefinition) ToRawNode() interface{} {
	return d.typ
}

type pointerTypeDefinition struct {
	typ *ast.StarExpr
}

func (d *pointerTypeDefinition) ReferentType() TypeDefinition {
	return exprToTypeDefinition(d.typ.X)
}

func (d *pointerTypeDefinition) ToRawNode() interface{} {
	return d.typ
}

type splatTypeDefinition struct {
	typ *ast.Ellipsis
}

func (d *splatTypeDefinition) ElemType() TypeDefinition {
	return exprToTypeDefinition(d.typ.Elt)
}
func (d *splatTypeDefinition) IsSplat() struct{} {
	return struct{}{}
}

type field struct {
	field *ast.Field
	name *ast.Ident
}

func (f *field) Doc() []string {
	// TODO: line comments?
	return extractCommentGroup(f.field.Doc)
}

// Name returns the name of the field, or nil for
// an anonymous field.
func (f *field) Name() Ident {
	if f.name == nil {
		return nil
	}

	return unqualifiedIdent(f.name.Name)
}

// Type returns the type of the field.
func (f *field) Type() TypeDefinition {
	return exprToTypeDefinition(f.field.Type)
}

func (f *field) Tag() reflect.StructTag {
	if f.field.Tag == nil {
		return reflect.StructTag("")
	}
	return reflect.StructTag(f.field.Tag.Value)
}

func (f *field) ToRawNode() interface{} {
	// we can't just return the underlying field,
	// because it might have been shared.
	return &ast.Field{
		Doc: f.field.Doc,
		Names: []*ast.Ident{f.name},
		Type: f.field.Type,
		Tag: f.field.Tag,
		Comment: f.field.Comment,
	}
}
