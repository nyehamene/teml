package ast

import (
	"fmt"
	"testing"

	"github.com/tel-lang/tel"
)

var valid = []string{
	`(package pp "a")`,
	`(package pp "a") (import i "b")`,
	`(package pp "a") (import i "b") (import i "c")`,
	`(package pp "a") (import i "b") (import i "c") (using [d e] i) (using f i)`,
	`(package pp "a") (import i "b") (using c i) (using d i)`,
	`(package pp "a") (import i "b") (using c i)`,
	`(package pp "a") (import i "b") (using [c d] i)`,
	`(package pp "a") (import i "b") (using [c, d] i)`,
	`(package pp "a") (import i "b") (using [c,] i)`,
	`(package pp "a") (import i "b") (using c i) (document [])`,
	`(package pp "a") (import i "b") (using c i) (component Foo [])`,
	`(package pp "a") (import i "b") (document [])`,
	`(package pp "a") (import i "b") (component Foo [])`,
	`(package pp "a") (document [])`,
	`(package pp "a") (document Foo [])`,
	`(package pp "a") (document Foo [a: A])`,
	`(package pp "a") (document Foo [a: A, b: B])`,
	`(package pp "a") (document Foo [a: A,])`,
	`(package pp "a") (document Foo [a: a.A])`,
	`(package pp "a") (document Foo [a: a.A, b: b.B])`,
	`(package pp "a") (document Foo [] (div) (div))`,
	`(package pp "a") (document Foo [] (div))`,
	`(package pp "a") (document Foo [] (div) (div))`,
	`(package pp "a") (document Foo [] (div #a{}))`,
	`(package pp "a") (document Foo [] (div #a.b{}))`,
	`(package pp "a") (document Foo [] (div #a.b.c{}))`,
	`(package pp "a") (document Foo [] (foo.div {}))`,
	`(package pp "a") (document Foo [] (foo.bar.div {}))`,
	`(package pp "a") (document Foo [] (div {a: "b", b: true, c: false, d: 100, e: 10}))`,
	`(package pp "a") (document Foo [] (div {a: "b" b: true c: false d: 100 e: 10}))`,
	`(package pp "a") (document Foo [] "foo")`,
	"(package pp \"a\") (document Foo [] -- foo\n)",
	"(package pp \"a\") (document Foo [] (div) \"foo\" -- foo\n)",
	`(package pp "a") (component Foo [])`,
	`(package pp "a") (component Foo []) (component Foo [])`,
	`(package pp "a") (component Foo [a: A])`,
	`(package pp "a") (component Foo [a: A, b: B])`,
	`(package pp "a") (component Foo [a: A b: B])`,
	`(package pp "a") (component Foo [a: A,])`,
	`(package pp "a") (component Foo [a: a.A])`,
	`(package pp "a") (component Foo [] (div))`,
	`(package pp "a") (component Foo [] (div) (div))`,
	`(package pp "a") (component Foo [] (foo.div {}))`,
	`(package pp "a") (component Foo [] (foo.bar.div {}))`,
	`(package pp "a") (component Foo [] (div #a{}))`,
	`(package pp "a") (component Foo [] (div {a: "b", b: true, c: false, d: 100, e: 10}))`,
	`(package pp "a") (component Foo [] (div {a: "b" b: true c: false d: 100 e: 10}))`,
	`(package pp "a") (component Foo [] (div (div (div))))`,
	`(package pp "a") (component Foo [] (div "foo"))`,
	"(package pp \"a\") (component Foo [] (div -- foo\n))",
	"(package pp \"a\") (component Foo [] -- foo\n)",
	"(package pp \"a\") (component Foo [] (div) \"foo\" -- foo\n)",
	"(package pp \"a\") (component Foo [] (div (div) \"foo\" -- foo\n))",
	`(package pp "a") (component Foo [] (div {} #a{} "foo"))`,
	`(package pp "a") (component Foo [] (div {} "foo" {} (div)))`,
	`(package pp "a") (component Foo [] (div {} "foo \\(foo)"))`,
	"(package pp \"a\") (component Foo [] (div {} -- foo \\\\(foo)\n))",
	"(package pp \"a\") (component Foo [] (div {a: \"foo\\\\(b)\"} -- foo \\\\(foo)\n))",
	`(package pp "path") (enum E [A B])`,
	`(package pp "path") (enum E [A, B])`,
	`(package pp "path") (document C [] (if true (div)))`,
	`(package pp "path") (component C [] (if f (div) (span)))`,
	`(package pp "path") (document C [] (cond f true: (div)))`,
	`(package pp "path") (document C [] (cond f true: (div) false: (span) 0: (div) "foo": (span)))`,
	`(package pp "path") (component C [] (cond f true: (div), false: (span), 0: (div), "foo": (span),))`,
	`(package pp "path") (component C [] (div {a: (if true "foo")}))`,
	`(package pp "path") (document [] (div {a: (if f "foo" 100)}))`,
	`(package pp "path") (component C [] (div {a: (cond f true: "foo")}))`,
	`(package pp "path") (document [] (div {a: (cond f true: "foo" false: 100)}))`,
	`(package pp "path") (document [] (div {a: (cond f true: "foo", false: 100, 0: true, "foo": "bar",)}))`,
	`(package pp "path") (component A[] (B [name: "foo"]))`,
	`(package pp "path") (using X x)`,
	`(package pp "path") (using X x.x)`,
	`(package pp "path") (using [X] x)`,
	`(package pp "path") (using [X] x.x)`,
	`(package pp "path") (using [X, X] x.x)`,
	`(package pp "path") (using [X  X] x.x)`,
}

func TestParse_short_valid(t *testing.T) {
	for i, src := range valid {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ctx := tel.Context{}
			file := tel.NewFile("test", "valid.tel", []byte(src))
			files := tel.NewFileSet("test", file)
			mod := Package{
				Name: "valid",
				Path: "test/valid",
			}
			loadPackage(ctx, files, &mod)
			if mod.HasError() {
				t.Fatal("parser failed unexpectedly")
			}
		})
	}
}

var invalid = []string{
	"foo",
	`(import i "path")`,
	`(using a i)`,
	`(enum X [a b])`,
	`(document X[])`,
	`(component X[])`,
	`(div)`,
	"(package)",
	"(package p)",
	`(package p ""`,
	`(package p "")`,
	`(package pp "X") (import)`,
	`(package pp "X") (import i)`,
	`(package pp "X") (using)`,
	`(package pp "X") (using i)`,
	`(package pp "X") (using [])`,
	`(package pp "X") (using [i)`,
	`(package pp "X") (using ])`,
	`(package pp "X") (document)`,
	`(package pp "X") (document [)`,
	`(package pp "X") (document [a])`,
	`(package pp "X") (document [a:])`,
	`(package pp "X") (document [] (div {))`,
	`(package pp "X") (document [] (div #{}))`,
	`(package pp "X") (document [] (div #a.{}))`,
	`(package pp "X") (document [] (div {a}))`,
	`(package pp "X") (document [] (div {a:}))`,
	"(package pp \"X\") (document [] (div {a: -- foo\n}))",
	"(package pp \"X\") (document [] (div {a: -- foo\\(foo)\n}))",
	`(package pp "X") (document [] (div a))`,
	`(package pp "X") (component)`,
	`(package pp "X") (component A)`,
	`(package pp "X") (component A [)`,
	`(package pp "X") (component A [a])`,
	`(package pp "X") (component A [a:])`,
	`(package pp "X") (component F [] (div {))`,
	`(package pp "X") (component F [] (div #{}))`,
	`(package pp "X") (component F [] (div #a.{}))`,
	`(package pp "X") (component F [] (div #a.b{}))`,
	`(package pp "X") (component F [] (div {a}))`,
	`(package pp "X") (component F [] (div {a:}))`,
	"(package pp \"X\") (component F [] (div {a: -- foo\n}))",
	"(package pp \"X\") (component F [] (div {a: -- foo\\(foo)\n}))",
	`(package pp "X") (component F [] (div a))`,
	`(package pp "X") (div)`,
	`(package pp "path") (import i "path") (document []) (using a i)`,
	`(package pp "path") (import i "path") (component X []) (using a i)`,
	`(package pp "path") (document []) (using a i)`,
	`(package pp "path") (component X []) (using a i)`,
	`(package pp "path") (document []) (document [])`,
	`(package pp "path") (enum E [])`,
	`(package pp "path") (component A[] (B ["name": "foo"]))`,
	`(package pp "path") (component A[] (B [name: "foo"))`,
	`(package pp "path") (component A[] (B [name "foo"]))`,
	`(package pp "a") (component Foo [] (div #a.b{}))`,
	`(package pp "a") (component Foo [] (div #a.b.c{}))`,
	`(package pp "path") (using x)`,
	`(package pp "path") (using [] x)`,
}

func TestParse_short_invalid(t *testing.T) {
	for i, src := range invalid {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			ctx := tel.Context{
				Flag: tel.PreserveComment | tel.ExitOnError,
			}
			file := tel.NewFile("test", "invalid.tel", []byte(src))
			files := tel.NewFileSet("test", file)
			mod := Package{
				Name: "valid",
				Path: "test/valid",
			}
			loadPackage(ctx, files, &mod)
			if !mod.HasError() {
				t.Error("parser succeeded unexpectedly")
			}
		})
	}
}

func TestPreserveComment(t *testing.T) {
	sourcecode := `(package p "views") ;; preserve comment`
	file := tel.NewFile("test", "test.tel", []byte(sourcecode))
	files := tel.NewFileSet("test", file)

	t.Run("preserve comment when token.PreserveComment flag is given", func(t *testing.T) {
		const expectedComments = 1
		const expectedComment = ";; preserve comment"

		ctx := tel.NewContextWithFlags(tel.PreserveComment)
		pkg := Package{
			Name: "valid",
			Path: "test/valid",
		}

		loadPackage(ctx, files, &pkg)
		mod := pkg.Modules[0]

		if got := len(mod.Comments); got != expectedComments {
			t.Fatalf("expected %d comment(s) got %d", expectedComments, got)
		}

		cmt := mod.Comments[0].Text
		if cmt != expectedComment {
			t.Fatalf("expected comment %q got %q", expectedComment, cmt)
		}
	})

	t.Run("do not preserve comment if token.PreserveComment flag is not given", func(t *testing.T) {
		const expectedComments = 0

		ctx := tel.NewContext()
		pkg := Package{
			Name: "valid",
			Path: "test/valid",
		}

		loadPackage(ctx, files, &pkg)

		mod := pkg.Modules[0]
		if got := len(mod.Comments); got != expectedComments {
			t.Fatalf("expected %d comment(s) got %d", expectedComments, got)
		}
	})
}
