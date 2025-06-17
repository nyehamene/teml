package ast_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/eml-lang/teml/ast"
	"github.com/eml-lang/teml/token"
	"github.com/google/go-cmp/cmp"
)

var valid_short = []string{
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
	`(package p "a") (component Foo [])`,
	`(package p "a") (component Foo []) (component Foo [])`,
	`(package p "a") (component Foo [a: A])`,
	`(package p "a") (component Foo [a: A, b: B])`,
	`(package p "a") (component Foo [a: A b: B])`,
	`(package p "a") (component Foo [a: A,])`,
	`(package p "a") (component Foo [] (div))`,
	`(package p "a") (component Foo [] (div) (div))`,
	`(package p "a") (component Foo [] (div {a: b}))`,
	`(package p "a") (component Foo [] (div {a: b, c: d}))`,
	`(package p "a") (component Foo [] (div #a{}))`,
	`(package p "a") (component Foo [] (div #a/b{}))`,
	`(package p "a") (component Foo [] (div #a/b/c{a: b}))`,
}

func TestParse_short_test(t *testing.T) {
	for i, source := range valid_short {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			expected := []token.Kind{
				token.Ident,
				token.String,
			}

			tokens := token.Scan([]byte(source), 0)

			f, hasError := ast.ParseWithErrorHandler(*tokens, func(err string) {
				t.Error(err)
			})

			kinds := getKinds(f.Package())

			if hasError {
				t.Error("Parser failed unexpectedly")
			}

			if diff := cmp.Diff(expected, kinds); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestParse_package_redeclared(t *testing.T) {
	t.Skip("unneccessary in a recursive decent parser")

	source := `(package p "does/not/matter")
			   (package i "/does/not/matter")`
	tokens := token.Scan([]byte(source), 0)

	_, hasError := ast.ParseWithErrorHandler(*tokens, func(string) {})

	if !hasError {
		t.Error("Parser succeeded unexpectedly")
	}
}

func TestParse_package_error(t *testing.T) {
	t.Skip("needs error recovery or augument the parser to display only one error per line")

	source := []string{
		"(package) ;error: Package Declaration Error\n;desc: missing identifier",
		"(package p) ;error: Package Declaration error\n;desc: missing path string",
		"(package \"path/to/package\") ;error: Package Declaration error\n;desc: missing identifier",
	}

	for i, src := range source {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			tokens := token.Scan([]byte(src), token.PreserveComment)
			got := []string{}
			errlabels := []string{";error", ";desc"}

			_, hasError := ast.ParseWithErrorHandler(*tokens, func(err string) {
				lines := strings.SplitSeq(err, "\n")
				for ln := range lines {
					chunks := strings.SplitN(ln, ": ", 2)
					if len(chunks) != 2 {
						return
					}
					lbl := chunks[0]
					msg := chunks[1]

					for _, l := range errlabels {
						if l != lbl {
							continue
						}
						got = append(got, msg)
					}
				}
			})

			expected := getErrorMessagesFromComments(*tokens)

			if diff := cmp.Diff(expected, got); diff != "" {
				t.Error(diff)
			}

			if !hasError {
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
