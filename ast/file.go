package ast

import (
	"github.com/eml-lang/teml/internal/errors"
	"github.com/eml-lang/teml/token"
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

func ParseFile(toks *token.File, flags ...token.Flag) *File {
	var flag token.Flag

	for _, f := range flags {
		flag |= f
	}

	p := parser{
		src:  toks,
		flag: flag,
		dst:  &File{},
	}
	p.parse(flag)
	return p.dst
}
