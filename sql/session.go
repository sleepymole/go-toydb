package sql

import (
	"errors"

	"github.com/sleepymole/go-toydb/sql/ast"
	"github.com/sleepymole/go-toydb/sql/execution"
	"github.com/sleepymole/go-toydb/sql/parser"
	"github.com/sleepymole/go-toydb/sql/plan"
)

type Session struct {
	engine Engine
	txn    Txn
}

func NewSession(engine Engine) *Session {
	return &Session{engine: engine}
}

func (s *Session) Execute(query string) (*execution.ResultSet, error) {
	var p parser.Parser
	stmt, err := p.Parse(query)
	if err != nil {
		return nil, err
	}
	switch x := stmt.(type) {
	case *ast.BeginStmt:
		if s.txn != nil {
			return nil, errors.New("already in a transaction")
		}
		if x.ReadOnly {
			if x.AsOf > 0 {
				txn, err := s.engine.BeginAsOf(x.AsOf)
				if err != nil {
					return nil, err
				}
				s.txn = txn
				result := execution.MakeResultSetBegin(txn.Version(), true)
				return result, nil
			}
			txn, err := s.engine.BeginReadOnly()
			if err != nil {
				return nil, err
			}
			s.txn = txn
			result := execution.MakeResultSetBegin(txn.Version(), true)
			return result, nil
		}
		if x.AsOf > 0 {
			return nil, errors.New("can't start read-write transaction in a given version")
		}
		txn, err := s.engine.Begin()
		if err != nil {
			return nil, err
		}
		s.txn = txn
		result := execution.MakeResultSetBegin(txn.Version(), false)
		return result, nil
	case *ast.CommitStmt:
		if s.txn == nil {
			return nil, errors.New("not in a transaction")
		}
		version := s.txn.Version()
		if err := s.txn.Commit(); err != nil {
			return nil, err
		}
		s.txn = nil
		result := execution.MakeResultSetCommit(version)
		return result, nil
	case *ast.RollbackStmt:
		if s.txn == nil {
			return nil, errors.New("not in a transaction")
		}
		version := s.txn.Version()
		if err := s.txn.Rollback(); err != nil {
			return nil, err
		}
		s.txn = nil
		result := execution.MakeResultSetRollback(version)
		return result, nil
	case *ast.ExplainStmt:
		var result *execution.ResultSet
		if err := s.runInTxn(func(txn Txn) error {
			planner := plan.NewPlanner(NewCatalog(txn))
			node, err := planner.Build(x.Stmt)
			if err != nil {
				return err
			}
			result = execution.MakeResultSetExplain(node)
			return nil
		}, true); err != nil {
			return nil, err
		}
		return result, nil
	default:
		var result *execution.ResultSet
		_, preferReadOnly := stmt.(*ast.SelectStmt)
		if err := s.runInTxn(func(txn Txn) error {
			planner := plan.NewPlanner(NewCatalog(txn))
			node, err := planner.Build(stmt)
			if err != nil {
				return err
			}
			executor, err := execution.BuildExecutor(node)
			if err != nil {
				return err
			}
			result, err = executor.Execute()
			return err
		}, preferReadOnly); err != nil {
			return nil, err
		}
		return result, nil
	}
}

func (s *Session) runInTxn(f func(Txn) error, preferReadOnly bool) error {
	txn := s.txn
	if txn == nil {
		var err error
		if preferReadOnly {
			txn, err = s.engine.BeginReadOnly()
		} else {
			txn, err = s.engine.Begin()
		}
		if err != nil {
			return err
		}
		defer txn.Rollback()
	}
	return f(s.txn)
}
