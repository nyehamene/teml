package generator

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	html "github.com/eml-lang/teml/codegen/go/source"
	ast "github.com/eml-lang/teml/codegen/go/transpiler"
)

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

	for _, st := range fsrc.Structs {
		if err := g.writeStruct(st); err != nil {
			return err
		}
	}

	for _, m := range fsrc.Methods {
		if err := g.writeMethod(m); err != nil {
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

	if err := g.writeln("}"); err != nil {
		return err
	}

	return nil
}

func (g *generator) writeMethod(m ast.Method) error {
	methodReceiver := strings.ToLower(m.Type[0:1])
	header := fmt.Sprintf("func (%s %s) %s(ctx context.Context, w io.Writer) error {", methodReceiver, m.Type, m.Name)
	if err := g.writeln(header); err != nil {
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

	case ast.WriteLiteralString:
		err = g.writeln(fmt.Sprintf("_, %s := io.WriteString(w, %s)", t.Var, t.Literal))

	default:
		panic(fmt.Sprintf("unexpected stmt type: %v", reflect.TypeOf(stmt)))
	}

	if err != nil {
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
