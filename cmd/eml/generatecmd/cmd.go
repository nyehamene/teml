package generatecmd

import (
	"errors"
	"io"
	"os"

	parser "github.com/eml-lang/teml/ast"
	codegen "github.com/eml-lang/teml/codegen/go/generator"
	gotranspiler "github.com/eml-lang/teml/codegen/go/transpiler"
	transpiler "github.com/eml-lang/teml/transpiler"

	"github.com/eml-lang/teml/token"
)

type Arguments struct {
	Writer io.Writer
	File   string
	Root   *os.Root
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

	toks := token.Scan(buf, token.ReduceAlloc)
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

	return nil
}
