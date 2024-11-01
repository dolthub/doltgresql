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

package main

import (
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/testing/generation/utils"
)

// booleanValueGenerators contains an assortment of booleans that may be used for testing boolean types.
var booleanValueGenerators = utils.Or(
	utils.Text("false"),
	utils.Text("true"),
)

// float32ValueGenerators contains an assortment of numbers that may be used for testing REAL.
var float32ValueGenerators = utils.Or(
	utils.Text("0::float4"),
	utils.Text("-1::float4"),
	utils.Text("1::float4"),
	utils.Text("2::float4"),
	utils.Text("-2::float4"),
	utils.Text("5.25::float4"),
	utils.Text("10.87::float4"),
	utils.Text("-10::float4"),
	utils.Text("100::float4"),
	utils.Text("21050.48::float4"),
	utils.Text("100000::float4"),
	utils.Text("-1184280::float4"),
	utils.Text("2525280.279::float4"),
	utils.Text("-2147483648::float4"),
	utils.Text("2147483647.59024::float4"),
)

// float64ValueGenerators contains an assortment of numbers that may be used for testing DOUBLE PRECISION.
var float64ValueGenerators = utils.Or(
	utils.Text("0::float8"),
	utils.Text("-1::float8"),
	utils.Text("1::float8"),
	utils.Text("2::float8"),
	utils.Text("-2::float8"),
	utils.Text("5.25::float8"),
	utils.Text("10.87::float8"),
	utils.Text("-10::float8"),
	utils.Text("100::float8"),
	utils.Text("21050.48::float8"),
	utils.Text("100000::float8"),
	utils.Text("-1184280::float8"),
	utils.Text("2525280.279::float8"),
	utils.Text("-2147483648::float8"),
	utils.Text("2147483647.59024::float8"),
)

// int16ValueGenerators contains an assortment of numbers that may be used for testing SMALLINT.
var int16ValueGenerators = utils.Or(
	utils.Text("0::int2"),
	utils.Text("-1::int2"),
	utils.Text("1::int2"),
	utils.Text("2::int2"),
	utils.Text("-2::int2"),
	utils.Text("5::int2"),
	utils.Text("10::int2"),
	utils.Text("-10::int2"),
	utils.Text("100::int2"),
	utils.Text("2105::int2"),
	utils.Text("10000::int2"),
	utils.Text("-11842::int2"),
	utils.Text("25252::int2"),
	utils.Text("-32768::int2"),
	utils.Text("32767::int2"),
)

// int32ValueGenerators contains an assortment of numbers that may be used for testing INTEGER.
var int32ValueGenerators = utils.Or(
	utils.Text("0::int4"),
	utils.Text("-1::int4"),
	utils.Text("1::int4"),
	utils.Text("2::int4"),
	utils.Text("-2::int4"),
	utils.Text("5::int4"),
	utils.Text("10::int4"),
	utils.Text("-10::int4"),
	utils.Text("100::int4"),
	utils.Text("21050::int4"),
	utils.Text("100000::int4"),
	utils.Text("-1184280::int4"),
	utils.Text("2525280::int4"),
	utils.Text("-2147483648::int4"),
	utils.Text("2147483647::int4"),
)

// int64ValueGenerators contains an assortment of numbers that may be used for testing BIGINT.
var int64ValueGenerators = utils.Or(
	utils.Text("0::int8"),
	utils.Text("-1::int8"),
	utils.Text("1::int8"),
	utils.Text("2::int8"),
	utils.Text("-2::int8"),
	utils.Text("5::int8"),
	utils.Text("10::int8"),
	utils.Text("-10::int8"),
	utils.Text("1000::int8"),
	utils.Text("2105076::int8"),
	utils.Text("100000000::int8"),
	utils.Text("-5184226581::int8"),
	utils.Text("8525267290::int8"),
	utils.Text("-9223372036854775808::int8"),
	utils.Text("9223372036854775807::int8"),
)

// numericValueGenerators contains an assortment of numbers that may be used for testing NUMERIC.
var numericValueGenerators = utils.Or(
	utils.Text("0::numeric"),
	utils.Text("-1::numeric"),
	utils.Text("1::numeric"),
	utils.Text("2::numeric"),
	utils.Text("-2::numeric"),
	utils.Text("5::numeric"),
	utils.Text("10::numeric"),
	utils.Text("-10::numeric"),
	utils.Text("1000::numeric"),
	utils.Text("2105076::numeric"),
	utils.Text("100000000.2345862323456346511423652312416532::numeric"),
	utils.Text("-5184226581::numeric"),
	utils.Text("8525267290::numeric"),
	utils.Text("-79223372036854775808::numeric"),
	utils.Text("79223372036854775807::numeric"),
)

// stringValueGenerators contains an assortment of strings that may be used for testing string types.
var stringValueGenerators = utils.Or(
	utils.Text("''"),
	utils.Text("' '"),
	utils.Text("'0'"),
	utils.Text("'1'"),
	utils.Text("'a'"),
	utils.Text("'abc'"),
	utils.Text("'123'"),
	utils.Text("'value'"),
	utils.Text("'12345'"),
	utils.Text("'something'"),
	utils.Text("' something'"),
	utils.Text("'something '"),
	utils.Text("'123456789'"),
	utils.Text("'a group of words'"),
	utils.Text("'1234567890123456'"),
)

// uuidValueGenerators contains an assortment of strings that may be used for testing UUID types.
var uuidValueGenerators = utils.Or(
	utils.Text("'00000000-0000-0000-0000-000000000000'::uuid"),
	utils.Text("'3883f595-6b61-42ff-a82c-226ca0d93731'::uuid"),
	utils.Text("'ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid"),
)

// valueMappings contains the value generators for the given type.
var valueMappings = map[uint32]utils.StatementGenerator{
	pgtypes.Bool.OID:    booleanValueGenerators,
	pgtypes.Float32.OID: float32ValueGenerators,
	pgtypes.Float64.OID: float64ValueGenerators,
	pgtypes.Int16.OID:   int16ValueGenerators,
	pgtypes.Int32.OID:   int32ValueGenerators,
	pgtypes.Int64.OID:   int64ValueGenerators,
	pgtypes.Numeric.OID: numericValueGenerators,
	pgtypes.Uuid.OID:    uuidValueGenerators,
	pgtypes.VarChar.OID: stringValueGenerators,
}
