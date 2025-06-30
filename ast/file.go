package ast

import "github.com/eml-lang/teml/internal/slice"

type File struct {
	Package    Package
	Document   Document
	Imports    slice.Slice[Import]
	Usings     slice.Slice[Using]
	Components slice.Slice[Component]
	Errors     slice.Slice[ParseError]
}

func (p *File) HasError() bool {
	return p.Errors.Size() > 0
}
