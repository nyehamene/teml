package ast

import "fmt"

func (t TypeDeclaration) String() string {
	return fmt.Sprintf("<%s %s>", t.Kind, t.Name)
}

func (k DeclarationKind) String() string {
	switch k {
	case DocumentDeclaration:
		return "document"
	case ComponentDeclaration:
		return "component"
	}
	panic("Unreachable")
}

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

func (e TypeError) String() string {
	switch e {
	case ErrUndefined:
		return "Undefined"
	case ErrRecursiveDefinition:
		return "Undefined self"
	}
	panic("Unreachable")
}
