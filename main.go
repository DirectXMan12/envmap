package main

import (
	"os"
	"flag"
	"fmt"
	"strings"
	"bytes"
	"reflect"

	"github.com/directxman12/envmap/pkg/loader"
	"github.com/directxman12/envmap/pkg/convert"
	"github.com/directxman12/envmap/pkg/generate"
)

func hasDeriveGetters(comments []string) bool {
	for _, comment := range comments {
		if comment == "+derive:getters" {
			return true
		}
	}
	return false
}

func toPubName(name string) string {
	return strings.ToUpper(name[0:1])+name[1:]
}

type autoGetters struct {
	fieldNames []string
	fieldTypes []convert.TypeDefinition
	structName string
}

func (g *autoGetters) Doc() []string { return nil }
func (g *autoGetters) Name() convert.Ident {
	return convert.NewIdent(toPubName(g.structName))
}
func (g *autoGetters) IsAlias() bool { return false }
func (g *autoGetters) Type() convert.TypeDefinition { return g }
func (g *autoGetters) Methods() []convert.Field {
	res := make([]convert.Field, len(g.fieldNames))
	for i, origName := range g.fieldNames {
		res[i] = &getterMethod{
			name: origName,
			typ: &getterFuncType{g.fieldTypes[i]},
		}
	}
	return res
}

type getterMethod struct {
	name string
	typ convert.TypeDefinition
}
func (g *getterMethod) Doc() []string { return nil }
func (g *getterMethod) Name() convert.Ident {
	if g.name == "" { return convert.Anonymous }
	return convert.NewIdent(toPubName(g.name))
}
func (g *getterMethod) Type() convert.TypeDefinition { return g.typ }
func (g *getterMethod) Tag() reflect.StructTag { return reflect.StructTag("") }

type getterFuncType struct {
	resType convert.TypeDefinition
}
func (g *getterFuncType) Params() []convert.Field { return nil }
func (g *getterFuncType) Results() []convert.Field {
	return []convert.Field{
		&getterMethod{
			name: "",
			typ: g.resType,
		},
	}
}


func main() {
	flag.Parse()

	loader, errs := loader.FromArgs(flag.Args())
	if len(errs) > 0 {
		for _, err := range errs {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		return
	}

	files := loader.Files()

	for _, file := range files {
		decls := convert.FromRaw(file)

		for _, typeDecl := range decls.Types() {
			if !hasDeriveGetters(typeDecl.Doc()) {
				continue
			}

			// TODO: deal with other types
			structDefn, isStruct := typeDecl.Type().(convert.StructTypeDefinition)
			if !isStruct {
				continue
			}

			getterGen := &autoGetters{
				structName: typeDecl.Name().Name(),
			}
			for _, field := range structDefn.Fields() {
				getterGen.fieldNames = append(getterGen.fieldNames, field.Name().Name())
				getterGen.fieldTypes = append(getterGen.fieldTypes, field.Type())
			}

			res := &bytes.Buffer{}
			node := generate.FromTypeDeclaration(getterGen)
			if err := loader.Format(res, node); err != nil {
				fmt.Fprintf(os.Stdout, "error: %v", err)
				continue
			}
			fmt.Println(string(res.Bytes()))
		}
	}
}
