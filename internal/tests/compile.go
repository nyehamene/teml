package tests

import (
	"testing"

	"github.com/eml-lang/teml/token"
	transpiler "github.com/eml-lang/teml/transpiler"
)

type CompilationStage int
type CompilationOption func(*CompilationContext)

const (
	StageTokenized CompilationStage = iota
	StageParsed
	StageTransformed
	StageResolved
	StageTypeChecked
	StageGenerated
)

type CompilationContext struct {
	name          string
	stage         CompilationStage
	resultHandler ResultHandler
	errhandler    ErrorHandler
	resolverFlag  transpiler.Flag
	tokenizerFlag token.Flag
	parserFlag    token.Flag
}

func NewCompilationContext(opts ...CompilationOption) CompilationContext {
	ctx := CompilationContext{
		name:          "<unnamed>",
		stage:         StageParsed,
		resultHandler: func(string, CompilationStage, any) error { return nil },
		errhandler:    func(*testing.T, CompilationStage, bool, ErrorSeq) bool { return true },
		resolverFlag:  0,
		tokenizerFlag: 0,
		parserFlag:    0,
	}

	for _, opt := range opts {
		opt(&ctx)
	}

	return ctx
}
