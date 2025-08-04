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
	names    map[string]string
}

type TypeEnv struct {
	names NameEnv
	types map[string]TypeSymbol
}

func newRootNameEnv() NameEnv {
	return newNameEnv(nil, "")
}

func newNameEnv(parent *NameEnv, name string) NameEnv {
	env := NameEnv{
		Name:     name,
		parent:   parent,
		children: map[string]NameEnv{},
		names:    map[string]string{},
	}
	return env
}

func newTypeEnv(root NameEnv) TypeEnv {
	return TypeEnv{
		names: root,
		types: map[string]TypeSymbol{},
	}
}

func bindSame(env NameEnv, name string) error {
	if err := env.BindName(name, name); err != nil {
		return err
	}
	return nil
}

func bindBuiltInTypeNames(env NameEnv) error {
	errbool := bindSame(env, TypeBool.String())
	errstr := bindSame(env, TypeString.String())
	errnum := bindSame(env, TypeNumber.String())
	return errors.Join(errbool, errstr, errnum)
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
		_ = env.BindName(element.Tag, fqn)
	}
}

func bindNativeElementTypes(env TypeEnv) {
	for _, element := range NativeElements {
		fqn, ok := env.names.LookupName(element.Tag)
		if !ok {
			panic("unreachable")
		}
		_ = env.BindType(element, fqn)
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

func (e *NameEnv) BindName(name, fqn string) error {
	e.names[name] = fqn
	return nil
}

func (e *NameEnv) LookupName(name string) (string, bool) {
	resolved, ok := e.names[name]
	if !ok {
		if e.parent != nil {
			return e.parent.LookupName(name)
		}
		return "", false
	}
	return resolved, true
}

func (e *NameEnv) LookupNonNativeName(name string) (string, bool) {
	resolved, ok := e.LookupName(name)
	if !ok {
		if e.parent != nil {
			return e.parent.LookupNonNativeName(name)
		}
		return "", false
	}
	if strings.Contains(resolved, nsNative) {
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
