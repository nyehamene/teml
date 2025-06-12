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

func TestScan_ident(t *testing.T) {
	source := "foo foo_bar foo1 foo-bar"
	expected := token.Ident

	tokens := token.Scan([]byte(source))
	kinds := getKinds(tokens)

	for i, tok := range kinds {
		if expected != tok {
			t.Errorf("expected %s but got %s at %d", expected, tok, i)
		}
	}
}

func TestScan_bracket(t *testing.T) {
	source := "([{}])"
	expected := []token.Kind{
		token.ParanOpen,
		token.BracketOpen,
		token.BraceOpen,
		token.BraceClose,
		token.BracketClose,
		token.ParenClose,
	}

	tokens := token.Scan([]byte(source))
	kinds := getKinds(tokens)

	if diff := cmp.Diff(expected, kinds); diff != "" {
		t.Error(diff)
	}
}

func TestScan_delimiter(t *testing.T) {
	source := ",:/\\"
	expected := []token.Kind{
		token.Comma,
		token.Colon,
		token.FSlash,
		token.BSlash,
	}

	tokens := token.Scan([]byte(source))
	kinds := getKinds(tokens)

	if diff := cmp.Diff(expected, kinds); diff != "" {
		t.Error(diff)
	}
}

func TestScan_quoted_string(t *testing.T) {
	source := `"foo"`
	expected := token.String

	tokens := token.Scan([]byte(source))
	kinds := getKinds(tokens)

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func TestScan_line_string(t *testing.T) {
	source := "-- line 1"
	expected := token.StringLine

	tokens := token.Scan([]byte(source))
	kinds := getKinds(tokens)

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func TestScan_quoted_string_template_string(t *testing.T) {
	source := `"foo \(bar)"`
	expected := token.StringTempl

	tokens := token.Scan([]byte(source))
	kinds := getKinds(tokens)

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func TestScan_line_string_template_string(t *testing.T) {
	source := `-- foo \(bar)`
	expected := token.StringLineTempl

	tokens := token.Scan([]byte(source))
	kinds := getKinds(tokens)

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func getKinds(toks []token.Token) []token.Kind {
	kinds := make([]token.Kind, 0, len(toks))
	for _, tok := range toks {
		kinds = append(kinds, tok.Kind)
	}
	return kinds
}
