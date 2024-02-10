package set

import (
	"maps"
	"slices"
)

// Comparable is a type that can be compared for equality and hashed.
type Comparable[T any] interface {
	Equal(T) bool
	Hash() uint64
}

// HashSet is a set of T that satisfies the Comparable interface.
type HashSet[T Comparable[T]] struct {
	m map[uint64][]T // hash -> elements
	n int
}

// NewHashSet returns a new HashSet.
func NewHashSet[T Comparable[T]]() *HashSet[T] {
	return &HashSet[T]{m: make(map[uint64][]T)}
}

// NewHashSetOf returns a new HashSet constructed from the elements in slice.
func NewHashSetOf[T Comparable[T]](slice []T) *HashSet[T] {
	s := NewHashSet[T]()
	s.AddSlice(slice)
	return s
}

// Clone returns a new HashSet cloned from the elements in s.
func (s *HashSet[T]) Clone() *HashSet[T] {
	return &HashSet[T]{m: maps.Clone(s.m), n: s.n}
}

// Add adds e to s.
func (s *HashSet[T]) Add(e T) {
	h := e.Hash()
	if !slices.ContainsFunc(s.m[h], e.Equal) {
		s.m[h] = append(s.m[h], e)
		s.n++
	}
}

// AddSlice adds each element of es to s.
func (s *HashSet[T]) AddSlice(es []T) {
	for _, e := range es {
		s.Add(e)
	}
}

// AddSet adds each element of es to s.
func (s *HashSet[T]) AddSet(other *HashSet[T]) {
	for _, es := range other.m {
		for _, e := range es {
			s.Add(e)
		}
	}
}

// Slice returns the elements of the set as a slice. The elements will not be
// in any particular order.
func (s HashSet[T]) Slice() []T {
	slice := make([]T, 0, s.Len())
	for _, es := range s.m {
		slice = append(slice, es...)
	}
	return slice
}

// Delete removes e from the set.
func (s *HashSet[T]) Delete(e T) {
	h := e.Hash()
	s.m[h] = slices.DeleteFunc(s.m[h], func(x T) bool {
		if e.Equal(x) {
			s.n--
			return true
		}
		return false
	})
}

// Contains reports whether s contains e.
func (s *HashSet[T]) Contains(e T) bool {
	h := e.Hash()
	return slices.ContainsFunc(s.m[h], func(x T) bool {
		return e.Equal(x)
	})
}

// Len reports the number of items in s.
func (s *HashSet[T]) Len() int {
	return s.n
}

// Equal reports whether s is equal to other.
func (s *HashSet[T]) Equal(other *HashSet[T]) bool {
	if s.Len() != other.Len() {
		return false
	}
	for h, es := range s.m {
		otherEs := other.m[h]
		eq := func(a, b T) bool {
			return a.Equal(b)
		}
		if !slices.EqualFunc(es, otherEs, eq) {
			return false
		}
	}
	return true
}

// Range calls f for each element in s.
// If f returns false, Range stops the iteration.
func (s *HashSet[T]) Range(f func(T) bool) {
	for _, es := range s.m {
		for _, e := range es {
			if !f(e) {
				return
			}
		}
	}
}
