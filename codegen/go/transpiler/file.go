package ast

import (
	"github.com/eml-lang/teml/internal/flags"
	"github.com/eml-lang/teml/internal/source"
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
	NameContextStruct         = "RenderContext"
	NameContextOptionStruct   = "RenderContextOption"
	NameContextContructor     = "NewRenderContext"
	NameComponentRenderMethod = "Render"
	NameContextField          = "ctx"
	NameWriterField           = "writer"
	NameAttributeField        = "attrs"
	NameChildrenField         = "children"
	NameComponentInterface    = "Component"
	NameFuncComponentStruct   = "funccomponent"
)

// Deprecated: use Parse instead
//
// ParseFile0
func ParseFile0(src *ast.File) File {
	return parseFile(src)
}

func ParseFile(src source.File, cflags ...flags.Flag) File {
	astfile, _ := ast.TypecheckFile(src, cflags...)
	return parseFile(astfile)
}

func parseFile(fsrc *ast.File) File {
	psr := parser{}

	pkg := psr.parsePackage(fsrc.Package)

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
		Struct: Struct{Name: NameContextStruct,
			Fields: []StructField{
				{
					Name: NameContextField,
					Type: Var("context.Context"),
				},
				{
					Name: NameWriterField,
					Type: Var("io.Writer"),
				},
				{
					Name: NameAttributeField,
					Type: Var("map[string]string"),
				},
				{
					Name: NameChildrenField,
					Type: Var("[]Component"),
				},
			},
		},
		Constructor: NameContextContructor,
		OptionType:  NameContextOptionStruct,
	}
	return context
}
