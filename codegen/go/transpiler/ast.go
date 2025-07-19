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

type Struct struct {
	Name   string
	Fields []StructField
}

type StructField struct {
	Name string
	Type string
}

type Method struct {
	Name string
	Type string
	Body []Stmt
}

type Stmt interface {
	stmt()
}

func (ReturnNil) stmt()          {}
func (ReturnIfNotNil) stmt()     {}
func (WriteLiteralString) stmt() {}

type WriteLiteralString struct {
	Literal string
	Var     string
}

type ReturnNil struct{}
type ReturnIfNotNil string
