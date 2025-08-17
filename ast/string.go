package ast

func (k TemplateKind) String() string {
	switch k {
	case DocumentTemplate:
		return "document"
	case ComponentTemplate:
		return "component"
	}
	panic("Unreachable")
}
