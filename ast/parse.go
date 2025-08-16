package ast

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/tel-lang/tel"
	"github.com/tel-lang/tel/ast/token"
)

type parser struct {
	src          *token.File
	mod          *Module
	cur          int
	flag         tel.Flag
	templateKind TemplateKind
}

var (
	eof token.Token = token.Token{Kind: -1, Pos: token.Pos{Offset: -1, Length: -1, Line: -1, Column: -1}}
)

func (p *parser) parse() {
	assert(p != nil, "parser is nil")

	type DeclarationKind int
	const (
		KindType DeclarationKind = iota
		KindUsing
		KindImport
		KindPackage
	)
	var (
		previousDeclaration   = KindPackage
		alreadyParseDocument  = false
		hasPackageDeclaration = false
		exitOnError           = p.flag&tel.ExitOnError != 0
	)

	for !p.eof() {
		var (
			decl, ok                     = p.parseDeclaration()
			currentDeclarationIsDocument = false
			currentDeclaration           DeclarationKind
		)
		if !ok && exitOnError {
			return
		}
		if !ok {
			continue
		}
		switch d := decl.(type) {
		case PackageDecl:
			p.mod.Package = d
			currentDeclaration = KindPackage

		case ImportDecl:
			p.mod.Imports = append(p.mod.Imports, d)
			currentDeclaration = KindImport

		case UsingDecl:
			p.mod.Usings = append(p.mod.Usings, d)
			currentDeclaration = KindUsing

		case TemplateDecl:
			p.mod.Templates = append(p.mod.Templates, d)
			currentDeclarationIsDocument = d.Kind == TemplateDocument
			currentDeclaration = KindType

		case EnumDecl:
			p.mod.Enums = append(p.mod.Enums, d)
			currentDeclaration = KindType

		default:
			p.addError("invalid declaration")
			continue
		}

		if currentDeclaration != KindPackage && !hasPackageDeclaration {
			p.addError("missing package declaration")
		} else if currentDeclaration > previousDeclaration {
			p.addError("unexpected declaration")
		}
		if alreadyParseDocument && currentDeclarationIsDocument {
			p.addError("document already declared")
		}
		if currentDeclaration == KindPackage {
			hasPackageDeclaration = true
		}
		if currentDeclarationIsDocument {
			alreadyParseDocument = true
		}
		previousDeclaration = currentDeclaration
	}

	tel.Assert(p.eof(), "expected eof")
}

func (p *parser) parsePackage() (PackageDecl, bool) {
	p.expect(token.Package)

	ident, ok := p.parseVar()
	if !ok {
		return PackageDecl{}, false
	}

	nameTok, ok := p.expect(token.String)
	if !ok {
		return PackageDecl{}, false
	}

	quotedID := p.mustExtractSourceText(nameTok)
	id := quotedID[1 : len(quotedID)-1]
	if id == "" {
		p.addError("package name is empty")
	}
	return PackageDecl{Var: ident, ID: id}, true
}

func (p *parser) parseImport() (ImportDecl, bool) {
	tel.Assert(p.peek().Kind == token.Import, "expected import keyword")
	// consume import keyword
	p.advance()

	ident, ok := p.parseVar()
	if !ok {
		return ImportDecl{}, false
	}

	path, ok := p.expect(token.String)
	if !ok {
		return ImportDecl{}, false
	}
	pathStr := String(p.mustExtractSourceText(path))
	return ImportDecl{Var: ident, Path: pathStr}, true
}

func (p *parser) parseUsing() (UsingDecl, bool) {
	p.expect(token.Using)

	var idents []Var

	if p.peek().Kind == token.BracketOpen {
		p.advance()
		for p.peek().Kind == token.Ident {
			ident, _ := p.parseVar()
			idents = append(idents, ident)
			if ch := p.peek(); ch.Kind == token.Comma {
				p.advance()
			}
		}
		if _, ok := p.expect(token.BracketClose); !ok {
			return UsingDecl{}, false
		}
		if len(idents) == 0 {
			p.addError("empty import alias list")
			return UsingDecl{}, false
		}

	} else {
		ident, ok := p.parseVar()
		if !ok {
			return UsingDecl{}, false
		}
		idents = append(idents, ident)
	}

	from, ok := p.parseMemberAccess()
	node := UsingDecl{
		Aliases: idents,
		From:    from,
	}
	return node, ok
}

func (p *parser) parseEnumDeclaration() (EnumDecl, bool) {
	p.expect(token.Enum)

	typeIdent, exists := p.parseVar()
	if !exists {
		return EnumDecl{}, false
	}

	p.expect(token.BracketOpen)

	var constants []FieldDecl

	for p.peek().Kind == token.Ident {
		ident, _ := p.parseVar()
		field := FieldDecl{
			Var:  ident,
			Type: typeIdent,
		}
		constants = append(constants, field)
		if p.peek().Kind == token.Comma {
			p.advance()
		}
	}

	if len(constants) == 0 {
		p.addError("empty enum constant")
	}

	p.expect(token.BracketClose)

	enum := EnumDecl{Var: typeIdent, Constants: constants}
	return enum, true
}

func (p *parser) parseTemplateDeclaration(kind TemplateKind) (TemplateDecl, bool) {
	switch kind {
	case TemplateComponent:
		p.expect(token.Component)
	case TemplateDocument:
		p.expect(token.Document)
	default:
		panic(fmt.Sprintf("unexpected ast.TemplateKind: %#v", kind))
	}

	p.templateKind = kind

	// ident is required for component but optional for document
	var ident Var
	if kind == TemplateDocument {
		if p.peek().Kind == token.Ident {
			ident, _ = p.parseVar()
		} else {
			cur := p.peek()
			name := strings.ToUpper(p.src.Filename[0:1]) + p.src.Filename[1:]
			ident = Var{
				Name:      name,
				Line:      cur.Line,
				Col:       cur.Column,
				namespace: p.src.Filename,
			}
		}
	} else {
		i, ok := p.parseVar()
		if !ok {
			return TemplateDecl{}, false
		}
		ident = i
	}

	properties, ok := p.parseProperties()
	if !ok {
		return TemplateDecl{}, false
	}

	var children []Stmt
	for !p.eof() {
		if ch := p.peek(); ch.Kind == token.ParenClose {
			break
		}
		templ, ok := p.parseTemplate()
		if !ok {
			return TemplateDecl{}, false
		}
		children = append(children, templ)
	}

	t := TemplateDecl{Kind: kind, Var: ident, Properties: properties, Stmts: children}
	return t, true
}

func (p *parser) parseProperties() ([]FieldDecl, bool) {
	var props []FieldDecl

	if _, ok := p.expect(token.BracketOpen); !ok {
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

	if _, ok := p.expect(token.BracketClose); !ok {
		return nil, false
	}

	return props, true
}

func (p *parser) parseProperty() (FieldDecl, bool) {
	var ident Var
	var Type Expr
	var ok bool

	if ident, ok = p.parseVar(); !ok {
		return FieldDecl{}, false
	}

	if _, ok = p.expect(token.Colon); !ok {
		return FieldDecl{}, false
	}

	// parse property type
	if Type, ok = p.parsePropertyType(); !ok {
		return FieldDecl{}, false
	}

	prop := FieldDecl{Var: ident, Type: Type}
	return prop, true
}

func (p *parser) parsePropertyType() (Expr, bool) {
	typeid, ok := p.parseMemberAccess()
	if !ok {
		return nil, false
	}
	switch pt := typeid.(type) {
	case Var:
		return pt, true
	case MemberAccess:
		return pt, true
	default:
		panic(fmt.Sprintf("unexpected member access type %v", reflect.TypeOf(typeid)))
	}
}

func (p *parser) parseDeclaration() (Node, bool) {
	var node Node
	var ok bool

	if _, ok := p.expect(token.ParenOpen); !ok {
		p.recover()
		return badNode{}, false
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
		node, ok = p.parseTemplateDeclaration(TemplateDocument)
	case token.Component:
		node, ok = p.parseTemplateDeclaration(TemplateComponent)
	case token.Enum:
		node, ok = p.parseEnumDeclaration()
	case token.Ident:
		node, ok = badNode{}, false

		if _, ok := p.parseElement(true); ok {
			p.addError("unexpected element declaration")
		} else {
			p.addError("invalid declaration")
		}

	case token.String, token.StringLine, token.StringTempl, token.StringLineTempl:
		p.addError("unexpected content")
		p.advance()
		node, ok = badNode{}, false

	default:
		p.addError("invalid declaration")
		node, ok = badNode{}, false
	}

	if !ok {
		return node, false
	}

	if _, ok = p.expect(token.ParenClose); !ok {
		return badNode{}, false
	}

	return node, ok
}

func (p *parser) parseCondElement() (Cond, bool) {
	var (
		cond    Expr
		options []Case
		ok      bool
	)

	p.expect(token.ParenOpen)
	p.expect(token.Cond)

	if cond, ok = p.parseExpr(); !ok {
		return Cond{}, false
	}

	for !p.eof() {
		if p.peek().Kind == token.ParenClose {
			break
		}

		cond, ok := p.parseExpr()
		if !ok {
			return Cond{}, false
		}

		p.expect(token.Colon)

		content, ok := p.parseTemplate()
		if !ok {
			return Cond{}, false
		}
		if p.peek().Kind == token.Comma {
			p.advance()
		}

		option := Case{
			Cond:   cond,
			Branch: content,
		}

		options = append(options, option)
	}

	if len(options) == 0 {
		p.addError("no case statement")
	}

	p.expect(token.ParenClose)

	e := Cond{
		Target: cond,
		Cases:  options,
	}

	return e, true
}

func (p *parser) parseIfElement() (If, bool) {
	var (
		cond       Expr
		thenBranch Stmt
		elseBranch Stmt
		ok         bool
	)

	if _, ok = p.expect(token.ParenOpen); !ok {
		return If{}, false
	}
	if _, ok = p.expect(token.If); !ok {
		return If{}, false
	}
	if cond, ok = p.parseExpr(); !ok {
		p.addError("invalid/missing condition expression")
		return If{}, false
	}
	if thenBranch, ok = p.parseTemplate(); !ok {
		p.addError("invalid then-branch")
		return If{}, false
	}

	if ch := p.peek(); ch.Kind != token.ParenClose {
		elseBranch, ok = p.parseTemplate()
		if !ok {
			return If{}, false
		}
	}

	if _, ok = p.expect(token.ParenClose); !ok {
		return If{}, false
	}

	return If{Cond: cond, Then: thenBranch, Else: elseBranch}, true
}

func (p *parser) parseElement(skipParenOpen bool) (Element, bool) {
	var ident Expr
	var attributes []Attr
	var parameters []ParameterDecl
	var children []Stmt
	var ok bool

	if !skipParenOpen {
		if _, ok = p.expect(token.ParenOpen); !ok {
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

		case token.BracketOpen:
			if parameters != nil {
				// TODO add test case
				p.addError("parameter already defined")
			}
			params, ok := p.parseElementParameters()
			if !ok {
				return Element{}, false
			}
			parameters = params

		case token.BraceOpen:
			attrs, ok := p.parseAttr()
			if !ok {
				return Element{}, false
			}
			attributes = append(attributes, attrs)

		case token.Hash:
			directive, directiveOk := p.parseDirective()
			attr, attrOk := p.parseAttr()
			if !directiveOk || !attrOk {
				return Element{}, false
			}
			attr.Directive = directive
			attributes = append(attributes, attr)

		default:
			templ, ok := p.parseTemplate()
			if !ok {
				return Element{}, false
			}
			children = append(children, templ)
		}
	}

	if _, ok = p.expect(token.ParenClose); !ok {
		return Element{}, false
	}

	if p.templateKind == TemplateComponent {
		for _, attr := range attributes {
			switch attr.Directive.(type) {
			case MemberAccess:
				p.addError("qualified tagged attributes not allowed in a component")
				return Element{}, false
			}
		}
	}

	e := Element{Tag: ident, Parameter: parameters, Attributes: attributes, Children: children}
	return e, true
}

func (p *parser) parseElementParameters() ([]ParameterDecl, bool) {
	tel.Assert(p.peek().Kind == token.BracketOpen, "expected [")

	pos := p.peek().Pos
	// consume [
	p.advance()

	parameters := []ParameterDecl{}
	for !p.eof() {
		ch := p.peek()
		if ch.Kind == token.BracketClose {
			break
		}
		if ch.Kind == token.ParenClose {
			break
		}

		var name Var
		var value Expr
		var ok bool

		name, ok = p.parseVar()
		if !ok {
			return []ParameterDecl{}, false
		}

		p.expect(token.Colon)

		value, ok = p.parseExpr()
		if !ok {
			return []ParameterDecl{}, false
		}

		if ch := p.peek(); ch.Kind == token.Comma {
			p.advance()
		}

		param := ParameterDecl{Var: Var(name), Value: value}
		parameters = append(parameters, param)
	}

	if ch := p.peek(); ch.Kind != token.BracketClose {
		cur := p.cur
		// TODO add a test case to check that the reported offset points to the position of [
		// in the element definition
		p.cur = pos.Offset
		p.addError("unterminated element parameters")
		p.cur = cur
		return nil, false
	}

	// consume ]
	p.advance()

	return parameters, true
}

func (p *parser) parseDirective() (Expr, bool) {
	p.expect(token.Hash)
	if p.peek().Kind == token.BraceOpen {
		p.addError("missing identifier")
		return nil, false
	}
	expr, ok := p.parseExpr()
	if !ok {
		return nil, false
	}
	return expr, true
}

func (p *parser) parseAttr() (Attr, bool) {
	var ok bool

	if _, ok = p.expect(token.BraceOpen); !ok {
		return Attr{}, false
	}

	var attrs []KeyVal
	for !p.eof() {
		if ch := p.peek(); ch.Kind != token.Ident {
			break
		}

		attr, ok := p.parseKeyVal()
		if !ok {
			return Attr{}, false
		}

		attrs = append(attrs, attr)
	}

	if _, ok = p.expect(token.BraceClose); !ok {
		return Attr{}, false
	}

	attrset := Attr{Entries: attrs}
	return attrset, true
}

func (p *parser) parseKeyVal() (KeyVal, bool) {
	var (
		key   Var
		value Expr
		ok    bool
	)

	if key, ok = p.parseVar(); !ok {
		return KeyVal{}, false
	}

	if _, ok = p.expect(token.Colon); !ok {
		return KeyVal{}, false
	}

	value, ok = p.parseExpr()
	if value == nil || !ok {
		return KeyVal{}, false
	}

	if ch := p.peek(); ch.Kind == token.Comma {
		p.advance()
	}

	return KeyVal{Key: key, Value: value}, true
}

func (p *parser) parseTemplate() (Stmt, bool) {
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
		var kind TextKind
		if ch.Kind == token.String {
			kind = QuotedText
		} else {
			kind = QuotedTemplateText
		}
		text := Text{Kind: kind, Value: p.mustExtractSourceText(ch)}
		return Text(text), true

	case token.StringLine,
		token.StringLineTempl:
		p.advance()

		var kind TextKind
		if ch.Kind == token.StringLine {
			kind = LineText
		} else {
			kind = LineTemplateText
		}

		lines := []Text{}

		for !p.eof() {
			ch := p.peek()
			if ch.Kind != token.StringLine && ch.Kind != token.StringLineTempl {
				break
			}

			text := Text{Kind: kind, Value: p.mustExtractSourceText(ch)}
			lines = append(lines, text)
			p.advance()
		}

		return TextGroup{lines}, true

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
	case token.True:
		expr, ok = True, true

	case token.False:
		expr, ok = False, true

	case token.Number:
		text := p.mustExtractSourceText(ch)
		// TODO support decimal numbers
		num, err := strconv.Atoi(text)
		if err != nil {
			p.addError(fmt.Sprintf("invalid number: %v", text))
		}
		expr, ok = Number(num), true

	case token.String, token.StringTempl:
		text := p.mustExtractSourceText(ch)
		expr, ok = String(text), true

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
	switch ch := p.peek(); ch.Kind {
	case token.ParenClose,
		token.BraceClose:
		p.addError("missing expression")
		return nil, false

	case token.Ident:
		return p.parseMemberAccess()

	case token.ParenOpen:
		switch ch := p.peekNext(); ch.Kind {
		case token.If:
			return p.parseIfExpression()
		case token.Cond:
			return p.parseCondExpression()
		default:
			panic(fmt.Sprintf("unexpected token %v", ch.Kind))
		}
	default:
		return p.parseLiteral()
	}
}

func (p *parser) parseMemberAccess() (Expr, bool) {
	var (
		left Expr
		ok   bool
	)
	left, ok = p.parseVar()
	if !ok {
		return nil, false
	}
	for p.peek().Kind == token.Dot {
		p.advance()
		right, ok := p.parseVar()
		if !ok {
			return nil, false
		}
		left = MemberAccess{
			Object: left,
			Member: right,
		}
	}
	return left, true
}

func (p *parser) parseVar() (Var, bool) {
	tok, ok := p.expect(token.Ident)
	if !ok {
		return Var{}, false
	}

	text := p.mustExtractSourceText(tok)
	line, col := tok.Line, tok.Column
	v := Var{
		Name: text,
		Line: line,
		Col:  col,
	}
	return v, true
}

func (p *parser) parseCondExpression() (Expr, bool) {
	var cond Expr
	var options []CaseExpr
	var ok bool

	if _, ok = p.expect(token.ParenOpen); !ok {
		return nil, false
	}

	if _, ok = p.expect(token.Cond); !ok {
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

		p.expect(token.Colon)

		value, ok := p.parseExpr()
		if !ok {
			return nil, false
		}

		if ch := p.peek(); ch.Kind == token.Comma {
			p.advance()
		}

		option := CaseExpr{Cond: lit, Branch: value}
		options = append(options, option)
	}

	if _, ok = p.expect(token.ParenClose); !ok {
		return nil, false
	}

	e := CondExpr{Target: cond, Cases: options}
	return e, true
}

func (p *parser) parseIfExpression() (Expr, bool) {
	var cond Expr
	var thenBranch Expr
	var elseBranch Expr
	var ok bool

	if _, ok = p.expect(token.ParenOpen); !ok {
		return nil, false
	}

	if _, ok = p.expect(token.If); !ok {
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

	if _, ok := p.expect(token.ParenClose); !ok {
		return nil, false
	}

	return IfExpr{Cond: cond, Then: thenBranch, Else: elseBranch}, true
}

func (p *parser) expect(k token.Kind) (token.Token, bool) {
	ch := p.peek()
	if ch.Kind != k {
		p.addError(fmt.Sprintf("expected %v got %v", k, ch.Kind))
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
		if next >= len(p.src.Tokens) {
			break
		}
		ch := p.src.Tokens[next]
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

	if next >= len(p.src.Tokens) {
		return eof
	}
	tok := p.src.Tokens[next]

	return tok
}

func (p *parser) peek() token.Token {
	if p.eof() {
		return eof
	}

	next := p.cur
	if next >= len(p.src.Tokens) {
		return eof
	}
	node := p.src.Tokens[next]

	return node
}

func (p *parser) eof() bool {
	e := p.cur >= p.src.Size()
	return e
}

func (p *parser) mustExtractSourceText(tok token.Token) string {
	text, ok := p.src.Text(tok)
	if !ok {
		panic(fmt.Sprintf("unable to extract text for: %v", tok))
	}
	return text
}

func (p *parser) addError(msg string) {
	tok := p.peek()

	if tok == eof {
		p.mod.error(errdescString(msg))
		return
	}

	line, col := tok.Line, tok.Column
	err := asterror{Message: msg, Line: line, Col: col}
	p.mod.error(errdesc(err))
}
