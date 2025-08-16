package token

import (
	"path"

	"github.com/tel-lang/tel"
)

type tokenizer struct {
	f   *File
	cur int
}

// Deprecated:
// Scan
func Scan(buf []byte, name string, flags ...tel.Flag) *File {
	input := tel.File{
		Path:    name,
		Name:    path.Base(name),
		Content: buf,
	}
	return ScanInput(input, flags...)
}

func ScanInput(src tel.File, flags ...tel.Flag) *File {
	var file *File
	var flag tel.Flag

	for _, f := range flags {
		flag |= f
	}

	buf := src.Content

	if flag&tel.ReduceAlloc != 0 {
		_, size := count(buf)
		file = newFile(buf, src, size)
	} else {
		file = newFile(buf, src, 0)
	}

	scan(file, flag)
	return file
}

func count(src []byte) (lines int, size int) {
	f := &File{src: src}
	t := tokenizer{f: f}

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

	return lines, size
}

func scan(f *File, flags tel.Flag) {
	t := tokenizer{f: f}

	curLineOffset := 0
	lineCount := 1

	for {
		t.skipSpace()

		if t.eof() {
			break
		}

		start := t.cur
		kind := t.next()
		end := t.cur

		if kind == Newline {
			curLineOffset = t.cur
			lineCount += 1

			addLine := flags & tel.PreserveNewline
			if addLine == 0 {
				continue
			}
		} else if kind == Comment {
			addComment := flags & tel.PreserveComment
			if addComment == 0 {
				continue
			}
		}

		length := end - start
		col := start - curLineOffset
		pos := Pos{
			Offset: start,
			Column: col,
			Length: length,
			Line:   lineCount,
		}
		tok := Token{
			Pos:  pos,
			Kind: kind,
		}

		if kind == Comment {
			f.Comments = append(f.Comments, tok)
			continue
		}

		f.Tokens = append(f.Tokens, tok)
	}
}

func (t *tokenizer) next() Kind {
	var kind Kind

	ch := t.peek()
	startOffset := t.cur

	if isAlpha(ch) {
		k := t.ident()
		lexeme := t.f.src[startOffset:t.cur]

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
	var kind Kind

	ch := t.peek()

	switch ch {
	case '[':
		kind = BracketOpen
		t.advance()
	case ']':
		kind = BracketClose
		t.advance()
	case '(':
		kind = ParenOpen
		t.advance()
	case ')':
		kind = ParenClose
		t.advance()
	case '{':
		kind = BraceOpen
		t.advance()
	case '}':
		kind = BraceClose
		t.advance()
	case ',':
		kind = Comma
		t.advance()
	case ':':
		kind = Colon
		t.advance()
	case '.':
		kind = Dot
		t.advance()
	case '\\':
		kind = BSlash
		t.advance()
	case '\n':
		kind = Newline
		t.advance()
	case '#':
		kind = Hash
		t.advance()
	case '"':
		kind = t.string()
		t.advance()
	case ';':
		kind = t.comment()
	case '-':
		if ch := t.peekNext(); ch == '-' {
			kind = t.stringLine()
			break
		}
		fallthrough
	default:
		kind = Invalid
		t.advance()
	}

	return kind
}

func (t *tokenizer) stringLine() Kind {
	tel.Assert(
		(t.peek() == '-' && t.peekNext() == '-'),
		"expected --",
	)

	t.advance()
	t.advance()

	isTempl := false

	for !t.eof() {
		ch := t.peek()
		if ch == '\n' {
			break
		}
		if ch == '\\' && t.peekNext() == '(' {
			isTempl = true
		}
		t.advance()
	}

	if isTempl {
		return StringLineTempl
	}

	return StringLine
}

func (t *tokenizer) number() Kind {
	tel.Assert(isDigit(t.peek()), "expected a digit")

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
	tel.Assert(t.peek() == '"', "expected \"")

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
	tel.Assert(isAlpha(t.peek()), "expected alpha")

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

func (t *tokenizer) comment() Kind {
	tel.Assert(t.peek() == ';', "expected ;")

	t.advance()

	for !t.eof() {
		if ch := t.peek(); ch == '\n' {
			break
		}
		t.advance()
	}

	return Comment
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
	ch := t.f.src[next]
	return ch
}

func (t *tokenizer) peekNext() byte {
	size := len(t.f.src)
	next := t.cur + 1

	if next >= size {
		return 0
	}

	ch := t.f.src[next]
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
	size := len(t.f.src)
	return cur >= size
}
