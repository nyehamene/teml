package ast

import (
	"errors"
	"fmt"
	"reflect"

	perrors "github.com/eml-lang/teml/internal/errors"
)

type resolver struct {
	src *File
}

func (r *resolver) resolvePackage(env Env) {
	p := r.src.Package
	r.bindVar(env, p.Ident)
}

func (r *resolver) resolveDeclarations(env Env) []Declaration {
	tmpls := r.src.Declarations
	for _, rec := range tmpls {
		r.resolveDeclaration(env, rec)
	}
	return tmpls
}

func (r *resolver) resolveDeclaration(env Env, rec Declaration) Declaration {
	switch t := rec.(type) {
	case Document:
		r.bindVar(env, t.Ident)
	case Component:
		r.bindVar(env, t.Ident)
	default:
		panic(fmt.Sprintf("unexpected declaration: %v", reflect.TypeOf(rec)))
	}
	return rec
}

func (r *resolver) resolveProperties(env Env, props []Property) {
	for _, prop := range props {
		r.resolvePropertyType(env, prop.Type)
		r.bindVar(env, prop.Ident)
	}
}

func (r *resolver) resolveStmts(env Env, stmts []Stmt) {
	for _, stmt := range stmts {
		r.resolveElement(env, stmt.Element)
	}
}

func (r *resolver) resolveElement(env Env, expr Element) {
	switch t := expr.(type) {
	case TextElement:
	case TextGroupElement:
	case ComponentElement, NativeElement, NumberElement, StringElement:
		// NOTE these are not produce by the parser

	case genericElement:
		r.resolveExpr(env, t.Tag)
		r.resolveParameter(env, t.Parameter)
		r.resolveAttributes(env, t.Attributes)
		r.resolveStmts(env, t.Body)

	case IFElement:
		r.resolveExpr(env, t.Cond)
		r.resolveElement(env, t.Then)
		if t.Else != nil {
			r.resolveElement(env, t.Else)
		}

	case CondElement:
		r.resolveExpr(env, t.Target)
		for _, c := range t.Cases {
			r.resolveExpr(env, c.Cond)
			r.resolveElement(env, c.Branch.Element)
		}

	default:
		panic(fmt.Sprintf("unexpected expression statement: %v", reflect.TypeOf(expr)))
	}
}

func (r *resolver) resolveParameter(env Env, params []KeyVal) {
	keyEnv := defaultEnv{env: map[string]Symbol{}}
	r.resolveEntries(env, keyEnv, params)
}

func (r *resolver) resolveAttributes(env Env, attrs []Attr) {
	keyEnv := defaultEnv{env: map[string]Symbol{}}
	for _, attr := range attrs {
		r.resolveEntries(env, keyEnv, attr.Entries)
	}
}

func (r *resolver) resolveEntries(env, keyEnv Env, attrs []KeyVal) {
	for i := range attrs {
		entry := attrs[i]
		r.bindVar(keyEnv, entry.Key)

		switch t := entry.Value.(type) {
		case Var:
			r.resolveVar(env, t)
		case MemberAccess:
			panic(errors.ErrUnsupported.Error())
		}
	}
}

func (r *resolver) resolveExpr(env Env, expr Expr) {
	switch t := expr.(type) {
	case String:
	case Number:
	case Bool:
	case Var:
		r.resolveVar(env, t)
	case MemberAccess:
		r.resolveExpr(env, t.Object)
		r.resolveVar(env, t.Member)
	case Enum:
		// NOTE do nothing
	case IFExpr:
		r.resolveExpr(env, t.Cond)
		r.resolveExpr(env, t.Then)
		r.resolveExpr(env, t.Else)
	case CondExpr:
		r.resolveExpr(env, t.Target)
		for _, c := range t.Cases {
			r.resolveExpr(env, c.Cond)
			r.resolveExpr(env, c.Branch)
		}
	}
}

func (r *resolver) resolvePropertyType(env Env, t PropertyType) {
	switch tt := t.(type) {
	case Var:
		r.resolveVar(env, tt)
	case MemberAccess:
		r.resolveExpr(env, tt.Object)
		r.resolveVar(env, tt.Member)
	case Enum:
		r.resolveExpr(env, tt)
	default:
		panic(fmt.Sprintf("unexpected property type: %v", reflect.TypeOf(t)))
	}
}

func (r *resolver) resolveVar(env Env, v Var) {
	if _, err := env.Lookup(v.Name); err != nil {
		r.addError(ErrUndeclared, v)
	}
}

func (r *resolver) bindVar(env Env, v Var) {
	if err := env.Bind(TypeUnchecked, v.Name); err != nil {
		r.addError(err, v)
	}
}

func (r *resolver) addError(errkind SymbolError, n Var) {
	var err perrors.Error
	name := n.Name
	line := n.Pos.Line
	col := n.Pos.Col

	switch errkind {
	case ErrUndeclared:
		msg := fmt.Sprintf("undeclared var %v (%d, %d)", name, line, col)
		err = perrors.Error{Message: msg}

	case ErrDuplicateDeclaration:
		msg := fmt.Sprintf("duplicate var %v (%d, %d)", name, line, col)
		err = perrors.Error{Message: msg}

	default:
		panic(fmt.Sprintf("unexpected error: %v at %s (%d, %d)", errkind, name, line, col))
	}
	r.src.errs = append(r.src.errs, err)
}
