package ast

import (
	"fmt"
	"reflect"
)

type resolver struct {
	context *entityContext
}

func resolvePackage(ctx *entityContext) {
	r := &resolver{context: ctx}
	r.resolve()
}

func (r *resolver) resolve() {
	assert(r != nil, "resolver is nil")
	assert(r.context != nil, "context is nil")
	assert(r.context.pkg != nil, "package is nil")
	assert(r.context.scope != nil, "scope is nil")

	pkg := r.context.pkg
	pkgScope := r.context.scope

	// bind modules and package declarations
	for _, mod := range pkg.Modules {
		modScope := newScope(pkgScope)

		pkgEntity := &TypePackage{
			Var:   mod.Package.Var,
			scope: pkgScope,
			mod:   mod,
		}
		modEntity := &TypeModule{
			decl:  mod,
			Name:  mod.Name,
			pkg:   pkgEntity,
			scope: modScope,
		}

		if err := modScope.bind(mod.Package.Name, pkgEntity); err != nil {
			r.errorVar(err, mod.Package.Var)
		}
		if err := pkgScope.bind(mod.Name, modEntity); err != nil {
			r.error(err)
		}

		r.context.modules = append(r.context.modules, modEntity)

		// bind enums
		for _, decl := range mod.Enums {
			enumScope := newScope(modScope)
			enum := &TypeEnum{
				Var:   decl.Var,
				scope: enumScope,
				decl:  &decl,
			}
			if err := modScope.bind(decl.Name, enum); err != nil {
				r.errorVar(err, decl.Var)
			}
			r.bindEnum(enum)
		}

		// bind templates
		for _, decl := range mod.Templates {
			tmplScope := newScope(modScope)
			tmpl := &TypeTemplate{
				Var:        decl.Var,
				Kind:       decl.Kind,
				scope:      tmplScope,
				fieldCount: len(decl.Properties),
				decl:       &decl,
			}

			if err := modScope.bind(decl.Name, tmpl); err != nil {
				r.errorVar(err, decl.Var)
			}

			// bind parameters
			r.bindFields(tmpl, decl.Properties)

			// resolve statements
			r.resolveStmts(tmplScope, decl.Stmts)

			modEntity.templates = append(modEntity.templates, tmpl)
		}
	}

	assert(r.context.modules != nil, "context modules is nil")
}

func (r *resolver) bindEnum(enum *TypeEnum) {
	assert(r != nil, "resolver is nil")
	assert(enum != nil, "enum entity is nil")
	assert(enum.scope != nil, "enum scope is nil")
	assert(enum.decl != nil, "enum decl is nil")

	for _, c := range enum.decl.Constants {
		field := &TypeField{
			Var:  c.Var,
			Type: enum,
			decl: &c,
		}
		if err := enum.scope.bind(c.Name, field); err != nil {
			r.error(err)
		}
		enum.fields = append(enum.fields, field)
	}
}

func (r *resolver) resolveStmts(scope *Scope, stmts []Stmt) {
	assert(r != nil, "resolver is nil")
	assert(scope != nil, "scope is nil")

	for _, stmt := range stmts {
		r.resolveElement(scope, stmt)
	}
}

func (r *resolver) resolveElement(scope *Scope, element Stmt) {
	assert(r != nil, "resolver is nil")
	assert(scope != nil, "scope is nil")

	switch t := element.(type) {
	case ComponentElement,
		PropertyElement,
		HTMLElement,
		NumberElement,
		StringElement:
		// NOTE these nodes are not produce by the parser
		panic("unreachable")

	case Text,
		TextGroup,
		If,
		Cond:
		// no oop

	case Element:
		r.bindParams(scope, t.Parameter)
		r.bindAttributes(scope, t.Attributes)
		r.resolveStmts(scope, t.Children)

	default:
		panic(fmt.Sprintf("unexpected expression statement: %v", reflect.TypeOf(element)))
	}
}

func (r *resolver) bindFields(t *TypeTemplate, decls []FieldDecl) {
	assert(r != nil, "resolver is nil")
	assert(t != nil, "type is nil")
	assert(t.scope != nil, "scope is nil")

	// detect duplicate fields
	for _, decl := range decls {
		field := &TypeField{
			Var:  decl.Var,
			decl: &decl,
		}
		if err := t.scope.bind(decl.Name, field); err != nil {
			r.errorVar(err, decl.Var)
		}
		t.fields = append(t.fields, field)
	}
}

func (r *resolver) bindParams(_ *Scope, decls []ParameterDecl) {
	assert(r != nil, "resolver is nil")

	if len(decls) == 0 {
		return
	}

	dummyScope := newScope(nil)

	for _, decl := range decls {
		// detect duplicate parameters
		param := &TypeField{Var: decl.Var}
		if err := dummyScope.bind(decl.Name, param); err != nil {
			r.errorVar(err, decl.Var)
		}
	}
}

func (r *resolver) bindAttributes(_ *Scope, attrs []Attr) {
	assert(r != nil, "resolver is nil")

	if len(attrs) == 0 {
		return
	}

	dummyScope := newScope(nil)

	for _, attr := range attrs {
		for _, entry := range attr.Entries {
			// detect duplicate attributes
			entity := &TypeField{Var: entry.Key}
			if err := dummyScope.bind(entry.Key.Name, entity); err != nil {
				r.errorVar(err, entry.Key)
			}
		}
	}
}
