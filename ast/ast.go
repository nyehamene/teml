package ast

import "github.com/eml-lang/teml/token"

type File struct {
	Pkg Package
}

type Node interface {
	node()
}

type Package struct {
	Ident token.Token
	Path  token.Token
}

type IntErrorNode int

const (
	EOF IntErrorNode = iota
	unexpectedTokenError
)

func NewFile() *File {
	f := File{}
	return &f
}

func (Package) node() {}

func (IntErrorNode) node() {}
