package ast

import ast "github.com/eml-lang/teml/transpiler"

type File struct {
	Package     Package
	Imports     []Import
	TypeAliases []TypeAlias
	Structs     []Struct
	Methods     []Method
}

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
	methods := make([]Method, 0, len(fsrc.Declarations))

	for _, decl := range fsrc.Declarations {
		// reset temp err variable name counter
		errvarCount = 0

		node := psr.parseStruct(decl)
		method := psr.parseMethod(decl)
		structs = append(structs, node)
		methods = append(methods, method)
	}

	fdest := File{
		Package:     pkg,
		Imports:     imports,
		TypeAliases: typeAliases,
		Structs:     structs,
		Methods:     methods,
	}
	return fdest
}

func addDefaultImports(imports *[]Import) {
	context := Import{Name: "context", Path: doubleQuoteString("context")}
	io := Import{Name: "io", Path: doubleQuoteString("io")}

	*imports = append(*imports, context, io)
}
