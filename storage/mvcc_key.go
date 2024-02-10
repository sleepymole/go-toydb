package storage

const (
	mvccNextVersionKeyTag    = byte(0x00)
	mvccTxnActiveKeyTag      = byte(0x01)
	mvccTxnActiveSnapshotKey = byte(0x02)
	mvccTxnWriteKeyTag       = byte(0x03)
	mvccVersionedKeyTag      = byte(0x04)
	mvccUnversionedKeyTag    = byte(0x05)
)

var (
	mvccNextVersionKey     = []byte{mvccNextVersionKeyTag}
	mvccTxnActiveKeyPrefix = []byte{mvccTxnActiveKeyTag}
	mvccVersionedKeyPrefix = []byte{mvccVersionedKeyTag}
)

func encodeMVCCTxnActiveKey(version MVCCVersion) []byte {
	panic("implement me")
}

func decodeMVCCTxnActiveKey(b []byte) (MVCCVersion, error) {
	panic("implement me")
}

func encodeMVCCTxnActiveSnapshotKey(version MVCCVersion) []byte {
	panic("implement me")
}

func decodeMVCCTxnActiveSnapshotKey(b []byte) (MVCCVersion, error) {
	panic("implement me")
}

func encodeMVCCTxnWriteKey(version MVCCVersion, key []byte) []byte {
	panic("implement me")
}

func encodeMVCCTxnWriteKeyPrefix(version MVCCVersion) []byte {
	panic("implement me")
}

func decodeMVCCTxnWriteKey(b []byte) (_ MVCCVersion, key []byte, _ error) {
	panic("implement me")
}

func encodeMVCCVersionedKey(key []byte, version MVCCVersion) []byte {
	panic("implement me")
}

func decodeMVCCVersionedKey(b []byte) (key []byte, _ MVCCVersion, _ error) {
	panic("implement me")
}

func encodeMVCCUnversionedKey(key []byte) []byte {
	panic("implement me")
}

func decodeMVCCUnversionedKey(b []byte) (key []byte, _ error) {
	panic("implement me")
}
