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

const (
	errorLabel = "SYNTAX ERROR"
)

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

	ident := p.expect(token.Ident, "invalid package identifier")
	path := p.expect(token.String, "invalid package path string")

	p.expect(token.ParenClose, "invalid delimiter")

	pkg := Package{Ident: ident, Path: path}
	p.file.Pkg = pkg
}

func (p *parser) expect(k token.Kind, msg ...string) pToken {
	ch := p.peek()
	if ch.Kind != k {
		m := strings.Join(msg, " ")
		p.printerr(fmt.Sprintf("error at %s. expected %s\n%s", ch.Kind, k, m))
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
