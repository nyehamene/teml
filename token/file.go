package token

import "iter"

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

func (f *Tokenized) Tokens() iter.Seq[Token] {
	return func(yield func(Token) bool) {
		for _, t := range f.tokens {
			if !yield(t) {
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

func (f *Tokenized) add(kind Kind, pos Pos, text string) Position {
	assert(
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
