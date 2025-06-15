package main_test

import (
	_ "embed"
	"testing"

	"github.com/eml-lang/teml/token"
)

//go:embed app.teml
var examplefile []byte

func TestScan(t *testing.T) {
	f := token.Scan(examplefile, 0)
	for _, tok := range f.Tokens() {
		if tok.Kind == token.Invalid {
			t.Fail()
		}
	}
}

func BenchmarkScan(b *testing.B) {
	for b.Loop() {
		_ = token.Scan(examplefile, 0)
	}
}

func BenchmarkScanReduceAlloc(b *testing.B) {
	for b.Loop() {
		_ = token.Scan(examplefile, token.ReduceAlloc)
	}
}
