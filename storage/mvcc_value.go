package storage

import (
	"github.com/emirpasic/gods/v2/sets"
	"github.com/samber/mo"
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

func encodeMVCCTxnActiveSnapshotValue(active sets.Set[MVCCVersion]) []byte {
	panic("implement me")
}

func decodeMVCCTxnActiveSnapshotValue(b []byte) (sets.Set[MVCCVersion], error) {
	panic("implement me")
}

func encodeMVCCVersionedValue(value mo.Option[[]byte]) []byte {
	panic("implement me")
}

func decodeMVCCVersionedValue(b []byte) (mo.Option[[]byte], error) {
	panic("implement me")
}
