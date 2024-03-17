package sql

import (
	"errors"

	"github.com/emirpasic/gods/v2/sets"
	"github.com/emirpasic/gods/v2/sets/treeset"
	"github.com/sleepymole/go-toydb/sql/ast"
	"github.com/sleepymole/go-toydb/sql/catalog"
	"github.com/sleepymole/go-toydb/sql/values"
	"github.com/sleepymole/go-toydb/util/itertools"
)

var ErrNotFound = errors.New("sql: not found")

// ID is a unique identifier for a row in a table.
type ID = *values.Value

// IDSet is a set of row identifiers.
type IDSet = sets.Set[ID]

// NewIDSet returns a new IDSet.
func NewIDSet() IDSet {
	cmp := func(a, b ID) int {
		result, _ := a.TryCompare(b)
		return result
	}
	return treeset.NewWith(cmp)
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
	Create(table string, row ast.Row) error
	Delete(table, id ID) error
	Read(table string, id ID) (ast.Row, error)
	ReadIndex(table, column string, value *values.Value) (IDSet, error)
	Update(table string, id ID, row ast.Row) error
	Scan(table string, filter ast.Expr) (ast.Rows, error)
	IndexScan(table, column string) (IndexScan, error)
}

// IndexEntry holds the value of an indexed column and the set
// of row identifiers that have that value.
type IndexEntry struct {
	Value *values.Value
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
