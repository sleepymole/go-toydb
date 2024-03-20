package sql

import (
	"github.com/sleepymole/go-toydb/raft"
	"github.com/sleepymole/go-toydb/storage"
)

type RaftState struct {
	raft.State

	engine storage.Engine
}

func NewRaftState(engine storage.Engine) *RaftState {
	return &RaftState{engine: engine}
}

type RaftEngine struct {
	Engine

	raftServer *raft.Server
}

func NewRaftEngine(raftServer *raft.Server) *RaftEngine {
	return &RaftEngine{raftServer: raftServer}
}

type RaftTxn struct {
	Txn
}
