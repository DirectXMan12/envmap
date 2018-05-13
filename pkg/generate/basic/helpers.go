package basic

import (
	"reflect"

	"github.com/directxman12/envmap/pkg/convert"
)

// This file defines "helper" structs which implement interfaces
// from convert, but don't necessarily correspond one-to-one to
// the structure of those interfaces.  This papers over some of
// the quirks of the Go AST.

// NewMethod returns convert.Field that defines an interface method.
func NewInterfaceMethod(name string, params []convert.TypeDefinition, paramNames []string, returns []convert.TypeDefinition, returnNames []string) convert.Field {
	return &method{
		name: name,
		params: params,
		paramNames: paramNames,
		returns: returns,
		returnNames: returnNames,
	}
}

// method defines a field and underlying type definition for an interface method.
type method struct {
	name string

	params []convert.TypeDefinition
	paramNames []string

	returns []convert.TypeDefinition
	returnNames []string
}

func (m *method) Name() convert.Ident { return convert.NewIdent(m.name) }
func (m *method) Type() convert.TypeDefinition { return m }
func (m *method) Tag() reflect.StructTag { return reflect.StructTag("") }
func (m *method) Params() []convert.Field {
	res := make([]convert.Field, len(m.params))
	for i, param := range m.params {
		var name convert.Ident = convert.Anonymous
		if len(m.paramNames) > i && m.returnNames[i] != "" {
			name = convert.NewIdent(m.paramNames[i])
		}
		res[i] = &field{
			typ: param,
			name: name,
		}
	}
	return res
}
func (m *method) Results() []convert.Field {
	res := make([]convert.Field, len(m.returns))
	for i, param := range m.returns {
		var name convert.Ident = convert.Anonymous
		if len(m.returnNames) > i && m.returnNames[i] != "" {
			name = convert.NewIdent(m.returnNames[i])
		}
		res[i] = &field{
			typ: param,
			name: name,
		}
	}
	return res
}

// field represents a basic field
type field struct {
	name convert.Ident
	typ convert.TypeDefinition
	tag reflect.StructTag
}
