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
	"math"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initBitLength registers the functions to the catalog.
func initBitLength() {
	framework.RegisterFunction(bit_length_text)
	framework.RegisterFunction(bit_length_bit)
}

// bit_length_text represents the PostgreSQL function of the same name, taking the same parameters.
var bit_length_text = framework.Function1{
	Name:       "bit_length",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		result, err := octet_length_text.Callable(ctx, t, val1)
		if err != nil {
			return nil, err
		}
		return result.(int32) * 8, nil
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
		return bitStringLength(uint64(len(val1str)))
	},
}

// bitStringLength converts a bit string's length to the int4 returned by PostgreSQL's bit string length functions.
func bitStringLength(length uint64) (int32, error) {
	if length > math.MaxInt32 {
		return 0, pgtypes.ErrOutOfRange.New("integer")
	}
	return int32(length), nil
}
