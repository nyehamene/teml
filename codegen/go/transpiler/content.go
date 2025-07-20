package ast

import (
	"fmt"

	ast "github.com/eml-lang/teml/transpiler"
)

func createAttributes(kvs map[string]string) []ast.Attr {
	entries := []ast.KeyVal{}
	for k, v := range kvs {
		entry := ast.KeyVal{Key: ast.Var{Name: k}, Value: ast.String(doubleQuoteString(v))}
		entries = append(entries, entry)
	}
	attrs := []ast.Attr{
		{
			Tag:     nil,
			Entries: entries,
		},
	}
	return attrs
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

var tempvarCount = 0

func makeTempVar() string {
	name := fmt.Sprintf("temp%d", tempvarCount)
	tempvarCount += 1
	return name
}
