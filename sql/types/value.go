package types

import (
	"cmp"
	"encoding/binary"
	"hash/fnv"
	"math"
	"strconv"
)

type Value struct {
	typ DataType
	x   int64
	s   string
}

func (v *Value) DataType() DataType {
	return v.typ
}

func (v *Value) IsNull() bool {
	return v.typ == Null
}

func (v *Value) IsBoolean() bool {
	return v.typ == Boolean
}

func (v *Value) IsInteger() bool {
	return v.typ == Integer
}

func (v *Value) IsFloat() bool {
	return v.typ == Float
}

func (v *Value) IsString() bool {
	return v.typ == String
}

func (v *Value) Boolean() bool {
	return v.x != 0
}

func (v *Value) Integer() int64 {
	return v.x
}

func (v *Value) Float() float64 {
	return math.Float64frombits(uint64(v.x))
}

func (v *Value) String() string {
	switch v.typ {
	case Null:
		return "NULL"
	case Boolean:
		if v.Boolean() {
			return "TRUE"
		} else {
			return "FALSE"
		}
	case Integer:
		return strconv.FormatInt(v.x, 10)
	case Float:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64)
	case String:
		return v.s
	default:
		return v.typ.String()
	}
}

func (v *Value) Compare(other *Value) (int, bool) {
	if other.IsNull() {
		if v.IsNull() {
			return 0, true
		}
		return 1, true
	}
	switch v.typ {
	case Null:
		return -1, true
	case Boolean:
		if !other.IsBoolean() {
			return 0, false
		}
		if v.Boolean() == other.Boolean() {
			return 0, true
		} else if v.Boolean() {
			return 1, true
		} else {
			return -1, true
		}
	case Integer:
		if other.IsInteger() {
			return cmp.Compare(v.x, other.x), true
		}
		if other.IsFloat() {
			return cmp.Compare(float64(v.x), other.Float()), true
		}
		return 0, false
	case Float:
		if other.IsInteger() {
			return cmp.Compare(v.Float(), float64(other.x)), true
		}
		if other.IsFloat() {
			return cmp.Compare(v.Float(), other.Float()), true
		}
		return 0, false
	case String:
		if !other.IsString() {
			return 0, false
		}
		return cmp.Compare(v.s, other.s), true
	default:
		return 0, false
	}
}

func (v *Value) Equal(other *Value) bool {
	c, ok := v.Compare(other)
	return ok && c == 0
}

func (v *Value) Hash() uint64 {
	h := fnv.New64a()
	h.Write([]byte{byte(v.typ)})
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(v.x))
	h.Write(buf[:])
	h.Write([]byte(v.s))
	return h.Sum64()
}

var (
	NullValue  = &Value{typ: Null}
	TrueValue  = &Value{typ: Boolean, x: 1}
	FalseValue = &Value{typ: Boolean}
)

func MakeBooleanValue(b bool) *Value {
	if b {
		return TrueValue
	}
	return FalseValue
}

func MakeFloatValue(f float64) *Value {
	return &Value{typ: Float, x: int64(math.Float64bits(f))}
}

func MakeIntegerValue(i int64) *Value {
	return &Value{typ: Integer, x: i}
}

func MakeStringValue(s string) *Value {
	return &Value{typ: String, s: s}
}
