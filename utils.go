package zeno

import (
	"strconv"
	"strings"
)

// toType tries to convert a string to a primitive type T.
// If conversion fails, it returns the zero value of T.
func toType[T any](s string) T {
	var zero T
	switch any(zero).(type) {
	case int:
		v, _ := strconv.Atoi(s)
		return any(v).(T)
	case int64:
		v, _ := strconv.ParseInt(s, 10, 64)
		return any(v).(T)
	case float64:
		v, _ := strconv.ParseFloat(s, 64)
		return any(v).(T)
	case float32:
		v, _ := strconv.ParseFloat(s, 32)
		return any(float32(v)).(T)
	case bool:
		v, _ := strconv.ParseBool(strings.ToLower(s))
		return any(v).(T)
	case string:
		return any(s).(T)
	case uint:
		v, _ := strconv.ParseUint(s, 10, 64)
		return any(uint(v)).(T)
	case uint64:
		v, _ := strconv.ParseUint(s, 10, 64)
		return any(v).(T)
	case uint32:
		v, _ := strconv.ParseUint(s, 10, 32)
		return any(uint32(v)).(T)
	case int32:
		v, _ := strconv.ParseInt(s, 10, 32)
		return any(int32(v)).(T)
	case int16:
		v, _ := strconv.ParseInt(s, 10, 16)
		return any(int16(v)).(T)
	case uint16:
		v, _ := strconv.ParseUint(s, 10, 16)
		return any(uint16(v)).(T)
	case int8:
		v, _ := strconv.ParseInt(s, 10, 8)
		return any(int8(v)).(T)
	case uint8:
		v, _ := strconv.ParseUint(s, 10, 8)
		return any(uint8(v)).(T)
	default:
		return zero
	}
}
