package ast

import (
	"github.com/eml-lang/teml/internal/errors"
)

type File struct {
	Package    Package
	Document   Document
	Imports    []Import
	Usings     []Using
	Components []Component
	Errors     []errors.Error
}

func (p *File) HasError() bool {
	return len(p.Errors) > 0
}

func (p *File) HasDocument() bool {
	return p.Document.IsNamed() || p.Document.Properties.Size() > 0 || p.Document.Children.Size() > 0
}
