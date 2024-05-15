// Copyright 2024 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bytes"
	"math"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestWriterReader tests that the writer can round-trip with the reader, preserving data while writing every data type.
func TestWriterReader(t *testing.T) {
	writer := NewWriter(0)
	writer.Bool(false)
	writer.Bool(true)
	writer.Int8(-128)
	writer.Int8(-57)
	writer.Int8(0)
	writer.Int8(35)
	writer.Int8(127)
	writer.Int16(-32768)
	writer.Int16(-25756)
	writer.Int16(0)
	writer.Int16(11527)
	writer.Int16(32767)
	writer.Int32(-2147483648)
	writer.Int32(-4764286)
	writer.Int32(0)
	writer.Int32(17395298)
	writer.Int32(2147483647)
	writer.Int64(-9223372036854775808)
	writer.Int64(-716243597812645)
	writer.Int64(0)
	writer.Int64(12784331275341)
	writer.Int64(9223372036854775807)
	writer.Uint8(0)
	writer.Uint8(126)
	writer.Uint8(255)
	writer.Uint16(0)
	writer.Uint16(8465)
	writer.Uint16(65535)
	writer.Uint32(0)
	writer.Uint32(217357235)
	writer.Uint32(4294967295)
	writer.Uint64(0)
	writer.Uint64(12465143251276)
	writer.Uint64(18446744073709551615)
	writer.Float32(-1716.5625)
	writer.Float32(0)
	writer.Float32(27465)
	writer.Float64(-104451.625)
	writer.Float64(0)
	writer.Float64(26112.40625)
	writer.VariableInt(-9223372036854775808)
	writer.VariableInt(-716243597812645)
	writer.VariableInt(0)
	writer.VariableInt(12784331275341)
	writer.VariableInt(9223372036854775807)
	writer.VariableUint(0)
	writer.VariableUint(12465143251276)
	writer.VariableUint(18446744073709551615)
	writer.String("")
	writer.String("abc123")
	writer.String("This is a full sentence. 素晴らしい")
	writer.BoolSlice([]bool{true, false, false, true})
	writer.Int8Slice([]int8{-128, -57, 0, 35, 127})
	writer.Int16Slice([]int16{-32768, -25756, 0, 11527, 32767})
	writer.Int32Slice([]int32{-2147483648, -4764286, 0, 17395298, 2147483647})
	writer.Int64Slice([]int64{-9223372036854775808, -716243597812645, 0, 12784331275341, 9223372036854775807})
	writer.Uint8Slice([]uint8{0, 126, 255})
	writer.Uint16Slice([]uint16{0, 8465, 65535})
	writer.Uint32Slice([]uint32{0, 217357235, 4294967295})
	writer.Uint64Slice([]uint64{0, 12465143251276, 18446744073709551615})
	writer.Float32Slice([]float32{-1716.5625, 0, 27465})
	writer.Float64Slice([]float64{-104451.625, 0, 26112.40625})
	writer.VariableIntSlice([]int64{-9223372036854775808, -716243597812645, 0, 12784331275341, 9223372036854775807})
	writer.VariableUintSlice([]uint64{0, 12465143251276, 18446744073709551615})
	writer.StringSlice([]string{"This", "is", "a", "string", "test."})
	reader := NewReader(writer.Data())
	require.Equal(t, false, reader.Bool())
	require.Equal(t, true, reader.Bool())
	require.Equal(t, int8(-128), reader.Int8())
	require.Equal(t, int8(-57), reader.Int8())
	require.Equal(t, int8(0), reader.Int8())
	require.Equal(t, int8(35), reader.Int8())
	require.Equal(t, int8(127), reader.Int8())
	require.Equal(t, int16(-32768), reader.Int16())
	require.Equal(t, int16(-25756), reader.Int16())
	require.Equal(t, int16(0), reader.Int16())
	require.Equal(t, int16(11527), reader.Int16())
	require.Equal(t, int16(32767), reader.Int16())
	require.Equal(t, int32(-2147483648), reader.Int32())
	require.Equal(t, int32(-4764286), reader.Int32())
	require.Equal(t, int32(0), reader.Int32())
	require.Equal(t, int32(17395298), reader.Int32())
	require.Equal(t, int32(2147483647), reader.Int32())
	require.Equal(t, int64(-9223372036854775808), reader.Int64())
	require.Equal(t, int64(-716243597812645), reader.Int64())
	require.Equal(t, int64(0), reader.Int64())
	require.Equal(t, int64(12784331275341), reader.Int64())
	require.Equal(t, int64(9223372036854775807), reader.Int64())
	require.Equal(t, uint8(0), reader.Uint8())
	require.Equal(t, uint8(126), reader.Uint8())
	require.Equal(t, uint8(255), reader.Uint8())
	require.Equal(t, uint16(0), reader.Uint16())
	require.Equal(t, uint16(8465), reader.Uint16())
	require.Equal(t, uint16(65535), reader.Uint16())
	require.Equal(t, uint32(0), reader.Uint32())
	require.Equal(t, uint32(217357235), reader.Uint32())
	require.Equal(t, uint32(4294967295), reader.Uint32())
	require.Equal(t, uint64(0), reader.Uint64())
	require.Equal(t, uint64(12465143251276), reader.Uint64())
	require.Equal(t, uint64(18446744073709551615), reader.Uint64())
	require.Equal(t, float32(-1716.5625), reader.Float32())
	require.Equal(t, float32(0), reader.Float32())
	require.Equal(t, float32(27465), reader.Float32())
	require.Equal(t, float64(-104451.625), reader.Float64())
	require.Equal(t, float64(0), reader.Float64())
	require.Equal(t, float64(26112.40625), reader.Float64())
	require.Equal(t, int64(-9223372036854775808), reader.VariableInt())
	require.Equal(t, int64(-716243597812645), reader.VariableInt())
	require.Equal(t, int64(0), reader.VariableInt())
	require.Equal(t, int64(12784331275341), reader.VariableInt())
	require.Equal(t, int64(9223372036854775807), reader.VariableInt())
	require.Equal(t, uint64(0), reader.VariableUint())
	require.Equal(t, uint64(12465143251276), reader.VariableUint())
	require.Equal(t, uint64(18446744073709551615), reader.VariableUint())
	require.Equal(t, "", reader.String())
	require.Equal(t, "abc123", reader.String())
	require.Equal(t, "This is a full sentence. 素晴らしい", reader.String())
	require.Equal(t, []bool{true, false, false, true}, reader.BoolSlice())
	require.Equal(t, []int8{-128, -57, 0, 35, 127}, reader.Int8Slice())
	require.Equal(t, []int16{-32768, -25756, 0, 11527, 32767}, reader.Int16Slice())
	require.Equal(t, []int32{-2147483648, -4764286, 0, 17395298, 2147483647}, reader.Int32Slice())
	require.Equal(t, []int64{-9223372036854775808, -716243597812645, 0, 12784331275341, 9223372036854775807}, reader.Int64Slice())
	require.Equal(t, []uint8{0, 126, 255}, reader.Uint8Slice())
	require.Equal(t, []uint16{0, 8465, 65535}, reader.Uint16Slice())
	require.Equal(t, []uint32{0, 217357235, 4294967295}, reader.Uint32Slice())
	require.Equal(t, []uint64{0, 12465143251276, 18446744073709551615}, reader.Uint64Slice())
	require.Equal(t, []float32{-1716.5625, 0, 27465}, reader.Float32Slice())
	require.Equal(t, []float64{-104451.625, 0, 26112.40625}, reader.Float64Slice())
	require.Equal(t, []int64{-9223372036854775808, -716243597812645, 0, 12784331275341, 9223372036854775807}, reader.VariableIntSlice())
	require.Equal(t, []uint64{0, 12465143251276, 18446744073709551615}, reader.VariableUintSlice())
	require.Equal(t, []string{"This", "is", "a", "string", "test."}, reader.StringSlice())
}

// TestWriterOrder ensures that some types that are traditionally not byte-comparable when serialized are
// byte-comparable using the writer.
func TestWriterOrder(t *testing.T) {
	t.Run("int8", func(t *testing.T) {
		var serializedData [][]byte
		integers := []int8{8, -1, 99, 0, -45, -103, 4, 127, 1, 10, 0, -128, -33, 56}
		for _, integer := range integers {
			writer := NewWriter(1)
			writer.Int8(integer)
			serializedData = append(serializedData, writer.Data())
		}
		sort.Slice(integers, func(i, j int) bool {
			return integers[i] < integers[j]
		})
		sort.Slice(serializedData, func(i, j int) bool {
			return bytes.Compare(serializedData[i], serializedData[j]) == -1
		})
		deserializedIntegers := make([]int8, len(integers))
		for i, data := range serializedData {
			reader := NewReader(data)
			deserializedIntegers[i] = reader.Int8()
		}
		require.Equal(t, integers, deserializedIntegers)
	})
	t.Run("int16", func(t *testing.T) {
		var serializedData [][]byte
		integers := []int16{8, -1, 999, 0, -455, -103, 4, 32767, 1, 100, 0, -32768, -33, 56}
		for _, integer := range integers {
			writer := NewWriter(2)
			writer.Int16(integer)
			serializedData = append(serializedData, writer.Data())
		}
		sort.Slice(integers, func(i, j int) bool {
			return integers[i] < integers[j]
		})
		sort.Slice(serializedData, func(i, j int) bool {
			return bytes.Compare(serializedData[i], serializedData[j]) == -1
		})
		deserializedIntegers := make([]int16, len(integers))
		for i, data := range serializedData {
			reader := NewReader(data)
			deserializedIntegers[i] = reader.Int16()
		}
		require.Equal(t, integers, deserializedIntegers)
	})
	t.Run("int32", func(t *testing.T) {
		var serializedData [][]byte
		integers := []int32{8, -1, 999, 0, -4559775, -103, 4, 2147483647, 1, 100, 0, -2147483648, -33, 5667845}
		for _, integer := range integers {
			writer := NewWriter(4)
			writer.Int32(integer)
			serializedData = append(serializedData, writer.Data())
		}
		sort.Slice(integers, func(i, j int) bool {
			return integers[i] < integers[j]
		})
		sort.Slice(serializedData, func(i, j int) bool {
			return bytes.Compare(serializedData[i], serializedData[j]) == -1
		})
		deserializedIntegers := make([]int32, len(integers))
		for i, data := range serializedData {
			reader := NewReader(data)
			deserializedIntegers[i] = reader.Int32()
		}
		require.Equal(t, integers, deserializedIntegers)
	})
	t.Run("int64", func(t *testing.T) {
		var serializedData [][]byte
		integers := []int64{8, -1, 999, 0, -4559775534255, -103, 4, 9223372036854775807, 1, 100, 0, -9223372036854775808, -33, 566782356235645}
		for _, integer := range integers {
			writer := NewWriter(8)
			writer.Int64(integer)
			serializedData = append(serializedData, writer.Data())
		}
		sort.Slice(integers, func(i, j int) bool {
			return integers[i] < integers[j]
		})
		sort.Slice(serializedData, func(i, j int) bool {
			return bytes.Compare(serializedData[i], serializedData[j]) == -1
		})
		deserializedIntegers := make([]int64, len(integers))
		for i, data := range serializedData {
			reader := NewReader(data)
			deserializedIntegers[i] = reader.Int64()
		}
		require.Equal(t, integers, deserializedIntegers)
	})
	t.Run("float32", func(t *testing.T) {
		var serializedData [][]byte
		floats := []float32{8, -1.123479648834723, 999, 0, -4559775534255, -103.72453, 4, 3.40282347e+38,
			9223372036854775807, 1, -3.40282347e+38, 100.1386723987453, 0, -9223372036854775808, -33,
			float32(math.Inf(1)), float32(math.Inf(-1)), 566782356235645.1345}
		for _, integer := range floats {
			writer := NewWriter(4)
			writer.Float32(integer)
			serializedData = append(serializedData, writer.Data())
		}
		sort.Slice(floats, func(i, j int) bool {
			return floats[i] < floats[j]
		})
		sort.Slice(serializedData, func(i, j int) bool {
			return bytes.Compare(serializedData[i], serializedData[j]) == -1
		})
		deserializedFloats := make([]float32, len(floats))
		for i, data := range serializedData {
			reader := NewReader(data)
			deserializedFloats[i] = reader.Float32()
		}
		require.Equal(t, floats, deserializedFloats)
	})
	t.Run("float64", func(t *testing.T) {
		var serializedData [][]byte
		floats := []float64{8, -1.123479648834723, 999, 0, -4559775534255, -103.72453, 4, 1.7976931348623157e+308,
			9223372036854775807, 1, -1.7976931348623157e+308, 100.1386723987453, 0, -9223372036854775808, -33,
			math.Inf(1), math.Inf(-1), 566782356235645.1345}
		for _, integer := range floats {
			writer := NewWriter(8)
			writer.Float64(integer)
			serializedData = append(serializedData, writer.Data())
		}
		sort.Slice(floats, func(i, j int) bool {
			return floats[i] < floats[j]
		})
		sort.Slice(serializedData, func(i, j int) bool {
			return bytes.Compare(serializedData[i], serializedData[j]) == -1
		})
		deserializedFloats := make([]float64, len(floats))
		for i, data := range serializedData {
			reader := NewReader(data)
			deserializedFloats[i] = reader.Float64()
		}
		require.Equal(t, floats, deserializedFloats)
	})
}
