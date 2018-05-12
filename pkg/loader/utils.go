package loader

import (
	"os"
	"sync"
)

// FromArgs is a convinience function to create a new loader, and populate
// it from a set of passed-in command-line arguments.  Loading is done in parallel.
func FromArgs(args []string) (Manipulator, []error) {
	loader := NewLoader()
	if len(args) == 0 {
		// load from stdin
		if err := loader.FromReader(os.Stdin, "<standard input>"); err != nil {
			return loader, []error{err}
		}
		return loader, nil
	}

	errChan := make(chan error, 10)
	var wg sync.WaitGroup

	var allErrs []error
	for _, path := range args {
		info, err := os.Stat(path)
		if err != nil {
			allErrs = append(allErrs, err)
			continue
		}
		wg.Add(1)
		if info.IsDir() {
			go func(p string) {
				defer wg.Done()
				for _, err := range loader.FromDirectory(p) {
					errChan <- err
				}
			}(path)
			continue
		}
		go func(p string) {
			defer wg.Done()
			if err := loader.FromFile(p); err != nil {
				errChan <- err
			}
		}(path)
	}

	go func() {
		defer close(errChan)
		wg.Wait()
	}()

	for err := range errChan {
		allErrs = append(allErrs, err)
	}

	return loader, allErrs
}
