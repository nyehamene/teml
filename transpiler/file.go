package ast

import (
	"fmt"
	"reflect"

	"github.com/eml-lang/teml/ast"
	"github.com/eml-lang/teml/internal/errors"
)

type Pos struct {
	Start int
	End   int
	Line  int
	Col   int
}

type Flag uint

type File struct {
	Package      Package
	Imports      []Import
	Usings       []Using
	Declarations []Declaration
	errs         []errors.Error
}

const (
	FlagNoBuiltinElement Flag = 1 << iota
	FlagNoBuiltinType
)

func (f *File) HasError() bool {
	if f == nil {
		return false
	}
	return len(f.errs) > 0
}

func (f *File) Errors() func(func(int, errors.Error) bool) {
	return func(yield func(int, errors.Error) bool) {
		for i, err := range f.errs {
			if !yield(i, err) {
				break
			}
		}
	}
}

func ParseFile(src *ast.File) *File {
	f := parseFile(src)
	return f
}

func parseFile(src *ast.File) *File {
	p := &parser{src: src}

	f := &File{
		errs: []errors.Error{},
	}

	f.Package = p.parsePackage()
	f.Imports = p.parseImports()
	f.Usings = p.parseUsings()

	decls := p.parseDeclarations()

	f.Declarations = decls
	return f
}

func ResolveFile(src *File, flags ...Flag) Env {
	var flag Flag
	for _, f := range flags {
		flag |= f
	}

	e := resolveFile(src, flag)
	return e
}

func resolveFile(f *File, flag Flag) Env {
	env := createEnv(nil)

	if flag&FlagNoBuiltinType == 0 {
		env = bindBuiltinTypes(env)
	}

	r := &resolver{src: f}
	r.resolvePackage(env)
	// TODO handle import and using declarations
	f.Declarations = r.resolveDeclarations(env)

	for _, tmpl := range f.Declarations {
		var name string
		var props []Property
		var stmts []Stmt

		switch t := tmpl.(type) {
		case Document:
			name = t.Ident.Name
			props = t.Properties
			stmts = t.Stmts

		case Component:
			name = t.Ident.Name
			props = t.Properties
			stmts = t.Stmts

		default:
			panic(fmt.Sprintf("unexpected declaration: %v", reflect.TypeOf(t)))
		}

		nestedEnv := env.Nest(name)
		r.resolveProperties(nestedEnv, props)

		if flag&FlagNoBuiltinElement == 0 {
			bindNativeElements(nestedEnv)
		}

		r.resolveStmts(nestedEnv, stmts)
	}

	return env
}

func TypecheckFile(src *File, env Env) Env {
	t := typechecker{src: src}
	e := t.typecheckFile(env)
	return e
}
