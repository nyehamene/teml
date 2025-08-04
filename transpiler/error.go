package ast

type SymbolError int

const (
	ErrUndeclared SymbolError = iota
	ErrDuplicateDeclaration
	ErrInvalidPackageName
	ErrNamespaceNotfound
	ErrRecursiveDefinition
	ErrInvalidElementTag
)

func (e SymbolError) Error() string {
	switch e {
	case ErrUndeclared:
		return "undeclared"
	case ErrDuplicateDeclaration:
		return "duplicate declaration"
	case ErrInvalidPackageName:
		return "invalid package name"
	case ErrNamespaceNotfound:
		return "namespace not found"
	case ErrRecursiveDefinition:
		return "recursive declaration"
	case ErrInvalidElementTag:
		return "invalid element tag"
	}
	panic("unreachable")
}
