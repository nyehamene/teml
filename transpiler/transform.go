package ast

import (
	"errors"
	"fmt"
	"reflect"
)

func (t *typechecker) transformElement(compIdent Var, node genericElement) Element {
	var transformedTo Element = node

	switch element := node.Tag.(type) {
	case String, Number, Bool, Enum, IFExpr, CondExpr:
		panic("unreachable")

	case Var:
		resolvedTagType, resolvedTag := t.getElementTagType(compIdent, element)
		transformedTo = t.transformElementByTagType(compIdent, node, resolvedTagType, resolvedTag)

	case MemberAccess:
		// TODO transform element with member access tag expression
		// NOTE can be transformed to either component or instance element
		panic(errors.ErrUnsupported)
	}

	return transformedTo
}

// func (t *typechecker) transformElementByTagType(ident Var, element genericElement, tag TypeSymbol, property bool) Element
func (t *typechecker) transformElementByTagType(ident Var, element genericElement, tagtype TypeSymbol, tag Binding) Element {
	var transformedTo Element = element

	switch tagtype := tagtype.(type) {
	case TypePackage:
		t.addError(ErrInvalidElementTag, ident)

	case TypeEnum:
		t.addError(ErrInvalidElementTag, ident)

	case NativeElementType:
		// TODO fail if parameter is not empty
		transformedTo = NativeElement{
			Tag:        element.Tag,
			Attributes: element.Attributes,
			Body:       element.Body,
		}

	case TypeDeclaration:
		if tag.IsType {
			// TODO fail if body is not empty
			t.typecheckElementParameter(t.env.names, element)
			transformedTo = ComponentElement{
				Tag:        element.Tag,
				Parameters: element.Parameter,
				Attributes: element.Attributes,
			}
		} else {
			// TODO fail if parameter is not empty
			transformedTo = PropertyElement{
				Tag:        element.Tag,
				Attributes: element.Attributes,
				Body:       element.Body,
			}
		}

	case BuiltinType:
		switch tagtype {
		case TypeBool:
			t.addError(ErrInvalidElementTag, ident)

		case TypeString:
			transformedTo = StringElement{
				Tag:        element.Tag,
				Attributes: element.Attributes,
			}
		case TypeNumber:
			transformedTo = NumberElement{
				Tag:        element.Tag,
				Attributes: element.Attributes,
			}
		default:
			panic(fmt.Sprintf("unexpected builtin type: %v", reflect.TypeOf(tag)))
		}

	default:
		panic(fmt.Sprintf("unexpected element symbol type: %v", reflect.TypeOf(tag)))
	}

	return transformedTo
}

func (t *typechecker) typecheckElementParameter(env NameEnv, element genericElement) {
	tag, ok := t.getName(element.Tag)
	if !ok {
		// TODO replace Var{} below with the tag ident
		t.addError(ErrInvalidElementTag, Var{})
	}

	resolvedTagName, ok := env.LookupName(tag.Name)
	if !ok {
		t.addError(ErrUndeclared, tag)
	}

	resolvedEnv, ok := env.LookupNameEnv(resolvedTagName.ID)
	if !ok {
		t.addError(ErrNamespaceNotfound, tag)
	}

	for _, p := range element.Parameter {
		t.typecheckComponentElementParameter(resolvedEnv, p)
	}
}

func (t *typechecker) getName(expr Expr) (Var, bool) {
	switch node := expr.(type) {
	case MemberAccess:
		// objName, ok := r.getName(node.Object)
		// if !ok {
		// 	return "", false
		// }

		// resolvedObjName, ok := env.LookupName(objName)
		// if !ok {
		// 	return "", false
		// }

		// objEnv, ok := env.LookupNameEnv(resolvedObjName)
		// if !ok {
		// 	return "", false
		// }

		// resolvedName, ok := objEnv.LookupName(node.Member.Name)
		// if !ok {
		// 	return "", false
		// }

		// return resolvedName, true
		// TODO TDB
		return Var{}, false

	case Var:
		return node, true

	default:
		return Var{}, false
	}
}

func (t *typechecker) typecheckComponentElementParameter(env NameEnv, p KeyVal) {
	resolvedKey, ok := env.LookupName(p.Key.Name)
	if !ok {
		t.addError(ErrUndeclared, p.Key)
		return
	}

	keyType, ok := t.lookupType(resolvedKey.ID)
	if !ok {
		t.addError(ErrUndeclaredType, p.Key)
		return
	}

	valueType, ok := t.getExprType(env, p.Value)
	if !ok {
		t.addError(ErrUndeclaredType, p.Key)
		return
	}

	if ok := t.matchType(keyType, valueType); !ok {
		t.addError(ErrTypeMismatch, p.Key)
	}
}

func (t *typechecker) getExprType(env NameEnv, expr Expr) (TypeSymbol, bool) {
	switch node := expr.(type) {
	case CondExpr:
		// TODO tbd
		return nil, false
	case Enum:
		// TODO tbd
		return nil, false
	case IFExpr:
		// TODO tbd
		return nil, false
	case MemberAccess:
		// TODO tbd
		return nil, false
	case Bool:
		return TypeBool, true
	case Number:
		return TypeNumber, true
	case String:
		return TypeString, true
	case Var:
		resolvedName, ok := env.LookupName(node.Name)
		if !ok {
			return nil, false
		}
		exprType, ok := t.lookupType(resolvedName.ID)
		if !ok {
			return nil, false
		}

		return exprType, true
	default:
		panic(fmt.Sprintf("unexpected ast.Expr: %#v", node))
	}
}

func (t *typechecker) matchType(t1, t2 TypeSymbol) bool {
	if t1 == t2 {
		return true
	}
	return false
}

func (t *typechecker) getElementTagType(compIdent Var, elementIdent Var) (TypeSymbol, Binding) {
	resolvedComponent, ok := t.lookupName(compIdent.Name)
	if !ok {
		t.addError(ErrUndeclared, compIdent)
		return nil, Binding{}
	}

	env, ok := t.lookupNameEnv(resolvedComponent.ID)
	if !ok {
		// TODO include elementIdent (Binding)
		t.addError(ErrUndeclared, compIdent)
		return nil, Binding{}
	}

	resolvedElement, ok := env.LookupName(elementIdent.Name)
	if !ok {
		// TODO include elementIdent (Binding)
		t.addError(ErrUndeclared, compIdent)
		return nil, Binding{}
	}

	resolvedType, ok := t.lookupType(resolvedElement.ID)
	if !ok {
		// TODO include elementIdent (Binding)
		t.addError(ErrUndeclared, compIdent)
		return nil, Binding{}
	}

	return resolvedType, resolvedElement
}
