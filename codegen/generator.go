package transpiler

import "io"

type Generator interface {
	Render(io.StringWriter) error
}
