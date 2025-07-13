package ast

import "fmt"

func (t BuiltinType) String() string {
	switch t {
	case TypeBool:
		return "Bool"
	case TypeNumber:
		return "Number"
	case TypeString:
		return "String"
	case TypeUnchecked:
		return "<unchecked>"
	}
	panic("Unreachable")
}

func (e ResolutionError) String() string {
	switch e {
	case ErrUndeclared:
		return "Undeclared var"
	case ErrDuplicateDeclaration:
		return "Duplicate declaration"
	}
	panic("Unreachable")
}
