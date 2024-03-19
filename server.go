package toydb

import (
	"context"

	"github.com/sleepymole/go-toydb/api/dbpb"
)

type Server struct {
	dbpb.UnimplementedDBServer
}

func (s *Server) Serve() error {
	panic("implement me")
}

func (s *Server) Execute(ctx context.Context, req *dbpb.ExecuteRequest) (*dbpb.ExecuteResponse, error) {
	panic("implement me")
}

func (s *Server) Status(ctx context.Context, req *dbpb.StatusRequest) (*dbpb.StatusResponse, error) {
	panic("implement me")
}
