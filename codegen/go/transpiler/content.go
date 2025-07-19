package ast

import "fmt"

func formatParagraph(text string) string {
	return "<p>" + text + "</p>"
}

func stripDoubleQuote(quotedstr string) string {
	striped := quotedstr[1 : len(quotedstr)-1]
	return striped
}

var errvarCount = 0

func makeErrVar() string {
	name := fmt.Sprintf("err%d", errvarCount)
	errvarCount += 1
	return name
}
