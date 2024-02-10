package types

import (
	"fmt"
)

type DataType uint8

const (
	DataTypeNull DataType = iota
	DataTypeBoolean
	DataTypeInteger
	DataTypeFloat
	DataTypeString
)

func (d DataType) String() string {
	switch d {
	case DataTypeBoolean:
		return "BOOLEAN"
	case DataTypeInteger:
		return "INTEGER"
	case DataTypeFloat:
		return "FLOAT"
	case DataTypeString:
		return "STRING"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", d)
	}
}
