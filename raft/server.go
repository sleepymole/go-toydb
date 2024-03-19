package raft

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/emirpasic/gods/v2/sets/hashset"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/sleepymole/go-toydb/api/raftpb"
	"github.com/sleepymole/go-toydb/util/assert"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const tickInterval = time.Millisecond * 100

func genertateRequestID() RequestID {
	uuid := uuid.New()
	return RequestID(uuid[:])
}

type Server struct {
	raftpb.UnimplementedRaftServer

	node       *Node
	msgInCh    chan *raftpb.Message
	msgOutCh   chan *raftpb.Message
	clientCh   chan lo.Tuple2[*raftpb.ClientRequest, chan *raftpb.ClientResponse]
	peerConns  map[NodeID]*grpc.ClientConn
	msgClients map[NodeID]raftpb.Raft_SendMessagesClient
}

func NewServer(id NodeID, peers map[NodeID]string, log *Log, state State) (*Server, error) {
	msgOutCh := make(chan *raftpb.Message, 1024)
	peerIDs := hashset.New[NodeID]()
	peerConns := make(map[NodeID]*grpc.ClientConn)

	var allIsWell bool
	defer func() {
		if !allIsWell {
			for _, conn := range peerConns {
				_ = conn.Close()
			}
		}
	}()

	for id := range peers {
		peerIDs.Add(id)
		conn, err := grpc.Dial(peers[id], grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return nil, fmt.Errorf("failed to dial peer %d: %v", id, err)
		}
		peerConns[id] = conn
	}
	node, err := NewNode(id, peerIDs, log, state, msgOutCh)
	if err != nil {
		return nil, err
	}
	msgInCh := make(chan *raftpb.Message, 1024)
	clientCh := make(chan lo.Tuple2[*raftpb.ClientRequest, chan *raftpb.ClientResponse], 1024)

	allIsWell = true
	return &Server{
		node:       node,
		msgInCh:    msgInCh,
		msgOutCh:   msgOutCh,
		clientCh:   clientCh,
		peerConns:  peerConns,
		msgClients: make(map[NodeID]raftpb.Raft_SendMessagesClient),
	}, nil
}

func (s *Server) Serve(l net.Listener) error {
	grpcServer := grpc.NewServer()
	raftpb.RegisterRaftServer(grpcServer, s)
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		<-ctx.Done()
		grpcServer.Stop()
		return nil
	})
	g.Go(func() error {
		return s.eventLoop(ctx)
	})
	return g.Wait()
}

func (s *Server) Messages(stream raftpb.Raft_SendMessagesServer) error {
	ctx := stream.Context()
	for {
		msg, err := stream.Recv()
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case s.msgInCh <- msg:
		}
	}
}

func (s *Server) Mutate(ctx context.Context, command []byte) ([]byte, error) {
	req := &raftpb.ClientRequest{
		Id:      genertateRequestID(),
		Type:    raftpb.ClientRequest_MUTATE,
		Command: command,
	}
	resp, err := s.clientRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Result, nil
}

func (s *Server) Query(ctx context.Context, command []byte) ([]byte, error) {
	req := &raftpb.ClientRequest{
		Id:      genertateRequestID(),
		Type:    raftpb.ClientRequest_QUERY,
		Command: command,
	}
	resp, err := s.clientRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Result, nil
}

func (s *Server) Status(ctx context.Context) (*raftpb.Status, error) {
	req := &raftpb.ClientRequest{
		Id:   genertateRequestID(),
		Type: raftpb.ClientRequest_STATUS,
	}
	resp, err := s.clientRequest(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Status, nil
}

func (s *Server) clientRequest(ctx context.Context, req *raftpb.ClientRequest) (*raftpb.ClientResponse, error) {
	respCh := make(chan *raftpb.ClientResponse, 1)
	defer close(respCh)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case s.clientCh <- lo.T2(req, respCh):
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case resp := <-respCh:
		if resp.Error != "" {
			return nil, errors.New(resp.Error)
		}
		return resp, nil
	}
}

func (s *Server) eventLoop(ctx context.Context) error {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	respChs := make(map[string]chan *raftpb.ClientResponse)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
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
			if err := s.sendMsg(msg.To, msg); err != nil {
				slog.Error("failed to send message", slog.Uint64("to", uint64(msg.To)), slog.Any("error", err))
			}
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

func (s *Server) sendMsg(to NodeID, msg *raftpb.Message) error {
	client, err := s.getMsgClient(to)
	if err != nil {
		return err
	}
	if err := client.Send(msg); err != nil {
		_ = client.CloseSend()
		delete(s.msgClients, to)
		return err
	}
	return nil
}

func (s *Server) getMsgClient(peer NodeID) (raftpb.Raft_SendMessagesClient, error) {
	client, ok := s.msgClients[peer]
	if ok {
		return client, nil
	}
	conn, ok := s.peerConns[peer]
	if !ok {
		return nil, fmt.Errorf("no connection to peer %d", peer)
	}
	client, err := raftpb.NewRaftClient(conn).SendMessages(context.TODO())
	if err != nil {
		return nil, err
	}
	s.msgClients[peer] = client
	return client, nil
}
