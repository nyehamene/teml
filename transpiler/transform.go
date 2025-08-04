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
			// TODO typecheck parameters
			// TODO fail if body is not empty
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
