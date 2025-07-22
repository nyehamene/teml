package ast

type TypeAlias struct {
	Name string
	Type string
}

type Package struct {
	Name string
}

type Import struct {
	Name string
	Path string
}

type RenderContextStruct struct {
	Struct
	Constructor string
	OptionType  string
}

type Struct struct {
	Name   string
	Fields []StructField
}

type StructField struct {
	Name string
	Type Type
}

type RenderMethod struct {
	//Name of the method
	Name string
	//Type the method is defined for
	Type string
	//Receiver variable name for the method
	Receiver string
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
func (FormatNumber) stmt()           {}
func (CallRenderMethod) stmt()       {}
func (CopyRenderContext) stmt()      {}
func (If) stmt()                     {}
func (Cond) stmt()                   {}
func (MapInstance) stmt()            {}
func (StructInstance) stmt()         {}

type StringLiteral struct {
	Value    string
	Variable string
}

type StringMemberAccessExpr struct {
	Value    string
	Variable string
}

type NumberMemberAccessExpr struct {
	Value    string
	Variable string
}

type InheritAttributes struct {
	//Variable for the error returned from the attributes
	Variable string
}

type FormatNumber struct {
	Value    string
	Variable string
}

type ReturnNil struct{}
type ReturnIfNotNil string

type CallRenderMethod struct {
	//Name of method to call
	Name string
	//Receiver variable name
	Receiver string
	//Variable stores the error return from the function call
	Variable string
	//Context variable name
	Context string
}

type BlankVar string

type MapInstance struct {
	Variable string
	Entries  []SetMapEntry
}

type StructInstance struct {
	Variable   string
	Type       string
	Parameters []SetStructField
}

type SetStructField struct {
	Struct string
	Name   string
	Value  Expr
}

type SetMapEntry struct {
	//Map of the map variable
	Map string
	//Key of the entry
	Key Var
	//Value of the entry
	Value Expr
}

type CopyRenderContext struct {
	Attrs    string
	Variable string
}

type If struct {
	//Cond expression
	Cond string
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
type Number string
type Bool string
type Var string

type Type interface {
	kind()
	Name() string
}

func (String) kind() {}
func (Number) kind() {}
func (Bool) kind()   {}
func (Var) kind()    {}
func (Enum) kind()   {}

func (s String) Name() string { return string(s) }
func (n Number) Name() string { return string(n) }
func (b Bool) Name() string   { return string(b) }
func (v Var) Name() string    { return string(v) }
func (e Enum) Name() string   { return e.TypeName }

type Enum struct {
	//TypeName
	TypeName string
	// TODO add the constant value type. accept only number and string (maybe accept bool)
	Constants []Expr
}
