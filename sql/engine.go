package sql

import (
	"errors"

	"github.com/sleepymole/go-toydb/sql/catalog"
	"github.com/sleepymole/go-toydb/sql/types"
	"github.com/sleepymole/go-toydb/util/itertools"
	"github.com/sleepymole/go-toydb/util/set"
)

var ErrNotFound = errors.New("sql: not found")

// ID is a unique identifier for a row in a table.
type ID = *types.Value

// IDSet is a set of row identifiers.
type IDSet = *set.HashSet[ID]

// NewIDSet returns a new IDSet.
func NewIDSet() IDSet {
	return set.NewHashSet[ID]()
}

type Engine interface {
	Begin() (Txn, error)
	BeginReadOnly() (Txn, error)
	BeginAsOf(version uint64) (Txn, error)
}

type Txn interface {
	Version() uint64
	ReadOnly() bool
	Commit() error
	Rollback() error
	Create(table string, row types.Row) error
	Delete(table, id ID) error
	Read(table string, id ID) (types.Row, error)
	ReadIndex(table, column string, value *types.Value) (IDSet, error)
	Update(table string, id ID, row types.Row) error
	Scan(table string, filter types.Expr) (types.Rows, error)
	IndexScan(table, column string) (IndexScan, error)
}

// IndexEntry holds the value of an indexed column and the set
// of row identifiers that have that value.
type IndexEntry struct {
	Value *types.Value
	IDSet IDSet
}

// IndexScan is an iterator over the index entries of a table.
type IndexScan itertools.Iterator[*IndexEntry]

type Catalog struct {
	txn Txn
}

var _ catalog.Catalog = &Catalog{}

func NewCatalog(txn Txn) *Catalog {
	return &Catalog{txn: txn}
}

func (c *Catalog) CreateTable(table *catalog.Table) error {
	panic("implement me")
}

func (c *Catalog) DeleteTable(table string) error {
	panic("implement me")
}

func (c *Catalog) ReadTable(table string) (*catalog.Table, error) {
	panic("implement me")
}

func (c *Catalog) ScanTables() (catalog.Tables, error) {
	panic("implement me")
}

func (c *Catalog) TableReferences(table string, withSelf bool) ([]*catalog.TableReference, error) {
	panic("implement me")
}
