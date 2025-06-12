package token

import "fmt"

type tokenizer struct {
	src []byte
	cur int
}

func Scan(src []byte) []Token {
	t := tokenizer{src: src}
	toks := []Token{}

	for !t.eof() {
		tok := t.next()
		toks = append(toks, tok)
	}

	return toks
}

func (t *tokenizer) next() Token {
	t.skipSpace()

	ch := t.peek()
	startOffset := t.cur

	if isAlpha(ch) {
		kind := t.ident()
		lexeme := t.src[startOffset:t.cur]
		if kw, ok := isKeyword(lexeme); ok {
			return newToken(kw, startOffset, t.cur)
		} else {
			return newToken(kind, startOffset, t.cur)
		}
	}

	//
	// NOTE avoid infine loop
	// This path is only taken if there is an invalid char
	//
	t.advance()
	fmt.Printf("invalid: %s\n", string(ch))
	return newToken(Invalid, startOffset, t.cur)
}

func (t *tokenizer) ident() Kind {
	start := t.cur

	for !t.eof() {
		ch := t.peek()
		if !isAlphaNumeric(ch) && ch != '-' {
			break
		}
		t.advance()
	}

	if empty := start == t.cur; empty {
		return Invalid
	}

	return Ident
}

func (t *tokenizer) skipSpace() {
	isSpace := map[byte]bool{
		' ':  true,
		'\t': true,
		'\v': true,
		'\r': true,
		'\f': true,
	}
	for !t.eof() {
		c := t.peek()
		if !isSpace[c] {
			break
		}
		t.advance()
	}
}

func (t *tokenizer) peek() byte {
	if t.eof() {
		return 0
	}
	next := t.cur
	c := t.src[next]
	return c
}

func (t *tokenizer) advance() {
	if t.eof() {
		return
	}
	next := t.cur + 1
	t.cur = next
}

func (t *tokenizer) eof() bool {
	cur := t.cur
	size := len(t.src)
	return cur >= size
}
