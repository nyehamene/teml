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

type ErrorHandler func(t *testing.T, ctx CompilationContext) bool
type ErrorSeq = func(yield func(int, perrors.Error) bool)

type CompilationStage int

const (
	StageParsed CompilationStage = iota
	StageTransformed
	StageResolved
	StageTypechecked
	StageGenerated
)

type CompilationContext struct {
	stage    CompilationStage
	hasError bool
	errseq   ErrorSeq
}

func (ctx CompilationContext) Stage() CompilationStage {
	return ctx.stage
}

func (ctx CompilationContext) HasError() bool {
	return ctx.hasError
}

func (ctx CompilationContext) Errors() ErrorSeq {
	return ctx.errseq
}

func NewErrorHandler(targetStage CompilationStage, failOnSuccessFlag bool) ErrorHandler {
	return func(t *testing.T, ctx CompilationContext) bool {
		lbl := "unknown stage"

		switch ctx.stage {
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

		if ctx.stage != targetStage {
			failOnError(t, lbl, ctx.hasError, ctx.errseq)
			return true
		}

		if failOnSuccessFlag {
			failOnSuccess(t, lbl, ctx.hasError, ctx.errseq)
		} else {
			failOnError(t, lbl, ctx.hasError, ctx.errseq)
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

		t.Run(file, func(t *testing.T) {
			tok := token.Scan(buf, 0)

			astp := parser.ParseFile(tok, 0)
			if !errhandler(t, CompilationContext{stage: StageParsed, hasError: astp.HasError(), errseq: astp.Errors.Each()}) {
				return
			}

			nodes := transpiler.ParseFile(astp, tok)
			if !errhandler(t, CompilationContext{stage: StageTransformed, hasError: nodes.HasError(), errseq: nodes.Errors()}) {
				return
			}

			env := transpiler.ResolveFile(nodes, flag)
			if !errhandler(t, CompilationContext{stage: StageResolved, hasError: nodes.HasError(), errseq: nodes.Errors()}) {
				return
			}

			_ = transpiler.TypecheckFile(nodes, env)
			if !errhandler(t, CompilationContext{stage: StageTypechecked, hasError: nodes.HasError(), errseq: nodes.Errors()}) {
				return
			}

			w := strings.Builder{}
			gonodes := gotranspiler.Parse(nodes)
			err := codegen.Generate(&w, &gonodes)

			if err := outputhandler(file, w.String()); err != nil {
				t.Error(err)
			}

			if !errhandler(t, CompilationContext{stage: StageGenerated, hasError: err != nil, errseq: errorSeqFunc(err)}) {
				return
			}
		})
	}
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
