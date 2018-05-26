package generators

import (
	"log"
	"strings"
	"fmt"
	"go/ast"
	"text/template"
	"bytes"

	"github.com/directxman12/envmap/pkg/traverse"
)

const (
	wrapperGenTagLabel = "envmap:gen-wrapper"
)

func firstField(t *traverse.TypeDeclaration) (string, string, bool) {
	structDef, isStruct := t.Spec.Type.(*ast.StructType)
	if !isStruct { return "", "", false }
	if structDef.Fields == nil || len(structDef.Fields.List) != 1 {
		log.Fatalf("cannot use gen-wrapper on a struct with anything other than one embedding")
	}

	field := structDef.Fields.List[0]
	if len(field.Names) != 0 {
		log.Fatalf("cannot use gen-wrapper on a struct with anything other than one embedding")
	}
	astTypePkg, astTypeName, isIdent := traverse.AsQualifiedIdent(field.Type)
	if !isIdent {
		log.Fatalf("cannot use gen-wrapper on a struct with anything other than one embedding")
	}
	return astTypePkg, astTypeName, true
}

type genInfo struct {
	InnerType string
	FuncName string
	TypeName string
	InterfaceName string
	InnerIsPointer bool
	Targets []targetType
}

var (
	structFuncTempl = template.Must(template.New("struct func").Delims("[", "]").Parse(`
func Raw[.InterfaceName](o [.InnerType]) [.TypeName] {
	return [.TypeName]{o}
}`))
	structRenderTempl = template.Must(template.New("struct render").Delims("[", "]").Parse(`
[range .Targets]
[- if (not .IsExtra)]
func (w [$.TypeName]) Render[.Name]() [.QualName] {
	return w.[.Name]
}[- end -]
[- end -]`))
	wrapperRenderTempl = template.Must(template.New("wrapper render").Delims("[", "]").Parse(`
[range .Targets]
[- if (not .IsExtra)]
func (w [$.TypeName]) Render[.Name]() [.QualName] {
[- if $.InnerIsPointer]
	res := [$.InnerType](*w)
	return &res
[- else]
	return [$.InnerType](w)
[- end]
}[- end -]
[- end -]
`))
	interfaceTempl = template.Must(template.New("interface").Delims("[", "]").Parse(`
type [.InterfaceName] interface {[range .Targets]
	Render[.Name]() [.QualName]
[- end]
}`))
)

type targetType struct {
	IsPointer bool
	Pkg string
	Name string
	IsExtra bool
}

func (t targetType) QualName() string {
	if t.Pkg == "" {
		return t.Name
	}
	res := fmt.Sprintf("%s.%s", t.Pkg, t.Name)
	if t.IsPointer {
		return "*"+res
	}
	return res
}

func GenWrappers(trav *traverse.Traverser) ([]string, []ImportSpec) {
	var gen []string
	for _, decl := range trav.Types().WithTag(wrapperGenTagLabel) {
		_, _, kindValRaw := decl.Tag(wrapperGenTagLabel)[0].Split()
		var kindValParts []string
		var baseIsPointer bool
		if len(kindValRaw) > 0 {
			kindValParts = strings.Split(kindValRaw, ",")
			if kindValParts[0] == "pointer" {
				baseIsPointer = true
				kindValParts = kindValParts[1:]
			}
		}

		var targets []targetType
		for _, kindVal := range kindValParts {
			nameParts := strings.SplitN(kindVal, ".", 2)
			isPointer := false
			isExtra := false
			if strings.HasPrefix(nameParts[0], "extra:") {
				isExtra = true
				nameParts[0] = nameParts[0][6:]
			}
			if nameParts[0][0] == '*' {
				isPointer = true
				nameParts[0] = nameParts[0][1:]
			}
			if len(nameParts) == 2 {
				targets = append(targets, targetType{Pkg: nameParts[0], Name: nameParts[1], IsPointer: isPointer, IsExtra: isExtra})
			} else {
				targets = append(targets, targetType{Name: nameParts[0], IsPointer: isPointer, IsExtra: isExtra})
			}
		}

		typeName := decl.Name()
		astTypePkg, astTypeName, isStruct := firstField(&decl)
		var funcTempl, renderTempl *template.Template
		info := genInfo{
			TypeName: typeName,
		}
		if isStruct {
			funcTempl = structFuncTempl
			renderTempl = structRenderTempl
			info.InterfaceName = strings.Title(typeName)
		} else {
			if !strings.HasPrefix(typeName, "Raw") {
				log.Fatalf("cannot generate wrapper for type %q: not of form Raw<somename>", typeName)
			}
			info.InterfaceName = typeName[3:]
			var isIdent bool
			astTypePkg, astTypeName, isIdent = traverse.AsQualifiedIdent(decl.Spec.Type)
			if !isIdent {
				log.Fatalf("cannot generate wrapper for type %q: it's not a wrapper on an ident", typeName)
			}
			renderTempl = wrapperRenderTempl
		}

		targets = append(targets, targetType{Pkg: astTypePkg, Name: astTypeName, IsPointer: baseIsPointer})

		info.InnerType = astTypeName
		if astTypePkg != "" {
			info.InnerType = fmt.Sprintf("%s.%s", astTypePkg, astTypeName)
		}
		if baseIsPointer {
			info.TypeName = "*"+info.TypeName
			info.InnerIsPointer = true
		}
		info.Targets = targets

		var out bytes.Buffer
		if err := interfaceTempl.Execute(&out, info); err != nil {
			log.Fatal(err)
		}
		if funcTempl != nil {
			if err := funcTempl.Execute(&out, info); err != nil {
				log.Fatal(err)
			}
		}
		if err := renderTempl.Execute(&out, info); err != nil {
			log.Fatal(err)
		}
		gen = append(gen, string(out.Bytes()))
	}

	return gen, []ImportSpec{{Path: "go/ast"}}
}
