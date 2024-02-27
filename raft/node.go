package raft

import (
	"math/rand/v2"

	"github.com/sleepymole/go-toydb/util/set"
)

type (
	NodeID uint8
	Ticks  uint8
)

const (
	heartbeatInterval       Ticks = 3
	electionTimeoutRangeMin Ticks = 10
	electionTimeoutRangeMax Ticks = 20
)

func randElectionTimeout() Ticks {
	return electionTimeoutRangeMin + rand.N(electionTimeoutRangeMax-electionTimeoutRangeMin)
}

type Status struct {
	Server        NodeID
	Leader        NodeID
	Term          Term
	NodeLastIndex map[NodeID]Index
	CommitIndex   Index
	ApplyIndex    Index
	Storage       string
	StorageSize   int64
}

type Role uint8

type Node interface {
	ID() NodeID
	Step(msg *Message) (Node, error)
	Tick() (Node, error)
}

var _ Node = &Follower{}
var _ Node = &Candidate{}
var _ Node = &Leader{}

type nodeBase struct {
	id    NodeID
	peers set.Set[NodeID]
	log   *Log
}

func (b *nodeBase) ID() NodeID {
	return b.id
}

func (b *nodeBase) init(
	id NodeID,
	peers set.Set[NodeID],
	log *Log,
	state State,
	nodeCh chan<- *Message,
) error {
	b.id = id
	b.peers = peers
	b.log = log

	stateCh := make(chan Instruction, 16)
	d := NewDriver(id, stateCh, nodeCh)
	if err := d.ApplyLog(state, log); err != nil {
		return err
	}
	go d.Drive(state)
	return nil
}

type Follower struct {
	nodeBase

	leader          NodeID
	leaderSeen      Ticks
	electionTimeout Ticks
	votedFor        NodeID
	forwarded       set.Set[RequestID]
}

func (f *Follower) Step(msg *Message) (Node, error) {
	panic("implement me")
}

func (f *Follower) Tick() (Node, error) {
	panic("implement me")
}

type Candidate struct {
	nodeBase

	votes            set.Set[NodeID]
	electionDuration Ticks
	electionTimeout  Ticks
}

func (c *Candidate) Step(msg *Message) (Node, error) {
	panic("implement me")
}

func (c *Candidate) Tick() (Node, error) {
	panic("implement me")
}

type progress struct {
	next  Index
	match Index
}

type Leader struct {
	nodeBase

	progress       map[NodeID]progress
	sinceHeartbeat Ticks
}

func (l *Leader) Step(msg *Message) (Node, error) {
	panic("implement me")
}

func (l *Leader) Tick() (Node, error) {
	panic("implement me")
}
