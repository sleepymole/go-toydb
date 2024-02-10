package sql

import "github.com/sleepymole/go-toydb/storage"

// KVEngine is a SQL engine based on an underlying MVCC key/value store.
type KVEngine struct {
	Engine
	kv *storage.MVCC
}

func NewKVEngine(engine storage.Engine) *KVEngine {
	return &KVEngine{kv: storage.NewMVCC(engine)}
}

type KVTxn struct {
	Txn
}
