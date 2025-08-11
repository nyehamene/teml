package compile

import (
	"os"
	"path"
)

type ErrorSeq = func(yield func(int, error) bool)

func readFile(f string) []byte {
	buf, err := os.ReadFile(f)
	if err != nil {
		panic(err)
	}

	return buf
}

func getFiles(ctx CompilationContext, basepath string) []string {
	entries, err := os.ReadDir(basepath)
	if err != nil {
		panic("could not open testdata dir")
	}

	validFiles := []string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		if n := e.Name(); ctx.filterFunc(n) {
			qname := path.Join(basepath, e.Name())
			validFiles = append(validFiles, qname)
		}
	}

	return validFiles
}

func newErrorSeq(errs []error) ErrorSeq {
	return func(yield func(int, error) bool) {
		for i, err := range errs {
			if !yield(i, err) {
				break
			}
		}
	}
}

func newErrorSeq1(err error) ErrorSeq {
	return func(yield func(int, error) bool) {
		yield(0, err)
	}
}
