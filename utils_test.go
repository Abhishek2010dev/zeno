package zeno

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToType(t *testing.T) {
	assert.Equal(t, 123, toType[int]("123"))
	assert.Equal(t, int64(-456), toType[int64]("-456"))
	assert.Equal(t, float64(3.14), toType[float64]("3.14"))
	assert.Equal(t, float32(2.718), toType[float32]("2.718"))
	assert.Equal(t, true, toType[bool]("true"))
	assert.Equal(t, false, toType[bool]("false"))
	assert.Equal(t, "hello", toType[string]("hello"))

	assert.Equal(t, uint(42), toType[uint]("42"))
	assert.Equal(t, uint8(255), toType[uint8]("255"))
	assert.Equal(t, int8(-128), toType[int8]("-128"))
	assert.Equal(t, int16(32767), toType[int16]("32767"))
	assert.Equal(t, uint16(65535), toType[uint16]("65535"))
	assert.Equal(t, int32(2147483647), toType[int32]("2147483647"))
	assert.Equal(t, uint32(4294967295), toType[uint32]("4294967295"))
	assert.Equal(t, int64(9223372036854775807), toType[int64]("9223372036854775807"))
	assert.Equal(t, uint64(18446744073709551615), toType[uint64]("18446744073709551615"))

	// Invalid inputs should return zero value
	assert.Equal(t, 0, toType[int]("invalid"))
	assert.Equal(t, float64(0), toType[float64]("invalid"))
	assert.Equal(t, false, toType[bool]("invalid"))

	// For string, it always returns the same string
	assert.Equal(t, "invalid", toType[string]("invalid"))
}
