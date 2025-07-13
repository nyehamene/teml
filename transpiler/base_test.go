package ast_test

import (
	"embed"
	"path"
	"strings"
	"testing"

	"github.com/eml-lang/teml/token"

	parser "github.com/eml-lang/teml/ast"
	perrors "github.com/eml-lang/teml/internal/errors"
	transpiler "github.com/eml-lang/teml/transpiler"
)

type CompilationStage int

const (
	ResolutionStage CompilationStage = iota
	TypecheckerStage
)

func readFile(dir embed.FS, f string) []byte {
	buf, err := dir.ReadFile(f)
	if err != nil {
		panic(err)
	}

	return buf
}

func getValidFiles(dir embed.FS, basepath string) []string {
	return getFiles(dir, basepath, func(name string) bool {
		return strings.HasPrefix(name, "valid")
	})
}

func getInvalidFiles(dir embed.FS, basepath string) []string {
	return getFiles(dir, basepath, func(name string) bool {
		return !strings.HasPrefix(name, "valid")
	})
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

func checkValidFiles(t *testing.T, fs embed.FS, path string, stage CompilationStage) {
	t.Helper()

	files := getValidFiles(fs, path)
	if len(files) == 0 {
		t.Fatal("no valid source file to test")
	}

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

		t.Run(file, func(t *testing.T) {
			buf := readFile(fs, file)
			tok := token.Scan(buf, 0)

			astp := parser.ParseFile(tok, 0)
			failOnError(t, "parsing", astp.HasError(), astp.Errors.Each())

			nodes := transpiler.ParseFile(astp, tok)
			failOnError(t, "tranformation", nodes.HasError(), nodes.Errors())

			env := transpiler.ResolveFile(nodes, flag)
			failOnError(t, "resolution", nodes.HasError(), nodes.Errors())

			if stage == ResolutionStage {
				return
			}

			_ = transpiler.TypecheckFile(nodes, env)
			failOnError(t, "type checker", nodes.HasError(), nodes.Errors())
		})
	}
}

func checkInvalidFiles(t *testing.T, fs embed.FS, path string, stage CompilationStage) {
	t.Helper()

	files := getInvalidFiles(fs, path)
	if len(files) == 0 {
		t.Fatal("no invalid source file to test")
	}

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

		t.Run(file, func(t *testing.T) {
			buf := readFile(fs, file)
			tok := token.Scan(buf, 0)

			astp := parser.ParseFile(tok, 0)
			failOnError(t, "parsing", astp.HasError(), astp.Errors.Each())

			nodes := transpiler.ParseFile(astp, tok)
			failOnError(t, "transformation", nodes.HasError(), nodes.Errors())

			env := transpiler.ResolveFile(nodes, flag)

			if stage == ResolutionStage {
				failOnSuccess(t, "resolution", nodes.HasError())
				return
			} else {
				failOnError(t, "resolution", nodes.HasError(), nodes.Errors())
			}

			_ = transpiler.TypecheckFile(nodes, env)
			failOnSuccess(t, "type checker", nodes.HasError())
		})
	}
}

func failOnError(t *testing.T, label string, hasError bool, errs func(yield func(int, perrors.Error) bool)) {
	t.Helper()
	for _, err := range errs {
		t.Error(err)
	}
	if hasError {
		t.Fatalf("%s failed unexpectedly", label)
	}
}

func failOnSuccess(t *testing.T, label string, hasError bool) {
	t.Helper()
	if !hasError {
		t.Fatalf("%s succeeded unexpectedly", label)
	}
}
