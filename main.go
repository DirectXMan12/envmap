package main

import (
	"os"
	"flag"
	"fmt"
	"strings"
	"bytes"

	"github.com/directxman12/envmap/pkg/loader"
	"github.com/directxman12/envmap/pkg/convert"
)

func loadFromArgs(loader loader.Loader, args []string) []error {
	// TODO: load all concurrently

	if len(args) == 0 {
		// load from stdin
		if err := loader.FromReader(os.Stdin, "<standard input>"); err != nil {
			return []error{err}
		}
		return nil
	}

	var allErrs []error
	for _, path := range args {
		info, err := os.Stat(path)
		if err != nil {
			allErrs = append(allErrs, err)
			continue
		}
		if info.IsDir() {
			errs := loader.FromDirectory(path)
			allErrs = append(allErrs, errs...)
			continue
		}
		if err := loader.FromFile(path); err != nil {
			allErrs = append(allErrs, err)
		}
	}

	return allErrs
}

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

func main() {
	flag.Parse()

	loader := loader.NewLoader()
	errs := loadFromArgs(loader, flag.Args())
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

			interfaceName := toPubName(typeDecl.Name().ShortName())

			res := &bytes.Buffer{}
			// TODO: generation facilities
			fmt.Fprintf(res, "type %s interface {\n", interfaceName)
			for _, field := range structDefn.Fields() {
				fmt.Fprintf(res, "\t%s() ", toPubName(field.Name().ShortName()))
				asRawer, isRawable := field.Type().(convert.AsRawAST)
				if !isRawable {
					fmt.Fprintf(res, "<shrug>\n")
					continue
				}
				if err := loader.Format(res, asRawer.ToRawNode()); err != nil {
					fmt.Fprintf(os.Stderr, "error: %v", err)
					return
				}
				fmt.Fprintf(res, "\n")
			}
			fmt.Fprintf(res, "}")

			fmt.Println(string(res.Bytes()))
		}
	}
}
