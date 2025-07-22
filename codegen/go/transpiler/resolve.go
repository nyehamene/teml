package ast

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	ast "github.com/eml-lang/teml/transpiler"
)

func (p *parser) resolveType(node ast.Var) Type {
	switch node.Name {
	case "String":
		return String("string")
	case "Number":
		return Number("int")
	case "Bool":
		return Bool("bool")
	}
	return Var(node.Name)
}

func (p *parser) resolveFieldType(structname, field string, expr ast.PropertyType) Type {
	switch t := expr.(type) {
	case ast.Var:
		return p.resolveType(t)

	case ast.Enum:
		typename := structname + strings.ToUpper(field[0:1]) + field[1:]

		constants := []Expr{}
		for _, c := range t.Constants {
			constant := p.resolveValue(c)
			constants = append(constants, constant)
		}

		e := Enum{
			TypeName:  typename,
			Constants: []Expr{},
		}
		return e

	case ast.MemberAccess:
		panic(errors.ErrUnsupported)

	default:
		panic(fmt.Sprintf("unexpected propety type: %v", reflect.TypeOf(expr)))
	}
}

func (p *parser) resolveName(expr ast.Expr) string {
	switch t := expr.(type) {
	case ast.Var:
		return p.resolveType(t).Name()
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

func (p *parser) resolveValue(expr ast.Expr) Expr {
	switch t := expr.(type) {
	case ast.Var:
		return Var(t.Name)
	case ast.String:
		return String(t.Value())
	case ast.Number:
		return Number(t)
	case ast.Bool:
		return Bool(t)
	case ast.MemberAccess:
		panic(errors.ErrUnsupported)
	case ast.IFExpr:
		panic(errors.ErrUnsupported)
	case ast.CondExpr:
		panic(errors.ErrUnsupported)
	case ast.Enum:
		panic(errors.ErrUnsupported)
	default:
		panic(errors.ErrUnsupported)
	}
}

func (p *parser) resolveAttributeValue(expr ast.Expr) Expr {
	switch t := expr.(type) {
	case ast.Var:
		return Var(t.Name)
	case ast.String:
		return String(t)
	case ast.Number:
		return Number(t)
	case ast.Bool:
		return Bool(t)
	case ast.MemberAccess:
		panic(errors.ErrUnsupported)
	case ast.IFExpr:
		panic(errors.ErrUnsupported)
	case ast.CondExpr:
		panic(errors.ErrUnsupported)
	case ast.Enum:
		panic(errors.ErrUnsupported)
	default:
		panic(errors.ErrUnsupported)
	}
}

func (p *parser) resolveIfCond(m methodinfo, expr ast.Expr) string {
	switch t := expr.(type) {
	case ast.Var:
		return m.receiver + "." + t.Name
	case ast.Bool:
		return string(t)
	case ast.String:
		panic(errors.ErrUnsupported)
	case ast.Number:
		panic(errors.ErrUnsupported)
	case ast.MemberAccess:
		panic(errors.ErrUnsupported)
	case ast.IFExpr:
		panic(errors.ErrUnsupported)
	case ast.CondExpr:
		panic(errors.ErrUnsupported)
	case ast.Enum:
		panic(errors.ErrUnsupported)
	default:
		panic(errors.ErrUnsupported)
	}
}

func (p *parser) resolveCondTarget(m methodinfo, expr ast.Expr) Expr {
	switch t := expr.(type) {
	case ast.Var:
		return Var(m.receiver + "." + t.Name)
	case ast.Bool:
		return Bool(t)
	case ast.String:
		return String(t)
	case ast.Number:
		return Number(t)
	case ast.MemberAccess:
		panic(errors.ErrUnsupported)
	case ast.IFExpr:
		panic(errors.ErrUnsupported)
	case ast.CondExpr:
		panic(errors.ErrUnsupported)
	case ast.Enum:
		panic(errors.ErrUnsupported)
	default:
		panic(errors.ErrUnsupported)
	}
}

func (p *parser) resolveCondCase(expr ast.Expr) Expr {
	switch t := expr.(type) {
	case ast.Bool:
		return Bool(t)
	case ast.String:
		return String(t)
	case ast.Number:
		return Number(t)
	case ast.Var:
		panic(errors.ErrUnsupported)
	case ast.MemberAccess:
		panic(errors.ErrUnsupported)
	case ast.IFExpr:
		panic(errors.ErrUnsupported)
	case ast.CondExpr:
		panic(errors.ErrUnsupported)
	case ast.Enum:
		panic(errors.ErrUnsupported)
	default:
		panic(errors.ErrUnsupported)
	}
}
