package ast_test

import (
	"embed"
	"strings"
	"testing"

	"github.com/eml-lang/teml/internal/tests"
)

//go:embed testdata/typecheck
var typecheckerBaseFS embed.FS

// typecheckerBasepath the path to the folder containing test source files
var typecheckerBasepath = "testdata/typecheck"

func TestTypecheckValid(t *testing.T) {
	const stage = tests.StageTypeChecked
	fileFilter := func(filename string) bool { return strings.Contains(filename, "valid") }
	errhandler := tests.NewErrorHandler(stage, false)
	outhandler := func(name string, stage tests.CompilationStage, result any) error { return nil }

	tests.CompileFiles(t, typecheckerBaseFS, typecheckerBasepath, fileFilter, outhandler, errhandler)
}

func TestTypecheckInvalid(t *testing.T) {
	const stage = tests.StageTypeChecked
	fileFilter := func(filename string) bool { return !strings.Contains(filename, "valid") }
	errhandler := tests.NewErrorHandler(stage, true)
	outhandler := func(name string, stage tests.CompilationStage, result any) error { return nil }

	tests.CompileFiles(t, typecheckerBaseFS, typecheckerBasepath, fileFilter, outhandler, errhandler)
}
