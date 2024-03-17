package execution

import (
	"github.com/sleepymole/go-toydb/sql/ast"
	"github.com/sleepymole/go-toydb/sql/plan"
)

type Executor interface {
	Execute() (*ResultSet, error)
}

func BuildExecutor(node plan.Node) (Executor, error) {
	panic("implement me")
}

type ResultSetType uint8

const (
	ResultSetBegin ResultSetType = iota
	ResultSetCommit
	ResultSetRollback
	ResultSetCreate
	ResultSetDelete
	ResultSetUpdate
	ResultSetCreateTable
	ResultSetDropTable
	ResultSetQuery
	ResultSetExplain
)

type ResultSet struct {
	Type     ResultSetType
	Version  uint64    // For BEGIN, COMMIT, and ROLLBACK
	ReadOnly bool      // For BEGIN
	Count    int64     // For CREATE, DELETE, INSERT, and UPDATE
	Table    string    // For CREATE TABLE, DROP TABLE
	Columns  []string  // For QUERY
	Rows     ast.Rows  // For QUERY
	Node     plan.Node // For EXPLAIN
}

func MakeBeginResultSet(version uint64, readOnly bool) *ResultSet {
	return &ResultSet{Type: ResultSetBegin, Version: version, ReadOnly: readOnly}
}

func MakeCommitResultSet(version uint64) *ResultSet {
	return &ResultSet{Type: ResultSetCommit, Version: version}
}

func MakeRollbackResultSet(version uint64) *ResultSet {
	return &ResultSet{Type: ResultSetRollback, Version: version}
}

func MakeCreateResultSet(count int64) *ResultSet {
	return &ResultSet{Type: ResultSetCreate, Count: count}
}

func MakeDeleteResultSet(count int64) *ResultSet {
	return &ResultSet{Type: ResultSetDelete, Count: count}
}

func MakeUpdateResultSet(count int64) *ResultSet {
	return &ResultSet{Type: ResultSetUpdate, Count: count}
}

func MakeCreateTableResultSet(table string) *ResultSet {
	return &ResultSet{Type: ResultSetCreateTable, Table: table}
}

func MakeDropTableResultSet(table string) *ResultSet {
	return &ResultSet{Type: ResultSetDropTable, Table: table}
}

func MakeQueryResultSet(columns []string, rows ast.Rows) *ResultSet {
	return &ResultSet{Type: ResultSetQuery, Columns: columns, Rows: rows}
}

func MakeExplainResultSet(node plan.Node) *ResultSet {
	return &ResultSet{Type: ResultSetExplain, Node: node}
}
