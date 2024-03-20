package toydb

import (
	"context"

	"github.com/sleepymole/go-toydb/toypb"
)

type Client struct {
}

func (c *Client) Execute(ctx context.Context, query string) (*toypb.ExecuteResponse, error) {
	panic("implement me")
}

func (c *Client) Mutate(ctx context.Context, query string) (*toypb.MutateResult, error) {
	panic("implement me")
}

func (c *Client) Query(ctx context.Context, query string) (*toypb.QueryResult, error) {
	panic("implement me")
}

func (c *Client) Status(ctx context.Context) (*toypb.StatusResponse, error) {
	panic("implement me")
}
