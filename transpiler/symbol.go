package ast

import "github.com/eml-lang/teml/ast"

type Symbol interface {
	symbol()
}

func (BuiltinType) symbol()       {}
func (NativeElementType) symbol() {}
func (TypePackage) symbol()       {}
func (TypeDeclaration) symbol()   {}
func (TypeEnum) symbol()          {}

type TypePackage struct {
	Path String
}

type TypeEnum struct {
	ConstantType Symbol
}

type TypeDeclaration struct {
	Kind   ast.TemplateKind
	Name   string
	TypeId string
}

// builtin types
type BuiltinType int

const (
	TypeBool BuiltinType = iota
	TypeNumber
	TypeString
)

type TypeSymbol interface {
	Symbol
	typeSymbol()
}

func (BuiltinType) typeSymbol()       {}
func (NativeElementType) typeSymbol() {}
func (TypePackage) typeSymbol()       {}
func (TypeDeclaration) typeSymbol()   {}
func (TypeEnum) typeSymbol()          {}
