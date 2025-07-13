package ast_test

import (
	"embed"
	"strings"
	"testing"

	"github.com/eml-lang/teml/internal/tests"
)

//go:embed testdata/resolver
var resolverBaseFS embed.FS

// resolverBasepath the path to the folder containing test source path
var resolverBasepath = "testdata/resolver"

func TestResolveValid(t *testing.T) {
	const stage = tests.StageResolved
	fileFilter := func(filename string) bool { return strings.Contains(filename, "valid") }
	errhandler := tests.NewErrorHandler(stage, false)
	outhandler := func(string, string) error { return nil }

	tests.CompileFiles(t, resolverBaseFS, resolverBasepath, fileFilter, outhandler, errhandler)
}

func TestResolveInvalid(t *testing.T) {
	const stage = tests.StageResolved
	fileFilter := func(filename string) bool { return !strings.Contains(filename, "valid") }
	errhandler := tests.NewErrorHandler(stage, true)
	outhandler := func(string, string) error { return nil }

	tests.CompileFiles(t, resolverBaseFS, resolverBasepath, fileFilter, outhandler, errhandler)
}
