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

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// typeCodec builds a storage codec from the given serialization functions and comparison.
func typeCodec(serialize func(any) ([]byte, error), deserialize func([]byte) (any, error), cmp func(any, any) int32, vectorIndexable bool) *pgtypes.TypeCodec {
	return &pgtypes.TypeCodec{
		Serialize:   serialize,
		Deserialize: deserialize,
		SerializedCompare: func(left []byte, right []byte) (int, error) {
			a, err := deserialize(left)
			if err != nil {
				return 0, err
			}
			b, err := deserialize(right)
			if err != nil {
				return 0, err
			}
			return int(cmp(a, b)), nil
		},
		VectorIndexable: vectorIndexable,
	}
}

// vectorSerialize serializes a vector as packed little-endian float32 elements, matching the format that GMS's
// sql.EncodeVector produces.
func vectorSerialize(val any) ([]byte, error) {
	vals, ok := val.([]float32)
	if !ok {
		return nil, errors.Errorf("unexpected vector value of type %T", val)
	}
	data := make([]byte, len(vals)*4)
	for i, v := range vals {
		binary.LittleEndian.PutUint32(data[i*4:], math.Float32bits(v))
	}
	return data, nil
}

// vectorDeserialize deserializes packed little-endian float32 elements.
func vectorDeserialize(data []byte) (any, error) {
	if len(data)%4 != 0 {
		return nil, errors.Errorf("malformed serialized vector of %d bytes", len(data))
	}
	vals := make([]float32, len(data)/4)
	for i := range vals {
		vals[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:]))
	}
	return vals, nil
}

// halfvecSerialize serializes a halfvec as packed little-endian half-precision bits.
func halfvecSerialize(val any) ([]byte, error) {
	vals, ok := val.([]float32)
	if !ok {
		return nil, errors.Errorf("unexpected halfvec value of type %T", val)
	}
	data := make([]byte, len(vals)*2)
	for i, v := range vals {
		binary.LittleEndian.PutUint16(data[i*2:], float32ToHalfBits(v))
	}
	return data, nil
}

// halfvecDeserialize deserializes packed little-endian half-precision bits, widening every element to float32.
func halfvecDeserialize(data []byte) (any, error) {
	if len(data)%2 != 0 {
		return nil, errors.Errorf("malformed serialized halfvec of %d bytes", len(data))
	}
	vals := make([]float32, len(data)/2)
	for i := range vals {
		vals[i] = halfBitsToFloat32(binary.LittleEndian.Uint16(data[i*2:]))
	}
	return vals, nil
}

// sparsevecSerialize serializes a sparsevec as little-endian {dim int32, nnz int32, indices [nnz]int32,
// values [nnz]float32}.
func sparsevecSerialize(val any) ([]byte, error) {
	sv, ok := val.(sparseVector)
	if !ok {
		return nil, errors.Errorf("unexpected sparsevec value of type %T", val)
	}
	data := make([]byte, 8+len(sv.Indices)*8)
	binary.LittleEndian.PutUint32(data, uint32(sv.Dim))
	binary.LittleEndian.PutUint32(data[4:], uint32(len(sv.Indices)))
	for i, index := range sv.Indices {
		binary.LittleEndian.PutUint32(data[8+i*4:], uint32(index))
	}
	valuesOffset := 8 + len(sv.Indices)*4
	for i, v := range sv.Values {
		binary.LittleEndian.PutUint32(data[valuesOffset+i*4:], math.Float32bits(v))
	}
	return data, nil
}

// sparsevecDeserialize deserializes the little-endian sparsevec format that sparsevecSerialize produces.
func sparsevecDeserialize(data []byte) (any, error) {
	if len(data) < 8 {
		return nil, errors.Errorf("malformed serialized sparsevec of %d bytes", len(data))
	}
	sv := sparseVector{Dim: int32(binary.LittleEndian.Uint32(data))}
	nnz := int(binary.LittleEndian.Uint32(data[4:]))
	if len(data) != 8+nnz*8 {
		return nil, errors.Errorf("malformed serialized sparsevec of %d bytes with %d non-zero elements", len(data), nnz)
	}
	if nnz == 0 {
		return sv, nil
	}
	sv.Indices = make([]int32, nnz)
	sv.Values = make([]float32, nnz)
	for i := range sv.Indices {
		sv.Indices[i] = int32(binary.LittleEndian.Uint32(data[8+i*4:]))
	}
	valuesOffset := 8 + nnz*4
	for i := range sv.Values {
		sv.Values[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[valuesOffset+i*4:]))
	}
	return sv, nil
}
