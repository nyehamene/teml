package generator_test

import (
	_ "embed"
	"strings"
	"testing"

	past "github.com/eml-lang/teml/ast"
	gen "github.com/eml-lang/teml/codegen/go/generator"
	ast "github.com/eml-lang/teml/codegen/go/transpiler"
	"github.com/eml-lang/teml/internal/source"
	"github.com/eml-lang/teml/token"
	tast "github.com/eml-lang/teml/transpiler"
)

//go:embed test-template/native-element-attr/source.teml
var content []byte

func TestDebug(t *testing.T) {
	tsrc := source.File{
		Path:    "test.teml",
		Name:    "test",
		Content: content,
	}
	tfile := token.ScanInput(tsrc, token.PreserveComment|token.ReduceAlloc)
	pfile := past.ParseFile(tfile)

	for _, err := range pfile.Errors {
		t.Error(err)
	}
	if pfile.HasError() {
		t.Fatal()
	}

	afile := tast.ParseFile(pfile)
	for _, err := range afile.Errors() {
		t.Error(err)
	}
	if afile.HasError() {
		t.Fatal()
	}

	env := tast.ResolveFile(afile)
	for _, err := range afile.Errors() {
		t.Error(err)
	}
	if afile.HasError() {
		t.Fatal()
	}

	_ = tast.TypecheckFile(afile, env)
	for _, err := range afile.Errors() {
		t.Error(err)
	}
	if afile.HasError() {
		t.Fatal()
	}

	file := ast.Parse(afile)

	var w strings.Builder
	err := gen.Generate(&w, &file)
	if err != nil {
		t.Fatal(err)
	}
}
