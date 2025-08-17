package ast

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/eml-lang/teml/ast"
	"github.com/eml-lang/teml/internal/flags"
	"github.com/eml-lang/teml/internal/source"
	"github.com/eml-lang/teml/token"

	past "github.com/eml-lang/teml/ast"
)

type testtable struct {
	path                string
	declCount           int
	targetDeclIndex     int
	targetStmtIndex     int
	expectedElementType transpiler.Element
	flag                flags.Flag
}

func TestTransform(t *testing.T) {
	tests := []testtable{
		{
			path:                "./testdata/transform/string_element.teml",
			declCount:           1,
			targetDeclIndex:     0,
			targetStmtIndex:     0,
			expectedElementType: transpiler.StringElement{},
			flag:                flags.FlagNoNativeElement,
		},
		{
			path:                "./testdata/transform/number_element.teml",
			declCount:           1,
			targetDeclIndex:     0,
			targetStmtIndex:     0,
			expectedElementType: transpiler.NumberElement{},
			flag:                flags.FlagNoNativeElement,
		},
		{
			path:                "./testdata/transform/native_element.teml",
			declCount:           1,
			targetDeclIndex:     0,
			targetStmtIndex:     0,
			expectedElementType: transpiler.NativeElement{},
		},
		{
			path:                "./testdata/transform/property_element.teml",
			declCount:           2,
			targetDeclIndex:     1,
			targetStmtIndex:     0,
			expectedElementType: transpiler.PropertyElement{},
			flag:                flags.FlagNoNativeElement,
		},
		{
			path:                "./testdata/transform/component_element.teml",
			declCount:           2,
			targetDeclIndex:     1,
			targetStmtIndex:     0,
			expectedElementType: transpiler.ComponentElement{},
			flag:                flags.FlagNoNativeElement,
		},
	}

//go:embed testdata/transform/native_element.teml
var nativeElement []byte

//go:embed testdata/transform/instance_element.teml
var instanceElement []byte

//go:embed testdata/transform/component_element.teml
var componentElement []byte

func TestTransformToStringElement(t *testing.T) {
	stmt := parseSource(t, stringElement, 1, 0, FlagNoNativeElement)
	var element StringElement
	var ok bool

	if element, ok = stmt.Element.(StringElement); !ok {
		t.Fatalf("expected %v got %v", reflect.TypeOf(element), reflect.TypeOf(stmt.Element))
	}
}

func TestTransformToNumberElement(t *testing.T) {
	stmt := parseSource(t, numberElement, 1, 0, FlagNoNativeElement)
	var element NumberElement
	var ok bool

	if element, ok = stmt.Element.(NumberElement); !ok {
		t.Fatalf("expected %v got %v", reflect.TypeOf(element), reflect.TypeOf(stmt.Element))
	}
}

func TestTransformToNativeElement(t *testing.T) {
	stmt := parseSource(t, nativeElement, 1, 0, 0)
	var element NativeElement
	var ok bool

	if element, ok = stmt.Element.(NativeElement); !ok {
		t.Fatalf("expected %v got %v", reflect.TypeOf(element), reflect.TypeOf(stmt.Element))
	}
}

// TODO rename to TestTransformToPropertyElement
func TestTransformToComponentElement(t *testing.T) {
	stmt := parseSource(t, componentElement, 2, 1, FlagNoNativeElement)
	var element PropertyElement
	var ok bool

	if element, ok = stmt.Element.(PropertyElement); !ok {
		t.Fatalf("expected %v got %v", reflect.TypeOf(element), reflect.TypeOf(stmt.Element))
	}
}

// TODO rename to TestTransformToComponentElement
func TestTransformToInstanceElement(t *testing.T) {
	stmt := parseSource(t, instanceElement, 2, 1, FlagNoNativeElement)
	var element ComponentElement
	var ok bool

	if element, ok = stmt.Element.(ComponentElement); !ok {
		t.Fatalf("expected %v got %v", reflect.TypeOf(element), reflect.TypeOf(stmt.Element))
	}
}

func parseSource(t *testing.T, src []byte, decls int, targetDecl int, flag Flag) Stmt {
	t.Helper()

	file := source.File{
		Path:    "test.teml",
		Name:    "test",
		Content: src,
	}

	templ := decls[test.targetDeclIndex]
	if templ.Kind != ast.ComponentTemplate {
		t.Errorf("expected %s", ast.ComponentTemplate)
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
	gotElemnetType := reflect.TypeOf(stmt.Element)
	if gotElemnetType != expectedElementType {
		t.Errorf("expected %v", expectedElementType)
		t.Errorf("got %v", gotElemnetType)
	}
	if astt.HasError() {
		t.Fatal("ast resolver failed unexpected")
	}

	_ = TypecheckFile(astt, renv, flag)
	for _, err := range astt.Errors() {
		t.Error(err)
	}
	if astt.HasError() {
		t.Fatal("ast resolver failed unexpected")
	}

	if l := len(astt.Declarations); l != decls {
		t.Fatalf("expected %d declarations got %d", decls, l)
	}

	decl := astt.Declarations[targetDecl]
	var com Component
	var ok bool

	if com, ok = decl.(Component); !ok {
		t.Fatalf("expected %v got %v", reflect.TypeOf(com), reflect.TypeOf(decl))
	}

	if l := len(com.Stmts); l != 1 {
		t.Fatalf("expected 1 stmt got %d", l)
	}

	stmt := com.Stmts[0]
	return stmt
}
