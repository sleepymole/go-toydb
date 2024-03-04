package raft

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/emirpasic/gods/v2/sets"
	"github.com/emirpasic/gods/v2/sets/hashset"
	"github.com/samber/mo"
	"github.com/sleepymole/go-toydb/storage"
	"github.com/sleepymole/go-toydb/util/assert"
	"github.com/sleepymole/go-toydb/util/itertools"
)

var ErrAbort = errors.New("raft: abort")

type NodeID uint8

const (
	defaultHeartbeatTick = 3
	defaultElectionTick  = 10
)

func electionTimeout() int {
	return defaultElectionTick + rand.N(defaultElectionTick)
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

const (
	Follower Role = iota
	Candidate
	Leader
)

type Node struct {
	id      NodeID
	peers   sets.Set[NodeID]
	term    Term
	log     *Log
	state   State
	nodeCh  chan<- *Message
	stateCh chan<- Instruction
	role    Role

	// The following fields are used when the node is a follower.
	leader     mo.Option[NodeID]
	leaderSeen int
	votedFor   mo.Option[NodeID]
	forwarded  sets.Set[RequestID]

	// The following fields are used when the node is a candidate.
	electionElapsed int
	electionTimeout int
	votesReceived   sets.Set[NodeID]

	// The following fields are used when the node is a leader.
	progress         map[NodeID]*progress
	heartbeatElapsed int
}

type progress struct {
	// next is the next index to replicate to the peer.
	next Index
	// last is the last index known to be replicated to the peer.
	// The indexes between last and next are in flight.
	last Index
}

func NewNode(
	id NodeID,
	peers sets.Set[NodeID],
	log *Log,
	state State,
	nodeCh chan<- *Message,
) (*Node, error) {
	stateCh := make(chan Instruction, 1)
	driver := NewDriver(id, stateCh, nodeCh)
	if err := driver.ApplyLog(state, log); err != nil {
		return nil, err
	}
	go driver.Drive(state)

	term, votedFor, err := log.GetTerm()
	if err != nil {
		return nil, err
	}
	n := &Node{
		id:       id,
		peers:    peers,
		term:     term,
		log:      log,
		state:    state,
		nodeCh:   nodeCh,
		stateCh:  stateCh,
		role:     Follower,
		votedFor: votedFor,
	}
	return n, nil
}

func (n *Node) ID() NodeID {
	return n.id
}

func (n *Node) Role() Role {
	return n.role
}

func (n *Node) Step(m *Message) error {
	// Drop messages from past terms.
	if m.Term < n.term && m.Term > 0 {
		return nil
	}

	// If we receive a message from a future term, become a leaderless
	// follower in it and step the message. If the message is a Heartbeat
	// or AppendEntries from the leader, stepping it will follow the leader.
	if m.Term > n.term {
		if err := n.becomeFollower(mo.None[NodeID](), m.Term); err != nil {
			return err
		}
	}

	switch n.role {
	case Follower:
		return n.stepFollower(m)
	case Candidate:
		return n.stepCadidate(m)
	case Leader:
		return n.stepLeader(m)
	default:
		panic("unreachable")
	}
}

func (n *Node) stepFollower(m *Message) error {
	assert.True(m.Client || m.Term == n.term)

	// Record when we last saw a message from the leader (if any).
	if n.isLeader(m.From) {
		n.leaderSeen = 0
	}

	switch e := m.Event.(type) {
	case *Heartbeat:
		if n.leader.IsAbsent() {
			if err := n.becomeFollower(mo.Some(m.From), m.Term); err != nil {
				return err
			}
		} else if n.leader.MustGet() != m.From {
			return fmt.Errorf("received heartbeat from unexpected leader %d", m.From)
		}

		// Advance commit index and apply entries if possible.
		hasCommitted, err := n.log.Has(e.CommitIndex, e.CommitTerm)
		if err != nil {
			return err
		}
		oldCommitIndex, _ := n.log.CommitIndex()
		if hasCommitted && e.CommitIndex > oldCommitIndex {
			if err := n.log.Commit(e.CommitIndex); err != nil {
				return err
			}
			it, err := n.log.Scan(oldCommitIndex+1, e.CommitIndex+1)
			if err != nil {
				return err
			}
			if err := itertools.Walk(it, func(entry *Entry) error {
				n.stateCh <- &ApplyInstruction{Entry: entry}
				return nil
			}); err != nil {
				return err
			}
		}
		return n.send(m.From, &ConfirmLeader{
			CommitIndex:  e.CommitIndex,
			HasCommitted: hasCommitted,
		})
	case *AppendEntries:
		if n.leader.IsAbsent() {
			if err := n.becomeFollower(mo.Some(m.From), m.Term); err != nil {
				return err
			}
		} else if n.leader.MustGet() != m.From {
			return fmt.Errorf("received heartbeat from unexpected leader %d", m.From)
		}
		// Append the entries, if possible.
		if e.BaseIndex > 0 {
			hasBase, err := n.log.Has(e.BaseIndex, e.BaseTerm)
			if err != nil {
				return err
			}
			if !hasBase {
				return n.send(m.From, &RejectEntries{})
			}
		}
		lastIndex, err := n.log.Splice(e.Entries)
		if err != nil {
			return err
		}
		return n.send(m.From, &AcceptEntries{
			LastIndex: lastIndex,
		})
	case *SolicitVote:
		// If we already voted for someone else, ignore it.
		if n.votedFor.IsPresent() && n.votedFor.MustGet() != m.From {
			return nil
		}
		// Only vote if the candidate's log is at least as up-to-date as ours.
		logIndex, logTerm := n.log.LastIndex()
		if e.LastTerm > logTerm || e.LastTerm == logTerm && e.LastIndex >= logIndex {
			if err := n.log.SetTerm(n.term, mo.Some(m.From)); err != nil {
				return err
			}
			n.votedFor = mo.Some(m.From)
			return n.send(m.From, &GrantVote{})
		}
		return nil
	case *GrantVote:
		// We may receive a vote after we lost the election and followed
		// a different leader. Ignore it.
		return nil
	case *ClientRequest:
		assert.True(m.Client, "client request from non-client")
		if n.leader.IsPresent() {
			n.forwarded.Add(e.ID)
			return n.send(m.From, m.Event)
		} else {
			return n.sendToClient(&ClientResponse{
				ID:  e.ID,
				Err: ErrAbort,
			})
		}
	case *ClientResponse:
		assert.True(n.isLeader(m.From), "client response from non-leader")
		if n.forwarded.Contains(e.ID) {
			n.forwarded.Remove(e.ID)
			return n.sendToClient(m.Event)
		}
		return nil
	default:
		return fmt.Errorf("received unexpected event %T", e)
	}
}

func (n *Node) stepCadidate(m *Message) error {
	assert.True(m.Client || m.Term == n.term)

	_, isHeartbeat := m.Event.(*Heartbeat)
	_, isAppendEntries := m.Event.(*AppendEntries)
	// If we receive a heartbeat or append entries in this term, we lost the
	// election and have a new leader. Follow the leader and step the message.
	if isHeartbeat || isAppendEntries {
		if err := n.becomeFollower(mo.Some(m.From), m.Term); err != nil {
			return err
		}
		return n.stepFollower(m)
	}

	switch e := m.Event.(type) {
	case *SolicitVote:
		// Ignore other candidates' solicitations.
		return nil
	case *GrantVote:
		n.votesReceived.Add(m.From)
		if n.votesReceived.Size() >= n.quorum() {
			return n.becomeLeader()
		}
		return nil
	case *ClientRequest:
		assert.True(m.Client, "client request from non-client")
		// Abort any inbound requests when campaigning.
		return n.sendToClient(&ClientResponse{
			ID:  e.ID,
			Err: ErrAbort,
		})
	default:
		return fmt.Errorf("received unexpected event %T", e)
	}
}

func (n *Node) stepLeader(m *Message) error {
	assert.True(m.Client || m.Term == n.term)

	switch e := m.Event.(type) {
	case *Heartbeat:
		return fmt.Errorf("saw other leader %d in term %d", m.From, m.Term)
	case *AppendEntries:
		return fmt.Errorf("saw other leader %d in term %d", m.From, m.Term)
	case *ConfirmLeader:
		// A follower received one of our heartbeats and confirmed that we
		// are its leader. If it doesn't have the commit index in its local
		// log, replicate the log to it.
		n.stateCh <- &VoteInstruction{
			Term:   m.Term,
			Index:  e.CommitIndex,
			NodeID: m.From,
		}
		if !e.HasCommitted {
			return n.sendLog(m.From)
		}
		return nil
	case *AcceptEntries:
		// A follower appended log entries we sent it. Record its progress
		// and attempt to commit new entries.
		lastIndex, _ := n.log.LastIndex()
		assert.True(e.LastIndex <= lastIndex, "follower's last index is ahead of ours")

		if e.LastIndex > n.progress[m.From].next {
			n.progress[m.From].last = e.LastIndex
			n.progress[m.From].next = e.LastIndex + 1
			_, err := n.maybeCommit()
			return err
		}
		return nil
	case *RejectEntries:
		// A follower rejected log entries we sent it, typically because it
		// does not have the base index in its log. Try to replicate from
		// the previous entry.
		//
		// Here perform linear probing, as described in the raft paper, can
		// be very slow with long divergent logs, but we keep it simple.
		n.progress[m.From].next--
		return n.sendLog(m.From)
	case *ClientRequest:
		switch e.Type {
		case RequestQuery:
			panic("implement me")
		case RequestMutate:
			panic("implement me")
		case RequestStatus:
			panic("implement me")
		default:
			return fmt.Errorf("unknown request type %d", e.Type)
		}
	case *SolicitVote:
		// Ignore it since we won the election.
		return nil
	case *GrantVote:
		// Ignore it since we won the election.
		return nil
	default:
		return fmt.Errorf("received unexpected event %T", e)
	}
}

func (n *Node) Tick() error {
	switch n.role {
	case Follower:
		n.leaderSeen++
		if n.leaderSeen >= defaultElectionTick {
			return n.becomeCadidate()
		}
		return nil
	case Candidate:
		n.electionElapsed++
		if n.electionElapsed >= n.electionTimeout {
			return n.campaign()
		}
		return nil
	case Leader:
		n.heartbeatElapsed++
		if n.heartbeatElapsed >= defaultHeartbeatTick {
			return n.heartbeat()
		}
		return nil
	default:
		panic("unreachable")
	}
}

func (n *Node) becomeFollower(leader mo.Option[NodeID], term Term) error {
	assert.True(term >= n.term)

	if err := n.abortForwarded(); err != nil {
		return err
	}
	n.forwarded = hashset.New[RequestID]()

	if leader.IsPresent() {
		if term == n.term && n.leader != leader {
			return fmt.Errorf(
				"multiple leaders in same term %d, %d vs %d",
				n.term, n.leader.MustGet(), leader.MustGet(),
			)
		}
		n.term = term
		n.leader = leader
		n.votedFor = leader
	} else {
		if term == n.term {
			return fmt.Errorf("cannot become leaderless follower in current term %d", n.term)
		}
		n.term = term
		n.leader = mo.None[NodeID]()
		n.votedFor = mo.None[NodeID]()
	}
	return n.log.SetTerm(n.term, n.votedFor)
}

func (n *Node) becomeCadidate() error {
	if err := n.abortForwarded(); err != nil {
		return err
	}
	n.role = Candidate
	n.electionTimeout = electionTimeout()
	n.votesReceived.Clear()
	return n.campaign()
}

func (n *Node) becomeLeader() error {
	panic("implement me")
}

func (n *Node) isLeader(id NodeID) bool {
	return n.leader.IsPresent() && n.leader.MustGet() == id
}

func (n *Node) quorum() int {
	return n.peers.Size()/2 + 1
}

func (n *Node) campaign() error {
	panic("implement me")
}

func (n *Node) heartbeat() error {
	commitIndex, commitTerm := n.log.CommitIndex()
	n.peers.Add()
	for _, peer := range n.peers.Values() {
		if err := n.send(peer, &Heartbeat{
			CommitIndex: commitIndex,
			CommitTerm:  commitTerm,
		}); err != nil {
			return err
		}
	}
	n.heartbeatElapsed = 0
	return nil
}

func (n *Node) propose(command []byte) (Index, error) {
	index, err := n.log.Append(n.term, command)
	if err != nil {
		return 0, err
	}
	for _, peer := range n.peers.Values() {
		if err := n.sendLog(peer); err != nil {
			return 0, err
		}
	}
	return index, nil
}

func (n *Node) maybeCommit() (Index, error) {
	var lastIndexes []Index
	for _, progress := range n.progress {
		lastIndexes = append(lastIndexes, progress.last)
	}
	slices.SortFunc(lastIndexes, func(a, b Index) int {
		return cmp.Compare(b, a)
	})
	commitIndex := lastIndexes[n.quorum()-1]

	// A 0 commit index means we haven't committed anything yet.
	if commitIndex == 0 {
		return 0, nil
	}
	// Make sure the commit index does not regress.
	oldCommitIndex, _ := n.log.CommitIndex()
	assert.True(
		oldCommitIndex <= commitIndex,
		"commit index regressed from %d to %d",
		oldCommitIndex, commitIndex,
	)

	entry, err := n.log.Get(commitIndex)
	if err != nil {
		if err == storage.ErrNotFound {
			return 0, fmt.Errorf("missing entry at index %d", commitIndex)
		}
		return 0, err
	}
	// We can only safely commit up to an entry from our own term,
	// see figure 8 in the raft paper.
	if entry.Term != n.term {
		return oldCommitIndex, nil
	}

	if commitIndex > oldCommitIndex {
		if err := n.log.Commit(commitIndex); err != nil {
			return 0, err
		}
		it, err := n.log.Scan(oldCommitIndex+1, commitIndex+1)
		if err != nil {
			return 0, err
		}
		if err := itertools.Walk(it, func(entry *Entry) error {
			n.stateCh <- &ApplyInstruction{Entry: entry}
			return nil
		}); err != nil {
			return 0, err
		}
	}
	return commitIndex, nil
}

func (n *Node) abortForwarded() error {
	for _, id := range n.forwarded.Values() {
		if err := n.sendToClient(&ClientResponse{
			ID:  id,
			Err: ErrAbort,
		}); err != nil {
			return err
		}
	}
	n.forwarded.Clear()
	return nil
}

func (n *Node) send(to NodeID, event Event) error {
	n.nodeCh <- &Message{
		Term:  n.term,
		From:  n.id,
		To:    to,
		Event: event,
	}
	return nil
}

func (n *Node) sendToClient(event Event) error {
	n.nodeCh <- &Message{
		Term:   n.term,
		Client: true,
		Event:  event,
	}
	return nil
}

func (n *Node) sendLog(peer NodeID) error {
	var (
		baseIndex Index
		baseTerm  Term
	)
	progress, ok := n.progress[peer]
	if !ok {
		return fmt.Errorf("unknown peer %d", peer)
	}
	if progress.next > 1 {
		entry, err := n.log.Get(progress.next - 1)
		if err != nil {
			if err == storage.ErrNotFound {
				return fmt.Errorf("missing entry at index %d", progress.next-1)
			}
			return err
		}
		baseIndex = progress.next - 1
		baseTerm = entry.Term
	}

	it, err := n.log.Scan(baseIndex+1, math.MaxUint64)
	if err != nil {
		return err
	}
	entries, err := itertools.ToSlice(it)
	if err != nil {
		return err
	}
	return n.send(peer, &AppendEntries{
		BaseIndex: baseIndex,
		BaseTerm:  baseTerm,
		Entries:   entries,
	})
}
