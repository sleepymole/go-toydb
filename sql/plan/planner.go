package plan

import (
	"github.com/sleepymole/go-toydb/sql/ast"
	"github.com/sleepymole/go-toydb/sql/catalog"
)

type Planner struct {
	cat catalog.Catalog
}

func NewPlanner(cat catalog.Catalog) *Planner {
	return &Planner{cat: cat}
}

func (p *Planner) Build(stmt ast.Stmt) (Node, error) {
	node, err := p.buildStmt(stmt)
	if err != nil {
		return nil, err
	}
	return optimize(node, p.cat)
}

func (p *Planner) buildStmt(stmt ast.Stmt) (Node, error) {
	panic("implement me")
}

type Scope struct {
}
