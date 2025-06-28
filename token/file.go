package token

import (
	"iter"

	"github.com/eml-lang/teml/internal/assert"
	"github.com/eml-lang/teml/internal/slice"
)

func NewFile(src []byte, size int, lines int) *File {
	f := File{src: src}
	f.Tokens = slice.Sized[Token](size)
	f.Pos = slice.Sized[Pos](size)
	f.Lines = slice.Sized[int](lines)
	return &f
}

type File struct {
	Tokens slice.Slice[Token]
	Pos    slice.Slice[Pos]
	Lines  slice.Slice[int]
	src    []byte
}

type Pos struct {
	Start int
	End   int
}

func (f *File) Size() int {
	s := f.Tokens.Size()
	return s
}

func (f *File) Texts() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, tok := range f.Tokens.Each() {
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
		f.Tokens.Size() == f.Pos.Size(),
		"len of tokens and text are do not match",
	)

	for i, tok := range f.Tokens.Each() {
		if target != tok {
			continue
		}

		pos, ok := f.Pos.Item(i)
		if !ok {
			return "", false
		}

		txt := f.src[pos.Start:pos.End]
		return string(txt), true
	}
	return "", false
}

func (f *File) add(kind Kind, pos Pos) Position {
	assert.Assert(
		f.Tokens.Size() == f.Pos.Size(),
		"expect tokens, pos, and text to have the same len",
	)

	position := f.Tokens.Size()
	p := Position(position)
	tok := newToken(kind, p)

	f.Tokens.Add(tok)
	f.Pos.Add(pos)
	return Position(position)
}

func (f *File) addLine(line int) {
	f.Lines.Add(line)
}
