package ast_test

import (
	_ "embed"
	"testing"

	"github.com/eml-lang/teml/internal/tests"
)

//go:embed testdata/source.teml
var source []byte

func TestParse(t *testing.T) {
	tests.CompileSource(t, source, tests.SetStage(tests.StageGenerated))
}
