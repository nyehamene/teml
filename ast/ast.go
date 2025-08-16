package ast

type Node interface {
	node()
}

func (PackageDecl) node()  {}
func (ImportDecl) node()   {}
func (UsingDecl) node()    {}
func (TemplateDecl) node() {}
func (EnumDecl) node()     {}
func (badNode) node()      {}

type Stmt interface {
	stmt()
}

func (Text) stmt()             {}
func (TextGroup) stmt()        {}
func (If) stmt()               {}
func (Cond) stmt()             {}
func (Element) stmt()          {}
func (StringElement) stmt()    {}
func (NumberElement) stmt()    {}
func (PropertyElement) stmt()  {}
func (ComponentElement) stmt() {}
func (HTMLElement) stmt()      {}

type Expr interface {
	expr()
}

func (Var) expr()            {}
func (Number) expr()         {}
func (Bool) expr()           {}
func (String) expr()         {}
func (MemberAccess) expr()   {}
func (IfExpr) expr()         {}
func (CondExpr) expr()       {}
func (StringTemplate) expr() {}

type PackageDecl struct {
	Var
	ID string
}

type ImportDecl struct {
	Var
	Path String
}

type UsingDecl struct {
	From    Expr
	Aliases []Var
}

type TemplateDecl struct {
	Var
	Properties []FieldDecl
	Stmts      []Stmt
	Kind       TemplateKind
}

type FieldDecl struct {
	Var
	Type Expr
}

type EnumDecl struct {
	Var
	Constants []FieldDecl
}

type Element struct {
	Tag        Expr
	Parameter  []ParameterDecl
	Attributes []Attr
	Children   []Stmt
}

type ParameterDecl struct {
	Var
	Value Expr
}

type NumberElement struct {
	Tag        Var
	Attributes []Attr
}

type StringElement struct {
	Tag        Var
	Attributes []Attr
}

type HTMLElement struct {
	Tag        Var
	Attributes []Attr
	Children   []Stmt
}

type PropertyElement struct {
	Tag        Var
	Attributes []Attr
	Children   []Stmt
}

type ComponentElement struct {
	Tag        Var
	Parameters []ParameterDecl
	Attributes []Attr
}

type If struct {
	Cond Expr
	Then Stmt
	Else Stmt
}

type Cond struct {
	Target Expr
	Cases  []Case
}

type Case struct {
	Cond   Expr
	Branch Stmt
}

type Attr struct {
	// TBD: create a different type to represent attribute tags
	// Tags should have the format:
	// #<member_access>:<directive>
	// Ex:
	// #foo.bar:style
	// #baz:style
	// #main.(foo,baz):style
	//   which is equivalent to
	//   #main.foo:style
	//   #main.baz:style
	Directive Expr
	Entries   []KeyVal
}

type KeyVal struct {
	Key   Var
	Value Expr
}

type Var struct {
	Name      string
	Line      int
	Col       int
	namespace string
}

//go:generate stringer -type=TemplateKind
type TemplateKind uint8

const (
	TemplateDocument TemplateKind = iota
	TemplateComponent
)

//go:generater stringer -type=TextKind
type TextKind uint8

const (
	QuotedText TextKind = iota
	LineText
	QuotedTemplateText
	LineTemplateText
)

type Text struct {
	Kind  TextKind
	Value string
}

type TextGroup struct {
	Lines []Text
}

type IfExpr struct {
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

type MemberAccess struct {
	Object Expr
	Member Var
}

type Comment struct {
	Text string
	Line int
	Col  int
}

type badNode struct{}

type String string
type StringTemplate string
type Number int

//go:generate stringer -type=Bool
type Bool uint8

const (
	False Bool = iota
	True
)
