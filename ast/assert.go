package ast

import (
	"fmt"
	"reflect"
)

func assert(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}

func assertValue(cond, value bool, msg string) {
	if !cond && value {
		panic(msg)
	}
}

func mustCastType[T Entity](entity Entity) T {
	var t T
	var ok bool
	t, ok = entity.(T)
	if !ok {
		panic(fmt.Sprintf("expected type %v got %v", reflect.TypeOf(t), reflect.TypeOf(entity)))
	}
	return t
}
