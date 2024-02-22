package itertools

import (
	"errors"
)

// Iterator is an interface for iterating over a sequence of items.
//
// Its cursor starts before the first item, so the first call to
// Next advances it to the first item.
type Iterator[T any] interface {
	// Next advances the iterator to the next item, which will then
	// be available through the Item method. It returns false when
	// the iterator stops producing items, either due to an error
	// or because the sequence has ended.
	Next() bool
	// Item returns the current item. It is only valid to call this
	// method after a call to Next has returned true.
	//
	// Implementations should make sure that the returned item is
	// immutable and remains valid after the iterator advances or
	// is closed.
	Item() T
	// Error returns the error, if any, that was encountered during
	// iteration. It should be checked after Next returns false.
	Error() error
	// Close releases any resources associated with the iterator.
	Close() error
}

// ErrStop is a sentinel error value that can be returned by a
// function to stop an iteration.
var ErrStop = errors.New("stop")

// Walk calls the given function for each item in the iterator.
func Walk[T any](it Iterator[T], f func(T) error) error {
	defer it.Close()
	for it.Next() {
		if err := f(it.Item()); err != nil {
			if err == ErrStop {
				return nil
			}
			return err
		}
	}
	return it.Error()
}

// Slice returns an iterator over the elements of a slice.
func Slice[T any](slice []T) Iterator[T] {
	return &sliceIter[T]{slice: slice}
}

type sliceIter[T any] struct {
	slice  []T
	cursor int
}

func (it *sliceIter[T]) Next() bool {
	it.cursor++
	return it.cursor <= len(it.slice)
}

func (it *sliceIter[T]) Item() T {
	return it.slice[it.cursor-1]
}

func (it *sliceIter[T]) Error() error {
	return nil
}

func (it *sliceIter[T]) Close() error {
	return nil
}

// ToSlice collects the items from an iterator into a slice.
func ToSlice[T any](it Iterator[T]) ([]T, error) {
	defer it.Close()
	var slice []T
	for it.Next() {
		slice = append(slice, it.Item())
	}
	if err := it.Error(); err != nil {
		return nil, err
	}
	return slice, nil
}

// Map returns an iterator that applies a function to each item in
// the input iterator.
func Map[T, U any](it Iterator[T], f func(T) (U, error)) Iterator[U] {
	return &mapIter[T, U]{it: it, f: f}
}

type mapIter[T, U any] struct {
	it   Iterator[T]
	f    func(T) (U, error)
	item U
	err  error
}

func (m *mapIter[T, U]) Next() bool {
	if m.err != nil {
		return false
	}
	if m.it.Next() {
		item, err := m.f(m.it.Item())
		if err != nil {
			var zero U
			m.item = zero
			m.err = err
			return false
		}
		m.item = item
		return true
	}
	return false
}

func (m *mapIter[T, U]) Item() U {
	return m.item
}

func (m *mapIter[T, U]) Error() error {
	if m.err != nil {
		return m.err
	}
	return m.it.Error()
}

func (m *mapIter[T, U]) Close() error {
	return m.it.Close()
}

// FilterMap returns an iterator that applies a function to each item in
// the input iterator and yields the result if the function returns true.
func FilterMap[T, U any](it Iterator[T], f func(T) (U, bool, error)) Iterator[U] {
	return &filterMap[T, U]{it: it, f: f}
}

type filterMap[T, U any] struct {
	it   Iterator[T]
	f    func(T) (U, bool, error)
	item U
	err  error
}

func (f *filterMap[T, U]) Next() bool {
	if f.err != nil {
		return false
	}
	for f.it.Next() {
		item, ok, err := f.f(f.it.Item())
		if err != nil {
			f.err = err
			return false
		}
		if ok {
			f.item = item
			return true
		}
	}
	return false
}

func (f *filterMap[T, U]) Item() U {
	return f.item
}

func (f *filterMap[T, U]) Error() error {
	if f.err != nil {
		return f.err
	}
	return f.it.Error()
}

func (f *filterMap[T, U]) Close() error {
	return f.it.Close()
}

// SkipWhile returns an iterator that skips items from the input
// iterator while a predicate function returns true.
func SkipWhile[T any](it Iterator[T], f func(T) bool) Iterator[T] {
	return &skipWhile[T]{it: it, f: f}
}

type skipWhile[T any] struct {
	it      Iterator[T]
	f       func(T) bool
	skipped bool
}

func (s *skipWhile[T]) Next() bool {
	if s.skipped {
		return s.it.Next()
	}
	for s.it.Next() {
		if !s.f(s.it.Item()) {
			s.skipped = true
			return true
		}
	}
	return false
}

func (s *skipWhile[T]) Item() T {
	return s.it.Item()
}

func (s *skipWhile[T]) Error() error {
	return s.it.Error()
}

func (s *skipWhile[T]) Close() error {
	return s.it.Close()
}

// TakeWhile returns an iterator that yields items from the input
// iterator while a predicate function returns true.
func TakeWhile[T any](it Iterator[T], f func(T) bool) Iterator[T] {
	return &takeWhile[T]{it: it, f: f}
}

type takeWhile[T any] struct {
	it  Iterator[T]
	f   func(T) bool
	eof bool
}

func (t *takeWhile[T]) Next() bool {
	if t.eof {
		return false
	}
	if t.it.Next() && t.f(t.it.Item()) {
		return true
	}
	t.eof = true
	return false
}

func (t *takeWhile[T]) Item() T {
	return t.it.Item()
}

func (t *takeWhile[T]) Error() error {
	return t.it.Error()
}

func (t *takeWhile[T]) Close() error {
	return t.it.Close()
}
