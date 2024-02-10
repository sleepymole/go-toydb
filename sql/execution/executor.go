package execution

import (
	"github.com/sleepymole/go-toydb/sql/plan"
	"github.com/sleepymole/go-toydb/sql/types"
)

type Executor interface {
	Execute() (*ResultSet, error)
}

func BuildExecutor(node plan.Node) (Executor, error) {
	panic("implement me")
}

type ResultSetType uint8

const (
	ResultSetTypeBegin ResultSetType = iota
	ResultSetTypeCommit
	ResultSetTypeRollback
	ResultSetTypeCreate
	ResultSetTypeDelete
	ResultSetTypeUpdate
	ResultSetTypeCreateTable
	ResultSetTypeDropTable
	ResultSetTypeQuery
	ResultSetTypeExplain
)

type ResultSet struct {
	Type     ResultSetType
	Version  uint64     // For BEGIN, COMMIT, and ROLLBACK
	ReadOnly bool       // For BEGIN
	Count    int64      // For CREATE, DELETE, INSERT, and UPDATE
	Table    string     // For CREATE TABLE, DROP TABLE
	Columns  []string   // For QUERY
	Rows     types.Rows // For QUERY
	Node     plan.Node  // For EXPLAIN
}

func MakeResultSetBegin(version uint64, readOnly bool) *ResultSet {
	return &ResultSet{Type: ResultSetTypeBegin, Version: version, ReadOnly: readOnly}
}

func MakeResultSetCommit(version uint64) *ResultSet {
	return &ResultSet{Type: ResultSetTypeCommit, Version: version}
}

func MakeResultSetRollback(version uint64) *ResultSet {
	return &ResultSet{Type: ResultSetTypeRollback, Version: version}
}

func MakeResultSetCreate(count int64) *ResultSet {
	return &ResultSet{Type: ResultSetTypeCreate, Count: count}
}

func MakeResultSetDelete(count int64) *ResultSet {
	return &ResultSet{Type: ResultSetTypeDelete, Count: count}
}

func MakeResultSetUpdate(count int64) *ResultSet {
	return &ResultSet{Type: ResultSetTypeUpdate, Count: count}
}

func MakeResultSetCreateTable(table string) *ResultSet {
	return &ResultSet{Type: ResultSetTypeCreateTable, Table: table}
}

func MakeResultSetDropTable(table string) *ResultSet {
	return &ResultSet{Type: ResultSetTypeDropTable, Table: table}
}

func MakeResultSetQuery(columns []string, rows types.Rows) *ResultSet {
	return &ResultSet{Type: ResultSetTypeQuery, Columns: columns, Rows: rows}
}

func MakeResultSetExplain(node plan.Node) *ResultSet {
	return &ResultSet{Type: ResultSetTypeExplain, Node: node}
}
