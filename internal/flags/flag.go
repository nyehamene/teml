package flags

type Flag uint

const (
	PreserveNewline Flag = 1 << iota
	PreserveComment
	ExitOnError
	HideErrors
	ReduceAlloc
	FlagNoNativeElement
	FlagNoBuiltinType
)
