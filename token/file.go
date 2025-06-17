package token

import (
	"iter"

	"github.com/eml-lang/teml/assert"
)

type Tokenized struct {
	tokens []Token
	pos    []Pos
	lines  []int
	src    []byte
}

type Pos struct {
	Start int
	End   int
}

func (f *Tokenized) adjustSize(size int, lines int) {
	f.tokens = make([]Token, 0, size)
	f.pos = make([]Pos, 0, size)
	f.lines = make([]int, 0, lines)
}

func (f *Tokenized) TokensFrom(pos int) iter.Seq2[int, Token] {
	return func(yield func(int, Token) bool) {
		for i := 0; pos < len(f.tokens); pos++ {
			if ch := f.tokens[pos]; !yield(i, ch) {
				break
			}
		}
	}
}

func (f *Tokenized) Tokens() iter.Seq2[int, Token] {
	return func(yield func(int, Token) bool) {
		for i, t := range f.tokens {
			if !yield(i, t) {
				break
			}
		}
	}
}

func (f Tokenized) Texts() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, tok := range f.Tokens() {
			if txt, ok := f.Text(tok); ok {
				if !yield(txt) {
					return
				}
			}
		}
	}
}

func (f Tokenized) Posses() iter.Seq[Pos] {
	return func(yield func(Pos) bool) {
		for _, t := range f.pos {
			if !yield(t) {
				return
			}
		}
	}
}

func (f Tokenized) Lines() iter.Seq[int] {
	return func(yield func(int) bool) {
		for _, l := range f.lines {
			if !yield(l) {
				return
			}
		}
	}
}

func (f Tokenized) Size() int {
	s := len(f.tokens)
	return s
}

func (f Tokenized) Token(i int) (Token, bool) {
	if i >= f.Size() {
		return Token{}, false
	}
	tok := f.tokens[i]
	return tok, true
}

func (f Tokenized) Text(target Token) (string, bool) {
	assert.Assert(
		len(f.tokens) == len(f.pos),
		"len of tokens and text are do not match",
	)

	for i, tok := range f.Tokens() {
		if target != tok {
			continue
		}
		pos := f.pos[i]
		txt := f.src[pos.Start:pos.End]
		return string(txt), true
	}
	return "", false
}

func (f *Tokenized) add(kind Kind, pos Pos) Position {
	assert.Assert(
		len(f.tokens) == len(f.pos),
		"expect tokens, pos, and text to have the same len",
	)

	position := len(f.tokens)
	p := Position(position)
	tok := newToken(kind, p)

	f.tokens = append(f.tokens, tok)
	f.pos = append(f.pos, pos)
	return Position(position)
}

func (f *Tokenized) addLine(line int) {
	f.lines = append(f.lines, line)
}
