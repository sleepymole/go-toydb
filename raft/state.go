package raft

import (
	"errors"

	"github.com/emirpasic/gods/v2/sets"
)

// InternalError is an error that is used to wrap internal errors.
// It is used to distinguish internal errors from user errors.
type InternalError struct {
	err error
}

func (e *InternalError) Error() string {
	return e.err.Error()
}

func (e *InternalError) Unwrap() error {
	return e.err
}

func MakeInternalError(err error) error {
	return &InternalError{err: err}
}

func IsInternalError(err error) bool {
	var e *InternalError
	return errors.As(err, &e)
}

// State represents a raft state machine. The caller of
// raft is responsible for implementing the state machine.
type State interface {
	// AppliedIndex returns the last applied index.
	AppliedIndex() Index
	// Apply applies a log to the state machine. Apply should
	// be deterministic and idempotent.
	//
	// If InternalError is returned, the raft node will be terminated.
	// Any other error is considered applied and returned to the caller.
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
	ID     RequestID
	NodeID NodeID
	Index  Index
}

type QueryInstruction struct {
	ID      RequestID
	NodeID  NodeID
	Command RequestID
	Term    Term
	Index   Index
	Quorum  uint8
}

type StatusInstruction struct {
	ID     RequestID
	NodeID NodeID
	Status *Status
}

type VoteInstruction struct {
	Term   Term
	Index  Index
	NodeID NodeID
}

type Query struct {
	ID      RequestID
	Term    Term
	NodeID  NodeID
	Command []byte
	Quorum  uint8
	Votes   sets.Set[NodeID]
}

// Driver is a driver for driving the raft state machine.
// Taking operations from the stateCh and sending results
// via the nodeCh.
type Driver struct {
	nodeID  NodeID
	stateCh <-chan Instruction
	nodeCh  chan<- *Message
	notify  map[Index]struct {
		id     []byte
		nodeID NodeID
	}
	// queries any
}

func NewDriver(nodeID NodeID, stateCh <-chan Instruction, nodeCh chan<- *Message) *Driver {
	return &Driver{
		nodeID:  nodeID,
		stateCh: stateCh,
		nodeCh:  nodeCh,
		notify: make(map[Index]struct {
			id     []byte
			nodeID NodeID
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
