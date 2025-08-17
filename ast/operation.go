package ast

import (
	"fmt"
)

func (t Template) IsNamed() bool {
	return t.Ident.Name != ""
}

func (t Text) StripDelimiter() string {
	str := t.Value
	switch t.Kind {
	case LineTemplateText, LineText:
		return str[2:]

	case QuotedTemplateText, QuotedText:
		return str[1 : len(str)-1]

	default:
		panic(fmt.Sprintf("unexpected ast.TextKind: %#v", t.Kind))
	}
}
