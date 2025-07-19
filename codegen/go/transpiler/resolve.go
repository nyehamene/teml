package ast

import (
	"errors"
	"fmt"
	"reflect"

	ast "github.com/eml-lang/teml/transpiler"
)

func (p *parser) resolveType(expr ast.PropertyType) string {
	switch t := expr.(type) {
	case ast.Var:
		switch t.Name {
		case "String":
			return "string"
		case "Number":
			return "int"
		}
		return t.Name

	case ast.MemberAccess:
		panic(errors.ErrUnsupported)

	case ast.Enum:
		panic(errors.ErrUnsupported)

	default:
		panic(fmt.Sprintf("unexpected propety type: %v", reflect.TypeOf(expr)))
	}
}

func (p *parser) resolveName(expr ast.Expr) string {
	switch t := expr.(type) {
	case ast.Var:
		return t.Name
	case ast.MemberAccess:
		panic(errors.ErrUnsupported)
	case ast.String:
		panic(fmt.Sprintf("expected to match a name but got string %v", t))
	case ast.Number:
		panic(fmt.Sprintf("expected to match a name but got number %v", t))
	case ast.Bool:
		panic(fmt.Sprintf("expected to match a name but got bool %v", t))
	case ast.IFExpr:
		panic(fmt.Sprintf("expected to match a name but got if expression %v", t))
	case ast.CondExpr:
		panic(fmt.Sprintf("expected to match a name but got cond expression %v", t))
	case ast.Enum:
		panic(fmt.Sprintf("expected to match a name but got enum expression %v", t))
	default:
		panic(fmt.Sprintf("expected expression type: %v", reflect.TypeOf(expr)))
	}
}
