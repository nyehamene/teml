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

const defaultTestFileName = "render_test.go"

var usageText = `usage: eml <command> [<args>...]

eml - build  HTML UIs in a simple templating language

commands:
  generate  Generate code for your platform from eml files

supported platforms:
  go
  bun
`

func main() {
	// srcout := strings.Builder{}
	srcout := bytes.Buffer{}
	testout := bytes.Buffer{}

	wd, errcwd := os.Getwd()
	if errcwd != nil {
		panic(errcwd)
	}

	root, errroot := os.OpenRoot(wd)
	if errroot != nil {
		panic(errroot)
	}
	defer root.Close()

	result := run(root, &srcout, &testout, os.Stderr, os.Args)

	switch t := result.(type) {
	case cmd.Error:
		panic(t.Result)
	case cmd.File:
		err := writeFile(root, t, &srcout)
		if err != nil {
			panic(err)
		}
		println("source code written to:", t.File)

		testFile := filepath.Join(filepath.Dir(t.File), defaultTestFileName)
		writeFile(root, cmd.NewFile(testFile), &testout)
		println("test code written to:", testFile)
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

func run(root *os.Root, stdout, testout, stderr io.Writer, args []string) cmd.Result {
	if len(args) < 2 {
		io.WriteString(stderr, usageText)
		os.Exit(64)
	}

	switch cmdarg := args[1]; cmdarg {
	case "generate":
		return generateCmd(root, stdout, testout, args[2:])

	default:
		err := fmt.Errorf("Unexpected command: %s", cmdarg)
		return cmd.NewError(err)
	}
}

func generateCmd(root *os.Root, stdout, testout io.Writer, args []string) cmd.Result {
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
		Writer:     stdout,
		TestWriter: testout,
		File:       file,
		Root:       root,
	}

	result := generatecmd.Generate(cmdargs)
	if result != nil {
		return cmd.NewError(result)
	}

	outfile := strings.Replace(file, path.Ext(file), ".go", 1)
	return cmd.NewFile(outfile)
}
