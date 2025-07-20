package generator

import (
	"fmt"
	"reflect"

	ast "github.com/eml-lang/teml/codegen/go/transpiler"
)

func (g *generator) resolveValue(expr ast.Expr) string {
	switch t := expr.(type) {
	case ast.String:
		return fmt.Sprintf("%q", t)
	case ast.Number:
		return fmt.Sprintf("%s", t)
	case ast.Var:
		return fmt.Sprintf("%s", t)
	case ast.Bool:
		return fmt.Sprintf("%s", t)
	default:
		panic(fmt.Sprintf("unexpected expression type: %v", reflect.TypeOf(expr)))

	}
}

func (g *generator) resolveSwitchTarget(expr ast.Expr) string {
	switch t := expr.(type) {
	case ast.String:
		return fmt.Sprintf("%s", t)
	case ast.Number:
		return fmt.Sprintf("%s", t)
	case ast.Var:
		return fmt.Sprintf("%s", t)
	case ast.Bool:
		return fmt.Sprintf("%s", t)
	default:
		panic(fmt.Sprintf("unexpected expression type: %v", reflect.TypeOf(expr)))

	}
}
