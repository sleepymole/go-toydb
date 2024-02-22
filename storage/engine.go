package storage

import (
	"errors"

	"github.com/sleepymole/go-toydb/util/itertools"
	"github.com/sleepymole/go-toydb/util/rangeutil"
)

// ErrNotFound is returned when a key is not found in the engine.
var ErrNotFound = errors.New("storage: not found")

// KeyValue represents a key-value pair.
type KeyValue struct {
	Key   []byte
	Value []byte
}

// MakeKeyValue creates a KeyValue from the given key and value.
func MakeKeyValue(key, value []byte) KeyValue {
	return KeyValue{
		Key:   key,
		Value: value,
	}
}

// Engine is a key/value storage engine, where both keys and values are
// arbitrary byte string between 0 B and 2 GB, stored in lexicographical
// key order. Writes are only guaranteed durable after calling Flush().
//
// This interface is designed to be used by a single goroutine, and
// implementations are not required to be thread-safe.
type Engine interface {
	// Get returns the value for the given key, or any error that occurred.
	// If the key does not exist, it returns ErrNotFound.
	Get(key []byte) ([]byte, error)
	// Set sets the value for the given key.
	Set(key, value []byte) error
	// Delete deletes the value for the given key. Deleting a non-existing key
	// is a no-op.
	Delete(key []byte) error
	// Flush flushes all the pending writes to the underlying storage medium.
	Flush() error
	// Scan returns an iterator that yields all key-value pairs with a key
	// that satisfies the given range bounds, in ascending order.
	Scan(start, end rangeutil.Bound) (itertools.Iterator[KeyValue], error)
	// ReverseScan returns an iterator that yields all key-value pairs with a
	// key that satisfies the given range bounds, in descending order.
	ReverseScan(start, end rangeutil.Bound) (itertools.Iterator[KeyValue], error)
	// ScanPrefix returns an iterator that yields all key-value pairs with a
	// key that has the given prefix, in ascending order.
	ScanPrefix(prefix []byte) (itertools.Iterator[KeyValue], error)
	// ReverseScanPrefix returns an iterator that yields all key-value pairs
	// with a key that has the given prefix, in descending order.
	ReverseScanPrefix(prefix []byte) (itertools.Iterator[KeyValue], error)
	// Status returns the current status of the engine.
	Status() (*EngineStatus, error)
	// Close closes the engine.
	Close() error
}

// EngineStatus contains the current status of an engine.
type EngineStatus struct {
	// Name is the name of the engine.
	Name string
	// Keys is the number of live keys in the engine.
	Keys int64
	// Size is the logical size of live key/values pairs.
	Size int64
	// TotalDiskSize is the on-disk size of all data, live and garbage.
	TotalDiskSize int64
	// LiveDiskSize is the on-disk size of live data.
	LiveDiskSize int64
	// GarbageDiskSize is the on-disk size of garbage data.
	GarbageDiskSize int64
}

type engineBase struct {
	self Engine
}

func (b *engineBase) init(self Engine) {
	b.self = self
}

func (b *engineBase) ScanPrefix(prefix []byte) (itertools.Iterator[KeyValue], error) {
	next := prefixNext(prefix)
	if len(next) == 0 {
		return b.self.Scan(rangeutil.RangeFrom(prefix))
	}
	return b.self.Scan(rangeutil.Range(prefix, next))
}

func (b *engineBase) ReverseScanPrefix(prefix []byte) (itertools.Iterator[KeyValue], error) {
	next := prefixNext(prefix)
	if len(next) == 0 {
		return b.self.Scan(rangeutil.RangeFrom(prefix))
	}
	return b.self.Scan(rangeutil.Range(prefix, next))

}

func prefixNext(prefix []byte) []byte {
	next := make([]byte, 0, len(prefix))
	for i := len(prefix) - 1; i >= 0; i-- {
		if prefix[i] < 0xff {
			next = append(next, prefix[:i]...)
			next = append(next, prefix[i]+1)
			return next
		}
	}
	return nil
}
