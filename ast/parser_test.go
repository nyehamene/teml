package ast_test

import (
	"fmt"
	"testing"

	"github.com/eml-lang/teml/ast"
	"github.com/eml-lang/teml/token"
	"github.com/google/go-cmp/cmp"
)

func TestParse_package(t *testing.T) {
	source := `(package p "path/to/pkg")`
	expected := []token.Kind{
		token.Ident,
		token.String,
	}

	tokens := token.Scan([]byte(source))

	f, hasError := ast.ParseWithErrorHandler(*tokens, func(err string) {
		t.Error(err)
	})

	kinds := getKinds(f.Pkg)

	if hasError {
		t.Error("Parser failed unexpectedly")
	}

	if diff := cmp.Diff(expected, kinds); diff != "" {
		t.Error(diff)
	}

}

func TestParse_package_error(t *testing.T) {
	source := []string{
		`(package)`,
		`(package p)`,
		`(package "path/to/package")`,
	}

	for i, src := range source {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			tokens := token.Scan([]byte(src))
			_, hasError := ast.ParseWithErrorHandler(*tokens, func(string) {})

			if !hasError {
				t.Error("Parser succeeded unexpectedly")
			}
		})
	}
}

func getKinds(n ast.Node) []token.Kind {
	switch t := n.(type) {
	case ast.Package:
		k := []token.Kind{t.Ident.Kind, t.Path.Kind}
		return k
	}
	return nil
}
