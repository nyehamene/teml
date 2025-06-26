package ast

import (
	"fmt"
	"iter"
	"log"
	"strconv"
	"strings"
	"testing"

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
	`(package p "path") (component C [a: (enum "A" "B")])`,
	`(package p "path") (component C [a: (enum "A", "B")])`,
	`(package p "path") (component C [a: (enum 0 1)])`,
	`(package p "path") (document C [] (if true (div)))`,
	`(package p "path") (component C [] (if f (div) (span)))`,
}

func TestParse_short_valid(t *testing.T) {
	for i, source := range valid {
		t.Run(fmt.Sprintf("%d %s", i, source), func(t *testing.T) {

			tokens := token.Scan([]byte(source), 0)

			_, hasError := ParseWithErrorHandler(*tokens, 0, func(err string) {
				t.Error(err)
			})

			if hasError {
				t.Error("Parser failed unexpectedly")
			}

		})
	}
}

var invalid = []string{
	"(package) ;desc: missing package identifier",
	"(package p) ;desc: missing package path",
	`(package p "" ;desc: missing closing parenthesis`,
	`(import) ;desc: missing import identifier`,
	`(import i) ;desc: missing import path`,
	"(using) ;desc: missing import alias",
	"(using i) ;desc: missing package to alias from",
	"(using []) ;desc: empty import alias list",
	"(using [i) ;desc: missing closing bracket",
	"(using ]) ;desc: missing import alias",
	"(document) ;desc: missing opening bracket",
	"(document [) ;desc: missing closing bracket",
	"(document [a]) ;desc: missing type separator",
	"(document [a:]) ;desc: missing property type",
	"(document [] (div {)) ;desc: missing closing brace",
	"(document [] (div #{})) ;desc: missing identifier",
	"(document [] (div #a/{})) ;desc: missing identifier",
	"(document [] (div {a})) ;desc: missing attribute value separator",
	"(document [] (div {a:})) ;desc: missing expression",
	"(document [] (div {a: -- foo\n})) ;desc: invalid expression",
	"(document [] (div {a: -- foo\\(foo)\n})) ;desc: invalid expression",
	"(document [] (div a)) ;desc: invalid content",
	"(component) ;desc: missing component identifier",
	"(component A) ;desc: missing opening bracket",
	"(component A [) ;desc: missing closing bracket",
	"(component A [a]) ;desc: missing type separator",
	"(component A [a:]) ;desc: missing property type",
	"(component F [] (div {)) ;desc: missing closing brace",
	"(component F [] (div #{})) ;desc: missing identifier",
	"(component F [] (div #a/{})) ;desc: missing identifier",
	"(component F [] (div {a})) ;desc: missing attribute value separator",
	"(component F [] (div {a:})) ;desc: missing expression",
	"(component F [] (div {a: -- foo\n})) ;desc: invalid expression",
	"(component F [] (div {a: -- foo\\(foo)\n})) ;desc: invalid expression",
	"(component F [] (div a)) ;desc: invalid content",
	"(div) ;desc: unexpected element declaration",
	"foo ;desc: missing opening parenthesis",
	`(import i "path") ;desc: missing package declaration`,
	`(package p "path") (using a i) ;desc: missing import declaration`,
	`(package p "path") (import i "path") (document []) (using a i) ;desc: unexpected using declaration`,
	`(package p "path") (document []) (document []) ;desc: duplicate document declaration`,
	`(package p "path") (component C [a: (enum "A" 10)]) ;desc: mismatch enum constant type`,
}

func TestParse_short_invalid(t *testing.T) {
	for i, source := range invalid {
		t.Run(fmt.Sprintf("%d %s", i, source), func(t *testing.T) {

			tokens := token.Scan([]byte(source), token.PreserveComment)
			goterrmsgs := map[string]string{}

			_, hasError := ParseWithErrorHandler(*tokens, token.ExitOnError, func(err string) {
				for e := range getEntriesFromString(err) {
					goterrmsgs[e.key] = e.value
				}
			})

			if !hasError {
				t.Error("Parser succeeded unexpectedly")
			}

			checkedAtLeastOneError := false
			for e := range getErrorMessagesFromComment(tokens) {
				got := goterrmsgs[e.key]
				if expected := e.value; expected != got {
					t.Errorf("expected %s but got %s", expected, got)
				}
				checkedAtLeastOneError = true
			}

			if !checkedAtLeastOneError {
				t.Fail()
			}
		})
	}
}

var valid_count = []string{
	`(document [a: A, b: B, c: C]) ;document_property: 3`,
	`(document [] (div {a: true, b: false, c: 100})) ;document_attribute: 3`,
	`(document [] (one) (two) (three)) ;document_content: 3`,
	`(document [] (div (one (one)))) ;document_content: 1`,
	`(document [] (div (one) (two) (three))) ;document_nested: 3`,
	`(document [] (div "one" "two")) ;document_nested: 2`,
	"(document [] (div --one\n --two\n --three\n --four\n)) ;document_nested: 4",
	"(document [] (div \"one\" --one\n (one))) ;document_nested: 3",

	`(component F [a: A, b: B]) ;property: 2`,
	`(component F [] (div {a: true, b: false})) ;attribute: 2`,
	`(component F [] (one) (two)) ;content: 2`,
	`(component F [] (div (one (one)))) ;content: 1`,
	`(component F [] (div (one) (two) (three))) ;nested: 3`,
	`(component F [] (div "one" "two")) ;nested: 2`,
	"(component F [] (div --one\n --two\n --three\n --four\n)) ;nested: 4",
	"(component F [] (div \"one\" --one\n (one))) ;nested: 3",
}

func TestValidCounting(t *testing.T) {
	for i, source := range valid_count {
		t.Run(fmt.Sprintf("%d %s", i, source), func(t *testing.T) {

			tokens := token.Scan([]byte(source), token.PreserveComment)
			f, hasError := Parse(*tokens, 0)

			if hasError {
				t.Error("parser failed unexpectedly")
			}

			for e := range getCountFromComment(tokens) {
				document := false

				switch e.key {
				case "document_property":
					document = true
					fallthrough

				case "property":
					props, ok := getFirstProperties(f, document)
					if !ok {
						t.Error("no properties found")
					}

					if l := len(props); e.value != l {
						t.Errorf("expected properties %d gut %d", e.value, l)
					}

				case "document_attribute":
					document = true
					fallthrough

				case "attribute":
					attrs, ok := getFirstAttributesAttributes(f, document)
					if !ok {
						t.Error("no element found")
					}

					if l := len(attrs); e.value != l {
						t.Errorf("expected attributes %d got %d", e.value, l)
					}

				case "document_content":
					document = true
					fallthrough

				case "content":
					children, ok := getChildren(f, document)
					if !ok {
						t.Error("no children found")
					}

					if l := len(children); e.value != l {
						t.Errorf("expected children %d got %d", e.value, l)
					}

				case "document_nested":
					document = true
					fallthrough

				case "nested":
					el, ok := getFirstElementChildren(f, document)
					if !ok {
						t.Error("no element found")
					}

					if l := len(el); e.value != l {
						t.Errorf("expected nested children %d got %d", e.value, l)
					}

				default:
					t.Errorf("unexpected count: %s", e.key)
				}
			}
		})
	}
}

func FuzzParse(f *testing.F) {
	for _, source := range valid {
		f.Add(source)
	}

	for _, source := range valid_count {
		f.Add(source)
	}

	f.Fuzz(func(t *testing.T, source string) {
		tokens := token.Scan([]byte(source), 0)
		_, hasError := parse(*tokens, 0, func(s string) {})

		if hasError {
			t.Error("parse failed unexpectedly")
		}
	})
}

func getFirstAttributesAttributes(f *File, document bool) ([]Attribute, bool) {
	e, ok := getFirstElement(f, document)
	if !ok {
		return nil, false
	}
	for _, c := range e.children {
		switch t := c.(type) {
		case TaggedAttributeSet:
			return t.attributes, true
		case UntaggedAttributeSet:
			return t.attributes, true
		}
	}
	return nil, false
}

func getChildren(f *File, document bool) ([]Content, bool) {
	if document {
		return f.document.children, true
	}

	c, ok := getFirstComponent(f)
	if !ok {
		return nil, false
	}

	return c.children, true
}

func getFirstProperties(f *File, document bool) ([]Property, bool) {
	if document {
		return f.document.properties, true
	}

	c, ok := getFirstComponent(f)
	if !ok {
		return nil, false
	}
	return c.properties, true
}

func getFirstComponent(f *File) (Component, bool) {
	if len(f.components) == 0 {
		return Component{}, false
	}

	return f.components[0], true
}

func getFirstElementChildren(f *File, document bool) ([]Content, bool) {
	e, ok := getFirstElement(f, document)
	if !ok {
		return nil, false
	}

	return e.children, true
}

func getFirstElement(f *File, document bool) (Element, bool) {
	getfirst := func(children []Content) (Element, bool) {
		for _, c := range children {
			switch t := c.(type) {
			case Element:
				return t, true
			}
		}
		return Element{}, false
	}

	if document {
		return getfirst(f.document.children)
	}

	c, ok := getFirstComponent(f)
	if !ok {
		return Element{}, false
	}

	return getfirst(c.children)
}

type entry[T any] struct {
	key   string
	value T
}

func getCountFromComment(f *token.Tokenized) iter.Seq[entry[int]] {
	return func(yield func(entry[int]) bool) {
		for tok := range getKind(f, token.Comment) {
			cmt, ok := f.Text(tok)
			if !ok {
				continue
			}

			for estr := range getEntriesFromString(cmt) {

				count, err := strconv.Atoi(estr.value)
				if err != nil {
					log.Println(err)
					continue
				}

				eint := entry[int]{key: estr.key, value: count}

				if !yield(eint) {
					break
				}
			}
		}
	}
}

func getErrorMessagesFromComment(f *token.Tokenized) iter.Seq[entry[string]] {
	return func(yield func(entry[string]) bool) {
		for tok := range getKind(f, token.Comment) {
			cmt, ok := f.Text(tok)
			if !ok {
				continue
			}

			for e := range getEntriesFromString(cmt) {
				if !yield(e) {
					break
				}
			}
		}
	}
}

func getKind(f *token.Tokenized, kind token.Kind) iter.Seq[token.Token] {
	return func(yield func(token.Token) bool) {
		for _, tok := range f.Tokens() {
			if tok.Kind != kind {
				continue
			}

			if !yield(tok) {
				break
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
