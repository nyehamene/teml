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
}

var (
	tokenEOF pToken = pToken{Kind: -1, Pos: -1}
)

func Parse(toks token.Tokenized) (*File, bool) {
	printerr := func(s string) {
		log.Println(s)
	}
	return ParseWithErrorHandler(toks, printerr)
}

func ParseWithErrorHandler(toks token.Tokenized, printerr func(string)) (*File, bool) {
	f := &File{}
	ok := parse(toks, f, printerr)
	return f, ok

}

func parse(toks token.Tokenized, f *File, printerr func(string)) (hasError bool) {
	p := parser{
		src:  toks,
		file: f,
	}

	p.printerr = func(s string) {
		p.hasError = true
		printerr(s)
	}

	defer func() {
		hasError = p.hasError
	}()

	p.parsePackage()
	p.parseImport()
	p.parseDeclarations()

	assert.Assert(p.eof(), "expected eof")

	return
}

func (p *parser) parsePackage() {
	p.expect(token.ParanOpen, "invalid delimiter")
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

		if ch.Kind != token.ParanOpen ||
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

		if ch.Kind != token.ParanOpen ||
			next.Kind != token.Using {
			break
		}

		p.advance()
		p.advance()

		// TODO respect ReduceAlloc flag
		idents := []token.Token{}

		ch = p.peek()

		if ch.Kind == token.BracketOpen {

			p.advance()

			for !p.eof() {

				id := p.peek()
				if id.Kind == token.BracketClose {
					break
				}

				ident := p.expect(token.Ident, "missing identifier")
				idents = append(idents, ident)
			}

			p.expect(token.BracketClose, "missing closing bracket ']'")

		} else {
			ident := p.expect(token.Ident, errfmt.title("Using Declaration Error"), errfmt.desc("missing identifier"))
			idents = append(idents, ident)
		}

		from := p.expect(token.Ident, errfmt.title("Using Declaration Error"), errfmt.desc("missing import identifier"))

		p.expect(token.ParenClose, errfmt.title("Using Declaration Error"), errfmt.desc("missing closing parenthesis ')'"))

		use := Using{From: from}
		for _, ident := range idents {
			use.idents = append(use.idents, ident)

		}

		p.file.usings = append(p.file.usings, use)
	}
}

func (p *parser) parseComponent() {
	assert.Assert(p.peek().Kind == token.Component, "expected keyword 'component'")
	p.advance()

	ident := p.expect(token.Ident, "missing component identifier '('")
	cmp := Component{Ident: ident}

	p.expect(token.BracketOpen, "missing opening square bracket '['")

	p.parseProperties(&cmp)

	p.expect(token.BracketClose, "missing closing square bracket ']'")

	p.file.components = append(p.file.components, cmp)
}

func (p *parser) parseProperties(cmp *Component) {

	for !p.eof() {

		if ch := p.peek(); ch.Kind == token.BracketClose {
			break
		}

		ident := p.expect(token.Ident, "missing property identifier")
		p.expect(token.Colon, "missing type separator ':'")
		Type := p.expect(token.Ident, "missing property type")

		pty := Property{Ident: ident, Type: Type}
		cmp.properties = append(cmp.properties, pty)

		if ch := p.peek(); ch.Kind == token.Comma {
			p.advance()
		}
	}

}

func (p *parser) parseDeclarations() {
	for !p.eof() {
		p.expect(token.ParanOpen, "missing opening parenthesis '('")

		ch := p.peek()

		switch ch.Kind {
		case token.Component:
			p.parseComponent()
		default:
			// TODO error
			return
		}

		p.expect(token.ParenClose, "missing closing parenthesis ')'")
	}
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
