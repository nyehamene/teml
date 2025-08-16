package ast

import (
	"fmt"
)

//go:generate stringer -type=TypeKind
type TypeKind int8

const (
	KindInvalid TypeKind = iota
	KindBool
	KindNumber
	KindString
	KindHTML
	KindPackage
	KindModule
	KindTemplate
	KindEnum
	KindField
)

type Type interface {
	typeKind() TypeKind
	entityKind() EntityKind
}

func (*TypeBool) typeKind() TypeKind        { return KindBool }
func (*TypeNumber) typeKind() TypeKind      { return KindNumber }
func (*TypeString) typeKind() TypeKind      { return KindString }
func (*HTMLElementType) typeKind() TypeKind { return KindHTML }
func (*TypePackage) typeKind() TypeKind     { return KindPackage }
func (*TypeModule) typeKind() TypeKind      { return KindModule }
func (*TypeTemplate) typeKind() TypeKind    { return KindTemplate }
func (*TypeEnum) typeKind() TypeKind        { return KindEnum }

type TypePackage struct {
	Var
	scope *Scope
	mod   *Module
}

type TypeModule struct {
	Name      string
	decl      *Module
	scope     *Scope
	pkg       *TypePackage
	templates []*TypeTemplate
}

type TypeEnum struct {
	Var
	decl   *EnumDecl
	scope  *Scope
	fields []*TypeField
}

type TypeTemplate struct {
	Var
	Kind       TemplateKind
	fieldCount int
	decl       *TemplateDecl
	scope      *Scope
	fields     []*TypeField
}

type TypeField struct {
	Var
	decl *FieldDecl
	Type Type
}

type TypeBool struct{}
type TypeNumber struct{}
type TypeString struct{}

var (
	builtinBool   = &TypeBool{}
	builtinNumber = &TypeNumber{}
	builtinString = &TypeString{}
)

func typesEqual(x, y Type) bool {
	if x == nil || y == nil {
		return false
	}
	if x == y {
		return true
	}
	if x.typeKind() != y.typeKind() {
		return false
	}

	switch typeKind := x.typeKind(); typeKind {
	case KindBool,
		KindField,
		KindHTML,
		KindInvalid,
		KindModule,
		KindNumber,
		KindPackage,
		KindString:

		panic("unreachable")

	case KindEnum:
		xt := x.(*TypeEnum)
		yt := y.(*TypeEnum)
		// TBD(fix): compare the qualified type names
		return xt.Var == yt.Var

	case KindTemplate:
		xt := x.(*TypeTemplate)
		yt := y.(*TypeTemplate)
		// TBD(fix): compare the qualified type names
		return xt.Var == yt.Var

	default:
		panic(fmt.Sprintf("unexpected ast.TypeKind: %#v", typeKind))
	}
}
