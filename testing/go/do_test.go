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

// TestDoStatements validates anonymous PL/pgSQL block execution.
func TestDoStatements(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "empty block",
			Assertions: []ScriptTestAssertion{
				{Query: `DO $$ BEGIN NULL; END $$;`},
			},
		},
		{
			Name: "declarations and mutations",
			SetUpScript: []string{
				`CREATE TABLE do_values (v INT PRIMARY KEY)`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `DO $$ DECLARE i INT := 1; BEGIN INSERT INTO do_values VALUES (i); END $$;`},
				{Query: `SELECT * FROM do_values`, Expected: []sql.Row{{int64(1)}}},
			},
		},
		{
			Name:        "explicit language",
			SetUpScript: []string{`CREATE TABLE do_language (v INT)`},
			Assertions: []ScriptTestAssertion{
				{Query: `DO LANGUAGE plpgsql $$ BEGIN INSERT INTO do_language VALUES (1); END $$;`},
				{Query: `DO $$ BEGIN INSERT INTO do_language VALUES (2); END $$ LANGUAGE plpgsql;`},
				{Query: `SELECT * FROM do_language ORDER BY v`, Expected: []sql.Row{{int64(1)}, {int64(2)}}},
			},
		},
		{
			Name: "PostgreSQL documentation example",
			SetUpScript: []string{
				`CREATE ROLE webuser`,
				`CREATE TABLE do_source (v INT)`,
				`CREATE VIEW do_view AS SELECT * FROM do_source`,
				`CREATE TABLE do_audit (view_name TEXT)`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `DO $$DECLARE r record;
BEGIN
    FOR r IN SELECT table_schema, table_name FROM information_schema.tables
             WHERE table_type = 'VIEW' AND table_schema = 'public'
    LOOP
        EXECUTE 'GRANT ALL ON ' || quote_ident(r.table_schema) || '.' || quote_ident(r.table_name) || ' TO webuser';
        INSERT INTO do_audit VALUES (r.table_name);
    END LOOP;
END$$;`},
				{Query: `SELECT * FROM do_audit`, Expected: []sql.Row{{"do_view"}}},
			},
		},
		{
			Name: "unsupported languages",
			Assertions: []ScriptTestAssertion{
				{
					Query:           `DO LANGUAGE sql 'SELECT 1';`,
					ExpectedErr:     `language "sql" does not support inline code execution`,
					ExpectedErrCode: "0A000",
				},
				{
					Query:           `DO LANGUAGE missing_language 'BEGIN NULL; END';`,
					ExpectedErr:     `language "missing_language" does not exist`,
					ExpectedErrCode: "42704",
				},
			},
		},
		{
			Name:        "dynamic command expression errors and bindings",
			SetUpScript: []string{`CREATE TABLE do_dynamic (v INT)`},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `DO $$ BEGIN EXECUTE NULL; END $$;`,
					ExpectedErr:     "query string argument of EXECUTE is null",
					ExpectedErrCode: "22004",
				},
				{
					Query: `DO $$ DECLARE
a TEXT := 'I'; b TEXT := 'N'; c TEXT := 'S'; d TEXT := 'E'; e TEXT := 'R';
f TEXT := 'T'; g TEXT := ' '; h TEXT := 'O'; i TEXT := ''; j TEXT := ''; k TEXT := '10';
BEGIN EXECUTE a || b || c || d || e || f || g || a || b || f || h || i || j || g || 'do_dynamic VALUES (' || k || ')'; END $$;`,
				},
				{Query: `SELECT * FROM do_dynamic`, Expected: []sql.Row{{int64(10)}}},
			},
		},
		{
			Name:        "statement failure is atomic",
			SetUpScript: []string{`CREATE TABLE do_atomic (v INT PRIMARY KEY)`},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `DO $$ BEGIN INSERT INTO do_atomic VALUES (1); RAISE EXCEPTION 'boom'; END $$;`,
					ExpectedErr:     "boom",
					ExpectedErrCode: "P0001",
				},
				{Query: `SELECT * FROM do_atomic`, Expected: []sql.Row{}},
			},
		},
		{
			Name: "dynamic execute using expressions",
			SetUpScript: []string{
				`CREATE TABLE do_using (v INT)`,
				`CREATE TABLE do_repeated_using (total INT, value TEXT)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `DO $$ DECLARE r RECORD; __dynamic_using_0__ INT := 40; __dynamic_using_1__ RECORD; BEGIN
EXECUTE 'INSERT INTO do_using VALUES ($1), ($2), ($3)' USING 1 + 1, length('abc'), NULL;
EXECUTE 'SELECT $1 AS v' INTO r USING 5 + 1;
INSERT INTO do_using VALUES (r.v);
EXECUTE 'INSERT INTO do_using VALUES ($1), ($2)' USING __dynamic_using_0__ + 1, __dynamic_using_0__ + 2;
EXECUTE 'SELECT $1' INTO __dynamic_using_0__ USING 43;
EXECUTE 'SELECT $1 AS v' INTO __dynamic_using_1__ USING 44;
INSERT INTO do_using VALUES (__dynamic_using_0__), (__dynamic_using_1__.v);
END $$;`,
				},
				{
					Query: `DO $$ DECLARE total INT; value TEXT; r RECORD; BEGIN
EXECUTE 'SELECT $1 + $1, $2::text' INTO total, value USING 6, NULL;
INSERT INTO do_repeated_using VALUES (total, value);
EXECUTE 'SELECT $1 + $1 AS total, $2::text AS value' INTO r USING 7, NULL;
INSERT INTO do_repeated_using VALUES (r.total, r.value);
EXECUTE 'INSERT INTO do_using VALUES ($1), ($1), ($10), ($10)' USING 1, 2, 3, 4, 5, 6, 7, 8, 9, 10;
END $$;`,
				},
				{Query: `SELECT * FROM do_repeated_using ORDER BY total`, Expected: []sql.Row{{int64(12), nil}, {int64(14), nil}}},
				{Query: `SELECT v, count(*) FROM do_using WHERE v IN (1, 10) GROUP BY v ORDER BY v`, Expected: []sql.Row{{int64(1), int64(2)}, {int64(10), int64(2)}}},
				{
					Query:    `SELECT * FROM do_using WHERE v IS NOT NULL ORDER BY v`,
					Expected: []sql.Row{{int64(1)}, {int64(1)}, {int64(2)}, {int64(3)}, {int64(6)}, {int64(10)}, {int64(10)}, {int64(41)}, {int64(42)}, {int64(43)}, {int64(44)}},
				},
				{Query: `SELECT count(*) FROM do_using WHERE v IS NULL`, Expected: []sql.Row{{int64(1)}}},
			},
		},
	})
}
