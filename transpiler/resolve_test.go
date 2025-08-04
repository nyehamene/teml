package ast

import (
	"embed"
	"fmt"
	"strings"

	"github.com/eml-lang/teml/ast"
	"github.com/eml-lang/teml/internal/source"
	"github.com/eml-lang/teml/token"
)

func TestResolverValid(t *testing.T) {
	const testdata = "./testdata/resolver/valid.teml"
	src, err := source.OpenFile(testdata)
	if err != nil {
		t.Fatal(err)
	}

	toks := token.ScanInput(src, token.PreserveComment)
	p_ast := ast.ParseFile(toks)
	for _, err := range p_ast.Errors {
		t.Error(err)
	}
	if p_ast.HasError() {
		t.Fatalf("parser failed unexpectedly")
	}

	t_ast := ParseFile(p_ast)
	for _, err := range t_ast.Errors() {
		t.Error(err)
	}
	if t_ast.HasError() {
		t.Fatal("transpiler failed unexpectedly")
	}

	r_env := ResolveFile(t_ast)
	for _, err := range t_ast.Errors() {
		t.Error(err)
	}
	if t_ast.HasError() {
		t.Fatal("transpiler failed unexpectedly")
	}

	expectedBindings := extractTestdataFromComment(toks)
	for _, binding := range expectedBindings {
		resolvedScope, ok := r_env.LookupName(binding.Scope)
		if !ok {
			t.Errorf("could not resolve scope %q", binding.Scope)
			continue
		}

		resolvedEnv, ok := r_env.LookupNameEnv(resolvedScope)
		if !ok {
			t.Errorf("could not resolve env %q", resolvedScope)
			continue
		}

		resolvedName, ok := resolvedEnv.LookupName(binding.Name)
		if !ok {
			t.Errorf("could not resolve name %q", binding.Name)
			continue
		}

		if binding.FQN != resolvedName {
			t.Errorf("expected %q", binding.FQN)
			t.Errorf("got %q", resolvedName)
		}
	}
}

func TestResolverInvalid(t *testing.T) {
	// nobuiltinElement := strings.Contains(file, "nobuiltin_element")
	// nobuiltinTypes := strings.Contains(file, "nobuiltin_type")

	// var flag Flag
	// if nobuiltinElement {
	// 	flag |= FlagNoNativeElement
	// }

	// if nobuiltinTypes {
	// 	flag |= FlagNoBuiltinType
	// }
}

type LookupOp struct {
	Scope string
	Name  string
	FQN   string
}

func extractTestdataFromComment(file *token.File) []LookupOp {
	var ops []LookupOp

	for _, tok := range file.Tokens {
		if tok.Kind != token.Comment {
			continue
		}

		// current token is a comment
		cmt, ok := file.Text(tok)
		if !ok {
			panic("could not get comment text")
		}

		const opMarker = ";lookup"
		if !strings.HasPrefix(cmt, opMarker) {
			continue
		}

		// comment format: lookup <scope> <name> <fqn>
		// example: ;lookup Home username views.Home.username
		cmt = strings.Replace(cmt, opMarker, "", 1)

		for data := range strings.SplitSeq(cmt, ";") {
			data = strings.TrimLeft(data, " ")
			chunks := strings.Split(data, " ")

			if explen, gotlen := 3, len(chunks); gotlen != explen {
				panic(fmt.Sprintf("invalid comment format. expected %d tokens got %d => %v", explen, gotlen, chunks))
			}

			scope := chunks[0]
			name := chunks[1]
			fqn := chunks[2]

			op := LookupOp{Scope: scope, Name: name, FQN: fqn}
			ops = append(ops, op)
		}
	}

	return ops
}
