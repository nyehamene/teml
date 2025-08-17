package ast

import "fmt"

func (t TypeDeclaration) String() string {
	return fmt.Sprintf("<%s %s>", t.Kind, t.Name)
}

func (t BuiltinType) String() string {
	switch t {
	case TypeBool:
		return "Bool"
	case TypeNumber:
		return "Number"
	case TypeString:
		return "String"
	}
	panic("Unreachable")
}
