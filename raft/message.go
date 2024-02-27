package raft

import (
	"encoding/binary"
	"hash/fnv"
)

// Address is the address of a message.
// Only one of Broadcast, NodeID, and Client is set.
type Address struct {
	// Broadcast is true if the message is being sent to all peers.
	Broadcast bool
	// NodeID is the ID of the specified node(local or remote).
	NodeID NodeID
	// Client is true if the message is being sent or received by a client.
	// In this case, only ClientRequest and ClientResponse events are allowed.
	Client bool
}

func (a Address) Equal(other Address) bool {
	return a.Broadcast == other.Broadcast && a.NodeID == other.NodeID && a.Client == other.Client
}

func (a Address) Hash() uint64 {
	h := fnv.New64a()
	binary.Write(h, binary.BigEndian, a.Broadcast)
	binary.Write(h, binary.BigEndian, uint64(a.NodeID))
	binary.Write(h, binary.BigEndian, a.Client)
	return h.Sum64()
}

type Message struct {
	Term  Term
	From  Address
	To    Address
	Event Event
}

type Event interface {
	isEvent()
}

var _ Event = &Heartbeat{}
var _ Event = &ConfirmLeader{}
var _ Event = &SolicitVote{}
var _ Event = &GrantVote{}
var _ Event = &AppendEntries{}
var _ Event = &RejectEntries{}
var _ Event = &ClientRequest{}
var _ Event = &ClientResponse{}

func (*Heartbeat) isEvent()      {}
func (*ConfirmLeader) isEvent()  {}
func (*SolicitVote) isEvent()    {}
func (*GrantVote) isEvent()      {}
func (*AppendEntries) isEvent()  {}
func (*RejectEntries) isEvent()  {}
func (*ClientRequest) isEvent()  {}
func (*ClientResponse) isEvent() {}

type Heartbeat struct {
	CommitIndex Index
	CommitTerm  Term
}

type ConfirmLeader struct {
	CommitIndex  Index
	HasCommitted bool
}

type SolicitVote struct {
	LastIndex Index
	LastTerm  Term
}

type GrantVote struct{}

type AppendEntries struct {
	BaseIndex Index
	BaseTerm  Term
	Entries   []Entry
}

type RejectEntries struct{}

type ClientRequest struct {
	ID      RequestID
	Type    RequestType
	Payload []byte
}

type ClientResponse struct {
	ID      RequestID
	Type    RequestType
	Payload []byte
	Error   error
}

type RequestType uint8

const (
	RequestTypeQuery RequestType = iota
	RequestTypeMutate
	RequestTypeStatus
)
