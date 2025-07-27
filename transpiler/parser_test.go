package ast_test

import (
	_ "embed"
	"testing"

	parser "github.com/eml-lang/teml/ast"
	transpiler "github.com/eml-lang/teml/transpiler"

	"github.com/eml-lang/teml/token"
)

//go:embed testdata/code.teml
var source []byte

func TestParse(t *testing.T) {
	f := parse(source)

	if f == nil {
		t.Error("source file parser failed unexpectedly")
	}

	if f.HasError() {
		t.Error("parser failed unexpectedly")
	}
}

func parse(src []byte) *transpiler.File {
	toks := token.Scan(src)
	asts := parser.ParseFile(toks)
	if asts.HasError() {
		return nil
	}
	f := transpiler.ParseFile(asts, toks)
	return f
}
