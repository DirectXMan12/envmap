package generators

import (
	"fmt"
	"log"
	"strings"

	"github.com/directxman12/envmap/pkg/traverse"
)

const (
	renderGenTagLabel = "envmap:render"
)

func GenRender(trav *traverse.Traverser) ([]string, []ImportSpec) {
	var gen []string
	for _, decl := range trav.Types().WithTag(renderGenTagLabel) {
		for _, tag := range decl.Tag(renderGenTagLabel) {
			typeName := decl.Name()

			_, _, infoRaw := tag.Split()
			infoParts := strings.SplitN(infoRaw, ",", 2)
			if len(infoParts) < 2{
				log.Fatalf("must specify arguments for render gen as `fieldName,typeName` for type %s", typeName)
			}
			fieldName := infoParts[0]
			retType := infoParts[1]
			dotIndex := strings.IndexByte(retType, '.')
			if dotIndex < 0 {
				log.Fatalf("must specify a type name for render gen of `pkg.Type` for type %s", typeName)
			}
			targetName := string(retType[dotIndex+1:])

			gen = append(gen, fmt.Sprintf(`
func (b *%[1]s) Render%[2]s() %[3]s {
	return b.%[4]s
}`, typeName, targetName, retType, fieldName))
		}
	}

	return gen, nil
}

