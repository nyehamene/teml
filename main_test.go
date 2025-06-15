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
	for tok := range f.Tokens() {
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

func BenchmarkScanCountFirst(b *testing.B) {
	for b.Loop() {
		_ = token.ScanCountFirst(examplefile, 0)
	}
}
