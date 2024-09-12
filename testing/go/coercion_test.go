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

func TestCoercion(t *testing.T) {
	RunScriptsWithoutNormalization(t, []ScriptTest{
		{
			Name: "Raw Literals",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 0`,
					Expected: []sql.Row{{int32(0)}},
				},
				{
					Query:    `SELECT 0.5`,
					Expected: []sql.Row{{Numeric("0.5")}},
				},
				{
					Query:    `SELECT 0.50`,
					Expected: []sql.Row{{Numeric("0.50")}},
				},
				{
					Query:    `SELECT -0.5`,
					Expected: []sql.Row{{Numeric("-0.5")}},
				},
				{
					Query:    `SELECT 12345671297673227365.5123624235623456`,
					Expected: []sql.Row{{Numeric("12345671297673227365.5123624235623456")}},
				},
				{
					Query:    `SELECT 1`,
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    `SELECT -1`,
					Expected: []sql.Row{{int32(-1)}},
				},
				{
					Query:    `SELECT 70000`,
					Expected: []sql.Row{{int32(70000)}},
				},
				{
					Query:    `SELECT 5000000000`,
					Expected: []sql.Row{{int64(5000000000)}},
				},
				{
					Query:    `SELECT 9223372036854775808`,
					Expected: []sql.Row{{Numeric("9223372036854775808")}},
				},
				{
					Query:    `SELECT ''`,
					Expected: []sql.Row{{""}},
				},
				{
					Query:    `SELECT 'test'`,
					Expected: []sql.Row{{"test"}},
				},
				{
					Query:    `SELECT '0'`,
					Expected: []sql.Row{{"0"}},
				},
			},
		},
		{
			Name: "Math Functions",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT abs(1)`,
					Expected: []sql.Row{{int32(1)}},
				},
				{
					Query:    `SELECT abs(1.5)`,
					Expected: []sql.Row{{Numeric("1.5")}},
				},
				{
					Query:    `SELECT abs(5000000000)`,
					Expected: []sql.Row{{int64(5000000000)}},
				},
				{
					Query:    `SELECT abs(9223372036854775808)`,
					Expected: []sql.Row{{Numeric("9223372036854775808")}},
				},
				{
					Query:    `SELECT abs('1')`,
					Expected: []sql.Row{{float64(1)}},
				},
				{
					Query:    `SELECT abs('1.5')`,
					Expected: []sql.Row{{float64(1.5)}},
				},
				{
					Query:    `SELECT abs('12345671297673227365.5123624235623456')`,
					Expected: []sql.Row{{float64(1.2345671297673228e+19)}},
				},
				{
					Query:    `SELECT factorial('1')`,
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:       `SELECT factorial('1.5')`,
					ExpectedErr: "invalid input",
				},
			},
		},
	})
}
