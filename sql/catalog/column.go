package catalog

import (
	"github.com/sleepymole/go-toydb/sql/values"
)

type Column struct {
	Name       string
	Type       values.Type
	PrimaryKey bool
	Nullable   bool
	Default    *values.Value
	Unique     bool
	References string
	Index      bool
}

func (c *Column) String() string {
	panic("implement me")
}
