package ast

import (
	"errors"
	"strings"
)

const nsNative = "<native>"

type NameEnv struct {
	Name     string
	parent   *NameEnv
	children map[string]NameEnv
	names    map[string]Binding
}

type TypeEnv struct {
	names NameEnv
	types map[string]TypeSymbol
}

type Binding struct {
	ID     string
	IsType bool
}

func newNameRootEnv() NameEnv {
	return newNameEnv(nil, "")
}

func newNameEnv(parent *NameEnv, name string) NameEnv {
	env := NameEnv{
		Name:     name,
		parent:   parent,
		children: map[string]NameEnv{},
		names:    map[string]Binding{},
	}
	return env
}

func newTypeEnv(root NameEnv) TypeEnv {
	return TypeEnv{
		names: root,
		types: map[string]TypeSymbol{},
	}
}

func bindSame(env NameEnv, name string) {
	env.BindName(name, name)
}

func bindBuiltInTypeNames(env NameEnv) {
	bindSame(env, TypeBool.String())
	bindSame(env, TypeString.String())
	bindSame(env, TypeNumber.String())
}

func bindBuiltInTypes(env TypeEnv) error {
	errbool := env.BindType(TypeBool, TypeBool.String())
	errstr := env.BindType(TypeString, TypeString.String())
	errnum := env.BindType(TypeNumber, TypeNumber.String())
	return errors.Join(errbool, errstr, errnum)
}

func bindNativeElementNames(env NameEnv) {
	for _, element := range NativeElements {
		fqn := nsNative + "." + element.Tag
		env.BindName(element.Tag, fqn)
	}
}

func bindNativeElementTypes(env TypeEnv) {
	for _, element := range NativeElements {
		resolved, ok := env.names.LookupName(element.Tag)
		if !ok {
			panic("unreachable")
		}
		// NOTE since binding has IsType field, might want to check
		// if the resolved name is actually a type and report an error.
		_ = env.BindType(element, resolved.ID)
	}
}

func (e *NameEnv) Nest(name string) NameEnv {
	env := newNameEnv(e, name)
	e.children[name] = env
	// add to root
	root := e
	for root.parent != nil {
		root = root.parent
	}
	root.children[name] = env
	return env
}

func (e *NameEnv) BindName(name, fqn string) {
	binding := Binding{ID: fqn, IsType: false}
	e.names[name] = binding
}

func (e *NameEnv) BindTypeName(name, fqn string) {
	binding := Binding{ID: fqn, IsType: true}
	e.names[name] = binding
}

func (e *NameEnv) LookupName(name string) (Binding, bool) {
	resolved, ok := e.names[name]
	if !ok {
		if e.parent != nil {
			return e.parent.LookupName(name)
		}
		return Binding{}, false
	}
	return resolved, true
}

func (e *NameEnv) LookupNonNativeName(name string) (Binding, bool) {
	resolved, ok := e.LookupName(name)
	if !ok {
		if e.parent != nil {
			return e.parent.LookupNonNativeName(name)
		}
		return Binding{}, false
	}
	if strings.Contains(resolved.ID, nsNative) {
		return e.parent.LookupNonNativeName(name)
	}
	return resolved, true
}

func (e *NameEnv) LookupNameEnv(name string) (NameEnv, bool) {
	env, ok := e.children[name]
	if !ok {
		if e.parent != nil {
			return e.parent.LookupNameEnv(name)
		}
		return NameEnv{}, false
	}
	return env, true
}

func (e *TypeEnv) LookupType(name string) (TypeSymbol, bool) {
	sym, ok := e.types[name]
	if !ok {
		return nil, false
	}
	return sym, true
}

func (e *TypeEnv) BindType(sym TypeSymbol, name string) error {
	_, existing := e.LookupType(name)
	if existing {
		// TODO compare the existing type with the new type
		_ = existing
	}

	// bind if does not already exist
	e.types[name] = sym
	return nil
}
