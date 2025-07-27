package token

import (
	"fmt"
	"iter"

	"github.com/eml-lang/teml/internal/assert"
)

func NewFile(src []byte, name string, size int, lines int) *File {
	f := File{src: src, Name: name}
	f.Tokens = make([]Token, 0, size)
	f.Pos = make([]Pos, 0, size)
	f.Lines = make([]int, 0, lines)
	// NOTE for better line and column numbering
	if len(src) > 0 {
		f.Lines = append(f.Lines, -1)
	}
	return &f
}

type File struct {
	Name   string
	Tokens []Token
	Pos    []Pos
	Lines  []int
	src    []byte
}

type Pos struct {
	Start int
	End   int
}

func (f *File) Size() int {
	s := len(f.Tokens)
	return s
}

func (f *File) Texts() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, tok := range f.Tokens {
			if txt, ok := f.Text(tok); ok {
				if !yield(txt) {
					return
				}
			}
		}
	}
}

func (f File) Text(target Token) (string, bool) {
	assert.Assert(
		len(f.Tokens) == len(f.Pos),
		"len of tokens and text are do not match",
	)

	for i, tok := range f.Tokens {
		if target != tok {
			continue
		}

		if i >= len(f.Pos) || i < 0 {
			return "", false
		}
		pos := f.Pos[i]

		txt := f.src[pos.Start:pos.End]
		return string(txt), true
	}
	return "", false
}

func (f *File) Line(tok Token) (line int, col int) {
	if int(tok.Pos) >= len(f.Pos) || int(tok.Pos) < 0 {
		panic(fmt.Sprintf("could not get positon of token: %v", tok))
	}
	pos := f.Pos[tok.Pos]

	lst := 0
	offset := 0

	for _, ln := range f.Lines {
		if ln > pos.Start {
			break
		}
		lst += 1
		offset = ln
	}

	return lst, pos.Start - offset
}

func (f *File) add(kind Kind, pos Pos) Position {
	assert.Assert(
		len(f.Tokens) == len(f.Pos),
		"expect tokens, pos, and text to have the same len",
	)

	position := len(f.Tokens)
	p := Position(position)
	tok := newToken(kind, p)

	f.Tokens = append(f.Tokens, tok)
	f.Pos = append(f.Pos, pos)
	return p
}

func (f *File) addLine(line int) {
	f.Lines = append(f.Lines, line)
}
