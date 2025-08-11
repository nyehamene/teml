package ast

type Flag uint

const (
	FlagNoNativeElement Flag = 1 << iota
	FlagNoBuiltinType
)

type Pos struct {
	Start int
	End   int
	Line  int
	Col   int
}

type File struct {
	Name         string
	Package      Package
	Imports      []Import
	Usings       []Using
	Declarations []Declaration
	errs         []error
	Comments     []Comment
}

func (f *File) HasError() bool {
	if f == nil {
		return false
	}
	return len(f.errs) > 0
}

func (f *File) Errors() func(func(int, error) bool) {
	return func(yield func(int, error) bool) {
		for i, err := range f.errs {
			if !yield(i, err) {
				break
			}
		}
	}
}
