package main

import (
	"flag"
	"log"
	"net"
	"os"

	"github.com/sleepymole/go-toydb/raft"
	"github.com/sleepymole/go-toydb/raft/raftpb"
	"github.com/sleepymole/go-toydb/sql"
	"github.com/sleepymole/go-toydb/sql/sqlpb"
	"github.com/sleepymole/go-toydb/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"
)

type config struct {
	ID               uint64            `yaml:"id"`
	Peers            map[uint64]string `yaml:"peers"`
	ListenSQL        string            `yaml:"listen_sql"`
	ListenRaft       string            `yaml:"listen_raft"`
	LogLevel         string            `yaml:"log_level"`
	DataDir          string            `yaml:"data_dir"`
	CompactThreshold float64           `yaml:"compact_threshold"`
	Sync             bool              `yaml:"sync"`
	StorageRaft      string            `yaml:"storage_raft"`
	StorageSQL       string            `yaml:"storage_sql"`
}

func defaultConfig() config {
	return config{
		ID:               1,
		Peers:            nil,
		ListenSQL:        ":9605",
		ListenRaft:       ":9706",
		LogLevel:         "info",
		DataDir:          "data",
		CompactThreshold: 0.2,
		Sync:             true,
		StorageRaft:      "memory",
		StorageSQL:       "memory",
	}
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file")
	flag.Parse()

	cfg := defaultConfig()
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err != nil {
			log.Fatal(err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Fatal(err)
		}
	}

	var err error
	var raftLog *raft.Log
	switch cfg.StorageRaft {
	case "memory":
		raftLog, err = raft.NewLog(storage.NewMemory(), cfg.Sync)
		if err != nil {
			log.Fatal(err)
		}
	case "bitcask":
		fallthrough
	default:
		log.Fatalf("storage engine %s is not supported", cfg.StorageRaft)
	}

	var raftState *sql.RaftState
	switch cfg.StorageSQL {
	case "memory":
		raftState = sql.NewRaftState(storage.NewMemory())
	case "bitcask":
		fallthrough
	default:
		log.Fatalf("storage engine %s is not supported", cfg.StorageSQL)
	}

	peerConns := make(map[raft.NodeID]*grpc.ClientConn)
	for id, addr := range cfg.Peers {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatal(err)
		}
		peerConns[id] = conn
	}
	raftServer, err := raft.NewServer(cfg.ID, raftLog, raftState, peerConns)
	if err != nil {
		log.Fatal(err)
	}
	startRaftServer(raftServer, cfg.ListenRaft)

	sqlServer := sql.NewServer(raftServer)
	startSQLServer(sqlServer, cfg.ListenSQL)

	// TODO: handle signals
	select {}
}

func startRaftServer(raftServer *raft.Server, addr string) error {
	grpcServer := grpc.NewServer()
	raftpb.RegisterRaftServer(grpcServer, raftServer)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	go grpcServer.Serve(ln)
	return nil
}

func startSQLServer(sqlServer *sql.Server, addr string) {
	grpcServer := grpc.NewServer()
	sqlpb.RegisterSQLServer(grpcServer, sqlServer)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	go grpcServer.Serve(ln)
}
