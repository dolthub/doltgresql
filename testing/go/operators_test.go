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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestOperators(t *testing.T) {
	RunScriptsWithoutNormalization(t, []ScriptTest{
		{
			Name: "Addition",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 1::float4 + 2::float4;`,
					Expected: []sql.Row{{float32(3)}},
				},
				{
					Query:    `SELECT 1::float4 + 2::float8;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::float4 + 2::int2;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::float4 + 2::int4;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::float4 + 2::int8;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::float4 + 2::numeric;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::float8 + 2::float4;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::float8 + 2::float8;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::float8 + 2::int2;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::float8 + 2::int4;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::float8 + 2::int8;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::float8 + 2::numeric;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::int2 + 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::int2 + 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::int2 + 2::int2;`,
					Expected: []sql.Row{{int16(3)}},
				},
				{
					Query:    `SELECT 1::int2 + 2::int4;`,
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    `SELECT 1::int2 + 2::int8;`,
					Expected: []sql.Row{{int64(3)}},
				},
				{
					Query:    `SELECT 1::int2 + 2::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT 1::int4 + 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::int4 + 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::int4 + 2::int2;`,
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    `SELECT 1::int4 + 2::int4;`,
					Expected: []sql.Row{{int32(3)}},
				},
				{
					Query:    `SELECT 1::int4 + 2::int8;`,
					Expected: []sql.Row{{int64(3)}},
				},
				{
					Query:    `SELECT 1::int4 + 2::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT 1::int8 + 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::int8 + 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::int8 + 2::int2;`,
					Expected: []sql.Row{{int64(3)}},
				},
				{
					Query:    `SELECT 1::int8 + 2::int4;`,
					Expected: []sql.Row{{int64(3)}},
				},
				{
					Query:    `SELECT 1::int8 + 2::int8;`,
					Expected: []sql.Row{{int64(3)}},
				},
				{
					Query:    `SELECT 1::int8 + 2::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT 1::numeric + 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::numeric + 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::numeric + 2::int2;`,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT 1::numeric + 2::int4;`,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT 1::numeric + 2::int8;`,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT 1::numeric + 2::numeric;`,
					Expected: []sql.Row{{Numeric("3")}},
				},
			},
		},
		{
			Name: "Subtraction",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 1::float4 - 2::float4;`,
					Expected: []sql.Row{{float32(-1)}},
				},
				{
					Query:    `SELECT 1::float4 - 2::float8;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::float4 - 2::int2;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::float4 - 2::int4;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::float4 - 2::int8;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::float4 - 2::numeric;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::float8 - 2::float4;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::float8 - 2::float8;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::float8 - 2::int2;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::float8 - 2::int4;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::float8 - 2::int8;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::float8 - 2::numeric;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::int2 - 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::int2 - 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::int2 - 2::int2;`,
					Expected: []sql.Row{{int16(-1)}},
				},
				{
					Query:    `SELECT 1::int2 - 2::int4;`,
					Expected: []sql.Row{{int32(-1)}},
				},
				{
					Query:    `SELECT 1::int2 - 2::int8;`,
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    `SELECT 1::int2 - 2::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    `SELECT 1::int4 - 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::int4 - 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::int4 - 2::int2;`,
					Expected: []sql.Row{{int32(-1)}},
				},
				{
					Query:    `SELECT 1::int4 - 2::int4;`,
					Expected: []sql.Row{{int32(-1)}},
				},
				{
					Query:    `SELECT 1::int4 - 2::int8;`,
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    `SELECT 1::int4 - 2::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    `SELECT 1::int8 - 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::int8 - 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::int8 - 2::int2;`,
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    `SELECT 1::int8 - 2::int4;`,
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    `SELECT 1::int8 - 2::int8;`,
					Expected: []sql.Row{{int64(-1)}},
				},
				{
					Query:    `SELECT 1::int8 - 2::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    `SELECT 1::numeric - 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::numeric - 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::numeric - 2::int2;`,
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    `SELECT 1::numeric - 2::int4;`,
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    `SELECT 1::numeric - 2::int8;`,
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    `SELECT 1::numeric - 2::numeric;`,
					Expected: []sql.Row{{Numeric("-1")}},
				},
			},
		},
		{
			Name: "Multiplication",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 1::float4 * 2::float4;`,
					Expected: []sql.Row{{float32(2)}},
				},
				{
					Query:    `SELECT 1::float4 * 2::float8;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::float4 * 2::int2;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::float4 * 2::int4;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::float4 * 2::int8;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::float4 * 2::numeric;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::float8 * 2::float4;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::float8 * 2::float8;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::float8 * 2::int2;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::float8 * 2::int4;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::float8 * 2::int8;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::float8 * 2::numeric;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::int2 * 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::int2 * 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::int2 * 2::int2;`,
					Expected: []sql.Row{{int16(2)}},
				},
				{
					Query:    `SELECT 1::int2 * 2::int4;`,
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    `SELECT 1::int2 * 2::int8;`,
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    `SELECT 1::int2 * 2::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 1::int4 * 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::int4 * 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::int4 * 2::int2;`,
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    `SELECT 1::int4 * 2::int4;`,
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    `SELECT 1::int4 * 2::int8;`,
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    `SELECT 1::int4 * 2::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 1::int8 * 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::int8 * 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::int8 * 2::int2;`,
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    `SELECT 1::int8 * 2::int4;`,
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    `SELECT 1::int8 * 2::int8;`,
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    `SELECT 1::int8 * 2::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 1::numeric * 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::numeric * 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::numeric * 2::int2;`,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 1::numeric * 2::int4;`,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 1::numeric * 2::int8;`,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 1::numeric * 2::numeric;`,
					Expected: []sql.Row{{Numeric("2")}},
				},
			},
		},
		{
			Name: "Division",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 8::float4 / 2::float4;`,
					Expected: []sql.Row{{float32(4)}},
				},
				{
					Query:    `SELECT 8::float4 / 2::float8;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::float4 / 2::int2;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::float4 / 2::int4;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::float4 / 2::int8;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::float4 / 2::numeric;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::float8 / 2::float4;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::float8 / 2::float8;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::float8 / 2::int2;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::float8 / 2::int4;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::float8 / 2::int8;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::float8 / 2::numeric;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::int2 / 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::int2 / 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::int2 / 2::int2;`,
					Expected: []sql.Row{{int16(4)}},
				},
				{
					Query:    `SELECT 8::int2 / 2::int4;`,
					Expected: []sql.Row{{int32(4)}},
				},
				{
					Query:    `SELECT 8::int2 / 2::int8;`,
					Expected: []sql.Row{{int64(4)}},
				},
				{
					Query:    `SELECT 8::int2 / 2::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("4")}},
				},
				{
					Query:    `SELECT 8::int4 / 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::int4 / 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::int4 / 2::int2;`,
					Expected: []sql.Row{{int32(4)}},
				},
				{
					Query:    `SELECT 8::int4 / 2::int4;`,
					Expected: []sql.Row{{int32(4)}},
				},
				{
					Query:    `SELECT 8::int4 / 2::int8;`,
					Expected: []sql.Row{{int64(4)}},
				},
				{
					Query:    `SELECT 8::int4 / 2::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("4")}},
				},
				{
					Query:    `SELECT 8::int8 / 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::int8 / 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::int8 / 2::int2;`,
					Expected: []sql.Row{{int64(4)}},
				},
				{
					Query:    `SELECT 8::int8 / 2::int4;`,
					Expected: []sql.Row{{int64(4)}},
				},
				{
					Query:    `SELECT 8::int8 / 2::int8;`,
					Expected: []sql.Row{{int64(4)}},
				},
				{
					Query:    `SELECT 8::int8 / 2::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("4")}},
				},
				{
					Query:    `SELECT 8::numeric / 2::float4;`,
					Skip:     true,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::numeric / 2::float8;`,
					Skip:     true,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::numeric / 2::int2;`,
					Expected: []sql.Row{{Numeric("4")}},
				},
				{
					Query:    `SELECT 8::numeric / 2::int4;`,
					Expected: []sql.Row{{Numeric("4")}},
				},
				{
					Query:    `SELECT 8::numeric / 2::int8;`,
					Expected: []sql.Row{{Numeric("4")}},
				},
				{
					Query:    `SELECT 8::numeric / 2::numeric;`,
					Expected: []sql.Row{{Numeric("4")}},
				},
			},
		},
		{
			Name: "Mod",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 11::int2 % 3::int2;`,
					Expected: []sql.Row{{int16(2)}},
				},
				{
					Query:    `SELECT 11::int2 % 3::int4;`,
					Skip:     true,
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    `SELECT 11::int2 % 3::int8;`,
					Skip:     true,
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    `SELECT 11::int2 % 3::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 11::int4 % 3::int2;`,
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    `SELECT 11::int4 % 3::int4;`,
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    `SELECT 11::int4 % 3::int8;`,
					Skip:     true,
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    `SELECT 11::int4 % 3::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 11::int8 % 3::int2;`,
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    `SELECT 11::int8 % 3::int4;`,
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    `SELECT 11::int8 % 3::int8;`,
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    `SELECT 11::int8 % 3::numeric;`,
					Skip:     true,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 11::numeric % 3::int2;`,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 11::numeric % 3::int4;`,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 11::numeric % 3::int8;`,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 11::numeric % 3::numeric;`,
					Expected: []sql.Row{{Numeric("2")}},
				},
			},
		},
		{
			Name: "Shift Left",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 5::int2 << 2::int2;`,
					Expected: []sql.Row{{int16(20)}},
				},
				{
					Query:    `SELECT 5::int2 << 2::int4;`,
					Expected: []sql.Row{{int16(20)}},
				},
				{
					Query:       `SELECT 5::int2 << 2::int8;`,
					ExpectedErr: "does not exist",
				},
				{
					Query:    `SELECT 5::int4 << 2::int2;`,
					Expected: []sql.Row{{int32(20)}},
				},
				{
					Query:    `SELECT 5::int4 << 2::int4;`,
					Expected: []sql.Row{{int32(20)}},
				},
				{
					Query:       `SELECT 5::int4 << 2::int8;`,
					ExpectedErr: "does not exist",
				},
				{
					Query:    `SELECT 5::int8 << 2::int2;`,
					Expected: []sql.Row{{int64(20)}},
				},
				{
					Query:    `SELECT 5::int8 << 2::int4;`,
					Expected: []sql.Row{{int64(20)}},
				},
				{
					Query:       `SELECT 5::int8 << 2::int8;`,
					ExpectedErr: "does not exist",
				},
			},
		},
		{
			Name: "Shift Right",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 17::int2 >> 2::int2;`,
					Expected: []sql.Row{{int16(4)}},
				},
				{
					Query:    `SELECT 17::int2 >> 2::int4;`,
					Expected: []sql.Row{{int16(4)}},
				},
				{
					Query:       `SELECT 17::int2 >> 2::int8;`,
					ExpectedErr: "does not exist",
				},
				{
					Query:    `SELECT 17::int4 >> 2::int2;`,
					Expected: []sql.Row{{int32(4)}},
				},
				{
					Query:    `SELECT 17::int4 >> 2::int4;`,
					Expected: []sql.Row{{int32(4)}},
				},
				{
					Query:       `SELECT 17::int4 >> 2::int8;`,
					ExpectedErr: "does not exist",
				},
				{
					Query:    `SELECT 17::int8 >> 2::int2;`,
					Expected: []sql.Row{{int64(4)}},
				},
				{
					Query:    `SELECT 17::int8 >> 2::int4;`,
					Expected: []sql.Row{{int64(4)}},
				},
				{
					Query:       `SELECT 17::int8 >> 2::int8;`,
					ExpectedErr: "does not exist",
				},
			},
		},
		{
			Name: "Bit And",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 13::int2 & 7::int2;`,
					Expected: []sql.Row{{int16(5)}},
				},
				{
					Query:    `SELECT 13::int2 & 7::int4;`,
					Skip:     true,
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    `SELECT 13::int2 & 7::int8;`,
					Skip:     true,
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    `SELECT 13::int4 & 7::int2;`,
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    `SELECT 13::int4 & 7::int4;`,
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    `SELECT 13::int4 & 7::int8;`,
					Skip:     true,
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    `SELECT 13::int8 & 7::int2;`,
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    `SELECT 13::int8 & 7::int4;`,
					Expected: []sql.Row{{int64(5)}},
				},
				{
					Query:    `SELECT 13::int8 & 7::int8;`,
					Expected: []sql.Row{{int64(5)}},
				},
			},
		},
		{
			Name: "Bit Or",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 13::int2 | 7::int2;`,
					Expected: []sql.Row{{int16(15)}},
				},
				{
					Query:    `SELECT 13::int2 | 7::int4;`,
					Skip:     true,
					Expected: []sql.Row{{int32(15)}},
				},
				{
					Query:    `SELECT 13::int2 | 7::int8;`,
					Skip:     true,
					Expected: []sql.Row{{int64(15)}},
				},
				{
					Query:    `SELECT 13::int4 | 7::int2;`,
					Expected: []sql.Row{{int32(15)}},
				},
				{
					Query:    `SELECT 13::int4 | 7::int4;`,
					Expected: []sql.Row{{int32(15)}},
				},
				{
					Query:    `SELECT 13::int4 | 7::int8;`,
					Skip:     true,
					Expected: []sql.Row{{int64(15)}},
				},
				{
					Query:    `SELECT 13::int8 | 7::int2;`,
					Expected: []sql.Row{{int64(15)}},
				},
				{
					Query:    `SELECT 13::int8 | 7::int4;`,
					Expected: []sql.Row{{int64(15)}},
				},
				{
					Query:    `SELECT 13::int8 | 7::int8;`,
					Expected: []sql.Row{{int64(15)}},
				},
			},
		},
		{
			Name: "Bit Xor",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 13::int2 # 7::int2;`,
					Expected: []sql.Row{{int16(10)}},
				},
				{
					Query:    `SELECT 13::int2 # 7::int4;`,
					Skip:     true,
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    `SELECT 13::int2 # 7::int8;`,
					Skip:     true,
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    `SELECT 13::int4 # 7::int2;`,
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    `SELECT 13::int4 # 7::int4;`,
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    `SELECT 13::int4 # 7::int8;`,
					Skip:     true,
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    `SELECT 13::int8 # 7::int2;`,
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    `SELECT 13::int8 # 7::int4;`,
					Expected: []sql.Row{{int64(10)}},
				},
				{
					Query:    `SELECT 13::int8 # 7::int8;`,
					Expected: []sql.Row{{int64(10)}},
				},
			},
		},
		{
			Name: "Negate",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT -(7::float4);`,
					Expected: []sql.Row{{float32(-7)}},
				},
				{
					Query:    `SELECT -(7::float8);`,
					Expected: []sql.Row{{float64(-7)}},
				},
				{
					Query:    `SELECT -(7::int2);`,
					Expected: []sql.Row{{int16(-7)}},
				},
				{
					Query:    `SELECT -(7::int4);`,
					Expected: []sql.Row{{int32(-7)}},
				},
				{
					Query:    `SELECT -(7::int8);`,
					Expected: []sql.Row{{int64(-7)}},
				},
				{
					Query:    `SELECT -(7::numeric);`,
					Expected: []sql.Row{{Numeric("-7")}},
				},
			},
		},
		{
			Name: "Unary Plus",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT +(7::float4);`,
					Expected: []sql.Row{{float32(7)}},
				},
				{
					Query:    `SELECT +(7::float8);`,
					Expected: []sql.Row{{float64(7)}},
				},
				{
					Query:    `SELECT +(7::int2);`,
					Expected: []sql.Row{{int16(7)}},
				},
				{
					Query:    `SELECT +(7::int4);`,
					Expected: []sql.Row{{int32(7)}},
				},
				{
					Query:    `SELECT +(7::int8);`,
					Expected: []sql.Row{{int64(7)}},
				},
				{
					Query:    `SELECT +(7::numeric);`,
					Expected: []sql.Row{{Numeric("7")}},
				},
			},
		},
		{
			Name: "Table Columns",
			SetUpScript: []string{
				`DROP TABLE IF EXISTS table_col_checks;`,
				`CREATE TABLE table_col_checks (v1 INT4, v2 INT8, v3 FLOAT4);`,
				`INSERT INTO table_col_checks VALUES (1, 2, 3), (4, 5, 6);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT v1 + v2 FROM table_col_checks ORDER BY v1;`,
					Expected: []sql.Row{{int64(3)}, {int64(9)}},
				},
				{
					Query:    `SELECT v2 - v1 FROM table_col_checks ORDER BY v1;`,
					Expected: []sql.Row{{int64(1)}, {int64(1)}},
				},
				{
					Query:    `SELECT v3 * v3 FROM table_col_checks ORDER BY v1;`,
					Expected: []sql.Row{{float32(9)}, {float32(36)}},
				},
				{
					Query:    `SELECT v1 / 2::int4, v2 / 2::int8 FROM table_col_checks ORDER BY v1;`,
					Expected: []sql.Row{{int32(0), int64(1)}, {int32(2), int64(2)}},
				},
			},
		},
	})
}
