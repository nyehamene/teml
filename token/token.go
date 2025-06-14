package token

import "fmt"

type Token struct {
	Kind Kind
	Pos  Position
}

type Position int

//go:generate stringer -type=Kind
type Kind int

const (
	Invalid Kind = iota

	// {{ Bracket
	BracketOpen  // [
	BracketClose // ]
	ParanOpen    // (
	ParenClose   // )
	BraceOpen    // {
	BraceClose   // }
	// }} Barcket

	// {{ Delimiter
	Comma
	Colon
	BSlash
	FSlash
	Hyphen
	// }} Delimiter

	// {{ Keyword
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
	// }} Keyword

	Ident

	// {{ String
	String
	StringLine
	StringTempl
	StringLineTempl
	// }} String

	Number

	Newline
)

func newToken(kind Kind, pos Position) Token {
	t := Token{Kind: kind, Pos: pos}
	return t
}

func isAlphaNumeric(c byte) bool {
	// TODO handle other unicode letter
	return isAlpha(c) ||
		isDigit(c) ||
		c == '_'
}

func isAlpha(c byte) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z')
}

func isDigit(c byte) bool {
	return (c >= '0' && c <= '9')
}

var keywords = map[string]Kind{
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

var tokenString = map[Kind]string{
	Invalid: "invalid",

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

	Ident:           "ident",
	String:          "string",
	StringLine:      "l_string",
	StringTempl:     "t_string",
	StringLineTempl: "t_l_string",

	Number: "number",

	BracketOpen:  "[",
	BracketClose: "]",
	ParanOpen:    "(",
	ParenClose:   ")",
	BraceOpen:    "{",
	BraceClose:   "}",

	Comma:  ",",
	Colon:  ":",
	FSlash: "/",
	Hyphen: "-",

	Newline: "\\n",
}

func isKeyword(ident []byte) (Kind, bool) {
	kw, ok := keywords[string(ident)]
	if !ok {
		return Invalid, false
	}
	return kw, true
}

func (k Kind) String() string {
	kw, ok := tokenString[k]

	if !ok {
		msg := fmt.Sprintf("Invalid token kind: %d", int(k))
		panic(msg)
	}

	return kw
}
