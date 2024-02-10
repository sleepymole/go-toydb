package ast

import (
	"fmt"
)

type Expr interface {
	fmt.Stringer
	expr()
}

type FieldExpr struct {
}

type ColumnExpr struct {
}

type LiteralExpr interface {
	Expr
	literalExpr()
}

type LiteralNull struct{}

type LiteralBoolean bool

type LiteralInteger int64

type LiteralFloat float64

type LiteralString string

type FunctionExpr struct {
}

type OperationExpr interface {
	Expr
	operationExpr()
}

type AndExpr struct {
}

type NotExpr struct {
}

type OrExpr struct {
}

type EqualExpr struct {
}

type GreaterThanExpr struct {
}

type IsNullExpr struct {
}

type LessThanExpr struct {
}

type LessThanOrEqualExpr struct {
}

type NotEqualExpr struct {
}

type AddExpr struct {
}

type AssertExpr struct {
}

type DivideExpr struct {
}

type ExponentiateExpr struct {
}

type FactorialExpr struct {
}

type ModuloExpr struct {
}

type MultiplyExpr struct {
}

type NegateExpr struct {
}

type SubtractExpr struct {
}

type LikeExpr struct {
}
