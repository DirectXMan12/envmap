package generators

import (
	"github.com/directxman12/envmap/pkg/traverse"
)

type ImportSpec struct {
	Path string
}

type Generator func(t *traverse.Traverser) (generated []string, imports []ImportSpec)
