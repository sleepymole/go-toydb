package raft

import (
	"errors"

	"github.com/emirpasic/gods/v2/sets/hashset"
	"github.com/sleepymole/go-toydb/api/raftpb"
	"github.com/sleepymole/go-toydb/util/assert"
	"github.com/sleepymole/go-toydb/util/itertools"
	"github.com/sleepymole/go-toydb/util/protoutil"
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
	Apply(entry *raftpb.Entry) ([]byte, error)
	// Query queries the state machine with the given command.
	Query(command []byte) ([]byte, error)
}

// Instruction is the instructions sent to driver.
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

// AbortInstruction is the instruction to abort any pending operations.
type AbortInstruction struct{}

// ApplyInstruction is the instruction to apply a log entry.
type ApplyInstruction struct {
	Entry *raftpb.Entry
}

// NotifyInstruction is the instruction to notify the given node with
// the result of applying the entry at the given index.
type NotifyInstruction struct {
	ID     RequestID
	NodeID NodeID
	Index  Index
}

// QueryInstruction is the instruction to query the state machine when
// the given term and index has been confirmed by vote.
type QueryInstruction struct {
	ID      RequestID
	NodeID  NodeID
	Command RequestID
	Term    Term
	Index   Index
	Quorum  int
}

// StatusQuery is the instruction to extend the given server status and
// return it the the given node.
type StatusInstruction struct {
	ID     RequestID
	NodeID NodeID
	Status *raftpb.Status
}

// VoteInstruction is the instruction to vote for the given term and index.
type VoteInstruction struct {
	Term   Term
	Index  Index
	NodeID NodeID
}

type Query struct {
	ID      RequestID
	NodeID  NodeID
	Command []byte
	Term    Term
	Index   Index
	Quorum  int
	Votes   *hashset.Set[NodeID]
}

func (q *Query) Ready() bool {
	return q.Votes.Size() >= int(q.Quorum)
}

type Notify struct {
	ID     RequestID
	NodeID NodeID
	Index  Index
}

// Driver is a driver for driving the raft state machine.
// Taking operations from the stateCh and sending results
// via the nodeCh.
type Driver struct {
	nodeID   NodeID
	stateCh  <-chan Instruction
	msgCh    chan<- *raftpb.Message
	notifies []*Notify
	queries  []*Query
}

func NewDriver(nodeID NodeID, stateCh <-chan Instruction, msgCh chan<- *raftpb.Message) *Driver {
	return &Driver{
		nodeID:  nodeID,
		stateCh: stateCh,
		msgCh:   msgCh,
	}
}

func (d *Driver) Drive(state State) error {
	for instruction := range d.stateCh {
		if err := d.execute(instruction, state); err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) ApplyLog(state State, log *Log) error {
	appliedIndex := state.AppliedIndex()
	commitIndex, _ := log.CommitIndex()
	assert.True(
		appliedIndex <= commitIndex,
		"applied index(%d) above commit index(%d)",
		appliedIndex, commitIndex,
	)
	if appliedIndex < commitIndex {
		scan, err := log.Scan(appliedIndex+1, commitIndex+1)
		if err != nil {
			return err
		}
		if err := itertools.Walk(scan, func(entry *raftpb.Entry) error {
			return d.Apply(state, entry)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) Apply(state State, entry *raftpb.Entry) error {
	result, err := state.Apply(entry)
	if IsInternalError(err) {
		return err
	}
	if err := d.notifyApplied(entry.Index, result, err); err != nil {
		return err
	}
	return d.queryExecute(state)
}

func (d *Driver) execute(instruction Instruction, state State) error {
	switch instruction := instruction.(type) {
	case *AbortInstruction:
		if err := d.notifyAbort(); err != nil {
			return err
		}
		if err := d.queryAbort(); err != nil {
			return err
		}
		return nil
	case *ApplyInstruction:
		return d.Apply(state, instruction.Entry)
	case *NotifyInstruction:
		if instruction.Index > state.AppliedIndex() {
			d.notifies = append(d.notifies, &Notify{
				ID:     instruction.ID,
				NodeID: instruction.NodeID,
				Index:  instruction.Index,
			})
		} else {
			if err := d.sendResponse(instruction.NodeID, &raftpb.ClientResponse{
				Id:    instruction.ID,
				Type:  raftpb.ClientRequest_MUTATE,
				Error: requestAborted,
			}); err != nil {
				return err
			}
		}
		return nil
	case *QueryInstruction:
		d.queries = append(d.queries, &Query{
			ID:      instruction.ID,
			NodeID:  instruction.NodeID,
			Command: instruction.Command,
			Quorum:  instruction.Quorum,
			Term:    instruction.Term,
			Index:   instruction.Index,
			Votes:   hashset.New[NodeID](),
		})
		return nil
	case *StatusInstruction:
		status := protoutil.Clone(instruction.Status)
		status.ApplyIndex = state.AppliedIndex()
		return d.sendResponse(instruction.NodeID, &raftpb.ClientResponse{
			Id:     instruction.ID,
			Type:   raftpb.ClientRequest_STATUS,
			Status: status,
		})
	case *VoteInstruction:
		d.queryVote(instruction.Term, instruction.Index, instruction.NodeID)
		return d.queryExecute(state)
	default:
		panic("unreachable")
	}
}

func (d *Driver) notifyAbort() error {
	for _, notify := range d.notifies {
		if err := d.sendResponse(notify.NodeID, &raftpb.ClientResponse{
			Id:    notify.ID,
			Type:  raftpb.ClientRequest_MUTATE,
			Error: requestAborted,
		}); err != nil {
			return err
		}
	}
	d.notifies = nil
	return nil
}

func (d *Driver) notifyApplied(index Index, result []byte, err error) error {
	var ready, pending []*Notify
	for _, notify := range d.notifies {
		if notify.Index == index {
			ready = append(ready, notify)
		} else {
			pending = append(pending, notify)
		}
	}
	d.notifies = pending

	for _, notify := range ready {
		if err := d.sendResponse(notify.NodeID, &raftpb.ClientResponse{
			Id:     notify.ID,
			Type:   raftpb.ClientRequest_MUTATE,
			Result: result,
			Error:  requestAborted,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (d *Driver) queryAbort() error {
	for _, query := range d.queries {
		if err := d.sendResponse(query.NodeID, &raftpb.ClientResponse{
			Id:    query.ID,
			Type:  raftpb.ClientRequest_QUERY,
			Error: requestAborted,
		}); err != nil {
			return err
		}
	}
	d.queries = nil
	return nil
}

func (d *Driver) queryExecute(state State) error {
	var ready, pending []*Query
	for _, query := range d.queries {
		if query.Index <= state.AppliedIndex() {
			ready = append(ready, query)
		} else {
			pending = append(pending, query)
		}
	}
	d.queries = pending

	for _, query := range ready {
		result, err := state.Query(query.Command)
		if err != nil {
			if err := d.sendResponse(query.NodeID, &raftpb.ClientResponse{
				Id:    query.ID,
				Type:  raftpb.ClientRequest_QUERY,
				Error: requestAborted,
			}); err != nil {
				return err
			}
		} else {
			if err := d.sendResponse(query.NodeID, &raftpb.ClientResponse{
				Id:     query.ID,
				Type:   raftpb.ClientRequest_QUERY,
				Result: result,
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *Driver) queryVote(term Term, index Index, nodeID NodeID) {
	for _, query := range d.queries {
		if term >= query.Term && index >= query.Index {
			query.Votes.Add(nodeID)
		}
	}
}

func (d *Driver) sendResponse(to NodeID, resp *raftpb.ClientResponse) error {
	d.msgCh <- &raftpb.Message{
		From: d.nodeID,
		To:   to,
		Term: 0, // TODO: set correct term
		Event: &raftpb.Message_ClientResponse{
			ClientResponse: resp,
		},
	}
	return nil
}
