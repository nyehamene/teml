package slice

import (
	"iter"
)

func New[T any](data []T) Slice[T] {
	return Slice[T]{data}
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
	items []T
}

func (s *Slice[T]) Size() int {
	return len(s.items)
}

func (s *Slice[T]) Add(val T) *Slice[T] {
	s.items = append(s.items, val)
	return s
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
