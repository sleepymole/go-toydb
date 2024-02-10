package catalog

import "github.com/sleepymole/go-toydb/sql/types"

type Column struct {
	Name       string
	DataType   types.DataType
	PrimaryKey bool
	Nullable   bool
	Default    *types.Value
	Unique     bool
	References string
	Index      bool
}

func (c *Column) String() string {
	panic("implement me")
}
