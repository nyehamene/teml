package ast

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/eml-lang/teml/internal/source"

	"github.com/eml-lang/teml/internal/flags"
)

type resolver struct {
	scopeName string
}

// Deprecated: use ParseFile instead
// ResolveFile0
func ResolveFile0(src *File, cflags ...flags.Flag) NameEnv {
	var flag flags.Flag
	for _, f := range cflags {
		flag |= f
	}

	e := resolveFile(src, flag)
	return e
}

func ResolveFile(src source.File, cflags ...flags.Flag) (*File, NameEnv) {
	var flag flags.Flag
	for _, f := range cflags {
		flag |= f
	}

	file := ParseFile(src, cflags...)
	env := resolveFile(file, flag)
	return file, env
}

func resolveFile(f *File, flag flags.Flag) NameEnv {
	rootEnv := newNameRootEnv()
	if flag&flags.FlagNoBuiltinType == 0 {
		bindBuiltInTypeNames(rootEnv)
	}

	namespaceEnv := rootEnv.Nest(f.Name)
	nativeEnv := namespaceEnv.Nest(nsNative)
	if flag&flags.FlagNoNativeElement == 0 {
		bindNativeElementNames(nativeEnv)
	}

	r := &resolver{scopeName: f.Name}
	if err := r.bindPackage(f, namespaceEnv); err != nil {
		f.addError(err)
	}
	// TODO handle import and using declarations

	if errs := r.bindDeclarations(f, namespaceEnv); len(errs) > 0 {
		for _, err := range errs {
			f.addError(err)
		}
	}

	for _, decl := range f.Declarations {
		ident := Var(decl.Ident)
		props := decl.Properties
		stmts := decl.Stmts

		r.scopeName = r.getQualifiedName(ident)

		resolvedName, ok := namespaceEnv.LookupName(ident.Name)
		if !ok {
			f.addSymbolError(ErrUndeclared, ident)
			continue
		}

		var declEnv NameEnv
		if flag&flags.FlagNoNativeElement == 0 {
			declEnv = nativeEnv.Nest(resolvedName.ID)
		} else {
			declEnv = namespaceEnv.Nest(resolvedName.ID)
		}

		if errs := r.bindProperties(declEnv, props); len(errs) > 0 {
			for _, err := range errs {
				f.addError(err)
			}
		}
		if errs := r.resolveStmts(f, declEnv, stmts); len(errs) > 0 {
			for _, err := range errs {
				f.addError(err)
			}
		}

		r.scopeName = f.Name
	}

	return nativeEnv
}

func (t *resolver) bindPackage(f *File, env NameEnv) error {
	const isType = false
	return t.bindVar(env, f.Package.Ident, isType)
}

func (t *resolver) bindDeclarations(f *File, env NameEnv) []error {
	declarations := f.Declarations
	var errs []error
	for _, d := range declarations {
		if err := t.bindDeclaration(env, d); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (r *resolver) bindDeclaration(env NameEnv, rec Template) error {
	const isType = true
	return r.bindVar(env, rec.Ident, isType)
}

func (t *resolver) bindProperties(env NameEnv, props []Property) []error {
	const isType = false
	var errs []error
	for _, p := range props {
		if err := t.resolvePropertyType(env, p.Type); err != nil {
			errs = append(errs, err)
		}
		if err := t.bindVar(env, p.Ident, isType); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (t *resolver) resolveStmts(f *File, env NameEnv, stmts []Stmt) []error {
	var errs []error
	for _, stmt := range stmts {
		if err := t.resolveElement(f, env, stmt.Element); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (r *resolver) resolveElement(f *File, env NameEnv, expr Element) error {
	switch t := expr.(type) {
	case TextElement, TextGroupElement: // noop
	case PropertyElement, NativeElement, NumberElement, StringElement:
		// NOTE these are not produce by the parser
		panic("unreachable")

	case genericElement:
		if err := r.resolveExpr(env, t.Tag); err != nil {
			return err
		}
		if errs := r.bindAttributes(env, t.Attributes); len(errs) > 0 {
			for _, err := range errs {
				f.addError(err)
			}
		}
		if errs := r.resolveStmts(f, env, t.Body); len(errs) > 0 {
			for _, err := range errs {
				f.addError(err)
			}
		}

	case IFElement:
		if err := r.resolveExpr(env, t.Cond); err != nil {
			return err
		}
		if err := r.resolveElement(f, env, t.Then); err != nil {
			return err
		}
		if t.Else != nil {
			if err := r.resolveElement(f, env, t.Else); err != nil {
				return err
			}
		}

	case CondElement:
		if err := r.resolveExpr(env, t.Target); err != nil {
			return err
		}
		for _, c := range t.Cases {
			if err := r.resolveExpr(env, c.Cond); err != nil {
				return err
			}
			if err := r.resolveElement(f, env, c.Branch.Element); err != nil {
				return err
			}
		}

	default:
		panic(fmt.Sprintf("unexpected expression statement: %v", reflect.TypeOf(expr)))
	}
	return nil
}

func (t *resolver) bindAttributes(env NameEnv, attrs []Attr) []error {
	attributeEnv := newNameEnv(nil, "<attr>")
	var errs []error
	for _, attr := range attrs {
		if err := t.bindAttribute(env, attributeEnv, attr.Entries); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (r *resolver) bindAttribute(env, keyEnv NameEnv, attrs []KeyVal) error {
	const isType = false
	for i := range attrs {
		entry := attrs[i]
		if err := r.bindVar(keyEnv, entry.Key, isType); err != nil {
			return err
		}

		switch t := entry.Value.(type) {
		case Var:
			if err := r.resolveVar(env, t); err != nil {
				return err
			}
		case MemberAccess:
			panic(errors.ErrUnsupported)
		}
	}
	return nil
}

func (r *resolver) resolveExpr(env NameEnv, expr Expr) error {
	switch t := expr.(type) {
	case String, Number, Bool, Enum: // noop
	case Var:
		return r.resolveVar(env, t)
	case MemberAccess:
		if err := r.resolveExpr(env, t.Object); err != nil {
			return err
		}
		return r.resolveVar(env, t.Member)
	case IFExpr:
		if err := r.resolveExpr(env, t.Cond); err != nil {
			return err
		}
		if err := r.resolveExpr(env, t.Then); err != nil {
			return err
		}
		return r.resolveExpr(env, t.Else)
	case CondExpr:
		if err := r.resolveExpr(env, t.Target); err != nil {
			return err
		}
		for _, c := range t.Cases {
			if err := r.resolveExpr(env, c.Cond); err != nil {
				return err
			}
			if err := r.resolveExpr(env, c.Branch); err != nil {
				return err
			}
		}
	default:
		panic(fmt.Errorf("unexpected expression type :%v", reflect.TypeOf(expr)))
	}
	return nil
}

func (r *resolver) resolvePropertyType(env NameEnv, t PropertyType) error {
	switch tt := t.(type) {
	case Var:
		return r.resolveVar(env, tt)
	case MemberAccess:
		if err := r.resolveExpr(env, tt.Object); err != nil {
			return err
		}
		return r.resolveVar(env, tt.Member)
	case Enum:
		return r.resolveExpr(env, tt)
	default:
		panic(fmt.Errorf("unexpected property type: %v", reflect.TypeOf(t)))
	}
}

func (t *resolver) resolveVar(env NameEnv, v Var) error {
	if _, ok := env.LookupName(v.Name); !ok {
		return &SymbolError{err: ErrUndeclared, symbol: v}
	}
	return nil
}

func (t *resolver) bindVar(env NameEnv, v Var, isType bool) error {
	fqn := t.getQualifiedName(v)
	if isType {
		env.BindTypeName(v.Name, fqn)
	} else {
		env.BindName(v.Name, fqn)
	}
	return nil
}

func (t *resolver) getQualifiedName(v Var) string {
	fqn := t.scopeName + "." + v.Name
	return fqn
}

func (t *resolver) addError(errkind error, n Var) {
	var err error
	name := n.Name
	line := n.Line
	col := n.Col

	switch errkind {
	case ErrUndeclared:
		err = fmt.Errorf("undeclared var %v (%d, %d)", name, line, col)

	case ErrDuplicateDeclaration:
		err = fmt.Errorf("duplicate var %v (%d, %d)", name, line, col)

	case ErrNamespaceNotfound:
		err = fmt.Errorf("undeclared type %v (%d, %d)", name, line, col)

	default:
		panic(fmt.Errorf("unexpected error %v at %s (%d, %d)", errkind, name, line, col))
	}
	t.src.errs = append(t.src.errs, err)
}
