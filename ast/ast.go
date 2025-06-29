package ast

import (
	"github.com/eml-lang/teml/internal/slice"
	"github.com/eml-lang/teml/token"
)

type Node interface {
	node()
}

type PropertyType interface {
	constant()
}

type AttributeSet interface {
	content()
	attrs()
}

type Content interface {
	content()
}

type Expr interface {
	expr()
}

type File struct {
	Package    Package
	Document   Document
	Imports    slice.Slice[Import]
	Usings     slice.Slice[Using]
	Components slice.Slice[Component]
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
	Idents slice.Slice[token.Token]
	From   token.Token
}

type Document struct {
	Ident      token.Token
	Properties slice.Slice[Property]
	Children   slice.Slice[Content]
}

type Component struct {
	Ident      token.Token
	Properties slice.Slice[Property]
	Children   slice.Slice[Content]
}

type Property struct {
	Ident token.Token
	Type  PropertyType
}

type IdentPropertyType token.Token

type EnumPropertyType struct {
	Constants slice.Slice[token.Token]
}

type Element struct {
	Ident    Expr
	Children slice.Slice[Content]
}

type IfElement struct {
	cond       Expr
	thenBranch Content
	elseBranch Content
}

type CondElement struct {
	cond    Expr
	Options slice.Slice[CondElementOption]
}

type CondElementOption struct {
	constant Expr
	branch   Content
}

type TaggedAttributeSet struct {
	tag        Expr
	Attributes slice.Slice[Attribute]
}

type UntaggedAttributeSet struct {
	Attributes slice.Slice[Attribute]
}

type Attribute struct {
	Key   token.Token
	value Expr
}

type Text token.Token

type TextGroup []Text

type IntErrorNode int

type PrimaryExpr token.Token

type IfExpression struct {
	cond       Expr
	thenBranch Expr
	elseBranch Expr
}

type CondExpression struct {
	cond    Expr
	Options slice.Slice[CondExpressionOption]
}

type CondExpressionOption struct {
	constant Expr
	value    Expr
}

type BinaryExpr struct {
	left  Expr
	right Expr
}

const (
	badNode IntErrorNode = iota
)

func (Package) node()   {}
func (Import) node()    {}
func (Using) node()     {}
func (Document) node()  {}
func (Component) node() {}

func (IntErrorNode) node() {}

func (Text) content()                 {}
func (TextGroup) content()            {}
func (Element) content()              {}
func (IfElement) content()            {}
func (CondElement) content()          {}
func (TaggedAttributeSet) content()   {}
func (UntaggedAttributeSet) content() {}

func (TaggedAttributeSet) attrs()   {}
func (UntaggedAttributeSet) attrs() {}

func (b BinaryExpr) expr()     {}
func (p PrimaryExpr) expr()    {}
func (e IfExpression) expr()   {}
func (e CondExpression) expr() {}

func (i IdentPropertyType) constant() {}
func (e EnumPropertyType) constant()  {}
