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
	toks := token.Scan(examplefile, 0)
	for _, tok := range toks.Tokens {
		if tok.Kind == token.Invalid {
			t.Fatal()
		}
	}

	astp := parser.ParseFile(toks, 0)
	for _, err := range astp.Errors {
		t.Error(err.Message)
	}

	astn := transpiler.ParseFile(astp, toks)
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
	ft := token.Scan(examplefile, 0)
	fa := parser.ParseFile(ft, 0)
	fn := transpiler.ParseFile(fa, ft)
	transpiler.ResolveFile(fn)
}
