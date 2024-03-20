package toydb

import (
	"context"

	"github.com/sleepymole/go-toydb/toypb"
)

type Server struct {
	toypb.UnimplementedToyDBServer
}

func (s *Server) Serve() error {
	panic("implement me")
}

func (s *Server) Execute(ctx context.Context, req *toypb.ExecuteRequest) (*toypb.ExecuteResponse, error) {
	panic("implement me")
}

func (s *Server) Status(ctx context.Context, req *toypb.StatusRequest) (*toypb.StatusResponse, error) {
	panic("implement me")
}
