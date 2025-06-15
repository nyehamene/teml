package ast

import (
	"fmt"
	"log"
	"strings"

	"github.com/eml-lang/teml/token"
)

type pToken = token.Token

type parser struct {
	src      token.Tokenized
	cur      int
	printerr func(string)
	hasError bool
	file     *File
}

var (
	tokenEOF pToken = pToken{Kind: -1, Pos: -1}
)

func Parse(toks token.Tokenized) (*File, bool) {
	printerr := func(s string) {
		log.Println(s)
	}
	return ParseWithErrorHandler(toks, printerr)
}

func ParseWithErrorHandler(toks token.Tokenized, printerr func(string)) (*File, bool) {
	f := NewFile()
	ok := parse(toks, f, printerr)
	return f, ok

}

func parse(toks token.Tokenized, f *File, printerr func(string)) bool {
	p := parser{
		src:  toks,
		file: f,
	}

	p.printerr = func(s string) {
		p.hasError = true
		printerr(s)
	}

	p.parsePackage()

	return p.hasError
}

func (p *parser) parsePackage() {
	p.expect(token.ParanOpen, "invalid delimiter")
	p.expect(token.Package, "invalid declaration")

	ident := p.expect(token.Ident, errfmt.title("Package Declaration Error"), errfmt.desc("missing identifier"))
	path := p.expect(token.String, errfmt.title("Package Declaration Error"), errfmt.desc("missing path string"))

	p.expect(token.ParenClose, "invalid delimiter")

	pkg := Package{Ident: ident, Path: path}
	p.file.Pkg = pkg
}

func (p *parser) expect(k token.Kind, msgs ...errmessage) pToken {
	ch := p.peek()
	if ch.Kind != k {
		exp := errfmt.expect(k)
		got := errfmt.got(ch.Kind)

		errs := make([]string, 0, len(msgs)+2)
		errs = append(errs, msgs...)
		errs = append(errs, exp, got)
		msg := strings.Join(errs, "\n")
		p.printerr(msg)
	}
	p.advance()
	return ch
}

func (p *parser) advance() {
	if p.eof() {
		return
	}
	next := p.cur + 1
	p.cur = next
}

func (p *parser) peek() pToken {
	if p.eof() {
		return tokenEOF
	}

	next := p.cur
	if node, ok := p.src.Token(next); !ok {
		return tokenEOF
	} else {
		return node
	}
}

func (p *parser) eof() bool {
	e := p.cur >= p.src.Size()
	return e
}

type stringer interface {
	String() string
}
type errformatter struct{}
type errmessage = string

var errfmt errformatter

func (errformatter) title(msg string) errmessage {
	return fmt.Sprintf(";error: %s", msg)
}

func (errformatter) desc(msg string) errmessage {
	return fmt.Sprintf(";desc: %s", msg)
}

func (errformatter) expect(msg stringer) errmessage {
	return fmt.Sprintf(";expected: %s", msg)
}

func (errformatter) got(msg stringer) errmessage {
	return fmt.Sprintf(";got: %s", msg)
}
