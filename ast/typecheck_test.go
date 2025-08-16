package ast

import (
	"testing"

	"github.com/tel-lang/tel"
)

func TestTypecheck(t *testing.T) {
	tests := []struct {
		flag      tel.Flag
		wantError bool
		name      string
		desc      string
		source    string
		skipTest  bool
	}{
		{
			name: "valid",
			desc: "",
			source: `
				(package p "views")

				(enum E [one, two])

				(component A
					[ s: String
					, n: Number
					, b: Bool
					, e: E
					]
					(s)
					(n)
					(if b
						"true"
						"false")
					(cond e
						E.one: "one"
						E.two: "two"))

				(component Main []
					(A [ s: "foo"
				   	   , n: 100
				   	   , b: true
				   	   , e: E.one
				   	   ]))
				`,
			flag: tel.FlagNoNativeElement,
		},
		{
			name: "invalid_package_tag",
			desc: "cannot use a value of type Package as an element tag",
			source: `
				(package pp "views")
				(document [] (pp))
				`,
			flag:      tel.FlagNoBuiltinType | tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "invalid_package_tag",
			desc: "cannot use a value of type Package as an element tag",
			source: `
				(package pp "views")
				(component A [] (pp))
				`,
			flag:      tel.FlagNoBuiltinType | tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "invalid_document_tag",
			desc: "cannot use a value of type Document as an element tag",
			source: `
				(package pp "views")
				(document A [])
				(component B [] (A))
				`,
			flag:      tel.FlagNoBuiltinType | tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "invalid_bool_tag_in_document",
			desc: "cannot use a value of type Document as an element tag",
			source: `
				(package pp "views")
				(document A [b: Bool] (b))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "invalid_bool_tag_in_component",
			desc: "cannot use a value of type Bool as an element tag",
			source: `
				(package pp "views")
				(component A [b: Bool] (b))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "invalid_enum_tag",
			desc: "cannot use a value of type Enum as an element tag",
			source: `
				(package pp "views")
				(enum E [one, two])
				(document A [e: E] (e))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "invalid_enum_tag",
			desc: "cannot use a value of type Enum as an element tag",
			source: `
				(package pp "views")
				(enum E [one, two])
				(component A [e: E] (e))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "invalid_property_type",
			desc: "a document X cannot have a parameter y of type X",
			source: `
				(package p "views")
				(document A [a: A])
				`,
			flag:      tel.FlagNoBuiltinType | tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "recursive_declaration",
			desc: "a component X cannot have a parameter y of type X",
			source: `
				(package p "views")
				(component A [a: A])
				`,
			flag:      tel.FlagNoBuiltinType | tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "document_property",
			desc: "a document cannot be used as the type of a property",
			source: `
				(package p "views")
				(document B [])
				(component A [a: B])
				`,
			flag:      tel.FlagNoBuiltinType | tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "duplicate_property",
			desc: "duplicate property in component",
			source: `(package p "views")
				   (component A [x: Number, x: String])`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "duplicate_property",
			desc: "duplicate property in document",
			source: `(package p "views")
				   (document [x: Number, x: String])`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(component B [p: String])
				(component A [] (B [p: 100]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(component B [p: String])
				(component A [] (B [p: true]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(component B [p: Number])
				(component A [] (B [p: true]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(component B [p: Number])
				(component A [] (B [p: "10"]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(component B [p: Bool])
				(component A [] (B [p: 100]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(component B [p: Bool])
				(component A [] (B [p: "true"]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(enum E [one])
				(component B [p: E])
				(component A [] (B [p: "1"]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(enum E [one])
				(component B [p: E])
				(component A [] (B [p: true]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(enum E [one])
				(component B [p: E])
				(component A [] (B [p: 1]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(enum E [one])
				(component B [p: E])
				(component A [] (B [p: false]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(enum E [one])
				(component B [p: E])
				(component A [] (B [p: 2]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(enum E [one])
				(component B [p: E])
				(component A [] (B [p: 2]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(enum E [one])
				(component B [p: E])
				(component A [] (B [p: 2]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "invalid parameter type",
			source: `
				(package p "views")
				(enum P [one])
				(component B [p: P])
				(component A [] (B [p: p.two]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "missing parameter",
			source: `
				(package p "views")
				(component B [s: String, n: Number, b: Bool])
				(component A [] (B [s: "2"]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "missing parameter",
			source: `
				(package p "views")
				(component B [s: String, n: Number, b: Bool])
				(component A [] (B [s: "2", n: 100]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "type_mismatch",
			desc: "missing parameter",
			source: `
				(package p "views")
				(component B [s: String, n: Number, b: Bool])
				(component A [] (B [s: "2", b: true]))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "invalid_component",
			desc: "undeclared parameter",
			source: `(package p "views")
				   (component A [name: String])
				   (component B []
				     (A [ name: "john"
				        , age: 100
				        ]))`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "undeclared_element",
			desc: "undeclared native element",
			source: `(package p "views")
				     ;; error: div is undefined
				     (component A [] (div))
				`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "undeclared_element",
			desc: "undeclared native element",
			source: `(package p "views")
			         ;; error: div is undefined
			         (document A [] (div))
			         `,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "undefined_type",
			desc: "no built in type in component",
			source: `(package p "views")
				   ;; error: String is undefined
				   (component H [s: String])`,
			flag:      tel.FlagNoBuiltinType,
			wantError: true,
		},
		{
			name: "number_literal_element_tag",
			desc: "cannot use a number as an element tag",
			source: `(package p "views")
				   (component H [] (10))`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "string_literal_element_tag",
			desc: "cannot use a number as an element tag",
			source: `(package p "views")
				   (component H [] ("str"))`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "if_stmt",
			source: `(package p "views")
				     (component A []
				     	(if true (div)))`,
		},
		{
			name: "if_stmt_else",
			source: `(package p "views")
				     (component A []
				     	(if true
				     		(div)
				     		(span)))`,
		},
		{
			name: "if_stmt__property_cond",
			source: `(package p "views")
				     (component A [b: Bool]
				     	(if b
				     		(div)
				     		(span)))`,
		},
		{
			name: "if_stmt__invalid_property_cond",
			source: `(package p "views")
				     (component A []
				     	(if bb
				     		(div)
				     		(span)))`,
			wantError: true,
		},
		{
			name: "if_stmt__string_property_cond",
			source: `(package p "views")
				     (component A [b: String]
				     	(if b
				     		(div)
				     		(span)))`,
			wantError: true,
		},
		{
			name: "if_stmt__number_property_cond",
			source: `(package p "views")
				     (component A [b: Number]
				     	(if b
				     		(div)
				     		(span)))`,
			wantError: true,
		},
		{
			name: "if_stmt__enum_property_cond",
			source: `(package p "views")
					 (enum YesNo [Yes, No])
				     (component A [b: YesNo]
				     	(if b
				     		(div)
				     		(span)))`,
			wantError: true,
		},
		{
			name: "cond_stmt__bool_target",
			source: `(package p "views")
				     (component A [b: Bool]
				     	(cond b
				     		true: (div)
				     		false: (span)))`,
			wantError: true,
		},
		{
			name: "cond_stmt__invalid_target",
			source: `(package p "views")
				     (component A []
				     	(cond bb
				     		true: (div)
				     		false: (span)))`,
			wantError: true,
		},
		{
			name: "cond_stmt__no_case",
			source: `(package p "views")
					 (enum YesNo [Yes, No])
				     (component A [b: YesNo]
				     	(cond b))`,
			wantError: true,
		},
		{
			name: "cond_stmt__invalid_case_cond",
			source: `(package p "views")
					 (enum YesNo [Yes, No])
				     (component A [b: YesNo]
				     	(cond b
				     		Yes: (div)
				     		true: (span)))`,
			wantError: true,
		},
		{
			name: "cond_stmt__mismatch_case_cond",
			source: `(package p "views")
					 (enum YesNo [Yes, No])
				     (component A [b: YesNo]
				     	(cond b
				     		YesNo.Yes: (div)
				     		true: (span)))`,
			wantError: true,
		},
		{
			name: "cond_stmt",
			source: `(package p "views")
					 (enum YesNo [Yes, No])
				     (component A [b: YesNo]
				     	(cond b
				     		YesNo.Yes: (div)
				     		YesNo.No: (span)))`,
		},
		{
			name: "cond_stmt__implicit_selector",
			source: `(package p "views")
					 (enum YesNo [Yes, No])
				     (component A [b: YesNo]
				     	(cond b
				     		.Yes: (div)
				     		.No: (span)))`,
			skipTest: true,
		},
		{
			name: "cond_stmt__optional_semicolon",
			source: `(package p "views")
					 (enum YesNo [Yes, No])
				     (component A [b: YesNo]
				     	(cond b
				     		.Yes (div)
				     		.No (span)))`,
			skipTest: true,
		},
		{
			name: "number_element__reject_parameter",
			source: `(package pp "views")
				     (component A [c: Number]
				     	(c [k: "v"]))`,
			wantError: true,
		},
		{
			name: "number_element__reject_children",
			source: `(package pp "views")
				     (component A [c: Number]
				     	(c (div)))`,
			wantError: true,
		},
		{
			name: "string_element__reject_parameter",
			source: `(package pp "views")
				     (component A [c: String]
				     	(c [k: "v"]))`,
			wantError: true,
		},
		{
			name: "string_element__reject_children",
			source: `(package pp "views")
				     (component A [c: String]
				     	(c (div)))`,
			wantError: true,
		},
		{
			name: "html_element__reject_parameter",
			source: `(package pp "views")
				     (component A []
				     	(div [k: "v"]))`,
			wantError: true,
		},
		{
			name: "property_element__reject_parameter",
			source: `(package pp "views")
					 (component B [text: String])
				     (component A [b: B]
				     	(b [text: "str"]))`,
			wantError: true,
		},
		{
			name: "component_element__reject_children",
			source: `(package pp "views")
					 (component B [text: String])
				     (component A []
				     	(B [text: "str"]
				     		(div)))`,
			wantError: true,
		},
	}

	for _, test := range tests {
		name := test.name
		t.Run(name, func(t *testing.T) {
			if test.skipTest {
				t.Skip()
			}
			runTypechecker(t, test.name, test.source, test.wantError, test.flag)
		})
	}
}

func runTypechecker(t *testing.T, filename, content string, succeedOnError bool, flag tel.Flag) {
	t.Helper()

	files := tel.NewFileSet("test")
	files.AddString(filename, content)
	ctx := tel.NewContextWithFlags(tel.PreserveComment | flag)

	pkg := Package{
		Modules:  []*Module{},
		Name:     filename,
		Path:     "test/" + filename,
		FullName: filename,
		hasError: false,
	}

	LoadPackage(ctx, files, &pkg)

	if succeedOnError {
		if !pkg.HasError() {
			t.Fatal("transpiler succeeded unexpectedly")
		}
	} else {
		if pkg.HasError() {
			t.Fatal("transpiler failed unexpectedly")
		}
	}
}
