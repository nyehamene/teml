package ast

import "fmt"

func formatParagraph(text string) string {
	return "<p>" + text + "</p>\\n"
}

func stripDoubleQuote(quotedstr string) string {
	striped := quotedstr[1 : len(quotedstr)-1]
	return striped
}

func doubleQuoteString(s string) string {
	quoted := fmt.Sprintf("\"%s\"", s)
	return quoted
}

func escapeSurrounding(str string) string {
	prefix := string(str[0])
	suffix := string(str[len(str)-1])
	content := str[1 : len(str)-1]
	escape := "\\"
	txt := fmt.Sprintf("%s%s%s%s%s", escape, prefix, content, escape, suffix)
	return txt
}

var errvarCount = 0

func makeErrVar() string {
	name := fmt.Sprintf("err%d", errvarCount)
	errvarCount += 1
	return name
}
