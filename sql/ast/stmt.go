package ast

type Stmt interface{}

var _ Stmt = &BeginStmt{}
var _ Stmt = &CommitStmt{}
var _ Stmt = &RollbackStmt{}
var _ Stmt = &ExplainStmt{}
var _ Stmt = &CreateTableStmt{}
var _ Stmt = &DropTableStmt{}
var _ Stmt = &DeleteStmt{}
var _ Stmt = &InsertStmt{}
var _ Stmt = &UpdateExpr{}
var _ Stmt = &SelectStmt{}

type BeginStmt struct {
	ReadOnly bool
	AsOf     uint64
}

type CommitStmt struct{}

type RollbackStmt struct{}

type ExplainStmt struct {
	Stmt Stmt
}

type CreateTableStmt struct {
	Name    string
	Columns []*Column
}

type DropTableStmt struct {
	Name string
}

type DeleteStmt struct {
	Table string
	Where Expr
}

type InsertStmt struct {
	Table   string
	Columns []*Column
	Values  [][]Expr
}

type UpdateStmt struct {
	Table string
	Exprs []*UpdateExpr
	Where Expr
}

type UpdateExpr struct {
	Name string
	Expr Expr
}

type SelectStmt struct {
	Exprs   []*SelectExpr
	From    []TableExpr
	Where   Expr
	GroupBy []Expr
	Having  Expr
	OrderBy []*OrderBy
	Offset  Expr
	Limit   Expr
}

type SelectExpr struct {
	Expr string
	As   string
}

type TableExpr interface {
	tableExpr()
}

type AliasedTableExpr struct {
	Name  string
	Alias string
}

type JoinTableExpr struct {
	Left  TableExpr
	Right TableExpr
	Type  JoinType
	Cond  Expr
}

type JoinType uint8

const (
	CrossJoin JoinType = iota
	InnerJoin
	LeftJoin
	RightJoin
)

type OrderBy struct {
	Expr      Expr
	Direction Direction
}

type Direction uint8

const (
	DirectionAscending Direction = iota
	DirectionDescending
)

type Column struct {
	Name       string
	DataType   string
	PrimaryKey bool
	Nullable   bool
	Default    Expr
	Unique     bool
	Index      bool
	References []string
}
