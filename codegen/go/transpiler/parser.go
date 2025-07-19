package ast

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	ast "github.com/eml-lang/teml/transpiler"
)

const defaultMethodName = "Render"

type parser struct {
	fsrc *ast.File
}

func (p *parser) parsePackage() Package {
	name := p.fsrc.Package.Path
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
		field := p.parseStructField(prop)
		fields = append(fields, field)
	}
	node := Struct{
		Name:   name,
		Fields: fields,
	}

	return node
}

func (p *parser) parseStructField(prop ast.Property) StructField {
	name := prop.Ident.Name
	// TODO resolve property type
	type0 := p.resolveType(prop.Type)
	node := StructField{
		Name: name,
		Type: type0,
	}
	return node
}

func (p *parser) parseMethod(decl ast.Declaration) Method {
	var methodname = defaultMethodName
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

	node := Method{
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
		errvar := makeErrVar()
		pe := doubleQuoteString(formatParagraph(stripDoubleQuote(elem.Text)))
		stmt1 := WriteLiteralString{Value: pe, Variable: errvar}
		stmt2 := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt1, stmt2)

	case ast.TextGroupElement:
		// open tag
		errvar := makeErrVar()
		stmt1 := WriteLiteralString{Value: doubleQuoteString("<p>\\n"), Variable: errvar}
		ret1 := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt1, ret1)

		// paragraph lines
		for _, line := range elem.Lines {
			errvar := makeErrVar()
			line = line + "\\n"
			txt := doubleQuoteString(line)
			stmt2 := WriteLiteralString{Value: txt, Variable: errvar}
			ret2 := ReturnIfNotNil(errvar)
			*stmts = append(*stmts, stmt2, ret2)
		}

		// close tag
		errvar = makeErrVar()
		stmt3 := WriteLiteralString{Value: doubleQuoteString("</p>\\n"), Variable: errvar}
		ret3 := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt3, ret3)

	case ast.NumberElement:
		panic(errors.ErrUnsupported)

	case ast.StringElement:
		// begin open tag
		errvar := makeErrVar()
		stmt1 := WriteLiteralString{Value: doubleQuoteString("<p"), Variable: errvar}
		ret1 := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt1, ret1)

		// attribues
		for _, attr := range elem.Attributes {
			p.parseAttr(attr, stmts)
		}

		// end open tag
		errvar = makeErrVar()
		stmt3 := WriteLiteralString{Value: doubleQuoteString(">"), Variable: errvar}
		ret3 := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt3, ret3)

		// content: member access expression
		name := p.resolveName(elem.Tag)
		memberAccess := m.receiver + "." + name
		errvar = makeErrVar()
		stmt4 := WriteStringExpr{Value: memberAccess, Variable: errvar}
		ret4 := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt4, ret4)

		// close tag
		errvar = makeErrVar()
		stmt5 := WriteLiteralString{Value: doubleQuoteString("</p>\\n"), Variable: errvar}
		ret5 := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt5, ret5)

	case ast.ComponentElement:
		panic(errors.ErrUnsupported)
	case ast.InstanceElement:
		panic(errors.ErrUnsupported)
	case ast.NativeElement:
		panic(errors.ErrUnsupported)
	case ast.IFElement:
		panic(errors.ErrUnsupported)
	case ast.CondElement:
		panic(errors.ErrUnsupported)
	default:
		panic(fmt.Sprintf("unexpected element type: %v", reflect.TypeOf(node)))
	}
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

	attr := " " + key + "=" + value

	errvar := makeErrVar()
	stmt := WriteLiteralString{Value: attr, Variable: errvar}
	ret := ReturnIfNotNil(errvar)
	*stmts = append(*stmts, stmt, ret)
}
