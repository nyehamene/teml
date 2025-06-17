package ast

import (
	"fmt"
	"log"
	"strings"

	"github.com/eml-lang/teml/assert"
	"github.com/eml-lang/teml/token"
)

type pToken = token.Token

type parser struct {
	src      token.Tokenized
	cur      int
	printerr func(string)
	hasError bool
	file     *File
	flag     token.Flags
}

var (
	tokenEOF pToken = pToken{Kind: -1, Pos: -1}
)

func Parse(toks token.Tokenized, flag token.Flags) (*File, bool) {
	printerr := func(s string) {
		log.Println(s)
	}
	return ParseWithErrorHandler(toks, flag, printerr)
}

func ParseWithErrorHandler(toks token.Tokenized, flag token.Flags, printerr func(string)) (*File, bool) {
	f := &File{}
	ok := parse(toks, f, flag, printerr)
	return f, ok

}

func parse(toks token.Tokenized, f *File, flag token.Flags, printerr func(string)) (hasError bool) {
	p := parser{
		src:  toks,
		file: f,
		flag: flag,
	}

	p.printerr = func(s string) {
		p.hasError = true
		printerr(s)
	}

	defer func() {
		hasError = p.hasError
	}()

	if flag&token.ReduceAlloc == 0 {
		f.adjustSize(toks)
	}

	p.parsePackage()
	p.parseImport()
	p.parseDeclarations()

	assert.Assert(p.eof(), "expected eof")

	return
}

func (p *parser) parsePackage() {
	p.expect(token.ParenOpen, "invalid delimiter")
	p.expect(token.Package, "invalid declaration")

	ident := p.expect(token.Ident, errfmt.title("Package Declaration Error"), errfmt.desc("missing identifier"))
	path := p.expect(token.String, errfmt.title("Package Declaration Error"), errfmt.desc("missing path string"))

	p.expect(token.ParenClose, "invalid delimiter")

	pkg := Package{Ident: ident, Path: path}
	p.file.pkg = pkg
}

func (p *parser) parseImport() {

	for !p.eof() {

		ch := p.peek()
		next := p.peekNext()

		if ch.Kind != token.ParenOpen ||
			next.Kind != token.Import {
			break
		}

		p.advance()
		p.advance()

		ident := p.expect(token.Ident, errfmt.title("Import Declaration Error"), errfmt.desc("missing identifier"))
		path := p.expect(token.String, errfmt.title("Import Declaration Error"), errfmt.desc("missing path string"))

		p.expect(token.ParenClose, errfmt.title("Import Declaration Error"), errfmt.desc("missing closing parenthesis ')'"))

		imp := Import{Ident: ident, Path: path}
		p.file.imports = append(p.file.imports, imp)

	}

	p.parseUsings()
}

func (p *parser) parseUsings() {
	for !p.eof() {
		ch := p.peek()
		next := p.peekNext()

		if ch.Kind != token.ParenOpen ||
			next.Kind != token.Using {
			break
		}

		p.advance()
		p.advance()

		use := Using{}

		ch = p.peek()

		if ch.Kind == token.BracketOpen {

			if p.flag&token.ReduceAlloc != 0 {
				use.adjustSize(p.cur, p.src)
			}

			p.advance()

			for !p.eof() {

				id := p.peek()
				if id.Kind == token.BracketClose {
					break
				}

				ident := p.expect(token.Ident, "missing identifier")
				use.idents = append(use.idents, ident)
			}

			p.expect(token.BracketClose, "missing closing bracket ']'")

		} else {
			ident := p.expect(token.Ident, errfmt.title("Using Declaration Error"), errfmt.desc("missing identifier"))
			use.idents = []token.Token{ident}
		}

		use.From = p.expect(token.Ident, errfmt.title("Using Declaration Error"), errfmt.desc("missing import identifier"))

		p.expect(token.ParenClose, errfmt.title("Using Declaration Error"), errfmt.desc("missing closing parenthesis ')'"))

		p.file.usings = append(p.file.usings, use)
	}
}

func (p *parser) parseComponent() {
	assert.Assert(p.peek().Kind == token.Component, "expected keyword 'component'")
	p.advance()

	ident := p.expect(token.Ident, "missing component identifier '('")
	cmp := Component{Ident: ident}

	p.expect(token.BracketOpen, "missing opening square bracket '['")

	cmp.properties = p.parseProperties()

	p.expect(token.BracketClose, "missing closing square bracket ']'")

	cmp.children = p.parseChildren()

	p.file.components = append(p.file.components, cmp)
}

func (p *parser) parseDocument() {
	assert.Assert(p.peek().Kind == token.Document, "expected keyword 'document'")

	p.advance()

	doc := Document{}

	if ch := p.peek(); ch.Kind == token.Ident {
		p.advance()
		doc.Ident = ch
	}

	p.expect(token.BracketOpen, "missing opening bracket '[")

	doc.properties = p.parseProperties()

	p.expect(token.BracketClose, "missing closing bracket '[")

	doc.children = p.parseChildren()

	p.file.document = doc
}

func (p *parser) parseProperties() []Property {
	var props []Property

	if p.flag&token.ReduceAlloc != 0 {
		props = createSizedPropertySlice(p.cur, p.src)
	}

	for !p.eof() {

		if ch := p.peek(); ch.Kind == token.BracketClose {
			break
		}

		prop := p.parseProperty()
		props = append(props, prop)
	}

	return props
}

func (p *parser) parseProperty() Property {
	ident := p.expect(token.Ident, "missing property identifier")
	p.expect(token.Colon, "missing type separator ':'")
	Type := p.expect(token.Ident, "missing property type")

	prop := Property{Ident: ident, Type: Type}

	if ch := p.peek(); ch.Kind == token.Comma {
		p.advance()
	}
	return prop
}

func (p *parser) parseDeclarations() {
	for !p.eof() {
		p.expect(token.ParenOpen, "missing opening parenthesis '('")

		ch := p.peek()

		switch ch.Kind {
		case token.Component:
			p.parseComponent()
		case token.Document:
			p.parseDocument()
		default:
			// TODO error
			return
		}

		p.expect(token.ParenClose, "missing closing parenthesis ')'")
	}
}

func (p *parser) parseElement() Element {
	p.expect(token.ParenOpen, "missing opening parenthesis '('")

	e := Element{}
	e.Ident = p.parseQualifiedName()

	// children
	e.children = p.parseChildren()

	p.expect(token.ParenClose, "missing closing parenthesis ')'")
	return e
}

func (p *parser) parseAttribute() Attribute {
	// tag
	attr := Attribute{}

	var tag Expr
	if ch := p.peek(); ch.Kind == token.Hash {
		p.advance()
		tag = p.parseQualifiedName()
	}

	if ch := p.peek(); ch.Kind == token.BraceOpen {
		p.advance()
		for !p.eof() {
			if ch := p.peek(); ch.Kind == token.BraceClose {
				break
			}

			attr.tag = tag
			attr.Ident = p.expect(token.Ident, "missing attribute key")

			p.expect(token.Colon, "missing attribute value separator ':'")

			attr.Value = p.parseExpr()

			if ch := p.peek(); ch.Kind == token.Comma {
				p.advance()
			}
		}
		p.expect(token.BraceClose, "missing closing square bracket ']'")
	}

	return attr
}

func (p *parser) parseChildren() []Content {
	var content []Content

	if p.flag&token.ReduceAlloc != 0 {
		content = createSizedContentSlize(p.cur, p.src)
	}

loop:
	for !p.eof() {
		switch ch := p.peek(); ch.Kind {
		case token.ParenOpen:
			child := p.parseElement()
			content = append(content, child)
		case token.Hash, token.BraceOpen:
			attr := p.parseAttribute()
			content = append(content, attr)
		case token.String, token.StringLine, token.StringTempl, token.StringLineTempl:
			p.advance()
			text := Text(ch)
			content = append(content, text)
		default:
			break loop
		}
	}
	return content
}

func (p *parser) parseQualifiedName() Expr {
	assert.Assert(p.peek().Kind == token.Ident, "expected identifier")
	val := p.expect(token.Ident, "missing identifier")

	var expr Expr = PrimaryExpr(val)

	for p.peek().Kind == token.FSlash {
		p.advance()
		right := PrimaryExpr(p.expect(token.Ident, "missing identifier"))
		expr = BinaryExpr{left: expr, right: right}
	}

	return expr
}

func (p *parser) parseExpr() Expr {
	switch ch := p.peek(); ch.Kind {
	case token.String, token.True, token.False, token.Number, token.StringTempl:
		p.advance()
		return PrimaryExpr(ch)
	}
	return nil
}

func (p *parser) expect(k token.Kind, msgs ...errmessage) pToken {
	ch := p.peek()
	if ch.Kind != k {
		exp := errfmt.expect(k)
		got := errfmt.got(ch.Kind)

		errs := make([]string, 0, len(msgs)+2)
		errs = append(errs, msgs...)
		errs = append(errs, exp, got)
		msg := strings.Join(errs, "\n")
		p.printerr(msg)
	}
	p.advance()
	return ch
}

func (p *parser) advance() {
	if p.eof() {
		return
	}
	next := p.cur + 1

	for ch := p.peek(); ch.Kind == token.Comment; {
		next += 1
	}

	p.cur = next
}

func (p *parser) peekNext() pToken {
	next := p.cur + 1
	size := p.src.Size()
	if next >= size {
		return tokenEOF
	}

	ch, ok := p.src.Token(next)
	if !ok {
		// unreachable
		assert.Assert(false, "unreachable")
		return tokenEOF
	}
	return ch
}

func (p *parser) peek() pToken {
	if p.eof() {
		return tokenEOF
	}

	next := p.cur
	if node, ok := p.src.Token(next); !ok {
		return tokenEOF
	} else {
		return node
	}
}

func (p *parser) eof() bool {
	e := p.cur >= p.src.Size()
	return e
}

type stringer interface {
	String() string
}
type errformatter struct{}
type errmessage = string

var errfmt errformatter

func (errformatter) title(msg string) errmessage {
	return fmt.Sprintf(";error: %s", msg)
}

func (errformatter) desc(msg string) errmessage {
	return fmt.Sprintf(";desc: %s", msg)
}

func (errformatter) expect(msg stringer) errmessage {
	return fmt.Sprintf(";expected: %s", msg)
}

func (errformatter) got(msg stringer) errmessage {
	return fmt.Sprintf(";got: %s", msg)
}
