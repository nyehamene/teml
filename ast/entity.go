package ast

import (
	"github.com/tel-lang/tel"
)

//go:generate stringer -type=EntityKind
type EntityKind int8

const (
	EntityInvalid EntityKind = iota
	EntityBuiltin
	EntityModule
	EntityPackage
	EntityType
	EntityVar
)

type entityContext struct {
	tel.Context
	pkg       *Package
	modules   []*TypeModule
	scope     *Scope
}

type Entity interface {
	entityKind() EntityKind
}

func (*TypeBool) entityKind() EntityKind        { return EntityBuiltin }
func (*TypeNumber) entityKind() EntityKind      { return EntityBuiltin }
func (*TypeString) entityKind() EntityKind      { return EntityBuiltin }
func (*HTMLElementType) entityKind() EntityKind { return EntityBuiltin }
func (*TypePackage) entityKind() EntityKind     { return EntityPackage }
func (*TypeModule) entityKind() EntityKind      { return EntityModule }
func (*TypeTemplate) entityKind() EntityKind    { return EntityType }
func (*TypeEnum) entityKind() EntityKind        { return EntityType }
func (*TypeField) entityKind() EntityKind       { return EntityVar }
