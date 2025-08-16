package generator

import (
	"io"

	"github.com/tel-lang/tel/go/transpiler"
)

const (
	ContextStructName       = "RenderContext"
	ContextVarName          = "ctx"
	ContextFieldName        = "ctx"
	ContextOptionStructName = "RenderContextOption"
	ContextParameterName    = ContextVarName

	WriterFieldName     = "writer"
	WriterVarName       = "w"
	WriterParameterName = WriterVarName

	AttributesFieldName     = "attributes"
	AttributesParameterName = "attrs"

	ChildrenFieldName     = "children"
	ChildrenParameterName = ChildrenFieldName

	ComponentStructName   = "Component"
	FuncComponentTypeName = "ComponentFunc"

	RenderMethodName = "Render"
)

func WriteTel(w io.Writer, pkg string) error {
	imports := []string{
		`"context"`,
		`"io"`,
		`"fmt"`,
	}

	r := runtime{
		PackageName:             pkg,
		Imports:                 imports,
		ContextOptionStructName: ContextOptionStructName,
		ContextVarName:          ContextVarName,
		ContextStructName:       ContextStructName,
		ContextFieldName:        ContextFieldName,
		WriterFieldName:         WriterFieldName,
		AttributesFieldName:     AttributesFieldName,
		ChildrenFieldName:       ChildrenFieldName,
		FuncComponentTypeName:   FuncComponentTypeName,
		OptionSetters: []optionSetter{
			{
				ParameterName: "cc",
				ParameterType: "context.Context",
				FieldName:     ContextFieldName,
				FuncName:      "SetCtx",
			},
			{
				ParameterName: WriterParameterName,
				ParameterType: "io.Writer",
				FieldName:     WriterFieldName,
				FuncName:      "SetWriter",
			},
			{
				ParameterName: AttributesParameterName,
				ParameterType: "map[string]string",
				FieldName:     AttributesFieldName,
				FuncName:      "SetAttributes",
			},
			{
				ParameterName: ChildrenParameterName,
				ParameterType: "[]" + ComponentStructName,
				FieldName:     ChildrenFieldName,
				FuncName:      "setChildren",
			},
		},
		GlobalVars:          []globalVars{},
		ComponentStructName: ComponentStructName,
	}

	if err := writeRuntime(w, r); err != nil {
		return err
	}
	return nil

}

func WriteNamespace(w io.Writer, src transpiler.File) error {
	n := namespace{
		writer:               NewSourceWriter(w),
		ContextParameterName: ContextParameterName,
		ContextFieldName:     ContextFieldName,
		AttributesFieldName:  AttributesFieldName,
		ChildrenFieldName:    ChildrenFieldName,
		WriterFieldName:      WriterFieldName,
		ContextVarName:       ContextVarName,
		WriterVarName:        WriterVarName,
		ContextStructName:    ContextStructName,
		RenderMethodName:     RenderMethodName,
		ComponentFunc:        FuncComponentTypeName,
	}

	Init(n)

	if err := writeNamespace(n, src); err != nil {
		return err
	}
	return nil
}
