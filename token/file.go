package token

import (
	"iter"

	"github.com/eml-lang/teml/assert"
)

type Tokenized struct {
	tokens []Token
	pos    []Pos
	text   []string
	lines  []int
}

type Pos struct {
	Start int
	End   int
}

func NewFile() *Tokenized {
	f := Tokenized{}
	return &f
}

func InitFile(size int, lines int) *Tokenized {
	f := Tokenized{}
	f.tokens = make([]Token, 0, size)
	f.pos = make([]Pos, 0, size)
	f.text = make([]string, 0, size)
	f.lines = make([]int, 0, lines)
	return &f
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
		for _, t := range f.text {
			if !yield(t) {
				return
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
		len(f.tokens) == len(f.text),
		"len of tokens and text are do not match",
	)

	for i, tok := range f.Tokens() {
		if target != tok {
			continue
		}
		txt := f.text[i]
		return txt, true
	}
	return "", false
}

func (f *Tokenized) add(kind Kind, pos Pos, text string) Position {
	assert.Assert(
		(len(f.tokens) == len(f.pos) && len(f.pos) == len(f.text)),
		"expect tokens, pos, and text to have the same len",
	)
	position := len(f.tokens)
	p := Position(position)
	tok := newToken(kind, p)

	f.tokens = append(f.tokens, tok)
	f.pos = append(f.pos, pos)
	f.text = append(f.text, text)
	return Position(position)
}

func (f *Tokenized) addLine(line int) {
	f.lines = append(f.lines, line)
}
