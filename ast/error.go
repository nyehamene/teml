package ast

import (
	"fmt"

	"github.com/eml-lang/teml/token"
)

type ParseError struct {
	Line    int
	Col     int
	Pos     token.Position
	Message string
}

type errformatter struct{}
type errmessage = string

var errfmt errformatter

func (errformatter) desc(msg string) errmessage {
	return fmt.Sprintf(";desc: %s", msg)
}
