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
	Type string
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

func (ReturnNil) stmt()                 {}
func (ReturnIfNotNil) stmt()            {}
func (WriteLiteralString) stmt()        {}
func (WriteStringMemberAccess) stmt()   {}
func (WriteNumberMemberAccess) stmt()   {}
func (WriteInheritedAttributes) stmt()  {}
func (FormatNumber) stmt()              {}
func (CallRenderFunction) stmt()        {}
func (MapVar) stmt()                    {}
func (MapEntry) stmt()                  {}
func (CopyContextWithAttributes) stmt() {}
func (If) stmt()                        {}

type WriteLiteralString struct {
	Value    string
	Variable string
}

type WriteStringMemberAccess struct {
	Value    string
	Variable string
}

type WriteNumberMemberAccess struct {
	Value    string
	Variable string
}

type WriteInheritedAttributes struct {
	//Variable for the error returned from the attributes
	Variable string
}

type FormatNumber struct {
	Value    string
	Variable string
}

type ReturnNil struct{}
type ReturnIfNotNil string

type CallRenderFunction struct {
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

type MapVar struct {
	Variable string
}

type MapEntry struct {
	//Name of the map variable
	Name string
	//Key of the entry
	Key string
	//Value of the entry
	Value Expr
}

type CopyContextWithAttributes struct {
	Attrs    string
	Variable string
}

type If struct {
	//Cond expression
	Cond string
	Then []Stmt
	Else []Stmt
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
