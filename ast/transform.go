package ast

import (
	"fmt"
)

func (t *typechecker) transformElement(scope *Scope, node Element) (Stmt, bool) {
	// TBD(fix): find a better way to get the location of the element tag in source.
	// Right now location information is only stored in Vars
	tag, isVar := t.asVar(scope, node.Tag)
	if !isVar {
		t.error(fmt.Errorf("%w: %v", ErrInvalidElementTag, node.Tag))
		return node, false
	}

	tagEntity, exists := scope.lookup(tag.Name)
	if !exists {
		t.errorVar(ErrUndeclared, tag)
		return node, false
	}

	var (
		element            Stmt = node
		field                   = false
		rejectParameter         = false
		rejectAttribute         = false
		rejectChildren          = false
		transformationOkay      = false
	)

check:
	switch typeid := tagEntity.(type) {
	case *TypeBool:
		t.errorVar(ErrInvalidElementTag, tag)

	case *TypeNumber:
		rejectParameter = true
		rejectChildren = true
		transformationOkay = true
		element = NumberElement{
			Tag:        tag,
			Attributes: node.Attributes,
		}

	case *TypeString:
		rejectParameter = true
		rejectChildren = true
		transformationOkay = true
		element = StringElement{
			Tag:        tag,
			Attributes: node.Attributes,
		}

	case *HTMLElementType:
		rejectParameter = true
		transformationOkay = true
		element = HTMLElement{
			Tag:        tag,
			Attributes: node.Attributes,
			Children:   node.Children,
		}

	case *TypeEnum:
		t.errorVar(ErrInvalidElementTag, tag)

	case *TypeField:
		field = true
		tagEntity = typeid.Type
		goto check

	case *TypeModule:
		t.errorVar(ErrInvalidElementTag, tag)

	case *TypePackage:
		t.errorVar(ErrInvalidElementTag, tag)

	case *TypeTemplate:
		transformationOkay = true
		if field {
			rejectParameter = true
			element = PropertyElement{
				Tag:        tag,
				Attributes: node.Attributes,
				Children:   node.Children,
			}
		} else {
			rejectChildren = true
			element = ComponentElement{
				Tag:        typeid.Var,
				Parameters: node.Parameter,
				Attributes: node.Attributes,
			}
		}

	default:
		panic(fmt.Sprintf("unexpected ast.Entity: %#v", typeid))
	}

	if rejectAttribute && len(node.Attributes) != 0 {
		// TBD(feat): get the attribute position in the source
		t.errorVar(ErrUnexpectedAttribute, tag)
	}
	if rejectParameter && len(node.Parameter) != 0 {
		// TBD(feat): get the parameter position in the source
		t.errorVar(ErrUnexpectedParameter, tag)
	}
	if rejectChildren && len(node.Children) != 0 {
		// TBD(feat): get the children position in the source
		t.errorVar(ErrUnexpectedChildren, tag)
	}

	return element, transformationOkay
}
