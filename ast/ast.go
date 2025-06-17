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
	elements   []Element
}

type Component struct {
	Ident      token.Token
	properties []Property
	elements   []Element
}

type Property struct {
	Ident token.Token
	Type  token.Token
}

type Element struct {
	Ident      Expr
	attributes []Attribute
}

type Attribute struct {
	tag   Expr
	Ident token.Token
	Value token.Token
}

type PrimaryExpr token.Token

type propertyholder interface {
	addProperty(Property)
}

type elementholder interface {
	addElement(Element)
}

type IntErrorNode int

type Expr interface {
	expr()
}

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

func (b BinaryExpr) expr()  {}
func (p PrimaryExpr) expr() {}

func (f *File) Package() Package {
	return f.pkg
}

func (c *Component) addProperty(p Property) {
	c.properties = append(c.properties, p)
}

func (c *Document) addProperty(p Property) {
	c.properties = append(c.properties, p)
}

func (c *Component) addElement(e Element) {
	c.elements = append(c.elements, e)
}

func (c *Document) addElement(e Element) {
	c.elements = append(c.elements, e)
}
