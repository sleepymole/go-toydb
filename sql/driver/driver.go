package driver

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/url"

	"github.com/sleepymole/go-toydb/sql/sqlpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const driverName = "toydb"

func init() {
	sql.Register(driverName, &Driver{})
}

type Driver struct{}

var _ driver.DriverContext = Driver{}

func (Driver) Open(dsn string) (driver.Conn, error) {
	return nil, errors.New("driver: Open is deprecated and should use OpenConnector instead")
}

func (Driver) OpenConnector(dsn string) (driver.Connector, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	return &connector{addr: u.Host}, nil
}

type connector struct {
	addr string
}

var _ driver.Connector = &connector{}

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	grpcConn, err := grpc.DialContext(ctx, c.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &conn{grpcConn: grpcConn}, nil
}

func (c *connector) Driver() driver.Driver {
	return Driver{}
}

type conn struct {
	grpcConn *grpc.ClientConn
	cli      sqlpb.SQL_ExecuteClient
	closed   bool
}

var _ driver.Conn = &conn{}
var _ driver.Pinger = &conn{}
var _ driver.SessionResetter = &conn{}
var _ driver.ConnPrepareContext = &conn{}
var _ driver.ExecerContext = &conn{}
var _ driver.QueryerContext = &conn{}
var _ driver.ConnBeginTx = &conn{}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("driver: Prepare is deprecated and should use PrepareContext instead")
}

func (c *conn) Begin() (driver.Tx, error) {
	return nil, errors.New("driver: Begin is deprecated and should use BeginTx instead")
}

func (c *conn) Close() error {
	c.closed = true
	err := c.cli.CloseSend()
	return errors.Join(err, c.grpcConn.Close())
}

func (c *conn) Ping(ctx context.Context) error {
	_, err := c.grabExecClient()
	return err
}

func (c *conn) grabExecClient() (sqlpb.SQL_ExecuteClient, error) {
	if c.cli != nil {
		return c.cli, nil
	}
	cli, err := sqlpb.NewSQLClient(c.grpcConn).Execute(context.Background())
	if err != nil {
		return nil, err
	}
	c.cli = cli
	return c.cli, nil
}

func (c *conn) maybeCloseExecClient() error {
	if c.cli != nil {
		err := c.cli.CloseSend()
		c.cli = nil
		return err
	}
	return nil
}

func (c *conn) ResetSession(ctx context.Context) error {
	if c.closed {
		return driver.ErrBadConn
	}
	return c.maybeCloseExecClient()
}

func (c *conn) IsValid() bool {
	return !c.closed
}

func (c *conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	panic("implement me")
}

func (c *conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	panic("implement me")
}

func (c *conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	panic("implement me")
}

func (c *conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	panic("implement me")
}

type stmt struct {
	driver.Stmt
	driver.StmtExecContext
	driver.StmtQueryContext
}

var _ driver.Stmt = &stmt{}
var _ driver.StmtExecContext = &stmt{}
var _ driver.StmtQueryContext = &stmt{}

type tx struct {
	driver.Tx
}

var _ driver.Tx = &tx{}
