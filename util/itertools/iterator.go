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
	return &sliceIterator[T]{slice: slice}
}

type sliceIterator[T any] struct {
	slice  []T
	cursor int
}

func (it *sliceIterator[T]) Next() bool {
	it.cursor++
	return it.cursor <= len(it.slice)
}

func (it *sliceIterator[T]) Item() T {
	return it.slice[it.cursor-1]
}

func (it *sliceIterator[T]) Error() error {
	return nil
}

func (it *sliceIterator[T]) Close() error {
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
	return &mapIterator[T, U]{it: it, f: f}
}

type mapIterator[T, U any] struct {
	it   Iterator[T]
	f    func(T) (U, error)
	item U
	err  error
}

func (m *mapIterator[T, U]) Next() bool {
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

func (m *mapIterator[T, U]) Item() U {
	return m.item
}

func (m *mapIterator[T, U]) Error() error {
	if m.err != nil {
		return m.err
	}
	return m.it.Error()
}

func (m *mapIterator[T, U]) Close() error {
	return m.it.Close()
}
