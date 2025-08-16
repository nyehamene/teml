package ast

import (
	"errors"
	"fmt"
)

type asterror struct {
	Message string
	Line    int
	Col     int
}

func (err asterror) Error() string {
	errmsg := fmt.Sprintf("%s (%d, %d)", err.Message, err.Line, err.Col)
	return errmsg
}

func errdesc(err asterror) error {
	return errdescString(err.Error())
}

func errdescString(err string) error {
	return fmt.Errorf(";desc: %s", err)
}

type SymbolError struct {
	err    error
	symbol Var
}

type SymbolErrors []error

//go:generate stringer -type=Error -linecomment
type Error int

const (
	ErrUndeclared          Error = iota // Undeclared
	ErrUnbound                          // Unbound
	ErrDuplicate                        // DuplicateVar
	ErrUndeclaredNamespace              // UndeclaredNamespace
	ErrRecursiveDefinition              // RecursiveDefinition
	ErrInvalidEnumConstant              // InvalidEnumConstant
	ErrUndeclaredType                   // UndeclaredType
	ErrNotObject                        // NotObject
	ErrUnexpectedChildren               // UnexpectedChildren
	ErrUnexpectedAttribute              // UnexpectedAttribute
	ErrUnexpectedParameter              // UnexpectedParameter

	ErrInvalidElementTag // InvalidElementTag

	ErrUnexpectedParameterInNativeElement   // UnexpectedParameterInNativeElement
	ErrUnexpectedParameterInPropertyElement // UnexpectedParameterInPropertyElement

	ErrMismatchType             // MismatchType
	ErrMismatchPackageName      // MismatchPackageName
	ErrMismatchEnumConstantType // MismatchEnumConstantType
	ErrMismatchName             // MismatchName
	ErrMismatchNamespace        // MismatchNamespace
	ErrDuplicateParameter       // DuplicateParameter
	ErrMissingParameter         // MissingParameter
	ErrUndeclaredParameter      // MissingUndeclaredParameter

	ErrInvalidFieldType // InvalidFieldType
	ErrNotType          // NotType
	ErrNotBool          // NotBool
	ErrNotEnum          // NotEnum

	ErrGeneric // Generic
)

func printError(e error) {
	fmt.Println(e.Error())
}

func (s SymbolError) Error() string {
	name := s.symbol.Name
	line := s.symbol.Line
	col := s.symbol.Col
	err := fmt.Errorf("%w at %s (%d, %d)", s.err, name, line, col)
	return err.Error()
}

func (s SymbolErrors) Error() string {
	if len(s) == 0 {
		return ""
	}

	var err error
	for _, e := range s {
		err = errors.Join(err, e)
	}
	return err.Error()
}

func (e Error) Error() string {
	return e.String()
}

func (p *Package) HasError() bool {
	return p != nil && p.hasError
}

func (m *Module) HasError() bool {
	return m != nil && m.hasError
}

func (p *Package) error(err error) {
	p.hasError = true
	printError(err)
}

func (m *Module) error(err error) {
	m.hasError = true
	printError(err)
}

func (r *resolver) error(err error) {
	r.context.pkg.error(err)
}

func (t *typechecker) error(err error) {
	t.context.pkg.error(err)
}

func (r *Package) errorVar(err error, ident Var) {
	r.error(&SymbolError{err: err, symbol: ident})
}

func (r *resolver) errorVar(err error, ident Var) {
	r.error(&SymbolError{err: err, symbol: ident})
}

func (t *typechecker) errorVar(err error, ident Var) {
	t.error(&SymbolError{err: err, symbol: ident})
}
