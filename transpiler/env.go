package ast

import perrors "github.com/eml-lang/teml/internal/errors"

type Env interface {
	Lookup(string) (Symbol, SymbolError)
	LookupEnv(string) (Env, bool)
	Bind(Symbol, string) SymbolError
	BindEnv(Env, string) SymbolError
	Nest(string) Env
}

type defaultEnv struct {
	parent Env
	env    map[string]Symbol
	nested map[string]Env
	errs   []perrors.Error
}

func (r defaultEnv) Lookup(name string) (Symbol, SymbolError) {
	resolved, ok := r.env[name]
	switch ok {
	case true:
		return resolved, nil

	case false:
		if r.parent == nil {
			break
		}
		return r.parent.Lookup(name)
	}

	return nil, ErrUndeclared
}

func (r defaultEnv) LookupEnv(name string) (Env, bool) {
	env, ok := r.nested[name]
	return env, ok
}

func (r defaultEnv) Bind(sym Symbol, name string) SymbolError {
	if existing, ok := r.env[name]; ok {
		switch existing {
		case TypeUnchecked:
			switch sym {
			case TypeUnchecked:
				return ErrDuplicateDeclaration
			}
		}
	}
	r.env[name] = sym
	return nil
}

func (r defaultEnv) BindEnv(env Env, name string) SymbolError {
	if _, ok := r.nested[name]; ok {
		return ErrDuplicateDeclaration
	}
	r.nested[name] = env
	return nil
}

func (r defaultEnv) Nest(name string) Env {
	env := createEnv(r)
	r.BindEnv(env, name)
	return env
}

func createEnv(parent Env) Env {
	env := defaultEnv{
		parent: parent,
		env:    map[string]Symbol{},
		nested: map[string]Env{},
	}
	return env
}

func bindBuiltinTypes(env Env) Env {
	env.Bind(TypeString, "String")
	env.Bind(TypeNumber, "Number")
	env.Bind(TypeBool, "Bool")
	return env

}

func bindNativeElements(env Env) Env {
	for _, e := range NativeElements {
		env.Bind(e, e.Tag)
	}
	return env
}
