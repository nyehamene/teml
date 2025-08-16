package token

import (
	"iter"

	"github.com/tel-lang/tel"
)

func newFile(src []byte, file tel.File, size int) *File {
	f := &File{
		src:      src,
		Filename: file.Name,
		Path:     file.Path,
		Tokens:   make([]Token, 0, size),
	}
	return f
}

type File struct {
	Path     string
	Filename string
	Tokens   []Token
	src      []byte
	Comments []Token
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
	var (
		tokenEnd   = target.Offset + target.Length
		tokenStart = target.Offset
	)
	if tokenStart >= len(f.src) {
		return "", false // invalid start offset
	}
	if tokenEnd > len(f.src) {
		return "", false // invalid end offset
	}
	text := f.src[tokenStart:tokenEnd]
	return string(text), true
}
