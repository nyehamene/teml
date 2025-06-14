package token

type tokenizer struct {
	src []byte
	cur int
}

func Scan(src []byte) *Tokenized {
	f := NewFile()
	scan(src, f)
	return f
}

func ScanCountFirst(src []byte) *Tokenized {
	t := tokenizer{src: src}

	lines := 0
	size := 0
	for {
		t.skipSpace()
		if t.eof() {
			break
		}

		k := t.next()
		if k == Newline {
			lines += 1
		}

		size += 1
	}

	f := InitFile(size, lines)
	scan(src, f)

	return f
}

func scan(src []byte, f *Tokenized) {
	t := tokenizer{src: src}

	for {
		t.skipSpace()

		if t.eof() {
			break
		}

		start := t.cur
		kind := t.next()
		end := t.cur

		pos := Pos{Start: start, End: end}
		text := string(t.src[start:end])

		if kind == Newline {
			f.addLine(start)
		}

		f.add(kind, pos, text)
	}
}

func (t *tokenizer) next() Kind {
	ch := t.peek()
	startOffset := t.cur
	kind := Invalid

	if isAlpha(ch) {
		k := t.ident()
		lexeme := t.src[startOffset:t.cur]

		if kw, ok := isKeyword(lexeme); ok {
			kind = kw
		} else {
			kind = k
		}
	} else if isDigit(ch) {
		kind = t.number()
	} else {
		kind = t.singleChars()
	}

	return kind
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
	case '\n':
		kind = Newline
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

	isTempl := false

	for !t.eof() {
		ch := t.peek()
		if ch == '\\' && t.peekNext() == '(' {
			isTempl = true
		}
		if t.peekNext() == '\n' {
			break
		}
		t.advance()
	}

	if isTempl {
		return StringLineTempl
	}

	return StringLine
}

func (t *tokenizer) number() Kind {
	assert(isDigit(t.peek()), "expected a digit")

	t.advance()

	for !t.eof() {
		if ch := t.peek(); !isDigit(ch) {
			break
		}
		t.advance()
	}

	// match decimal number
	if ch := t.peek(); ch == '.' {
		if ch := t.peekNext(); !isDigit(ch) {
			// consume '.'
			t.advance()
			return Invalid
		}

		// consume '.'
		t.advance()

		for !t.eof() {
			if ch := t.peek(); !isDigit(ch) {
				break
			}
			t.advance()
		}
	}

	return Number
}

func (t *tokenizer) string() Kind {
	assert(t.peek() == '"', "expected \"")

	t.advance()

	isTempl := false

	for !t.eof() {
		ch := t.peek()
		if ch == '"' {
			break
		}
		if ch == '\n' {
			break
		}
		if ch == '\\' && t.peekNext() == '(' {
			isTempl = true
		}
		t.advance()
	}

	if isTempl {
		return StringTempl
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
