package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/eml-lang/teml/cmd/eml/cmd"
	"github.com/eml-lang/teml/cmd/eml/generatecmd"
	"github.com/eml-lang/teml/internal/assert"
)

var usageText = `usage: eml <command> [<args>...]

eml - build  HTML UIs in a simple templating language

commands:
  generate  Generate code for your platform from eml files

supported platforms:
  go
  bun
`

func main() {
	// w := strings.Builder{}
	w := bytes.Buffer{}

	wd, errcwd := os.Getwd()
	if errcwd != nil {
		panic(errcwd)
	}

	root, errroot := os.OpenRoot(wd)
	if errroot != nil {
		panic(errroot)
	}
	defer root.Close()

	result := run(root, &w, os.Stderr, os.Args)

	switch t := result.(type) {
	case cmd.Error:
		panic(t.Result)
	case cmd.File:
		err := writeFile(root, t, &w)
		if err != nil {
			panic(err)
		}
		writeTestFile(root, t)
		println("wrote source code to:", t.File)
	default:
		panic(fmt.Sprintf("unexpected cmd result type: %s", reflect.TypeOf(result)))
	}
}

func writeFile(root *os.Root, f cmd.File, r io.Reader) error {
	assert.Assert(root != nil, "root is nil")
	assert.Assert(f.File != "", "filename is empty")

	nf, ferr := root.Create(f.File)
	if ferr != nil {
		return ferr
	}
	defer nf.Close()

	buf, errbuf := io.ReadAll(r)
	if errbuf != nil {
		return errbuf
	}

	// TODO format source
	// fbuf, errfbuf := format.Source(buf)
	// if errfbuf != nil {
	// 	return errfbuf
	// }
	//
	// _, errwrt := nf.Write(fbuf)
	// if errwrt != nil {
	// 	return errwrt
	// }

	// TODO remove when formatting source
	_, errwrt := nf.Write(buf)
	if errwrt != nil {
		return errwrt
	}

	return nil
}

func writeTestFile(root *os.Root, f cmd.File) error {
	assert.Assert(root != nil, "root is nil")
	assert.Assert(f.File != "", "filename is empty")

	path := filepath.Dir(f.File)
	path = filepath.Join(path, "render_test.go")

	nf, ferr := root.Create(path)
	if ferr != nil {
		return ferr
	}
	defer nf.Close()

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
	data := A{}

	ctx := context.Background()
	w := strings.Builder{}

	err := data.Render(ctx, &w)
	if err != nil {
		t.Fatal(err)
	}

	gotHTML := w.String()

	if diff := cmp.Diff(expectedHTML, gotHTML); diff != "" {
		t.Error(diff)
	}
}
	`

	_, errwrt := io.WriteString(nf, source)
	if errwrt != nil {
		return errwrt
	}

	return nil
}

func run(root *os.Root, w, stderr io.Writer, args []string) cmd.Result {
	if len(args) < 2 {
		io.WriteString(stderr, usageText)
		os.Exit(64)
	}

	switch cmdarg := args[1]; cmdarg {
	case "generate":
		return generateCmd(root, w, args[2:])

	default:
		err := fmt.Errorf("Unexpected command: %s", cmdarg)
		return cmd.NewError(err)
	}
}

func generateCmd(root *os.Root, w io.Writer, args []string) cmd.Result {
	var file string

	flags := flag.NewFlagSet("generate", flag.ExitOnError)
	flags.StringVar(&file, "f", "", "")

	if err := flags.Parse(args); err != nil {
		return cmd.NewError(err)
	}

	// if f, err := wdRel(file); err != nil {
	// 	return cmd.NewError(err)
	// } else {
	// 	file = f
	// }

	cmdargs := generatecmd.Arguments{
		Writer: w,
		File:   file,
		Root:   root,
	}

	result := generatecmd.Generate(cmdargs)
	if result != nil {
		return cmd.NewError(result)
	}

	outfile := strings.Replace(file, path.Ext(file), ".go", 1)
	return cmd.NewFile(outfile)
}
