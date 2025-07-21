package ast

import (
	"github.com/eml-lang/teml/internal/assert"
	"github.com/eml-lang/teml/internal/errors"
	"github.com/eml-lang/teml/internal/slice"
	"github.com/eml-lang/teml/token"
)

type errmessage = string

type parser struct {
	src  *token.File
	dst  *File
	cur  int
	flag token.Flags
}

var (
	eof token.Token = token.Token{Kind: -1, Pos: -1}
)

func ParseFile(toks *token.File, flag token.Flags) *File {
	p := parser{
		src:  toks,
		flag: flag,
		dst:  &File{},
	}
	p.parse(flag)
	return p.dst
}

func (p *parser) parse(flag token.Flags) {
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
				return
			}
			continue
		}

		switch d := decl.(type) {
		case Package:
			p.dst.Package = d
			lastOrder = OrderPackage

		case Import:
			p.dst.Imports = append(p.dst.Imports, d)
			if lastOrder != OrderPackage && lastOrder != OrderImport {
				p.addError("missing package declaration")
			}
			lastOrder = OrderImport

		case Using:
			p.dst.Usings = append(p.dst.Usings, d)
			if lastOrder == OrderDeclaration {
				p.addError("unexpected using declaration")
			} else if lastOrder != OrderImport {
				p.addError("missing import declaration")
			}

		case Document:
			p.dst.Document = d
			if hasDocument {
				p.addError("duplicate document declaration")
			}

			lastOrder = OrderDeclaration
			hasDocument = true

		case Component:
			p.dst.Components = append(p.dst.Components, d)
			lastOrder = OrderDeclaration

		default:
			p.addError("invalid declaration")
		}
	}

	assert.Assert(p.eof(), "expected eof")
}

func (p *parser) parsePackage() (Package, bool) {
	assert.Assert(p.peek().Kind == token.Package, "expected package keyword")

	var ident Var
	var path token.Token
	var ok bool

	// consume package keyword
	p.advance()

	if ident, ok = p.parseVar(); !ok {
		return Package{}, false
	}

	if path, ok = p.expect(token.String, "missing package path"); !ok {
		return Package{}, false
	}

	return Package{Ident: ident, Path: path}, true
}

func (p *parser) parseImport() (Import, bool) {
	assert.Assert(p.peek().Kind == token.Import, "expected import keyword")

	var ident Var
	var path token.Token
	var ok bool

	// consume import keyword
	p.advance()

	if ident, ok = p.parseVar(); !ok {
		return Import{}, false
	}

	if path, ok = p.expect(token.String, "missing import path"); !ok {
		return Import{}, false
	}

	return Import{Ident: ident, Path: Constant(path)}, true
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

			ident, ok := p.parseVar()
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
		ident, ok := p.parseVar()
		if !ok {
			return Using{}, false
		}
		u.Idents.Add(ident)
	}

	var ok bool
	var from Var

	if from, ok = p.parseVar(); !ok {
		return Using{}, false
	}

	u.From = from
	return u, true
}

func (p *parser) parseComponent() (Component, bool) {
	assert.Assert(p.peek().Kind == token.Component, "expected component keyword")

	var ident Var
	var properties []Property
	var children []Content
	var ok bool

	// consume component keyword
	p.advance()

	if ident, ok = p.parseVar(); !ok {
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

	var ident Var
	var properties []Property
	var children []Content
	var ok bool

	// consume document keyword
	p.advance()

	if ch := p.peek(); ch.Kind == token.Ident {
		ident, _ = p.parseVar()
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
	var ident Var
	var Type PropertyType
	var ok bool

	if ident, ok = p.parseVar(); !ok {
		return Property{}, false
	}

	if _, ok = p.expect(token.Colon, "missing type separator"); !ok {
		return Property{}, false
	}

	// parse property type
	if Type, ok = p.parsePropertyType(); !ok {
		return Property{}, false
	}

	prop := Property{Ident: ident, Type: Type}
	return prop, true
}

func (p *parser) parsePropertyType() (PropertyType, bool) {
	switch ch := p.peek(); ch.Kind {
	case token.Ident:
		switch ch := p.peekNext(); ch.Kind {
		case token.FSlash:
			var left Expr

			left, _ = p.parseVar()
			for p.peek().Kind == token.FSlash {
				p.advance() // consume forward slash

				right, ok := p.parseVar()
				if !ok {
					return nil, false
				}
				left = MemberAccess{Object: left, Member: right}
			}

			return left.(MemberAccess), true

		default:
			ident, _ := p.parseVar()
			return ident, true
		}

	case token.ParenOpen:
		p.advance()
		if _, ok := p.expect(token.Enum, "missing enum keyword"); !ok {
			return nil, false
		}

		enumtype := Enum{}
		var constantKind *token.Kind

		for !p.eof() {
			var constant token.Token
			var ok bool

			if ch := p.peek(); ch.Kind == token.ParenClose {
				break
			}

			if constant, ok = p.parseEnumConstant(); !ok {
				return nil, false
			}

			if constantKind != nil && constant.Kind != *constantKind {
				p.addError("mismatch enum constant type")
			} else {
				constantKind = &constant.Kind
			}

			enumtype.Constants.Add(Constant(constant))
		}

		// TODO fail if enum constants is empty

		if _, ok := p.expect(token.ParenClose, "missing close parenthesis"); !ok {
			return nil, false
		}

		return enumtype, true

	default:
		p.addError("missing property type")
		return nil, false
	}
}

func (p *parser) parseEnumConstant() (token.Token, bool) {
	switch ch := p.peek(); ch.Kind {
	case token.String,
		token.Number:

		p.advance()
		// consume semicolon
		if ch := p.peek(); ch.Kind == token.Comma {
			p.advance()
		}
		return ch, true
	default:
		p.addError("invalid enum constant")
		return token.Token{}, false
	}
}

func (p *parser) parseDeclaration() (Node, bool) {
	var node Node
	var ok bool

	if _, ok := p.expect(token.ParenOpen, "missing opening parenthesis"); !ok {
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
	var options []CaseElement
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

		option := CaseElement{
			Cond:   lit,
			Branch: content,
		}

		options = append(options, option)
	}

	if _, ok := p.expect(token.ParenClose, "missing closing parenthese"); !ok {
		return CondElement{}, false
	}

	e := CondElement{
		Target: cond,
		Cases:  slice.New(options),
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

	return IfElement{Cond: cond, Then: thenBranch, Else: elseBranch}, true
}

func (p *parser) parseElement(skipParenOpen bool) (Element, bool) {
	var ident Expr
	var attributes []AttributeSet
	var children []Content
	var ok bool

	if !skipParenOpen {
		if _, ok = p.expect(token.ParenOpen, "missing opening parenthesis"); !ok {
			return Element{}, false
		}
	}

	if ident, ok = p.parseExpr(); !ok {
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
			attributes = append(attributes, attrs)

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

	e := Element{Ident: ident, Attributes: slice.New(attributes), Children: slice.New(children)}
	return e, true
}

func (p *parser) parseAttributes() (AttributeSet, bool) {
	var tag Expr
	var ok bool

	if ch := p.peek(); ch.Kind == token.Hash {
		p.advance()
		if ch := p.peek(); ch.Kind == token.BraceOpen {
			p.addError("missing identifier")
			return nil, false
		}

		if tag, ok = p.parseExpr(); !ok {
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
		attrset = TaggedAttributeSet{Tag: tag, Attributes: slice.New(attrs)}
	} else {
		attrset = UntaggedAttributeSet{Attributes: slice.New(attrs)}
	}

	return attrset, true
}

func (p *parser) parseAttribute() (Attribute, bool) {
	var key Var
	var value Expr
	var ok bool

	if key, ok = p.parseVar(); !ok {
		return Attribute{}, false
	}

	if _, ok = p.expect(token.Colon, "missing attribute value separator"); !ok {
		return Attribute{}, false
	}

	value, ok = p.parseExpr()
	if value == nil || !ok {
		return Attribute{}, false
	}

	if ch := p.peek(); ch.Kind == token.Comma {
		p.advance()
	}

	return Attribute{Key: key, Value: value}, true
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

	case token.String, token.StringTempl:
		p.advance()
		text := Text(ch)
		return Text(text), true

	case token.StringLine,
		token.StringLineTempl:

		p.advance()
		textGroup := TextGroup{}
		textGroup = append(textGroup, Text(ch))

		for !p.eof() {
			ch := p.peek()
			if ch.Kind != token.StringLine && ch.Kind != token.StringLineTempl {
				break
			}

			textGroup = append(textGroup, Text(ch))
			p.advance()
		}

		return textGroup, true

	case token.Ident:
		p.addError("identifier is not a valid template content")

	default:
		p.addError("no content")
	}

	return nil, false
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

		expr, ok = Constant(ch), true

	case token.StringLine,
		token.StringLineTempl:

		p.addError("line string literal is not a valid expression")

	default:
		p.addError("invalid expression")
		return nil, false
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
		expr, ok = p.parseVar()

		// parse member access
		for !p.eof() {
			ch := p.peek()
			if ch.Kind != token.FSlash {
				break
			}

			p.advance()

			var right Var
			if right, ok = p.parseVar(); !ok {
				return nil, false
			}

			expr = MemberAccess{Object: expr, Member: right}
		}

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

func (p *parser) parseVar() (Var, bool) {
	tok := p.peek()
	if tok.Kind != token.Ident {
		p.addError("missing identifier")
		return Var{}, false
	}
	p.advance()
	return Var(tok), true
}

func (p *parser) parseCondExpression() (Expr, bool) {
	var cond Expr
	var options []Case
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

		option := Case{Cond: lit, Branch: value}
		options = append(options, option)
	}

	if _, ok = p.expect(token.ParenClose, "missing closing parenthesis"); !ok {
		return nil, false
	}

	e := CondExpression{Target: cond, Cases: slice.New(options)}
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

	return IfExpression{Cond: cond, Then: thenBranch, Else: elseBranch}, true
}

func (p *parser) expect(k token.Kind, msg errmessage) (token.Token, bool) {
	ch := p.peek()
	if ch.Kind != k {
		p.addError(msg)
		// p.addError(k.String())
		// p.addError(ch.Kind.String())
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

func (p *parser) addError(msg string) {
	tok := p.peek()

	if tok == eof {
		p.dst.Errors = append(p.dst.Errors, errors.Error{Message: errors.Desc(msg)})
		return
	}

	position, ok := p.src.Pos.Item(tok.Pos)
	if !ok {
		panic("Invalid token postion")
	}

	line := -1
	col := -1
	{
		lst := -1
		for i, l := range p.src.Lines.Each() {
			if l > position.Start {
				break
			}
			line = i
			lst = l
		}
		col = lst - position.Start
	}

	err := errors.Error{Line: line, Col: col, Message: errors.Desc(msg)}
	p.dst.Errors = append(p.dst.Errors, err)
}
