package sql

import (
	"context"

	"github.com/sleepymole/go-toydb/raft"
	"github.com/sleepymole/go-toydb/raft/raftpb"
	"github.com/sleepymole/go-toydb/storage"
)

type RaftState struct {
	raft.State

	engine storage.Engine
}

func NewRaftState(engine storage.Engine) *RaftState {
	return &RaftState{engine: engine}
}

type RaftClient interface {
	Mutate(ctx context.Context, command []byte) ([]byte, error)
	Query(ctx context.Context, command []byte) ([]byte, error)
	Status(ctx context.Context) (*raftpb.Status, error)
}

type RaftEngine struct {
	Engine

	raftCli RaftClient
}

func NewRaftEngine(raftCli RaftClient) *RaftEngine {
	return &RaftEngine{raftCli: raftCli}
}

type RaftTxn struct {
	Txn
}
