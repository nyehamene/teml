package ast

import "github.com/eml-lang/teml/ast"

type Node interface {
	code()
}

func (Package) code()  {}
func (Import) code()   {}
func (Using) code()    {}
func (Template) code() {}

type Element interface {
	element()
}

func (TextElement) element()      {}
func (TextGroupElement) element() {}
func (NumberElement) element()    {}
func (StringElement) element()    {}
func (PropertyElement) element()  {}
func (ComponentElement) element() {}
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
	Path  String
}

type Import struct {
	Ident Var
	Path  String
}

type Using struct {
	From   Var
	Idents []Var
}

type Template struct {
	Kind       ast.TemplateKind
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
	Text String
}

type TextGroupElement struct {
	Lines []String
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

type PropertyElement struct {
	Tag        Expr
	Attributes []Attr
	Body       []Stmt
}

type ComponentElement struct {
	Tag        Expr
	Parameters []KeyVal
	Attributes []Attr
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
type Number int
type Bool ast.Bool

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

type Var ast.Var

type Enum struct {
	Constants []EnumConstant
}

type KeyVal struct {
	Key   Var
	Value Expr
}

type EnumConstant Expr

type Attr struct {
	Tag     Expr
	Entries []KeyVal
}

type Comment = ast.Comment

func (s String) Value() string {
	// strip double quoted
	return string(s[1 : len(s)-1])
}

func Join(xs []String, sep string) String {
	var result String
	for _, s := range xs {
		if result != "" {
			result += String(sep)
		}
		result += s
	}
	return result
}

func (v Var) Join(other Var, sep string) string {
	name := v.Name + sep + other.Name
	return name
}
