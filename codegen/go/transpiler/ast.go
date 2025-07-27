package ast

import ast "github.com/eml-lang/teml/transpiler"

type TypeAlias struct {
	Name Var
	Type Var
}

type Package struct {
	Name Var
}

type Import struct {
	Name Var
	Path String
}

type RenderContextStruct struct {
	Struct
	Constructor Var
	OptionType  Var
}

type Struct struct {
	Name   Var
	Fields []StructField
}

type StructField struct {
	Name Var
	Type Type
}

type RenderMethod struct {
	//Name of the method
	Name Var
	//Type the method is defined for
	Type Var
	//Receiver variable name for the method
	Receiver Var
	//Body
	Body []Stmt
}

type Stmt interface {
	stmt()
}

func (ReturnNil) stmt()              {}
func (ReturnIfNotNil) stmt()         {}
func (StringLiteral) stmt()          {}
func (StringMemberAccessExpr) stmt() {}
func (NumberMemberAccessExpr) stmt() {}
func (InheritAttributes) stmt()      {}
func (InheritChildren) stmt()        {}
func (CallRenderMethod) stmt()       {}
func (CopyRenderContext) stmt()      {}
func (If) stmt()                     {}
func (Cond) stmt()                   {}
func (MapInstance) stmt()            {}
func (StructInstance) stmt()         {}
func (SliceInstance) stmt()          {}
func (ComponentInstance) stmt()      {}
func (BlankVar) stmt()               {}

type StringLiteral struct {
	Value String
	Error Var
}

type StringMemberAccessExpr struct {
	Value Var
	Error Var
}

type NumberMemberAccessExpr struct {
	Value    Var
	Error    Var
	Variable Var
}

type InheritAttributes struct {
	//Error for the error returned from the attributes
	Error Var
}

type InheritChildren struct {
	Context Var
}

type Children struct{}

type ReturnNil struct{}
type ReturnIfNotNil string

type CallRenderMethod struct {
	//Name of method to call
	Name Var
	//Receiver variable name
	Receiver Var
	//Error stores the error return from the function call
	Error Var
	//Context variable name
	Context Var
}

type BlankVar string

type MapInstance struct {
	Variable Var
	Entries  []SetMapEntry
}

type StructInstance struct {
	Variable   Var
	Type       Var
	Parameters []SetStructField
}

type SliceInstance struct {
	Variable Var
	Type     Var
	Values   []Var
}

type ComponentInstance struct {
	Variable Var
	Stmts    []Stmt
}

type SetStructField struct {
	Struct Var
	Name   Var
	Value  Expr
}

type SetMapEntry struct {
	//Map of the map variable
	Map Var
	//Key of the entry
	Key Var
	//Value of the entry
	Value Expr
}

type CopyRenderContext struct {
	Attrs    Var
	Variable Var
	Children Var
}

type If struct {
	//Cond expression
	Cond Expr
	Then []Stmt
	Else []Stmt
}

type Cond struct {
	Target Expr
	Cases  []Case
}

type Case struct {
	Match  Expr
	Branch []Stmt
}

type Expr interface {
	expr()
}

func (String) expr() {}
func (Number) expr() {}
func (Var) expr()    {}
func (Bool) expr()   {}

type String string
type Number int
type Bool ast.Bool
type Var string

type Type interface {
	kind()
	Name() Var
}

func (Var) kind()  {}
func (Enum) kind() {}

func (v Var) Name() Var  { return v }
func (e Enum) Name() Var { return e.TypeName }

type Enum struct {
	//TypeName
	TypeName Var
	// TODO add the constant value type. accept only number and string (maybe accept bool)
	Constants []Expr
}
