package ast

type Symbol interface {
	symbol()
}

func (BuiltinType) symbol()       {}
func (NativeElementType) symbol() {}
func (TypePackage) symbol()       {}
func (TypeDeclaration) symbol()   {}
func (TypeEnum) symbol()          {}
func (TypeVar) symbol()           {}

type TypeVar struct {
	Type Symbol
}

type TypePackage struct {
	Path string
}

type TypeEnum struct {
	ConstantType Symbol
}

type TypeDeclaration struct {
	Kind DeclarationKind
	Name string
}

type DeclarationKind int

const (
	DocumentDeclaration DeclarationKind = iota
	ComponentDeclaration
)

// builtin types
type BuiltinType int

const (
	TypeBool BuiltinType = iota
	TypeNumber
	TypeString
	TypeUnchecked
)
