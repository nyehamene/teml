package ast_test

import (
	_ "embed"
	"path"
	"reflect"
	"testing"

	"github.com/tel-lang/tel"
	"github.com/tel-lang/tel/ast"
)

type testtable struct {
	name                string
	source              string
	declCount           int
	targetDeclIndex     int
	targetStmtIndex     int
	expectedElementType ast.Stmt
	flag                tel.Flag
}

func TestTransform(t *testing.T) {
	tests := []testtable{
		{
			name:                "native_element",
			source:              `(package p "views") (component C [] (div))`,
			declCount:           1,
			targetDeclIndex:     0,
			targetStmtIndex:     0,
			expectedElementType: ast.HTMLElement{},
			flag:                tel.FlagNoBuiltinType,
		},
		{
			name:                "string_element",
			source:              `(package p "views") (component C [text: String] (text))`,
			declCount:           1,
			targetDeclIndex:     0,
			targetStmtIndex:     0,
			expectedElementType: ast.StringElement{},
			flag:                tel.FlagNoNativeElement,
		},
		{
			name:                "number_element",
			source:              `(package p "views") (component C [num: Number] (num))`,
			declCount:           1,
			targetDeclIndex:     0,
			targetStmtIndex:     0,
			expectedElementType: ast.NumberElement{},
			flag:                tel.FlagNoNativeElement,
		},
		{
			name:                "property_element",
			source:              `(package p "views") (component A []) (component C [p: A] (p))`,
			declCount:           2,
			targetDeclIndex:     1,
			targetStmtIndex:     0,
			expectedElementType: ast.PropertyElement{},
			flag:                tel.FlagNoNativeElement,
		},
		{
			name:                "component_element",
			source:              `(package p "views") (component A []) (component C [] (A))`,
			declCount:           2,
			targetDeclIndex:     1,
			targetStmtIndex:     0,
			expectedElementType: ast.ComponentElement{},
			flag:                tel.FlagNoNativeElement,
		},
	}

	for _, test := range tests {
		name := path.Base(test.name)

		pkg := ast.Package{
			Modules:  []*ast.Module{},
			Name:     name,
			Path:     "test/" + name,
			FullName: name,
		}

		t.Run(name, func(t *testing.T) {
			files := tel.NewFileSet(name)
			files.AddString(test.name, test.source)
			ctx := tel.NewContext()
			ast.LoadPackage(ctx, files, &pkg)
			mod := pkg.Modules[0]
			validateTypeCheckResult(t, mod, test)
		})
	}
}

func validateTypeCheckResult(t *testing.T, mod *ast.Module, test testtable) {
	t.Helper()

	decls := mod.Templates
	if l := len(decls); l != test.declCount {
		t.Errorf("expected %d declarations", test.declCount)
		t.Errorf("got %d declarations", l)
		return
	}

	templ := decls[test.targetDeclIndex]
	if templ.Kind != ast.TemplateComponent {
		t.Errorf("expected %s", ast.TemplateComponent)
		t.Errorf("got %s", templ.Kind)
		return
	}

	if l := len(templ.Stmts); l < test.targetStmtIndex {
		t.Errorf("expected %d at least statements", test.targetStmtIndex)
		t.Errorf("got %d statements", l)
		return
	}

	stmt := templ.Stmts[test.targetStmtIndex]
	expectedElementType := reflect.TypeOf(test.expectedElementType)
	gotElemnetType := reflect.TypeOf(stmt)
	if gotElemnetType != expectedElementType {
		t.Errorf("expected %v", expectedElementType)
		t.Errorf("got %v", gotElemnetType)
	}
}
