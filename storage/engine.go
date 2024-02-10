package storage

import (
	"errors"

	"github.com/sleepymole/go-toydb/util/itertools"
)

var ErrNotFound = errors.New("storage: not found")

type KeyValue struct {
	Key   []byte
	Value []byte
}

func MakeKeyValue(key, value []byte) KeyValue {
	return KeyValue{
		Key:   key,
		Value: value,
	}
}

type Engine interface {
	Get(key []byte) ([]byte, error)
	Set(key, value []byte) error
	Delete(key []byte) error
	Flush() error
	Scan(start, end []byte) (itertools.Iterator[KeyValue], error)
	ReverseScan(start, end []byte) (itertools.Iterator[KeyValue], error)
	Status() (*EngineStatus, error)
}

type EngineStatus struct {
	Name            string
	Keys            int64
	Size            int64
	TotalDiskSize   int64
	LiveDiskSize    int64
	GarbageDiskSize int64
}

func ScanPrefix(engine Engine, prefix []byte) (itertools.Iterator[KeyValue], error) {
	return engine.Scan(prefix, prefixNext(prefix))
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
