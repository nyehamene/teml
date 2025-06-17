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
