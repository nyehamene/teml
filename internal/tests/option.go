package tests

import (
	"github.com/eml-lang/teml/token"
	transpiler "github.com/eml-lang/teml/transpiler"
)

func SetName(name string) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.name = name
	}
}

func SetStage(stage CompilationStage) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.stage = stage
	}
}

func SetResultHandler(h ResultHandler) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.resultHandler = h
	}
}

func SetErrorHandler(h ErrorHandler) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.errhandler = h
	}
}

func SetResolverFlag(f transpiler.Flag) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.resolverFlag = f
	}
}

func SetTokenizerFlag(f token.Flag) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.tokenizerFlag = f
	}
}

func SetParserFlag(f token.Flag) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.parserFlag = f
	}
}
