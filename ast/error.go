package ast

import (
	"fmt"
)

type asterror struct {
	Message string
	Line    int
	Col     int
}

func (err asterror) Error() string {
	errmsg := fmt.Sprintf("%s (%d, %d)", err.Message, err.Line, err.Col)
	return errmsg
}

func errdesc(err asterror) error {
	return fmt.Errorf(";desc: %s", err)
}

func errdescString(err string) error {
	return fmt.Errorf(";desc: %s", err)
}
