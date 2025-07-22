package generator_test

import (
	_ "embed"
	"strings"
	"testing"

	past "github.com/eml-lang/teml/ast"
	gen "github.com/eml-lang/teml/codegen/go/generator"
	ast "github.com/eml-lang/teml/codegen/go/transpiler"
	"github.com/eml-lang/teml/token"
	tast "github.com/eml-lang/teml/transpiler"
)

//go:embed test-component/native-element-attr/source.teml
var source []byte

func TestDebug(t *testing.T) {
	tfile := token.Scan(source, token.PreserveComment|token.ReduceAlloc)
	pfile := past.ParseFile(tfile, 0)

	for _, err := range pfile.Errors {
		t.Error(err)
	}
	if pfile.HasError() {
		t.Fatal()
	}

	afile := tast.ParseFile(pfile, tfile)
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
