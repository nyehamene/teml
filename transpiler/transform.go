package ast

import (
	"errors"
	"fmt"
	"reflect"
)

type TypeBinding struct {
	Type TypeSymbol
	Name Binding
}

func (t *typechecker) transformStmts(src *File) error {
	var err error

	for _, d := range src.Declarations {
		var typeIdent Var
		var stmts []Stmt

		switch tt := d.(type) {
		case Document:
			typeIdent = Var(tt.Ident)
			stmts = tt.Stmts

		case Component:
			typeIdent = Var(tt.Ident)
			stmts = tt.Stmts

		default:
			panic(fmt.Sprintf("unexpected declaration: %v", reflect.TypeOf(t)))
		}

		for i, stmt := range stmts {
			transformed, err_t := t.transformElement(typeIdent, stmt.Element)
			if err_t != nil {
				err = errors.Join(err_t)
			}
			stmts[i] = Stmt{transformed}
		}
	}

	if err != nil {
		return err
	}
	return nil
}

func (t *typechecker) transformElement(compIdent Var, node Element) (Element, error) {
	var transformed Element = node
	var err error

	switch element := node.(type) {
	case TextElement, TextGroupElement:

	case ComponentElement, NativeElement, NumberElement, PropertyElement, StringElement:
		panic("unreachable")

	case CondElement:
		for i, c := range element.Cases {
			transformedBranch, err_t := t.transformElement(compIdent, c.Branch.Element)
			if err_t != nil {
				err = errors.Join(err, err_t)
			}
			element.Cases[i] = CaseStmt{
				Cond:   c.Cond,
				Branch: Stmt{Element: transformedBranch},
			}
		}

		transformed = element

	case IFElement:
		var transformedThen Element
		var transformedElse Element
		var err_t error

		transformedThen, err_t = t.transformElement(compIdent, element.Then)
		if err_t != nil {
			err = errors.Join(err, err_t)
		}

		if element.Else != nil {
			transformedElse, err_t = t.transformElement(compIdent, element.Else)
		}
		if err_t != nil {
			err = errors.Join(err, err_t)
		}

		transformed = IFElement{
			Cond: element.Cond,
			Then: transformedThen,
			Else: transformedElse,
		}

	case genericElement:
		transformed, err = t.transformGenericElement(compIdent, element)

	default:
		panic(fmt.Sprintf("unexpected ast.Element: %#v", element))
	}

	if err != nil {
		return transformed, err
	}
	return transformed, nil
}

func (t *typechecker) transformGenericElement(compIdent Var, node genericElement) (Element, error) {
	switch element := node.Tag.(type) {
	case String, Number, Bool, Enum, IFExpr, CondExpr:
		panic("unreachable")

	case Var:
		var nodeIdent Var = element
		return t.transformGenerictElementByTagType(compIdent, nodeIdent, node)

	case MemberAccess:
		panic(errors.ErrUnsupported)
	}

	return node, nil
}

func (t *typechecker) transformGenerictElementByTagType(compIdent, elementIdent Var, element genericElement) (Element, error) {
	var transformedTo Element = element
	var err error

	binding, err := t.getElementTagType(compIdent, elementIdent)
	if err != nil {
		return element, err
	}

	switch tagtype := binding.Type.(type) {
	case TypePackage:
		err = SymbolError{ErrPackageElementTag, elementIdent}

	case TypeEnum:
		err = SymbolError{ErrEnumElementTag, elementIdent}

	case NativeElementType:
		if len(element.Parameter) != 0 {
			err = SymbolError{ErrParameterInNativeElement, elementIdent}
		}
		transformedTo = NativeElement{
			Tag:        element.Tag,
			Attributes: element.Attributes,
			Body:       element.Body,
		}

	case TypeDeclaration:
		if binding.Name.IsType {
			err = t.typecheckElementParameter(t.env.names, element)
			transformedTo = ComponentElement{
				Tag:        element.Tag,
				Parameters: element.Parameter,
				Attributes: element.Attributes,
			}
		} else {
			if len(element.Parameter) != 0 {
				err = SymbolError{ErrParameterInPropertyElement, elementIdent}
			}
			transformedTo = PropertyElement{
				Tag:        element.Tag,
				Attributes: element.Attributes,
				Body:       element.Body,
			}
		}

	case BuiltinType:
		switch tagtype {
		case TypeBool:
			err = SymbolError{ErrBoolElementTag, elementIdent}

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
			panic(fmt.Sprintf("unexpected builtin type: %v", reflect.TypeOf(binding.Name)))
		}

	default:
		panic(fmt.Sprintf("unexpected element symbol type: %v", reflect.TypeOf(binding.Name)))
	}

	if err != nil {
		return transformedTo, err
	}
	return transformedTo, nil
}

func (t *typechecker) typecheckElementParameter(env NameEnv, element genericElement) error {
	tag, err := t.getName(element.Tag)
	if err != nil {
		return err
	}

	resolvedTagName, ok := env.LookupName(tag.Name)
	if !ok {
		return SymbolError{ErrUndeclared, tag}
	}

	resolvedEnv, ok := env.LookupNameEnv(resolvedTagName.ID)
	if !ok {
		return SymbolError{ErrNamespaceNotfound, tag}
	}

	for _, p := range element.Parameter {
		if err := t.typecheckComponentElementParameter(resolvedEnv, p); err != nil {
			t.addError(err)
		}
	}

	return nil
}

func (t *typechecker) getName(expr Expr) (Var, error) {
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
		panic(errors.ErrUnsupported)

	case Var:
		return node, nil

	default:
		panic(fmt.Sprintf("unexpected element tag %#v", reflect.TypeOf(expr)))
	}
}

func (t *typechecker) typecheckComponentElementParameter(env NameEnv, p KeyVal) error {
	resolvedKey, ok := env.LookupName(p.Key.Name)
	if !ok {
		return SymbolError{ErrUndeclared, p.Key}
	}

	keyType, ok := t.lookupType(resolvedKey.ID)
	if !ok {
		return SymbolError{ErrUndeclaredType, p.Key}
	}

	valueType, ok := t.getExprType(env, p.Value)
	if !ok {
		return SymbolError{ErrUndeclaredType, p.Key}
	}

	if ok := t.matchType(keyType, valueType); !ok {
		return SymbolError{ErrTypeMismatch, p.Key}
	}
	return nil
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
		return t.lookupType(resolvedName.ID)
	case MemberAccess:
		panic(errors.ErrUnsupported)
	case Enum:
		constantType := t.typecheckEnumConstants(tt.Constants)
		return TypeEnum{ConstantType: constantType}, true
	case IFExpr:
		thenType, ok := t.getExprType(env, tt.Then)
		if !ok {
			return nil, false
		}

		return exprType, true
	default:
		panic(fmt.Sprintf("unexpected ast.Expr: %#v", node))
	}
}

func (t *typechecker) matchType(t1, t2 TypeSymbol) bool {
	return t1 == t2
}

// TODO rename to resolveElementTagType
func (t *typechecker) getElementTagType(compIdent, elementIdent Var) (TypeBinding, error) {
	resolvedComponent, ok := t.lookupName(compIdent.Name)
	if !ok {
		err := SymbolError{ErrUndeclared, compIdent}
		return TypeBinding{}, err
	}

	componentEnv, ok := t.lookupNameEnv(resolvedComponent.ID)
	if !ok {
		err := SymbolError{ErrNamespaceNotfound, compIdent}
		return TypeBinding{}, err
	}

	resolvedElement, ok := componentEnv.LookupName(elementIdent.Name)
	if !ok {
		err := SymbolError{ErrUndeclared, elementIdent}
		return TypeBinding{}, err
	}

	resolvedType, ok := t.lookupType(resolvedElement.ID)
	if !ok {
		err := SymbolError{ErrUndeclared, elementIdent}
		return TypeBinding{}, err
	}

	binding := TypeBinding{Type: resolvedType, Name: resolvedElement}
	return binding, nil
}