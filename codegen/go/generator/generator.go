package generator

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	html "github.com/eml-lang/teml/codegen/go/source"
	ast "github.com/eml-lang/teml/codegen/go/transpiler"
)

const NameRenderContextVar = "rdc"
const NamedRenderCopyMethod = "copyRenderContextWithAttributes"

func Generate(stdout io.Writer, fsrc *ast.File) error {
	w := html.NewSourceWriter(stdout)
	g := generator{
		writer: w,
	}

	if err := g.writePackage(&fsrc.Package); err != nil {
		return err
	}

	for _, i := range fsrc.Imports {
		if err := g.writeImport(i); err != nil {
			return err
		}
	}

	if err := g.writeRenderContextStruct(fsrc.RenderContext); err != nil {
		return err
	}

	if err := g.writeComponentInterface(); err != nil {
		return nil
	}

	for _, st := range fsrc.Structs {
		if err := g.writeStruct(st); err != nil {
			return err
		}
	}

	for _, gvar := range fsrc.GlobalVars {
		if err := g.writeGlobalVar(gvar); err != nil {
			return err
		}
	}

	for _, m := range fsrc.Methods {
		if err := g.writeRenderMethod(fsrc.RenderContext, m); err != nil {
			return err
		}
	}

	return nil
}

type generator struct {
	writer html.SourceWriter
}

func (g *generator) writePackage(pkg *ast.Package) error {
	if err := g.writeln("package " + ast.ResolveValue(pkg.Name)); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeImport(i ast.Import) error {
	if err := g.writeln("import " + ast.ResolveValue(i.Name) + " " + ast.ResolveValue(i.Path)); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeStruct(st ast.Struct) error {
	if err := g.writeln("type " + ast.ResolveValue(st.Name) + " struct {"); err != nil {
		return err
	}

	for _, field := range st.Fields {
		if err := g.writeField(field); err != nil {
			return err
		}
	}

	if err := g.writeln("}"); err != nil {
		return err
	}

	if err := g.writeEnumType(st.Fields); err != nil {
		return err
	}

	return nil
}

func (g *generator) writeEnumType(fields []ast.StructField) error {
	enums := []ast.Enum{}
	for _, field := range fields {
		switch t := field.Type.(type) {
		case ast.Var: // noop
		case ast.Enum:
			enums = append(enums, t)
		default:
			panic(fmt.Sprintf("unexpected field type: %v", reflect.TypeOf(field)))
		}
	}

	for _, enum := range enums {
		if err := g.writefn("type %s string", enum.Name()); err != nil {
			return err
		}
		for _, c := range enum.Constants {
			value := ast.ResolveValue(c)
			name := ast.ResolveEnumContantAsName(c)
			name = ast.ResolveValue(enum.TypeName) + name
			if err := g.writefn("const %s = %s", name, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *generator) writeField(f ast.StructField) error {
	if err := g.writeln(ast.ResolveValue(f.Name) + " " + ast.ResolveType(f.Type)); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeComponentInterface() error {
	src := `
	type %s interface {
		%s(%s) error
	}

	type %s func() error

	func (fc %[4]s) %[2]s(ctx %[3]s) error {
		return fc()
	}
	`
	err := g.writefn(src, ast.NameComponentInterface, ast.NameComponentRenderMethod, ast.NameContextStruct, ast.NameFuncComponentStruct)
	if err != nil {
		return err
	}
	return nil
}

func (g *generator) writeRenderContextStruct(rc ast.RenderContextStruct) error {
	if err := g.writeStruct(rc.Struct); err != nil {
		return err
	}

	// write option type
	if err := g.writefn("type %s func(ctx *%s)", rc.OptionType, rc.Name); err != nil {
		return err
	}

	// write consturctor
	if err := g.writefn("func %s(opts ...%s) %s {", rc.Constructor, rc.OptionType, rc.Name); err != nil {
		return err
	}
	if err := g.writefn("ctx := %s{}", rc.Name); err != nil {
		return err
	}
	if err := g.writeln("for _, opt := range opts {"); err != nil {
		return err
	}
	if err := g.writeln("opt(&ctx)"); err != nil {
		return err
	}
	if err := g.writeln("}"); err != nil {
		return err
	}
	if err := g.writeln("return ctx"); err != nil {
		return err
	}
	if err := g.writeln("}"); err != nil {
		return err
	}

	// write an option setter function for each field in the struct
	for _, opt := range rc.Fields {
		name := ast.ResolveValue(opt.Name)
		funcname := strings.ToUpper(name[0:1]) + name[1:]
		argname := "val"

		if err := g.writefn("func Set%s(%s %s) %s {", funcname, argname, opt.Type, rc.OptionType); err != nil {
			return err
		}
		if err := g.writefn("return func(ctx *%s) {", rc.Name); err != nil {
			return err
		}
		if err := g.writefn("ctx.%s = %s", opt.Name, argname); err != nil {
			return err
		}
		if err := g.writeln("}"); err != nil {
			return err
		}
		if err := g.writeln("}"); err != nil {
			return err
		}
	}

	return nil
}

func (g *generator) writeRenderMethod(rc ast.RenderContextStruct, m ast.RenderMethod) error {
	header := fmt.Sprintf("func (%s %s) %s(%s %s) error {", m.Receiver, m.Type, m.Name, NameRenderContextVar, rc.Name)
	if err := g.writeln(header); err != nil {
		return err
	}

	if err := g.writefn("ctx := %s.ctx", NameRenderContextVar); err != nil {
		return err
	}
	if err := g.writefn("w := %s.writer", NameRenderContextVar); err != nil {
		return err
	}

	lastStmtIdx := len(m.Body) - 1
	body := make([]ast.Stmt, 0, len(m.Body)+2)
	body = append(body, m.Body[0:lastStmtIdx]...)
	body = append(body, ast.BlankVar("ctx"), ast.BlankVar("w"), m.Body[lastStmtIdx])

	for _, stmt := range body {
		if err := g.writeStmt(stmt); err != nil {
			return err
		}
	}

	if err := g.writeln("}"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeInherittedChildren(ctxVar ast.Var) error {
	if err := g.writefn("%s := %s{", ctxVar, ast.NameContextStruct); err != nil {
		return err
	}
	if err := g.writefn("%s: %s.%[1]s,", ast.NameContextField, NameRenderContextVar); err != nil {
		return err
	}
	if err := g.writefn("%s: %s.%[1]s,", ast.NameWriterField, NameRenderContextVar); err != nil {
		return err
	}
	if err := g.writeln("}"); err != nil {
		return err
	}
	if err := g.writefn("for _, child := range %s.%s {", NameRenderContextVar, ast.NameChildrenField); err != nil {
		return err
	}
	if err := g.writefn("err := child.%s(%s)", ast.NameComponentRenderMethod, ctxVar); err != nil {
		return err
	}
	if err := g.writeln("if err != nil {\nreturn err\n}"); err != nil {
		return err
	}
	if err := g.writeln("}"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeStmt(stmt ast.Stmt) error {
	var err error

	switch t := stmt.(type) {
	case ast.BlankVar:
		err = g.writefn("_ = %s", t)

	case ast.ReturnNil:
		err = g.writeln("return nil")

	case ast.ReturnIfNotNil:
		err = g.writeln("if " + string(t) + " != nil {")
		if err != nil {
			break
		}
		err = g.writeln("return " + string(t))
		if err != nil {
			break
		}
		err = g.writeln("}")

	case ast.MapInstance:
		err = g.writeMapInstance(t)

	case ast.CopyRenderContext:
		err = g.writeCopyRenderContext(t)

	case ast.StringLiteral:
		err = g.writefn("_, %s := io.WriteString(w, %s)", t.Error, t.Value)

	case ast.StringMemberAccessExpr:
		// TODO sanitize user input (t.Value)
		err = g.writefn("_, %s := io.WriteString(w, %s)", t.Error, t.Value)

	case ast.NumberMemberAccessExpr:
		// TODO sanitize user input (t.Value)
		err = g.writefn("%s := fmt.Sprintf(%q, %s)", t.Variable, "%d", t.Value)
		if err != nil {
			return err
		}
		err = g.writefn("_, %s := io.WriteString(w, %s)", t.Error, t.Variable)

	case ast.InheritAttributes:
		err = g.writeInherittedAttributes(t)

	case ast.InheritChildren:
		err = g.writeInherittedChildren(t.Context)

	case ast.CallRenderMethod:
		err = g.writefn("%s := %s.%s(%s)", t.Error, t.Receiver, t.Name, t.Context)

	case ast.If:
		err = g.writeIfStmt(t)

	case ast.Cond:
		err = g.writeCond(t)

	case ast.StructInstance:
		err = g.writeStructInstance(t)

	case ast.SliceInstance:
		err = g.writeSliceInstance(t)

	case ast.ComponentInstance:
		err = g.writeFuncComponent(t)

	default:
		panic(fmt.Sprintf("unexpected stmt type: %v", reflect.TypeOf(stmt)))
	}

	if err != nil {
		return err
	}

	return nil
}

func (g *generator) writeCopyRenderContext(node ast.CopyRenderContext) error {
	if err := g.writefn("%s := %s{", node.Variable, ast.NameContextStruct); err != nil {
		return err
	}
	if err := g.writefn("%s: %s.%[1]s,", ast.NameWriterField, NameRenderContextVar); err != nil {
		return err
	}
	if err := g.writefn("%s: %s.%[1]s,", ast.NameContextField, NameRenderContextVar); err != nil {
		return err
	}
	if err := g.writefn("%s: %s,", ast.NameAttributeField, node.Attrs); err != nil {
		return err
	}
	if err := g.writefn("%s: %s,", ast.NameChildrenField, node.Children); err != nil {
		return err
	}
	if err := g.writeln("}"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeMapInstance(node ast.MapInstance) error {
	if err := g.writefn("%s := map[string]string{", node.Variable); err != nil {
		return err
	}
	for _, entry := range node.Entries {
		if err := g.writefn("%s: %s,", ast.ResolveMapKey(entry.Key), ast.ResolveValue(entry.Value)); err != nil {
			return err
		}
	}
	if err := g.writeln("}"); err != nil {
		return err
	}

	return nil
}

func (g *generator) writeFuncComponent(node ast.ComponentInstance) error {
	if err := g.writefn("var %s %s = func() error {", node.Variable, ast.NameFuncComponentStruct); err != nil {
		return err
	}

	for _, stmt := range node.Stmts {
		if err := g.writeStmt(stmt); err != nil {
			return err
		}
	}

	if err := g.writeln("}"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeSliceInstance(node ast.SliceInstance) error {
	if err := g.writefn("%s := []%s{", node.Variable, node.Type); err != nil {
		return err
	}
	for _, value := range node.Values {
		if err := g.writefn("%s,", value); err != nil {
			return err
		}
	}
	if err := g.writeln("}"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeStructInstance(node ast.StructInstance) error {
	if err := g.writefn("%s := %s{", node.Variable, node.Type); err != nil {
		return err
	}

	for _, field := range node.Parameters {
		if err := g.writefn("%s: %s,", field.Name, ast.ResolveValue(field.Value)); err != nil {
			return err
		}
	}

	if err := g.writeln("}"); err != nil {
		return err
	}

	return nil
}

func (g *generator) writeCond(condsmt ast.Cond) error {
	target := ast.ResolveValue(condsmt.Target)

	if err := g.writefn("switch %s {", target); err != nil {
		return err
	}

	for _, c := range condsmt.Cases {
		match := ast.ResolveSwitchTarget(c.Match)

		if err := g.writefn("case %s:", match); err != nil {
			return err
		}
		for _, stmt := range c.Branch {
			if err := g.writeStmt(stmt); err != nil {
				return err
			}
		}
	}

	if err := g.writeln("default:"); err != nil {
		return err
	}

	if err := g.writefn("panic(%s)", fmt.Sprintf("fmt.Sprintf(%q, %s)", "unexpected enum value: %s", target)); err != nil {
		return err
	}

	if err := g.writeln("}"); err != nil {
		return err
	}

	return nil
}

func (g *generator) writeIfStmt(ifsmt ast.If) error {
	if err := g.writefn("if %s {", ifsmt.Cond); err != nil {
		return err
	}

	for _, stmt := range ifsmt.Then {
		err := g.writeStmt(stmt)
		if err != nil {
			return err
		}
	}

	// then
	if len(ifsmt.Else) > 0 {
		if err := g.writeln("} else {"); err != nil {
			return err
		}

		// else
		for _, stmt := range ifsmt.Else {
			err := g.writeStmt(stmt)
			if err != nil {
				return err
			}
		}
	}

	if err := g.writeln("}"); err != nil {
		return err
	}

	return nil
}

func (g *generator) writeInherittedAttributes(node ast.InheritAttributes) error {
	err := g.writefn("for key, val := range %s.%s {", NameRenderContextVar, ast.NameAttributeField)
	if err != nil {
		return err
	}
	err = g.writefn("var %s error", node.Error)
	if err != nil {
		return err
	}
	err = g.writeln("attr := fmt.Sprintf(\" %s=%s\", key, val)")
	if err != nil {
		return err
	}
	err = g.writefn("_, %s = io.WriteString(w, attr)", node.Error)
	if err != nil {
		return err
	}
	err = g.writefn("if %s != nil {", node.Error)
	if err != nil {
		return err
	}
	err = g.writefn("return %s", node.Error)
	if err != nil {
		return err
	}
	err = g.writeln("}")
	if err != nil {
		return err
	}
	err = g.writeln("}")
	if err != nil {
		return err
	}
	return nil
}

func (g *generator) writeGlobalVar(gvar ast.BlankVar) error {
	if err := g.writeln("var _ = " + string(gvar)); err != nil {
		return err
	}
	return nil
}

func (g *generator) write(s string) error {
	if _, err := g.writer.Write(s); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeln(s string) error {
	if err := g.write(s + "\n"); err != nil {
		return err
	}
	return nil
}

//go:format writefn printf 1 2
func (g *generator) writefn(format string, args ...any) error {
	if err := g.writeln(fmt.Sprintf(format, args...)); err != nil {
		return err
	}
	return nil
}
