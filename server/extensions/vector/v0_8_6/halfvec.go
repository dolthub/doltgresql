// Copyright 2026 Dolthub, Inc.
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

package v0_8_6

import (
	"encoding/binary"
	"math"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
)

// float32ToHalfBits converts a float32 to the bits of the nearest half-precision value, rounding to nearest even.
func float32ToHalfBits(f float32) uint16 {
	bits := math.Float32bits(f)
	sign := uint16(bits>>16) & 0x8000
	exp := int32(bits>>23) & 0xff
	mant := bits & 0x7fffff
	if exp == 0xff {
		if mant == 0 {
			return sign | 0x7c00
		}
		return sign | 0x7e00
	}
	halfExp := exp - 127 + 15
	if halfExp >= 0x1f {
		return sign | 0x7c00
	}
	if halfExp <= 0 {
		if halfExp < -10 {
			return sign
		}
		mant |= 0x800000
		shift := uint32(14 - halfExp)
		half := uint16(mant >> shift)
		roundBit := uint32(1) << (shift - 1)
		if mant&roundBit != 0 && mant&(3*roundBit-1) != 0 {
			half++
		}
		return sign | half
	}
	half := uint16(halfExp<<10) | uint16(mant>>13)
	const roundBit = uint32(0x1000)
	if mant&roundBit != 0 && mant&(3*roundBit-1) != 0 {
		half++
	}
	return sign | half
}

// halfBitsToFloat32 converts the bits of a half-precision value to the float32 it represents exactly.
func halfBitsToFloat32(half uint16) float32 {
	sign := uint32(half&0x8000) << 16
	exp := uint32(half>>10) & 0x1f
	mant := uint32(half & 0x3ff)
	switch exp {
	case 0x1f:
		return math.Float32frombits(sign | 0x7f800000 | mant<<13)
	case 0:
		return math.Float32frombits(sign | math.Float32bits(float32(mant)*0x1p-24))
	default:
		return math.Float32frombits(sign | (exp+112)<<23 | mant<<13)
	}
}

// halfQuantize rounds the given value to half precision, erroring when a finite value overflows the half range.
func halfQuantize(f float32) (float32, error) {
	rounded := halfBitsToFloat32(float32ToHalfBits(f))
	if math.IsInf(float64(rounded), 0) && !math.IsInf(float64(f), 0) {
		return 0, errValueOutOfRange("overflow")
	}
	return rounded, nil
}

// halfvecIn implements halfvec_in, which parses the "[x,y,...]" text representation of a half-precision vector.
func halfvecIn(ctx *sql.Context, args ...any) (any, error) {
	input := args[0].(string)
	tokens, err := splitDenseLiteral("halfvec", input)
	if err != nil {
		return nil, err
	}
	result := make([]float32, len(tokens))
	for i, token := range tokens {
		val, err := parseElement("halfvec", input, token)
		if err != nil {
			return nil, err
		}
		if err = checkElement("halfvec", val); err != nil {
			return nil, err
		}
		rounded := halfBitsToFloat32(float32ToHalfBits(val))
		if math.IsInf(float64(rounded), 0) {
			return nil, errors.Errorf(`"%s" is out of range for type halfvec`, token)
		}
		result[i] = rounded
	}
	if err = checkExpectedDims(args[2].(int32), len(result)); err != nil {
		return nil, err
	}
	return result, nil
}

// halfvecTypmodIn implements halfvec_typmod_in, which parses the dimensions modifier.
func halfvecTypmodIn(ctx *sql.Context, args ...any) (any, error) {
	return typmodInValue("halfvec", maxDenseDims, args[0].([]any))
}

// halfvecRecv implements halfvec_recv, which decodes the binary representation of a half-precision vector.
func halfvecRecv(ctx *sql.Context, args ...any) (any, error) {
	data := args[0].([]byte)
	if len(data) < 4 {
		return nil, errors.New("insufficient data left in message")
	}
	dim := int16(binary.BigEndian.Uint16(data))
	unused := int16(binary.BigEndian.Uint16(data[2:]))
	if dim < 1 {
		return nil, errAtLeastOneDim("halfvec")
	}
	if dim > maxDenseDims {
		return nil, errTooManyDims("halfvec", maxDenseDims)
	}
	if unused != 0 {
		return nil, errors.Errorf("expected unused to be 0, not %d", unused)
	}
	if len(data) != 4+int(dim)*2 {
		return nil, errors.New("insufficient data left in message")
	}
	result := make([]float32, dim)
	for i := range result {
		result[i] = halfBitsToFloat32(binary.BigEndian.Uint16(data[4+i*2:]))
		if err := checkElement("halfvec", result[i]); err != nil {
			return nil, err
		}
	}
	if err := checkExpectedDims(args[2].(int32), len(result)); err != nil {
		return nil, err
	}
	return result, nil
}

// halfvecSend implements halfvec_send, which encodes the binary representation of a half-precision vector.
func halfvecSend(ctx *sql.Context, args ...any) (any, error) {
	vals := args[0].([]float32)
	data := make([]byte, 4+len(vals)*2)
	binary.BigEndian.PutUint16(data, uint16(len(vals)))
	for i, val := range vals {
		binary.BigEndian.PutUint16(data[4+i*2:], float32ToHalfBits(val))
	}
	return data, nil
}

// halfvecToVector implements halfvec_to_vector, whose values are already exact float32s.
func halfvecToVector(ctx *sql.Context, args ...any) (any, error) {
	vals := args[0].([]float32)
	if err := checkExpectedDims(args[1].(int32), len(vals)); err != nil {
		return nil, err
	}
	return vals, nil
}

// vectorToHalfvec implements vector_to_halfvec, which rounds every element to half precision.
func vectorToHalfvec(ctx *sql.Context, args ...any) (any, error) {
	vals := args[0].([]float32)
	result := make([]float32, len(vals))
	for i, val := range vals {
		rounded, err := halfQuantize(val)
		if err != nil {
			return nil, err
		}
		result[i] = rounded
	}
	if err := checkExpectedDims(args[1].(int32), len(result)); err != nil {
		return nil, err
	}
	return result, nil
}
