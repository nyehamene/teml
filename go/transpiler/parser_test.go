package transpiler_test

import (
	"testing"

	"github.com/tel-lang/tel"
	"github.com/tel-lang/tel/go/transpiler"
)

func TestGenerateNamespace(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		wantError bool
		flag      tel.Flag
	}{
		{
			name: "empty parameter and body",
			src: `(package pp "mypackage")
				  (component EmptyParameterAndBody [])`,
			flag: tel.FlagNoBuiltinType & tel.FlagNoNativeElement,
		},
		{
			name: "enum parameter empty body",
			src: `(package pp "mypackage")
				  (enum YesNo [Yes, No])
				  (component EnumParameterEmptyBody [e: YesNo])`,
			flag: tel.FlagNoNativeElement,
		},
		{
			name: "empty body",
			src: `(package pp "mypackage")
				  (component EmptyBody [n: Number, s: String, b: Bool])`,
			flag: tel.FlagNoNativeElement,
		},
		{
			name: "many enum parameters empty body",
			src: `(package pp "mypackage")
				  (enum YesNo [Yes, No])
				  (enum RGB [Red, Green, Blue])
				  (component ManyEnumParametersEmptyBody [choice: YesNo, rgb: RGB])`,
			flag: tel.FlagNoNativeElement,
		},
		{
			name: "empty parameter native element",
			src: `(package pp "mypackage")
				  (component EmptyParameterNativeElement []
				     (div "hello world"))`,
			flag: tel.FlagNoBuiltinType,
		},
		{
			name: "empty parameter native element",
			src: `(package pp "mypackage")
				  (component EmptyParameterNativeElement []
					(div "hello world"))`,
			flag:      tel.FlagNoNativeElement,
			wantError: true,
		},
		{
			name: "empty parameter string element",
			src: `(package pp "mypackage")
				  (component A [])
				  (component EmptyParameterComponentElement [a: A]
					"hello, world")`,
			flag: tel.FlagNoBuiltinType & tel.FlagNoNativeElement,
		},
		{
			name: "empty parameter text element",
			src: `(package pp "mypackage")
				  (component A [])
				  (component EmptyParameterComponentElement [a: A]
					-- hello, world
					-- how are you
					)`,
			flag: tel.FlagNoBuiltinType & tel.FlagNoNativeElement,
		},
		{
			name: "empty parameter component element",
			src: `(package pp "mypackage")
				  (component A [])
				  (component EmptyParameterComponentElement []
					(A))`,
			flag: tel.FlagNoBuiltinType & tel.FlagNoNativeElement,
		},
		{
			name: "bool parameter property element",
			src: `(package pp "mypackage")
				  (component EmptyParameterComponentElement [b: Bool]
					(if b "string" (div)))`,
			flag: tel.FlagNoBuiltinType & tel.FlagNoNativeElement,
		},
		{
			name: "enum parameter property element",
			src: `(package pp "mypackage")
				  (component C [])
				  (enum Type [String, Text, Native, Component])
				  (component EmptyParameterComponentElement [c: C, e: Type]
					(cond e
						"string": "one"
						"text":   -- two
						"native": (div)
						"component": (c)))`,
			flag: tel.FlagNoBuiltinType & tel.FlagNoNativeElement,
		},
		{
			name: "component parameter property element",
			src: `(package pp "mypackage")
				  (component A [])
				  (component EmptyParameterComponentElement [a: A]
					(a))`,
			flag: tel.FlagNoBuiltinType & tel.FlagNoNativeElement,
		},
		{
			name: "string parameter property element",
			src: `(package pp "mypackage")
				  (component EmptyParameterComponentElement [s: String]
					(s))`,
			flag: tel.FlagNoBuiltinType & tel.FlagNoNativeElement,
		},
		{
			name: "number parameter property element",
			src: `(package pp "mypackage")
				  (component EmptyParameterComponentElement [n: Number]
					(n))`,
			flag: tel.FlagNoBuiltinType & tel.FlagNoNativeElement,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			src := parseString(test.name, test.src, test.flag)
			checkOutput(t, src, test.wantError)
		})
	}
}

func TestGeneratePackage(t *testing.T) {
	type sourcefile struct {
		name    string
		content string
	}

	type testcase struct {
		flag    tel.Flag
		path    string
		sources []sourcefile
	}

	tests := testcase{
		flag: 0,
		path: "test",
		sources: []sourcefile{
			{
				name: "view.tel",
				content: `(package pp "demo")
						  (component Foo [])
						  (component Bar [])`,
			},
			{
				name: "main.tel",
				content: `(package pp "demo")
						  (component Main
						  	[ foo: pp.view.Foo
						    , bar: pp.view.Bar
						    ]
						  	(foo)
						  	(bar))`,
			},
		},
	}

	files := tel.NewFileSet(tests.path)
	for _, src := range tests.sources {
		files.AddString(src.name, src.content)
	}

	const wantError = false
	parsedFiles := parsePackage(files, tests.flag)
	checkOutput(t, parsedFiles, wantError)
}

func parsePackage(files tel.FileSet, flag tel.Flag) transpiler.FileSet {
	ctx := tel.NewContextWithFlags(flag)
	return transpiler.ParseFile(ctx, files)
}

func checkOutput(t *testing.T, src transpiler.FileSet, wantError bool) {
	t.Helper()

	if wantError && !src.Error {
		t.Fatal("parser succeeded unexpectedly")
	} else {
		t.Error("parser failed unexpectedly")
	}
}

func parseString(name, content string, flag tel.Flag) transpiler.FileSet {
	file := tel.File{
		Path:    "test/" + name + ".tel",
		Name:    name,
		Content: []byte(content),
	}
	pkg := tel.FileSet{
		Path:  "test",
		Files: []tel.File{file},
	}
	return parsePackage(pkg, flag)
}
