package tests

import (
	"github.com/eml-lang/teml/internal/flags"
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

func SetResolverFlag(f flags.Flag) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.resolverFlag = f
	}
}

func SetTypecheckerFlag(f flags.Flag) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.typecheckerFlag = f
	}
}

func SetTranspilerFlag(f flags.Flag) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.resolverFlag = f
		ctx.typecheckerFlag = f
	}
}

func SetTokenizerFlag(f flags.Flag) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.tokenizerFlag = f
	}
}

func SetParserFlag(f flags.Flag) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.parserFlag = f
	}
}

func SetSourceFlag(f flags.Flag) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.tokenizerFlag = f
		ctx.parserFlag = f
	}
}

func SetFileFilter(f func(string) bool) CompilationOption {
	return func(ctx *CompilationContext) {
		ctx.filterFunc = f
	}
}
