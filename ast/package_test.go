package ast

import (
	"testing"

	"github.com/tel-lang/tel"
)

func TestPackage(t *testing.T) {
	type Module struct {
		dir    string
		name   string
		source string
	}
	type Test struct {
		modules   []Module
		name      string
		desc      string
		flag      tel.Flag
		wantError bool
		skip      bool
	}
	tests := []Test{
		{
			name: "load multiple modules",
			modules: []Module{
				{
					dir:  "demo",
					name: "foo",
					source: `
						(package foo "demo")
						(component Foo [])
					`,
				},
				{
					dir:  "demo",
					name: "bar",
					source: `
						(package bar "demo")
						(component Bar [])
					`,
				},
				{
					dir:  "demo",
					name: "main",
					source: `
						(package pp "demo")
						(component Main
							[ foo: pp.foo.Foo
							, bar: pp.bar.Bar
							])
					`,
				},
			},
		},
		{
			name: "module property type",
			desc: "cannot use a value of type Module as a property type",
			modules: []Module{
				{
					dir:  "demo",
					name: "foo",
					source: `
						(package foo "demo")
						(component Foo [])
					`,
				},
				{
					dir:  "demo",
					name: "main",
					source: `
						(package pp "demo")
						(component Main [foo: pp.foo])
					`,
				},
			},
			wantError: true,
		},
		{
			name: "module element tag",
			desc: "cannot use a value of type Module as an element tag",
			modules: []Module{
				{
					dir:  "demo",
					name: "foo",
					source: `
						(package foo "demo")
						(component Foo [])
					`,
				},
				{
					dir:  "demo",
					name: "main",
					source: `
						(package pp "demo")
						(component Main [] (pp.foo))
					`,
				},
			},
			wantError: true,
		},
		{
			name: "using single",
			modules: []Module{
				{
					dir:  "demo",
					name: "model",
					source: `
						(package foo "demo")
						(component Foo [])
					`,
				},
				{
					dir:  "demo",
					name: "main",
					source: `
						(package pp "demo")
						(using Foo pp.model)
						(component Main [f: Foo])
					`,
				},
			},
		},
		{
			name: "using single",
			modules: []Module{
				{
					dir:  "demo",
					name: "model",
					source: `
						(package foo "demo")
						(component Foo [])
					`,
				},
				{
					dir:  "demo",
					name: "main",
					source: `
						(package pp "demo")
						(using [Foo] pp.model)
						(component Main [f: Foo])
					`,
				},
			},
		},
		{
			name: "using many",
			modules: []Module{
				{
					dir:  "demo",
					name: "model",
					source: `
						(package foo "demo")
						(component Foo [])
						(component Bar [])
					`,
				},
				{
					dir:  "demo",
					name: "main",
					source: `
						(package pp "demo")
						(using [Foo, Bar] pp.model)
						(component Main [f: Foo, b: Bar])
					`,
				},
			},
		},
		{
			name: "using many undeclared",
			modules: []Module{
				{
					dir:  "demo",
					name: "model",
					source: `
						(package foo "demo")
						(component Foo [])
					`,
				},
				{
					dir:  "demo",
					name: "main",
					source: `
						(package pp "demo")
						(using [Foo, Bar] pp.model)
						(component Main [f: Foo])
					`,
				},
			},
			wantError: true,
		},
		{
			name: "different package names in a directory",
			modules: []Module{
				{
					dir:  "demo",
					name: "model",
					source: `
						(package pp "foo")
						(component Foo [])
					`,
				},
				{
					dir:  "demo",
					name: "main",
					source: `
						(package pp "bar")
						(component Bar [])
					`,
				},
			},
			wantError: true,
		},
	}

	for _, test := range tests {
		ctx := tel.NewContextWithFlags(test.flag)
		src := tel.FileSet{
			Path:  "test/pkg",
			Files: []tel.File{},
		}
		for _, mod := range test.modules {
			file := tel.NewFile(mod.dir, mod.name, []byte(mod.source))
			src.Files = append(src.Files, file)
		}

		pkg := &Package{
			Name:     "test",
			Path:     "test/pkg",
			FullName: "test/pkg",
			hasError: false,
		}

		t.Run(test.name, func(t *testing.T) {
			if test.skip {
				t.Skip()
			}

			LoadPackage(ctx, src, pkg)

			if test.wantError {
				if !pkg.hasError {
					t.Error("load package succeeded unexpectedly")
				}
			} else {
				if pkg.hasError {
					t.Error("load package failed unexpectedly")
				}
			}
		})
	}
}
