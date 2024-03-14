package raft

import (
	"encoding/gob"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/emirpasic/gods/v2/sets/hashset"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/sleepymole/go-toydb/raft/raftpb"
	"github.com/sleepymole/go-toydb/util/assert"
)

const tickInterval = time.Millisecond * 100

func genertateRequestID() RequestID {
	uuid := uuid.New()
	return RequestID(uuid[:])
}

type Server struct {
	node      *Node
	peers     map[NodeID]string
	peerConns map[NodeID]net.Conn
	msgInCh   chan *raftpb.Message
	msgOutCh  chan *raftpb.Message
	clientCh  chan lo.Tuple2[*raftpb.ClientRequest, chan *raftpb.ClientResponse]
}

func NewServer(id NodeID, peers map[NodeID]string, log *Log, state State) (*Server, error) {
	msgOutCh := make(chan *raftpb.Message, 1024)
	peerSet := hashset.New[NodeID]()
	for id := range peers {
		peerSet.Add(id)
	}
	node, err := NewNode(id, peerSet, log, state, msgOutCh)
	if err != nil {
		return nil, err
	}
	msgInCh := make(chan *raftpb.Message, 1024)
	clientCh := make(chan lo.Tuple2[*raftpb.ClientRequest, chan *raftpb.ClientResponse], 1024)
	return &Server{
		node:     node,
		peers:    peers,
		msgInCh:  msgInCh,
		msgOutCh: msgOutCh,
		clientCh: clientCh,
	}, nil
}

func (s *Server) Serve(l net.Listener) error {
	go s.eventLoop()

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.receiveLoop(conn)
	}
}

func (s *Server) Mutate(command []byte) ([]byte, error) {
	req := &raftpb.ClientRequest{
		Id:      genertateRequestID(),
		Type:    raftpb.ClientRequest_MUTATE,
		Command: command,
	}
	respCh := make(chan *raftpb.ClientResponse, 1)
	s.clientCh <- lo.T2(req, respCh)
	resp := <-respCh
	if resp.Error != "" {
		return resp.Result, errors.New(resp.Error)
	}
	return resp.Result, nil
}

func (s *Server) Query(command []byte) ([]byte, error) {
	req := &raftpb.ClientRequest{
		Id:      genertateRequestID(),
		Type:    raftpb.ClientRequest_QUERY,
		Command: command,
	}
	respCh := make(chan *raftpb.ClientResponse, 1)
	s.clientCh <- lo.T2(req, respCh)
	resp := <-respCh
	if resp.Error != "" {
		return resp.Result, errors.New(resp.Error)
	}
	return resp.Result, nil
}

func (s *Server) Status() (*raftpb.Status, error) {
	req := &raftpb.ClientRequest{
		Id:   genertateRequestID(),
		Type: raftpb.ClientRequest_STATUS,
	}
	respCh := make(chan *raftpb.ClientResponse, 1)
	s.clientCh <- lo.T2(req, respCh)
	resp := <-respCh
	if resp.Error != "" {
		return resp.Status, errors.New(resp.Error)
	}
	return resp.Status, nil
}

func (s *Server) eventLoop() error {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	respChs := make(map[string]chan *raftpb.ClientResponse)

	for {
		select {
		case <-ticker.C:
			if err := s.node.Tick(); err != nil {
				return err
			}
		case msg, ok := <-s.msgInCh:
			if !ok {
				return nil
			}
			if err := s.node.Step(msg); err != nil {
				return err
			}
		case msg, ok := <-s.msgOutCh:
			if !ok {
				return nil
			}
			if msg.To == 0 {
				if resp := msg.GetClientResponse(); resp != nil {
					respCh, ok := respChs[string(resp.Id)]
					if ok {
						respCh <- resp
						delete(respChs, string(resp.Id))
					}
				}
				continue
			}
			assert.True(msg.To != 0, "non-client message must have a destination")
			s.send(msg.To, msg)
		case tuple, ok := <-s.clientCh:
			if !ok {
				return nil
			}
			req, respCh := tuple.Unpack()
			msg := &raftpb.Message{
				Event: &raftpb.Message_ClientRequest{
					ClientRequest: req,
				},
			}
			if err := s.node.Step(msg); err != nil {
				return err
			}
			respChs[string(req.Id)] = respCh
		}
	}
}

func (s *Server) send(to NodeID, msg *raftpb.Message) {
	conn, err := s.getConn(to)
	if err != nil {
		slog.Error("failed to get connection to peer %d: %v", to, err)
		return
	}

	// TODO: Use gRPC instead of sending gob-encoded messages over TCP.
	encoder := gob.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		_ = conn.Close()
		slog.Error("failed to encode message: %v", err)
		return
	}
	s.returnConn(to, conn)
}

func (s *Server) getConn(id NodeID) (net.Conn, error) {
	conn, ok := s.peerConns[id]
	if ok {
		return conn, nil
	}
	addr, ok := s.peers[id]
	if !ok {
		return nil, fmt.Errorf("address of node %d not found", id)
	}
	return net.Dial("tcp", addr)
}

func (s *Server) returnConn(id NodeID, conn net.Conn) {
	s.peerConns[id] = conn
}

func (s *Server) receiveLoop(conn net.Conn) {
	defer conn.Close()

	// TODO: Use gRPC instead of sending gob-encoded messages over TCP.
	decoder := gob.NewDecoder(conn)
	for {
		var msg raftpb.Message
		if err := decoder.Decode(&msg); err != nil {
			slog.Error("failed to decode message: %v", err)
			return
		}
		s.msgInCh <- &msg
	}
}
