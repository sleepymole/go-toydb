package storage

import (
	"bytes"
	"slices"

	"github.com/google/btree"
	"github.com/sleepymole/go-toydb/util/itertools"
)

type MemoryEngine struct {
	data *btree.BTreeG[KeyValue]
}

var _ Engine = (*MemoryEngine)(nil)

func NewMemoryEngine() *MemoryEngine {
	return &MemoryEngine{
		data: btree.NewG(2, func(a, b KeyValue) bool {
			return bytes.Compare(a.Key, b.Key) < 0
		}),
	}
}

func (m *MemoryEngine) Get(key []byte) ([]byte, error) {
	kv, ok := m.data.Get(MakeKeyValue(key, nil))
	if !ok {
		return nil, ErrNotFound
	}
	value := bytes.Clone(kv.Value)
	return value, nil
}

func (m *MemoryEngine) Set(key, value []byte) error {
	kv := MakeKeyValue(slices.Clone(key), slices.Clone(value))
	m.data.ReplaceOrInsert(kv)
	return nil
}

func (m *MemoryEngine) Delete(key []byte) error {
	m.data.Delete(MakeKeyValue(key, nil))
	return nil
}

func (m *MemoryEngine) Flush() error {
	return nil
}

func (m *MemoryEngine) Scan(start, end []byte) (itertools.Iterator[KeyValue], error) {
	scan := newMemoryEngineScan(m.data, slices.Clone(start), slices.Clone(end), false)
	return scan, nil
}

func (m *MemoryEngine) ReverseScan(start, end []byte) (itertools.Iterator[KeyValue], error) {
	scan := newMemoryEngineScan(m.data, slices.Clone(start), slices.Clone(end), true)
	return scan, nil
}

func (m *MemoryEngine) Status() (*EngineStatus, error) {
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

type memoryEngineScan struct {
	data       *btree.BTreeG[KeyValue]
	start, end []byte
	reverse    bool
	kv         KeyValue
	kvCh       chan KeyValue
	stop       chan struct{}
	done       chan struct{}
}

var _ itertools.Iterator[KeyValue] = &memoryEngineScan{}

func newMemoryEngineScan(
	data *btree.BTreeG[KeyValue],
	start, end []byte,
	reverse bool,
) *memoryEngineScan {
	return &memoryEngineScan{
		data:    data,
		start:   start,
		end:     end,
		reverse: reverse,
		stop:    make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
}

func (m *memoryEngineScan) Next() bool {
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

func (m *memoryEngineScan) run() {
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

	from := MakeKeyValue(m.start, nil)
	if len(m.end) > 0 {
		to := MakeKeyValue(m.end, nil)
		if m.reverse {
			m.data.DescendRange(from, to, iterFunc)
		} else {
			m.data.AscendRange(from, to, iterFunc)
		}
	} else {
		if m.reverse {
			m.data.DescendLessOrEqual(from, iterFunc)
		} else {
			m.data.AscendGreaterOrEqual(from, iterFunc)
		}
	}
}

func (m *memoryEngineScan) Item() KeyValue {
	return m.kv
}

func (m *memoryEngineScan) Error() error {
	return nil
}

func (m *memoryEngineScan) Close() error {
	select {
	case m.stop <- struct{}{}:
	default:
	}
	<-m.done
	return nil
}
