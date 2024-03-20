package raft

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/sleepymole/go-toydb/raft/raftpb"
	"github.com/sleepymole/go-toydb/storage"
	"github.com/sleepymole/go-toydb/util/itertools"
	"github.com/sleepymole/go-toydb/util/rangeutil"
)

const (
	entryKeyTag    = byte(0x00)
	termVoteKeyTag = byte(0x01)
	commitKeyTag   = byte(0x02)
)

var (
	entryKeyPrefix = []byte{entryKeyTag}
	termVoteKey    = []byte{termVoteKeyTag}
	commitKey      = []byte{commitKeyTag}
)

func encodeEntryKey(index Index) []byte {
	buf := make([]byte, 9)
	buf[0] = entryKeyTag
	binary.BigEndian.PutUint64(buf[1:], uint64(index))
	return buf
}

func decodeEntryKey(b []byte) (Index, error) {
	if len(b) != 9 || b[0] != entryKeyTag {
		return 0, errors.New("invalid entry key")
	}
	return Index(binary.BigEndian.Uint64(b[1:])), nil
}

func encodeEntryValue(term Term, command []byte) []byte {
	buf := make([]byte, 8+len(command))
	binary.BigEndian.PutUint64(buf, uint64(term))
	copy(buf[8:], command)
	return buf
}

func decodeEntryValue(b []byte) (Term, []byte, error) {
	if len(b) < 8 {
		return 0, nil, errors.New("invalid entry value, too short")
	}
	term := Term(binary.BigEndian.Uint64(b))
	command := make([]byte, len(b)-8)
	copy(command, b[8:])
	return term, command, nil
}

func decodeEntry(key, value []byte) (*raftpb.Entry, error) {
	index, err := decodeEntryKey(key)
	if err != nil {
		return nil, err
	}
	term, command, err := decodeEntryValue(value)
	if err != nil {
		return nil, err
	}
	return &raftpb.Entry{Index: index, Term: term, Command: command}, nil
}

func encodeTermVoteValue(term Term, votedFor NodeID) []byte {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf, uint64(term))
	binary.BigEndian.PutUint64(buf[8:], uint64(votedFor))
	return buf
}

func decodeTermVoteValue(b []byte) (Term, NodeID, error) {
	if len(b) != 16 {
		return 0, 0, errors.New("invalid term vote value")
	}
	term := Term(binary.BigEndian.Uint64(b))
	votedFor := NodeID(binary.BigEndian.Uint64(b[8:]))
	return term, votedFor, nil
}

func encodeCommitValue(index Index, term Term) []byte {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf, uint64(index))
	binary.BigEndian.PutUint64(buf[8:], uint64(term))
	return buf
}

func decodeCommitValue(b []byte) (Index, Term, error) {
	if len(b) != 16 {
		return 0, 0, errors.New("invalid commit value")
	}
	index := Index(binary.BigEndian.Uint64(b))
	term := Term(binary.BigEndian.Uint64(b[8:]))
	return index, term, nil
}

// Log is a persistent log of entries. It is the source of truth for
// the state machine. Each entry is identified by its index and term.
type Log struct {
	// engine is the underlying storage engine for the log.
	engine storage.Engine
	// lastIndex is the index of the last stored entry.
	lastIndex Index
	// lastTerm is the term of the last stored entry.
	lastTerm Term
	// commitIndex is the index of the last committed entry.
	commitIndex Index
	// commitTerm is the term of the last committed entry.
	commitTerm Term
	// sync indicates whether to flush the log to disk after each write.
	sync bool
}

func NewLog(engine storage.Engine, sync bool) (*Log, error) {
	lastIndex, lastTerm, err := loadLastIndex(engine)
	if err != nil {
		return nil, err
	}
	commitIndex, commitTerm, err := loadCommitIndex(engine)
	if err != nil {
		return nil, err
	}

	log := &Log{
		engine:      engine,
		lastIndex:   lastIndex,
		lastTerm:    lastTerm,
		commitIndex: commitIndex,
		commitTerm:  commitTerm,
		sync:        sync,
	}
	return log, nil
}

func loadLastIndex(engine storage.Engine) (Index, Term, error) {
	it, err := engine.ReverseScanPrefix(entryKeyPrefix)
	if err != nil {
		return 0, 0, err
	}
	defer it.Close()

	if !it.Next() {
		return 0, 0, it.Error()
	}
	entry, err := decodeEntry(it.Item().Key, it.Item().Value)
	if err != nil {
		return 0, 0, err
	}
	return entry.Index, entry.Term, nil
}

func loadCommitIndex(engine storage.Engine) (Index, Term, error) {
	value, err := engine.Get(commitKey)
	if err != nil {
		if err == storage.ErrNotFound {
			err = nil
		}
		return 0, 0, err
	}
	return decodeCommitValue(value)
}

func (l *Log) Status() (*storage.EngineStatus, error) {
	return l.engine.Status()
}

func (l *Log) CommitIndex() (Index, Term) {
	return l.commitIndex, l.commitTerm
}

func (l *Log) LastIndex() (Index, Term) {
	return l.lastIndex, l.lastTerm
}

func (l *Log) GetTerm() (_ Term, votedFor NodeID, _ error) {
	value, err := l.engine.Get(termVoteKey)
	if err != nil {
		return 0, votedFor, err
	}
	return decodeTermVoteValue(value)
}

func (l *Log) SetTerm(term Term, votedFor NodeID) error {
	value := encodeTermVoteValue(term, votedFor)
	if err := l.engine.Set(termVoteKey, value); err != nil {
		return err
	}
	return l.maybeFlush()
}

func (l *Log) Append(term Term, command []byte) (Index, error) {
	index := l.lastIndex + 1
	key := encodeEntryKey(index)
	value := encodeEntryValue(term, command)
	if err := l.engine.Set(key, value); err != nil {
		return 0, err
	}
	if err := l.maybeFlush(); err != nil {
		return 0, err
	}
	l.lastIndex = index
	l.lastTerm = term
	return index, nil
}

func (l *Log) Commit(index Index) error {
	if index < l.commitIndex {
		return fmt.Errorf("commit index regression: %d -> %d", l.commitIndex, index)
	}
	entry, err := l.Get(index)
	if err != nil {
		if err == storage.ErrNotFound {
			return fmt.Errorf("cannot commit non-existant index %d", index)
		}
		return err
	}
	value := encodeCommitValue(index, l.lastTerm)
	if err := l.engine.Set(commitKey, value); err != nil {
		return err
	}
	if err := l.maybeFlush(); err != nil {
		return err
	}
	l.commitIndex = entry.Index
	l.commitTerm = entry.Term
	return nil
}

func (l *Log) maybeFlush() error {
	if l.sync {
		return l.engine.Flush()
	}
	return nil
}

func (l *Log) Get(index Index) (*raftpb.Entry, error) {
	key := encodeEntryKey(index)
	value, err := l.engine.Get(key)
	if err != nil {
		return nil, err
	}
	return decodeEntry(key, value)
}

func (l *Log) Has(index Index, term Term) (bool, error) {
	entry, err := l.Get(index)
	if err != nil {
		if err == storage.ErrNotFound {
			return index == 0 && term == 0, nil
		}
		return false, err
	}
	return entry.Term == term, nil
}

func (l *Log) Scan(start, end Index) (itertools.Iterator[*raftpb.Entry], error) {
	from := encodeEntryKey(start)
	to := encodeEntryKey(end)
	it, err := l.engine.Scan(rangeutil.Range(from, to))
	if err != nil {
		return nil, err
	}
	return itertools.Map(it, func(kv storage.KeyValue) (*raftpb.Entry, error) {
		return decodeEntry(kv.Key, kv.Value)
	}), nil
}

// Splice splices a set log entries into the log and returns the index
// of the last entry.
//
// The entries must be contiguous, and the first entry's index must be
// at most lastIndex+1 and at least commitIndex+1. When splicing, the
// log will overwrite any existing entries and truncate any existing
// entries after the last spliced entry.
func (l *Log) Splice(entries []*raftpb.Entry) (Index, error) {
	if len(entries) == 0 {
		return l.lastIndex, nil
	}
	if entries[0].Index == 0 {
		return 0, errors.New("spliced entries must begin after index 0")
	}
	if entries[0].Index > l.lastIndex+1 {
		return 0, errors.New("spliced entries must not begin after last index + 1")
	}
	if entries[0].Index <= l.commitIndex {
		return 0, errors.New("spliced entries must not begin after commit index")
	}
	for i := 1; i < len(entries); i++ {
		if entries[i].Index != entries[i-1].Index+1 {
			return 0, errors.New("spliced entries must be contiguous")
		}
	}
	lastIndex := entries[len(entries)-1].Index
	lastTerm := entries[len(entries)-1].Term

	// We must delete the entries in reverse order, making sure the
	// remaining entries are always contiguous. Otherwise, we could
	// end up with a hole in the log if an error occurs.
	for i := l.lastIndex; i >= entries[0].Index; i-- {
		if err := l.engine.Delete(encodeEntryKey(i)); err != nil {
			return 0, err
		}
	}

	for _, entry := range entries {
		key := encodeEntryKey(entry.Index)
		value := encodeEntryValue(entry.Term, entry.Command)
		if err := l.engine.Set(key, value); err != nil {
			return 0, err
		}
	}
	l.lastIndex = lastIndex
	l.lastTerm = lastTerm
	return lastIndex, nil
}
