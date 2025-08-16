package transpiler

import (
	"path/filepath"

	"github.com/tel-lang/tel"
	"github.com/tel-lang/tel/ast"
)

type FileSet struct {
	Files   []File
	Path    string
	Package string
	Error   bool
}

type File struct {
	Name          string
	Package       PackageVar
	Imports       []Import
	TypeAliases   []TypeAlias
	Structs       []Struct
	Methods       []RenderMethod
	GlobalVars    []BlankVar
	RenderContext RenderContextStruct
}

const (
	//Deprecated:
	NameContextStruct = "RenderContext"
	//Deprecated:
	NameContextOptionStruct = "RenderContextOption"
	//Deprecated:
	NameContextContructor = "NewRenderContext"
	//Deprecated:
	NameComponentRenderMethod = "Render"
	//Deprecated:
	NameContextField = "ctx"
	//Deprecated:
	NameWriterField = "writer"
	//Deprecated:
	NameAttributeField = "attrs"
	//Deprecated:
	NameChildrenField = "children"
	//Deprecated:
	NameComponentInterface = "Component"
)

func ParseFile(ctx tel.Context, src tel.FileSet) FileSet {
	name := filepath.Base(src.Path)

	pkg := ast.Package{
		Name:     name,
		Path:     src.Path,
		FullName: "", // TBD(fix): get from tel.pack file
	}

	ast.LoadPackage(ctx, src, &pkg)
	generateSRC := FileSet{
		Files:   []File{},
		Error:   pkg.HasError(),
		Path:    pkg.Path,
		Package: pkg.FullName,
	}
	if generateSRC.HasError() {
		return generateSRC
	}
	for _, mod := range pkg.Modules {
		generatedFile := parseModule(mod)
		generateSRC.Files = append(generateSRC.Files, generatedFile)
	}
	return generateSRC
}

func parseModule(mod *ast.Module) File {
	psr := parser{}

	pkg := PackageVar{Var(mod.Package.Name)}

	imports := make([]Import, 0, len(mod.Imports))

	for _, i := range mod.Imports {
		node := psr.parseImport(i)
		imports = append(imports, node)
	}

	typeAliasCount := 0
	for _, u := range mod.Usings {
		typeAliasCount += len(u.Aliases)
	}
	typeAliases := make([]TypeAlias, 0, typeAliasCount)
	for _, u := range mod.Usings {
		for typealias := range psr.parseUsing(u) {
			typeAliases = append(typeAliases, typealias)
		}
	}

	structs := make([]Struct, 0, len(mod.Templates))
	methods := make([]RenderMethod, 0, len(mod.Templates))

	for _, decl := range mod.Templates {
		// reset temp variable counter
		errvarCount = 0
		tempvarCount = 0

		node := psr.parseStruct(decl)
		method := psr.parseMethod(decl)
		structs = append(structs, node)
		methods = append(methods, method)
	}

	// NOTE prevents 'unused variable error'
	// globalVars := []BlankVar{
	// 	BlankVar("fmt.Append"),
	// 	BlankVar("io.EOF"),
	// }
	var globalVars []BlankVar

	fdest := File{
		Name:          mod.Name,
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

func (f *FileSet) HasError() bool {
	return f.Error
}
