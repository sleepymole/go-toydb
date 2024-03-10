package raft

import (
	"encoding/gob"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/emirpasic/gods/v2/sets/hashset"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/sleepymole/go-toydb/util/assert"
)

const tickInterval = time.Millisecond * 100

type Server struct {
	node      *Node
	addr      string
	peers     map[NodeID]string
	peerConns map[NodeID]net.Conn
	msgInCh   chan *Message
	msgOutCh  chan *Message
	clientCh  chan lo.Tuple2[*ClientRequest, chan *ClientResponse]
}

func NewServer(
	id NodeID,
	addr string,
	peers map[NodeID]string,
	log *Log,
	state State,
) (*Server, error) {
	msgOutCh := make(chan *Message, 1024)
	peerSet := hashset.New[NodeID]()
	for id := range peers {
		peerSet.Add(id)
	}
	node, err := NewNode(id, peerSet, log, state, msgOutCh)
	if err != nil {
		return nil, err
	}
	msgInCh := make(chan *Message, 1024)
	clientCh := make(chan lo.Tuple2[*ClientRequest, chan *ClientResponse], 1024)
	return &Server{
		node:     node,
		addr:     addr,
		peers:    peers,
		msgInCh:  msgInCh,
		msgOutCh: msgOutCh,
		clientCh: clientCh,
	}, nil
}

func (s *Server) Serve() error {
	go s.eventLoop()

	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		go s.receiveLoop(conn)
	}
}

func (s *Server) Mutate(command []byte) ([]byte, error) {
	req := &ClientRequest{
		ID:      makeRequestID(),
		Type:    RequestMutate,
		Command: command,
	}
	respCh := make(chan *ClientResponse, 1)
	s.clientCh <- lo.T2(req, respCh)
	resp := <-respCh
	return resp.Result, resp.Err
}

func (s *Server) Query(command []byte) ([]byte, error) {
	req := &ClientRequest{
		ID:      makeRequestID(),
		Type:    RequestQuery,
		Command: command,
	}
	respCh := make(chan *ClientResponse, 1)
	s.clientCh <- lo.T2(req, respCh)
	resp := <-respCh
	return resp.Result, resp.Err
}

func (s *Server) Status() (*Status, error) {
	req := &ClientRequest{
		ID:   makeRequestID(),
		Type: RequestStatus,
	}
	respCh := make(chan *ClientResponse, 1)
	s.clientCh <- lo.T2(req, respCh)
	resp := <-respCh
	return resp.Status, resp.Err
}

func makeRequestID() RequestID {
	uuid := uuid.New()
	return RequestID(uuid[:])
}

func (s *Server) eventLoop() error {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	respChs := make(map[string]chan *ClientResponse)

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
			if msg.Client {
				resp, ok := msg.Event.(*ClientResponse)
				if ok {
					respCh, ok := respChs[string(resp.ID)]
					if ok {
						respCh <- resp
						delete(respChs, string(resp.ID))
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
			msg := &Message{
				Client: true,
				Event:  req,
			}
			if err := s.node.Step(msg); err != nil {
				return err
			}
			respChs[string(req.ID)] = respCh
		}
	}
}

func (s *Server) send(to NodeID, msg *Message) {
	conn, err := s.getConn(to)
	if err != nil {
		slog.Error("failed to get connection to peer %d: %v", to, err)
		return
	}

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

	decoder := gob.NewDecoder(conn)
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			slog.Error("failed to decode message: %v", err)
			return
		}
		s.msgInCh <- &msg
	}
}
