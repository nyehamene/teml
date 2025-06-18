package ast_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/eml-lang/teml/ast"
	"github.com/eml-lang/teml/token"
)

var valid = []string{
	`(package p "a")`,
	`(package p "a") (import i "b")`,
	`(package p "a") (import i "b") (import i "c")`,
	`(package p "a") (import i "b") (import i "c") (using [d e] i) (using f i)`,
	`(package p "a") (import i "b") (using c i) (using d i)`,
	`(package p "a") (import i "b") (using c i)`,
	`(package p "a") (import i "b") (using [c d] i)`,
	`(package p "a") (import i "b") (using c i) (document [])`,
	`(package p "a") (import i "b") (using c i) (component Foo [])`,
	`(package p "a") (import i "b") (document [])`,
	`(package p "a") (import i "b") (component Foo [])`,
	`(package p "a") (document [])`,
	`(package p "a") (document Foo [])`,
	`(package p "a") (document Foo [a: A])`,
	`(package p "a") (document Foo [a: A, b: B])`,
	`(package p "a") (document Foo [a: A,])`,
	`(package p "a") (document Foo [] (div) (div))`,
	`(package p "a") (document Foo [] (div))`,
	`(package p "a") (document Foo [] (div) (div))`,
	`(package p "a") (document Foo [] (div #a{}))`,
	`(package p "a") (document Foo [] (div #a/b{}))`,
	`(package p "a") (document Foo [] (div #a/b/c{}))`,
	`(package p "a") (document Foo [] (foo/div {}))`,
	`(package p "a") (document Foo [] (foo/bar/div {}))`,
	`(package p "a") (document Foo [] (div {a: "b", b: true, c: false, d: 100, e: 10.1}))`,
	`(package p "a") (document Foo [] (div {a: "b" b: true c: false d: 100 e: 10.1}))`,
	`(package p "a") (document Foo [] "foo")`,
	"(package p \"a\") (document Foo [] -- foo\n)",
	"(package p \"a\") (document Foo [] (div) \"foo\" -- foo\n)",
	`(package p "a") (component Foo [])`,
	`(package p "a") (component Foo []) (component Foo [])`,
	`(package p "a") (component Foo [a: A])`,
	`(package p "a") (component Foo [a: A, b: B])`,
	`(package p "a") (component Foo [a: A b: B])`,
	`(package p "a") (component Foo [a: A,])`,
	`(package p "a") (component Foo [] (div))`,
	`(package p "a") (component Foo [] (div) (div))`,
	`(package p "a") (component Foo [] (foo/div {}))`,
	`(package p "a") (component Foo [] (foo/bar/div {}))`,
	`(package p "a") (component Foo [] (div #a{}))`,
	`(package p "a") (component Foo [] (div #a/b{}))`,
	`(package p "a") (component Foo [] (div #a/b/c{}))`,
	`(package p "a") (component Foo [] (div {a: "b", b: true, c: false, d: 100, e: 10.1}))`,
	`(package p "a") (component Foo [] (div {a: "b" b: true c: false d: 100 e: 10.1}))`,
	`(package p "a") (component Foo [] (div (div (div))))`,
	`(package p "a") (component Foo [] (div "foo"))`,
	"(package p \"a\") (component Foo [] (div -- foo\n))",
	"(package p \"a\") (component Foo [] -- foo\n)",
	"(package p \"a\") (component Foo [] (div) \"foo\" -- foo\n)",
	"(package p \"a\") (component Foo [] (div (div) \"foo\" -- foo\n))",
	`(package p "a") (component Foo [] (div {} #a{} "foo"))`,
	`(package p "a") (component Foo [] (div {} "foo" {} (div)))`,
	`(package p "a") (component Foo [] (div {} "foo \(foo)"))`,
	"(package p \"a\") (component Foo [] (div {} -- foo \\(foo)\n))",
	"(package p \"a\") (component Foo [] (div {a: \"foo\\(b)\"} -- foo \\(foo)\n))",
}

func TestParse_short_valid(t *testing.T) {
	for i, source := range valid {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {

			tokens := token.Scan([]byte(source), 0)

			_, hasError := ast.ParseWithErrorHandler(*tokens, 0, func(err string) {
				t.Error(err)
			})

			if hasError {
				t.Error("Parser failed unexpectedly")
			}

		})
	}
}

var invalid = []string{
	// "(package) ;error: DECLARATION ERROR\n;desc: missing identifier",
}

func TestParse_short_invalid(t *testing.T) {
	for i, source := range invalid {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {

			tokens := token.Scan([]byte(source), 0)

			_, hasError := ast.ParseWithErrorHandler(*tokens, 0, func(err string) {
				t.Error(err)
			})

			if hasError {
				t.Error("Parser succeeded unexpectedly")
			}
		})
	}
}

func getKinds(n ast.Node) []token.Kind {
	switch t := n.(type) {
	case ast.Package:
		k := []token.Kind{t.Ident.Kind, t.Path.Kind}
		return k
	}
	return nil
}

func getErrorMessagesFromComments(f token.Tokenized) []string {
	cmts := []string{}
	for _, tok := range f.Tokens() {
		if tok.Kind != token.Comment {
			continue
		}
		cmt, ok := f.Text(tok)
		if !ok {
			continue
		}

		msg := strings.Split(cmt, ": ")[1]
		cmts = append(cmts, msg)
	}
	return cmts
}
