package ast

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/tel-lang/tel"
)

func TestResolverValid(t *testing.T) {
	const source = `
			(package pp "resolver")

			(document D [fst: A, snd: B]
			  (fst)
			  (snd))

			(component A [name: String]
			  (div {id: name}
			  	(name)))

			(component B []
			  (div))

			(component C []
			  (A [name: "string"]))
			`

	files := tel.NewFileSet("test")
	files.AddString("demo.tel", source)
	ctx := tel.NewContextWithFlags(tel.PreserveComment)
	pkg := runPackageResolver(ctx, files)
	if pkg.HasError() {
		t.Fatal("resolver failed unexpectedly")
	}
}

func TestResolverInvalid(t *testing.T) {
	sources := []struct {
		Desc string
		Name string
		Text string
		Flag tel.Flag
	}{
		{
			Name: "duplicate_attribute",
			Desc: "duplicate attribute in document template",
			Text: `(package p "views")
				   (document [] (div {id: "one" id: "two"}))`,
		},
		{
			Name: "duplicate_attribute",
			Desc: "duplicate attribute in named document template",
			Text: `(package p "views")
				   (document A [] (div {id: "one" id: "two"}))`,
		},
		{
			Name: "duplicate_attribute",
			Desc: "duplicate attribute in component template",
			Text: `(package p "views")
				   (component A [] (div {id: "one" id: "two"}))`,
		},
		{
			Name: "duplicate_attribute",
			Desc: "duplicate attribute in separate attribute set in document template",
			Text: `(package p "views")
				   (document [] (div {id: "one"} {id: "two"}))`,
		},
		{
			Name: "duplicate_attribute",
			Desc: "duplicate attribute in separate attribute set in named document template",
			Text: `(package p "views")
				   (document A [] (div {id: "one"} {id: "two"}))`,
		},
		{
			Name: "duplicate_attribute",
			Desc: "duplicate attribute in separate attribute set in component template",
			Text: `(package p "views")
				   (component A [] (div {id: "one"} {id: "two"}))`,
		},
		{
			Name: "duplicate_declaration",
			Desc: "duplicate component declaration",
			Text: `(package p "views")
				   (component A [])
				   (component A [])`,
		},
		{
			Name: "duplicate_declaration",
			Desc: "duplicate declaration",
			Text: `(package p "views")
				   (document A [])
				   (component A [])`,
		},
		{
			Name: "invalid_component",
			Desc: "duplicate parameter",
			Text: `(package p "views")
				   (component A [name: String])
				   (component B []
				     (A [ name: "john"
				        , name: "doe"
				        ]))`,
		},
	}
	for _, src := range sources {
		t.Run(src.Desc, func(t *testing.T) {
			files := tel.NewFileSet("test/" + src.Name)
			files.AddString(src.Name, src.Text)
			ctx := tel.NewContextWithFlags(src.Flag)
			pkg := runPackageResolver(ctx, files)
			if !pkg.HasError() {
				t.Error("ResolveFile() succeeded unexpectedly")
			}
		})
	}
}

type LookupOp struct {
	Name      string
	Scope     string
	Namespace string
}

func (b LookupOp) String() string {
	if b.Namespace == "" {
		return fmt.Sprintf("%s in %s", b.Name, b.Scope)
	}
	return fmt.Sprintf("%s.%s in %s", b.Namespace, b.Name, b.Scope)
}

func runPackageResolver(ctx tel.Context, srcsets tel.FileSet) Package {
	name := filepath.Base(srcsets.Path)
	builtinScope := newBuiltinScope(ctx)
	scope := newScope(builtinScope)

	pkg := Package{
		Name: name,
		Path: srcsets.Path,
	}
	context := &entityContext{
		Context: ctx,
		pkg:     &pkg,
		scope:   scope,
	}

	loadPackage(ctx, srcsets, &pkg)
	resolvePackage(context)
	return pkg
}
