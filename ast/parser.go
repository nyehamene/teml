package ast

import (
	"fmt"
	"log"

	"github.com/eml-lang/teml/internal/assert"
	"github.com/eml-lang/teml/internal/slice"
	"github.com/eml-lang/teml/token"
)

type parser struct {
	src      token.File
	cur      int
	printerr func(string)
	hasError bool
	flag     token.Flags
}

var (
	eof token.Token = token.Token{Kind: -1, Pos: -1}
)

func Parse(toks token.File, flag token.Flags) (*File, bool) {
	printerr := func(s string) {
		log.Println(s)
	}
	return ParseWithErrorHandler(toks, flag, printerr)
}

func ParseWithErrorHandler(toks token.File, flag token.Flags, printerr func(string)) (*File, bool) {
	f, ok := parse(toks, flag, printerr)
	return f, ok

}

func parse(toks token.File, flag token.Flags, printerr func(string)) (*File, bool) {
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
			f.Package = d
			lastOrder = OrderPackage

		case Import:
			f.Imports.Add(d)
			if lastOrder != OrderPackage && lastOrder != OrderImport {
				p.addError("missing package declaration")
			}
			lastOrder = OrderImport

		case Using:
			f.Usings.Add(d)
			if lastOrder == OrderDeclaration {
				p.addError("unexpected using declaration")
			} else if lastOrder != OrderImport {
				p.addError("missing import declaration")
			}

		case Document:
			f.Document = d
			if hasDocument {
				p.addError("duplicate document declaration")
			}

			lastOrder = OrderDeclaration
			hasDocument = true

		case Component:
			f.Components.Add(d)
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

			u.Idents.Add(ident)

			if ch := p.peek(); ch.Kind == token.Comma {
				p.advance()
			}
		}

		if _, ok := p.expect(token.BracketClose, "missing closing bracket"); !ok {
			return Using{}, false
		}

		if u.Idents.Size() == 0 {
			p.addError("empty import alias list")
			return Using{}, false
		}

	} else {
		ident, ok := p.expect(token.Ident, "missing import alias")
		if !ok {
			return Using{}, false
		}
		u.Idents.Add(ident)
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

	for !p.eof() {
		if ch := p.peek(); ch.Kind == token.ParenClose {
			break
		}

		templ, ok := p.parseTemplate()
		if !ok {
			return Component{}, false
		}
		children = append(children, templ)
	}

	c := Component{Ident: ident, Properties: slice.New(properties), Children: slice.New(children)}

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

	for !p.eof() {
		if ch := p.peek(); ch.Kind == token.ParenClose {
			break
		}

		templ, ok := p.parseTemplate()
		if !ok {
			return Document{}, false
		}
		children = append(children, templ)
	}

	d := Document{Ident: ident, Properties: slice.New(properties), Children: slice.New(children)}
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
				enumtype.Constants.Add(ch)

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

func (p *parser) parseCondElement() (CondElement, bool) {
	var cond Expr
	var options []CondElementOption
	var ok bool

	if _, ok = p.expect(token.ParenOpen, "missing opening parenthesis"); !ok {
		return CondElement{}, false
	}

	if _, ok = p.expect(token.Cond, "missing cond keyword"); !ok {
		return CondElement{}, false
	}

	if cond, ok = p.parseExpr(); !ok {
		return CondElement{}, false
	}

	for !p.eof() {

		if ch := p.peek(); ch.Kind == token.ParenClose {
			break
		}

		lit, ok := p.parseLiteral()
		if !ok {
			return CondElement{}, false
		}

		p.expect(token.Colon, "missing colon")

		content, ok := p.parseTemplate()
		if !ok {
			return CondElement{}, false
		}

		if ch := p.peek(); ch.Kind == token.Comma {
			p.advance()
		}

		option := CondElementOption{
			constant: lit,
			branch:   content,
		}

		options = append(options, option)
	}

	if _, ok := p.expect(token.ParenClose, "missing closing parenthese"); !ok {
		return CondElement{}, false
	}

	e := CondElement{
		cond:    cond,
		Options: slice.New(options),
	}

	return e, true
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

loop:
	for !p.eof() {
		switch ch := p.peek(); ch.Kind {
		case token.ParenClose:
			break loop

		case token.Hash, token.BraceOpen:
			attrs, ok := p.parseAttributes()
			if !ok {
				return Element{}, false
			}
			children = append(children, attrs)

		default:
			templ, ok := p.parseTemplate()
			if !ok {
				return Element{}, false
			}
			children = append(children, templ)
		}
	}

	if _, ok = p.expect(token.ParenClose, "missing closing parenthesis"); !ok {
		return Element{}, false
	}

	e := Element{Ident: ident, Children: slice.New(children)}
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
		attrset = TaggedAttributeSet{tag: tag, Attributes: slice.New(attrs)}
	} else {
		attrset = UntaggedAttributeSet{Attributes: slice.New(attrs)}
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

		switch ch := p.peekNext(); ch.Kind {

		// if element
		case token.If:
			if cond, ok := p.parseIfElement(); ok {
				return cond, true
			}

		// cond element
		case token.Cond:
			if cond, ok := p.parseCondElement(); ok {
				return cond, true
			}

		// element
		case token.Ident:
			if child, ok := p.parseElement(false); ok {
				return child, true
			}
		}

	case token.String,
		token.StringLine,
		token.StringTempl,
		token.StringLineTempl:

		p.advance()
		text := Text(ch)
		return text, true

	case token.Ident:
		p.addError("invalid content")

	default:
		p.addError("missing content")
	}

	return nil, false
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

func (p *parser) parseLiteral() (Expr, bool) {
	var expr Expr
	var ok bool

	switch ch := p.peek(); ch.Kind {
	case token.True,
		token.False,
		token.Number,
		token.String,
		token.StringTempl:

		expr, ok = PrimaryExpr(ch), true

	default:
		p.addError("invalid expression")
	}

	p.advance()
	return expr, ok
}

func (p *parser) parseExpr() (Expr, bool) {
	var expr Expr
	var ok bool

	switch ch := p.peek(); ch.Kind {
	case token.ParenClose,
		token.BraceClose:
		expr, ok = nil, false
		p.addError("missing expression")

	case token.Ident:
		p.advance()
		expr, ok = PrimaryExpr(ch), true

	case token.ParenOpen:

		switch ch := p.peekNext(); ch.Kind {
		case token.If:
			expr, ok = p.parseIfExpression()
		case token.Cond:
			expr, ok = p.parseCondExpression()
		}

	default:
		expr, ok = p.parseLiteral()
	}

	return expr, ok
}

func (p *parser) parseCondExpression() (Expr, bool) {
	var cond Expr
	var options []CondExpressionOption
	var ok bool

	if _, ok = p.expect(token.ParenOpen, "missing opening parenthesis"); !ok {
		return nil, false
	}

	if _, ok = p.expect(token.Cond, "missing cond keyword"); !ok {
		return nil, false
	}

	if cond, ok = p.parseExpr(); !ok {
		return nil, false
	}

	for !p.eof() {
		if ch := p.peek(); ch.Kind == token.ParenClose {
			break
		}

		lit, ok := p.parseLiteral()
		if !ok {
			return nil, false
		}

		p.expect(token.Colon, "missing colon")

		value, ok := p.parseExpr()
		if !ok {
			return nil, false
		}

		if ch := p.peek(); ch.Kind == token.Comma {
			p.advance()
		}

		option := CondExpressionOption{constant: lit, value: value}
		options = append(options, option)
	}

	if _, ok = p.expect(token.ParenClose, "missing closing parenthesis"); !ok {
		return nil, false
	}

	e := CondExpression{cond: cond, Options: slice.New(options)}
	return e, true
}

func (p *parser) parseIfExpression() (Expr, bool) {
	var cond Expr
	var thenBranch Expr
	var elseBranch Expr
	var ok bool

	if _, ok = p.expect(token.ParenOpen, "missing opening parenthesis"); !ok {
		return nil, false
	}

	if _, ok = p.expect(token.If, "missing if keyword"); !ok {
		return nil, false
	}

	if cond, ok = p.parseExpr(); !ok {
		return nil, false
	}

	if thenBranch, ok = p.parseExpr(); !ok {
		return nil, false
	}

	if ch := p.peek(); ch.Kind != token.ParenClose {
		if elseBranch, ok = p.parseExpr(); !ok {
			return nil, false
		}
	}

	if _, ok := p.expect(token.ParenClose, "missing closing parenthesis"); !ok {
		return nil, false
	}

	return IfExpression{cond: cond, thenBranch: thenBranch, elseBranch: elseBranch}, true
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
		ch, ok := p.src.Tokens.Item(next)
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

	tok, ok := p.src.Tokens.Item(next)
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
	node, ok := p.src.Tokens.Item(next)
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
