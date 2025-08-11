package ast

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
func (Number) expr()         {}
func (Bool) expr()           {}
func (String) expr()         {}
func (StringTemplate) expr() {}
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
	Path  String
}

type Import struct {
	Ident Var
	Path  String
}

type Using struct {
	Idents []Var
	From   Var
}

type Document struct {
	Ident      Var
	Properties []Property
	Children   []Content
}

type Component struct {
	Ident      Var
	Properties []Property
	Children   []Content
}

type Property struct {
	Ident Var
	Type  PropertyType
}

type Enum struct {
	Constants []Expr
}

type Element struct {
	Ident      Expr
	Parameter  []ElementParameter
	Attributes []AttributeSet
	Children   []Content
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
	Cases  []CaseElement
}

type CaseElement struct {
	Cond   Expr
	Branch Content
}

type TaggedAttributeSet struct {
	Tag        Expr
	Attributes []Attribute
}

type UntaggedAttributeSet struct {
	Attributes []Attribute
}

type Attribute struct {
	Key   Var
	Value Expr
}

type Var struct {
	Name string
	Line int
	Col  int
}

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

type TextGroup []Text

type IfExpression struct {
	Cond Expr
	Then Expr
	Else Expr
}

type CondExpression struct {
	Target Expr
	Cases  []Case
}

type Case struct {
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

type String string
type StringTemplate string
type Number int
type Bool uint8

const (
	False Bool = iota
	True
)

type IntErrorNode int

const (
	badNode IntErrorNode = iota
)
