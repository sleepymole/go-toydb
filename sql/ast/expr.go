package ast

import (
	"fmt"

	"github.com/sleepymole/go-toydb/sql/values"
	"github.com/sleepymole/go-toydb/util/itertools"
)

type Row []*values.Value

type Rows itertools.Iterator[Row]

type Expr interface {
	fmt.Stringer
	Eval(row Row) (*values.Value, error)
}

var _ Expr = &ConstantExpr{}
var _ Expr = &FieldExpr{}
var _ Expr = &AndExpr{}
var _ Expr = &NotExpr{}
var _ Expr = &OrExpr{}
var _ Expr = &EqualExpr{}
var _ Expr = &GreatThanExpr{}
var _ Expr = &IsNullExpr{}
var _ Expr = &LessThanExpr{}
var _ Expr = &AddExpr{}
var _ Expr = &AssertExpr{}
var _ Expr = &DivideExpr{}
var _ Expr = &ExponentiateExpr{}
var _ Expr = &FactorialExpr{}
var _ Expr = &ModuloExpr{}
var _ Expr = &MultiplyExpr{}
var _ Expr = &NegateExpr{}
var _ Expr = &SubtractExpr{}

type ConstantExpr struct {
	Value *values.Value
}

func (c *ConstantExpr) String() string {
	return c.Value.String()
}

func (c *ConstantExpr) Eval(row Row) (*values.Value, error) {
	return c.Value, nil
}

type FieldExpr struct {
	Index int
	Table string
	Name  string
}

func (f *FieldExpr) String() string {
	if f.Table != "" && f.Name != "" {
		return fmt.Sprintf("%s.%s", f.Table, f.Name)
	}
	if f.Name != "" {
		return f.Name
	}
	return fmt.Sprintf("#%d", f.Index+1)
}

func (f *FieldExpr) Eval(row Row) (*values.Value, error) {
	return row[f.Index], nil
}

type AndExpr struct {
	Lhs Expr
	Rhs Expr
}

func (a *AndExpr) String() string {
	return fmt.Sprintf("(%s AND %s)", a.Lhs, a.Rhs)
}

func (a *AndExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type NotExpr struct {
	Expr
	Sub Expr
}

func (n *NotExpr) String() string {
	return fmt.Sprintf("NOT %s", n.Sub)
}

func (n *NotExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type OrExpr struct {
	Lhs Expr
	Rhs Expr
}

func (o *OrExpr) String() string {
	return fmt.Sprintf("(%s OR %s)", o.Lhs, o.Rhs)
}

func (o *OrExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type EqualExpr struct {
	Lhs Expr
	Rhs Expr
}

func (e *EqualExpr) String() string {
	return fmt.Sprintf("(%s = %s)", e.Lhs, e.Rhs)
}

func (e *EqualExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type GreatThanExpr struct {
	Lhs Expr
	Rhs Expr
}

func (g *GreatThanExpr) String() string {
	return fmt.Sprintf("(%s > %s)", g.Lhs, g.Rhs)
}

func (g *GreatThanExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type IsNullExpr struct {
	Expr Expr
}

func (i *IsNullExpr) String() string {
	return fmt.Sprintf("%s IS NULL", i.Expr)
}

func (i *IsNullExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type LessThanExpr struct {
	Lhs Expr
	Rhs Expr
}

func (l *LessThanExpr) String() string {
	return fmt.Sprintf("(%s < %s)", l.Lhs, l.Rhs)
}

func (l *LessThanExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type AddExpr struct {
	Lhs Expr
	Rhs Expr
}

func (a *AddExpr) String() string {
	return fmt.Sprintf("(%s + %s)", a.Lhs, a.Rhs)
}

func (a *AddExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type AssertExpr struct {
	Expr Expr
}

func (a *AssertExpr) String() string {
	return fmt.Sprintf("ASSERT(%s)", a.Expr)
}

func (a *AssertExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type DivideExpr struct {
	Lhs Expr
	Rhs Expr
}

func (d *DivideExpr) String() string {
	return fmt.Sprintf("(%s / %s)", d.Lhs, d.Rhs)
}

func (d *DivideExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type ExponentiateExpr struct {
	Base Expr
	Exp  Expr
}

func (e *ExponentiateExpr) String() string {
	return fmt.Sprintf("(%s ^ %s)", e.Base, e.Exp)
}

func (e *ExponentiateExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type FactorialExpr struct {
	Expr Expr
}

func (f *FactorialExpr) String() string {
	return fmt.Sprintf("%s!", f.Expr)
}

func (f *FactorialExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type ModuloExpr struct {
	Lhs Expr
	Rhs Expr
}

func (m *ModuloExpr) String() string {
	return fmt.Sprintf("(%s %% %s)", m.Lhs, m.Rhs)
}

func (m *ModuloExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type MultiplyExpr struct {
	Lhs Expr
	Rhs Expr
}

func (m *MultiplyExpr) String() string {
	return fmt.Sprintf("(%s * %s)", m.Lhs, m.Rhs)
}

func (m *MultiplyExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type NegateExpr struct {
	Sub Expr
}

func (n *NegateExpr) String() string {
	return fmt.Sprintf("-%s", n.Sub)
}

func (n *NegateExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type SubtractExpr struct {
	Lhs Expr
	Rhs Expr
}

func (s *SubtractExpr) String() string {
	return fmt.Sprintf("(%s - %s)", s.Lhs, s.Rhs)
}

func (s *SubtractExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}

type LikeExpr struct {
	Lhs Expr
	Rhs Expr
}

func (l *LikeExpr) String() string {
	return fmt.Sprintf("(%s LIKE %s)", l.Lhs, l.Rhs)
}

func (l *LikeExpr) Eval(row Row) (*values.Value, error) {
	panic("implement me")
}
