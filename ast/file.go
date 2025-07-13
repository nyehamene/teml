package ast

import (
	"github.com/eml-lang/teml/internal/errors"
	"github.com/eml-lang/teml/internal/slice"
)

type File struct {
	Package    Package
	Document   Document
	Imports    slice.Slice[Import]
	Usings     slice.Slice[Using]
	Components slice.Slice[Component]
	Errors     slice.Slice[errors.Error]
}

func (p *File) HasError() bool {
	return p.Errors.Size() > 0
}

func (p *File) HasDocument() bool {
	return p.Document.IsNamed() || p.Document.Properties.Size() > 0 || p.Document.Children.Size() > 0
}
