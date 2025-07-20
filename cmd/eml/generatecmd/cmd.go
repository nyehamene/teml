package generatecmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	parser "github.com/eml-lang/teml/ast"
	codegen "github.com/eml-lang/teml/codegen/go/generator"
	gotranspiler "github.com/eml-lang/teml/codegen/go/transpiler"
	transpiler "github.com/eml-lang/teml/transpiler"

	"github.com/eml-lang/teml/token"
)

type Arguments struct {
	Writer     io.Writer
	TestWriter io.Writer
	File       string
	Root       *os.Root
}

func Generate(args Arguments) error {
	var buf []byte
	var err error

	w := args.Writer
	file := args.File
	root := args.Root

	f, err := root.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	buf, err = io.ReadAll(f)
	if err != nil {
		return err
	}

	toks := token.Scan(buf, token.ReduceAlloc|token.PreserveComment)
	astp := parser.ParseFile(toks, 0)
	for _, errast := range astp.Errors.Each() {
		err = errors.Join(errast)
	}
	if astp.HasError() {
		return err
	}

	astf := transpiler.ParseFile(astp, toks)
	for _, errast := range astf.Errors() {
		err = errors.Join(errast)
	}
	if astf.HasError() {
		return err
	}

	envr := transpiler.ResolveFile(astf)
	for _, errast := range astf.Errors() {
		err = errors.Join(errast)
	}
	if astf.HasError() {
		return err
	}

	_ = transpiler.TypecheckFile(astf, envr)
	for _, errast := range astf.Errors() {
		err = errors.Join(errast)
	}
	if astf.HasError() {
		return err
	}

	gofile := gotranspiler.Parse(astf)
	err = codegen.Generate(w, &gofile)
	if err != nil {
		return err
	}

	testdata := getTestDataFromComment(toks)
	err = writeTestFile(args.TestWriter, testdata)
	if err != nil {
		return err
	}

	return nil
}

func writeTestFile(stdout io.Writer, data string) error {
	source := `
package views

import (
	"context"
	_ "embed"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

//go:embed expected.html
var expectedHTML string

func TestRender(t *testing.T) {
	data := %s

	ctx := context.Background()
	w := strings.Builder{}

	rdrctx := %s{
		ctx: ctx,
		writer: &w,
	}

	err := data.%s(rdrctx)
	if err != nil {
		t.Fatal(err)
	}

	gotHTML := w.String()

	if diff := cmp.Diff(expectedHTML, gotHTML); diff != "" {
		println(gotHTML)
		t.Error(diff)
	}
}
	`

	_, errwrt := io.WriteString(
		stdout,
		fmt.Sprintf(source, data, gotranspiler.RenderContext, gotranspiler.RenderComponentMethod),
	)
	if errwrt != nil {
		return errwrt
	}

	return nil
}

func getTestDataFromComment(toks *token.File) string {
	for _, tok := range toks.Tokens.Each() {
		if tok.Kind != token.Comment {
			continue
		}

		cmt, ok := toks.Text(tok)
		if !ok {
			continue
		}

		if !strings.Contains(cmt, "test:instance") {
			continue
		}

		cmt = strings.Replace(cmt, ";;test:instance", "", 1)
		cmt = strings.TrimLeft(cmt, " ")
		return cmt
	}
	return ""
}
