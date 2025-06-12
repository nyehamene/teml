package token

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
	} else {
		kind := t.singleChars()
		return newToken(kind, startOffset, t.cur)
	}
}

func (t *tokenizer) singleChars() Kind {
	ch := t.peek()
	kind := Invalid

	switch ch {
	case '[':
		kind = BracketOpen
	case ']':
		kind = BracketClose
	case '(':
		kind = ParanOpen
	case ')':
		kind = ParenClose
	case '{':
		kind = BraceOpen
	case '}':
		kind = BraceClose
	case ',':
		kind = Comma
	case ':':
		kind = Colon
	case '/':
		kind = FSlash
	case '\\':
		kind = BSlash
	case '"':
		kind = t.string()
	case '-':
		if ch := t.peekNext(); ch == '-' {
			kind = t.stringLine()
		}
	}

	t.advance()
	return kind
}

func (t *tokenizer) stringLine() Kind {
	assert(
		(t.peek() == '-' && t.peekNext() == '-'),
		"expected --",
	)

	t.advance()
	t.advance()

	for !t.eof() {
		ch := t.peek()
		if ch == '\n' {
			break
		}
		t.advance()
	}

	return StringLine
}

func (t *tokenizer) string() Kind {
	assert(t.peek() == '"', "expected \"")

	t.advance()

	for !t.eof() {
		ch := t.peek()
		if ch == '"' {
			break
		}
		if ch == '\n' {
			break
		}
		t.advance()
	}

	if ch := t.peek(); ch != '"' {
		return Invalid
	}

	return String
}

func (t *tokenizer) ident() Kind {
	assert(isAlpha(t.peek()), "expected alpha")

	t.advance()

	for !t.eof() {
		ch := t.peek()
		if !isAlphaNumeric(ch) && ch != '-' {
			break
		}
		t.advance()
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
	ch := t.src[next]
	return ch
}

func (t *tokenizer) peekNext() byte {
	size := len(t.src)
	next := t.cur + 1

	if next >= size {
		return 0
	}

	ch := t.src[next]
	return ch
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
