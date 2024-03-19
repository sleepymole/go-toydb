package raft

import (
	"bytes"
	"cmp"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"

	"github.com/emirpasic/gods/v2/sets"
	"github.com/emirpasic/gods/v2/sets/treeset"
	"github.com/sleepymole/go-toydb/api/raftpb"
	"github.com/sleepymole/go-toydb/storage"
	"github.com/sleepymole/go-toydb/util/assert"
	"github.com/sleepymole/go-toydb/util/itertools"
)

const (
	defaultHeartbeatTick = 3
	defaultElectionTick  = 10
	requestAborted       = "request aborted"
)

func electionTimeout() int {
	return defaultElectionTick + rand.N(defaultElectionTick)
}

type (
	NodeID    = uint64
	Term      = uint64
	Index     = uint64
	RequestID = []byte
)

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
	msgCh   chan<- *raftpb.Message
	stateCh chan<- Instruction
	role    Role

	// The following fields are used when the node is a follower.
	leader     NodeID
	leaderSeen int
	votedFor   NodeID
	forwarded  *treeset.Set[RequestID]

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
	msgCh chan<- *raftpb.Message,
) (*Node, error) {
	stateCh := make(chan Instruction, 1)
	driver := NewDriver(id, stateCh, msgCh)
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
		msgCh:    msgCh,
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

func (n *Node) Step(m *raftpb.Message) error {
	// Drop messages from past terms.
	if m.Term < n.term && m.Term > 0 {
		return nil
	}

	// If we receive a message from a future term, become a leaderless
	// follower in it and step the message. If the message is a Heartbeat
	// or AppendEntries from the leader, stepping it will follow the leader.
	if m.Term > n.term {
		if err := n.becomeFollower(0, m.Term); err != nil {
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

func (n *Node) stepFollower(m *raftpb.Message) error {
	assert.True(m.Term == 0 || m.Term == n.term)

	// Record when we last saw a message from the leader (if any).
	if n.leader != 0 && n.leader == m.From {
		n.leaderSeen = 0
	}

	if event := m.GetHeartbeat(); event != nil {
		if n.leader == 0 {
			if err := n.becomeFollower(m.From, m.Term); err != nil {
				return err
			}
		} else if n.leader != m.From {
			return fmt.Errorf("received heartbeat from unexpected leader %d", m.From)
		}

		// Advance commit index and apply entries if possible.
		hasCommitted, err := n.log.Has(event.CommitIndex, event.CommitTerm)
		if err != nil {
			return err
		}
		oldCommitIndex, _ := n.log.CommitIndex()
		if hasCommitted && event.CommitIndex > oldCommitIndex {
			if err := n.log.Commit(event.CommitIndex); err != nil {
				return err
			}
			it, err := n.log.Scan(oldCommitIndex+1, event.CommitIndex+1)
			if err != nil {
				return err
			}
			if err := itertools.Walk(it, func(entry *raftpb.Entry) error {
				n.stateCh <- &ApplyInstruction{Entry: entry}
				return nil
			}); err != nil {
				return err
			}
		}
		return n.send(m.From, &raftpb.ConfirmLeader{
			CommitIndex:  event.CommitIndex,
			HasCommitted: hasCommitted,
		})
	} else if event := m.GetAppendEntries(); event != nil {
		if n.leader == 0 {
			if err := n.becomeFollower(m.From, m.Term); err != nil {
				return err
			}
		} else if n.leader != m.From {
			return fmt.Errorf("received heartbeat from unexpected leader %d", m.From)
		}
		if len(event.Entries) > 0 {
			base := event.Entries[0].Index - 1
			if base > 0 {
				hasBase, err := n.log.Has(base, event.Entries[0].Term)
				if err != nil {
					return err
				}
				if !hasBase {
					return n.send(m.From, &raftpb.RejectEntries{})
				}
			}
		}
		lastIndex, err := n.log.Splice(event.Entries)
		if err != nil {
			return err
		}
		return n.send(m.From, &raftpb.AcceptEntries{
			LastIndex: lastIndex,
		})
	} else if event := m.GetSolicitVote(); event != nil {
		// If we already voted for someone else, ignore it.
		if n.votedFor != 0 && n.votedFor != m.From {
			return nil
		}
		// Only vote if the candidate's log is at least as up-to-date as ours.
		logIndex, logTerm := n.log.LastIndex()
		if event.LastTerm > logTerm || event.LastTerm == logTerm && event.LastIndex >= logIndex {
			if err := n.log.SetTerm(n.term, m.From); err != nil {
				return err
			}
			n.votedFor = m.From
			return n.send(m.From, &raftpb.GrantVote{})
		}
		return nil
	} else if event := m.GetGrantVote(); event != nil {
		// We may receive a vote after we lost the election and followed
		// a different leader. Ignore it.
		return nil
	} else if event := m.GetClientRequest(); event != nil {
		assert.True(m.From != 0, "client request from non-client")
		if n.leader != 0 {
			n.forwarded.Add(event.Id)
			return n.send(m.From, event)
		} else {
			return n.sendToClient(&raftpb.ClientResponse{
				Id:    event.Id,
				Error: requestAborted,
			})
		}
	} else if event := m.GetClientResponse(); event != nil {
		assert.True(n.isLeader(m.From), "client response from non-leader")
		if n.forwarded.Contains(event.Id) {
			n.forwarded.Remove(event.Id)
			return n.sendToClient(event)
		}
		return nil
	} else {
		return fmt.Errorf("received unexpected event %T", m.GetEvent())
	}
}

func (n *Node) stepCadidate(m *raftpb.Message) error {
	assert.True(m.Term == 0 || m.Term == n.term)

	// If we receive a heartbeat or append entries in this term, we lost the
	// election and have a new leader. Follow the leader and step the message.
	if m.GetHeartbeat() != nil || m.GetAppendEntries() != nil {
		if err := n.becomeFollower(m.From, m.Term); err != nil {
			return err
		}
		return n.stepFollower(m)
	}

	if event := m.GetSolicitVote(); event != nil {
		// Ignore other candidates' solicitations.
		return nil
	} else if event := m.GetGrantVote(); event != nil {
		n.votesReceived.Add(m.From)
		if n.votesReceived.Size() >= n.quorum() {
			return n.becomeLeader()
		}
		return nil
	} else if event := m.GetClientRequest(); event != nil {
		assert.True(m.Term == 0, "client request from non-client")
		// Abort any inbound requests when campaigning.
		return n.sendToClient(&raftpb.ClientResponse{
			Id:    event.Id,
			Type:  event.Type,
			Error: requestAborted,
		})
	} else {
		return fmt.Errorf("received unexpected event %T", m.GetEvent())
	}
}

func (n *Node) stepLeader(m *raftpb.Message) error {
	assert.True(m.Term == 0 || m.Term == n.term)

	if event := m.GetHeartbeat(); event != nil {
		return fmt.Errorf("saw other leader %d in term %d", m.From, m.Term)
	} else if event := m.GetAppendEntries(); event != nil {
		return fmt.Errorf("saw other leader %d in term %d", m.From, m.Term)
	} else if event := m.GetConfirmLeader(); event != nil {
		// A follower received one of our heartbeats and confirmed that we
		// are its leader. If it doesn't have the commit index in its local
		// log, replicate the log to it.
		n.stateCh <- &VoteInstruction{
			Term:   m.Term,
			Index:  event.CommitIndex,
			NodeID: m.From,
		}
		if !event.HasCommitted {
			return n.sendLog(m.From)
		}
		return nil
	} else if event := m.GetAcceptEntries(); event != nil {
		// A follower appended log entries we sent it. Record its progress
		// and attempt to commit new entries.
		lastIndex, _ := n.log.LastIndex()
		assert.True(event.LastIndex <= lastIndex, "follower's last index is ahead of ours")

		if event.LastIndex > n.progress[m.From].next {
			n.progress[m.From].last = event.LastIndex
			n.progress[m.From].next = event.LastIndex + 1
			return n.maybeCommit()
		}
		return nil
	} else if event := m.GetRejectEntries(); event != nil {
		// A follower rejected log entries we sent it, typically because it
		// does not have the base index in its log. Try to replicate from
		// the previous entry.
		//
		// Here perform linear probing, as described in the raft paper, can
		// be very slow with long divergent logs, but we keep it simple.
		n.progress[m.From].next--
		return n.sendLog(m.From)
	} else if event := m.GetClientRequest(); event != nil {
		switch event.Type {
		case raftpb.ClientRequest_QUERY:
			commitIndex, _ := n.log.CommitIndex()
			n.stateCh <- &QueryInstruction{
				ID:      event.Id,
				NodeID:  m.From,
				Command: event.Command,
				Term:    n.term,
				Index:   commitIndex,
				Quorum:  n.quorum(),
			}
			n.stateCh <- &VoteInstruction{
				Term:   n.term,
				Index:  commitIndex,
				NodeID: n.id,
			}
			return n.heartbeat()
		case raftpb.ClientRequest_MUTATE:
			index, err := n.propose(event.Command)
			if err != nil {
				return err
			}
			n.stateCh <- &NotifyInstruction{
				ID:     event.Id,
				NodeID: m.From,
				Index:  index,
			}
			if n.peers.Empty() {
				return n.maybeCommit()
			}
			return nil
		case raftpb.ClientRequest_STATUS:
			engineStatus, err := n.log.Status()
			if err != nil {
				return err
			}
			nodeLastIndex := make(map[NodeID]Index)
			for peer, progress := range n.progress {
				nodeLastIndex[peer] = progress.last
			}
			nodeLastIndex[n.id], _ = n.log.LastIndex()
			commitIndex, _ := n.log.CommitIndex()
			status := &raftpb.Status{
				Server:        n.id,
				Leader:        n.id,
				Term:          n.term,
				NodeLastIndex: nodeLastIndex,
				CommitIndex:   commitIndex,
				Storage:       engineStatus.Name,
				StorageSize:   engineStatus.Size,
			}
			n.stateCh <- &StatusInstruction{
				ID:     event.Id,
				NodeID: m.From,
				Status: status,
			}
			return nil
		default:
			return fmt.Errorf("unknown request type %d", event.Type)
		}
	} else if event := m.GetSolicitVote(); event != nil {
		// Ignore it since we won the election.
		return nil
	} else if event := m.GetGrantVote(); event != nil {
		// Ignore it since we won the election.
		return nil
	} else {
		return fmt.Errorf("received unexpected event %T", event)
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

func (n *Node) becomeFollower(leader NodeID, term Term) error {
	assert.True(term >= n.term)

	if err := n.abortForwarded(); err != nil {
		return err
	}
	n.forwarded = treeset.NewWith(func(a, b RequestID) int {
		return bytes.Compare(a, b)
	})

	if leader != 0 {
		if term == n.term && n.leader != leader {
			return fmt.Errorf(
				"multiple leaders in same term %d, %d vs %d",
				n.term, n.leader, leader,
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
		n.leader = 0
		n.votedFor = 0
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
	n.progress = make(map[NodeID]*progress)
	for _, peer := range n.peers.Values() {
		n.progress[peer] = &progress{
			next: 1,
			last: 0,
		}
	}
	n.role = Leader
	if err := n.heartbeat(); err != nil {
		return err
	}

	// Propose an empty command when assuming leadership, to disambiguate
	// previous entries in the log. See section 8 of the raft paper.
	_, err := n.propose(nil)
	return err
}

func (n *Node) isLeader(id NodeID) bool {
	return n.leader != 0 && n.leader == id
}

func (n *Node) quorum() int {
	return n.peers.Size()/2 + 1
}

func (n *Node) campaign() error {
	n.term++
	n.votesReceived.Clear()
	n.votesReceived.Add(n.id) // Vote for ourselves.
	if err := n.log.SetTerm(n.term, n.id); err != nil {
		return err
	}
	lastIndex, lastTerm := n.log.LastIndex()
	for _, peer := range n.peers.Values() {
		if err := n.send(peer, &raftpb.SolicitVote{
			LastIndex: lastIndex,
			LastTerm:  lastTerm,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) heartbeat() error {
	commitIndex, commitTerm := n.log.CommitIndex()
	n.peers.Add()
	for _, peer := range n.peers.Values() {
		if err := n.send(peer, &raftpb.Heartbeat{
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

func (n *Node) maybeCommit() error {
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
		return nil
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
			return fmt.Errorf("missing entry at index %d", commitIndex)
		}
		return err
	}
	// We can only safely commit up to an entry from our own term,
	// see figure 8 in the raft paper.
	if entry.Term != n.term {
		return nil
	}

	if commitIndex > oldCommitIndex {
		if err := n.log.Commit(commitIndex); err != nil {
			return err
		}
		it, err := n.log.Scan(oldCommitIndex+1, commitIndex+1)
		if err != nil {
			return err
		}
		if err := itertools.Walk(it, func(entry *raftpb.Entry) error {
			n.stateCh <- &ApplyInstruction{Entry: entry}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func (n *Node) abortForwarded() error {
	for _, id := range n.forwarded.Values() {
		if err := n.sendToClient(&raftpb.ClientResponse{
			Id:    id,
			Error: requestAborted,
		}); err != nil {
			return err
		}
	}
	n.forwarded.Clear()
	return nil
}

func (n *Node) send(to NodeID, event raftpb.Eventer) error {
	n.msgCh <- &raftpb.Message{
		Term:  n.term,
		From:  n.id,
		To:    to,
		Event: event.GetEvent(),
	}
	return nil
}

func (n *Node) sendToClient(resp *raftpb.ClientResponse) error {
	n.msgCh <- &raftpb.Message{
		Term: n.term,
		From: n.id,
		To:   0,
		Event: &raftpb.Message_ClientResponse{
			ClientResponse: resp,
		},
	}
	return nil
}

func (n *Node) sendLog(peer NodeID) error {
	progress, ok := n.progress[peer]
	if !ok {
		return fmt.Errorf("unknown peer %d", peer)
	}

	it, err := n.log.Scan(progress.next, math.MaxUint64)
	if err != nil {
		return err
	}
	entries, err := itertools.ToSlice(it)
	if err != nil {
		return err
	}
	return n.send(peer, &raftpb.AppendEntries{
		Entries: entries,
	})
}
