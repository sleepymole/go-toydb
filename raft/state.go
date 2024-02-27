package raft

import (
	"github.com/sleepymole/go-toydb/util/set"
)

// State represents a raft-machine state machine.
type State interface {
	// AppliedIndex returns the last index from the state machine.
	AppliedIndex() Index
	// Apply applies a log to the state machine. The log entry is
	// always applied no matter what the result is (success or failure).
	// The returned value is the result and the error during the application
	// of the log entry.
	Apply(entry *Entry) ([]byte, error)
	// Query queries the state machine with the given command.
	Query(command []byte) ([]byte, error)
}

type Instruction interface {
	isInstruction()
}

var _ Instruction = &AbortInstruction{}
var _ Instruction = &ApplyInstruction{}
var _ Instruction = &NotifyInstruction{}
var _ Instruction = &QueryInstruction{}
var _ Instruction = &StatusInstruction{}
var _ Instruction = &VoteInstruction{}

func (*AbortInstruction) isInstruction()  {}
func (*ApplyInstruction) isInstruction()  {}
func (*NotifyInstruction) isInstruction() {}
func (*QueryInstruction) isInstruction()  {}
func (*StatusInstruction) isInstruction() {}
func (*VoteInstruction) isInstruction()   {}

type AbortInstruction struct{}

type ApplyInstruction struct {
	Entry *Entry
}

type NotifyInstruction struct {
	ID      []byte
	Address Address
	Index   Index
}

type QueryInstruction struct {
	ID      []byte
	Address Address
	Command []byte
	Term    Term
	Index   Index
	Quorum  uint8
}

type StatusInstruction struct {
	ID      []byte
	Address Address
	Status  *Status
}

type VoteInstruction struct {
	Term    Term
	Index   Index
	Address Address
}

type Query struct {
	ID      []byte
	Term    Term
	Address Address
	Command []byte
	Quorum  uint8
	Votes   set.HashSet[Address]
}

// Driver is a driver for driving the raft state machine.
// Taking operations from the stateCh and sending results
// via the nodeCh.
type Driver struct {
	nodeID  NodeID
	stateCh <-chan Instruction
	nodeCh  chan<- *Message
	notify  map[Index]struct {
		address Address
		id      []byte
	}
	// queries any
}

func NewDriver(nodeID NodeID, stateCh <-chan Instruction, nodeCh chan<- *Message) *Driver {
	return &Driver{
		nodeID:  nodeID,
		stateCh: stateCh,
		nodeCh:  nodeCh,
		notify: make(map[Index]struct {
			address Address
			id      []byte
		}),
	}
}

func (d *Driver) Drive(state State) error {
	panic("implement me")
}

func (d *Driver) ApplyLog(state State, log *Log) error {
	panic("implement me")
}

func (d *Driver) Apply(state State, entry *Entry) error {
	panic("implement me")
}
