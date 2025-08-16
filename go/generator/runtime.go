package generator

import (
	"io"
	"text/template"
)

type runtime struct {
	PackageName string
	Imports     []string

	ContextOptionStructName string

	ContextVarName string

	ContextStructName   string
	ContextFieldName    string
	WriterFieldName     string
	AttributesFieldName string
	ChildrenFieldName   string

	ComponentStructName   string
	FuncComponentTypeName string

	OptionSetters []optionSetter
	GlobalVars    []globalVars
}

type optionSetter struct {
	ParameterName string
	ParameterType string
	FieldName     string
	FuncName      string
}

type globalVars struct {
	Name  string
	Value string
}

func writeRuntime(w io.Writer, r runtime) error {
	const sourceTemplate = `package {{.PackageName}}

import (
{{- range .Imports}}
	{{.}}
{{- end}}
)

type {{.ContextStructName}} struct {
	{{.ContextFieldName}}     context.Context
	{{.WriterFieldName}}      io.Writer
	{{.AttributesFieldName}}  map[string]string
	{{.ChildrenFieldName}}    []Component
}

type {{.ContextOptionStructName}} func( {{- .ContextVarName}} *{{.ContextStructName -}} )

type {{.ComponentStructName}} interface {
	Render({{.ContextStructName}}) error
}

type {{.FuncComponentTypeName}} func() error

func (fc {{.FuncComponentTypeName}}) Render(ctx {{.ContextStructName}}) error {
	return fc()
}

func New {{- .ContextStructName -}} (opts ...{{.ContextOptionStructName -}} ) {{.ContextStructName}} {
	{{.ContextVarName}} := {{.ContextStructName -}} {}
	for _, opt := range opts {
		opt(& {{- .ContextVarName -}} )
	}
	return {{.ContextVarName}}
}

{{- with $parent := . }}
	{{- range .OptionSetters}}
func {{.FuncName -}} ( {{- .ParameterName}} {{.ParameterType -}} ) {{$parent.ContextOptionStructName}} {
	return func( {{- $parent.ContextVarName}} * {{- $parent.ContextStructName -}} ) {
		{{$parent.ContextVarName -}} . {{- .FieldName}} = {{.ParameterName}}
	}
}
	{{- end}}
{{- end}}

func writeString(w io.Writer, str string) error {
	_, err := io.WriteString(w, str)
	return err
}

func writeNumber(w io.Writer, str any) error {
	num := fmt.Sprintf("%d", str)
	_, err := io.WriteString(w, num)
	return err
}

//go:format 1 2
func formatString(format string, args ...any) string {
	return fmt.Sprintf(format, args...)
}
`
	t := template.New("runtime")
	t, err := t.Parse(sourceTemplate)
	if err != nil {
		return err
	}

	if err := t.Execute(w, r); err != nil {
		return err
	}
	return nil
}
