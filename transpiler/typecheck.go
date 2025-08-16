package ast

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/eml-lang/teml/internal/flags"
	"github.com/eml-lang/teml/internal/source"
)

type typechecker struct {
	env TypeEnv
}

// Deprecated: use TypecheckFile instead
//
// TypecheckFile0
func TypecheckFile0(src *File, names NameEnv, cflags ...flags.Flag) TypeEnv {
	var flag flags.Flag
	for _, f := range cflags {
		flag |= f
	}

	env := newTypeEnv(names)
	if flag&flags.FlagNoBuiltinType == 0 {
		bindBuiltInTypes(env)
	}
	if flag&flags.FlagNoNativeElement == 0 {
		bindNativeElementTypes(env)
	}

	t := typechecker{env: env}
	t.typecheckFile(src)
	return env
}

func TypecheckFile(src source.File, cflags ...flags.Flag) (*File, TypeEnv) {
	var flag flags.Flag
	for _, f := range cflags {
		flag |= f
	}

	file, bindings := ResolveFile(src, cflags...)
	env := newTypeEnv(bindings)
	if flag&flags.FlagNoBuiltinType == 0 {
		bindBuiltInTypes(env)
	}
	if flag&flags.FlagNoNativeElement == 0 {
		bindNativeElementTypes(env)
	}

	t := typechecker{env: env}
	t.typecheckFile(file)
	t.transformStmts(file)
	return file, env
}

func (t *typechecker) typecheckFile(f *File) {
	if err := t.typecheckPackage(f.Package); err != nil {
		f.addError(err)
	}
	if errs := t.typecheckDeclaration(f.Declarations); len(errs) > 0 {
		for _, err := range errs {
			f.addError(err)
		}
	}

	for _, declaration := range f.Declarations {
		var typeIdent Var
		var props []Property
		var stmts []Stmt

		switch tt := declaration.(type) {
		case Document:
			typeIdent = Var(tt.Ident)
			props = tt.Properties
			stmts = tt.Stmts

		case Component:
			typeIdent = Var(tt.Ident)
			props = tt.Properties
			stmts = tt.Stmts

		default:
			panic(fmt.Sprintf("unexpected declaration: %v", reflect.TypeOf(t)))
		}

		resolvedName, ok := t.lookupName(typeIdent.Name)
		if !ok {
			f.addSymbolError(ErrUndeclared, typeIdent)
			return
		}

		resolvedType, ok := t.lookupType(resolvedName.ID)
		if !ok {
			f.addSymbolError(ErrUndeclared, typeIdent)
			return
		}

		if errs := t.typecheckProperties(typeIdent, resolvedType, props); len(errs) > 0 {
			for _, err := range errs {
				f.addError(err)
			}
		}
		t.typecheckStmts(f, typeIdent, stmts)
	}
}

func (t *typechecker) typecheckPackage(pkg Package) error {
	resolvedName, ok := t.lookupNonNativeName(pkg.Ident.Name)
	if !ok {
		return &SymbolError{err: ErrUndeclared, symbol: pkg.Ident}
	}
	pkgtype := TypePackage{Path: pkg.Path}
	t.bind(pkgtype, resolvedName.ID)
	return nil
}

func (t *typechecker) typecheckDeclaration(decls []Declaration) []error {
	var errs []error
	for _, d := range decls {
		var node Var
		var kind DeclarationKind

		switch tt := d.(type) {
		case Document:
			node = Var(tt.Ident)
			kind = DocumentDeclaration

		case Component:
			node = Var(tt.Ident)
			kind = ComponentDeclaration

		default:
			panic(fmt.Sprintf("unexpected declaration: %v", reflect.TypeOf(t)))
		}

		resolvedName, ok := t.lookupName(node.Name)
		if !ok {
			errs = append(errs, &SymbolError{err: ErrUndeclared, symbol: node})
			return errs
		}

		sym := TypeDeclaration{
			Kind:   kind,
			Name:   node.Name,
			TypeId: resolvedName.ID,
		}

		t.bind(sym, resolvedName.ID)
	}
	return errs
}

func (t *typechecker) typecheckProperties(decl Var, typesym TypeSymbol, props []Property) []error {
	var errs []error
	for _, p := range props {
		if err := t.typecheckProperty(decl, typesym, p); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (t *typechecker) typecheckProperty(decl Var, sym Symbol, property Property) error {
	var resolvedType TypeSymbol

	switch typeNode := property.Type.(type) {
	case Var:
		resolvedTypeName, ok := t.lookupName(typeNode.Name)
		if !ok {
			return &SymbolError{err: ErrUndeclared, symbol: typeNode}
		}

		resolvedType, ok = t.lookupType(resolvedTypeName.ID)
		if !ok {
			return &SymbolError{err: ErrUndeclared, symbol: typeNode}
		}

		// a component/document type cannot appear in it's own property list
		if resolvedType == sym {
			return &SymbolError{err: ErrRecursiveDefinition, symbol: typeNode}
		}

		// a document cannot be used as a property type
		switch resolvedType := resolvedType.(type) {
		case TypeDeclaration:
			if resolvedType.Kind == DocumentDeclaration {
				return &SymbolError{err: ErrDocumentElementTag, symbol: typeNode}
			}
		}

	case Enum:
		constantType := t.typecheckEnumConstants(typeNode.Constants)
		resolvedType = TypeEnum{ConstantType: constantType}

	case MemberAccess:
		panic(errors.ErrUnsupported)

	default:
		panic(fmt.Sprintf("unepxected property type: %v", reflect.TypeOf(typeNode)))
	}

	resolvedScopeName, ok := t.lookupName(decl.Name)
	if !ok {
		panic("unreachable")
	}

	scopeEnv, ok := t.lookupNameEnv(resolvedScopeName.ID)
	if !ok {
		panic("unreachable")
	}

	resolvedName, ok := scopeEnv.LookupName(property.Ident.Name)
	if !ok {
		return &SymbolError{err: ErrUndeclared, symbol: property.Ident}
	}

	t.bind(resolvedType, resolvedName.ID)
	return nil
}

func (t *typechecker) typecheckElementParameters(f *File, element genericElement) error {
	if len(element.Parameter) == 0 {
		return nil
	}

	decl, err := t.getName(element.Tag)
	if err != nil {
		return err
	}

	resolvedDeclName, exists := t.lookupName(decl.Name)
	if !exists {
		return &SymbolError{err: ErrUndeclared, symbol: decl}
	}

	resolvedScope, exists := t.lookupNameEnv(resolvedDeclName.ID)
	if !exists {
		return &SymbolError{err: ErrNamespaceNotfound, symbol: decl}
	}

	for _, p := range element.Parameter {
		valueType, exists := t.getExprType(t.env.names, p.Value)
		if !exists {
			f.addSymbolError(ErrUndeclaredType, p.Key)
			continue
		}

		resolvedKey, exists := resolvedScope.LookupName(p.Key.Name)
		if !exists {
			f.addSymbolError(ErrUndeclared, p.Key)
			continue
		}

		keyType, exists := t.lookupType(resolvedKey.ID)
		if !exists {
			f.addSymbolError(ErrUndeclared, p.Key)
			continue
		}

		if keyType != valueType {
			f.addSymbolError(ErrTypeMismatch, p.Key)
		}
	}

	return nil
}

func (t *typechecker) typecheckEnumConstants(cons []EnumConstant) Symbol {

	validateType := func(ex, got Symbol) Symbol {
		if ex == nil {
			return got
		}

		if ex != got {
			panic(fmt.Sprintf("Conflicting enum constant types: %v and %v are the same", ex, got))
		}
		return ex
	}

	var constantType Symbol
	for _, c := range cons {
		switch c.(type) {
		case String:
			constantType = validateType(constantType, TypeString)
		case Number:
			constantType = validateType(constantType, TypeNumber)
		case Bool:
		case IFExpr:
		case CondExpr:
		case Var:
		case MemberAccess:
		case Enum:
		default:
			panic(fmt.Sprintf("unexpected enum constant: %v", reflect.TypeOf(c)))
		}
	}

	return constantType
}

func (t *typechecker) typecheckStmts(f *File, decl Var, stmts []Stmt) {
	for i, stmt := range stmts {
		element := t.typecheckElement(f, decl, stmt.Element)
		stmts[i] = Stmt{element}
	}
}

func (t *typechecker) typecheckElement(f *File, decl Var, node Element) Element {
	switch element := node.(type) {
	case TextElement:
		return node

	case TextGroupElement:
		return node

	case ComponentElement, PropertyElement, NativeElement, NumberElement, StringElement:
		// NOTE these are not produce by the parser
		panic("Unreachable")

	case genericElement:
		if err := t.typecheckTagExpr(decl, element.Tag); err != nil {
			f.addError(err)
		}
		if err := t.typecheckElementParameters(f, element); err != nil {
			f.addError(err)
		}
		t.typecheckAttributes(element.Attributes)
		t.typecheckStmts(f, decl, element.Body)
		return node

	case IFElement:
		t.typecheckExpr(element.Cond)
		t.typecheckElement(f, decl, element.Then)
		if element.Else != nil {
			t.typecheckElement(f, decl, element.Else)
		}
		return element

	case CondElement:
		t.typecheckExpr(element.Target)
		for _, c := range element.Cases {
			t.typecheckExpr(c.Cond)
			t.typecheckElement(f, decl, c.Branch.Element)
		}
		return element
	}
	panic(fmt.Sprintf("Unreachable: %v", reflect.TypeOf(node)))
}

// TODO replace with getExprType
func (t *typechecker) typecheckExpr(expr Expr) {
	switch tt := expr.(type) {
	case String, Number, Bool:
		// no oop

	case IFExpr:
		// TODO type check if expression

	case CondExpr:
		// TODO type check cond expression

	case Var:
		// TODO typecheck Var

	case MemberAccess:
		panic(errors.ErrUnsupported)

	case Enum:
		t.typecheckEnumConstants(tt.Constants)

	default:
		panic(fmt.Sprintf("unexpected expr type: %v", reflect.TypeOf(tt)))
	}
}

func (t *typechecker) typecheckTagExpr(ident Var, expr Expr) error {
	resolvedIdentName, ok := t.lookupName(ident.Name)
	if !ok {
		return &SymbolError{err: ErrUndeclared, symbol: ident}
	}

	env, ok := t.lookupNameEnv(resolvedIdentName.ID)
	if !ok {
		return &SymbolError{err: ErrUndeclared, symbol: ident}
	}

	switch node := expr.(type) {
	case Var:
		resolvedTag, ok := env.LookupName(node.Name)
		if !ok {
			return &SymbolError{err: ErrUndeclared, symbol: node}
		}

		tagtype, ok := t.lookupType(resolvedTag.ID)
		if !ok {
			return &SymbolError{err: ErrUndeclared, symbol: node}
		}

		if err := t.typecheckTag(tagtype); err != nil {
			return &SymbolError{err: err, symbol: node}
		}
	case MemberAccess:
		// TODO type check member access tag expression
		panic(errors.ErrUnsupported)
	}
	return nil
}

func (t *typechecker) typecheckTag(sym TypeSymbol) error {
	switch typeNode := sym.(type) {
	case NativeElementType:
		// noop
	case TypeEnum:
		return ErrEnumElementTag
	case TypePackage:
		return ErrPackageElementTag
	case BuiltinType:
		if sym == TypeBool {
			return ErrBoolElementTag
		}
	case TypeDeclaration:
		switch typeNode.Kind {
		case DocumentDeclaration:
			return ErrDocumentElementTag
		}
	default:
		panic(fmt.Sprintf("unexpected type symbol: %#v", typeNode))
	}

	return nil
}

func (t *typechecker) typecheckAttributes(attrs []Attr) {
	for _, attr := range attrs {
		t.typecheckAttribute(attr.Entries)
	}
}

func (t *typechecker) typecheckAttribute(kvs []KeyVal) {
	for _, kv := range kvs {
		t.typecheckExpr(kv.Value)
	}
}

func (t *typechecker) bind(sym TypeSymbol, name string) {
	t.env.BindType(sym, name)
}

func (t *typechecker) lookupName(name string) (ResolvedName, bool) {
	return t.env.LookupName(name)
}

func (t *typechecker) lookupNonNativeName(name string) (ResolvedName, bool) {
	return t.env.LookupNonNativeName(name)
}

func (t *typechecker) lookupType(name string) (TypeSymbol, bool) {
	return t.env.LookupType(name)
}

func (t *typechecker) lookupNameEnv(name string) (NameEnv, bool) {
	return t.env.LookupNameEnv(name)
}

func (t *typechecker) getName(expr Expr) (Var, error) {
	switch tt := expr.(type) {
	case Var:
		return tt, nil
	case MemberAccess:
		return t.getName(tt.Object)
	default:
		return Var{}, fmt.Errorf("cannot get name from %v", reflect.TypeOf(expr))
	}
}

func (t *typechecker) getExprType(env NameEnv, expr Expr) (TypeSymbol, bool) {
	switch tt := expr.(type) {
	case String:
		return TypeString, true
	case Number:
		return TypeNumber, true
	case Bool:
		return TypeBool, true
	case Var:
		resolvedName, ok := env.LookupName(tt.Name)
		if !ok {
			return nil, false
		}
		return t.lookupType(resolvedName.ID)
	case MemberAccess:
		panic(errors.ErrUnsupported)
	case Enum:
		constantType := t.validateEnumConstants(tt.Constants)
		return TypeEnum{ConstantType: constantType}, true
	case IFExpr:
		thenType, ok := t.getExprType(env, tt.Then)
		if !ok {
			return nil, false
		}
		elseType, ok := t.getExprType(env, tt.Else)
		if !ok {
			return nil, false
		}
		if thenType != elseType {
			return nil, false
		}
		return thenType, true
	case CondExpr:
		var exprType TypeSymbol
		for i, c := range tt.Cases {
			caseType, ok := t.getExprType(env, c.Branch)
			if !ok {
				return nil, false
			}
			if i == 0 {
				exprType = caseType
				continue
			}
			if exprType != caseType {
				return nil, false
			}
		}
		return exprType, true
	default:
		panic(fmt.Sprintf("unexpected expression type: %v", reflect.TypeOf(expr)))
	}
}