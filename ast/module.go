package ast

import (
	"fmt"

	"github.com/tel-lang/tel"
	"github.com/tel-lang/tel/ast/token"
)

type Module struct {
	Imports   []ImportDecl
	Usings    []UsingDecl
	Enums     []EnumDecl
	Templates []TemplateDecl
	Comments  []Comment
	Package   PackageDecl
	Name      string
	hasError  bool
}

func loadModule(ctx tel.Context, src tel.File) *Module {
	toks := token.ScanInput(src, ctx.Flag)
	mod := &Module{Name: toks.Filename}
	p := parser{
		src:  toks,
		flag: ctx.Flag,
		mod:  mod,
	}
	p.parse()

	// preserve comment
	if len(toks.Comments) > 0 {
		for _, tok := range toks.Comments {
			text, ok := toks.Text(tok)
			if !ok {
				fmt.Printf("Could not get comment text for %#v\n", tok)
				continue
			}
			line, col := tok.Line, tok.Column
			cmt := Comment{Text: text, Line: line, Col: col}
			mod.Comments = append(mod.Comments, cmt)
		}
	}

	return mod
}
