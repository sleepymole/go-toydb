package parser

import (
	"github.com/sleepymole/go-toydb/sql/ast"
)

type Parser struct {
}

func (p *Parser) Parse(query string) (ast.Stmt, error) {
	panic("implement me")
}
