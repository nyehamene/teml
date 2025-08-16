package tel

//go:generate stringer -type=Flag
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
