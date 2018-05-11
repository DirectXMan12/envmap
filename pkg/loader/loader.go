package loader

import (
	"io"
	"sync"

	"go/ast"
	"go/parser"
	"go/token"
	"go/build"
	"go/format"
)

// Loader knows how to load ASTs from files, directories, or arbitrary text streams.
type Loader interface {
	// FromReader loads source code from the given reader, assigning it the given name.
	FromReader(src io.Reader, name string) error
	// FromFile loads source code from the given file path
	FromFile(filename string) error
	// FromDirectory loads source code from the given directory.
	FromDirectory(directory string) []error
	// Files returns the parsed ASTs.  Do not call further
	// parse functions until this returns.
	Files() []*ast.File
}

type Dumper interface {
	// Format formats the given "node" (as per "go/format".Node),
	// writing the result to dst.
	Format(dst io.Writer, node interface{}) error
}

type Manipulator interface {
	Loader
	Dumper
}

// fileLoader loads content into a set of ast.Files, for use later.
// files are parsed in parallel by FromDirectory, and all From* methods
// may be called concurrently.
type fileLoader struct {
	files []*ast.File

	fileChan chan *ast.File
	waitForFiles sync.WaitGroup

	fileSet *token.FileSet  // doesn't need special locking
	buildContext build.Context
}

func (f *fileLoader) FromReader(srcReader io.Reader, name string) error {
	return f.parse(name, srcReader)
}

func (f *fileLoader) FromFile(filename string) error {
	return f.parse(filename, nil)
}

func (f *fileLoader) FromDirectory(directory string) []error {
	pkginfo, err := f.buildContext.ImportDir(directory, 0 /* No special flags */)
	if _, noGoFound := err.(*build.NoGoError); err != nil && !noGoFound {
		return []error{err}
	}
	// TODO: what do we do about no go found?

	var filenames []string
	filenames = append(filenames, pkginfo.GoFiles...)
	filenames = append(filenames, pkginfo.CgoFiles...)

	// TODO: consider test files and xtest files?

	// parse all the files in parallel
	errChan := make(chan error, len(filenames))
	var wg sync.WaitGroup

	for _, filename := range filenames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			errChan <- f.FromFile(name)
		}(filename)
	}
	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

// parse parses a "file" with the given name.  The source may be nil (to load
// from an actual file) or an io.Reader, string, or []byte.  The actual ast is
// sent over a channel, so this is safe to call concurrently.
func (f *fileLoader) parse(name string, src interface{}) error {
	f.waitForFiles.Add(1)

	file, err := parser.ParseFile(f.fileSet, name, src, parser.ParseComments)
	if err != nil {
		return err
	}
	f.fileChan <- file
	return nil
}

// receiveFiles watches the file channel and
// stores the ASTs it receives.
func (f *fileLoader) receiveFiles() {
	for file := range f.fileChan {
		f.files = append(f.files, file)
		f.waitForFiles.Done()
	}
}

func (f *fileLoader) Files() []*ast.File {
	f.waitForFiles.Wait()
	return f.files
}

func (f *fileLoader) Format(dst io.Writer, node interface{}) error {
	return format.Node(dst, f.fileSet, node)
}

// NewLoader returns a new Loader which can parse files concurrently.
func NewLoader() Manipulator {
	loader := &fileLoader{
		fileSet: token.NewFileSet(),
		buildContext: build.Default,
		fileChan: make(chan *ast.File, 10),
	}

	go loader.receiveFiles()

	return loader
}
