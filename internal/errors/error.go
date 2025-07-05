package errors

import (
	"fmt"

	"github.com/eml-lang/teml/token"
)

type Error struct {
	Line    int
	Col     int
	Pos     token.Position
	Message string
}

func Desc(msg string) string {
	return fmt.Sprintf(";desc: %s", msg)
}

func (p Error) String() string {
	return p.Message
}
