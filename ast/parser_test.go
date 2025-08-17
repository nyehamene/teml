package ast_test

import (
	"fmt"
	"iter"
	"strings"
	"testing"

	"github.com/eml-lang/teml/ast"
	"github.com/eml-lang/teml/internal/flags"
	"github.com/eml-lang/teml/internal/source"
)

var valid = []string{
	`(package p "a")`,
	`(package p "a") (import i "b")`,
	`(package p "a") (import i "b") (import i "c")`,
	`(package p "a") (import i "b") (import i "c") (using [d e] i) (using f i)`,
	`(package p "a") (import i "b") (using c i) (using d i)`,
	`(package p "a") (import i "b") (using c i)`,
	`(package p "a") (import i "b") (using [c d] i)`,
	`(package p "a") (import i "b") (using [c, d] i)`,
	`(package p "a") (import i "b") (using [c,] i)`,
	`(package p "a") (import i "b") (using c i) (document [])`,
	`(package p "a") (import i "b") (using c i) (component Foo [])`,
	`(package p "a") (import i "b") (document [])`,
	`(package p "a") (import i "b") (component Foo [])`,
	`(package p "a") (document [])`,
	`(package p "a") (document Foo [])`,
	`(package p "a") (document Foo [a: A])`,
	`(package p "a") (document Foo [a: A, b: B])`,
	`(package p "a") (document Foo [a: A,])`,
	`(package p "a") (document Foo [a: a.A])`,
	`(package p "a") (document Foo [a: a.A, b: b.B])`,
	`(package p "a") (document Foo [] (div) (div))`,
	`(package p "a") (document Foo [] (div))`,
	`(package p "a") (document Foo [] (div) (div))`,
	`(package p "a") (document Foo [] (div #a{}))`,
	`(package p "a") (document Foo [] (div #a.b{}))`,
	`(package p "a") (document Foo [] (div #a.b.c{}))`,
	`(package p "a") (document Foo [] (foo.div {}))`,
	`(package p "a") (document Foo [] (foo.bar.div {}))`,
	`(package p "a") (document Foo [] (div {a: "b", b: true, c: false, d: 100, e: 10}))`,
	`(package p "a") (document Foo [] (div {a: "b" b: true c: false d: 100 e: 10}))`,
	`(package p "a") (document Foo [] "foo")`,
	"(package p \"a\") (document Foo [] -- foo\n)",
	"(package p \"a\") (document Foo [] (div) \"foo\" -- foo\n)",
	`(package p "a") (component Foo [])`,
	`(package p "a") (component Foo []) (component Foo [])`,
	`(package p "a") (component Foo [a: A])`,
	`(package p "a") (component Foo [a: A, b: B])`,
	`(package p "a") (component Foo [a: A b: B])`,
	`(package p "a") (component Foo [a: A,])`,
	`(package p "a") (component Foo [a: a.A])`,
	`(package p "a") (component Foo [] (div))`,
	`(package p "a") (component Foo [] (div) (div))`,
	`(package p "a") (component Foo [] (foo.div {}))`,
	`(package p "a") (component Foo [] (foo.bar.div {}))`,
	`(package p "a") (component Foo [] (div #a{}))`,
	`(package p "a") (component Foo [] (div {a: "b", b: true, c: false, d: 100, e: 10}))`,
	`(package p "a") (component Foo [] (div {a: "b" b: true c: false d: 100 e: 10}))`,
	`(package p "a") (component Foo [] (div (div (div))))`,
	`(package p "a") (component Foo [] (div "foo"))`,
	"(package p \"a\") (component Foo [] (div -- foo\n))",
	"(package p \"a\") (component Foo [] -- foo\n)",
	"(package p \"a\") (component Foo [] (div) \"foo\" -- foo\n)",
	"(package p \"a\") (component Foo [] (div (div) \"foo\" -- foo\n))",
	`(package p "a") (component Foo [] (div {} #a{} "foo"))`,
	`(package p "a") (component Foo [] (div {} "foo" {} (div)))`,
	`(package p "a") (component Foo [] (div {} "foo \\(foo)"))`,
	"(package p \"a\") (component Foo [] (div {} -- foo \\\\(foo)\n))",
	"(package p \"a\") (component Foo [] (div {a: \"foo\\\\(b)\"} -- foo \\\\(foo)\n))",
	`(package p "path") (component C [a: (enum "A" "B")])`,
	`(package p "path") (component C [a: (enum "A", "B")])`,
	`(package p "path") (component C [a: (enum 0 1)])`,
	`(package p "path") (document C [] (if true (div)))`,
	`(package p "path") (component C [] (if f (div) (span)))`,
	`(package p "path") (document C [] (cond f true: (div)))`,
	`(package p "path") (document C [] (cond f true: (div) false: (span) 0: (div) "foo": (span)))`,
	`(package p "path") (component C [] (cond f true: (div), false: (span), 0: (div), "foo": (span),))`,
	`(package p "path") (component C [] (div {a: (if true "foo")}))`,
	`(package p "path") (document [] (div {a: (if f "foo" 100)}))`,
	`(package p "path") (component C [] (div {a: (cond f true: "foo")}))`,
	`(package p "path") (document [] (div {a: (cond f true: "foo" false: 100)}))`,
	`(package p "path") (document [] (div {a: (cond f true: "foo", false: 100, 0: true, "foo": "bar",)}))`,
	`(package p "path") (component A[] (B [name: "foo"]))`,
}

func TestParse_short_valid(t *testing.T) {
	for i, src := range valid {
		t.Logf("(%d) %s", i, src)
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			fsrc := source.NewFile("valid.teml", []byte(src))
			file := ast.ParseFile(fsrc)
			for _, err := range file.Errors {
				t.Error(err)
			}
			if file.HasError() {
				t.Fatal("parser failed unexpectedly")
			}
		})
	}
}

var invalid = []string{
	"(package) ;desc: missing identifier",
	"(package p) ;desc: missing package path",
	`(package p "" ;desc: missing closing parenthesis`,
	`(import) ;desc: missing identifier`,
	`(import i) ;desc: missing import path`,
	"(using) ;desc: missing identifier",
	"(using i) ;desc: missing identifier",
	"(using []) ;desc: empty import alias list",
	"(using [i) ;desc: missing closing bracket",
	"(using ]) ;desc: missing identifier",
	"(document) ;desc: missing opening bracket",
	"(document [) ;desc: missing closing bracket",
	"(document [a]) ;desc: missing type separator",
	"(document [a:]) ;desc: missing property type",
	"(document [] (div {)) ;desc: missing closing brace",
	"(document [] (div #{})) ;desc: missing identifier",
	"(document [] (div #a.{})) ;desc: missing identifier",
	"(document [] (div {a})) ;desc: missing attribute value separator",
	"(document [] (div {a:})) ;desc: missing expression",
	"(document [] (div {a: -- foo\n})) ;desc: line string literal is not a valid expression",
	"(document [] (div {a: -- foo\\(foo)\n})) ;desc: line string literal is not a valid expression",
	"(document [] (div a)) ;desc: identifier is not a valid template content",
	"(component) ;desc: missing identifier",
	"(component A) ;desc: missing opening bracket",
	"(component A [) ;desc: missing closing bracket",
	"(component A [a]) ;desc: missing type separator",
	"(component A [a:]) ;desc: missing property type",
	"(component F [] (div {)) ;desc: missing closing brace",
	"(component F [] (div #{})) ;desc: missing identifier",
	"(component F [] (div #a.{})) ;desc: missing identifier",
	"(component F [] (div #a.b{})) ;desc: qualified tagged attributes not allowed in a component",
	"(component F [] (div {a})) ;desc: missing attribute value separator",
	"(component F [] (div {a:})) ;desc: missing expression",
	"(component F [] (div {a: -- foo\n})) ;desc: line string literal is not a valid expression",
	"(component F [] (div {a: -- foo\\(foo)\n})) ;desc: line string literal is not a valid expression",
	"(component F [] (div a)) ;desc: identifier is not a valid template content",
	"(div) ;desc: unexpected element declaration",
	"foo ;desc: missing opening parenthesis",
	`(import i "path") ;desc: missing package declaration`,
	`(package p "path") (using a i) ;desc: missing import declaration`,
	`(package p "path") (import i "path") (document []) (using a i) ;desc: unexpected using declaration`,
	`(package p "path") (document []) (document []) ;desc: duplicate document declaration`,
	`(package p "path") (component C [a: (enum "A" 10)]) ;desc: mismatch enum constant type`,
	`(package p "path") (component A[] (B ["name": "foo"])) ;desc: missing identifier`,
	`(package p "path") (component A[] (B [name: "foo")) ;desc: unterminated element parameters`,
	`(package p "path") (component A[] (B [name "foo"])) ;desc: missing parameter value separator`,
	`(package p "a") (component Foo [] (div #a.b{})) ;desc: qualified tagged attributes not allowed in a component`,
	`(package p "a") (component Foo [] (div #a.b.c{})) ;desc: qualified tagged attributes not allowed in a component`,
}

func TestParse_short_invalid(t *testing.T) {
	for i, src := range invalid {
		t.Logf("(%d) %s", i, src)
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {

			fsrc := source.NewFile("invalid.teml", []byte(src))
			goterrmsgs := map[string]string{}
			file := ast.ParseFile(fsrc, flags.PreserveComment, flags.ExitOnError)

			// collect error messages
			for _, err := range file.Errors {
				for e := range getEntriesFromString(err.Error()) {
					goterrmsgs[e.key] = e.value
				}
			}

			checkedAtLeastOneError := false
			for e := range getErrorMessagesFromComment(file.Comments) {
				got := goterrmsgs[e.key]
				expected := e.value
				if !strings.HasPrefix(got, expected) {
					// if expected := e.value; expected != got {
					t.Errorf("expected %s", expected)
					t.Errorf("got %s", got)
				}
				checkedAtLeastOneError = true
			}

			if !checkedAtLeastOneError {
				t.Fail()
			}
		})
	}
}

func TestPreserveComment(t *testing.T) {
	sourcecode := `(package p "views") ;; preserve comment`
	srcfile := source.NewFile("test.teml", []byte(sourcecode))

	t.Run("preserve comment when token.PreserveComment flag is given", func(t *testing.T) {
		const expectedComments = 1
		const expectedComment = ";; preserve comment"

		file := ast.ParseFile(srcfile, flags.PreserveComment)
		if got := len(file.Comments); got != expectedComments {
			t.Fatalf("expected %d comment(s) got %d", expectedComments, got)
		}

		cmt := file.Comments[0].Text
		if cmt != expectedComment {
			t.Fatalf("expected comment %q got %q", expectedComment, cmt)
		}
	})

	t.Run("do not preserve comment if token.PreserveComment flag is not given", func(t *testing.T) {
		const expectedComments = 0

		file := ast.ParseFile(srcfile)
		if got := len(file.Comments); got != expectedComments {
			t.Fatalf("expected %d comment(s) got %d", expectedComments, got)
		}
	})
}

type entry[T any] struct {
	key   string
	value T
}

func getErrorMessagesFromComment(cmts []ast.Comment) iter.Seq[entry[string]] {
	return func(yield func(entry[string]) bool) {
		for _, cmt := range cmts {
			for e := range getEntriesFromString(cmt.Text) {
				if !yield(e) {
					break
				}
			}
		}
	}
}

func getEntriesFromString(s string) iter.Seq[entry[string]] {
	return func(yield func(entry[string]) bool) {
		lines := strings.SplitSeq(s, "\n")
		for line := range lines {

			chunks := strings.SplitN(line, ": ", 2)
			key := chunks[0]
			value := chunks[1]

			// remove comment delimiter ;
			key = key[1:]

			e := entry[string]{key: key, value: value}

			if !yield(e) {
				break
			}
		}
	}
}
