package main_test

import (
	_ "embed"
	"testing"

	"github.com/eml-lang/teml/ast"
	"github.com/eml-lang/teml/token"
)

//go:embed app.teml
var examplefile []byte

func TestScanParse(t *testing.T) {
	f := token.Scan(examplefile, 0)
	for _, tok := range f.Tokens.Each() {
		if tok.Kind == token.Invalid {
			t.Fail()
		}
	}

	file := ast.ParseFile(f, 0)

	if file.HasError() {
		t.Fail()
	}
}

func BenchmarkScan(b *testing.B) {
	for b.Loop() {
		f := token.Scan(examplefile, 0)
		ast.ParseFile(f, 0)
	}
}

func BenchmarkScanReduceAlloc(b *testing.B) {
	for b.Loop() {
		f := token.Scan(examplefile, token.ReduceAlloc)
		ast.ParseFile(f, token.ReduceAlloc)
	}
}
