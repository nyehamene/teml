package ast

import (
	"github.com/eml-lang/teml/token"
)

type File struct {
	pkg        Package
	imports    []Import
	usings     []Using
	document   Document
	components []Component
}

type Node interface {
	node()
}

type Package struct {
	Ident token.Token
	Path  token.Token
}

type Import struct {
	Ident token.Token
	Path  token.Token
}

type Using struct {
	idents []token.Token
	From   token.Token
}

type Document struct {
	Ident      token.Token
	properties []Property
	children   []Content
}

type Component struct {
	Ident      token.Token
	properties []Property
	children   []Content
}

type Property struct {
	Ident token.Token
	Type  token.Token
}

type Element struct {
	Ident      Expr
	attributes []Attribute
	children   []Content
}

type Attribute struct {
	tag   Expr
	Ident token.Token
	Value Expr
}

type Content interface {
	content()
}

type Text token.Token

type IntErrorNode int

type Expr interface {
	expr()
}

type PrimaryExpr token.Token

type BinaryExpr struct {
	left  Expr
	right Expr
}

const (
	EOF IntErrorNode = iota
	unexpectedTokenError
)

func (Package) node()   {}
func (Import) node()    {}
func (Using) node()     {}
func (Document) node()  {}
func (Component) node() {}

func (IntErrorNode) node() {}

func (Text) content()      {}
func (Element) content()   {}
func (Attribute) content() {}

func (b BinaryExpr) expr()  {}
func (p PrimaryExpr) expr() {}

func (f *File) Package() Package {
	return f.pkg
}

func (f *File) adjustSize(tok token.Tokenized) {
	usings := 0
	imports := 0
	components := 0

	for _, tok := range tok.Tokens() {
		switch tok.Kind {
		case token.Using:
			usings += 1
		case token.Import:
			imports += 1
		case token.Component:
			components += 1
		}
	}

	if imports > 0 {
		f.imports = make([]Import, 0, imports)
	}

	if usings > 0 {
		f.usings = make([]Using, 0, usings)
	}

	if components > 0 {
		f.components = make([]Component, 0, components)
	}
}

func (f *Using) adjustSize(start int, tok token.Tokenized) {
	idents := 0

loop:
	for _, ch := range tok.TokensFrom(start) {

		switch ch.Kind {
		case token.ParenClose:
			break loop
		case token.Ident:
			idents += 1
		}
	}

	if idents == 0 {
		return
	}

	f.idents = make([]token.Token, 0, idents)
}

func createSizedPropertySlice(start int, tok token.Tokenized) []Property {
	props := 0
	p := parser{src: tok, cur: start}

	for !p.eof() {
		if ch := p.peek(); ch.Kind == token.BracketClose {
			break
		}
		p.parseProperty()
		props += 1
	}

	if props == 0 {
		return nil
	}

	return make([]Property, 0, props)
}

func createSizedContentSlize(start int, tok token.Tokenized) []Content {
	contents := 0
	openParans := 0

loop:
	for _, ch := range tok.TokensFrom(start) {
		switch ch.Kind {
		case token.ParenOpen:
			openParans += 1
			fallthrough
		case token.String, token.StringLine, token.StringTempl, token.StringLineTempl:
			contents += 1
		case token.ParenClose:
			if openParans == 0 {
				break loop
			}
			openParans -= 1
		}
	}

	if contents == 0 {
		return nil
	}

	return make([]Content, 0, contents)
}
