package generator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/tel-lang/tel/go/transpiler"
)

type namespace struct {
	writer               SourceWriter
	ContextParameterName string

	ContextStructName   string
	ContextFieldName    string
	AttributesFieldName string
	ChildrenFieldName   string
	WriterFieldName     string

	ContextVarName string
	WriterVarName  string

	RenderMethodName string

	ComponentFunc string
}

func writeNamespace(n namespace, src transpiler.File) error {
	w := n.writer.Unwrap()

	if err := packageDeclarationTemplate.Execute(w, src.Package); err != nil {
		return err
	}

	if err := n.executeImportDeclarationTemplate(w, src.Imports); err != nil {
		return err
	}

	for _, node := range src.Structs {
		if err := n.executeStructDefinitionTemplate(w, node); err != nil {
			return err
		}
	}

	for _, node := range src.Structs {
		if err := n.executeEnumDefinitionTemplate(w, node.Fields); err != nil {
			return err
		}
	}

	for _, gvar := range src.GlobalVars {
		if err := n.writeGlobalVar(gvar); err != nil {
			return err
		}
	}

	for _, m := range src.Methods {
		if err := n.executeRenderMethodTemplate(w, m); err != nil {
			return err
		}
	}

	return nil
}

func (n namespace) writeStmt(stmt transpiler.Stmt) (string, error) {
	switch node := stmt.(type) {

	case transpiler.ReturnNil:
		source := "\treturn nil"
		return source, nil

	case transpiler.BlankVar:
		w := &strings.Builder{}
		err := n.executeBlankVarTemplate(w, node)
		return w.String(), err

	case transpiler.CallRenderMethod:
		w := &strings.Builder{}
		err := n.executeCallRenderMethodTemplate(w, node)
		return w.String(), err

	case transpiler.StringLiteral:
		w := &strings.Builder{}
		err := n.executeStringLiteralTemplate(w, node)
		return w.String(), err

	case transpiler.StringMemberAccessExpr:
		// TODO sanitize user input (t.Value)
		w := &strings.Builder{}
		err := n.executeStringMemberAccessTemplate(w, node)
		return w.String(), err

	case transpiler.NumberMemberAccessExpr:
		w := &strings.Builder{}
		err := n.executeNumberMemberAccessExprTemplate(w, node)
		return w.String(), err

	case transpiler.InheritChildren:
		w := &strings.Builder{}
		err := n.executeInherittedChildrenTemplate(w, node)
		return w.String(), err

	case transpiler.CopyRenderContext:
		w := &strings.Builder{}
		err := n.executeCopyContextTemplate(w, node)
		return w.String(), err

	case transpiler.ReturnIfNotNil:
		w := &strings.Builder{}
		err := n.executeReturnIfNotNilTemplate(w, node)
		return w.String(), err

	case transpiler.MapInstance:
		w := &strings.Builder{}
		err := n.executeMapInstanceTemplate(w, node)
		return w.String(), err

	case transpiler.InheritAttributes:
		w := &strings.Builder{}
		err := n.executeInherittedAttributes(w, node)
		return w.String(), err

	case transpiler.If:
		w := &strings.Builder{}
		err := n.executeIfStmtTemplate(w, node)
		return w.String(), err

	case transpiler.Cond:
		w := &strings.Builder{}
		err := n.executeCondStmtTemplate(w, node)
		return w.String(), err

	case transpiler.StructInstance:
		w := &strings.Builder{}
		err := n.executeStructInstanceTemplate(w, node)
		return w.String(), err

	case transpiler.SliceInstance:
		w := &strings.Builder{}
		err := n.executeSliceInstanceTemplate(w, node)
		return w.String(), err

	case transpiler.ComponentInstance:
		w := &strings.Builder{}
		err := n.executeComponentInstanceTemplate(w, node)
		return w.String(), err

	default:
		panic(fmt.Sprintf("unexpected stmt type: %v", reflect.TypeOf(stmt)))
	}
}

func (n namespace) writeGlobalVar(gvar transpiler.BlankVar) error {
	if err := n.writeln("var _ = " + string(gvar)); err != nil {
		return err
	}
	return nil
}

func (n namespace) write(s string) error {
	if _, err := n.writer.Write(s); err != nil {
		return err
	}
	return nil
}

func (n namespace) writeln(s string) error {
	if err := n.write(s + "\n"); err != nil {
		return err
	}
	return nil
}
