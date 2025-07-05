package slice

import (
	"iter"
)

func New[T any](data []T) Slice[T] {
	return Slice[T]{items: data}
}

func Empty[T any]() Slice[T] {
	var slice Slice[T]
	return slice
}

func Sized[T any](size int) Slice[T] {
	slice := Slice[T]{}
	slice.items = make([]T, 0, size)
	return slice
}

type Slice[T any] struct {
	items  []T
	frozen bool
}

func (s *Slice[T]) Size() int {
	return len(s.items)
}

func (s *Slice[T]) Freeze() {
	s.frozen = true
}

func (s *Slice[T]) Add(val T) bool {
	if s.frozen {
		return false
	}
	s.items = append(s.items, val)
	return true
}

func (s *Slice[T]) Item(index int) (T, bool) {
	if l := len(s.items); index >= l || index < 0 {
		var zero T
		return zero, false
	}
	val := s.items[index]
	return val, true
}

func (s *Slice[T]) ItemsCopy() []T {
	dst := make([]T, len(s.items))
	copy(dst, s.items)
	return dst
}

func (s *Slice[T]) Each() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, t := range s.items {
			if !yield(i, t) {
				break
			}
		}
	}
}
