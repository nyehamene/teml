package ast

import (
	ast "github.com/eml-lang/teml/transpiler"
)

type File struct {
	Package       Package
	Imports       []Import
	TypeAliases   []TypeAlias
	Structs       []Struct
	Methods       []RenderMethod
	GlobalVars    []BlankVar
	RenderContext RenderContextStruct
}

const (
	RenderContext             = "RenderContext"
	RenderContextOption       = "RenderContextOption"
	RenderContextContructor   = "NewRenderContext"
	RenderComponentMethod     = "Render"
	RenderContextContextField = "ctx"
	RenderContextWriterField  = "writer"
	RenderContextAttrsField   = "attrs"
)

func Parse(fsrc *ast.File) File {
	psr := parser{fsrc}

	pkg := psr.parsePackage()

	imports := make([]Import, 0, len(fsrc.Imports)+2)

	addDefaultImports(&imports)

	for _, i := range fsrc.Imports {
		node := psr.parseImport(i)
		imports = append(imports, node)
	}

	typeAliasCount := 0
	for _, u := range fsrc.Usings {
		typeAliasCount += len(u.Idents)
	}
	typeAliases := make([]TypeAlias, 0, typeAliasCount)
	for _, u := range fsrc.Usings {
		for typealias := range psr.parseUsing(u) {
			typeAliases = append(typeAliases, typealias)
		}
	}

	structs := make([]Struct, 0, len(fsrc.Declarations))
	methods := make([]RenderMethod, 0, len(fsrc.Declarations))

	for _, decl := range fsrc.Declarations {
		// reset temp variable counter
		errvarCount = 0
		tempvarCount = 0

		node := psr.parseStruct(decl)
		method := psr.parseMethod(decl)
		structs = append(structs, node)
		methods = append(methods, method)
	}

	// NOTE prevents 'unused variable error'
	globalVars := []BlankVar{
		BlankVar("fmt.Append"),
	}

	fdest := File{
		Package:       pkg,
		Imports:       imports,
		TypeAliases:   typeAliases,
		RenderContext: createRenderContextStruct(),
		Structs:       structs,
		Methods:       methods,
		GlobalVars:    globalVars,
	}
	return fdest
}

func addDefaultImports(imports *[]Import) {
	context := Import{Name: "context", Path: doubleQuoteString("context")}
	io := Import{Name: "io", Path: doubleQuoteString("io")}
	fmt := Import{Name: "fmt", Path: doubleQuoteString("fmt")}

	*imports = append(*imports, context, io, fmt)
}

func createRenderContextStruct() RenderContextStruct {
	context := RenderContextStruct{
		Struct: Struct{Name: RenderContext,
			Fields: []StructField{
				{
					Name: RenderContextContextField,
					Type: "context.Context",
				},
				{
					Name: RenderContextWriterField,
					Type: "io.Writer",
				},
				{
					Name: RenderContextAttrsField,
					Type: "map[string]string",
				},
			},
		},
		Constructor: RenderContextContructor,
		OptionType:  RenderContextOption,
	}
	return context
}
