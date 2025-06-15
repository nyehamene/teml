package ast

import "github.com/eml-lang/teml/token"

type File struct {
	Pkg     Package
	imports []Import
}

type Node interface {
	node()
}

type Package struct {
	Ident token.Token
	Path  token.Token
}

type Import struct {
	Ident token.Token
	Path  token.Token
}

type IntErrorNode int

const (
	EOF IntErrorNode = iota
	unexpectedTokenError
)

func (Package) node() {}
func (Import) node()  {}

func (IntErrorNode) node() {}

func (f *File) addImport(imp Import) {
	f.imports = append(f.imports, imp)
}
