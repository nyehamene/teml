package ast

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/eml-lang/teml/internal/flags"
	"github.com/eml-lang/teml/internal/source"
)

func TestTypecheckValid(t *testing.T) {
	const succeedOnError = false
	runTypechecker(t, "testdata/typecheck/valid.teml", succeedOnError)
}

func TestTypecheckInvalid(t *testing.T) {
	const succeedOnError = true
	const basepath = "testdata/typecheck"

	entries, err := os.ReadDir(basepath)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			continue
		}
		if strings.Contains(name, "valid") {
			continue
		}

		testdata := path.Join(basepath, name)
		t.Run(testdata, func(t *testing.T) {
			// TODO handle skipped tests
			if strings.Contains(name, "skip") {
				t.Skip()
			}
			runTypechecker(t, testdata, succeedOnError)
		})
	}
}

func runTypechecker(t *testing.T, testdata string, succeedOnError bool) {
	t.Helper()
	t.Log(testdata)

	src, err := source.OpenFile(testdata)
	if err != nil {
		t.Fatal(err)
	}

	t_ast, _ := TypecheckFile(src, flags.PreserveComment)
	if succeedOnError {
		if !t_ast.HasError() {
			t.Fatal("transpiler succeeded unexpectedly")
		}
	} else {
		for _, err := range t_ast.Errors() {
			t.Error(err)
		}
		if t_ast.HasError() {
			t.Fatal("transpiler failed unexpectedly")
		}
	}
}
