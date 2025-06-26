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
	eof token.Token = token.Token{Kind: -1, Pos: -1}
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

	type Order int
	const (
		OrderNone Order = iota
		OrderPackage
		OrderImport
		OrderDeclaration
	)

	lastOrder := OrderNone
	hasDocument := false

	for !p.eof() {
		decl, ok := p.parseDeclaration()
		if !ok {
			if flag&token.ExitOnError != 0 {
				return f, p.hasError
			}
			continue
		}

		switch d := decl.(type) {
		case Package:
			f.pkg = d
			lastOrder = OrderPackage

		case Import:
			f.imports = append(f.imports, d)
			if lastOrder != OrderPackage && lastOrder != OrderImport {
				p.addError("missing package declaration")
			}
			lastOrder = OrderImport

		case Using:
			f.usings = append(f.usings, d)
			if lastOrder == OrderDeclaration {
				p.addError("unexpected using declaration")
			} else if lastOrder != OrderImport {
				p.addError("missing import declaration")
			}

		case Document:
			f.document = d
			if hasDocument {
				p.addError("duplicate document declaration")
			}

			lastOrder = OrderDeclaration
			hasDocument = true

		case Component:
			f.components = append(f.components, d)
			lastOrder = OrderDeclaration

		default:
			p.addError("invalid declaration")
		}
	}

	assert.Assert(p.eof(), "expected eof")

	return f, p.hasError
}

func (p *parser) parsePackage() (Package, bool) {
	assert.Assert(p.peek().Kind == token.Package, "expected package keyword")

	var ident token.Token
	var path token.Token
	var ok bool

	// consume package keyword
	p.advance()

	if ident, ok = p.expect(token.Ident, "missing package identifier"); !ok {
		return Package{}, false
	}

	if path, ok = p.expect(token.String, "missing package path"); !ok {
		return Package{}, false
	}

	return Package{Ident: ident, Path: path}, true
}

func (p *parser) parseImport() (Import, bool) {
	assert.Assert(p.peek().Kind == token.Import, "expected import keyword")

	var ident token.Token
	var path token.Token
	var ok bool

	// consume import keyword
	p.advance()

	if ident, ok = p.expect(token.Ident, "missing import identifier"); !ok {
		return Import{}, false
	}

	if path, ok = p.expect(token.String, "missing import path"); !ok {
		return Import{}, false
	}

	return Import{Ident: ident, Path: path}, true
}

func (p *parser) parseUsing() (Using, bool) {
	assert.Assert(p.peek().Kind == token.Using, "expected using keyword")

	var u Using

	// consume using keyword
	p.advance()

	if ch := p.peek(); ch.Kind == token.BracketOpen {

		p.advance()

		for !p.eof() {

			if ch := p.peek(); ch.Kind != token.Ident {
				break
			}

			ident, ok := p.expect(token.Ident, "missing import alias")
			if !ok {
				return Using{}, false
			}

			u.idents = append(u.idents, ident)

			if ch := p.peek(); ch.Kind == token.Comma {
				p.advance()
			}
		}

		if _, ok := p.expect(token.BracketClose, "missing closing bracket"); !ok {
			return Using{}, false
		}

		if len(u.idents) == 0 {
			p.addError("empty import alias list")
			return Using{}, false
		}

	} else {
		ident, ok := p.expect(token.Ident, "missing import alias")
		if !ok {
			return Using{}, false
		}

		u.idents = append(u.idents, ident)
	}

	var ok bool
	if u.From, ok = p.expect(token.Ident, "missing package to alias from"); !ok {
		return Using{}, false
	}

	return u, true
}

func (p *parser) parseComponent() (Component, bool) {
	assert.Assert(p.peek().Kind == token.Component, "expected component keyword")

	var ident token.Token
	var properties []Property
	var children []Content
	var ok bool

	// consume component keyword
	p.advance()

	if ident, ok = p.expect(token.Ident, "missing component identifier"); !ok {
		return Component{}, false
	}

	if properties, ok = p.parseProperties(); !ok {
		return Component{}, false
	}

	if children, ok = p.parseChildren(); !ok {
		return Component{}, false
	}

	c := Component{Ident: ident, properties: properties, children: children}

	return c, true
}

func (p *parser) parseDocument() (Document, bool) {
	assert.Assert(p.peek().Kind == token.Document, "expected document keyword")

	var ident token.Token
	var properties []Property
	var children []Content
	var ok bool

	// consume document keyword
	p.advance()

	if ch := p.peek(); ch.Kind == token.Ident {
		p.advance()
		ident = ch
	}

	if properties, ok = p.parseProperties(); !ok {
		return Document{}, false
	}

	if children, ok = p.parseChildren(); !ok {
		return Document{}, false
	}

	d := Document{Ident: ident, properties: properties, children: children}
	return d, true
}

func (p *parser) parseProperties() ([]Property, bool) {
	var props []Property

	if _, ok := p.expect(token.BracketOpen, "missing opening bracket"); !ok {
		return nil, false
	}

	for !p.eof() {

		if ch := p.peek(); ch.Kind != token.Ident {
			break
		}

		prop, ok := p.parseProperty()
		if !ok {
			return nil, false
		}

		if ch := p.peek(); ch.Kind == token.Comma {
			p.advance()
		}

		props = append(props, prop)
	}

	if _, ok := p.expect(token.BracketClose, "missing closing bracket"); !ok {
		return nil, false
	}

	return props, true
}

func (p *parser) parseProperty() (Property, bool) {
	var ident token.Token
	// TODO type should be a qualified identifier
	var Type PropertyType
	var ok bool

	if ident, ok = p.expect(token.Ident, "missing property identifier"); !ok {
		return Property{}, false
	}

	if _, ok = p.expect(token.Colon, "missing type separator"); !ok {
		return Property{}, false
	}

	// parse property type
	switch ch := p.peek(); ch.Kind {
	case token.Ident:
		p.advance()
		Type = IdentPropertyType(ch)

	case token.ParenOpen:
		p.advance()

		if _, ok := p.expect(token.Enum, "missing enum keyword"); !ok {
			return Property{}, false
		}

		enumtype := EnumPropertyType{}

		var constantKind *token.Kind

	loop:
		for !p.eof() {
			ch := p.peek()

			switch ch.Kind {
			case token.ParenClose:
				break loop

			case token.String, token.Number:
				if constantKind != nil && ch.Kind != *constantKind {
					p.addError("mismatch enum constant type")
				} else {
					constantKind = &ch.Kind
				}
				enumtype.constants = append(enumtype.constants, ch)

			default:
				p.addError("invalid enum constant")
				return Property{}, false
			}

			// reached only when a valid constant is matched
			p.advance()

			if ch := p.peek(); ch.Kind == token.Comma {
				p.advance()
			}
		}

		// TODO fail is enum constants is empty

		if _, ok := p.expect(token.ParenClose, "missing close parenthesis"); !ok {
			return Property{}, false
		}

		Type = enumtype

	default:
		p.addError("missing property type")
		return Property{}, false
	}

	prop := Property{Ident: ident, Type: Type}
	return prop, true
}

func (p *parser) parseDeclaration() (Node, bool) {
	var node Node
	var ok bool

	if _, ok := p.expect(token.ParenOpen, "missing opening parenthesis"); !ok {
		// TODO move to parse method
		p.recover()
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
	case token.Ident:
		node, ok = badNode, false

		if _, ok := p.parseElement(true); ok {
			p.addError("unexpected element declaration")
		} else {
			p.addError("invalid declaration")
		}

	case token.String, token.StringLine, token.StringTempl, token.StringLineTempl:
		p.addError("unexpected content")
		p.advance()
		node, ok = badNode, false

	default:
		p.addError("invalid declaration")
		node, ok = badNode, false
	}

	if !ok {
		return node, false
	}

	if _, ok = p.expect(token.ParenClose, "missing closing parenthesis"); !ok {
		return badNode, false
	}

	return node, ok
}

func (p *parser) parseIfElement() (IfElement, bool) {
	var cond Expr
	var thenBranch Content
	var elseBranch Content
	var ok bool

	if _, ok = p.expect(token.ParenOpen, "missing opening parenthesis"); !ok {
		return IfElement{}, false
	}

	if _, ok = p.expect(token.If, "missing if keyword"); !ok {
		return IfElement{}, false
	}

	if cond, ok = p.parseExpr(); !ok {
		p.addError("invalid/missing condition expression")
		return IfElement{}, false
	}

	if thenBranch, ok = p.parseTemplate(); !ok {
		p.addError("missing content")
		return IfElement{}, false
	}

	if ch := p.peek(); ch.Kind != token.ParenClose {
		elseBranch, ok = p.parseTemplate()
		if !ok {
			return IfElement{}, false
		}
	}

	if _, ok = p.expect(token.ParenClose, "missing closing parenthesis"); !ok {
		return IfElement{}, false
	}

	return IfElement{cond: cond, thenBranch: thenBranch, elseBranch: elseBranch}, true
}

func (p *parser) parseElement(skipParenOpen bool) (Element, bool) {
	var ident Expr
	var children []Content
	var ok bool

	if !skipParenOpen {
		if _, ok = p.expect(token.ParenOpen, "missing opening parenthesis"); !ok {
			return Element{}, false
		}
	}

	if ident, ok = p.parseQualifiedName(); !ok {
		return Element{}, false
	}

	if children, ok = p.parseChildren(); !ok {
		return Element{}, false
	}

	if _, ok = p.expect(token.ParenClose, "missing closing parenthesis"); !ok {
		return Element{}, false
	}

	e := Element{Ident: ident, children: children}
	return e, true
}

func (p *parser) parseAttributes() (AttributeSet, bool) {
	var tag Expr
	var ok bool

	if ch := p.peek(); ch.Kind == token.Hash {
		p.advance()
		if tag, ok = p.parseQualifiedName(); !ok {
			return nil, false
		}
	}

	if _, ok = p.expect(token.BraceOpen, "missing opening brace"); !ok {
		return nil, false
	}

	var attrs []Attribute
	for !p.eof() {
		if ch := p.peek(); ch.Kind != token.Ident {
			break
		}

		attr, ok := p.parseAttribute()
		if !ok {
			return nil, false
		}

		attrs = append(attrs, attr)
	}

	if _, ok = p.expect(token.BraceClose, "missing closing brace"); !ok {
		return nil, false
	}

	var attrset AttributeSet
	if tag != nil {
		attrset = TaggedAttributeSet{tag: tag, attributes: attrs}
	} else {
		attrset = UntaggedAttributeSet{attributes: attrs}
	}

	return attrset, true
}

func (p *parser) parseAttribute() (Attribute, bool) {
	var key token.Token
	var value Expr
	var ok bool

	if key, ok = p.expect(token.Ident, "missing attribute key"); !ok {
		return Attribute{}, false
	}

	if _, ok = p.expect(token.Colon, "missing attribute value separator"); !ok {
		return Attribute{}, false
	}

	value, ok = p.parseExpr()
	if value == nil {
		return Attribute{}, false
	}

	if ch := p.peek(); ch.Kind == token.Comma {
		p.advance()
	}

	return Attribute{Key: key, value: value}, true
}

func (p *parser) parseTemplate() (Content, bool) {
	switch ch := p.peek(); ch.Kind {
	case token.ParenOpen:

		// if element
		switch ch := p.peekNext(); ch.Kind {
		case token.If:
			cond, ok := p.parseIfElement()
			if !ok {
				return nil, false
			}

			return cond, true

			// element
		case token.Ident:
			child, ok := p.parseElement(false)
			if !ok {
				return nil, false
			}

			return child, true
		}

	case token.String,
		token.StringLine,
		token.StringTempl,
		token.StringLineTempl:

		p.advance()
		text := Text(ch)
		return text, true
	}

	p.addError("invalid content")
	return nil, false
}

func (p *parser) parseChildren() ([]Content, bool) {
	var content []Content

	for !p.eof() {
		switch ch := p.peek(); ch.Kind {
		case token.Hash, token.BraceOpen:
			attrs, ok := p.parseAttributes()
			if !ok {
				return content, false
			}

			content = append(content, attrs)

		case token.ParenClose:
			return content, true

		default:
			templ, ok := p.parseTemplate()
			if !ok {
				return content, false
			}

			content = append(content, templ)
		}
	}

	return content, true
}

func (p *parser) parseQualifiedName() (Expr, bool) {
	var val token.Token
	var ok bool

	if val, ok = p.expect(token.Ident, "missing identifier"); !ok {
		return nil, false
	}

	var expr Expr = PrimaryExpr(val)

	for p.peek().Kind == token.FSlash {
		p.advance()

		if val, ok = p.expect(token.Ident, "missing identifier"); !ok {
			return nil, false
		}

		right := PrimaryExpr(val)
		expr = BinaryExpr{left: expr, right: right}
	}

	return expr, true
}

func (p *parser) parseExpr() (Expr, bool) {
	var expr Expr
	var ok bool

	switch ch := p.peek(); ch.Kind {
	case token.True,
		token.False,
		token.Number,
		token.Ident,
		token.String,
		token.StringTempl:
		expr, ok = PrimaryExpr(ch), true

	case token.ParenClose,
		token.BraceClose:
		expr, ok = nil, false
		p.addError("missing expression")
	default:
		expr, ok = nil, false
		p.addError("invalid expression")
	}

	p.advance()
	return expr, ok
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

func (p *parser) recover(continueFrom ...token.Kind) {
outer:
	for !p.eof() {
		ch := p.peek()
		for _, tok := range continueFrom {
			if ch.Kind == tok {
				break outer
			}
			if ch.Kind == token.ParenOpen {
				break outer
			}
		}
		p.advance()
	}
}

func (p *parser) advance() {
	if p.eof() {
		return
	}
	next := p.cur + 1
	p.cur = next

	p.skipComments()
}

func (p *parser) skipComments() {
	for !p.eof() {
		next := p.cur
		ch, ok := p.src.Token(next)
		if !ok {
			break
		}
		if ch.Kind != token.Comment {
			break
		}
		p.cur += 1
	}
}

func (p *parser) peekNext() token.Token {
	next := p.cur + 1
	if l := p.src.Size(); next >= l {
		return eof
	}

	tok, ok := p.src.Token(next)
	if !ok {
		return eof
	}

	return tok
}

func (p *parser) peek() token.Token {
	if p.eof() {
		return eof
	}

	next := p.cur
	node, ok := p.src.Token(next)
	if !ok {
		return eof
	}

	return node
}

func (p *parser) eof() bool {
	e := p.cur >= p.src.Size()
	return e
}

func (p *parser) addError(msg errmessage) {
	p.hasError = true
	p.printerr(errfmt.desc(msg))
}

type stringer interface {
	String() string
}

type errformatter struct{}
type errmessage = string

var errfmt errformatter

func (errformatter) desc(msg string) errmessage {
	return fmt.Sprintf(";desc: %s", msg)
}

func (errformatter) expect(msg stringer) errmessage {
	return fmt.Sprintf(";expected: %s", msg)
}

func (errformatter) got(msg stringer) errmessage {
	return fmt.Sprintf(";got: %s", msg)
}
