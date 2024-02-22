package storage

import (
	"bytes"
	"slices"

	"github.com/google/btree"
	"github.com/sleepymole/go-toydb/util/itertools"
	"github.com/sleepymole/go-toydb/util/rangeutil"
)

type Memory struct {
	engineBase

	data *btree.BTreeG[KeyValue]
}

var _ Engine = &Memory{}

func NewMemoryEngine() *Memory {
	m := &Memory{
		data: btree.NewG(2, func(a, b KeyValue) bool {
			return bytes.Compare(a.Key, b.Key) < 0
		}),
	}
	m.init(m)
	return m
}

func (m *Memory) Get(key []byte) ([]byte, error) {
	kv, ok := m.data.Get(MakeKeyValue(key, nil))
	if !ok {
		return nil, ErrNotFound
	}
	value := bytes.Clone(kv.Value)
	return value, nil
}

func (m *Memory) Set(key, value []byte) error {
	kv := MakeKeyValue(slices.Clone(key), slices.Clone(value))
	m.data.ReplaceOrInsert(kv)
	return nil
}

func (m *Memory) Delete(key []byte) error {
	m.data.Delete(MakeKeyValue(key, nil))
	return nil
}

func (m *Memory) Flush() error {
	return nil
}

func (m *Memory) Scan(start, end rangeutil.Bound) (itertools.Iterator[KeyValue], error) {
	start = start.Clone()
	var scan itertools.Iterator[KeyValue] = newMemoryScan(m.data, start.Value(), start.Unbounded(), true /* ascending */)
	if start.Excluded() {
		scan = itertools.SkipWhile(scan, func(kv KeyValue) bool {
			return !bytes.Equal(kv.Key, start.Value())
		})
	}
	if end.Included() {
		scan = itertools.TakeWhile(scan, func(kv KeyValue) bool {
			return bytes.Compare(kv.Key, end.Value()) <= 0
		})
	} else if end.Excluded() {
		scan = itertools.TakeWhile(scan, func(kv KeyValue) bool {
			return bytes.Compare(kv.Key, end.Value()) < 0
		})
	}
	return scan, nil
}

func (m *Memory) ReverseScan(start, end rangeutil.Bound) (itertools.Iterator[KeyValue], error) {
	start = start.Clone()
	var scan itertools.Iterator[KeyValue] = newMemoryScan(m.data, start.Value(), start.Unbounded(), false /* ascending */)
	if start.Excluded() {
		scan = itertools.SkipWhile(scan, func(kv KeyValue) bool {
			return !bytes.Equal(kv.Key, start.Value())
		})
	}
	if end.Included() {
		scan = itertools.TakeWhile(scan, func(kv KeyValue) bool {
			return bytes.Compare(kv.Key, end.Value()) >= 0
		})
	} else if end.Excluded() {
		scan = itertools.TakeWhile(scan, func(kv KeyValue) bool {
			return bytes.Compare(kv.Key, end.Value()) > 0
		})
	}
	return scan, nil
}

func (m *Memory) Status() (*EngineStatus, error) {
	size := int64(0)
	m.data.Ascend(func(kv KeyValue) bool {
		size += int64(len(kv.Key)) + int64(len(kv.Value))
		return true
	})
	return &EngineStatus{
		Name:            "memory",
		Keys:            int64(m.data.Len()),
		Size:            size,
		TotalDiskSize:   0,
		LiveDiskSize:    0,
		GarbageDiskSize: 0,
	}, nil
}

func (m *Memory) Close() error {
	return nil
}

type memoryScan struct {
	data      *btree.BTreeG[KeyValue]
	start     []byte
	unbounded bool
	ascending bool
	kv        KeyValue
	kvCh      chan KeyValue
	stop      chan struct{}
	done      chan struct{}
}

var _ itertools.Iterator[KeyValue] = &memoryScan{}

func newMemoryScan(
	data *btree.BTreeG[KeyValue],
	start []byte,
	unbounded bool,
	ascending bool,
) *memoryScan {
	return &memoryScan{
		data:      data,
		start:     start,
		unbounded: unbounded,
		ascending: ascending,
		stop:      make(chan struct{}, 1),
		done:      make(chan struct{}),
	}
}

func (m *memoryScan) Next() bool {
	if m.kvCh == nil {
		m.kvCh = make(chan KeyValue, 1)
		go m.run()
	}
	kv, ok := <-m.kvCh
	if !ok {
		m.kv.Key = nil
		m.kv.Value = nil
		return false
	}
	m.kv = kv
	return true
}

func (m *memoryScan) run() {
	defer close(m.done)
	defer close(m.kvCh)

	iterFunc := func(kv KeyValue) bool {
		select {
		case m.kvCh <- kv:
			return true
		case <-m.stop:
			return false
		}
	}

	if m.unbounded {
		if m.ascending {
			m.data.Ascend(iterFunc)
		} else {
			m.data.Descend(iterFunc)
		}
	} else {
		start := MakeKeyValue(m.start, nil)
		if m.ascending {
			m.data.AscendGreaterOrEqual(start, iterFunc)
		} else {
			m.data.DescendLessOrEqual(start, iterFunc)
		}

	}
}

func (m *memoryScan) Item() KeyValue {
	return m.kv
}

func (m *memoryScan) Error() error {
	return nil
}

func (m *memoryScan) Close() error {
	select {
	case m.stop <- struct{}{}:
	default:
	}
	<-m.done
	return nil
}
