package generators

import (
	"strings"
	"fmt"

	"github.com/directxman12/envmap/pkg/traverse"
)

const (
	docGenTagLabel = "envmap:with-docs"
)

func GenDoc(trav *traverse.Traverser) ([]string, []ImportSpec) {
	var gen []string
	for _, decl := range trav.Types().WithTag(docGenTagLabel) {
		_, _, fieldName := decl.Tag(docGenTagLabel)[0].Split()
		typeName := decl.Name()
		objName := typeName
		if strings.HasSuffix(objName, "Builder") {
			objName = objName[:len(objName)-len("Builder")]
		}

		gen = append(gen, fmt.Sprintf(`
// WithDoc attaches documentation to this %[2]s.
func (b *%[1]s) WithDoc(lines ...string) *%[1]s {
	if len(lines) > 0 {
		b.%[3]s.Doc = createDocComment(lines)
	}
	return b
}`, typeName, objName, fieldName))
	}

	return gen, nil
}
