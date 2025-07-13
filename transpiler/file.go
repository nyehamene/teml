package ast

import (
	"fmt"
	"reflect"

	"github.com/eml-lang/teml/ast"
	"github.com/eml-lang/teml/internal/errors"
	"github.com/eml-lang/teml/token"
)

type Pos struct {
	Start int
	End   int
	Line  int
	Col   int
}

type File struct {
	Package      Package
	Imports      []Import
	Usings       []Using
	Declarations []Declaration
	errs         []errors.Error
}

func (f *File) HasError() bool {
	if f == nil {
		return false
	}
	return len(f.errs) > 0
}

func (f *File) Errors() func(func(int, errors.Error) bool) {
	return func(yield func(int, errors.Error) bool) {
		for i, err := range f.errs {
			if !yield(i, err) {
				break
			}
		}
	}
}

func ParseFile(src *ast.File, toks *token.File) *File {
	f := parseFile(src, toks)
	return f
}

func parseFile(src *ast.File, toks *token.File) *File {
	p := &parser{src: src, toks: toks}

	f := &File{
		errs: []errors.Error{},
	}

	f.Package = p.parsePackage()
	f.Imports = p.parseImports()
	f.Usings = p.parseUsings()

	decls := p.parseDeclarations()

	f.Declarations = decls
	return f
}

