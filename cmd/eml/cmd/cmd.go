package cmd

func NewError(err error) Result {
	return Error{Result: err}
}

func NewFile(name string) File {
	return File{
		File: name,
	}
}

type Result interface {
	cmd()
}

func (File) cmd()  {}
func (Error) cmd() {}

type File struct {
	File string
}

type Error struct {
	Result error
}
