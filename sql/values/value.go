package values

import (
	"cmp"
	"fmt"
	"math"
	"strconv"

	"github.com/sleepymole/go-toydb/sql/sqlpb"
)

type Type uint8

const (
	TypeNull Type = iota
	TypeBool
	TypeInt
	TypeFloat
	TypeString
)

func (t Type) String() string {
	switch t {
	case TypeNull:
		return "Null"
	case TypeBool:
		return "Bool"
	case TypeInt:
		return "Int"
	case TypeFloat:
		return "Float"
	case TypeString:
		return "String"
	default:
		return fmt.Sprintf("Unknown(%d)", t)
	}
}

type Value struct {
	typ Type
	x   int64
	s   string
}

func (v *Value) ToPB() *sqlpb.Value {
	switch v.typ {
	case TypeNull:
		return &sqlpb.Value{Union: &sqlpb.Value_Null{}}
	case TypeBool:
		return &sqlpb.Value{Union: &sqlpb.Value_Bool{Bool: v.Bool()}}
	case TypeInt:
		return &sqlpb.Value{Union: &sqlpb.Value_Int{Int: v.x}}
	case TypeFloat:
		return &sqlpb.Value{Union: &sqlpb.Value_Float{Float: v.Float()}}
	case TypeString:
		return &sqlpb.Value{Union: &sqlpb.Value_String_{String_: v.s}}
	default:
		return &sqlpb.Value{Union: &sqlpb.Value_Null{}}
	}
}

func (v *Value) Type() Type {
	return v.typ
}

func (v *Value) IsNull() bool {
	return v.typ == TypeNull
}

func (v *Value) IsBool() bool {
	return v.typ == TypeBool
}

func (v *Value) IsInt() bool {
	return v.typ == TypeInt
}

func (v *Value) IsFloat() bool {
	return v.typ == TypeFloat
}

func (v *Value) IsString() bool {
	return v.typ == TypeString
}

func (v *Value) Bool() bool {
	return v.x != 0
}

func (v *Value) Int() int64 {
	return v.x
}

func (v *Value) Float() float64 {
	return math.Float64frombits(uint64(v.x))
}

func (v *Value) String() string {
	switch v.typ {
	case TypeNull:
		return "NULL"
	case TypeBool:
		if v.Bool() {
			return "TRUE"
		} else {
			return "FALSE"
		}
	case TypeInt:
		return strconv.FormatInt(v.x, 10)
	case TypeFloat:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64)
	case TypeString:
		return v.s
	default:
		return fmt.Sprintf("Unknown(type=%d)", v.typ)
	}
}

func (v *Value) Equal(otherV *Value) bool {
	result, ok := v.TryCompare(otherV)
	return ok && result == 0
}

func (v *Value) TryCompare(otherV *Value) (int, bool) {
	if otherV.IsNull() {
		if v.IsNull() {
			return 0, true
		}
		return 1, true
	}
	switch v.typ {
	case TypeNull:
		return -1, true
	case TypeBool:
		if !otherV.IsBool() {
			return 0, false
		}
		if v.Bool() == otherV.Bool() {
			return 0, true
		} else if v.Bool() {
			return 1, true
		} else {
			return -1, true
		}
	case TypeInt:
		if otherV.IsInt() {
			return cmp.Compare(v.x, otherV.x), true
		}
		if otherV.IsFloat() {
			return cmp.Compare(float64(v.x), otherV.Float()), true
		}
		return 0, false
	case TypeFloat:
		if otherV.IsInt() {
			return cmp.Compare(v.Float(), float64(otherV.x)), true
		}
		if otherV.IsFloat() {
			return cmp.Compare(v.Float(), otherV.Float()), true
		}
		return 0, false
	case TypeString:
		if !otherV.IsString() {
			return 0, false
		}
		return cmp.Compare(v.s, otherV.s), true
	default:
		return 0, false
	}
}

var (
	Null  = &Value{typ: TypeNull}
	True  = &Value{typ: TypeBool, x: 1}
	False = &Value{typ: TypeBool}
)

func MakeBool(b bool) *Value {
	if b {
		return True
	}
	return False
}

func MakeInt(i int64) *Value {
	return &Value{typ: TypeInt, x: i}
}

func MakeFloat(f float64) *Value {
	return &Value{typ: TypeFloat, x: int64(math.Float64bits(f))}
}

func MakeString(s string) *Value {
	return &Value{typ: TypeString, s: s}
}
