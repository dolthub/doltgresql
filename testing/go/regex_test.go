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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestCaseInsensitiveRegex(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "case-insensitive regex",
		Assertions: []ScriptTestAssertion{
			{Query: `SELECT 'AbC' ~* '^abc$', 'AbC' !~* '^abc$';`, Expected: []sql.Row{{"t", "f"}}},
			{Query: `SELECT 'AbC' ~* '^xyz$', 'AbC' !~* '^xyz$';`, Expected: []sql.Row{{"f", "t"}}},
			{Query: `SELECT 'xxAbCyy' ~* 'abc', 'xxAbCyy' ~* '^abc$';`, Expected: []sql.Row{{"t", "f"}}},
			{Query: `SELECT '' ~* '', 'abc' ~* '', '' !~* '', '' ~* 'a';`, Expected: []sql.Row{{"t", "t", "f", "f"}}},
			{Query: `SELECT NULL ~* 'abc', 'abc' ~* NULL, NULL !~* 'abc', 'abc' !~* NULL;`, Expected: []sql.Row{{nil, nil, nil, nil}}},
			{Query: `SELECT NULL::text ~* '[', NULL::text !~* '[';`, Expected: []sql.Row{{nil, nil}}},
			{Query: `SELECT 'AbC123' ~* '^[a-z]+[[:digit:]]+$', 'AbC' ~* '^(xyz|abc)$';`, Expected: []sql.Row{{"t", "t"}}},
			{Query: `SELECT E'A\nB' ~* '^a.b$', E'A\nB' !~* '^a.b$';`, Expected: []sql.Row{{"t", "f"}}},
			{Query: `SELECT 'Éclair' ~* '^é', '日本語' ~* '^日本';`, Expected: []sql.Row{{"t", "t"}}},
			{Query: `SELECT 'AbC'::varchar ~* '^abc$', 'AbC'::name ~* '^abc$';`, Expected: []sql.Row{{"t", "t"}}},
			{Query: `SELECT 'pg_catalog'::name ~* '^PG_', 'public'::name !~* '^PG_';`, Expected: []sql.Row{{"t", "t"}}},
			{Query: `SELECT texticregexeq('AbC', '^abc$'), texticregexne('AbC', '^abc$');`, Expected: []sql.Row{{"t", "f"}}},
			{Query: `SELECT 'Application/JSON'::name ~* '^[A-Za-z0-9.-]+/[A-Za-z0-9.\+-]+$', 'application/ld+json'::name ~* '^[A-Za-z0-9.-]+/[A-Za-z0-9.\+-]+$', 'bad type'::name ~* '^[A-Za-z0-9.-]+/[A-Za-z0-9.\+-]+$';`, Expected: []sql.Row{{"t", "t", "f"}}},
			{Query: `CREATE TABLE regex_inputs (id int PRIMARY KEY, value text, pattern text);`},
			{Query: `INSERT INTO regex_inputs VALUES (1, 'AbC', '^abc$'), (2, 'AbC', '^xyz$'), (3, NULL, '['), (4, 'text', NULL);`},
			{Query: `SELECT id, value ~* pattern, value !~* pattern FROM regex_inputs ORDER BY id;`, Expected: []sql.Row{{int32(1), "t", "f"}, {int32(2), "f", "t"}, {int32(3), nil, nil}, {int32(4), nil, nil}}},
			{Query: `SELECT id FROM regex_inputs WHERE value ~* pattern ORDER BY id;`, Expected: []sql.Row{{int32(1)}}},
			{Query: `SELECT 'abc' ~* '[';`, ExpectedErr: "regular expression", ExpectedErrCode: "2201B"},
			{Query: `SELECT 'abc' !~* '(';`, ExpectedErr: "regular expression", ExpectedErrCode: "2201B"},
			{Query: `SELECT 123 ~* 'abc';`, ExpectedErr: "operator does not exist"},
			{Query: `SELECT 'abc' !~* 123;`, ExpectedErr: "operator does not exist"},
		},
	}})
}
