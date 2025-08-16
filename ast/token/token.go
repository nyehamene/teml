package token

type Pos struct {
	Offset int
	Length int
	Line   int
	Column int
}

type Token struct {
	Pos
	Kind Kind
}

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
	"package":   Package,
	"import":    Import,
	"using":     Using,
	"component": Component,
	"document":  Document,
	"true":      True,
	"false":     False,
	"if":        If,
	"cond":      Cond,
	"enum":      Enum,
}

func isKeyword(ident []byte) (Kind, bool) {
	kw, ok := keywords[string(ident)]
	if !ok {
		return Invalid, false
	}
	return kw, true
}
