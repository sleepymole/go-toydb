package catalog

import "github.com/sleepymole/go-toydb/util/itertools"

type Catalog interface {
	CreateTable(table *Table) error
	DeleteTable(table string) error
	ReadTable(table string) (*Table, error)
	ScanTables() (Tables, error)
	TableReferences(table string, withSelf bool) ([]*TableReference, error)
}

type Tables itertools.Iterator[*Table]

type TableReference struct {
	Table   string
	Columns []string
}
