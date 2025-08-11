package ast

import (
	"fmt"

	"github.com/eml-lang/teml/internal/source"
	"github.com/eml-lang/teml/token"

	cflags "github.com/eml-lang/teml/internal/flags"
)

type File struct {
	Name       string
	Package    Package
	Document   Document
	Imports    []Import
	Usings     []Using
	Components []Component
	Errors     []error
	Comments   []Comment
}

func (p *File) HasError() bool {
	return len(p.Errors) > 0
}

func (p *File) HasDocument() bool {
	return p.Document.IsNamed() || len(p.Document.Properties) > 0 || len(p.Document.Children) > 0
}

// Deprecated: use ParseFile(source.File, ...cflags.Flag) instead
func ParseFile0(toks *token.File, flags ...cflags.Flag) *File {
	var flag cflags.Flag

	for _, f := range flags {
		flag |= f
	}

	file := &File{Name: toks.Name}

	p := parser{
		src:  toks,
		flag: flag,
		dst:  file,
	}
	p.parse(flag)
	return p.dst
}

func ParseFile(src source.File, flags ...cflags.Flag) *File {
	var flag cflags.Flag
	for _, f := range flags {
		flag |= f
	}

	toks := token.ScanInput(src, flag)
	file := parseFile(toks, flag)

	// preserve comment
	if flag&cflags.PreserveComment != 0 {
		for _, tok := range toks.Tokens {
			if tok.Kind != token.Comment {
				continue
			}
			text, ok := toks.Text(tok)
			if !ok {
				fmt.Printf("Could not get comment text for %#v\n", tok)
				continue
			}
			line, col := toks.Line(tok)
			cmt := Comment{Text: text, Line: line, Col: col}
			file.Comments = append(file.Comments, cmt)
		}
	}

	return file
}

func parseFile(toks *token.File, flag cflags.Flag) *File {
	file := &File{Name: toks.Name}
	p := parser{src: toks, flag: flag, dst: file}
	p.parse(flag)
	return p.dst
}
