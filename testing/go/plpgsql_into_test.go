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

// TestPlpgsqlSelectInto covers what an INTO clause does to its targets when the query returns something
// other than exactly one row, and when it has several targets. All expectations were verified against
// PostgreSQL 16.
func TestPlpgsqlSelectInto(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "INTO targets when the query matches no rows",
			SetUpScript: []string{
				`CREATE TABLE k (id int, nm text);`,
				`INSERT INTO k VALUES (1, 'a'), (2, 'b');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// Without STRICT this is not an error: the target is set to NULL.
					Query: `CREATE FUNCTION f_scalar_miss() RETURNS int LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	SELECT id INTO v FROM k WHERE id = 99;
	RETURN coalesce(v, -1);
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_scalar_miss();`,
					Expected: []sql.Row{{-1}},
				},
				{
					// Every target is set to NULL, not just the first.
					Query: `CREATE FUNCTION f_two_miss() RETURNS text LANGUAGE plpgsql AS $$
DECLARE a int; b text;
BEGIN
	SELECT id, nm INTO a, b FROM k WHERE id = 99;
	RETURN coalesce(a::text, 'NULLa') || '/' || coalesce(b, 'NULLb');
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_two_miss();`,
					Expected: []sql.Row{{"NULLa/NULLb"}},
				},
			},
		},
		{
			Name: "INTO with several targets assigns every one of them",
			SetUpScript: []string{
				`CREATE TABLE k (id int, nm text);`,
				`INSERT INTO k VALUES (1, 'a'), (2, 'b');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION f_two() RETURNS text LANGUAGE plpgsql AS $$
DECLARE a int; b text;
BEGIN
	SELECT id, nm INTO a, b FROM k WHERE id = 1;
	RETURN coalesce(a::text, 'NULLa') || '/' || coalesce(b, 'NULLb');
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_two();`,
					Expected: []sql.Row{{"1/a"}},
				},
			},
		},
		{
			Name: "INTO keeps the first row when the query matches several",
			SetUpScript: []string{
				`CREATE TABLE k (id int, nm text);`,
				`INSERT INTO k VALUES (1, 'a'), (2, 'b');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// Without STRICT, extra rows are discarded rather than raising.
					Query: `CREATE FUNCTION f_scalar_multi() RETURNS int LANGUAGE plpgsql AS $$
DECLARE v int;
BEGIN
	SELECT id INTO v FROM k ORDER BY id;
	RETURN v;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_scalar_multi();`,
					Expected: []sql.Row{{1}},
				},
				{
					Query: `CREATE FUNCTION f_two_multi() RETURNS text LANGUAGE plpgsql AS $$
DECLARE a int; b text;
BEGIN
	SELECT id, nm INTO a, b FROM k ORDER BY id;
	RETURN coalesce(a::text, 'NULLa') || '/' || coalesce(b, 'NULLb');
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f_two_multi();`,
					Expected: []sql.Row{{"1/a"}},
				},
			},
		},
	})
}
