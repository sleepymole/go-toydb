package raft

import (
	"github.com/sleepymole/go-toydb/storage"
)

type Index uint64

type Entry struct {
	Index   Index
	Term    Term
	Command []byte
}

type Log struct {
	engine      storage.Engine
	lastIndex   Index
	lastTerm    Term
	commitIndex Index
	commitTerm  Term
	sync        bool
}

func (l *Log) Status() (*storage.EngineStatus, error) {
	return l.engine.Status()
}
