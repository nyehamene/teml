package token

import "fmt"

type Token struct {
	Kind Kind
	Pos  Position
}

type Position = int

//go:generate stringer -type=Kind
type Kind int

const (
	Invalid Kind = iota

	// BracketOpen [
	BracketOpen
	// BracketClose ]
	BracketClose
	// ParenOpen (
	ParenOpen
	// ParenClose )
	ParenClose
	// BraceOpen {
	BraceOpen
	// BraceClose }
	BraceClose

	Comma
	Colon
	BSlash
	Dot
	Hyphen
	Hash

	Package
	Import
	Using
	Component
	Document
	If
	Cond
	Enum

	Ident

	_literalBegin
	_constantBegin
	True
	False
	String
	Number
	_constantEnd

	StringLine
	StringTempl
	StringLineTempl
	_literalEnd

	Newline
	Comment
)

func newToken(kind Kind, pos Position) Token {
	t := Token{Kind: kind, Pos: pos}
	return t
}

func IsLiteral(k Kind) bool {
	return k > _literalBegin && k < _literalEnd
}

func IsConstant(k Kind) bool {
	return k > _constantBegin && k < _constantEnd
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
	True.String():      True,
	False.String():     False,
	If.String():        If,
	Cond.String():      Cond,
	Enum.String():      Enum,
}

var tokenString = map[Kind]string{
	Invalid: "invalid",

	Package:   "package",
	Import:    "import",
	Using:     "using",
	Component: "component",
	Document:  "document",
	True:      "true",
	False:     "false",
	If:        "if",
	Cond:      "cond",
	Enum:      "enum",

	Ident:           "ident",
	String:          "string",
	StringLine:      "l_string",
	StringTempl:     "t_string",
	StringLineTempl: "t_l_string",

	Number: "number",

	BracketOpen:  "[",
	BracketClose: "]",
	ParenOpen:    "(",
	ParenClose:   ")",
	BraceOpen:    "{",
	BraceClose:   "}",

	Comma:  ",",
	Colon:  ":",
	Dot:    ".",
	Hyphen: "-",
	Hash:   "#",

	Newline: "\\n",
	Comment: ";...",
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
		return fmt.Sprintf("%d", int(k))
	}

	return kw
}
