package sql

import (
	"context"

	"github.com/sleepymole/go-toydb/raft"
	"github.com/sleepymole/go-toydb/sql/sqlpb"
)

type Server struct {
	raftServer *raft.Server
}

var _ sqlpb.SQLServer = &Server{}

func NewServer(raftServer *raft.Server) *Server {
	return &Server{raftServer: raftServer}
}

func (s *Server) Serve(ctx context.Context) error {
	panic("implement me")
}

func (s *Server) Execute(stream sqlpb.SQL_ExecuteServer) error {
	panic("implement me")
}

func (s *Server) Status(ctx context.Context, req *sqlpb.StatusRequest) (*sqlpb.StatusResponse, error) {
	panic("implement me")
}
