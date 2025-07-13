package ast

import "github.com/eml-lang/teml/token"

type Node interface {
	code()
}

func (Package) code()   {}
func (Import) code()    {}
func (Using) code()     {}
func (Component) code() {}
func (Document) code()  {}

type Declaration interface {
	decl()
}

func (Document) decl()  {}
func (Component) decl() {}

type Element interface {
	element()
}

func (TextElement) element()      {}
func (NumberElement) element()    {}
func (StringElement) element()    {}
func (ComponentElement) element() {}
func (InstanceElement) element()  {}
func (genericElement) element()   {}
func (NativeElement) element()    {}
func (IFElement) element()        {}
func (CondElement) element()      {}

type Expr interface {
	expr()
}

func (String) expr()       {}
func (Number) expr()       {}
func (Bool) expr()         {}
func (IFExpr) expr()       {}
func (CondExpr) expr()     {}
func (Var) expr()          {}
func (MemberAccess) expr() {}
func (Enum) expr()         {}

type PropertyType interface {
	typeExpr()
}

func (Var) typeExpr()          {}
func (MemberAccess) typeExpr() {}
func (Enum) typeExpr()         {}

// ast
type Package struct {
	Ident Var
	Path  string
}

type Import struct {
	Ident Var
	Path  string
}

type Using struct {
	From   Var
	Idents []Var
}

type Document Component

type Component struct {
	Ident      Var
	Properties []Property
	Stmts      []Stmt
}

type Property struct {
	Ident Var
	Type  PropertyType
}

// statements
type Stmt struct {
	Element Element
}

type TextElement struct {
	Text string
}

type NumberElement struct {
	Tag        Expr
	Attributes []Attr
}

type StringElement struct {
	Tag        Expr
	Attributes []Attr
}

type genericElement struct {
	Tag        Expr
	Parameter  []KeyVal
	Attributes []Attr
	Body       []Stmt
}

type NativeElement struct {
	Tag        Expr
	Attributes []Attr
	Body       []Stmt
}

type ComponentElement struct {
	Tag        Expr
	Attributes []Attr
	Body       []Stmt
}

type InstanceElement struct {
	Tag        Expr
	Parameter  []KeyVal
	Attributes []Attr
	Body       []Stmt
}

type IFElement struct {
	Cond Expr
	Then Element
	Else Element
}

type CondElement struct {
	Target Expr
	Cases  []CaseStmt
}

type CaseStmt struct {
	Cond   Expr
	Branch Stmt
}

// expressions
type String string
type Number string
type Bool string

type MemberAccess struct {
	Object Expr
	Member Var
}

type IFExpr struct {
	Cond Expr
	Then Expr
	Else Expr
}

type CondExpr struct {
	Target Expr
	Cases  []CaseExpr
}

type CaseExpr struct {
	Cond   Expr
	Branch Expr
}

type Var struct {
	Name string
	Pos  Pos
}

type Constant token.Token

type Enum struct {
	Constants []EnumConstant
}

type KeyVal struct {
	Key   Var
	Value Expr
}

type EnumConstant Constant

type Attr struct {
	Tag     Expr
	Entries []KeyVal
}
