package token_test

import (
	"testing"

	"github.com/eml-lang/teml/token"
	"github.com/google/go-cmp/cmp"
)

func TestScan_keyword(t *testing.T) {
	source := "package import using component document and or not true false if"
	expected := []token.Kind{
		token.Package,
		token.Import,
		token.Using,
		token.Component,
		token.Document,
		token.And,
		token.Or,
		token.Not,
		token.True,
		token.False,
		token.If,
	}

	tokens := token.Scan([]byte(source))
	kinds := getKinds(tokens)

	if diff := cmp.Diff(expected, kinds); diff != "" {
		t.Error(diff)
	}
}

func getKinds(toks []token.Token) []token.Kind {
	kinds := make([]token.Kind, 0, len(toks))
	for _, tok := range toks {
		kinds = append(kinds, tok.Kind)
	}
	return kinds
}
