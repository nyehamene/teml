package generatecmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/eml-lang/teml/cmd/eml/cmd"
	"github.com/eml-lang/teml/internal/assert"
	"github.com/eml-lang/teml/internal/source"
	"github.com/eml-lang/teml/token"

	codegen "github.com/eml-lang/teml/codegen/go/generator"
	gotranspiler "github.com/eml-lang/teml/codegen/go/transpiler"
)

type Arguments struct {
	Writer     io.Writer
	TestWriter io.Writer
	File       string
	Root       *os.Root
}

func Generate(args Arguments) error {
	w := args.Stdout
	file := args.Path
	root := args.Root

	f, err := root.Open(file)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	buf, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	src := source.NewFile(file, buf)
	err = codegen.Generate(w, src)
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
	"fmt"

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

	expected := strings.TrimSpace(expectedHTML)
	expected = strings.ReplaceAll(expected, "\n", "")

	got := w.String()
	got = strings.TrimSpace(got)
	got = strings.ReplaceAll(got, "\n", "")

	if diff := cmp.Diff(expected, got); diff != "" {
		fmt.Println(got)
		t.Error(diff)
	}
}
	`

	_, errwrt := io.WriteString(
		stdout,
		fmt.Sprintf(source, data, gotranspiler.NameContextStruct, gotranspiler.NameComponentRenderMethod),
	)
	if errwrt != nil {
		return errwrt
	}

	return nil
}

func getTestDataFromComment(toks *token.File) string {
	for _, tok := range toks.Tokens {
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
