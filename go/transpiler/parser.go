package transpiler

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/tel-lang/tel/ast"
)

type parser struct{}

func (p *parser) parseImport(i ast.ImportDecl) Import {
	name := i.Name
	path := i.Path
	node := Import{Name: Var(name), Path: String(path)}
	return node
}

// parseUsing returns a type alias named name for a type in the package pkgname
func (p *parser) parseUsing(u ast.UsingDecl) func(func(TypeAlias) bool) {
	return func(yield func(TypeAlias) bool) {
		pkgname := p.getQualifiedName(u.From)
		for _, ident := range u.Aliases {
			aliasname := ident.Name
			typename := string(pkgname) + "." + aliasname

			alias := TypeAlias{
				Name: Var(aliasname),
				Type: Var(typename),
			}
			if !yield(alias) {
				break
			}
		}
	}
}

func (p *parser) parseStruct(decl ast.TemplateDecl) Struct {
	name := decl.Name
	props := decl.Properties

	fields := make([]StructField, 0, len(props))
	for _, prop := range props {
		field := p.parseStructField(name, prop)
		fields = append(fields, field)
	}
	node := Struct{
		Name:   Var(name),
		Fields: fields,
	}
	return node
}

func (p *parser) parseStructField(structname string, prop ast.FieldDecl) StructField {
	name := prop.Name
	type0 := p.parseFieldType(structname, name, prop.Type)
	node := StructField{
		Name: Var(name),
		Type: type0,
	}
	return node
}

type typeinfo struct {
	name string
}

func (m typeinfo) Receiver() Var {
	rcv := strings.ToLower(m.name[0:1])
	return Var(rcv)
}

func (m typeinfo) getMember(name Var) Var {
	member := m.Receiver() + "." + name
	return member
}

func (p *parser) parseMethod(decl ast.TemplateDecl) RenderMethod {
	typename := decl.Name
	stmts := decl.Stmts
	t := typeinfo{name: typename}

	// + the return statement at the end of the function
	body := make([]Stmt, 0, len(stmts)+1)
	for _, stmt := range stmts {
		p.parseStmt(t, stmt, &body)
	}

	childContentVar := makeTempVar()
	body = append(body, InheritChildren{Context: childContentVar})

	// add return statement
	body = append(body, ReturnNil{})
	node := RenderMethod{
		Name:     NameComponentRenderMethod,
		Type:     Var(typename),
		Receiver: t.Receiver(),
		Body:     body,
	}
	return node
}

func (p *parser) parseStmt(m typeinfo, node ast.Stmt, stmts *[]Stmt) {
	errvar := makeErrVar()

	switch elem := node.(type) {
	case ast.Text:
		stmt1 := StringLiteral{Value: String(elem.Value), Error: errvar}
		ret1 := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt1, ret1)

	case ast.TextGroup:
		// TODO join lines separated by a space " "
		text := join(elem, "")
		stmt2 := StringLiteral{Value: String(doubleQuoteString(text)), Error: errvar}
		ret2 := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt2, ret2)

	case ast.NumberElement:
		member := p.getQualifiedName(elem.Tag)
		memberAccess := m.getMember(member)
		tempvar := makeTempVar()
		stmt := NumberMemberAccessExpr{Value: memberAccess, Variable: tempvar, Error: errvar}
		ret := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt, ret)

	case ast.StringElement:
		member := p.getQualifiedName(elem.Tag)
		memberAccess := m.getMember(member)
		stmt := StringMemberAccessExpr{Value: memberAccess, Error: errvar}
		ret := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt, ret)

	case ast.HTMLElement:
		tag := p.getQualifiedName(elem.Tag)
		p.parseOpenTagWithAttributes(tag, elem.Attributes, stmts)

		for _, stmt := range elem.Children {
			p.parseStmt(m, stmt, stmts)
		}

		p.parseCloseTag(tag, stmts)

	case ast.PropertyElement:
		member := p.getQualifiedName(elem.Tag)
		p.parseCallRenderMethod(m, m.getMember(member), elem.Attributes, elem.Children, stmts)

	case ast.ComponentElement:
		cmptype := p.getQualifiedName(elem.Tag)
		cmpvar := makeTempVar()
		parameters := p.parseInstanceParameters(cmpvar, elem.Parameters)
		stmt1 := StructInstance{Type: cmptype, Variable: cmpvar, Parameters: parameters}
		*stmts = append(*stmts, stmt1)
		p.parseCallRenderMethod(m, cmpvar, elem.Attributes, nil, stmts)

	case ast.If:
		cond := p.parseIfCond(m, elem.Cond)

		// then branch
		thenBranch := []Stmt{}
		p.parseStmt(m, elem.Then, &thenBranch)

		// else branch
		elseBranch := []Stmt{}
		if elem.Else != nil {
			p.parseStmt(m, elem.Else, &elseBranch)
		}

		stmt1 := If{Cond: cond, Then: thenBranch, Else: elseBranch}
		*stmts = append(*stmts, stmt1)

	case ast.Cond:
		target := p.parseCondTarget(m, elem.Target)

		// cases
		cases := []Case{}
		for _, c := range elem.Cases {
			branch := []Stmt{}
			cond := p.parseCondCase(c.Cond)
			p.parseStmt(m, c.Branch, &branch)
			cases = append(cases, Case{Match: cond, Branch: branch})
		}

		stmt1 := Cond{Target: target, Cases: cases}
		*stmts = append(*stmts, stmt1)

	default:
		panic(fmt.Sprintf("unexpected element type: %v", reflect.TypeOf(node)))
	}
}

func (p *parser) parseCallRenderMethod(m typeinfo, receiver Var, attrs []ast.Attr, body []ast.Stmt, stmts *[]Stmt) {
	mapvar := makeTempVar()
	entries := p.parseMapEntries(mapvar, attrs)
	stmt1 := MapInstance{Variable: mapvar, Entries: entries}

	slicevar := makeTempVar()
	cmpvars := p.parseChildren(m, body, stmts)
	stmt2 := SliceInstance{Variable: slicevar, Values: cmpvars, Type: NameComponentInterface}

	ctxvar := makeTempVar()
	stmt3 := CopyRenderContext{Variable: ctxvar, Attrs: mapvar, Children: slicevar}

	errvar := makeErrVar()
	stmt4 := CallRenderMethod{Context: ctxvar, Name: NameComponentRenderMethod, Receiver: receiver, Error: errvar}
	ret4 := ReturnIfNotNil(errvar)
	// *stmts = append(*stmts, stmt1, stmt2, stmt3, stmt4, ret4, BlankVar(slicevar), BlankVar(mapvar))
	*stmts = append(*stmts, stmt1, stmt2, stmt3, stmt4, ret4)
}

func (p *parser) parseChildren(m typeinfo, body []ast.Stmt, stmts *[]Stmt) []Var {
	if len(body) == 0 {
		return nil
	}

	cmpvars := make([]Var, 0, len(body))
	for _, stmt := range body {
		cmpvar := makeTempVar()
		cmpvars = append(cmpvars, cmpvar)

		body := []Stmt{}
		p.parseStmt(m, stmt, &body)
		body = append(body, ReturnNil{})
		stmt := ComponentInstance{Variable: cmpvar, Stmts: body}
		*stmts = append(*stmts, stmt)
	}
	return cmpvars
}

func (p *parser) parseInstanceParameters(structvar Var, params []ast.ParameterDecl) []SetStructField {
	if len(params) == 0 {
		return nil
	}

	fields := make([]SetStructField, 0, len(params))
	for _, param := range params {
		field := SetStructField{
			Struct: structvar,
			Name:   Var(param.Name),
			Value:  p.parseValue(param.Value),
		}
		fields = append(fields, field)
	}
	return fields
}

func (p *parser) parseMapEntries(mapvar Var, attrs []ast.Attr) []SetMapEntry {
	if len(attrs) == 0 {
		return nil
	}

	entries := make([]SetMapEntry, 0, len(attrs))
	for _, attr := range attrs {
		for _, entry := range attr.Entries {
			key := doubleQuoteString(entry.Key.Name)
			value := p.parseAttrValue(entry.Value)
			value = doubleQuoteString(ResolveValue(value))
			entry := SetMapEntry{Name: mapvar, Key: Var(key), Value: value}
			entries = append(entries, entry)
		}
	}
	return entries
}

func (p *parser) parseOpenTagWithAttributes(name Var, attrs []ast.Attr, stmts *[]Stmt) {
	// begin open tag
	errvar := makeErrVar()
	stmt1 := StringLiteral{Value: doubleQuoteString(fmt.Sprintf("<%s", name)), Error: errvar}
	ret1 := ReturnIfNotNil(errvar)
	*stmts = append(*stmts, stmt1, ret1)

	errvar = makeErrVar()
	stmti := InheritAttributes{Error: errvar}
	*stmts = append(*stmts, stmti)

	// attribues
	for _, attr := range attrs {
		p.parseAttrs(attr.Entries, stmts)
	}

	// end open tag
	errvar = makeErrVar()
	stmt2 := StringLiteral{Value: doubleQuoteString(">"), Error: errvar}
	ret2 := ReturnIfNotNil(errvar)
	*stmts = append(*stmts, stmt2, ret2)
}

func (p *parser) parseCloseTag(name Var, stmts *[]Stmt) {
	errvar := makeErrVar()
	stmt := StringLiteral{Value: doubleQuoteString(fmt.Sprintf("</%s>", name)), Error: errvar}
	ret := ReturnIfNotNil(errvar)
	*stmts = append(*stmts, stmt, ret)
}

func (p *parser) parseAttrs(entries []ast.KeyVal, stmts *[]Stmt) {
	for _, entry := range entries {
		p.parseAttr(entry, stmts)
	}
}

func (p *parser) parseAttr(node ast.KeyVal, stmts *[]Stmt) {
	key := Var(node.Key.Name)
	value := p.parseAttrValue(node.Value)

	attr := " " + ResolveValue(key) + "=" + ResolveValue(value)
	attrstr := doubleQuoteString(attr)

	errvar := makeErrVar()
	stmt := StringLiteral{Value: attrstr, Error: errvar}
	ret := ReturnIfNotNil(errvar)
	*stmts = append(*stmts, stmt, ret)
}

func (p *parser) parseType(node ast.Var) Type {
	switch node.Name {
	case "String":
		return Var("string")
	case "Number":
		return Var("int")
	case "Bool":
		return Var("bool")
	}
	return Var(node.Name)
}

func (p *parser) parseFieldType(structname, field string, expr ast.Expr) Type {
	switch t := expr.(type) {
	case ast.Var:
		return p.parseType(t)

	case ast.MemberAccess:
		panic(errors.ErrUnsupported)

	default:
		panic(fmt.Sprintf("unexpected propety type: %v", reflect.TypeOf(expr)))
	}
}

func (p *parser) parseValue(expr ast.Expr) Expr {
	switch t := expr.(type) {
	case ast.Var:
		return Var(t.Name)
	case ast.String:
		return String(t)
	case ast.Number:
		return Number(t)
	case ast.Bool:
		return Bool(t)
	case ast.MemberAccess:
		panic(errors.ErrUnsupported)
	case ast.IfExpr:
		panic(errors.ErrUnsupported)
	case ast.CondExpr:
		panic(errors.ErrUnsupported)
	default:
		panic(errors.ErrUnsupported)
	}
}

func (p *parser) parseAttrValue(expr ast.Expr) Expr {
	switch t := expr.(type) {
	case ast.Var:
		return Var(t.Name)
	case ast.String:
		return escapeSurrounding(t)
	case ast.Number:
		return Number(t)
	case ast.Bool:
		return Bool(t)
	case ast.MemberAccess:
		panic(errors.ErrUnsupported)
	case ast.IfExpr:
		panic(errors.ErrUnsupported)
	case ast.CondExpr:
		panic(errors.ErrUnsupported)
	default:
		panic(errors.ErrUnsupported)
	}
}

func (p *parser) parseIfCond(m typeinfo, expr ast.Expr) Expr {
	switch t := expr.(type) {
	case ast.Var:
		return m.getMember(Var(t.Name))
	case ast.Bool:
		return Bool(t)
	case ast.String:
		panic(errors.ErrUnsupported)
	case ast.Number:
		panic(errors.ErrUnsupported)
	case ast.MemberAccess:
		panic(errors.ErrUnsupported)
	case ast.IfExpr:
		panic(errors.ErrUnsupported)
	case ast.CondExpr:
		panic(errors.ErrUnsupported)
	default:
		panic(errors.ErrUnsupported)
	}
}

func (p *parser) parseCondTarget(m typeinfo, expr ast.Expr) Expr {
	switch t := expr.(type) {
	case ast.Var:
		return Var(m.getMember(Var(t.Name)))
	case ast.Bool:
		return Bool(t)
	case ast.String:
		return String(t)
	case ast.Number:
		return Number(t)
	case ast.MemberAccess:
		panic(errors.ErrUnsupported)
	case ast.IfExpr:
		panic(errors.ErrUnsupported)
	case ast.CondExpr:
		panic(errors.ErrUnsupported)
	default:
		panic(errors.ErrUnsupported)
	}
}

func (p *parser) parseCondCase(expr ast.Expr) Expr {
	switch t := expr.(type) {
	case ast.Bool:
		return Bool(t)
	case ast.String:
		return String(t)
	case ast.Number:
		return Number(t)
	case ast.Var:
		panic(errors.ErrUnsupported)
	case ast.MemberAccess:
		panic(errors.ErrUnsupported)
	case ast.IfExpr:
		panic(errors.ErrUnsupported)
	case ast.CondExpr:
		panic(errors.ErrUnsupported)
	default:
		panic(errors.ErrUnsupported)
	}
}

func (p *parser) getQualifiedName(e ast.Expr) Var {
	switch v := e.(type) {
	case ast.MemberAccess:
		right := Var(v.Member.Name)
		left := p.getQualifiedName(v.Object)
		return left + "." + right
	case ast.Var:
		return Var(v.Name)
	default:
		panic(fmt.Sprintf("unexpected ast.Expr: %#v", v))
	}
}
