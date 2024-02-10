package storage

import (
	"github.com/samber/mo"
	"github.com/sleepymole/go-toydb/util/set"
)

var (
	mvccTxnActiveValue = []byte{}
	mvccTxnWriteValue  = []byte{}
)

func encodeMVCCNextVersionValue(version MVCCVersion) []byte {
	panic("implement me")
}

func decodeMVCCNextVersionValue(b []byte) (MVCCVersion, error) {
	panic("implement me")
}

func encodeMVCCTxnActiveSnapshotValue(active set.Set[MVCCVersion]) []byte {
	panic("implement me")
}

func decodeMVCCTxnActiveSnapshotValue(b []byte) (set.Set[MVCCVersion], error) {
	panic("implement me")
}

func encodeMVCCVersionedValue(value mo.Option[[]byte]) []byte {
	panic("implement me")
}

func decodeMVCCVersionedValue(b []byte) (mo.Option[[]byte], error) {
	panic("implement me")
}
