EnvMap: Go Code Generation Tools
================================

EnvMap is a library easily traversing and generating Go code.  It's
designed to feel vaguely like compile-time "reflection" (hence the name).

EnvMap wraps the Go AST.  `"pkg/loader".NewLoader` can be used to load
files into a Go AST.  The `"pkg/loader".FromArgs` implements the common
case for loading from a list of files, directories, or standard input
provided as command line arguments.

Once you have a Go AST, you can use `"pkg/convert".FromRaw` to convert it
into the forms defined in EnvMap.  Those forms (as interfaces) as live in
`"pkg/convert"`.

Those interfaces can in turn be implemented (or use the existing builder
implementation) to generate new Go ASTs.  `"pkg/generate".NewASTBuilder`
allows for constructing new Go ASTs from the interfaces in
`"pkg/convert"`.  You can either implement those interfaces yourself, or
use the builder implementations in `"pkg/generate/builder"`.
