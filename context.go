package tel

type Context struct {
	Flag Flag
}

func NewContext() Context {
	return Context{}
}

func NewContextWithFlags(f Flag, flags ...Flag) Context {
	for _, flag := range flags {
		f |= flag
	}
	return Context{Flag: f}
}
