package transpiler

import (
	"fmt"

	"github.com/tel-lang/tel/ast"
)

func (b Bool) String() string {
	switch ast.Bool(b) {
	case ast.False:
		return "false"
	case ast.True:
		return "true"
	default:
		panic(fmt.Sprintf("unexpected ast.Bool: %#v", b))
	}
}
