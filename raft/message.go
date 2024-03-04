package raft

type Message struct {
	Term Term
	From NodeID
	To   NodeID
	// If Client is true, it indicates that the message is
	// sent to or received from local client. In the case,
	// The Event must be ClientRequest or ClientResponse,
	// and From and To are zero.
	Client bool
	Event  Event
}

type Event interface {
	isEvent()
}

var _ Event = &Heartbeat{}
var _ Event = &ConfirmLeader{}
var _ Event = &SolicitVote{}
var _ Event = &GrantVote{}
var _ Event = &AppendEntries{}
var _ Event = &AcceptEntries{}
var _ Event = &RejectEntries{}
var _ Event = &ClientRequest{}
var _ Event = &ClientResponse{}

func (*Heartbeat) isEvent()      {}
func (*ConfirmLeader) isEvent()  {}
func (*SolicitVote) isEvent()    {}
func (*GrantVote) isEvent()      {}
func (*AppendEntries) isEvent()  {}
func (*AcceptEntries) isEvent()  {}
func (*RejectEntries) isEvent()  {}
func (*ClientRequest) isEvent()  {}
func (*ClientResponse) isEvent() {}

// Heartbeat is the event sent by the leader to the followers
// to maintain the leader's authority.
type Heartbeat struct {
	CommitIndex Index
	CommitTerm  Term
}

// ConfirmLeader is the event sent by the follower to the leader
// to confirm the leader's authority after receiving a heartbeat.
type ConfirmLeader struct {
	CommitIndex  Index
	HasCommitted bool
}

// SolicitVote is the event sent by the candidate to the followers
// to request their votes.
type SolicitVote struct {
	LastIndex Index
	LastTerm  Term
}

// GrantVote is the event sent by the follower to the candidate
// to grant its vote.
type GrantVote struct{}

// AppendEntries is the event sent by the leader to the followers
// to replicate log entries.
type AppendEntries struct {
	BaseIndex Index
	BaseTerm  Term
	Entries   []*Entry
}

// AcceptEntries is the event sent by the follower to the leader
// to accept a set of log entries.
type AcceptEntries struct {
	LastIndex Index
}

// RejectEntries is the event sent by the follower to the leader
// to reject a set of log entries.
type RejectEntries struct{}

type RequestID string

type RequestType uint8

const (
	RequestQuery RequestType = iota
	RequestMutate
	RequestStatus
)

type ClientRequest struct {
	ID      RequestID
	Type    RequestType
	Command []byte
}

type ClientResponse struct {
	ID      RequestID
	Type    RequestType
	Payload []byte
	Status  *Status
	Err     error
}
