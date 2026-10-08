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

package functions

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initBitLength registers the functions to the catalog.
func initBitLength() {
	framework.RegisterFunction(bit_length_text)
	framework.RegisterFunction(bit_length_bytea)
	framework.RegisterFunction(bit_length_bit)
}

// bit_length_text represents the PostgreSQL function of the same name, taking the same parameters.
var bit_length_text = framework.Function1{
	Name:       "bit_length",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return valueBitLength(ctx, val1)
	},
}

// bit_length_bytea represents the PostgreSQL function of the same name, taking the same parameters.
var bit_length_bytea = framework.Function1{
	Name:       "bit_length",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Bytea},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return valueBitLength(ctx, val1)
	},
}

// bit_length_bit represents the PostgreSQL function of the same name, taking the same parameters.
var bit_length_bit = framework.Function1{
	Name:       "bit_length",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Bit},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		val1str, err := framework.UnwrapString(ctx, val1)
		if err != nil {
			return nil, err
		}
		return lengthToInt32(int64(len(val1str)))
	},
}

// valueBitLength returns the number of bits in a string- or bytes-typed value, which is its byte length times eight.
func valueBitLength(ctx *sql.Context, val any) (int32, error) {
	length, err := framework.UnwrapByteLength(ctx, val)
	if err != nil {
		return 0, err
	}
	return lengthToInt32(length * 8)
}
