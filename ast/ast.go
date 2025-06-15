package ast

import (
	"github.com/eml-lang/teml/assert"
	"github.com/eml-lang/teml/token"
)

type File struct {
	Pkg     Package
	imports []Import
	usings  []Using
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

type Using struct {
	idents []token.Token
	From   token.Token
}

type IntErrorNode int

const (
	EOF IntErrorNode = iota
	unexpectedTokenError
)

func (Package) node() {}
func (Import) node()  {}
func (Using) node()   {}

func (IntErrorNode) node() {}

func (u *Using) addIdent(id token.Token) {
	assert.Assert(id.Kind == token.Ident, "expected ident")
	u.idents = append(u.idents, id)
}

func (f *File) addImport(imp Import) {
	f.imports = append(f.imports, imp)
}

func (f *File) addUsing(use Using) {
	f.usings = append(f.usings, use)
}
