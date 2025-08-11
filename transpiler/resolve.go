package ast

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/eml-lang/teml/internal/source"

	perrors "github.com/eml-lang/teml/internal/errors"
	cflags "github.com/eml-lang/teml/internal/flags"
)

type resolver struct {
	src       *File
	scopeName string
}

// Deprecated: use ParseFile instead
// ResolveFile0
func ResolveFile0(src *File, flags ...cflags.Flag) NameEnv {
	var flag cflags.Flag
	for _, f := range flags {
		flag |= f
	}

	e := resolveFile(src, flag)
	return e
}

func ResolveFile(src source.File, flags ...cflags.Flag) NameEnv {
	var flag cflags.Flag
	for _, f := range flags {
		flag |= f
	}

	file := ParseFile(src, flags...)
	e := resolveFile(file, flag)
	return e
}

func resolveFile(f *File, flag cflags.Flag) NameEnv {
	rootEnv := newNameRootEnv()
	if flag&cflags.FlagNoBuiltinType == 0 {
		bindBuiltInTypeNames(rootEnv)
	}

	namespaceEnv := rootEnv.Nest(f.Name)
	nativeEnv := namespaceEnv.Nest(nsNative)
	if flag&cflags.FlagNoNativeElement == 0 {
		bindNativeElementNames(nativeEnv)
	}

	r := &resolver{src: f, scopeName: f.Name}
	r.resolvePackage(namespaceEnv)
	// TODO handle import and using declarations

	f.Declarations = r.resolveDeclarations(namespaceEnv)

	for _, decl := range f.Declarations {
		var ident Var
		var props []Property
		var stmts []Stmt

		switch t := decl.(type) {
		case Document:
			ident = Var(t.Ident)
			props = t.Properties
			stmts = t.Stmts

		case Component:
			ident = Var(t.Ident)
			props = t.Properties
			stmts = t.Stmts

		default:
			panic(fmt.Sprintf("unexpected declaration: %v", reflect.TypeOf(t)))
		}

		r.scopeName = r.getQualifiedName(ident)

		resolvedName, ok := namespaceEnv.LookupName(ident.Name)
		if !ok {
			r.addError(ErrUndeclared, ident)
			continue
		}

		var declEnv NameEnv
		if flag&cflags.FlagNoNativeElement == 0 {
			declEnv = nativeEnv.Nest(resolvedName.ID)
		} else {
			declEnv = namespaceEnv.Nest(resolvedName.ID)
		}

		r.resolveProperties(declEnv, props)
		r.resolveStmts(declEnv, stmts)

		r.scopeName = f.Name
	}

	return nativeEnv
}

func (t *resolver) resolvePackage(env NameEnv) {
	const isType = false
	t.bindVar(env, t.src.Package.Ident, isType)
}

func (t *resolver) resolveDeclarations(env NameEnv) []Declaration {
	declarations := t.src.Declarations
	for _, d := range declarations {
		t.resolveDeclaration(env, d)
	}
	return declarations
}

func (r *resolver) resolveDeclaration(env NameEnv, rec Declaration) Declaration {
	const isType = true
	switch t := rec.(type) {
	case Document:
		r.bindVar(env, t.Ident, isType)
	case Component:
		r.bindVar(env, t.Ident, isType)
	default:
		panic(fmt.Sprintf("unexpected declaration: %v", reflect.TypeOf(rec)))
	}
	return rec
}

func (t *resolver) resolveProperties(env NameEnv, props []Property) {
	const isType = false
	for _, p := range props {
		t.resolvePropertyType(env, p.Type)
		t.bindVar(env, p.Ident, isType)
	}
}

func (t *resolver) resolveStmts(env NameEnv, stmts []Stmt) {
	for _, stmt := range stmts {
		t.resolveElement(env, stmt.Element)
	}
}

func (r *resolver) resolveElement(env NameEnv, expr Element) {
	switch t := expr.(type) {
	case TextElement, TextGroupElement: // noop
	case PropertyElement, NativeElement, NumberElement, StringElement:
		// NOTE these are not produce by the parser
		panic("unreachable")

	case genericElement:
		r.resolveExpr(env, t.Tag)
		// NOTE handle in the typecheck
		// r.resolveElementParameter(env, t)
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

func (t *resolver) resolveAttributes(env NameEnv, attrs []Attr) {
	attributeEnv := newNameEnv(nil, "<attr>")
	for _, attr := range attrs {
		t.resolveEntries(env, attributeEnv, attr.Entries)
	}
}

func (r *resolver) resolveEntries(env, keyEnv NameEnv, attrs []KeyVal) {
	const isType = false
	for i := range attrs {
		entry := attrs[i]
		r.bindVar(keyEnv, entry.Key, isType)

		switch t := entry.Value.(type) {
		case Var:
			r.resolveVar(env, t)
		case MemberAccess:
			panic(errors.ErrUnsupported)
		}
	}
}

func (r *resolver) resolveExpr(env NameEnv, expr Expr) {
	switch t := expr.(type) {
	case String, Number, Bool, Enum: // noop
	case Var:
		r.resolveVar(env, t)
	case MemberAccess:
		r.resolveExpr(env, t.Object)
		r.resolveVar(env, t.Member)
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
	default:
		panic(fmt.Errorf("unexpected expression type :%v", reflect.TypeOf(expr)))
	}
}

func (r *resolver) resolvePropertyType(env NameEnv, t PropertyType) {
	switch tt := t.(type) {
	case Var:
		r.resolveVar(env, tt)
	case MemberAccess:
		r.resolveExpr(env, tt.Object)
		r.resolveVar(env, tt.Member)
	case Enum:
		r.resolveExpr(env, tt)
	default:
		panic(fmt.Errorf("unexpected property type: %v", reflect.TypeOf(t)))
	}
}

func (t *resolver) resolveVar(env NameEnv, v Var) {
	if _, ok := env.LookupName(v.Name); !ok {
		t.addError(ErrUndeclared, v)
	}
}

func (t *resolver) bindVar(env NameEnv, v Var, isType bool) {
	fqn := t.getQualifiedName(v)
	if isType {
		env.BindTypeName(v.Name, fqn)
	} else {
		env.BindName(v.Name, fqn)
	}
}

func (t *resolver) getQualifiedName(v Var) string {
	fqn := t.scopeName + "." + v.Name
	return fqn
}

func (t *resolver) addError(errkind error, n Var) {
	var err perrors.Error
	name := n.Name
	line := n.Line
	col := n.Col

	switch errkind {
	case ErrUndeclared:
		msg := fmt.Sprintf("undeclared var %v (%d, %d)", name, line, col)
		err = perrors.Error{Message: msg}

	case ErrDuplicateDeclaration:
		msg := fmt.Sprintf("duplicate var %v (%d, %d)", name, line, col)
		err = perrors.Error{Message: msg}

	case ErrNamespaceNotfound:
		msg := fmt.Sprintf("undeclared type %v (%d, %d)", name, line, col)
		err = perrors.Error{Message: msg}

	default:
		panic(fmt.Errorf("unexpected error %v at %s (%d, %d)", errkind, name, line, col))
	}
	t.src.errs = append(t.src.errs, err)
}
