package ast

import (
	_ "embed"
	"reflect"
	"testing"

	"github.com/eml-lang/teml/token"

	past "github.com/eml-lang/teml/ast"
)

//go:embed testdata/transform/string_element.teml
var stringElement []byte

//go:embed testdata/transform/number_element.teml
var numberElement []byte

//go:embed testdata/transform/native_element.teml
var nativeElement []byte

//go:embed testdata/transform/instance_element.teml
var instanceElement []byte

//go:embed testdata/transform/component_element.teml
var componentElement []byte

func TestTransformToStringElement(t *testing.T) {
	stmt := parseSource(t, stringElement, 1, 0)
	var element StringElement
	var ok bool

	if element, ok = stmt.Element.(StringElement); !ok {
		t.Fatalf("expected %v got %v", reflect.TypeOf(element), reflect.TypeOf(stmt.Element))
	}
}

func TestTransformToNumberElement(t *testing.T) {
	stmt := parseSource(t, numberElement, 1, 0)
	var element NumberElement
	var ok bool

	if element, ok = stmt.Element.(NumberElement); !ok {
		t.Fatalf("expected %v got %v", reflect.TypeOf(element), reflect.TypeOf(stmt.Element))
	}
}

func TestTransformToNativeElement(t *testing.T) {
	stmt := parseSource(t, nativeElement, 1, 0)
	var element NativeElement
	var ok bool

	if element, ok = stmt.Element.(NativeElement); !ok {
		t.Fatalf("expected %v got %v", reflect.TypeOf(element), reflect.TypeOf(stmt.Element))
	}
}

func TestTransformToComponentElement(t *testing.T) {
	stmt := parseSource(t, componentElement, 2, 1)
	var element ComponentElement
	var ok bool

	if element, ok = stmt.Element.(ComponentElement); !ok {
		t.Fatalf("expected %v got %v", reflect.TypeOf(element), reflect.TypeOf(stmt.Element))
	}
}

func TestTransformToInstanceElement(t *testing.T) {
	stmt := parseSource(t, instanceElement, 2, 1)
	var element InstanceElement
	var ok bool

	if element, ok = stmt.Element.(InstanceElement); !ok {
		t.Fatalf("expected %v got %v", reflect.TypeOf(element), reflect.TypeOf(stmt.Element))
	}
}

func parseSource(t *testing.T, source []byte, decls int, targetDecl int) Stmt {
	t.Helper()

	toks := token.Scan(source, "test.teml")
	astp := past.ParseFile(toks)
	if astp.HasError() {
		t.Fatal("source parser failed unexpected")
	}

	astt := ParseFile(astp, toks)
	if astt.HasError() {
		t.Fatal("source transpiler failed unexpected")
	}

	renv := ResolveFile(astt)
	if astt.HasError() {
		t.Fatal("source transpiler failed unexpected")
	}

	_ = TypecheckFile(astt, renv)

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
