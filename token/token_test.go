package token_test

import (
	"iter"
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

	f := token.Scan([]byte(source), 0)
	kinds := getKinds(f.Tokens())

	if diff := cmp.Diff(expected, kinds); diff != "" {
		t.Error(diff)
	}
}

func TestScan_ident(t *testing.T) {
	source := "foo foo_bar foo1 foo-bar"
	expected := token.Ident

	f := token.Scan([]byte(source), 0)
	kinds := getKinds(f.Tokens())

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

	f := token.Scan([]byte(source), 0)
	kinds := getKinds(f.Tokens())

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

	f := token.Scan([]byte(source), 0)
	kinds := getKinds(f.Tokens())

	if diff := cmp.Diff(expected, kinds); diff != "" {
		t.Error(diff)
	}
}

func TestScan_quoted_string(t *testing.T) {
	source := `"foo"`
	expected := token.String

	f := token.Scan([]byte(source), 0)
	kinds := getKinds(f.Tokens())

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func TestScan_line_string(t *testing.T) {
	source := "-- line 1"
	expected := token.StringLine

	f := token.Scan([]byte(source), 0)
	kinds := getKinds(f.Tokens())

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func TestScan_quoted_string_template_string(t *testing.T) {
	source := `"foo \(bar)"`
	expected := token.StringTempl

	f := token.Scan([]byte(source), 0)
	kinds := getKinds(f.Tokens())

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func TestScan_line_string_template_string(t *testing.T) {
	source := `-- foo \(bar)`
	expected := token.StringLineTempl

	f := token.Scan([]byte(source), 0)
	kinds := getKinds(f.Tokens())

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

	f := token.Scan([]byte(source), token.PreserveNewline)
	kinds := getKinds(f.Tokens())

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

	f := token.Scan([]byte(source), 0)
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

	f := token.Scan([]byte(source), 0)
	kinds := getKinds(f.Tokens())

	if diff := cmp.Diff(expected, kinds); diff != "" {
		t.Error(diff)
	}
}

func TestScan_position(t *testing.T) {
	source := `package fooooo "foooo" -- foo`
	//         012345678901234567890123456789
	expected := []token.Pos{
		{0, 7},
		{8, 14},
		{15, 22},
		{23, 29},
	}

	f := token.Scan([]byte(source), 0)
	pos := getPosses(f.Posses())

	if diff := cmp.Diff(expected, pos); diff != "" {
		t.Error(diff)
	}
}

func getPosses(s iter.Seq[token.Pos]) []token.Pos {
	pos := []token.Pos{}
	for p := range s {
		pos = append(pos, p)
	}
	return pos
}

func TestScan_line(t *testing.T) {
	source := "package\nfoo\n"
	//         0123456.7890.12
	expected := []int{7, 11}

	f := token.Scan([]byte(source), 0)
	lines := getLines(f.Lines())

	if diff := cmp.Diff(expected, lines); diff != "" {
		t.Error(diff)
	}
}

func TestScan_number(t *testing.T) {
	source := "10 1.0"
	expected := token.Number

	f := token.Scan([]byte(source), 0)
	kinds := getKinds(f.Tokens())

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
		}
	}
}

func TestScan_comment(t *testing.T) {
	source := "; howdy"
	expected := token.Comment

	f := token.Scan([]byte(source), token.PreserveComment)
	kinds := getKinds(f.Tokens())

	if len(kinds) == 0 {
		t.Error("expected comment")
	}

	for i, got := range kinds {
		if expected != got {
			t.Errorf("expected %s but got %s at %d", expected, got, i)
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

	f := token.Scan([]byte(source), token.PreserveNewline)
	nl := getNewlines(f.Tokens())

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

	f := token.Scan([]byte(source), token.PreserveNewline)
	nl := getKinds(f.Tokens())

	if diff := cmp.Diff(expected, nl); diff != "" {
		t.Error(diff)
	}
}

func getNewlines(s iter.Seq2[int, token.Token]) []token.Kind {
	kinds := []token.Kind{}
	for _, tok := range s {
		if tok.Kind != token.Newline {
			continue
		}
		kinds = append(kinds, tok.Kind)
	}
	return kinds
}

func getLines(s iter.Seq[int]) []int {
	lines := []int{}
	for line := range s {
		lines = append(lines, line)
	}
	return lines
}

func getTexts(s iter.Seq[string]) []string {
	texts := []string{}
	for str := range s {
		texts = append(texts, str)
	}
	return texts
}

func getKinds(s iter.Seq2[int, token.Token]) []token.Kind {
	kinds := []token.Kind{}
	for _, tok := range s {
		kinds = append(kinds, tok.Kind)
	}
	return kinds
}
