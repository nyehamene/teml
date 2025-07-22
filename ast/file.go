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
	return p.Document.IsNamed() || len(p.Document.Properties) > 0 || len(p.Document.Children) > 0
}
