package token

import (
	"path"

	"github.com/eml-lang/teml/internal/assert"
	"github.com/eml-lang/teml/internal/source"
)

type tokenizer struct {
	f   *File
	cur int
}

type Flag uint

const (
	PreserveNewline Flag = 1 << iota
	PreserveComment
	ExitOnError
	HideErrors
	ReduceAlloc
)

// TODO remove
// @deprecate
func Scan(buf []byte, name string, flags ...Flag) *File {
	input := source.File{
		Path:    name,
		Name:    path.Base(name),
		Content: buf,
	}
	return ScanInput(input, flags...)
}

func ScanInput(src source.File, flags ...Flag) *File {
	var file *File
	var flag Flag

	for _, f := range flags {
		flag |= f
	}

	buf := src.Content

	if flag&ReduceAlloc != 0 {
		lines, size := count(buf)
		file = newFile(buf, src, size, lines)
	} else {
		file = newFile(buf, src, 0, 0)
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

func scan(f *File, flags Flag) {
	t := tokenizer{f: f}

	for {
		t.skipSpace()

		if t.eof() {
			break
		}

		start := t.cur
		kind := t.next()
		end := t.cur

		pos := Pos{Start: start, End: end}

		if kind == Newline {
			f.addLine(start)

			addLine := flags & PreserveNewline
			if addLine != 0 {
				f.add(kind, pos)
			}
			continue
		}

		if kind == Comment {
			addComment := flags & PreserveComment
			if addComment != 0 {
				f.add(kind, pos)
			}
			continue
		}

		f.add(kind, pos)
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
	assert.Assert(
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
	assert.Assert(isDigit(t.peek()), "expected a digit")

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
	assert.Assert(t.peek() == '"', "expected \"")

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
	assert.Assert(isAlpha(t.peek()), "expected alpha")

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
	assert.Assert(t.peek() == ';', "expected ;")

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
