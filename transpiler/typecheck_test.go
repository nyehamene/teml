package ast_test

import (
	"embed"
	"testing"
)

//go:embed testdata/typecheck
var typecheckerBaseFS embed.FS

// typecheckerBasepath the path to the folder containing test source files
var typecheckerBasepath = "testdata/typecheck"

func TestTypecheckValid(t *testing.T) {
	checkValidFiles(t, typecheckerBaseFS, typecheckerBasepath, TypecheckerStage)
}

func TestTypecheckInvalid(t *testing.T) {
	checkInvalidFiles(t, typecheckerBaseFS, typecheckerBasepath, TypecheckerStage)
}
