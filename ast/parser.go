package ast

import (
	"fmt"
	"log"

	"github.com/eml-lang/teml/assert"
	"github.com/eml-lang/teml/token"
)

type parser struct {
	src      token.Tokenized
	cur      int
	printerr func(string)
	hasError bool
	flag     token.Flags
}

var (
	tokenEOF token.Token = token.Token{Kind: -1, Pos: -1}
)

func Parse(toks token.Tokenized, flag token.Flags) (*File, bool) {
	printerr := func(s string) {
		log.Println(s)
	}
	return ParseWithErrorHandler(toks, flag, printerr)
}

func ParseWithErrorHandler(toks token.Tokenized, flag token.Flags, printerr func(string)) (*File, bool) {
	f, ok := parse(toks, flag, printerr)
	return f, ok

}

func parse(toks token.Tokenized, flag token.Flags, printerr func(string)) (*File, bool) {
	f := &File{}

	p := parser{
		src:  toks,
		flag: flag,
	}

	p.printerr = func(s string) {
		p.hasError = true
		printerr(s)
	}

	if flag&token.ReduceAlloc == 0 {
		f.adjustSize(toks)
	}

	for !p.eof() {
		decl, ok := p.parseDeclaration()
		if !ok {
			continue
		}

		switch d := decl.(type) {
		case Package:
			f.pkg = d
		case Import:
			f.imports = append(f.imports, d)
		case Using:
			f.usings = append(f.usings, d)
		case Component:
			f.components = append(f.components, d)
		}
	}

	assert.Assert(p.eof(), "expected eof")

	return f, p.hasError
}

func (p *parser) parsePackage() (Package, bool) {
	var ident token.Token
	var path token.Token
	var ok bool

	if _, ok := p.expect(token.Package, "missing package keyword"); !ok {
		return Package{}, false
	}

	if ident, ok = p.expect(token.Ident, "missing package identifier"); !ok {
		return Package{}, false
	}

	if path, ok = p.expect(token.String, "missing package path"); !ok {
		return Package{}, false
	}

	return Package{Ident: ident, Path: path}, true
}

func (p *parser) parseImport() (Import, bool) {
	var ident token.Token
	var path token.Token
	var ok bool

	if _, ok = p.expect(token.Import, "missing import keyword"); !ok {
		return Import{}, false
	}

	if ident, ok = p.expect(token.Ident, "missing import identifier"); !ok {
		return Import{}, false
	}

	if path, ok = p.expect(token.String, "missing import path string"); !ok {
		return Import{}, false
	}

	return Import{Ident: ident, Path: path}, true
}

func (p *parser) parseUsing() (Using, bool) {
	var u Using

	if _, ok := p.expect(token.Using, "missing using keyword"); !ok {
		return Using{}, false
	}

	if ch := p.peek(); ch.Kind == token.BracketOpen {

		p.advance()

		if p.flag&token.ReduceAlloc != 0 {
			u.adjustSize(p.cur, p.src)
		}

		for !p.eof() {

			if id := p.peek(); id.Kind == token.BracketClose {
				break
			}

			if ident, ok := p.expect(token.Ident, "missing identifier"); ok {
				u.idents = append(u.idents, ident)
			} else {
				return Using{}, false
			}
		}

		p.expect(token.BracketClose, "missing closing bracket")

	} else {
		if ident, ok := p.expect(token.Ident, "missing identifier"); ok {
			u.idents = append(u.idents, ident)
		} else {
			return Using{}, false
		}
	}

	var ok bool
	if u.From, ok = p.expect(token.Ident, "missing import identifier"); !ok {
		return Using{}, false
	}

	return u, true
}

func (p *parser) parseComponent() (Component, bool) {
	var ident token.Token
	var properties []Property
	var children []Content
	var ok bool

	if _, ok = p.expect(token.Component, "missing component keyword"); !ok {
		return Component{}, false
	}

	if ident, ok = p.expect(token.Ident, "missing component identifier"); !ok {
		return Component{}, false
	}

	if _, ok = p.expect(token.BracketOpen, "missing opening square bracket '['"); !ok {
		return Component{}, false
	}

	properties = p.parseProperties()

	if _, ok = p.expect(token.BracketClose, "missing closing square bracket ']'"); !ok {
		return Component{}, false
	}

	children = p.parseChildren()

	c := Component{Ident: ident, properties: properties, children: children}

	return c, true
}

func (p *parser) parseDocument() (Document, bool) {
	var ident token.Token
	var properties []Property
	var children []Content
	var ok bool

	if _, ok = p.expect(token.Document, "missing document keyword"); !ok {
		return Document{}, false
	}

	if ch := p.peek(); ch.Kind == token.Ident {
		p.advance()
		ident = ch
	}

	if _, ok = p.expect(token.BracketOpen, "missing opening bracket"); !ok {
		return Document{}, false
	}

	properties = p.parseProperties()

	if _, ok = p.expect(token.BracketClose, "missing closing bracket"); !ok {
		return Document{}, false
	}

	children = p.parseChildren()

	d := Document{Ident: ident, properties: properties, children: children}
	return d, true
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
	ident, _ := p.expect(token.Ident, "missing property identifier")

	p.expect(token.Colon, "missing type separator ':'")

	Type, _ := p.expect(token.Ident, "missing property type")

	prop := Property{Ident: ident, Type: Type}

	if ch := p.peek(); ch.Kind == token.Comma {
		p.advance()
	}
	return prop
}

func (p *parser) parseDeclaration() (Node, bool) {
	var node Node
	var ok bool

	if _, ok := p.expect(token.ParenOpen, "missing opening parenthesis"); !ok {
		return badNode, false
	}

	ch := p.peek()

	switch ch.Kind {
	case token.Package:
		node, ok = p.parsePackage()
	case token.Import:
		node, ok = p.parseImport()
	case token.Using:
		node, ok = p.parseUsing()
	case token.Document:
		node, ok = p.parseDocument()
	case token.Component:
		node, ok = p.parseComponent()
	default:
		node, ok = badNode, false
		p.advance()
	}

	if _, ok := p.expect(token.ParenClose, "missing closing parenthesis"); !ok {
		return badNode, false
	}

	return node, ok
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
			attr.Ident, _ = p.expect(token.Ident, "missing attribute key")

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
	val, _ := p.expect(token.Ident, "missing identifier")

	var expr Expr = PrimaryExpr(val)

	for p.peek().Kind == token.FSlash {
		p.advance()
		val, _ := p.expect(token.Ident, "missing identifier")
		right := PrimaryExpr(val)
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

func (p *parser) expect(k token.Kind, msg errmessage) (token.Token, bool) {
	ch := p.peek()
	if ch.Kind != k {
		p.printerr(errfmt.desc(msg))
		p.printerr(errfmt.expect(k))
		p.printerr(errfmt.got(ch.Kind))
		return ch, false
	}
	p.advance()
	return ch, true
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

func (p *parser) peekNext() token.Token {
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

func (p *parser) peek() token.Token {
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
