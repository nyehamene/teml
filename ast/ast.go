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

func (Var) expr()            {}
func (Constant) expr()       {}
func (MemberAccess) expr()   {}
func (IfExpression) expr()   {}
func (CondExpression) expr() {}
func (Enum) expr()           {}

type PropertyType interface {
	typeExpr()
}

func (Enum) typeExpr()         {}
func (Var) typeExpr()          {}
func (MemberAccess) typeExpr() {}

type Package struct {
	Ident Var
	Path  token.Token
}

type Import struct {
	Ident Var
	Path  Constant
}

type Using struct {
	Idents slice.Slice[Var]
	From   Var
}

type Document struct {
	Ident      Var
	Properties slice.Slice[Property]
	Children   slice.Slice[Content]
}

type Component struct {
	Ident      Var
	Properties slice.Slice[Property]
	Children   slice.Slice[Content]
}

type Property struct {
	Ident Var
	Type  PropertyType
}

type Enum struct {
	Constants slice.Slice[Constant]
}

type Element struct {
	Ident      Expr
	Parameter  slice.Slice[ElementParameter]
	Attributes slice.Slice[AttributeSet]
	Children   slice.Slice[Content]
}

type ElementParameter struct {
	Ident Var
	Value Expr
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
	Key   Var
	Value Expr
}

type Var token.Token

type Text token.Token

type TextGroup []Text

type IntErrorNode int

type Constant token.Token

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
