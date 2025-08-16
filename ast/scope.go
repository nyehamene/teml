package ast

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/tel-lang/tel"
)

type Scope struct {
	parent   *Scope
	elements map[string]Entity
}

func newBuiltinScope(ctx tel.Context) *Scope {
	scope := newScope(nil)
	if ctx.Flag&tel.FlagNoBuiltinType == 0 {
		if err := bindBuiltinTypes(scope); err != nil {
			panic(err)
		}
	}
	if ctx.Flag&tel.FlagNoNativeElement == 0 {
		if err := bindNativeHTMLElements(scope); err != nil {
			panic(err)
		}
	}
	return scope
}

func newScope(parent *Scope) *Scope {
	scope := &Scope{
		parent:   parent,
		elements: map[string]Entity{},
	}
	assert(scope != nil, "scope is nil")
	assert(scope.elements != nil, "elements is nil")
	return scope
}

func bindBuiltinTypes(scope *Scope) error {
	assert(scope != nil, "scope is nil")
	errbool := bindBuiltinType(scope, "Bool", builtinBool)
	errstr := bindBuiltinType(scope, "String", builtinString)
	errnum := bindBuiltinType(scope, "Number", builtinNumber)

	err := errors.Join(errbool, errstr, errnum)
	if err != nil {
		return err
	}
	return nil
}

func bindBuiltinType(scope *Scope, name string, t Entity) error {
	assert(scope != nil, "scope is nil")
	err := scope.bind(name, t)
	return err
}

func bindNativeHTMLElements(scope *Scope) (err error) {
	assert(scope != nil, "scope is nil")
	for _, element := range NativeHTMLElements {
		if bindErr := scope.bind(element.Tag, &element); bindErr != nil {
			err = errors.Join(bindErr)
		}
	}
	return
}

func getScope(entity Entity) (scope *Scope, exists bool) {
	defer assertValue(scope != nil, exists, "existing scope is nil")
	assert(entity != nil, "entity is nil")

	switch kind := entity.entityKind(); kind {
	case EntityBuiltin,
		EntityInvalid,
		EntityVar:
		return nil, false

	case EntityModule:
		typeid := entity.(*TypeModule)
		return typeid.scope, true

	case EntityPackage:
		typeid := entity.(*TypePackage)
		return typeid.scope, true

	case EntityType:
		if enum, ok := entity.(*TypeEnum); ok {
			return enum.scope, true
		}
		if typeid, ok := entity.(*TypeTemplate); ok {
			return typeid.scope, true
		}
		panic(fmt.Sprintf("unexpected kind ast.EntityType: %#v", reflect.TypeOf(entity)))

	default:
		panic(fmt.Sprintf("unexpected ast.EntityKind: %#v", kind))
	}
}

func (s *Scope) bind(name string, entity Entity) error {
	assert(s != nil, "scope is nil")
	assert(name != "", "name is empty")
	assert(entity != nil, "entity is nil")
	if _, exists := s.elements[name]; exists {
		return ErrDuplicate
	}
	s.elements[name] = entity
	return nil
}

func (s *Scope) rebind(name string, entity Entity) error {
	assert(s != nil, "scope is nil")
	assert(name != "", "name is empty")
	assert(entity != nil, "entity is nil")
	if _, exists := s.elements[name]; !exists {
		return ErrUnbound
	}
	s.elements[name] = entity
	return nil
}

func (s *Scope) lookupCurrent(name string) (entity Entity, exists bool) {
	assert(name != "", "name is empty")
	defer assertValue(entity != nil, exists, "existing entity is nil")

	if s == nil {
		return nil, false
	}

	entity, exists = s.elements[name]
	return entity, exists
}

func (s *Scope) lookup(name string) (Entity, bool) {
	assert(name != "", "name is empty")

	if s == nil {
		return nil, false
	}

	entity, exists := s.lookupCurrent(name)
	if !exists && s.parent != nil {
		return s.parent.lookup(name)
	}

	return entity, exists
}
