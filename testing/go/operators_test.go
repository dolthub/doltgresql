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
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::int2 + 2::float8;`,
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
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT 1::int4 + 2::float4;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::int4 + 2::float8;`,
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
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT 1::int8 + 2::float4;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::int8 + 2::float8;`,
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
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT 1::numeric + 2::float4;`,
					Expected: []sql.Row{{float64(3)}},
				},
				{
					Query:    `SELECT 1::numeric + 2::float8;`,
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
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::int2 - 2::float8;`,
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
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    `SELECT 1::int4 - 2::float4;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::int4 - 2::float8;`,
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
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    `SELECT 1::int8 - 2::float4;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::int8 - 2::float8;`,
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
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    `SELECT 1::numeric - 2::float4;`,
					Expected: []sql.Row{{float64(-1)}},
				},
				{
					Query:    `SELECT 1::numeric - 2::float8;`,
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
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::int2 * 2::float8;`,
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
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 1::int4 * 2::float4;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::int4 * 2::float8;`,
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
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 1::int8 * 2::float4;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::int8 * 2::float8;`,
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
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT 1::numeric * 2::float4;`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `SELECT 1::numeric * 2::float8;`,
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
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::int2 / 2::float8;`,
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
					Expected: []sql.Row{{Numeric("4")}},
				},
				{
					Query:    `SELECT 8::int4 / 2::float4;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::int4 / 2::float8;`,
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
					Expected: []sql.Row{{Numeric("4")}},
				},
				{
					Query:    `SELECT 8::int8 / 2::float4;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::int8 / 2::float8;`,
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
					Expected: []sql.Row{{Numeric("4")}},
				},
				{
					Query:    `SELECT 8::numeric / 2::float4;`,
					Expected: []sql.Row{{float64(4)}},
				},
				{
					Query:    `SELECT 8::numeric / 2::float8;`,
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
					Expected: []sql.Row{{int32(2)}},
				},
				{
					Query:    `SELECT 11::int2 % 3::int8;`,
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    `SELECT 11::int2 % 3::numeric;`,
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
					Expected: []sql.Row{{int64(2)}},
				},
				{
					Query:    `SELECT 11::int4 % 3::numeric;`,
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
			Name: "Less Than",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT false < true;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT true < false;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'abc'::bpchar < 'def'::bpchar;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'def'::bpchar < 'abc'::bpchar;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'abc'::"char" < 'def'::"char";`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT E'\\x01'::bytea < E'\\x02'::bytea;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT E'\\x02'::bytea < E'\\x01'::bytea;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2019-01-03'::date < '2020-07-15'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2020-02-05'::date < '2019-08-17'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2021-03-07'::date < '2022-09-19 04:19:19'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2022-04-09'::date < '2021-10-21 08:27:40'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2023-05-11'::date < '2024-11-23 12:35:54+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2024-06-13'::date < '2023-12-25 16:43:55+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 1.23::float4 < 4.56::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 4.56::float4 < 1.23::float4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 7.89::float4 < 9.01::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 9.01::float4 < 7.89::float8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 2.34::float8 < 5.67::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 5.67::float8 < 2.34::float4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 8.99::float8 < 9.01::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 9.01::float8 < 8.99::float8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 10::int2 < 29::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 29::int2 < 10::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 11::int2 < 28::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 28::int2 < 11::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 12::int2 < 27::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 27::int2 < 12::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 13::int4 < 26::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 26::int4 < 13::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 14::int4 < 25::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 25::int4 < 14::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 15::int4 < 24::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 24::int4 < 15::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 16::int8 < 23::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 23::int8 < 16::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 17::int8 < 22::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 22::int8 < 17::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 18::int8 < 21::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 21::int8 < 18::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '{"a":1}'::jsonb < '{"b":2}'::jsonb;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"b":2}'::jsonb < '{"a":1}'::jsonb;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'and'::name < 'then'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'then'::name < 'and'::name;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'cold'::name < 'dance'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'dance'::name < 'cold'::text;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 10.20::numeric < 20.10::numeric;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 20.10::numeric < 10.20::numeric;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 101::oid < 202::oid;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 202::oid < 101::oid;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'dog'::text < 'good'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'good'::text < 'dog'::name;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'hello'::text < 'world'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'world'::text < 'hello'::text;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '12:12:12'::time < '14:15:16'::time;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '14:15:16'::time < '12:12:12'::time;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2019-01-03 10:21:00'::timestamp < '2020-02-05'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2020-02-05 10:21:00'::timestamp < '2019-01-03'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2020-02-05 11:32:00'::timestamp < '2021-03-07 12:43:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp < '2020-02-05 11:32:00'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp < '2022-04-09 13:54:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2022-04-09 13:54:00'::timestamp < '2021-03-07 12:43:00+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2022-04-09 13:54:00+00'::timestamptz < '2023-05-11'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2023-05-11 13:54:00+00'::timestamptz < '2022-04-09'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2023-05-11 14:15:00+00'::timestamptz < '2024-06-13 13:54:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2024-06-13 13:54:00+00'::timestamptz < '2023-05-11 14:15:00'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2024-06-13 15:36:00+00'::timestamptz < '2025-07-15 14:15:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2025-07-15 14:15:00+00'::timestamptz < '2024-06-13 15:36:00+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '12:16:20+00'::timetz < '13:17:21+00'::timetz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '13:17:21+00'::timetz < '12:16:20+00'::timetz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid < '64b67ba1-e368-4cfd-ae6f-0c3e77716fb6'::uuid;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb6'::uuid < '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid;`,
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			Name: "Greater Than",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT false > true;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT true > false;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'abc'::bpchar > 'def'::bpchar;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'def'::bpchar > 'abc'::bpchar;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'abc'::"char" > 'def'::"char";`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT E'\\x01'::bytea > E'\\x02'::bytea;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT E'\\x02'::bytea > E'\\x01'::bytea;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2019-01-03'::date > '2020-07-15'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2020-02-05'::date > '2019-08-17'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2021-03-07'::date > '2022-09-19 04:19:19'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2022-04-09'::date > '2021-10-21 08:27:40'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2023-05-11'::date > '2024-11-23 12:35:54+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2024-06-13'::date > '2023-12-25 16:43:55+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 1.23::float4 > 4.56::float4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 4.56::float4 > 1.23::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 7.89::float4 > 9.01::float8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 9.01::float4 > 7.89::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 2.34::float8 > 5.67::float4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 5.67::float8 > 2.34::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 8.99::float8 > 9.01::float8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 9.01::float8 > 8.99::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 10::int2 > 29::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 29::int2 > 10::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 11::int2 > 28::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 28::int2 > 11::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 12::int2 > 27::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 27::int2 > 12::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 13::int4 > 26::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 26::int4 > 13::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 14::int4 > 25::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 25::int4 > 14::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 15::int4 > 24::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 24::int4 > 15::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 16::int8 > 23::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 23::int8 > 16::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 17::int8 > 22::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 22::int8 > 17::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 18::int8 > 21::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 21::int8 > 18::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"a":1}'::jsonb > '{"b":2}'::jsonb;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '{"b":2}'::jsonb > '{"a":1}'::jsonb;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'and'::name > 'then'::name;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'then'::name > 'and'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'cold'::name > 'dance'::text;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'dance'::name > 'cold'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 10.20::numeric > 20.10::numeric;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 20.10::numeric > 10.20::numeric;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 101::oid > 202::oid;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 202::oid > 101::oid;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'dog'::text > 'good'::name;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'good'::text > 'dog'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'hello'::text > 'world'::text;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'world'::text > 'hello'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '12:12:12'::time > '14:15:16'::time;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '14:15:16'::time > '12:12:12'::time;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2019-01-03 10:21:00'::timestamp > '2020-02-05'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2020-02-05 10:21:00'::timestamp > '2019-01-03'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2020-02-05 11:32:00'::timestamp > '2021-03-07 12:43:00'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp > '2020-02-05 11:32:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp > '2022-04-09 13:54:00+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2022-04-09 13:54:00'::timestamp > '2021-03-07 12:43:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2022-04-09 13:54:00+00'::timestamptz > '2023-05-11'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2023-05-11 13:54:00+00'::timestamptz > '2022-04-09'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2023-05-11 14:15:00+00'::timestamptz > '2024-06-13 13:54:00'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2024-06-13 13:54:00+00'::timestamptz > '2023-05-11 14:15:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2024-06-13 15:36:00+00'::timestamptz > '2025-07-15 14:15:00+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2025-07-15 14:15:00+00'::timestamptz > '2024-06-13 15:36:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '12:16:20+00'::timetz > '13:17:21+00'::timetz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '13:17:21+00'::timetz > '12:16:20+00'::timetz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid > '64b67ba1-e368-4cfd-ae6f-0c3e77716fb6'::uuid;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb6'::uuid > '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid;`,
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "Less Or Equal",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT false <= true;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT true <= true;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT true <= false;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'abc'::bpchar <= 'def'::bpchar;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'abc'::bpchar <= 'abc'::bpchar;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'def'::bpchar <= 'abc'::bpchar;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'abc'::"char" <= 'def'::"char";`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT E'\\x01'::bytea <= E'\\x02'::bytea;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT E'\\x01'::bytea <= E'\\x01'::bytea;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT E'\\x02'::bytea <= E'\\x01'::bytea;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2019-01-03'::date <= '2020-07-15'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2019-01-03'::date <= '2019-01-03'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2020-02-05'::date <= '2019-08-17'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2021-03-07'::date <= '2022-09-19 04:19:19'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2021-03-07'::date <= '2021-03-07 00:00:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2022-04-09'::date <= '2021-10-21 08:27:40'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2023-05-11'::date <= '2024-11-23 12:35:54+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2023-05-11'::date <= '2023-05-11 00:00:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2024-06-13'::date <= '2023-12-25 16:43:55+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 1.23::float4 <= 4.56::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 1.23::float4 <= 1.23::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 4.56::float4 <= 1.23::float4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 7.89::float4 <= 9.01::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 7.75::float4 <= 7.75::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 9.01::float4 <= 7.89::float8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 2.34::float8 <= 5.67::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 2.25::float8 <= 2.25::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 5.67::float8 <= 2.34::float4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 8.99::float8 <= 9.01::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 8.75::float8 <= 8.75::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 9.01::float8 <= 8.99::float8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 10::int2 <= 29::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 10::int2 <= 10::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 29::int2 <= 10::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 11::int2 <= 28::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 11::int2 <= 11::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 28::int2 <= 11::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 12::int2 <= 27::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 12::int2 <= 12::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 27::int2 <= 12::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 13::int4 <= 26::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 13::int4 <= 13::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 26::int4 <= 13::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 14::int4 <= 25::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 14::int4 <= 14::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 25::int4 <= 14::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 15::int4 <= 24::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 15::int4 <= 15::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 24::int4 <= 15::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 16::int8 <= 23::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 16::int8 <= 16::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 23::int8 <= 16::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 17::int8 <= 22::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 17::int8 <= 17::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 22::int8 <= 17::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 18::int8 <= 21::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 18::int8 <= 18::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 21::int8 <= 18::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '{"a":1}'::jsonb <= '{"b":2}'::jsonb;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"a":1}'::jsonb <= '{"a":1}'::jsonb;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"b":2}'::jsonb <= '{"a":1}'::jsonb;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'and'::name <= 'then'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'and'::name <= 'and'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'then'::name <= 'and'::name;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'cold'::name <= 'dance'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'cold'::name <= 'cold'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'dance'::name <= 'cold'::text;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 10.20::numeric <= 20.10::numeric;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 10.20::numeric <= 10.20::numeric;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 20.10::numeric <= 10.20::numeric;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 101::oid <= 202::oid;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 101::oid <= 101::oid;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 202::oid <= 101::oid;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'dog'::text <= 'good'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'dog'::text <= 'dog'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'good'::text <= 'dog'::name;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'hello'::text <= 'world'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'hello'::text <= 'hello'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'world'::text <= 'hello'::text;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '12:12:12'::time <= '14:15:16'::time;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '12:12:12'::time <= '12:12:12'::time;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '14:15:16'::time <= '12:12:12'::time;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2019-01-03 10:21:00'::timestamp <= '2020-02-05'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2019-01-03 00:00:00'::timestamp <= '2019-01-03'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2020-02-05 10:21:00'::timestamp <= '2019-01-03'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2020-02-05 11:32:00'::timestamp <= '2021-03-07 12:43:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2020-02-05 11:32:00'::timestamp <= '2020-02-05 11:32:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp <= '2020-02-05 11:32:00'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp <= '2022-04-09 13:54:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp <= '2021-03-07 12:43:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2022-04-09 13:54:00'::timestamp <= '2021-03-07 12:43:00+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2022-04-09 13:54:00+00'::timestamptz <= '2023-05-11'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2022-04-09 00:00:00+00'::timestamptz <= '2022-04-09'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2023-05-11 13:54:00+00'::timestamptz <= '2022-04-09'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2023-05-11 14:15:00+00'::timestamptz <= '2024-06-13 13:54:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2023-05-11 14:15:00+00'::timestamptz <= '2023-05-11 14:15:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2024-06-13 13:54:00+00'::timestamptz <= '2023-05-11 14:15:00'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2024-06-13 15:36:00+00'::timestamptz <= '2025-07-15 14:15:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2024-06-13 15:36:00+00'::timestamptz <= '2024-06-13 15:36:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2025-07-15 14:15:00+00'::timestamptz <= '2024-06-13 15:36:00+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '12:16:20+00'::timetz <= '13:17:21+00'::timetz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '12:16:20+00'::timetz <= '12:16:20+00'::timetz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '13:17:21+00'::timetz <= '12:16:20+00'::timetz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid <= '64b67ba1-e368-4cfd-ae6f-0c3e77716fb6'::uuid;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid <= '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb6'::uuid <= '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid;`,
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			Name: "Greater Or Equal",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT false >= true;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT true >= true;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT true >= false;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'abc'::bpchar >= 'def'::bpchar;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'abc'::bpchar >= 'abc'::bpchar;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'def'::bpchar >= 'abc'::bpchar;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'abc'::"char" >= 'def'::"char";`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT E'\\x01'::bytea >= E'\\x02'::bytea;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT E'\\x01'::bytea >= E'\\x01'::bytea;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT E'\\x02'::bytea >= E'\\x01'::bytea;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2019-01-03'::date >= '2020-07-15'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2019-01-03'::date >= '2019-01-03'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2020-02-05'::date >= '2019-08-17'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2021-03-07'::date >= '2022-09-19 04:19:19'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2021-03-07'::date >= '2021-03-07 00:00:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2022-04-09'::date >= '2021-10-21 08:27:40'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2023-05-11'::date >= '2024-11-23 12:35:54+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2023-05-11'::date >= '2023-05-11 00:00:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2024-06-13'::date >= '2023-12-25 16:43:55+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 1.23::float4 >= 4.56::float4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 1.23::float4 >= 1.23::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 4.56::float4 >= 1.23::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 7.89::float4 >= 9.01::float8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 7.75::float4 >= 7.75::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 9.01::float4 >= 7.89::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 2.34::float8 >= 5.67::float4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 2.25::float8 >= 2.25::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 5.67::float8 >= 2.34::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 8.99::float8 >= 9.01::float8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 8.75::float8 >= 8.75::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 9.01::float8 >= 8.99::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 10::int2 >= 29::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 10::int2 >= 10::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 29::int2 >= 10::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 11::int2 >= 28::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 11::int2 >= 11::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 28::int2 >= 11::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 12::int2 >= 27::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 12::int2 >= 12::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 27::int2 >= 12::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 13::int4 >= 26::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 13::int4 >= 13::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 26::int4 >= 13::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 14::int4 >= 25::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 14::int4 >= 14::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 25::int4 >= 14::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 15::int4 >= 24::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 15::int4 >= 15::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 24::int4 >= 15::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 16::int8 >= 23::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 16::int8 >= 16::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 23::int8 >= 16::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 17::int8 >= 22::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 17::int8 >= 17::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 22::int8 >= 17::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 18::int8 >= 21::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 18::int8 >= 18::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 21::int8 >= 18::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"a":1}'::jsonb >= '{"b":2}'::jsonb;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '{"a":1}'::jsonb >= '{"a":1}'::jsonb;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"b":2}'::jsonb >= '{"a":1}'::jsonb;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'and'::name >= 'then'::name;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'and'::name >= 'and'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'then'::name >= 'and'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'cold'::name >= 'dance'::text;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'cold'::name >= 'cold'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'dance'::name >= 'cold'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 10.20::numeric >= 20.10::numeric;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 10.20::numeric >= 10.20::numeric;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 20.10::numeric >= 10.20::numeric;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 101::oid >= 202::oid;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 101::oid >= 101::oid;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 202::oid >= 101::oid;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'dog'::text >= 'good'::name;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'dog'::text >= 'dog'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'good'::text >= 'dog'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'hello'::text >= 'world'::text;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'hello'::text >= 'hello'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'world'::text >= 'hello'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '12:12:12'::time >= '14:15:16'::time;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '12:12:12'::time >= '12:12:12'::time;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '14:15:16'::time >= '12:12:12'::time;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2019-01-03 10:21:00'::timestamp >= '2020-02-05'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2019-01-03 00:00:00'::timestamp >= '2019-01-03'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2020-02-05 10:21:00'::timestamp >= '2019-01-03'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2020-02-05 11:32:00'::timestamp >= '2021-03-07 12:43:00'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2020-02-05 11:32:00'::timestamp >= '2020-02-05 11:32:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp >= '2020-02-05 11:32:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp >= '2022-04-09 13:54:00+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp >= '2021-03-07 12:43:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2022-04-09 13:54:00'::timestamp >= '2021-03-07 12:43:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2022-04-09 13:54:00+00'::timestamptz >= '2023-05-11'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2022-04-09 00:00:00+00'::timestamptz >= '2022-04-09'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2023-05-11 13:54:00+00'::timestamptz >= '2022-04-09'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2023-05-11 14:15:00+00'::timestamptz >= '2024-06-13 13:54:00'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2023-05-11 14:15:00+00'::timestamptz >= '2023-05-11 14:15:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2024-06-13 13:54:00+00'::timestamptz >= '2023-05-11 14:15:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2024-06-13 15:36:00+00'::timestamptz >= '2025-07-15 14:15:00+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2024-06-13 15:36:00+00'::timestamptz >= '2024-06-13 15:36:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2025-07-15 14:15:00+00'::timestamptz >= '2024-06-13 15:36:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '12:16:20+00'::timetz >= '13:17:21+00'::timetz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '12:16:20+00'::timetz >= '12:16:20+00'::timetz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '13:17:21+00'::timetz >= '12:16:20+00'::timetz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid >= '64b67ba1-e368-4cfd-ae6f-0c3e77716fb6'::uuid;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid >= '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb6'::uuid >= '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid;`,
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "Equal",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT true = true;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT true = false;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'abc'::bpchar = 'abc'::bpchar;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'def'::bpchar = 'abc'::bpchar;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'abc'::"char" = 'abc'::"char";`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'def'::"char" = 'abc'::"char";`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT E'\\x01'::bytea = E'\\x01'::bytea;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT E'\\x02'::bytea = E'\\x01'::bytea;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2019-01-03'::date = '2019-01-03'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2020-02-05'::date = '2019-08-17'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2021-03-07'::date = '2021-03-07 00:00:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2022-04-09'::date = '2021-10-21 08:27:40'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2023-05-11'::date = '2023-05-11 00:00:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2024-06-13'::date = '2023-12-25 16:43:55+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 1.23::float4 = 1.23::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 4.56::float4 = 1.23::float4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 7.75::float4 = 7.75::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 9.01::float4 = 7.89::float8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 2.25::float8 = 2.25::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 5.67::float8 = 2.34::float4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 8.75::float8 = 8.75::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 9.01::float8 = 8.99::float8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 10::int2 = 10::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 29::int2 = 10::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 11::int2 = 11::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 28::int2 = 11::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 12::int2 = 12::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 27::int2 = 12::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 13::int4 = 13::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 26::int4 = 13::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 14::int4 = 14::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 25::int4 = 14::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 15::int4 = 15::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 24::int4 = 15::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 16::int8 = 16::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 23::int8 = 16::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 17::int8 = 17::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 22::int8 = 17::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 18::int8 = 18::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 21::int8 = 18::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '{"a":1}'::jsonb = '{"a":1}'::jsonb;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"b":2}'::jsonb = '{"a":1}'::jsonb;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'and'::name = 'and'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'then'::name = 'and'::name;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'cold'::name = 'cold'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'dance'::name = 'cold'::text;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 10.20::numeric = 10.20::numeric;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 20.10::numeric = 10.20::numeric;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 101::oid = 101::oid;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 202::oid = 101::oid;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'dog'::text = 'dog'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'good'::text = 'dog'::name;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'hello'::text = 'hello'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'world'::text = 'hello'::text;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '12:12:12'::time = '12:12:12'::time;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '14:15:16'::time = '12:12:12'::time;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2019-01-03 00:00:00'::timestamp = '2019-01-03'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2020-02-05 10:21:00'::timestamp = '2019-01-03'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2020-02-05 11:32:00'::timestamp = '2020-02-05 11:32:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp = '2020-02-05 11:32:00'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp = '2021-03-07 12:43:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2022-04-09 13:54:00'::timestamp = '2021-03-07 12:43:00+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2022-04-09 00:00:00+00'::timestamptz = '2022-04-09'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2023-05-11 13:54:00+00'::timestamptz = '2022-04-09'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2023-05-11 14:15:00+00'::timestamptz = '2023-05-11 14:15:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2024-06-13 13:54:00+00'::timestamptz = '2023-05-11 14:15:00'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2024-06-13 15:36:00+00'::timestamptz = '2024-06-13 15:36:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2025-07-15 14:15:00+00'::timestamptz = '2024-06-13 15:36:00+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '12:16:20+00'::timetz = '12:16:20+00'::timetz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '13:17:21+00'::timetz = '12:16:20+00'::timetz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid = '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb6'::uuid = '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid;`,
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			Name: "Not Equal Standard Syntax (<>)",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT true <> true;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT true <> false;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'abc'::bpchar <> 'abc'::bpchar;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'def'::bpchar <> 'abc'::bpchar;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'abc'::"char" <> 'abc'::"char";`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'def'::"char" <> 'abc'::"char";`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT E'\\x01'::bytea <> E'\\x01'::bytea;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT E'\\x02'::bytea <> E'\\x01'::bytea;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2019-01-03'::date <> '2019-01-03'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2020-02-05'::date <> '2019-08-17'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2021-03-07'::date <> '2021-03-07 00:00:00'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2022-04-09'::date <> '2021-10-21 08:27:40'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2023-05-11'::date <> '2023-05-11 00:00:00+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2024-06-13'::date <> '2023-12-25 16:43:55+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 1.23::float4 <> 1.23::float4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 4.56::float4 <> 1.23::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 7.75::float4 <> 7.75::float8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 9.01::float4 <> 7.89::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 2.25::float8 <> 2.25::float4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 5.67::float8 <> 2.34::float4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 8.75::float8 <> 8.75::float8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 9.01::float8 <> 8.99::float8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 10::int2 <> 10::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 29::int2 <> 10::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 11::int2 <> 11::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 28::int2 <> 11::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 12::int2 <> 12::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 27::int2 <> 12::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 13::int4 <> 13::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 26::int4 <> 13::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 14::int4 <> 14::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 25::int4 <> 14::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 15::int4 <> 15::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 24::int4 <> 15::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 16::int8 <> 16::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 23::int8 <> 16::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 17::int8 <> 17::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 22::int8 <> 17::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 18::int8 <> 18::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 21::int8 <> 18::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"a":1}'::jsonb <> '{"a":1}'::jsonb;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '{"b":2}'::jsonb <> '{"a":1}'::jsonb;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'and'::name <> 'and'::name;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'then'::name <> 'and'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'cold'::name <> 'cold'::text;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'dance'::name <> 'cold'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 10.20::numeric <> 10.20::numeric;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 20.10::numeric <> 10.20::numeric;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 101::oid <> 101::oid;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 202::oid <> 101::oid;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'dog'::text <> 'dog'::name;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'good'::text <> 'dog'::name;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 'hello'::text <> 'hello'::text;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 'world'::text <> 'hello'::text;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '12:12:12'::time <> '12:12:12'::time;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '14:15:16'::time <> '12:12:12'::time;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2019-01-03 00:00:00'::timestamp <> '2019-01-03'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2020-02-05 10:21:00'::timestamp <> '2019-01-03'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2020-02-05 11:32:00'::timestamp <> '2020-02-05 11:32:00'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp <> '2020-02-05 11:32:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2021-03-07 12:43:00'::timestamp <> '2021-03-07 12:43:00+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2022-04-09 13:54:00'::timestamp <> '2021-03-07 12:43:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2022-04-09 00:00:00+00'::timestamptz <> '2022-04-09'::date;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2023-05-11 13:54:00+00'::timestamptz <> '2022-04-09'::date;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2023-05-11 14:15:00+00'::timestamptz <> '2023-05-11 14:15:00'::timestamp;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2024-06-13 13:54:00+00'::timestamptz <> '2023-05-11 14:15:00'::timestamp;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '2024-06-13 15:36:00+00'::timestamptz <> '2024-06-13 15:36:00+00'::timestamptz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '2025-07-15 14:15:00+00'::timestamptz <> '2024-06-13 15:36:00+00'::timestamptz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '12:16:20+00'::timetz <> '12:16:20+00'::timetz;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '13:17:21+00'::timetz <> '12:16:20+00'::timetz;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid <> '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '64b67ba1-e368-4cfd-ae6f-0c3e77716fb6'::uuid <> '64b67ba1-e368-4cfd-ae6f-0c3e77716fb5'::uuid;`,
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "Not Equal Alternate Syntax (!=)", // This should be exactly equivalent to <>, so this is only a subset
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 10::int2 != 10::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 29::int2 != 10::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 11::int2 != 11::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 28::int2 != 11::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 12::int2 != 12::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 27::int2 != 12::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 13::int4 != 13::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 26::int4 != 13::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 14::int4 != 14::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 25::int4 != 14::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 15::int4 != 15::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 24::int4 != 15::int8;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 16::int8 != 16::int2;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 23::int8 != 16::int2;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 17::int8 != 17::int4;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 22::int8 != 17::int4;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT 18::int8 != 18::int8;`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT 21::int8 != 18::int8;`,
					Expected: []sql.Row{{"t"}},
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
					Expected: []sql.Row{{int32(5)}},
				},
				{
					Query:    `SELECT 13::int2 & 7::int8;`,
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
					Expected: []sql.Row{{int32(15)}},
				},
				{
					Query:    `SELECT 13::int2 | 7::int8;`,
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
					Expected: []sql.Row{{int32(10)}},
				},
				{
					Query:    `SELECT 13::int2 # 7::int8;`,
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
			Name: "Binary JSON",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT '[{"a":"foo"},{"b":"bar"},{"c":"baz"}]'::json -> 2;`,
					Expected: []sql.Row{{`{"c": "baz"}`}},
				},
				{
					Query:    `SELECT '[{"a":"foo"},{"b":"bar"},{"c":"baz"}]'::jsonb -> 2;`,
					Expected: []sql.Row{{`{"c": "baz"}`}},
				},
				{
					Query:    `SELECT '[{"a":"foo"},{"b":"bar"},{"c":"baz"}]'::json -> -3;`,
					Expected: []sql.Row{{`{"a": "foo"}`}},
				},
				{
					Query:    `SELECT '[{"a":"foo"},{"b":"bar"},{"c":"baz"}]'::jsonb -> -3;`,
					Expected: []sql.Row{{`{"a": "foo"}`}},
				},
				{
					Query:    `SELECT '{"a": {"b":"foo"}}'::json -> 'a';`,
					Expected: []sql.Row{{`{"b": "foo"}`}},
				},
				{
					Query:    `SELECT '{"a": {"b":"foo"}}'::jsonb -> 'a';`,
					Expected: []sql.Row{{`{"b": "foo"}`}},
				},
				{
					Query:    `SELECT '[1,2,3]'::json ->> 2;`,
					Expected: []sql.Row{{`3`}},
				},
				{
					Query:    `SELECT '[1,2,3]'::jsonb ->> 2;`,
					Expected: []sql.Row{{`3`}},
				},
				{
					Query:    `SELECT '{"a":1,"b":2}'::json ->> 'b';`,
					Expected: []sql.Row{{`2`}},
				},
				{
					Query:    `SELECT '{"a":1,"b":2}'::jsonb ->> 'b';`,
					Expected: []sql.Row{{`2`}},
				},
				{
					Query:    `SELECT '{"a": {"b": ["foo","bar"]}}'::json #> ARRAY['a','b','1']::text[];`,
					Expected: []sql.Row{{`"bar"`}},
				},
				{
					Query:    `SELECT '{"a": {"b": ["foo","bar"]}}'::json #> ARRAY['a','b','1'];`,
					Expected: []sql.Row{{`"bar"`}},
				},
				{
					Query:    `SELECT '{"a": {"b": ["foo","bar"]}}'::json #> '{a,b,1}';`,
					Expected: []sql.Row{{`"bar"`}},
				},
				{
					Query:    `SELECT '{"a": {"b": ["foo","bar"]}}'::jsonb #> ARRAY['a','b','1']::text[];`,
					Expected: []sql.Row{{`"bar"`}},
				},
				{
					Query:    `SELECT '{"a": {"b": ["foo","bar"]}}'::jsonb #> ARRAY['a','b','1'];`,
					Expected: []sql.Row{{`"bar"`}},
				},
				{
					Query:    `SELECT '{"a": {"b": ["foo","bar"]}}'::jsonb #> '{a,b,1}';`,
					Expected: []sql.Row{{`"bar"`}},
				},
				{
					Query:    `SELECT '{"a": {"b": ["foo","bar"]}}'::json #>> ARRAY['a','b','1']::text[];`,
					Expected: []sql.Row{{`bar`}},
				},
				{
					Query:    `SELECT '{"a": {"b": ["foo","bar"]}}'::json #>> ARRAY['a','b','1'];`,
					Expected: []sql.Row{{`bar`}},
				},
				{
					Query:    `SELECT '{"a": {"b": ["foo","bar"]}}'::json #>> '{a,b,1}';`,
					Expected: []sql.Row{{`bar`}},
				},
				{
					Query:    `SELECT '{"a": {"b": ["foo","bar"]}}'::jsonb #>> ARRAY['a','b','1']::text[];`,
					Expected: []sql.Row{{`bar`}},
				},
				{
					Query:    `SELECT '{"a": {"b": ["foo","bar"]}}'::jsonb #>> ARRAY['a','b','1'];`,
					Expected: []sql.Row{{`bar`}},
				},
				{
					Query:    `SELECT '{"a": {"b": ["foo","bar"]}}'::jsonb #>> '{a,b,1}';`,
					Expected: []sql.Row{{`bar`}},
				},
				{
					Query:    `SELECT '{"a":1, "b":2}'::jsonb @> '{"b":2}'::jsonb;`,
					Skip:     true,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"b":2}'::jsonb <@ '{"a":1, "b":2}'::jsonb;`,
					Skip:     true,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"a":1, "b":2}'::jsonb ? 'b';`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '["a", "b", "c"]'::jsonb ? 'b';`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"a":1, "b":2, "c":3}'::jsonb ?| ARRAY['b','d']::text[];`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '{"a":1, "b":2, "c":3}'::jsonb ?| ARRAY['b','d'];`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '["a", "b", "c"]'::jsonb ?& ARRAY['a','b']::text[];`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '["a", "b", "c"]'::jsonb ?& ARRAY['a','b'];`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT '["a", "b", "c"]'::jsonb ?& ARRAY['d','b']::text[];`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '["a", "b", "c"]'::jsonb ?& ARRAY['d','b'];`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT '["a", "b"]'::jsonb || '["a", "d"]'::jsonb;`,
					Expected: []sql.Row{{`["a", "b", "a", "d"]`}},
				},
				{
					Query:    `SELECT '{"a": "b"}'::jsonb || '{"c": "d"}'::jsonb;`,
					Expected: []sql.Row{{`{"a": "b", "c": "d"}`}},
				},
				{
					Query:    `SELECT '[1, 2]'::jsonb || '3'::jsonb;`,
					Expected: []sql.Row{{`[1, 2, 3]`}},
				},
				{
					Query:    `SELECT '{"a": "b"}'::jsonb || '42'::jsonb;`,
					Expected: []sql.Row{{`[{"a": "b"}, 42]`}},
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
		{
			Name: "Concatenate",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 'Hello, ' || 'World!';`,
					Expected: []sql.Row{{"Hello, World!"}},
				},
				{
					Query:    `SELECT '123' || '456';`,
					Expected: []sql.Row{{"123456"}},
				},
				{
					Query:    `SELECT 'foo' || 'bar' || 'baz';`,
					Expected: []sql.Row{{"foobarbaz"}},
				},
				{
					Query:    `SELECT 123 || '456';`,
					Expected: []sql.Row{{"123456"}},
				},
				{
					Query:    `SELECT '123' || 456;`,
					Expected: []sql.Row{{"123456"}},
				},
				{
					Query:    `SELECT '123' || 4.56;`,
					Expected: []sql.Row{{"1234.56"}},
				},
				{
					Query:    `SELECT 12.3 || '4.56';`,
					Expected: []sql.Row{{"12.34.56"}},
				},
				{
					Query:    `SELECT true || 'bar' || false;`,
					Expected: []sql.Row{{"truebarfalse"}},
				},
				{
					Query:    `SELECT '2000-01-01 00:00:00'::timestamp || ' happy new year';`,
					Expected: []sql.Row{{"2000-01-01 00:00:00 happy new year"}},
				},
				{
					Query:    `SELECT 'hello ' || '2000-01-01 00:00:00'::timestamp ;`,
					Expected: []sql.Row{{"hello 2000-01-01 00:00:00"}},
				},
				{
					Query:    `SELECT '2000-01-01'::timestamp || ' happy new year';`,
					Expected: []sql.Row{{"2000-01-01 00:00:00 happy new year"}},
				},
				{
					Query:    `SELECT '2000-01-01 00:00:00-08'::timestamptz || ' happy new year';`,
					Expected: []sql.Row{{"2000-01-01 00:00:00-08 happy new year"}},
				},
				{
					Query:    `SELECT 'hello ' || '2000-01-01 00:00:00-08'::timestamptz;`,
					Expected: []sql.Row{{"hello 2000-01-01 00:00:00-08"}},
				},
				{
					Query:    `SELECT '00:00:00'::time || ' midnight';`,
					Expected: []sql.Row{{"00:00:00 midnight"}},
				},
				{
					Query:    `SELECT 'midnight ' || '00:00:00'::time;`,
					Expected: []sql.Row{{"midnight 00:00:00"}},
				},
				{
					Query:    `SELECT '00:00:00-07'::timetz || ' midnight';`,
					Expected: []sql.Row{{"00:00:00-07 midnight"}},
				},
				{
					Query:    `SELECT 'midnight ' || '00:00:00-07'::timetz ;`,
					Expected: []sql.Row{{"midnight 00:00:00-07"}},
				},
				{
					Query:    `SELECT 'foo'::bytea || 'bar';`,
					Expected: []sql.Row{{[]byte{0x66, 0x6F, 0x6F, 0x62, 0x61, 0x72}}},
				},
				{
					Query:    `SELECT 'bar' || 'foo'::bytea;`,
					Expected: []sql.Row{{[]byte{0x62, 0x61, 0x72, 0x66, 0x6F, 0x6F}}},
				},
				{
					Query:    `SELECT '\xDEADBEEF'::bytea || '\xCAFEBABE'::bytea;`,
					Expected: []sql.Row{{[]byte{0xDE, 0xAD, 0xBE, 0xEF, 0xCA, 0xFE, 0xBA, 0xBE}}},
				},
			},
		},
	})
}
