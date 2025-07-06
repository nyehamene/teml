package ast

import (
	"github.com/eml-lang/teml/internal/slice"
	"github.com/eml-lang/teml/token"
)

type Node interface {
	node()
}

func (Package) node()      {}
func (Import) node()       {}
func (Using) node()        {}
func (Document) node()     {}
func (Component) node()    {}
func (IntErrorNode) node() {}

type PropertyType interface {
	proptype()
}

func (SimpleType) proptype()    {}
func (Enum) proptype()          {}
func (QualifiedType) proptype() {}

type AttributeSet interface {
	attrs()
}

func (TaggedAttributeSet) attrs()   {}
func (UntaggedAttributeSet) attrs() {}

type Content interface {
	content()
}

func (Text) content()        {}
func (TextGroup) content()   {}
func (Element) content()     {}
func (IfElement) content()   {}
func (CondElement) content() {}

type Expr interface {
	expr()
}

func (PrimaryExpr) expr()    {}
func (MemberAccess) expr()   {}
func (IfExpression) expr()   {}
func (CondExpression) expr() {}

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

type SimpleType token.Token

type Enum struct {
	Constants slice.Slice[token.Token]
}

type QualifiedType struct {
	Object PropertyType
	Name   PropertyType
}

type Element struct {
	Ident      Expr
	Attributes slice.Slice[AttributeSet]
	Children   slice.Slice[Content]
}

type IfElement struct {
	Cond Expr
	Then Content
	Else Content
}

type CondElement struct {
	Target Expr
	Cases  slice.Slice[CaseElement]
}

type CaseElement struct {
	Cond   Expr
	Branch Content
}

type TaggedAttributeSet struct {
	Tag        Expr
	Attributes slice.Slice[Attribute]
}

type UntaggedAttributeSet struct {
	Attributes slice.Slice[Attribute]
}

type Attribute struct {
	Key   token.Token
	Value Expr
}

type Text token.Token

type TextGroup []Text

type IntErrorNode int

type PrimaryExpr token.Token

type IfExpression struct {
	Cond Expr
	Then Expr
	Else Expr
}

type CondExpression struct {
	Target Expr
	Cases  slice.Slice[Case]
}

type Case struct {
	Cond   Expr
	Branch Expr
}

type MemberAccess struct {
	Object Expr
	Member Var
}

const (
	badNode IntErrorNode = iota
)

func (d Document) IsNamed() bool {
	return d.Ident.Kind != token.Invalid
}
