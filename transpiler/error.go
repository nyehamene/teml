package ast

import "fmt"

type SymbolError interface {
	Error() string
	symbolError()
}

func (ResolutionError) symbolError() {}

func (e ResolutionError) Error() string {
	return fmt.Sprintf("resolution error: %s", e.String())
}

type ResolutionError int

// resolution errors
const (
	ErrUndeclared ResolutionError = iota
	ErrDuplicateDeclaration
	ErrInvalidPackageName
)
