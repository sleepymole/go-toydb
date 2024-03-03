package types

import (
	"fmt"
)

type DataType uint8

const (
	Null DataType = iota
	Boolean
	Integer
	Float
	String
)

func (d DataType) String() string {
	switch d {
	case Null:
		return "NULL"
	case Boolean:
		return "BOOLEAN"
	case Integer:
		return "INTEGER"
	case Float:
		return "FLOAT"
	case String:
		return "STRING"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", d)
	}
}
