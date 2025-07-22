package ast

import (
	"fmt"

	ast "github.com/eml-lang/teml/transpiler"
)

func doubleQuoteString(s string) String {
	quoted := fmt.Sprintf("\"%s\"", s)
	return String(quoted)
}

func escapeSurrounding0(str ast.String) String {
	prefix := string(str[0])
	suffix := string(str[len(str)-1])
	content := str[1 : len(str)-1]
	escape := "\\"
	txt := fmt.Sprintf("%s%s%[1]s%[3]s%[2]s%[4]s%[4]s", prefix, escape, content, suffix)
	return String(txt)
}

func escapeSurrounding(str ast.String) String {
	prefix := string(str[0])
	suffix := string(str[len(str)-1])
	content := str[1 : len(str)-1]
	escape := "\\"
	txt := fmt.Sprintf("%s%s%s%[1]s%[4]s", escape, prefix, content, suffix)
	return String(txt)
}

var errvarCount = 0

func makeErrVar() Var {
	// TODO return Var
	name := fmt.Sprintf("err%d", errvarCount)
	errvarCount += 1
	return Var(name)
}

var tempvarCount = 0

func makeTempVar() Var {
	// TODO return Var
	name := fmt.Sprintf("temp%d", tempvarCount)
	tempvarCount += 1
	return Var(name)
}
