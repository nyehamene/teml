package token

import "github.com/eml-lang/teml/internal/source"

func newFile(src []byte, file source.File, size int, lines int) *File {
	f := &File{
		src:    src,
		Name:   file.Name,
		Path:   file.Path,
		Tokens: make([]Token, 0, size),
		Pos:    make([]Pos, 0, size),
		Lines:  make([]int, 0, lines),
	}
	// NOTE for one-based line numbering
	if len(src) > 0 {
		f.Lines = append(f.Lines, -1)
	}
	return f
}
