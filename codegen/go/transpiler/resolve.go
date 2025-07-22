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
		return string(t)
	case Var:
		return string(t)
	case Bool:
		return string(t)
	default:
		panic(fmt.Sprintf("unexpected expression type: %v", reflect.TypeOf(expr)))
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
		return fmt.Sprintf("%s", t)
	case Number:
		return fmt.Sprintf("%s", t)
	case Var:
		return fmt.Sprintf("%s", t)
	case Bool:
		return fmt.Sprintf("%s", t)
	default:
		panic(fmt.Sprintf("unexpected expression type: %v", reflect.TypeOf(expr)))

	}
}
