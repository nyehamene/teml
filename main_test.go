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
	for _, tok := range f.Tokens() {
		if tok.Kind == token.Invalid {
			t.Fail()
		}
	}

	_, hasError := ast.ParseWithErrorHandler(*f, 0, func(string) {
		t.Fail()
	})

	if hasError {
		t.Fail()
	}
}

func BenchmarkScan(b *testing.B) {
	for b.Loop() {
		f := token.Scan(examplefile, 0)
		_, _ = ast.Parse(*f, 0)
	}
}

func BenchmarkScanReduceAlloc(b *testing.B) {
	for b.Loop() {
		f := token.Scan(examplefile, token.ReduceAlloc)
		_, _ = ast.Parse(*f, token.ReduceAlloc)
	}
}
