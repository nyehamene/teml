package main_test

import (
	_ "embed"
	"testing"

	"github.com/eml-lang/teml/token"
)

//go:embed app.teml
var examplefile []byte

func TestScan(t *testing.T) {
	tokens := token.Scan(examplefile)
	for _, tok := range tokens {
		if tok.Kind == token.Invalid {
			t.Fail()
		}
	}
}
