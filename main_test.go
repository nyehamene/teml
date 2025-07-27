package main_test

import (
	_ "embed"
	"testing"

	"github.com/eml-lang/teml/token"

	parser "github.com/eml-lang/teml/ast"
	transpiler "github.com/eml-lang/teml/transpiler"
)

//go:embed app.teml
var examplefile []byte

func TestScanParse(t *testing.T) {
	toks := token.Scan(examplefile, "test.teml")
	for _, tok := range toks.Tokens {
		if tok.Kind == token.Invalid {
			t.Fatal()
		}
	}

	astp := parser.ParseFile(toks)
	for _, err := range astp.Errors {
		t.Error(err.Message)
	}

	astn := transpiler.ParseFile(astp)
	for _, err := range astn.Errors() {
		t.Error(err.Message)
	}

	transpiler.ResolveFile(astn)
	for _, err := range astn.Errors() {
		t.Error(err.Message)
	}
}

func BenchmarkScan(b *testing.B) {
	for b.Loop() {
		parseFile()
	}
}

func BenchmarkScanReduceAlloc(b *testing.B) {
	for b.Loop() {
		parseFile()
	}
}

func parseFile() {
	ft := token.Scan(examplefile, "test.teml")
	fa := parser.ParseFile(ft)
	fn := transpiler.ParseFile(fa)
	transpiler.ResolveFile(fn)
}
