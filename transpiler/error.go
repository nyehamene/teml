package ast

import "fmt"

type SymbolError struct {
	err    error
	symbol Var
}

type Error int

const (
	ErrUndeclared Error = iota
	ErrDuplicateDeclaration
	ErrNamespaceNotfound
	ErrRecursiveDefinition
	ErrInvalidElementTag
	ErrUndeclaredType
	ErrTypeMismatch
	ErrPackageElementTag
	ErrEnumElementTag
	ErrParameterInNativeElement
	ErrParameterInPropertyElement
	ErrBoolElementTag
	ErrDocumentElementTag
)

func (s SymbolError) Error() string {
	name := s.symbol.Name
	line := s.symbol.Line
	col := s.symbol.Col
	err := fmt.Sprintf("%w at %s (%d, %d)", s.err, name, line, col)
	return err
}

func (e Error) Error() string {
	switch e {
	case ErrUndeclared:
		return "undeclared variable"
	case ErrUndeclaredType:
		return "undeclared type"
	case ErrDuplicateDeclaration:
		return "duplicate declaration"
	case ErrNamespaceNotfound:
		return "namespace not found"
	case ErrRecursiveDefinition:
		return "recursive type declaration"
	case ErrInvalidElementTag:
		return "invalid element tag"
	case ErrBoolElementTag:
		return "invalid tag type (bool)"
	case ErrEnumElementTag:
		return "invalid tag type (enum)"
	case ErrPackageElementTag:
		return "invalid tag type (package)"
	case ErrDocumentElementTag:
		return "invalid tag type (document)"
	case ErrParameterInNativeElement:
		return "invalid native element (parameter)"
	case ErrParameterInPropertyElement:
		return "invalid property element (parameter)"
	case ErrTypeMismatch:
		return "type mismatch"
	default:
		panic(fmt.Sprintf("unexpected ast.Error: %#v", e))
	}
}

func (f *File) addError(err error) {
	f.errs = append(f.errs, err)
}

func (f *File) addSymbolError(err error, ident Var) {
	f.addError(&SymbolError{err: err, symbol: ident})
}
