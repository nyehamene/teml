package ast

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/eml-lang/teml/internal/flags"
	"github.com/eml-lang/teml/internal/source"
)

type typechecker struct {
	src *File
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

	t := typechecker{src: src, env: env}
	t.typecheckFile()
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

	t := typechecker{src: file, env: env}
	t.typecheckFile()
	return file, env
}

func (t *typechecker) typecheckFile() {
	t.typecheckPackage()
	t.typecheckDeclaration()

	for _, declaration := range t.src.Declarations {
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
			t.addError(ErrUndeclared, typeIdent)
			return
		}

		resolvedType, ok := t.lookupType(resolvedName.ID)
		if !ok {
			t.addError(ErrUndeclared, typeIdent)
			return
		}

		t.typecheckProperties(typeIdent, resolvedType, props)
		t.typecheckStmts(typeIdent, stmts)
	}
}

func (t *typechecker) typecheckPackage() {
	resolvedName, ok := t.lookupNonNativeName(t.src.Package.Ident.Name)
	if !ok {
		t.addError(ErrUndeclared, t.src.Package.Ident)
		return
	}
	path := t.src.Package.Path
	pkgtype := TypePackage{Path: path}
	t.bind(pkgtype, resolvedName.ID)
}

func (t *typechecker) typecheckDeclaration() {
	for _, decl := range t.src.Declarations {
		var node Var
		var kind DeclarationKind

		switch tt := decl.(type) {
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
			t.addError(ErrUndeclared, node)
			return
		}

		sym := TypeDeclaration{
			Kind:   kind,
			Name:   node.Name,
			TypeId: resolvedName.ID,
		}

		t.bind(sym, resolvedName.ID)
	}
}

func (t *typechecker) typecheckProperties(decl Var, typesym TypeSymbol, props []Property) {
	for _, p := range props {
		t.typecheckProperty(decl, typesym, p)
	}
}

func (t *typechecker) typecheckProperty(decl Var, sym Symbol, property Property) {
	var propertyType TypeSymbol

	switch typeNode := property.Type.(type) {
	case Var:
		resolvedName, ok := t.lookupName(typeNode.Name)
		if !ok {
			t.addError(ErrUndeclared, typeNode)
			return
		}

		propertyType, ok = t.lookupType(resolvedName.ID)
		if !ok {
			t.addError(ErrUndeclared, typeNode)
			return
		}

		// component cannot used as a type in its property list
		if propertyType == sym {
			t.addError(ErrRecursiveDefinition, typeNode)
			return
		}

		// a document cannot be used as a property type
		switch resolvedType := propertyType.(type) {
		case TypeDeclaration:
			if resolvedType.Kind == DocumentDeclaration {
				t.addError(ErrInvalidElementTag, typeNode)
				return
			}
		}

	case Enum:
		constantType := t.validateEnumConstants(typeNode.Constants)
		propertyType = TypeEnum{ConstantType: constantType}

	case MemberAccess:
		panic(errors.ErrUnsupported)

	default:
		panic(fmt.Sprintf("unepxected property type: %v", reflect.TypeOf(typeNode)))
	}

	resolvedEnvName, ok := t.lookupName(decl.Name)
	if !ok {
		panic("unreachable")
	}

	propertyEnv, ok := t.lookupNameEnv(resolvedEnvName.ID)
	if !ok {
		panic("unreachable")
	}

	resolvedName, ok := propertyEnv.LookupName(property.Ident.Name)
	if !ok {
		t.addError(ErrUndeclared, property.Ident)
		return
	}

	t.bind(propertyType, resolvedName.ID)
}

func (t *typechecker) validateEnumConstants(cons []EnumConstant) Symbol {

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

func (t *typechecker) typecheckStmts(decl Var, stmts []Stmt) {
	for i, stmt := range stmts {
		element := t.typecheckElement(decl, stmt.Element)
		stmts[i] = Stmt{element}
	}
}

func (t *typechecker) typecheckElement(decl Var, node Element) Element {
	switch element := node.(type) {
	case TextElement:
		return node

	case TextGroupElement:
		return node

	case ComponentElement, PropertyElement, NativeElement, NumberElement, StringElement:
		// NOTE these are not produce by the parser
		panic("Unreachable")

	case genericElement:
		t.typecheckTagExpr(decl, element.Tag)
		// TODO typecheck properties
		t.typecheckAttributes(decl, element.Attributes)
		t.typecheckStmts(decl, element.Body)
		return t.transformElement(decl, element)

	case IFElement:
		t.typecheckExpr(decl, element.Cond)
		t.typecheckElement(decl, element.Then)
		if element.Else != nil {
			t.typecheckElement(decl, element.Else)
		}
		return element

	case CondElement:
		t.typecheckExpr(decl, element.Target)
		for _, c := range element.Cases {
			t.typecheckExpr(decl, c.Cond)
			t.typecheckElement(decl, c.Branch.Element)
		}
		return element
	}
	panic(fmt.Sprintf("Unreachable: %v", reflect.TypeOf(node)))
}

// TODO replace with getExprType
func (t *typechecker) typecheckExpr(decl Var, expr Expr) {
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
		t.validateEnumConstants(tt.Constants)

	default:
		panic(fmt.Sprintf("unexpected expr type: %v", reflect.TypeOf(tt)))
	}
}

func (t *typechecker) typecheckTagExpr(ident Var, expr Expr) {
	resolvedIdentName, ok := t.lookupName(ident.Name)
	if !ok {
		t.addError(ErrUndeclared, ident)
		return
	}

	env, ok := t.lookupNameEnv(resolvedIdentName.ID)
	if !ok {
		t.addError(ErrUndeclared, ident)
		return
	}

	switch node := expr.(type) {
	case Var:
		resolvedTag, ok := env.LookupName(node.Name)
		if !ok {
			t.addError(ErrUndeclared, node)
			return
		}

		tagtype, ok := t.lookupType(resolvedTag.ID)
		if !ok {
			t.addError(ErrUndeclared, node)
			return
		}

		if err := t.typecheckTag(tagtype); err != nil {
			t.addError(err, node)
			return
		}
	case MemberAccess:
		// TODO type check member access tag expression
		panic(errors.ErrUnsupported)
	}
}

func (t *typechecker) typecheckTag(sym TypeSymbol) error {
	switch typeNode := sym.(type) {
	case NativeElementType:

	case TypeEnum, TypePackage:
		return ErrInvalidElementTag

	case BuiltinType:
		if sym == TypeBool {
			return ErrInvalidElementTag
		}

	case TypeDeclaration:
		switch typeNode.Kind {
		case DocumentDeclaration:
			return ErrInvalidElementTag
		}

	default:
		panic(fmt.Sprintf("unexpected type symbol: %#v", typeNode))
	}

	return nil
}

func (t *typechecker) typecheckAttributes(decl Var, attrs []Attr) {
	for _, attr := range attrs {
		t.typecheckKeyVals(decl, attr.Entries)
	}
}

func (t *typechecker) typecheckKeyVals(decl Var, kvs []KeyVal) {
	for _, kv := range kvs {
		t.typecheckExpr(decl, kv.Value)
	}
}

func (t *typechecker) bind(sym TypeSymbol, name string) {
	t.env.BindType(sym, name)
}

func (t *typechecker) lookupName(name string) (Binding, bool) {
	return t.env.names.LookupName(name)
}

func (t *typechecker) lookupNonNativeName(name string) (Binding, bool) {
	return t.env.names.LookupNonNativeName(name)
}

func (t *typechecker) lookupType(binding string) (TypeSymbol, bool) {
	return t.env.LookupType(binding)
}

func (t *typechecker) lookupNameEnv(binding string) (NameEnv, bool) {
	return t.env.names.LookupNameEnv(binding)
}

func (t *typechecker) addError(errkind error, node Var) {
	var err error
	name := node.Name
	line, col := node.Line, node.Col

	switch errkind {
	case ErrUndeclared:
		err = fmt.Errorf("undeclared type %v (%d, %d)", name, line, col)

	case ErrRecursiveDefinition:
		err = fmt.Errorf("recursive type %v (%d, %d)", name, line, col)

	case ErrInvalidElementTag:
		err = fmt.Errorf("type mismatch: element tag is not a component/element: %v (%d, %d)", name, line, col)

	case ErrTypeMismatch:
		err = fmt.Errorf("type mismatch: %v (%d, %d)", name, line, col)

	default:
		panic(fmt.Sprintf("unexpected error: \"%v\" at %s (%d, %d)", errkind, name, line, col))
	}
	t.src.errs = append(t.src.errs, err)
}
