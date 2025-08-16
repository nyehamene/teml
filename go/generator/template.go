package generator

import (
	"io"
	"text/template"

	"github.com/tel-lang/tel/go/transpiler"
)

const packageDeclarationTemplateSource = "package {{.Name}}\n"

const importDeclarationTemplateSource = `
{{- if .Imports -}}
import (
	{{- range .Imports}}
	{{.Name}} {{.Path}}
	{{- end}}
)
{{- end -}}
`

const blankVarTemplateSource = "_ = {{.VarName}}\n"

const stringLiteralTemplateSource = `
		{{.ErrorVarName}} := writeString( {{- .ContextVarName -}} . {{- .WriterFieldName}}, {{.Value -}} )
`

type stringExprTemplateModel struct {
	ErrorVarName    string
	ContextVarName  string
	WriterFieldName string
	Value           string
}

type blankVarTemplateModel struct {
	VarName string
}

const callRenderMethodTemplateSource = `
		{{.ErrorVarName}} := {{.ReceiverVarName}}.{{.RenderMethodName -}} ( {{- .ContextVarName -}} )
`

type callRenderMethodTemplateModel struct {
	ErrorVarName     string
	RenderMethodName string
	ReceiverVarName  string
	ContextVarName   string
}

type importDeclarationTemplateModel struct {
	Imports []transpiler.Import
}

const structDefinitionTemplateSource = `
type {{.Name}} struct {
	{{- range $field := .Fields }}
	{{.Name}} {{.Type.Name}}
	{{- end}}
}
`

const enumDefinitionTemplateSource = `
type {{.TypeName}} string

const (
	{{- $TypeName := .TypeName}}
	{{- range $constant := .Constants }}
	{{$TypeName -}} {{ResolveEnumConstantAsName $constant}} {{$TypeName}} = {{$constant}}
	{{- end}}
)
`

const renderMethodTemplateSource = `
func ( {{- .ComponentMethodReceiverName}} {{.ComponentTypeName -}} ) {{.MethodName -}} ( {{- .ContextParameterName}} {{.ContextStructName -}} ) error {
	{{- range $stmt := .Stmts}}
	{{- writeStmt $stmt}}
	{{- end}}
}
`

type renderMethodTemplateModel struct {
	ComponentMethodReceiverName string
	ComponentTypeName           string
	ContextParameterName        string
	ContextStructName           string
	MethodName                  string
	Stmts                       []transpiler.Stmt
}

const inherittedChildrenTemplateSource = `
		{{.ChildContextVar}} := {{.ContextStructName -}} {
			{{.ContextFieldName -}} : {{.ContextParameterName -}} . {{- .ContextFieldName}},
			{{.WriterFieldName -}} : {{.ContextParameterName -}} . {{- .WriterFieldName}},
		}

		for _, child := range {{.ContextParameterName -}} . {{- .ChildrenFieldName}} {
			if err := child.{{.RenderMethodName -}} ( {{- .ChildContextVar -}} ); err != nil {
				return err
			}
		}
`

type inherittedChildrenTemplateModel struct {
	ChildContextVar      string
	ContextStructName    string
	ContextFieldName     string
	ContextParameterName string
	WriterFieldName      string
	ChildrenFieldName    string
	RenderMethodName     string
}

const copyContenxtTemplateSource = `
		{{.VarName}} := {{.ContextStructName -}} {
			{{.ContextFieldName -}} : {{.ContextParameterName -}} . {{- .ContextFieldName}},
			{{.WriterFieldName -}} : {{.ContextParameterName -}} . {{- .WriterFieldName}},
			{{.AttributesFieldName -}} : {{.AttributesVarName}},
			{{.ChildrenFieldName -}} : {{.ChildrenVarName}},
		}
`

type copyContextTemplateModel struct {
	ContextStructName    string
	ContextFieldName     string
	ContextParameterName string
	WriterFieldName      string
	AttributesFieldName  string
	ChildrenFieldName    string
	AttributesVarName    string
	ChildrenVarName      string
	VarName              string
}

const returnIfNotNilTemplateSource = `
		if {{.VarName}} != nil {
			return {{.VarName}}
		}
`

type returnIfNotNilTemplateSourceModel struct {
	VarName string
}

const mapInstanceTemplateSource = `
		{{.VarName}} := map[string]string{
			{{- range .Entries}}
			{{ResolveMapKey .Key -}} : {{- ResolveValue .Value}},
			{{- end}}
		}
`

type mapInstanceTemplateModel struct {
	VarName string
	Entries []transpiler.SetMapEntry
}

const numberMemberAccessExprTemplateSource = `
		{{.ErrorVarName}} := writeNumber( {{- .ContextVarName -}} . {{- .WriterFieldName}}, {{.Value -}} )
`

type numberMemberAccessExprTemplateModel struct {
	Value           string
	ErrorVarName    string
	ContextVarName  string
	WriterFieldName string
}

const inherittedAttributesTemplateSource = `
		for key, val := range {{.ContextVarName}}.{{.AttributesVarName}} {
			attr := formatString(" %s=%s", key, val)
			if {{.ErrorVarName}} := writeString( {{- .ContextVarName -}} . {{- .WriterFieldName -}}, attr); {{.ErrorVarName}} != nil {
				return {{.ErrorVarName}}
			}
		}
`

type inherittedAttributesTemplateModel struct {
	ContextVarName    string
	AttributesVarName string
	WriterFieldName   string
	ErrorVarName      string
}

const ifStmtTemplateSource = `
		if {{.Cond}} {
			{{- range $stmt := .Stmts}}
			{{writeStmt $stmt}}
			{{- end}}
		} {{- if .Else }} else {
			{{- range $stmt := .Else}}
			{{writeStmt $stmt}}
			{{- end}}
		} {{- end}}
`

type ifStmtTemplateModel struct {
	Cond  string
	Stmts []transpiler.Stmt
	Else  []transpiler.Stmt
}

const condStmtTemplateSource = `
	switch {{.Cond}} {
		{{- range .Cases}}
	case {{.Match -}} :
		{{- range $stmt := .Branch -}}
		{{writeStmt $stmt}}
		{{- end}}
		{{- end}}
	default:
		panic(formatString("unexpected enum value %s", {{.Cond -}} ))
	}
	`

type condStmtTemplateModel struct {
	Cond  string
	Cases []transpiler.Case
}

const structInstanceTemplateSource = `
		{{.VarName}} := {{.TypeName -}} {
			{{- range .Fields}}
			{{.Name -}} : {{.Value}},
			{{- end}}
		}
		`

type structInstanceTemplateModel struct {
	VarName  string
	TypeName string
	Fields   []transpiler.SetStructField
}

const sliceInstanceTemplateSource = `
		{{.VarName}} := []{{.TypeName -}} {
			{{- range .Values}}
			{{resolve .}},
			{{- end}}
		}
		`

type sliceInstanceTemplateModel struct {
	VarName  string
	TypeName string
	Values   []transpiler.Var
}

const componentInstanceTemplateSource = `
		var {{.VarName}} {{.TypeName}} = func() error {
			{{- range $stmt := .Stmts}}
			{{writeStmt $stmt}}
			{{- end}}
		}
		`

type componentInstanceTemplateModel struct {
	VarName  string
	TypeName string
	Stmts    []transpiler.Stmt
}

var (
	packageDeclarationTemplate,
	importDeclarationTemplate,
	structDefinitionTemplate,
	enumDefinitionTemplate,
	renderMethodTemplate,
	inherittedChildrenTemplate,
	copyContextTemplate,
	returnIfNotNilTemplate,
	mapInstanceTemplate,
	numberMemberAccessExprTemplate,
	inherittedAttributesTemplate,
	ifStmtTemplate,
	condStmtTemplate,
	structInstanceTemplate,
	sliceInstanceTemplate,
	callRenderMethodTemplate,
	blankVarTemplate,
	stringExprTemplate,
	componentInstanceTemplate *template.Template
)

func Init(n namespace) {
	packageDeclarationTemplate = template.Must(template.New("package_decl").Parse(packageDeclarationTemplateSource))
	importDeclarationTemplate = template.Must(template.New("import_decl").Parse(importDeclarationTemplateSource))
	structDefinitionTemplate = template.Must(template.New("struct_def").Parse(structDefinitionTemplateSource))
	inherittedChildrenTemplate = template.Must(template.New("inheritted_children").Parse(inherittedChildrenTemplateSource))
	copyContextTemplate = template.Must(template.New("copy_context").Parse(copyContenxtTemplateSource))
	numberMemberAccessExprTemplate = template.Must(template.New("number_member_access_expr").Parse(numberMemberAccessExprTemplateSource))
	inherittedAttributesTemplate = template.Must(template.New("inheritted_attributes").Parse(inherittedAttributesTemplateSource))
	structInstanceTemplate = template.Must(template.New("struct_instance").Parse(structInstanceTemplateSource))
	returnIfNotNilTemplate = template.Must(template.New("return_if_nil").Parse(returnIfNotNilTemplateSource))
	callRenderMethodTemplate = template.Must(template.New("call_render_method").Parse(callRenderMethodTemplateSource))
	blankVarTemplate = template.Must(template.New("blank_var").Parse(blankVarTemplateSource))
	stringExprTemplate = template.Must(template.New("string_literal").Parse(stringLiteralTemplateSource))

	enumDefinitionTemplate = template.New("enum_def")
	enumDefinitionTemplate.Funcs(map[string]any{
		"ResolveEnumConstantAsName": transpiler.ResolveEnumConstantAsName,
	})
	enumDefinitionTemplate = template.Must(enumDefinitionTemplate.Parse(enumDefinitionTemplateSource))

	renderMethodTemplate = template.New("render_method")
	renderMethodTemplate.Funcs(map[string]any{
		"writeStmt": n.writeStmt,
	})
	renderMethodTemplate = template.Must(renderMethodTemplate.Parse(renderMethodTemplateSource))

	ifStmtTemplate = template.New("if_stmt")
	ifStmtTemplate.Funcs(map[string]any{
		"writeStmt": n.writeStmt,
	})
	ifStmtTemplate = template.Must(ifStmtTemplate.Parse(ifStmtTemplateSource))

	condStmtTemplate = template.New("cond_stmt")
	condStmtTemplate.Funcs(map[string]any{
		"writeStmt": n.writeStmt,
	})
	condStmtTemplate = template.Must(condStmtTemplate.Parse(condStmtTemplateSource))

	componentInstanceTemplate = template.New("component_instance")
	componentInstanceTemplate.Funcs(map[string]any{
		"writeStmt": n.writeStmt,
	})
	componentInstanceTemplate = template.Must(componentInstanceTemplate.Parse(componentInstanceTemplateSource))

	sliceInstanceTemplate = template.New("slice_stmt")
	sliceInstanceTemplate.Funcs(map[string]any{
		"resolve": transpiler.ResolveValue,
	})
	sliceInstanceTemplate = template.Must(sliceInstanceTemplate.Parse(sliceInstanceTemplateSource))

	mapInstanceTemplate = template.New("map_instance")
	mapInstanceTemplate.Funcs(map[string]any{
		"ResolveMapKey": transpiler.ResolveMapKey,
		"ResolveValue":  transpiler.ResolveValue,
	})
	mapInstanceTemplate = template.Must(mapInstanceTemplate.Parse(mapInstanceTemplateSource))
}

func (n namespace) executeStringExprTemplate(w io.Writer, model stringExprTemplateModel) error {
	if err := stringExprTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeStringLiteralTemplate(w io.Writer, node transpiler.StringLiteral) error {
	model := stringExprTemplateModel{
		ErrorVarName:    string(node.Error),
		Value:           string(node.Value),
		ContextVarName:  n.ContextParameterName,
		WriterFieldName: n.WriterFieldName,
	}
	return n.executeStringExprTemplate(w, model)
}

func (n namespace) executeStringMemberAccessTemplate(w io.Writer, node transpiler.StringMemberAccessExpr) error {
	model := stringExprTemplateModel{
		ErrorVarName:    string(node.Error),
		Value:           string(node.Value),
		ContextVarName:  n.ContextParameterName,
		WriterFieldName: n.WriterFieldName,
	}
	return n.executeStringExprTemplate(w, model)
}

func (n namespace) executeStructDefinitionTemplate(w io.Writer, node transpiler.Struct) error {
	if err := structDefinitionTemplate.Execute(w, node); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeEnumDefinitionTemplate(w io.Writer, fields []transpiler.StructField) error {
	var enums []transpiler.Enum
	for _, field := range fields {
		switch t := field.Type.(type) {
		case transpiler.Enum:
			enums = append(enums, t)
		}
	}

	if len(enums) == 0 {
		return nil
	}

	for _, e := range enums {
		if err := enumDefinitionTemplate.Execute(w, e); err != nil {
			return err
		}
	}

	return nil
}

func (n namespace) executeBlankVarTemplate(w io.Writer, node transpiler.BlankVar) error {
	model := blankVarTemplateModel{
		VarName: string(node),
	}
	if err := blankVarTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeCallRenderMethodTemplate(w io.Writer, node transpiler.CallRenderMethod) error {
	model := callRenderMethodTemplateModel{
		ErrorVarName:     string(node.Error),
		ReceiverVarName:  string(node.Receiver),
		ContextVarName:   string(node.Context),
		RenderMethodName: n.RenderMethodName,
	}
	if err := callRenderMethodTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeImportDeclarationTemplate(w io.Writer, i []transpiler.Import) error {
	model := importDeclarationTemplateModel{
		Imports: i,
	}
	if err := importDeclarationTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeRenderMethodTemplate(w io.Writer, m transpiler.RenderMethod) error {
	model := renderMethodTemplateModel{
		ComponentMethodReceiverName: string(m.Receiver),
		ComponentTypeName:           string(m.Type),
		Stmts:                       m.Body,
		ContextParameterName:        n.ContextParameterName,
		ContextStructName:           n.ContextStructName,
		MethodName:                  n.RenderMethodName,
	}
	if err := renderMethodTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeInherittedChildrenTemplate(w io.Writer, m transpiler.InheritChildren) error {
	model := inherittedChildrenTemplateModel{
		ChildContextVar:      string(m.Context),
		ContextStructName:    n.ContextStructName,
		ContextFieldName:     n.ContextFieldName,
		ContextParameterName: n.ContextParameterName,
		WriterFieldName:      n.WriterFieldName,
		ChildrenFieldName:    n.ChildrenFieldName,
		RenderMethodName:     n.RenderMethodName,
	}
	if err := inherittedChildrenTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeCopyContextTemplate(w io.Writer, m transpiler.CopyRenderContext) error {
	model := copyContextTemplateModel{
		ContextStructName:    n.ContextStructName,
		ContextFieldName:     n.ContextFieldName,
		ContextParameterName: n.ContextParameterName,
		WriterFieldName:      n.WriterFieldName,
		AttributesFieldName:  n.AttributesFieldName,
		ChildrenFieldName:    n.ChildrenFieldName,
		AttributesVarName:    string(m.Attrs),
		ChildrenVarName:      string(m.Children),
		VarName:              string(m.Variable),
	}
	if err := copyContextTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeReturnIfNotNilTemplate(w io.Writer, m transpiler.ReturnIfNotNil) error {
	model := returnIfNotNilTemplateSourceModel{
		VarName: string(m),
	}
	if err := returnIfNotNilTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeMapInstanceTemplate(w io.Writer, node transpiler.MapInstance) error {
	model := mapInstanceTemplateModel{
		VarName: string(node.Variable),
		Entries: node.Entries,
	}
	if err := mapInstanceTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeNumberMemberAccessExprTemplate(w io.Writer, m transpiler.NumberMemberAccessExpr) error {
	model := numberMemberAccessExprTemplateModel{
		Value:           string(m.Value),
		ErrorVarName:    string(m.Error),
		ContextVarName:  n.ContextVarName,
		WriterFieldName: n.WriterFieldName,
	}
	if err := numberMemberAccessExprTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeInherittedAttributes(w io.Writer, m transpiler.InheritAttributes) error {
	model := inherittedAttributesTemplateModel{
		ContextVarName:    n.ContextVarName,
		AttributesVarName: n.AttributesFieldName,
		WriterFieldName:   n.WriterFieldName,
		ErrorVarName:      string(m.Error.Name()),
	}
	if err := inherittedAttributesTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeIfStmtTemplate(w io.Writer, node transpiler.If) error {
	model := ifStmtTemplateModel{
		Cond:  transpiler.ResolveValue(node.Cond),
		Stmts: node.Then,
		Else:  node.Else,
	}
	if err := ifStmtTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeCondStmtTemplate(w io.Writer, m transpiler.Cond) error {
	model := condStmtTemplateModel{
		Cond:  transpiler.ResolveValue(m.Target),
		Cases: m.Cases,
	}
	if err := condStmtTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeStructInstanceTemplate(w io.Writer, m transpiler.StructInstance) error {
	model := structInstanceTemplateModel{
		VarName:  transpiler.ResolveValue(m.Variable),
		TypeName: transpiler.ResolveValue(m.Type),
		Fields:   m.Parameters,
	}
	if err := structInstanceTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeSliceInstanceTemplate(w io.Writer, m transpiler.SliceInstance) error {
	model := sliceInstanceTemplateModel{
		VarName:  transpiler.ResolveValue(m.Variable),
		TypeName: transpiler.ResolveValue(m.Type),
		Values:   m.Values,
	}
	if err := sliceInstanceTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}

func (n namespace) executeComponentInstanceTemplate(w io.Writer, m transpiler.ComponentInstance) error {
	model := componentInstanceTemplateModel{
		VarName:  transpiler.ResolveValue(m.Variable),
		TypeName: n.ComponentFunc,
		Stmts:    m.Stmts,
	}
	if err := componentInstanceTemplate.Execute(w, model); err != nil {
		return err
	}
	return nil
}
