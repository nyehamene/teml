package token

import "fmt"

type Token struct {
	Kind Kind
	Pos  Pos
}

type Pos struct {
	Start int
	End   int
}

type Kind int

const (
	Invalid Kind = iota

	//
	// {{ Keyword
	//
	keywordStart
	Package
	Import
	Using
	Component
	Document
	And
	Or
	Not
	True
	False
	If
	keywordEnd
	// }}
)

func newToken(kind Kind, start, end int) Token {
	pos := Pos{Start: start, End: end}
	t := Token{Kind: kind, Pos: pos}
	return t
}

func isLetter(c byte) bool {
	// TODO handle other unicode letter
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		c == '_'
}

var stringKeyword = map[string]Kind{
	Package.String():   Package,
	Import.String():    Import,
	Using.String():     Using,
	Component.String(): Component,
	Document.String():  Document,
	And.String():       And,
	Or.String():        Or,
	Not.String():       Not,
	True.String():      True,
	False.String():     False,
	If.String():        If,
}

var keywordString = map[Kind]string{
	Invalid:   "invalid",
	Package:   "package",
	Import:    "import",
	Using:     "using",
	Component: "component",
	Document:  "document",
	And:       "and",
	Or:        "or",
	Not:       "not",
	True:      "true",
	False:     "false",
	If:        "if",
}

func isKeyword(ident []byte) (Kind, bool) {
	id := string(ident)
	kw, ok := stringKeyword[id]
	if !ok {
		return Invalid, false
	}
	return kw, true
}

func (k Kind) String() string {
	kw, ok := keywordString[k]

	if !ok {
		msg := fmt.Sprintf("Invalid token kind: %d", int(k))
		panic(msg)
	}

	return kw
}
