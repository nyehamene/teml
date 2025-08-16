package tel_test

import (
	_ "embed"
	"io"
	"strings"
	"testing"

	"github.com/tel-lang/tel"
	"github.com/tel-lang/tel/go/generator"
	"github.com/tel-lang/tel/go/transpiler"
)

//go:embed grammar/example/app.tel
var examplefile []byte

func TestMain(t *testing.T) {
	telFile := &strings.Builder{}
	namespaceFile := &strings.Builder{}
	files := tel.NewFileSet("test")
	files.Add("app.tel", examplefile)
	run(t, telFile, namespaceFile, files)
}

func BenchmarkScan(b *testing.B) {
	telFile := &strings.Builder{}
	namespaceFile := &strings.Builder{}
	files := tel.NewFileSet("test")
	files.Add("app.tel", examplefile)

	for b.Loop() {
		run(b, telFile, namespaceFile, files)
	}
}

func BenchmarkScanReduceAllocTokenizer(b *testing.B) {
	telFile := &strings.Builder{}
	namespaceFile := &strings.Builder{}
	files := tel.NewFileSet("test")
	files.Add("app.tel", examplefile)

	for b.Loop() {
		run(b, telFile, namespaceFile, files)
	}
}

type runner interface {
	Helper()
	Fail()
	Error(...any)
	Fatal(...any)
}

func run(_ runner, tw, w io.Writer, fileset tel.FileSet) {
	ctx := tel.NewContext()
	src := transpiler.ParseFile(ctx, fileset)

	_ = generator.WriteTel(tw, src.Package)

	for _, file := range src.Files {
		_ = generator.WriteNamespace(w, file)
	}
}
