package ast

import (
	"fmt"
	"reflect"
)

func ResolveValue(expr Expr) string {
	switch t := expr.(type) {
	case String:
		return string(t)
	case Number:
		return fmt.Sprintf("%d", t)
	case Var:
		return string(t)
	case Bool:
		return t.String()
	default:
		panic(fmt.Sprintf("unexpected expression type: %v", reflect.TypeOf(expr)))
	}
}

func ResolveEnumContantAsName(expr Expr) string {
	switch t := expr.(type) {
	case String:
		// strip double quoted
		name := t[1 : len(t)-1]
		return string(name)
	default:
		return ResolveValue(expr)
	}
}

func ResolveType(t Type) string {
	switch tt := t.(type) {
	case Var:
		return string(tt)
	case Enum:
		return string(tt.TypeName)
	default:
		panic(fmt.Sprintf("unexpected type: %v", reflect.TypeOf(t)))
	}
}

func ResolveMapKey(expr Expr) string {
	return ResolveValue(expr)
}

func ResolveSwitchTarget(expr Expr) string {
	switch t := expr.(type) {
	case String:
		return string(t)
	case Number:
		return fmt.Sprintf("%d", t)
	case Var:
		return string(t)
	case Bool:
		return t.String()
	default:
		panic(fmt.Sprintf("unexpected expression type: %v", reflect.TypeOf(expr)))
	}
}
