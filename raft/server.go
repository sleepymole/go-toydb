package raft

import (
	"net"
)

type RequestID uint64

type Server struct {
	node  *Node
	peers map[NodeID]string
	msgCh chan *Message
}

func (s *Server) Serve(l *net.Listener) error {
	panic("implement me")
}
