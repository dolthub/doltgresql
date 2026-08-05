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

func TestRootObjectCollections(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "dropping every root object restores the committed root",
			SetUpScript: []string{
				`CREATE TABLE t1 (pk INTEGER PRIMARY KEY, v1 INTEGER);`,
				`CREATE TABLE cast_src (v TEXT);`,
				`CREATE TABLE cast_dst (v TEXT);`,
				`SELECT dolt_add('.');`,
				`SELECT dolt_commit('-m', 'initial');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE SEQUENCE s1;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{{"public.s1", "f", "new table"}},
				},
				{
					Query:    `DROP SEQUENCE s1;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE DOMAIN d1 AS INTEGER;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{{"public._d1", "f", "new table"}, {"public.d1", "f", "new table"}},
				},
				{
					Query:    `DROP DOMAIN d1;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE FUNCTION f1() RETURNS INTEGER AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{{"public.f1()", "f", "new table"}},
				},
				{
					Query:    `DROP FUNCTION f1;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE PROCEDURE p1() AS $$ BEGIN END; $$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{{"public.p1()", "f", "new table"}},
				},
				{
					Query:    `DROP PROCEDURE p1;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE FUNCTION tf1() RETURNS trigger AS $$ BEGIN RETURN NEW; END; $$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TRIGGER tr1 BEFORE INSERT ON t1 FOR EACH ROW EXECUTE FUNCTION tf1();`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP TRIGGER tr1 ON t1;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP FUNCTION tf1;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE FUNCTION cf1(cast_src) RETURNS cast_dst AS $$ SELECT ROW(($1).v)::cast_dst $$ LANGUAGE SQL;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE CAST (cast_src AS cast_dst) WITH FUNCTION cf1(cast_src);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP CAST (cast_src AS cast_dst);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP FUNCTION cf1;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "root objects created by another session are visible",
			SetUpScript: []string{
				`CREATE TABLE t1 (pk INTEGER PRIMARY KEY);`,
			},
			Assertions: []ScriptTestAssertion{
				{ // Read first, so that this session has cached every collection.
					Query:    `SELECT * FROM t1;`,
					Expected: []sql.Row{},
				},
				{ // A failed lookup must not leave this session holding the collection it searched.
					Query:       `SELECT nextval('s1');`,
					ExpectedErr: `does not exist`,
				},
				{
					Username: "postgres",
					Password: "password",
					Query:    `CREATE SEQUENCE s1;`,
					Expected: []sql.Row{},
				},
				{
					Username: "postgres",
					Password: "password",
					Query:    `CREATE DOMAIN d1 AS INTEGER;`,
					Expected: []sql.Row{},
				},
				{
					Username: "postgres",
					Password: "password",
					Query:    `CREATE FUNCTION f1() RETURNS INTEGER AS $$ BEGIN RETURN 42; END; $$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT nextval('s1');`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT f1();`,
					Expected: []sql.Row{{42}},
				},
				{
					Query:    `CREATE TABLE t2 (v d1);`,
					Expected: []sql.Row{},
				},
				{ // The value advanced by the other session must be visible here too.
					Username: "postgres",
					Password: "password",
					Query:    `SELECT nextval('s1');`,
					Expected: []sql.Row{{2}},
				},
			},
		},
		{
			Name: "root objects follow checkout, merge, and reset",
			SetUpScript: []string{
				`CREATE TABLE t1 (pk INTEGER PRIMARY KEY);`,
				`SELECT dolt_add('.');`,
				`SELECT dolt_commit('-m', 'initial');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT dolt_checkout('-b', 'other');`,
					Expected: []sql.Row{{[]any{int64(0), "Switched to branch 'other'"}}},
				},
				{
					Query:    `CREATE SEQUENCE s1;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE FUNCTION f1() RETURNS INTEGER AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT f1();`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:    `SELECT length(dolt_commit('-m', 'other')::text) = 32;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT dolt_checkout('main');`,
					Expected: []sql.Row{{[]any{int64(0), "Switched to branch 'main'"}}},
				},
				{
					Query:       `SELECT f1();`,
					ExpectedErr: `'f1' not found`,
				},
				{
					Query:       `SELECT nextval('s1');`,
					ExpectedErr: `does not exist`,
				},
				{ // The returned commit hash is not deterministic.
					Query:            `SELECT dolt_merge('other');`,
					SkipResultsCheck: true,
				},
				{
					Query:    `SELECT f1();`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT nextval('s1');`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `CREATE SEQUENCE s2;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_reset('--hard');`,
					Expected: []sql.Row{{int64(0)}},
				},
				{
					Query:       `SELECT nextval('s2');`,
					ExpectedErr: `does not exist`,
				},
				{ // The committed root objects survive the reset.
					Query:    `SELECT f1();`,
					Expected: []sql.Row{{1}},
				},
				{
					Query:    `SELECT nextval('s1');`,
					Expected: []sql.Row{{1}},
				},
			},
		},
	})
}
