package storagepb

import "slices"

func (m *MVCCTxnState) IsVisible(version uint64) bool {
	if slices.Contains(m.ActiveTxns, version) {
		return false
	}
	if m.ReadOnly {
		return version < m.Version
	}
	return version <= m.Version
}
