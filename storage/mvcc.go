package storage

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/samber/mo"
	"github.com/sleepymole/go-toydb/util/itertools"
	"github.com/sleepymole/go-toydb/util/rangeutil"
	"github.com/sleepymole/go-toydb/util/set"
)

var (
	ErrReadOnly     = errors.New("storage: read-only transcation")
	ErrSerializable = errors.New("storage: serialization failure, retry transcation")
)

type MVCCVersion uint64

type MVCC struct {
	mu struct {
		sync.RWMutex
		engine Engine
	}
}

func NewMVCC(engine Engine) *MVCC {
	m := &MVCC{}
	m.mu.engine = engine
	return m
}

func (m *MVCC) Begin() (*MVCCTxn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	version, err := m.nextVersionLocked()
	if err != nil {
		return nil, err
	}
	if err := m.setNextVersionLocked(version + 1); err != nil {
		return nil, err
	}

	active, err := m.scanActiveLocked()
	if err != nil {
		return nil, err
	}
	if active.Len() > 0 {
		if err := m.setTxnActiveSnapshotLocked(version, active); err != nil {
			return nil, err
		}
	}
	if err := m.setTxnActiveLocked(version); err != nil {
		return nil, err
	}

	return &MVCCTxn{
		mvcc: m,
		state: &MVCCTxnState{
			Version:  version,
			ReadOnly: false,
			Active:   active,
		},
	}, nil
}

func (m *MVCC) BeginReadOnly() (*MVCCTxn, error) {
	return m.beginReadOnlyInternal(mo.None[MVCCVersion]())
}

func (m *MVCC) BeginAsOf(version MVCCVersion) (*MVCCTxn, error) {
	return m.beginReadOnlyInternal(mo.Some(version))
}

// beginReadOnlyInternal begins a new read-only MVCCTxn. If version is
// given it will see the state as of the beginning of that version (ignoring
// writes at that version). In other words, it sees the same state as the
// read-write MVCCTxn at the version saw when it began.
func (m *MVCC) beginReadOnlyInternal(asOf mo.Option[MVCCVersion]) (*MVCCTxn, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	version, err := m.nextVersionLocked()
	if err != nil {
		return nil, err
	}

	// If requested, create the transcation as of a past version, restoring
	// the active snapshot as of the beginning of that version. Otherwise,
	// use the latest version and get the current, real-time snapshot.
	var active set.Set[MVCCVersion]
	if asOfVersion, ok := asOf.Get(); ok {
		if asOfVersion >= version {
			return nil, fmt.Errorf("version %d does not exists", asOfVersion)
		}
		version = asOfVersion
		active, err = m.getTxnActiveSnapshotLocked(version)
		if err != nil {
			return nil, err
		}
	} else {
		active, err = m.scanActiveLocked()
		if err != nil {
			return nil, err
		}
	}

	return &MVCCTxn{
		mvcc: m,
		state: &MVCCTxnState{
			Version:  version,
			ReadOnly: true,
			Active:   active,
		},
	}, nil
}

func (m *MVCC) Resume(state MVCCTxnState) (*MVCCTxn, error) {
	if !state.ReadOnly {
		m.mu.RLock()
		defer m.mu.RUnlock()

		hasTxnActive, err := m.hasTxnActiveLocked(state.Version)
		if err != nil {
			return nil, err
		}
		if !hasTxnActive {
			return nil, fmt.Errorf("no active MVCCTxn at version %d", state.Version)
		}
	}
	return &MVCCTxn{
		mvcc:  m,
		state: &state,
	}, nil
}

func (m *MVCC) GetUnversioned(key []byte) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	encodedKey := encodeMVCCUnversionedKey(key)
	return m.mu.engine.Get(encodedKey)
}

func (m *MVCC) SetUnversioned(key, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	encodedKey := encodeMVCCUnversionedKey(key)
	return m.mu.engine.Set(encodedKey, value)
}

func (m *MVCC) Status() (*MVCCStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	version, err := m.nextVersionLocked()
	if err != nil {
		return nil, err
	}
	active, err := m.scanActiveLocked()
	if err != nil {
		return nil, err
	}
	engineStatus, err := m.mu.engine.Status()
	if err != nil {
		return nil, err
	}
	return &MVCCStatus{
		Versions:   int64(version - 1),
		ActiveTxns: int64(active.Len()),
		Storage:    engineStatus,
	}, nil
}

func (m *MVCC) nextVersionLocked() (MVCCVersion, error) {
	value, err := m.mu.engine.Get(mvccNextVersionKey)
	if err != nil {
		if err == ErrNotFound {
			return 1, nil
		}
		return 0, err
	}
	return decodeMVCCNextVersionValue(value)
}

func (m *MVCC) setNextVersionLocked(version MVCCVersion) error {
	value := encodeMVCCNextVersionValue(version)
	return m.mu.engine.Set(mvccNextVersionKey, value)
}

func (m *MVCC) scanActiveLocked() (set.Set[MVCCVersion], error) {
	iter, err := m.mu.engine.ScanPrefix(mvccTxnActiveKeyPrefix)
	if err != nil {
		return nil, err
	}
	active := make(set.Set[MVCCVersion])
	walkFunc := func(kv KeyValue) error {
		version, err := decodeMVCCTxnActiveKey(kv.Key)
		if err != nil {
			return err
		}
		active.Add(version)
		return nil
	}
	if err := itertools.Walk(iter, walkFunc); err != nil {
		return nil, err
	}
	return active, nil
}

func (m *MVCC) setTxnActiveLocked(version MVCCVersion) error {
	key := encodeMVCCTxnActiveKey(version)
	return m.mu.engine.Set(key, mvccTxnActiveValue)
}

func (m *MVCC) hasTxnActiveLocked(version MVCCVersion) (bool, error) {
	key := encodeMVCCTxnActiveKey(version)
	_, err := m.mu.engine.Get(key)
	if err != nil {
		if err == ErrNotFound {
			err = nil
		}
		return false, err
	}
	return true, nil
}

func (m *MVCC) setTxnActiveSnapshotLocked(version MVCCVersion, active set.Set[MVCCVersion]) error {
	key := encodeMVCCTxnActiveSnapshotKey(version)
	value := encodeMVCCTxnActiveSnapshotValue(active)
	return m.mu.engine.Set(key, value)
}

func (m *MVCC) getTxnActiveSnapshotLocked(version MVCCVersion) (set.Set[MVCCVersion], error) {
	key := encodeMVCCTxnActiveSnapshotKey(version)
	value, err := m.mu.engine.Get(key)
	if err != nil {
		if err == ErrNotFound {
			err = nil
		}
		return nil, err
	}
	return decodeMVCCTxnActiveSnapshotValue(value)
}

// writeVersion writes a new version for a key at the given version. None writes
// a deletion tombstone. If a write conflict is found (either a newer or an
// uncommitted version), an serializable error is returned. Replacing our own
// uncommitted version is fine.
func (m *MVCC) writeVersion(key []byte, value mo.Option[[]byte], state *MVCCTxnState) error {
	minVersion := state.Version + 1
	for v := range state.Active {
		minVersion = min(minVersion, v)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check for write conflicts, i.e if the latest version of the key is
	// invisible to us (either a newer version, or an uncommitted version
	// in the past). We can only conflict with the latest key, since all
	// MVCCTxn enforce the same invariant.
	start := encodeMVCCVersionedKey(key, minVersion)
	end := encodeMVCCVersionedKey(key, math.MaxUint64)
	iter, err := m.mu.engine.ReverseScan(rangeutil.RangeInclusive(start, end))
	if err != nil {
		return err
	}
	walkFunc := func(kv KeyValue) error {
		_, version, err := decodeMVCCVersionedKey(kv.Key)
		if err != nil {
			return err
		}
		if !state.IsVisible(version) {
			return ErrSerializable
		}
		return nil
	}
	if err := itertools.Walk(iter, walkFunc); err != nil {
		return err
	}

	// Write the new version and its write record.
	if err := m.mu.engine.Set(
		encodeMVCCTxnWriteKey(state.Version, key),
		mvccTxnWriteValue,
	); err != nil {
		return err
	}
	return m.mu.engine.Set(
		encodeMVCCVersionedKey(key, state.Version),
		encodeMVCCVersionedValue(value),
	)
}

func (m *MVCC) get(key []byte, state *MVCCTxnState) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	start := encodeMVCCVersionedKey(key, 0)
	end := encodeMVCCVersionedKey(key, state.Version)
	iter, err := m.mu.engine.ReverseScan(rangeutil.Range(start, end))
	if err != nil {
		return nil, err
	}

	var value mo.Option[[]byte]
	walkFunc := func(kv KeyValue) error {
		_, version, err := decodeMVCCVersionedKey(iter.Item().Key)
		if err != nil {
			_ = iter.Close()
			return err
		}
		if state.IsVisible(version) {
			value, err = decodeMVCCVersionedValue(iter.Item().Value)
			if err != nil {
				return err
			}
			return itertools.ErrStop
		}
		return nil
	}
	if err := itertools.Walk(iter, walkFunc); err != nil {
		return nil, err
	}
	if value.IsAbsent() {
		return nil, ErrNotFound
	}
	return value.MustGet(), err
}

func (m *MVCC) commit(state *MVCCTxnState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	iter, err := m.mu.engine.ScanPrefix(encodeMVCCTxnWriteKeyPrefix(state.Version))
	if err != nil {
		return err
	}
	remove, err := itertools.ToSlice(
		itertools.Map(iter, func(kv KeyValue) ([]byte, error) {
			_, key, err := decodeMVCCTxnWriteKey(kv.Key)
			return key, err
		}),
	)

	for _, key := range remove {
		if err := m.mu.engine.Delete(key); err != nil {
			return err
		}
	}
	return m.mu.engine.Delete(encodeMVCCTxnActiveKey(state.Version))
}

func (m *MVCC) rollback(state *MVCCTxnState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	iter, err := m.mu.engine.ScanPrefix(encodeMVCCTxnWriteKeyPrefix(state.Version))
	if err != nil {
		return err
	}
	var rollback [][]byte
	walkFunc := func(kv KeyValue) error {
		_, key, err := decodeMVCCTxnWriteKey(kv.Key)
		if err != nil {
			return err
		}
		rollback = append(rollback, key)
		rollback = append(rollback, encodeMVCCVersionedKey(key, state.Version))
		return nil
	}
	if err := itertools.Walk(iter, walkFunc); err != nil {
		return err
	}

	for _, key := range rollback {
		if err := m.mu.engine.Delete(key); err != nil {
			return err
		}
	}
	return m.mu.engine.Delete(encodeMVCCTxnActiveKey(state.Version))
}

type MVCCStatus struct {
	Versions   int64
	ActiveTxns int64
	Storage    *EngineStatus
}

type MVCCTxn struct {
	mvcc  *MVCC
	state *MVCCTxnState
}

type MVCCTxnState struct {
	Version  MVCCVersion
	ReadOnly bool
	Active   set.Set[MVCCVersion]
}

func (m *MVCCTxnState) IsVisible(version MVCCVersion) bool {
	if m.Active.Contains(version) {
		return false
	}
	if m.ReadOnly {
		return version < m.Version
	}
	return version <= m.Version
}

func (t *MVCCTxn) Version() MVCCVersion {
	return t.state.Version
}

func (t *MVCCTxn) ReadOnly() bool {
	return t.state.ReadOnly
}

func (t *MVCCTxn) State() *MVCCTxnState {
	return t.state
}

func (t *MVCCTxn) Commit() error {
	if t.state.ReadOnly {
		return nil
	}
	return t.mvcc.commit(t.state)
}

func (t *MVCCTxn) Rollback() error {
	if t.state.ReadOnly {
		return nil
	}
	return t.mvcc.rollback(t.state)
}

func (t *MVCCTxn) Delete(key []byte) error {
	if t.state.ReadOnly {
		return ErrReadOnly
	}
	return t.mvcc.writeVersion(key, mo.None[[]byte](), t.state)
}

func (t *MVCCTxn) Set(key, value []byte) error {
	if t.state.ReadOnly {
		return ErrReadOnly
	}
	return t.mvcc.writeVersion(key, mo.Some(value), t.state)
}

func (t *MVCCTxn) Get(key []byte) ([]byte, error) {
	return t.mvcc.get(key, t.state)
}

func (t *MVCCTxn) Scan(start, end rangeutil.Bound) (itertools.Iterator[KeyValue], error) {
	return newMVCCScan(t.mvcc, start, end, false /* reverse */, t.state)
}

func (t *MVCCTxn) ReverseScan(start, end rangeutil.Bound) (itertools.Iterator[KeyValue], error) {
	return newMVCCScan(t.mvcc, start, end, true /* reverse */, t.state)
}

type mvccKeyValue struct {
	key     []byte
	value   mo.Option[[]byte]
	version MVCCVersion
}

func decodeMVCCKeyValue(encoded KeyValue) (mvccKeyValue, error) {
	key, version, err := decodeMVCCVersionedKey(encoded.Key)
	if err != nil {
		return mvccKeyValue{}, err
	}
	value, err := decodeMVCCVersionedValue(encoded.Value)
	if err != nil {
		return mvccKeyValue{}, err
	}
	return mvccKeyValue{
		key:     key,
		value:   value,
		version: version,
	}, nil
}

type mvccScan struct {
	mvcc    *MVCC
	reverse bool
	state   *MVCCTxnState
	kv      KeyValue

	inner      itertools.Iterator[mvccKeyValue]
	innerValid bool
}

func newMVCCScan(
	mvcc *MVCC,
	start, end rangeutil.Bound,
	reverse bool,
	state *MVCCTxnState,
) (*mvccScan, error) {
	startVersion, endVersion := MVCCVersion(0), MVCCVersion(math.MaxUint64)
	if reverse {
		startVersion, endVersion = math.MaxUint64, 0
	}

	if start.Included() || start.Excluded() {
		encoded := encodeMVCCVersionedKey(start.Value(), startVersion)
		start = rangeutil.Included(encoded)
	} else {
		encoded := encodeMVCCVersionedKey(start.Value(), startVersion)
		start = rangeutil.Excluded(encoded)
	}

	if end.Included() {
		encoded := encodeMVCCVersionedKey(end.Value(), endVersion)
		end = rangeutil.Excluded(encoded)
	} else if end.Excluded() {
		encoded := encodeMVCCVersionedKey(end.Value(), endVersion)
		end = rangeutil.Excluded(encoded)
	} else {
		next := prefixNext(mvccVersionedKeyPrefix)
		end = rangeutil.Excluded(next)
	}

	mvcc.mu.RLock()

	var rawIter itertools.Iterator[KeyValue]
	var err error
	if reverse {
		rawIter, err = mvcc.mu.engine.ReverseScan(start, end)
	} else {
		rawIter, err = mvcc.mu.engine.Scan(start, end)
	}
	if err != nil {
		mvcc.mu.RUnlock()
		return nil, err
	}

	inner := itertools.FilterMap(
		rawIter,
		func(encoded KeyValue) (mvccKeyValue, bool, error) {
			kv, err := decodeMVCCKeyValue(encoded)
			if err != nil {
				return mvccKeyValue{}, false, err
			}
			if state.IsVisible(kv.version) {
				return kv, true, nil
			}
			return mvccKeyValue{}, false, nil
		},
	)
	// Advance to the first visible key. It creates an invarint that the
	// inner iterator is always at a next visible key when called from Next.
	innerValid := inner.Next()

	return &mvccScan{
		mvcc:       mvcc,
		reverse:    reverse,
		state:      state,
		inner:      inner,
		innerValid: innerValid,
	}, nil
}

func (s *mvccScan) Next() bool {
	for s.innerValid {
		kv := s.inner.Item()
		// Scan over all the versions of the same key.
		s.innerValid = s.inner.Next()
		for s.innerValid && bytes.Equal(s.inner.Item().key, kv.key) {
			// Keep the latest version of the key.
			if !s.reverse {
				kv = s.inner.Item()
			}
			s.innerValid = s.inner.Next()
		}
		if kv.value.IsPresent() {
			s.kv = KeyValue{
				Key:   kv.key,
				Value: kv.value.MustGet(),
			}
			return true
		}
	}
	return false
}

func (s *mvccScan) Item() KeyValue {
	return s.kv
}

func (s *mvccScan) Error() error {
	return s.inner.Error()
}

func (s *mvccScan) Close() error {
	defer s.mvcc.mu.RUnlock()

	return s.inner.Close()
}
