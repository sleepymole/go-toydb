package sqlpb

func (v *Value) Type() ValueType_Enum {
	switch v.Union.(type) {
	case *Value_Null:
		return ValueType_NULL
	case *Value_Bool:
		return ValueType_BOOL
	case *Value_Int:
		return ValueType_INT
	case *Value_Float:
		return ValueType_FLOAT
	case *Value_String_:
		return ValueType_STRING
	default:
		panic("unreachable")
	}
}
