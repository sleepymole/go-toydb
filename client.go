package toydb

import (
	"context"

	"github.com/sleepymole/go-toydb/api/dbpb"
)

type Client struct {
}

func (c *Client) Execute(ctx context.Context, query string) (*dbpb.ExecuteResponse, error) {
	panic("implement me")
}

func (c *Client) Mutate(ctx context.Context, query string) (*dbpb.MutateResult, error) {
	panic("implement me")
}

func (c *Client) Query(ctx context.Context, query string) (*dbpb.QueryResult, error) {
	panic("implement me")
}

func (c *Client) Status(ctx context.Context) (*dbpb.StatusResponse, error) {
	panic("implement me")
}
