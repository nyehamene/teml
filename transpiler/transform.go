package ast

import (
	"errors"
	"fmt"
	"reflect"
)

func (t *typechecker) transformElement(decl Var, node genericElement) Element {
	var transformedTo Element = node

	switch element := node.Tag.(type) {
	case String, Number, Bool, Enum, IFExpr, CondExpr:
		panic("unreachable")

	case Var:
		tagtype := t.getElementTagType(decl, element)
		transformedTo = t.transformElementByTagType(element, node, tagtype, false)

	case MemberAccess:
		// TODO transform element with member access tag expression
		// NOTE can be transformed to either component or instance element
		panic(errors.ErrUnsupported)
	}

	return transformedTo
}

func (t *typechecker) transformElementByTagType(ident Var, element genericElement, tag TypeSymbol, property bool) Element {
	var transformedTo Element = element

	switch tagtype := tag.(type) {
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

	case PropertySymbol:
		transformedTo = t.transformElementByTagType(ident, element, tagtype.Type, true)

	case TypeDeclaration:
		if property {
			// TODO fail if parameter is not empty
			transformedTo = PropertyElement{
				Tag:        element.Tag,
				Attributes: element.Attributes,
				Body:       element.Body,
			}
		} else {
			// TODO fail if body is not empty
			t.typecheckElementParameter(t.env.names, element)
			transformedTo = ComponentElement{
				Tag:        element.Tag,
				Parameters: element.Parameter,
				Attributes: element.Attributes,
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

	resolvedEnv, ok := env.LookupNameEnv(resolvedTagName)
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

	keyType, ok := t.lookupType(resolvedKey)
	if !ok {
		t.addError(ErrUndeclaredType, p.Key)
		return
	}

	propertyType, ok := keyType.(PropertySymbol)
	if !ok {
		t.addError(ErrUnexpectedPropertyType, p.Key)
	}

	valueType, ok := t.getExprType(env, p.Value)
	if !ok {
		t.addError(ErrUndeclaredType, p.Key)
		return
	}

	if ok := t.matchType(propertyType.Type, valueType); !ok {
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
		exprType, ok := t.lookupType(resolvedName)
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

func (t *typechecker) getElementTagType(ident Var, elementIdent Var) TypeSymbol {
	resolvedIdentName, ok := t.lookupName(ident.Name)
	if !ok {
		t.addError(ErrUndeclared, ident)
		return nil
	}

	env, ok := t.lookupNameEnv(resolvedIdentName)
	if !ok {
		t.addError(ErrUndeclared, elementIdent)
		return nil
	}

	resolvedElementName, ok := env.LookupName(elementIdent.Name)
	if !ok {
		t.addError(ErrUndeclared, elementIdent)
		return nil
	}

	elementType, ok := t.lookupType(resolvedElementName)
	if !ok {
		t.addError(ErrUndeclared, elementIdent)
		return nil
	}

	return elementType
}
