package token_test

import (
	"iter"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/tel-lang/tel"
	"github.com/tel-lang/tel/ast/token"
)

func TestScan_keyword(t *testing.T) {
	source := "package import using component document true false if cond enum"
	expected := []token.Kind{
		token.Package,
		token.Import,
		token.Using,
		token.Component,
		token.Document,
		token.True,
		token.False,
		token.If,
		token.Cond,
		token.Enum,
	}

	f := token.Scan([]byte(source), "test.tel")
	kinds := getKinds(f.Tokens)

	if diff := cmp.Diff(expected, kinds); diff != "" {
		t.Error(diff)
	}
}

func TestScan_ident(t *testing.T) {
	source := "foo foo_bar foo1 foo-bar"
	expected := token.Ident

	f := token.Scan([]byte(source), "test.tel")
	kinds := getKinds(f.Tokens)

	for i, tok := range kinds {
		if expected != tok {
			t.Errorf("expected %s but got %s at %d", expected, tok, i)
		}
	}
}

func TestScan_bracket(t *testing.T) {
	source := "([{}])"
	expected := []token.Kind{
		token.ParenOpen,
		token.BracketOpen,
		token.BraceOpen,
		token.BraceClose,
		token.BracketClose,
		token.ParenClose,
	}

	f := token.Scan([]byte(source), "test.tel")
	kinds := getKinds(f.Tokens)

	if diff := cmp.Diff(expected, kinds); diff != "" {
		t.Error(diff)
	}
}

func TestScan_delimiter(t *testing.T) {
	source := ",:.\\"
	expected := []token.Kind{
		token.Comma,
		token.Colon,
		token.Dot,
		token.BSlash,
	}

	f := token.Scan([]byte(source), "test.tel")
	kinds := getKinds(f.Tokens)

	if diff := cmp.Diff(expected, kinds); diff != "" {
		t.Error(diff)
	}
}

func TestScan_quoted_string(t *testing.T) {
	source := "\"foo\""
	expected := token.String

	f := token.Scan([]byte(source), "test.tel")
	kinds := getKinds(f.Tokens)

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func TestScan_line_string(t *testing.T) {
	source := "-- line 1"
	expected := token.StringLine

	f := token.Scan([]byte(source), "test.tel")
	kinds := getKinds(f.Tokens)

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func TestScan_quoted_string_template_string(t *testing.T) {
	source := "\"foo \\(bar)\""
	expected := token.StringTempl

	f := token.Scan([]byte(source), "test.tel")
	kinds := getKinds(f.Tokens)

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func TestScan_line_string_template_string(t *testing.T) {
	source := "-- foo \\(bar)"
	expected := token.StringLineTempl

	f := token.Scan([]byte(source), "test.tel")
	kinds := getKinds(f.Tokens)

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func TestScan_newline(t *testing.T) {
	source := `

	`
	expected := token.Newline

	f := token.Scan([]byte(source), "test.tel", tel.PreserveNewline)
	kinds := getKinds(f.Tokens)

	if len(kinds) == 0 {
		t.Error("expected newline")
	}

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func TestScan_lexeme(t *testing.T) {
	source := `package foo "foo" -- foo`
	expected := []string{
		"package",
		"foo",
		`"foo"`,
		"-- foo",
	}

	f := token.Scan([]byte(source), "test.tel")
	texts := getTexts(f.Texts())

	if diff := cmp.Diff(expected, texts); diff != "" {
		t.Error(diff)
	}
}

func TestScan_line_string_line(t *testing.T) {
	source := `
	-- line 1
	`
	expected := []token.Kind{
		token.StringLine,
	}

	f := token.Scan([]byte(source), "test.tel")
	kinds := getKinds(f.Tokens)

	if diff := cmp.Diff(expected, kinds); diff != "" {
		t.Error(diff)
	}
}

func TestScan_position_and_length(t *testing.T) {
	source := `package fooooo "foooo" -- foo`
	//         012345678901234567890123456789
	expected := []token.Pos{
		{Offset: 0, Length: 7},
		{Offset: 8, Length: 6},
		{Offset: 15, Length: 7},
		{Offset: 23, Length: 6},
	}

	f := token.Scan([]byte(source), "test.tel")
	poses := make([]token.Pos, 0, len(expected))

	for _, tok := range f.Tokens {
		pos := token.Pos{
			Offset: tok.Offset,
			Length: tok.Length,
		}
		poses = append(poses, pos)
	}

	if diff := cmp.Diff(expected, poses); diff != "" {
		t.Error(diff)
	}
}

func TestScan_line(t *testing.T) {
	source := "package\nfoo 10\ntrue"
	//         0123456.7890.12
	expected := []struct {
		kind   token.Kind
		line   int
		column int
	}{
		{kind: token.Package, line: 1, column: 1},
		{kind: token.Ident, line: 2, column: 1},
		{kind: token.Number, line: 2, column: 5},
		{kind: token.True, line: 3, column: 1},
	}

	f := token.Scan([]byte(source), "test.tel")

	for i, tok := range f.Tokens {
		exp := expected[i]
		if exp.kind != tok.Kind {
			t.Errorf("expected token %v", exp.kind)
			t.Errorf("got token %v", tok.Kind)
		}
		if exp.line != tok.Line {
			t.Errorf("expected line %d", exp.line)
			t.Errorf("got line %d", tok.Line)
		}
	}
}

func TestScan_number(t *testing.T) {
	source := "10 1.0"
	expected := token.Number

	f := token.Scan([]byte(source), "test.tel")
	kinds := getKinds(f.Tokens)

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func TestScan_comment(t *testing.T) {
	source := "; howdy"
	expected := token.Comment

	f := token.Scan([]byte(source), "test.tel", tel.PreserveComment)

	if len(f.Comments) == 0 {
		t.Error("expected comment")
	}

	for i, got := range f.Comments {
		if expected != got.Kind {
			t.Errorf("expected %s but got %v at %d", expected, got, i)
		}
	}
}

func TestScan_newline_after_string_line(t *testing.T) {
	source := `
	-- 1 line
	`
	expected := []token.Kind{
		token.Newline,
		token.Newline,
	}

	f := token.Scan([]byte(source), "test.tel", tel.PreserveNewline)
	nl := getNewlines(f.Tokens)

	if diff := cmp.Diff(expected, nl); diff != "" {
		t.Error(diff)
	}
}

func TestScan_newline_after_comment(t *testing.T) {
	source := `
	; 1 line
	`
	expected := []token.Kind{
		token.Newline,
		token.Newline,
	}

	f := token.Scan([]byte(source), "test.tel", tel.PreserveNewline)
	nl := getKinds(f.Tokens)

	if diff := cmp.Diff(expected, nl); diff != "" {
		t.Error(diff)
	}
}

func getNewlines(s []token.Token) []token.Kind {
	kinds := []token.Kind{}
	for _, tok := range s {
		if tok.Kind != token.Newline {
			continue
		}
		kinds = append(kinds, tok.Kind)
	}
	return kinds
}

func getTexts(s iter.Seq[string]) []string {
	texts := []string{}
	for str := range s {
		texts = append(texts, str)
	}
	return texts
}

func getKinds(s []token.Token) []token.Kind {
	kinds := []token.Kind{}
	for _, tok := range s {
		kinds = append(kinds, tok.Kind)
	}
	return kinds
}
