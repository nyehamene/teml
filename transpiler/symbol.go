package ast

type Symbol interface {
	symbol()
}

func (BuiltinType) symbol()       {}

// builtin types
type BuiltinType int

const (
	TypeBool BuiltinType = iota
	TypeNumber
	TypeString
	TypeUnchecked
)
