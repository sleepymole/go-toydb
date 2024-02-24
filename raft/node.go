package raft

import "math/rand/v2"

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
	Step() Node
	Tick() Node
}
