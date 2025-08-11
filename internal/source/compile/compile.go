package compile

import (
	"fmt"
	"strings"

	"github.com/eml-lang/teml/internal/source"
	"github.com/eml-lang/teml/token"

	parser "github.com/eml-lang/teml/ast"
	codegen "github.com/eml-lang/teml/codegen/go/generator"
	gotranspiler "github.com/eml-lang/teml/codegen/go/transpiler"
	transpiler "github.com/eml-lang/teml/transpiler"
)

type ErrorHandler func(stage CompilationStage, errors ErrorSeq)

type ResultHandler func(result CompilationResult)

type CompilationResult interface {
	compile()
}

func (TokenizedResult) compile()   {}
func (ParsedResult) compile()      {}
func (TranspiledResult) compile()  {}
func (ResolvedResult) compile()    {}
func (TypeCheckedResult) compile() {}
func (GeneratedResult) compile()   {}

type TokenizedResult struct {
	File *token.File
}

type ParsedResult struct {
	File *parser.File
}

type TranspiledResult struct {
	File *transpiler.File
}

type ResolvedResult struct {
	File *transpiler.File
	Env  transpiler.NameEnv
}

type TypeCheckedResult struct {
	File *transpiler.File
	Env  transpiler.TypeEnv
}

type GeneratedResult struct {
	File *gotranspiler.File
}

type CompilationOption func(*CompilationContext)

type CompilationContext struct {
	name            string
	stage           CompilationStage
	errhandler      ErrorHandler
	resultHandler   ResultHandler
	resolverFlag    transpiler.Flag
	typecheckerFlag transpiler.Flag
	tokenizerFlag   token.Flag
	parserFlag      token.Flag
	filterFunc      func(string) bool
}

type CompilationStage int

const (
	StageTokenized CompilationStage = iota
	StageParsed
	StageTranspiled
	StageResolved
	StageTypeChecked
	StageGenerated
)

func NewCompilationContext(opts ...CompilationOption) CompilationContext {
	ctx := CompilationContext{
		name:          "<unnamed>",
		stage:         StageGenerated,
		errhandler:    func(CompilationStage, ErrorSeq) {},
		resultHandler: func(CompilationResult) {},
		resolverFlag:  0,
		tokenizerFlag: 0,
		parserFlag:    0,
		filterFunc:    func(string) bool { return true },
	}

	for _, opt := range opts {
		opt(&ctx)
	}

	return ctx
}

func Dir(ctx CompilationContext, basepath string) {
	files := getFiles(ctx, basepath)
	for i := range files {
		file := files[i]
		buf := readFile(file)
		src := source.NewFile(file, buf)
		Source(ctx, src)
	}
}

func Source(ctx CompilationContext, file source.File) CompilationResult {
	const stageFinished CompilationStage = -2
	const stageStart CompilationStage = -1

	var currentStage = stageStart
	var nextStage = StageTokenized

	var tokenizedFile *token.File
	var parsedFile *parser.File
	var transpiledFile *transpiler.File
	var generatedFile gotranspiler.File
	var nameEnv transpiler.NameEnv
	var typeEnv transpiler.TypeEnv
	var result CompilationResult

continueStage:
	if currentStage == ctx.stage {
		nextStage = stageFinished
	} else if nextStage != stageFinished {
		if currentStage != stageStart {
			ctx.resultHandler(result)
		}
		currentStage = nextStage
	}

	switch nextStage {
	case StageTokenized:
		tokenizedFile = token.ScanInput(file, ctx.tokenizerFlag)
		result = TokenizedResult{File: tokenizedFile}
		// ctx.resultHandler(result)
		nextStage = StageParsed
		goto continueStage

	case StageParsed:
		parsedFile = parser.ParseFile0(tokenizedFile, ctx.parserFlag)
		if parsedFile.HasError() {
			ctx.errhandler(currentStage, newErrorSeq(parsedFile.Errors))
			return nil
		}
		result = ParsedResult{File: parsedFile}
		// ctx.resultHandler(result)
		nextStage = StageTranspiled
		goto continueStage

	case StageTranspiled:
		transpiledFile = transpiler.ParseFile(parsedFile)
		if transpiledFile.HasError() {
			ctx.errhandler(currentStage, transpiledFile.Errors())
			return nil
		}
		result = TranspiledResult{File: transpiledFile}
		// ctx.resultHandler(result)
		nextStage = StageResolved
		goto continueStage

	case StageResolved:
		nameEnv = transpiler.ResolveFile(transpiledFile, ctx.resolverFlag)
		if transpiledFile.HasError() {
			ctx.errhandler(currentStage, transpiledFile.Errors())
			return nil
		}
		result = ResolvedResult{File: transpiledFile, Env: nameEnv}
		// ctx.resultHandler(result)
		nextStage = StageTypeChecked
		goto continueStage

	case StageTypeChecked:
		typeEnv = transpiler.TypecheckFile0(transpiledFile, nameEnv, ctx.typecheckerFlag)
		if transpiledFile.HasError() {
			ctx.errhandler(currentStage, transpiledFile.Errors())
			return nil
		}
		result = TypeCheckedResult{File: transpiledFile, Env: typeEnv}
		// ctx.resultHandler(result)
		nextStage = StageGenerated
		goto continueStage

	case StageGenerated:
		w := strings.Builder{}
		generatedFile = gotranspiler.ParseFile0(transpiledFile)
		err := codegen.Generate0(&w, &generatedFile)
		if err != nil {
			ctx.errhandler(currentStage, newErrorSeq1(err))
			return nil
		}
		result = GeneratedResult{File: &generatedFile}
		// ctx.resultHandler(result)
		nextStage = stageFinished
		goto continueStage

	case stageFinished:
		return result

	default:
		panic(fmt.Sprintf("unexpected compile.CompilationStage: %#v", currentStage))
	}
}
