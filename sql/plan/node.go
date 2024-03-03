package plan

import "fmt"

type Node interface {
}

type AggregationNode struct {
}

type CreateTableNode struct {
}

type DeleteNode struct {
}

type DropTableNode struct {
}

type FilterNode struct {
}

type HashJoinNode struct {
}

type IndexLookupNode struct {
}

type InsertNode struct {
}

type KeyLookupNode struct {
}

type LimitNode struct {
}

type NestedLookJoinNode struct {
}

type NothingNode struct {
}

type OffsetNode struct {
}

type OrderNode struct {
}

type ProjectionNode struct {
}

type ScanNode struct {
}

type UpdateNode struct {
}

type AggregationType uint8

const (
	AggregationAverage AggregationType = iota
	AggregationCount
	AggregationMax
	AggregationMin
	AggregationSum
)

func (a AggregationType) String() string {
	switch a {
	case AggregationAverage:
		return "average"
	case AggregationCount:
		return "count"
	case AggregationMax:
		return "maximum"
	case AggregationMin:
		return "minimum"
	case AggregationSum:
		return "sum"
	default:
		return fmt.Sprintf("unknown(%d)", a)
	}
}

type Direction uint8

const (
	DirectionAscending Direction = iota
	DirectionDescending
)

func (d Direction) String() string {
	switch d {
	case DirectionAscending:
		return "asc"
	case DirectionDescending:
		return "desc"
	default:
		return fmt.Sprintf("unknown(%d)", d)
	}
}
