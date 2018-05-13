package main

import (
	"os"
	"flag"
	"fmt"

	"go/format"

	"github.com/directxman12/envmap/pkg/loader"
	"github.com/directxman12/envmap/pkg/convert"
	"github.com/directxman12/envmap/pkg/generate"
	. "github.com/directxman12/envmap/pkg/generate/builder"
)

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
		//decls := convert.FromRaw(file)
		convert.FromRaw(file)

		testAST := Package("somepkg").

			Declare(
				Type("someType", Struct()).
					WithDoc("someType is a struct")).

			Declare(
				Type("someOtherType", Struct()).
					WithDoc("someOtherType is a struct")).
			
			Declare(
				Var("PublicVar", convert.NewIdent("int"), nil).
					WithDoc("PublicVar is a public int"))


		builder := generate.NewASTBuilder()
		node := builder.FromAST(testAST)
		fileSet := builder.FileSet()
		//node := builder.FromAST(decls)
		//if err := loader.Format(os.Stdout, node); err != nil {
		if err := format.Node(os.Stdout, fileSet, node); err != nil {
			fmt.Fprintf(os.Stderr, "error formatting: %v", err)
			continue
		}
	}
}
