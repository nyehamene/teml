package ast

import (
	"iter"

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
	Type  PropertyType
}

type PropertyType interface {
	constant()
}

type IdentPropertyType token.Token

type EnumPropertyType struct {
	constants []token.Token
}

type Element struct {
	Ident    Expr
	children []Content
}

type IfElement struct {
	cond       Expr
	thenBranch Content
	elseBranch Content
}

type CondElement struct {
	cond    Expr
	options []CondElementOption
}

type CondElementOption struct {
	constant Expr
	branch   Content
}

type AttributeSet interface {
	content()
	attrs()
}

type TaggedAttributeSet struct {
	tag        Expr
	attributes []Attribute
}

type UntaggedAttributeSet struct {
	attributes []Attribute
}

type Attribute struct {
	Key   token.Token
	value Expr
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
	badNode IntErrorNode = iota
	unexpectedTokenError
)

func (Package) node()   {}
func (Import) node()    {}
func (Using) node()     {}
func (Document) node()  {}
func (Component) node() {}

func (IntErrorNode) node() {}

func (Text) content()                 {}
func (Element) content()              {}
func (IfElement) content()            {}
func (CondElement) content()          {}
func (TaggedAttributeSet) content()   {}
func (UntaggedAttributeSet) content() {}

func (TaggedAttributeSet) attrs()   {}
func (UntaggedAttributeSet) attrs() {}

func (b BinaryExpr) expr()  {}
func (p PrimaryExpr) expr() {}

func (i IdentPropertyType) constant() {}
func (e EnumPropertyType) constant()  {}

func (f File) Package() Package {
	return f.pkg
}

func (f File) Document() Document {
	return f.document
}

func (f File) Components() iter.Seq2[int, Component] {
	return seq(f.components, func(c Component) Component {
		return c
	})
}

func (c Component) Properties() iter.Seq2[int, Property] {
	return seq(c.properties, func(p Property) Property {
		return p
	})
}

func seq[A any, E any](es []E, m func(E) A) iter.Seq2[int, A] {
	return func(yield func(int, A) bool) {
		for i, e := range es {
			a := m(e)
			if !yield(i, a) {
				break
			}
		}
	}
}
