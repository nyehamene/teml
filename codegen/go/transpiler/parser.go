package ast

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	ast "github.com/eml-lang/teml/transpiler"
)

type parser struct{}

func (p *parser) parsePackage(pkg ast.Package) Package {
	name := pkg.Path
	name = stripDoubleQuote(name)
	node := Package{name}
	return node
}

func (p *parser) parseImport(i ast.Import) Import {
	name := i.Ident.Name
	path := i.Path
	node := Import{Name: name, Path: path}
	return node
}

// parseUsing returns a type alias named name for a type in the package pkgname
func (p *parser) parseUsing(u ast.Using) func(func(TypeAlias) bool) {
	return func(yield func(TypeAlias) bool) {
		pkgname := u.From.Name
		for _, ident := range u.Idents {
			typename := ident.Name
			alias := TypeAlias{
				Name: typename,
				Type: fmt.Sprintf("%s.%s", pkgname, typename),
			}
			if !yield(alias) {
				break
			}
		}
	}
}

func (p *parser) parseStruct(decl ast.Declaration) Struct {
	var name string
	var props []ast.Property

	switch t := decl.(type) {
	case ast.Component:
		name = t.Ident.Name
		props = t.Properties
	case ast.Document:
		panic(errors.ErrUnsupported)
	default:
		panic(fmt.Sprintf("unexpected declaration type: %v", reflect.TypeOf(decl)))
	}

	fields := make([]StructField, 0, len(props))
	for _, prop := range props {
		field := p.parseStructField(name, prop)
		fields = append(fields, field)
	}
	node := Struct{
		Name:   name,
		Fields: fields,
	}

	return node
}

func (p *parser) parseStructField(structname string, prop ast.Property) StructField {
	name := prop.Ident.Name
	// TODO resolve property type
	type0 := p.resolveFieldType(structname, name, prop.Type)
	node := StructField{
		Name: name,
		Type: type0,
	}
	return node
}

func (p *parser) parseMethod(decl ast.Declaration) RenderMethod {
	var methodname = RenderComponentMethod
	var typename string
	var stmts []ast.Stmt

	switch t := decl.(type) {
	case ast.Component:
		typename = t.Ident.Name
		stmts = t.Stmts

	case ast.Document:
		panic(errors.ErrUnsupported)

	default:
		panic(fmt.Sprintf("unexpected declaration type: %v", reflect.TypeOf(decl)))
	}

	// method receiver variable name
	receiver := strings.ToLower(typename[0:1])

	// + the return statement at the end of the function
	body := make([]Stmt, 0, len(stmts)+1)

	m := methodinfo{
		receiver: receiver,
		typename: typename,
		name:     methodname,
	}
	for _, stmt := range stmts {
		p.parseStmt(m, stmt.Element, &body)
	}

	// add return statement
	body = append(body, ReturnNil{})

	node := RenderMethod{
		Name:     methodname,
		Type:     typename,
		Receiver: receiver,
		Body:     body,
	}
	return node
}

type methodinfo struct {
	receiver string
	typename string
	name     string
}

func (p *parser) parseStmt(m methodinfo, node ast.Element, stmts *[]Stmt) {
	switch elem := node.(type) {
	case ast.TextElement:
		const tag = "span"
		const inheritParentAttr = false
		attrs := createAttributes(map[string]string{"style": "display: contents"})
		// open tag
		p.parseOpenTagWithAttributes(tag, attrs, inheritParentAttr, stmts)

		// content
		errvar := makeErrVar()
		stmt1 := WriteLiteralString{Value: elem.Text, Variable: errvar}
		ret1 := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt1, ret1)

		// close tag
		p.parseCloseTag(tag, stmts)

	case ast.TextGroupElement:
		const tag = "span"
		const inheritParentAttr = false
		attrs := createAttributes(map[string]string{"style": "display: contents"})
		// open tag
		p.parseOpenTagWithAttributes(tag, attrs, inheritParentAttr, stmts)

		// paragraph lines
		for _, line := range elem.Lines {
			errvar := makeErrVar()
			txt := doubleQuoteString(line)
			stmt2 := WriteLiteralString{Value: txt, Variable: errvar}
			ret2 := ReturnIfNotNil(errvar)
			*stmts = append(*stmts, stmt2, ret2)
		}

		// close tag
		p.parseCloseTag(tag, stmts)

	case ast.NumberElement:
		// open tag
		const tag = "data"
		const inheritParentAttr = true
		p.parseOpenTagWithAttributes(tag, elem.Attributes, inheritParentAttr, stmts)

		// content
		member := p.resolveName(elem.Tag)
		memberAccess := m.receiver + "." + member

		tempvar := makeTempVar()
		stmt1 := FormatNumber{Value: memberAccess, Variable: tempvar}
		errvar := makeErrVar()
		stmt2 := WriteNumberMemberAccess{Value: tempvar, Variable: errvar}
		ret2 := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt1, stmt2, ret2)

		// close tag
		p.parseCloseTag(tag, stmts)

	case ast.StringElement:
		// open tag
		const tag = "span"
		const inheritParentAttr = true
		p.parseOpenTagWithAttributes(tag, elem.Attributes, inheritParentAttr, stmts)

		// content
		member := p.resolveName(elem.Tag)
		memberAccess := m.receiver + "." + member
		errvar := makeErrVar()
		stmt := WriteStringMemberAccess{Value: memberAccess, Variable: errvar}
		ret := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt, ret)

		// close tag
		p.parseCloseTag(tag, stmts)

	case ast.NativeElement:
		const inheritParentAttr = true
		// open tag
		tag := p.resolveName(elem.Tag)
		p.parseOpenTagWithAttributes(tag, elem.Attributes, inheritParentAttr, stmts)

		for _, stmt := range elem.Body {
			p.parseStmt(m, stmt.Element, stmts)
		}

		// close tag
		p.parseCloseTag(tag, stmts)

	case ast.ComponentElement:
		// attributes
		attrvar := p.parseAttribute(elem.Attributes, stmts)
		ctxvar := makeTempVar()
		stmt1 := CopyContextWithAttributes{Attrs: attrvar, Variable: ctxvar}

		// TODO render component body

		name := m.name
		member := p.resolveName(elem.Tag)
		memberAccess := m.receiver + "." + member
		errvar := makeErrVar()
		stmt2 := CallRenderFunction{Context: ctxvar, Name: name, Receiver: memberAccess, Variable: errvar}
		ret2 := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt1, stmt2, ret2)

	case ast.InstanceElement:
		// TODO generate html for instance element
		panic(errors.ErrUnsupported)

	case ast.IFElement:
		cond := p.resolveIfCond(m, elem.Cond)

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

	case ast.CondElement:
		target := p.resolveCondTarget(m, elem.Target)

		// cases
		cases := []Case{}
		for _, c := range elem.Cases {
			branch := []Stmt{}
			cond := p.resolveCondCase(c.Cond)
			p.parseStmt(m, c.Branch.Element, &branch)
			cases = append(cases, Case{Match: cond, Branch: branch})
		}

		stmt1 := Cond{Target: target, Cases: cases}
		*stmts = append(*stmts, stmt1)

	default:
		panic(fmt.Sprintf("unexpected element type: %v", reflect.TypeOf(node)))
	}
}

func (p *parser) parseAttribute(attrs []ast.Attr, stmts *[]Stmt) string {
	mapvar := makeTempVar()
	stmt1 := MapVar{Variable: mapvar}
	*stmts = append(*stmts, stmt1)

	for _, attr := range attrs {
		for _, entry := range attr.Entries {
			key := entry.Key.Name
			value := p.resolveValue(entry.Value)
			stmt2 := MapEntry{Name: mapvar, Key: key, Value: value}
			*stmts = append(*stmts, stmt2)
		}
	}
	return mapvar
}

func (p *parser) parseOpenTagWithAttributes(name string, attrs []ast.Attr, inherit bool, stmts *[]Stmt) {
	// begin open tag
	errvar := makeErrVar()
	stmt1 := WriteLiteralString{Value: doubleQuoteString(fmt.Sprintf("<%s", name)), Variable: errvar}
	ret1 := ReturnIfNotNil(errvar)
	*stmts = append(*stmts, stmt1, ret1)

	if inherit {
		errvar = makeErrVar()
		stmti := WriteInheritedAttributes{Variable: errvar}
		*stmts = append(*stmts, stmti)
	}

	// attribues
	for _, attr := range attrs {
		p.parseAttr(attr, stmts)
	}

	// end open tag
	errvar = makeErrVar()
	stmt2 := WriteLiteralString{Value: doubleQuoteString(">"), Variable: errvar}
	ret2 := ReturnIfNotNil(errvar)
	*stmts = append(*stmts, stmt2, ret2)
}

func (p *parser) parseCloseTag(name string, stmts *[]Stmt) {
	errvar := makeErrVar()
	stmt := WriteLiteralString{Value: doubleQuoteString(fmt.Sprintf("</%s>", name)), Variable: errvar}
	ret := ReturnIfNotNil(errvar)
	*stmts = append(*stmts, stmt, ret)
}

func (p *parser) parseAttr(attr ast.Attr, stmts *[]Stmt) {
	for _, entry := range attr.Entries {
		p.parseEntry(entry, stmts)
	}
}

func (p *parser) parseEntry(entry ast.KeyVal, stmts *[]Stmt) {
	key := entry.Key.Name
	var value string

	switch t := entry.Value.(type) {
	case ast.String:
		value = string(t)

	case ast.Number:
		value = string(t)

	case ast.Bool:
		value = string(t)

	case ast.Var:
		value = t.Name

	case ast.IFExpr:
		panic(errors.ErrUnsupported)

	case ast.CondExpr:
		panic(errors.ErrUnsupported)

	case ast.MemberAccess:
		panic(errors.ErrUnsupported)

	case ast.Enum:
		panic(errors.ErrUnsupported)

	default:
		panic(fmt.Sprintf("unexpected expression type: %v", reflect.TypeOf(entry.Value)))
	}

	attr := " " + key + "=" + escapeSurrounding(value)

	errvar := makeErrVar()
	stmt := WriteLiteralString{Value: doubleQuoteString(attr), Variable: errvar}
	ret := ReturnIfNotNil(errvar)
	*stmts = append(*stmts, stmt, ret)
}
