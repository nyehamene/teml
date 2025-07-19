package ast

import (
	"errors"
	"fmt"
	"reflect"

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
	type0 := "any"
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

	// + the return statement at the end of the function
	body := make([]Stmt, 0, len(stmts)+1)

	for _, stmt := range stmts {
		p.parseStmt(stmt.Element, &body)
	}

	// add return statement
	body = append(body, ReturnNil{})

	node := Method{
		Name: methodname,
		Type: typename,
		Body: body,
	}
	return node
}

func (p *parser) parseStmt(elem ast.Element, stmts *[]Stmt) {
	switch t := elem.(type) {
	case ast.TextElement:
		errvar := makeErrVar()
		pe := doubleQuoteString(formatParagraph(stripDoubleQuote(t.Text)))
		stmt1 := WriteLiteralString{Literal: pe, Var: errvar}
		stmt2 := ReturnIfNotNil(errvar)
		*stmts = append(*stmts, stmt1, stmt2)

	case ast.NumberElement:
		panic(errors.ErrUnsupported)
	case ast.StringElement:
		panic(errors.ErrUnsupported)
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
		panic(fmt.Sprintf("unexpected element type: %v", reflect.TypeOf(elem)))
	}
}
