package generator

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	html "github.com/eml-lang/teml/codegen/go/source"
	ast "github.com/eml-lang/teml/codegen/go/transpiler"
)

const RenderContextVar = "rdc"
const RenderCopyMethod = "copyRenderContextWithAttributes"

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
	if err := g.writeln("package " + pkg.Name); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeImport(i ast.Import) error {
	if err := g.writeln("import " + i.Name + " " + i.Path); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeStruct(st ast.Struct) error {
	if err := g.writeln("type " + st.Name + " struct {"); err != nil {
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
		case ast.String: // noop
		case ast.Number: // noop
		case ast.Bool: // noop
		case ast.Var: // noop
		case ast.Enum:
			enums = append(enums, t)
		default:
			panic(fmt.Sprintf("unexpected field type: %v", reflect.TypeOf(field)))
		}
	}

	for _, enum := range enums {
		if err := g.writeln(fmt.Sprintf("type %s string", enum.Name())); err != nil {
			return err
		}
		for _, c := range enum.Constants {
			value := g.resolveValue(c)
			name := enum.TypeName + value
			if err := g.writeln(fmt.Sprintf("const %s = %s", name, value)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *generator) writeField(f ast.StructField) error {
	if err := g.writeln(f.Name + " " + f.Type.Name()); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeComponentInterface() error {
	src := `
	type Component interface {
		%s(%s)
	}

	type funccomponent func() error

	func (fc funccomponent) %[1]s(ctx %[2]s) error {
		return fc()
	}
	`
	err := g.writeln(fmt.Sprintf(src, ast.RenderComponentMethod, ast.RenderContext, ast.RenderContextWriterField))
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
	if err := g.writeln(fmt.Sprintf("type %s func(ctx *%s)", rc.OptionType, rc.Name)); err != nil {
		return err
	}

	// write consturctor
	if err := g.writeln(fmt.Sprintf("func %s(opts ...%s) %s {", rc.Constructor, rc.OptionType, rc.Name)); err != nil {
		return err
	}
	if err := g.writeln(fmt.Sprintf("ctx := %s{}", rc.Name)); err != nil {
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
		funcname := strings.ToUpper(opt.Name[0:1]) + opt.Name[1:]
		argname := "val"

		if err := g.writeln(fmt.Sprintf("func Set%s(%s %s) %s {", funcname, argname, opt.Type, rc.OptionType)); err != nil {
			return err
		}
		if err := g.writeln(fmt.Sprintf("return func(ctx *%s) {", rc.Name)); err != nil {
			return err
		}
		if err := g.writeln(fmt.Sprintf("ctx.%s = %s", opt.Name, argname)); err != nil {
			return err
		}
		if err := g.writeln("}"); err != nil {
			return err
		}
		if err := g.writeln("}"); err != nil {
			return err
		}
	}

	// generate a function to make a copy of the render context
	copyFuncHeader := fmt.Sprintf("func %s(%s %s, attrs map[string]string) %s {",
		RenderCopyMethod,
		RenderContextVar,
		ast.RenderContext,
		ast.RenderContext,
	)
	if err := g.writeln(copyFuncHeader); err != nil {
		return err
	}
	const newctx = "ctxcopy"
	if err := g.writeln(fmt.Sprintf("%s := %s{}", newctx, ast.RenderContext)); err != nil {
		return err
	}
	for _, field := range rc.Fields {
		name := field.Name
		if name == ast.RenderContextAttrsField {
			continue
		}
		if err := g.writeln(fmt.Sprintf("%s.%s = %s.%s", newctx, name, RenderContextVar, name)); err != nil {
			return err
		}
	}

	if err := g.writeln(fmt.Sprintf("%s.%s = attrs", newctx, ast.RenderContextAttrsField)); err != nil {
		return err
	}
	if err := g.writeln(fmt.Sprintf("return %s", newctx)); err != nil {
		return err
	}

	if err := g.writeln("}"); err != nil {
		return err
	}

	return nil
}

func (g *generator) writeRenderMethod(rc ast.RenderContextStruct, m ast.RenderMethod) error {
	header := fmt.Sprintf("func (%s %s) %s(%s %s) error {", m.Receiver, m.Type, m.Name, RenderContextVar, rc.Name)
	if err := g.writeln(header); err != nil {
		return err
	}

	if err := g.writeln(fmt.Sprintf("ctx := %s.ctx", RenderContextVar)); err != nil {
		return err
	}
	if err := g.writeln(fmt.Sprintf("w := %s.writer", RenderContextVar)); err != nil {
		return err
	}

	if err := g.writeln("_ = ctx"); err != nil {
		return err
	}
	if err := g.writeln("_ = w"); err != nil {
		return err
	}

	for _, stmt := range m.Body {
		if err := g.writeStmt(stmt); err != nil {
			return err
		}
	}

	if err := g.writeln("}"); err != nil {
		return err
	}
	return nil
}

func (g *generator) writeStmt(stmt ast.Stmt) error {
	var err error

	switch t := stmt.(type) {
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

	case ast.FormatNumber:
		err = g.writeln(fmt.Sprintf("%s := fmt.Sprintf(%q, %s)", t.Variable, "%d", t.Value))

	case ast.MapInstance:
		err = g.writeMapInstance(t)

	case ast.CopyRenderContext:
		err = g.writeln(fmt.Sprintf("%s := %s(%s, %s)", t.Variable, RenderCopyMethod, RenderContextVar, t.Attrs))

	case ast.StringLiteral:
		err = g.writeln(fmt.Sprintf("_, %s := io.WriteString(w, %s)", t.Variable, t.Value))

	case ast.StringMemberAccessExpr:
		// TODO sanitize user input (t.Value)
		err = g.writeln(fmt.Sprintf("_, %s := io.WriteString(w, %s)", t.Variable, t.Value))

	case ast.NumberMemberAccessExpr:
		// TODO sanitize user input (t.Value)
		err = g.writeln(fmt.Sprintf("_, %s := io.WriteString(w, %s)", t.Variable, t.Value))

	case ast.InheritAttributes:
		err = g.writeInherittedAttributes(t)

	case ast.CallRenderMethod:
		err = g.writeln(fmt.Sprintf("%s := %s.%s(%s)", t.Variable, t.Receiver, t.Name, t.Context))

	case ast.If:
		err = g.writeIfStmt(t)

	case ast.Cond:
		err = g.writeCond(t)

	case ast.StructInstance:
		err = g.writeStructInstance(t)

	default:
		panic(fmt.Sprintf("unexpected stmt type: %v", reflect.TypeOf(stmt)))
	}

	if err != nil {
		return err
	}

	return nil
}

func (g *generator) writeMapInstance(node ast.MapInstance) error {
	if err := g.writeln(fmt.Sprintf("%s := map[string]string{", node.Variable)); err != nil {
		return err
	}

	for _, entry := range node.Entries {
		if err := g.writeln(fmt.Sprintf("%s: %s,", g.resolveMapKey(entry.Key), g.resolveValue(entry.Value))); err != nil {
			return err
		}
	}

	if err := g.writeln("}"); err != nil {
		return err
	}

	return nil
}

func (g *generator) writeStructInstance(node ast.StructInstance) error {
	if err := g.writeln(fmt.Sprintf("%s := %s{", node.Variable, node.Type)); err != nil {
		return err
	}

	for _, field := range node.Parameters {
		if err := g.writeln(fmt.Sprintf("%s: %s,", field.Name, g.resolveValue(field.Value))); err != nil {
			return err
		}
	}

	if err := g.writeln("}"); err != nil {
		return err
	}

	return nil
}

func (g *generator) writeCond(condsmt ast.Cond) error {
	target := g.resolveValue(condsmt.Target)

	if err := g.writeln(fmt.Sprintf("switch %s {", target)); err != nil {
		return err
	}

	for _, c := range condsmt.Cases {
		match := g.resolveSwitchTarget(c.Match)

		if err := g.writeln(fmt.Sprintf("case %s:", match)); err != nil {
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

	if err := g.writeln(fmt.Sprintf("panic(%s)", fmt.Sprintf("fmt.Sprintf(%q, %s)", "unexpected enum value: %s", target))); err != nil {
		return err
	}

	if err := g.writeln("}"); err != nil {
		return err
	}

	return nil
}

func (g *generator) writeIfStmt(ifsmt ast.If) error {
	if err := g.writeln(fmt.Sprintf("if %s {", ifsmt.Cond)); err != nil {
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
	err := g.writeln(fmt.Sprintf("for key, val := range %s.%s {", RenderContextVar, ast.RenderContextAttrsField))
	if err != nil {
		return err
	}
	err = g.writeln(fmt.Sprintf("var %s error", node.Variable))
	if err != nil {
		return err
	}
	err = g.writeln("attr := fmt.Sprintf(\" %s=%s\", key, val)")
	if err != nil {
		return err
	}
	err = g.writeln(fmt.Sprintf("_, %s = io.WriteString(w, attr)", node.Variable))
	if err != nil {
		return err
	}
	err = g.writeln(fmt.Sprintf("if %s != nil {", node.Variable))
	if err != nil {
		return err
	}
	err = g.writeln(fmt.Sprintf("return %s", node.Variable))
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
	g.writeln("var _ = " + string(gvar))
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
