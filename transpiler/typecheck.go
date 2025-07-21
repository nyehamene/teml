package ast

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/eml-lang/teml/internal/assert"
	perrors "github.com/eml-lang/teml/internal/errors"
)

type typechecker struct {
	src *File
}

func (t *typechecker) typecheckFile(env Env) Env {
	pkg := t.typecheckPackage(env)
	t.typecheckDeclaration(env, pkg)

	for _, tmpl := range t.src.Declarations {
		var name string
		var props []Property
		var stmts []Stmt

		switch tt := tmpl.(type) {
		case Document:
			name = tt.Ident.Name
			props = tt.Properties
			stmts = tt.Stmts

		case Component:
			name = tt.Ident.Name
			props = tt.Properties
			stmts = tt.Stmts

		default:
			panic(fmt.Sprintf("unexpected declaration: %v", reflect.TypeOf(t)))
		}

		dtype, err := env.Lookup(name)
		assert.Assert(err == nil, "expected nil")

		dtype = dtype.(TypeDeclaration)

		nestedEnv, ok := env.LookupEnv(name)
		if !ok {
			panic(fmt.Sprintf("invalid state: env not found for %s", name))
		}

		t.typecheckProperties(nestedEnv, dtype, props)
		t.typecheckStmts(nestedEnv, stmts)
	}

	return env
}

func (t *typechecker) typecheckPackage(env Env) string {
	f := t.src
	ident := f.Package.Ident.Name
	name := f.Package.Path

	// remove double quote
	name = name[1 : len(name)-1]

	pkgtype := TypePackage{Path: name}
	t.bind(env, pkgtype, ident)
	return name
}

func (t *typechecker) typecheckDeclaration(env Env, ns string) {
	for _, decl := range t.src.Declarations {
		var name string
		var sym Symbol

		switch tt := decl.(type) {
		case Document:
			name = tt.Ident.Name
			typename := ns + "/" + name
			sym = TypeDeclaration{
				Kind: DocumentDeclaration,
				Name: typename,
			}

		case Component:
			name = tt.Ident.Name
			typename := ns + "/" + name
			sym = TypeDeclaration{
				Kind: ComponentDeclaration,
				Name: typename,
			}

		default:
			panic(fmt.Sprintf("unexpected declaration: %v", reflect.TypeOf(t)))
		}

		t.bind(env, sym, name)
	}
}

func (t *typechecker) typecheckProperties(env Env, dtype Symbol, props []Property) {
	for _, p := range props {
		ident := p.Ident

		switch proptype := p.Type.(type) {
		case Var:
			resolvedType := t.lookup(env, proptype)

			switch resolvedType := resolvedType.(type) {
			case BuiltinType:
				switch resolvedType {
				case TypeUnchecked:
					t.addError(ErrUndefined, proptype)
				}

			case TypeDeclaration:
				switch resolvedType {
				case dtype:
					t.addError(ErrRecursiveDefinition, proptype)
				default:
					switch resolvedType.Kind {
					case DocumentDeclaration:
						t.addError(ErrMismatchElementTag, proptype)
					}
				}
			}

			t.bind(env, TypeVar{resolvedType}, ident.Name)

		case Enum:
			constantType := t.validateEnumConstants(proptype.Constants)
			enumtype := TypeEnum{
				ConstantType: constantType,
			}
			t.bind(env, enumtype, ident.Name)

		case MemberAccess:
			panic(errors.ErrUnsupported)

		default:
			panic(fmt.Sprintf("unepxected property type: %v", reflect.TypeOf(proptype)))
		}
	}
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

func (t *typechecker) typecheckStmts(env Env, stmts []Stmt) {
	for i, stmt := range stmts {
		element := t.typecheckElement(env, stmt.Element)
		stmts[i] = Stmt{element}
	}
}

func (t *typechecker) typecheckElement(env Env, expr Element) Element {
	switch tt := expr.(type) {
	case TextElement:
		return expr

	case TextGroupElement:
		return expr

	case InstanceElement, ComponentElement, NativeElement, NumberElement, StringElement:
		// NOTE these are not produce by the parser
		panic("Unreachable")

	case genericElement:
		t.typecheckTag(env, tt.Tag)
		// TODO typecheck properties
		t.typecheckAttributes(env, tt.Attributes)
		t.typecheckStmts(env, tt.Body)
		return t.transformElement(env, tt)

	case IFElement:
		t.typecheckExpr(env, tt.Cond)
		t.typecheckElement(env, tt.Then)
		if tt.Else != nil {
			t.typecheckElement(env, tt.Else)
		}
		return tt

	case CondElement:
		t.typecheckExpr(env, tt.Target)
		for _, c := range tt.Cases {
			t.typecheckExpr(env, c.Cond)
			t.typecheckElement(env, c.Branch.Element)
		}
		return tt
	}
	panic(fmt.Sprintf("Unreachable: %v", reflect.TypeOf(expr)))
}

func (t *typechecker) transformElement(env Env, node genericElement) Element {
	var element Element
	switch nodetype := node.Tag.(type) {
	case String, Number, Bool, Enum:
		// NOTE literals cannot be used as a tag name
		// NOTE an error should have been reported already
		// TODO add a test case
		element = node

	case IFExpr:
		// NOTE an if expression cannot be used as a tag name
		// TODO add a test case
		element = node

	case CondExpr:
		// NOTE an if expression cannot be used as a tag name
		// TODO add a test case
		element = node

	case Var:
		tag := t.lookup(env, nodetype)
		element = t.transformElementByTagType(env, node, tag, false)

	case MemberAccess:
		// TODO transform element with member access tag expression
		// NOTE can be transformed to either component or instance element
		panic(errors.ErrUnsupported)
	}

	return element
}

func (t *typechecker) transformElementByTagType(env Env, genElem genericElement, tag Symbol, component bool) Element {
	var element Element

	switch tagtype := tag.(type) {
	case TypePackage:
		// NOTE an package cannot be used as a tag
		// NOTE an error should have been reported already
		// TODO add a test case
		element = genElem

	case TypeEnum:
		// NOTE an enum cannot be used as a tag name
		// NOTE an error should have been reported already
		// TODO add a test case
		element = genElem

	case NativeElementType:
		element = NativeElement{
			Tag:        genElem.Tag,
			Attributes: genElem.Attributes,
			Body:       genElem.Body,
		}

	case TypeVar:
		return t.transformElementByTagType(env, genElem, tagtype.Type, true)

	case TypeDeclaration:
		if component {
			element = ComponentElement{
				Tag:        genElem.Tag,
				Attributes: genElem.Attributes,
				Body:       genElem.Body,
			}
		} else {
			// TODO typecheck parameters
			element = InstanceElement{
				Tag:        genElem.Tag,
				Parameter:  genElem.Parameter,
				Attributes: genElem.Attributes,
				Body:       genElem.Body,
			}
		}

	case BuiltinType:
		switch tagtype {
		case TypeString:
			element = StringElement{
				Tag:        genElem.Tag,
				Attributes: genElem.Attributes,
			}
		case TypeNumber:
			element = NumberElement{
				Tag:        genElem.Tag,
				Attributes: genElem.Attributes,
			}
		case TypeBool, TypeUnchecked:
			// NOTE an should have already been reported when type checking element tag name expression
			// TODO add a test case
			element = genElem
		default:
			panic(fmt.Sprintf("unexpected builtin type: %v", reflect.TypeOf(tag)))
		}

	default:
		panic(fmt.Sprintf("unexpected element symbol type: %v", reflect.TypeOf(tag)))
	}

	return element
}

func (t *typechecker) typecheckExpr(env Env, expr Expr) {
	switch tt := expr.(type) {
	case String, Number, Bool:
		// no oop

	case IFExpr:
		t.typecheckExpr(env, tt.Cond)
		t.typecheckExpr(env, tt.Then)
		t.typecheckExpr(env, tt.Else)

	case CondExpr:
		t.typecheckExpr(env, tt.Target)
		for _, c := range tt.Cases {
			t.typecheckExpr(env, c.Cond)
			t.typecheckExpr(env, c.Branch)
		}

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

func (t *typechecker) typecheckTag(env Env, expr Expr) {
	switch node := expr.(type) {
	case Var:
		nodetype := t.lookup(env, node)
		t.typecheckTagSymbol(env, node, nodetype)

	case MemberAccess:
		// TODO type check member access tag expression
		panic(errors.ErrUnsupported)
	}
}

func (t *typechecker) typecheckTagSymbol(env Env, node Var, sym Symbol) {
	switch symtype := sym.(type) {
	case BuiltinType:
		switch symtype {
		case TypeBool:
			t.addError(ErrMismatchElementTag, node)
		}

	case TypeEnum:
		t.addError(ErrMismatchElementTag, node)

	case TypePackage:
		t.addError(ErrMismatchElementTag, node)

	case TypeDeclaration:
		switch symtype.Kind {
		case DocumentDeclaration:
			t.addError(ErrMismatchElementTag, node)
		}

	case TypeVar:
		t.typecheckTagSymbol(env, node, symtype.Type)
	}
}

func (t *typechecker) typecheckAttributes(env Env, attrs []Attr) {
	for _, attr := range attrs {
		t.typecheckKeyVals(env, attr.Entries)
	}
}

func (t *typechecker) typecheckKeyVals(env Env, kvs []KeyVal) {
	for _, kv := range kvs {
		t.typecheckExpr(env, kv.Value)
	}
}

func (t *typechecker) bind(env Env, sym Symbol, name string) {
	err := env.Bind(sym, name)
	if err != nil {
		panic(fmt.Sprintf("failed to bind %s to %v: %v", name, sym, err))
	}
}

func (t *typechecker) lookup(env Env, node Var) Symbol {
	sym, err := env.Lookup(node.Name)
	if err != nil {
		t.addError(ErrUndeclared, node)
	}
	return sym
}

func (t *typechecker) addError(errkind SymbolError, n Var) {
	var err perrors.Error
	name := n.Name
	line := n.Pos.Line
	col := n.Pos.Col

	switch errkind {
	case ErrUndeclared:
		msg := fmt.Sprintf("undeclared type %v (%d, %d)", name, line, col)
		err = perrors.Error{Message: msg}

	case ErrUndefined:
		msg := fmt.Sprintf("undefined type %v (%d, %d)", name, line, col)
		err = perrors.Error{Message: msg}

	case ErrRecursiveDefinition:
		msg := fmt.Sprintf("recursive type %v (%d, %d)", name, line, col)
		err = perrors.Error{Message: msg}

	case ErrMismatchElementTag:
		msg := fmt.Sprintf("type mismatch: element tag is not a component/element: %v (%d, %d)", name, line, col)
		err = perrors.Error{Message: msg}

	default:
		panic(fmt.Sprintf("unexpected error: \"%v\" at %s (%d, %d)", errkind, name, line, col))
	}
	t.src.errs = append(t.src.errs, err)
}
