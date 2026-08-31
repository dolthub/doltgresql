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

// vectorIn implements vector_in, which parses the "[x,y,...]" text representation of a vector.
func vectorIn(ctx *sql.Context, args ...any) (any, error) {
	input := args[0].(string)
	tokens, err := splitDenseLiteral("vector", input)
	if err != nil {
		return nil, err
	}
	result := make([]float32, len(tokens))
	for i, token := range tokens {
		val, err := parseElement("vector", input, token)
		if err != nil {
			return nil, err
		}
		if err = checkElement("vector", val); err != nil {
			return nil, err
		}
		result[i] = val
	}
	if err = checkExpectedDims(args[2].(int32), len(result)); err != nil {
		return nil, err
	}
	return result, nil
}

// denseOut implements vector_out and halfvec_out, which render the "[x,y,...]" text representation.
func denseOut(ctx *sql.Context, args ...any) (any, error) {
	return formatDense(args[0].([]float32)), nil
}

// vectorTypmodIn implements vector_typmod_in, which parses the dimensions modifier.
func vectorTypmodIn(ctx *sql.Context, args ...any) (any, error) {
	return typmodInValue("vector", maxDenseDims, args[0].([]any))
}

// vectorRecv implements vector_recv, which decodes the binary representation of a vector.
func vectorRecv(ctx *sql.Context, args ...any) (any, error) {
	data := args[0].([]byte)
	if len(data) < 4 {
		return nil, errors.New("insufficient data left in message")
	}
	dim := int16(binary.BigEndian.Uint16(data))
	unused := int16(binary.BigEndian.Uint16(data[2:]))
	if dim < 1 {
		return nil, errAtLeastOneDim("vector")
	}
	if dim > maxDenseDims {
		return nil, errTooManyDims("vector", maxDenseDims)
	}
	if unused != 0 {
		return nil, errors.Errorf("expected unused to be 0, not %d", unused)
	}
	if len(data) != 4+int(dim)*4 {
		return nil, errors.New("insufficient data left in message")
	}
	result := make([]float32, dim)
	for i := range result {
		result[i] = math.Float32frombits(binary.BigEndian.Uint32(data[4+i*4:]))
		if err := checkElement("vector", result[i]); err != nil {
			return nil, err
		}
	}
	if err := checkExpectedDims(args[2].(int32), len(result)); err != nil {
		return nil, err
	}
	return result, nil
}

// vectorSend implements vector_send, which encodes the binary representation of a vector.
func vectorSend(ctx *sql.Context, args ...any) (any, error) {
	vals := args[0].([]float32)
	data := make([]byte, 4+len(vals)*4)
	binary.BigEndian.PutUint16(data, uint16(len(vals)))
	for i, val := range vals {
		binary.BigEndian.PutUint32(data[4+i*4:], math.Float32bits(val))
	}
	return data, nil
}
