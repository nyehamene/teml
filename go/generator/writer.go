package generator

import (
	"io"
	"unicode/utf8"
)

func NewSourceWriter(w io.Writer) SourceWriter {
	sw := SourceWriter{
		w: w,
	}
	return sw
}

type SourceWriter struct {
	w io.Writer
}

func (sw *SourceWriter) Write(s string) (int, error) {
	var err error
	utf8Bytes := make([]byte, 4)
	for _, c := range s {
		var n int

		ln := utf8.EncodeRune(utf8Bytes, c)
		n, err = sw.w.Write(utf8Bytes[:ln])
		if err != nil {
			return n, err
		}
	}

	return len(s), nil
}

// Deprecated: To be removed (tbd).
// Unwrap returns w.
func (sw *SourceWriter) Unwrap() io.Writer {
	return sw.w
}
