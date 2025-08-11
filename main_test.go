package main_test

import (
	_ "embed"
	"testing"

	"github.com/eml-lang/teml/internal/flags"
	"github.com/eml-lang/teml/internal/source"
	"github.com/eml-lang/teml/internal/source/compile"
)

//go:embed app.teml
var examplefile []byte

var sourceFile = source.File{
	Path:    "test.teml",
	Name:    "test",
	Content: examplefile,
}

func TestScanParse(t *testing.T) {
	toks := token.ScanInput(sourceFile)
	for _, tok := range toks.Tokens {
		if tok.Kind == token.Invalid {
			t.Fatal()
		}
	}

	astp := parser.ParseFile(toks)
	for _, err := range astp.Errors {
		t.Error(err.Message)
	}

	astn := transpiler.ParseFile(astp)
	for _, err := range astn.Errors() {
		t.Error(err)
	}

	transpiler.ResolveFile(astn)
	for _, err := range astn.Errors() {
		t.Error(err)
	}
}

func BenchmarkScan(b *testing.B) {
	for b.Loop() {
		parseFile(ctx)
	}
}

func BenchmarkScanReduceAllocTokenizer(b *testing.B) {
	ctx := compile.NewCompilationContext(
		compile.SetTokenizerFlag(flags.ReduceAlloc),
		compile.SetErrorHandler(errhandler),
	)
	for b.Loop() {
		parseFile(ctx)
	}
}

func BenchmarkScanReduceAllocParser(b *testing.B) {
	ctx := compile.NewCompilationContext(
		compile.SetParserFlag(flags.ReduceAlloc),
		compile.SetErrorHandler(errhandler),
	)
	for b.Loop() {
		parseFile(ctx)
	}
}

func BenchmarkScanReduceAlloc(b *testing.B) {
	ctx := compile.NewCompilationContext(
		compile.SetSourceFlag(flags.ReduceAlloc),
		compile.SetErrorHandler(errhandler),
	)
	for b.Loop() {
		parseFile()
	}
}

func parseFile() {
	ft := token.ScanInput(sourceFile)
	fa := parser.ParseFile(ft)
	fn := transpiler.ParseFile(fa)
	transpiler.ResolveFile(fn)
}
