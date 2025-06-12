package token

func assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

func not(cond bool) bool {
	return !cond
}
