package ast

import (
	"fmt"
	"reflect"

	"github.com/eml-lang/teml/ast"
	"github.com/eml-lang/teml/token"
)

type parser struct {
	src  *ast.File
	toks *token.File
}

func (p *parser) parsePackage() Package {
	pkg := p.src.Package

	ident := p.parseVar(pkg.Ident)
	path := p.text(pkg.Path)

	node := Package{Ident: ident, Path: String(path)}
	return node
}

func (p *parser) parseImports() []Import {
	if len(p.src.Imports) == 0 {
		return nil
	}

	nodes := make([]Import, 0, len(p.src.Imports))
	for _, imp := range p.src.Imports {
		ident := p.parseVar(imp.Ident)
		path := p.text(token.Token(imp.Path))
		node := Import{Ident: ident, Path: String(path)}
		nodes = append(nodes, node)
	}

	return nodes
}

func (p *parser) parseUsings() []Using {
	if len(p.src.Usings) == 0 {
		return nil
	}

	nodes := make([]Using, 0, len(p.src.Usings))
	for _, use := range p.src.Usings {
		node := p.parseUsing(use)
		nodes = append(nodes, node)
	}

	return nodes
}

func (p *parser) parseUsing(use ast.Using) Using {
	from := p.parseVar(use.From)
	idents := make([]Var, 0, len(use.Idents))

	for _, id := range use.Idents {
		ident := p.parseVar(id)
		idents = append(idents, ident)
	}

	node := Using{From: from, Idents: idents}
	return node
}

func (p *parser) parseDeclarations() []Declaration {
	size := len(p.src.Components)

	if p.src.HasDocument() {
		size += 1
	}

	tmpls := make([]Declaration, 0, size)

	if p.src.HasDocument() {
		node := p.parseDeclaration(p.src.Document)
		tmpls = append(tmpls, node)
	}

	for _, cmp := range p.src.Components {
		node := p.parseDeclaration(cmp)
		tmpls = append(tmpls, node)
	}

	return tmpls
}

func (p *parser) parseDeclaration(n ast.Node) Declaration {
	switch t := n.(type) {
	case ast.Document:
		node := p.parseDocument(t)
		return node
	case ast.Component:
		node := p.parseComponent(t)
		return node
	default:
		panic(fmt.Sprintf("unexpected ast node type: %v", reflect.TypeOf(t)))
	}
}

func (p *parser) parseDocument(d ast.Document) Document {
	var ident Var
	if d.IsNamed() {
		ident = p.parseVar(d.Ident)
	}

	// TODO use the file name if document is not named

	props := p.parseProperties(d.Properties)
	stmts := p.parseStmts(d.Children)
	node := Document{Ident: ident, Properties: props, Stmts: stmts}
	return node
}

func (p *parser) parseComponent(cmp ast.Component) Declaration {
	ident := p.parseVar(cmp.Ident)
	props := p.parseProperties(cmp.Properties)
	stmts := p.parseStmts(cmp.Children)
	node := Component{Ident: ident, Properties: props, Stmts: stmts}
	return node
}

func (p *parser) parseProperties(params []ast.Property) []Property {
	if len(params) == 0 {
		return nil
	}

	nodes := make([]Property, 0, len(params))
	for _, param := range params {
		ident := p.parseVar(param.Ident)
		typevar := p.parsePropertyType(param.Type)
		node := Property{Ident: ident, Type: typevar}
		nodes = append(nodes, node)
	}
	return nodes
}

func (p *parser) parsePropertyType(t ast.PropertyType) PropertyType {
	switch tt := t.(type) {
	case ast.Var:
		return p.parseVar(tt)
	case ast.MemberAccess:
		return p.parseMemberAccess(tt)
	case ast.Enum:
		return p.parseEnum(tt)
	}
	panic(fmt.Sprintf("unexpected property type: %v", reflect.TypeOf(t)))
}

func (p *parser) parseStmts(ch []ast.Content) []Stmt {
	if len(ch) == 0 {
		return nil
	}

	stmts := make([]Stmt, 0, len(ch))

	for _, c := range ch {
		expr := p.parseExprStmt(c)
		node := Stmt{expr}
		stmts = append(stmts, node)
	}

	return stmts
}

func (p *parser) parseExprStmt(c ast.Content) Element {
	switch t := c.(type) {
	case ast.Text:
		txt := p.text(token.Token(t))
		return TextElement{String(txt)}

	case ast.TextGroup:
		var tg ast.TextGroup = t

		lines := make([]String, 0, len(tg))
		for _, t := range tg {
			line := p.text(token.Token(t))
			// strip line string marker: --
			line = line[2:]
			lines = append(lines, String(line))
		}
		node := TextGroupElement{lines}
		return node

	case ast.Element:
		ident := p.parseIdentifier(t.Ident)
		props := p.parseElementProperties(t.Parameter)
		attrs := p.parseGenericAttributes(t.Attributes)
		block := p.parseBlock(t.Children)
		node := genericElement{Tag: ident, Parameter: props, Attributes: attrs, Body: block}
		return node

	case ast.IfElement:
		cond := p.parseExpr(t.Cond)
		then := p.parseExprStmt(t.Then)
		var els Element
		if t.Else != nil {
			els = p.parseExprStmt(t.Else)
		}
		return IFElement{Cond: cond, Then: then, Else: els}

	case ast.CondElement:
		target := p.parseExpr(t.Target)
		cases := make([]CaseStmt, 0, len(t.Cases))
		for _, c := range t.Cases {
			cond := p.parseExpr(c.Cond)
			stmt := p.parseExprStmt(c.Branch)
			branch := Stmt{Element: stmt}
			node := CaseStmt{Cond: cond, Branch: branch}
			cases = append(cases, node)
		}
		return CondElement{Target: target, Cases: cases}

	}
	panic(fmt.Sprintf("unexpected content: %v", reflect.TypeOf(c)))
}

func (p *parser) parseBlock(block []ast.Content) []Stmt {
	if len(block) == 0 {
		return nil
	}
	nodes := make([]Stmt, 0, len(block))
	for _, item := range block {
		expr := p.parseExprStmt(item)
		node := Stmt{Element: expr}
		nodes = append(nodes, node)
	}
	return nodes
}

func (p *parser) parseElementProperties(props []ast.ElementParameter) []KeyVal {
	if len(props) == 0 {
		return nil
	}
	nodes := make([]KeyVal, 0, len(props))
	for _, prop := range props {
		name := p.parseVar(prop.Ident)
		value := p.parseExpr(prop.Value)
		node := KeyVal{Key: name, Value: value}
		nodes = append(nodes, node)
	}
	return nodes
}

func (p *parser) parseGenericAttributes(attrs []ast.AttributeSet) []Attr {
	if len(attrs) == 0 {
		return nil
	}
	nodes := make([]Attr, 0, len(attrs))
	for _, attr := range attrs {
		switch attrtype := attr.(type) {
		case ast.TaggedAttributeSet:
			tag := p.parseExpr(attrtype.Tag)
			entries := p.parseAttributes(attrtype.Attributes)
			node := Attr{Tag: tag, Entries: entries}
			nodes = append(nodes, node)

		case ast.UntaggedAttributeSet:
			entries := p.parseAttributes(attrtype.Attributes)
			node := Attr{Entries: entries}
			nodes = append(nodes, node)

		default:
			panic(fmt.Sprintf("unexpected attribute set: %v", reflect.TypeOf(attr)))
		}
	}
	return nodes
}

func (p *parser) parseAttributes(attrs []ast.Attribute) []KeyVal {
	if len(attrs) == 0 {
		return nil
	}
	nodes := make([]KeyVal, 0, len(attrs))
	for _, attr := range attrs {
		key := p.parseVar(attr.Key)
		val := p.parseExpr(attr.Value)
		node := KeyVal{Key: key, Value: val}
		nodes = append(nodes, node)
	}
	return nodes
}

func (p *parser) parseExpr(e ast.Expr) Expr {
	switch t := e.(type) {
	case ast.Var:
		ident := p.parseVar(t)
		return ident

	case ast.Constant:
		switch t.Kind {
		case token.Ident:
			ident := p.parseVarFrom(token.Token(t))
			return ident

		case token.True, token.False:
			txt := p.text(token.Token(t))
			expr := Bool(txt)
			return expr

		case token.Number:
			num := p.text(token.Token(t))
			expr := Number(num)
			return expr

		case token.String:
			txt := p.text(token.Token(t))
			expr := String(txt)
			return expr
		}

	case ast.MemberAccess:
		expr := p.parseMemberAccess(t)
		return expr

	case ast.IfExpression:
		cond := p.parseExpr(t.Cond)
		then := p.parseExpr(t.Then)
		// FIX t.Else could be nil
		els := p.parseExpr(t.Else)
		expr := IFExpr{Cond: cond, Then: then, Else: els}
		return expr

	case ast.CondExpression:
		target := p.parseExpr(t.Target)
		cases := make([]CaseExpr, 0, len(t.Cases))
		for _, c := range t.Cases {
			cond := p.parseExpr(c.Cond)
			branch := p.parseExpr(c.Branch)
			node := CaseExpr{Cond: cond, Branch: branch}
			cases = append(cases, node)
		}
		expr := CondExpr{Target: target, Cases: cases}
		return expr

	case ast.Enum:
		return p.parseEnum(t)
	}

	panic(fmt.Sprintf("unexpected expression type: %v", reflect.TypeOf(e)))
}

func (p *parser) parseMemberAccess(m ast.MemberAccess) MemberAccess {
	object := p.parseIdentifier(m.Object)
	member := p.parseVar(m.Member)
	node := MemberAccess{Object: object, Member: member}
	return node
}

func (p *parser) parseIdentifier(e ast.Expr) Expr {
	switch t := e.(type) {
	case ast.Var:
		switch t.Kind {
		case token.Ident:
			ident := p.parseVarFrom(token.Token(t))
			return ident
		}
	case ast.MemberAccess:
		ident := p.parseMemberAccess(t)
		return ident
	}
	panic(fmt.Sprintf("invalid identifier expression: %v", reflect.TypeOf(e)))
}

func (p *parser) parseEnum(e ast.Enum) Enum {
	if len(e.Constants) == 0 {
		return Enum{}
	}

	constants := make([]EnumConstant, 0, len(e.Constants))
	for _, c := range e.Constants {
		constant := EnumConstant(p.parseExpr(c))
		constants = append(constants, constant)
	}

	node := Enum{Constants: constants}
	return node
}

func (p *parser) parseVarFrom(e token.Token) Var {
	// hack
	v := ast.Var(e)
	return p.parseVar(v)
}

func (p *parser) parseVar(tok ast.Var) Var {
	var pos token.Pos

	txt := p.text(token.Token(tok))
	if int(tok.Pos) >= len(p.toks.Pos) || int(tok.Pos) < 0 {
		panic(fmt.Sprintf("could not find tok position in source file: %v", tok))
	}
	pos = p.toks.Pos[tok.Pos]

	// TODO find a better way to store line & col number
	line, col := p.toks.Line(token.Token(tok))

	return Var{Name: txt, Pos: Pos{Start: pos.Start, End: pos.End, Line: line, Col: col}}
}

func (p *parser) text(tok token.Token) string {
	txt, ok := p.toks.Text(tok)
	if !ok {
		panic(fmt.Sprintf("text not found for: token %v", tok))
	}
	return txt
}
