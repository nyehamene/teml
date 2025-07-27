package tests

import (
	"embed"
	"path"
	"strings"
	"testing"

	codegen "github.com/eml-lang/teml/codegen/go/generator"
	gotranspiler "github.com/eml-lang/teml/codegen/go/transpiler"
	"github.com/eml-lang/teml/token"

	parser "github.com/eml-lang/teml/ast"
	perrors "github.com/eml-lang/teml/internal/errors"
	transpiler "github.com/eml-lang/teml/transpiler"
)

type ErrorSeq = func(yield func(int, perrors.Error) bool)

type ErrorHandler func(t *testing.T, stage CompilationStage, hasError bool, errors ErrorSeq) bool
type ResultHandler func(fn string, result string) error

func NewErrorHandler(targetStage CompilationStage, failOnSuccessFlag bool) ErrorHandler {
	return func(t *testing.T, stage CompilationStage, hasError bool, errors ErrorSeq) bool {
		lbl := "unknown stage"

		switch stage {
		case StageParsed:
			lbl = "parsing"
		case StageTransformed:
			lbl = "transformation"
		case StageResolved:
			lbl = "resolution"
		case StageTypechecked:
			lbl = "type checking"
		case StageGenerated:
			lbl = "generation"
		}

		if stage != targetStage {
			failOnError(t, lbl, hasError, errors)
			return true
		}

		if failOnSuccessFlag {
			failOnSuccess(t, lbl, hasError, errors)
		} else {
			failOnError(t, lbl, hasError, errors)
		}
		return false
	}
}

func CompileFiles(
	t *testing.T,
	basefs embed.FS,
	basepath string,
	filefilter func(string) bool,
	outputhandler func(fn string, out string) error,
	errhandler ErrorHandler,
) {
	t.Helper()

	files := getFiles(basefs, basepath, filefilter)

	for i := range files {
		file := files[i]
		nobuiltinElement := strings.Contains(file, "nobuiltin_element")
		nobuiltinTypes := strings.Contains(file, "nobuiltin_type")

		var flag transpiler.Flag
		if nobuiltinElement {
			flag |= transpiler.FlagNoBuiltinElement
		}

		if nobuiltinTypes {
			flag |= transpiler.FlagNoBuiltinType
		}

		buf := readFile(basefs, file)

		opts := []CompilationOption{
			SetName(file),
			SetResolverFlag(flag),
			SetResultHandler(outputhandler),
			SetErrorHandler(errhandler),
		}

		CompileSource(t, buf, opts...)
	}
}

func CompileSource(t *testing.T, buf []byte, opts ...CompilationOption) {
	t.Helper()

	ctx := NewCompilationContext(opts...)
	name := ctx.name

	t.Run(name, func(t *testing.T) {
		tok := token.Scan(buf)

		astp := parser.ParseFile(tok)
		if !ctx.errhandler(t, StageParsed, astp.HasError(), newErrorSeq(astp.Errors)) {
			return
		}

		nodes := transpiler.ParseFile(astp, tok)
		if !ctx.errhandler(t, StageTransformed, nodes.HasError(), nodes.Errors()) {
			return
		}

		env := transpiler.ResolveFile(nodes, ctx.resolverFlag)
		if !ctx.errhandler(t, StageResolved, nodes.HasError(), nodes.Errors()) {
			return
		}

		_ = transpiler.TypecheckFile(nodes, env)
		if !ctx.errhandler(t, StageTypechecked, nodes.HasError(), nodes.Errors()) {
			return
		}

		w := strings.Builder{}
		gonodes := gotranspiler.Parse(nodes)
		err := codegen.Generate(&w, &gonodes)
		if !ctx.errhandler(t, StageGenerated, err != nil, errorSeqFunc(err)) {
			return
		}

		if err = ctx.resultHandler(name, w.String()); err != nil {
			t.Error(err)
		}
	})
}

func errorSeqFunc(err error) ErrorSeq {
	return func(yield func(int, perrors.Error) bool) {
		if err == nil {
			return
		}
		perr := perrors.Error{Message: err.Error()}
		yield(0, perr)
	}
}

func readFile(dir embed.FS, f string) []byte {
	buf, err := dir.ReadFile(f)
	if err != nil {
		panic(err)
	}

	return buf
}

func getFiles(dir embed.FS, base string, filter func(string) bool) []string {
	entries, err := dir.ReadDir(base)
	if err != nil {
		panic("could not open testdata dir")
	}

	validFiles := []string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		if n := e.Name(); filter(n) {
			qname := path.Join(base, e.Name())
			validFiles = append(validFiles, qname)
		}
	}

	return validFiles
}

func failOnError(t *testing.T, label string, hasError bool, errs ErrorSeq) {
	t.Helper()
	for _, err := range errs {
		t.Error(err)
	}
	if hasError {
		t.Fatalf("%s failed unexpectedly", label)
	}
}

func failOnSuccess(t *testing.T, label string, hasError bool, _ ErrorSeq) {
	t.Helper()
	if !hasError {
		t.Fatalf("%s succeeded unexpectedly", label)
	}
}

func newErrorSeq(errs []perrors.Error) ErrorSeq {
	return func(yield func(int, perrors.Error) bool) {
		for i, err := range errs {
			if !yield(i, err) {
				break
			}
		}
	}
}
