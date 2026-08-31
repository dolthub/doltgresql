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

// TestPlpgsqlIdentifierCasing covers how a PL/pgSQL identifier is matched to the variable it names. An
// unquoted identifier folds to lowercase and a quoted one keeps its case, at every site that names a
// variable: a reference, an assignment target, an INTO target, a RAISE parameter, and a parameter name.
// All expectations were verified against PostgreSQL 16.
func TestPlpgsqlIdentifierCasing(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "references fold to the declared name",
			SetUpScript: []string{
				`CREATE TABLE k (id int, nm text);`,
				`INSERT INTO k VALUES (1, 'a');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION r1() RETURNS int LANGUAGE plpgsql AS $$
DECLARE MyVar int := 1; BEGIN RETURN MYVAR; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT r1();`,
					Expected: []sql.Row{{1}},
				},
				{
					// A quoted declaration keeps its case, and a quoted reference finds it.
					Query: `CREATE FUNCTION r2() RETURNS int LANGUAGE plpgsql AS $$
DECLARE "MyVar" int := 1; BEGIN RETURN "MyVar"; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT r2();`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "assignment and INTO targets fold too",
			SetUpScript: []string{
				`CREATE TABLE k (id int, nm text);`,
				`INSERT INTO k VALUES (1, 'a');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION r3() RETURNS int LANGUAGE plpgsql AS $$
DECLARE myvar int; BEGIN MyVar := 7; RETURN myvar; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT r3();`,
					Expected: []sql.Row{{7}},
				},
				{
					Query: `CREATE FUNCTION r4() RETURNS int LANGUAGE plpgsql AS $$
DECLARE "MyVar" int; BEGIN "MyVar" := 7; RETURN "MyVar"; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT r4();`,
					Expected: []sql.Row{{7}},
				},
				{
					Query: `CREATE FUNCTION r7() RETURNS int LANGUAGE plpgsql AS $$
DECLARE myvar int; BEGIN SELECT id INTO MyVar FROM k; RETURN myvar; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT r7();`,
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "parameters and RAISE arguments fold too",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION r5(MyParam int) RETURNS int LANGUAGE plpgsql AS $$
BEGIN RETURN MYPARAM; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT r5(7);`,
					Expected: []sql.Row{{7}},
				},
				{
					Query: `CREATE FUNCTION r6("MyParam" int) RETURNS int LANGUAGE plpgsql AS $$
BEGIN RETURN "MyParam"; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT r6(7);`,
					Expected: []sql.Row{{7}},
				},
				{
					Query: `CREATE FUNCTION r8() RETURNS void LANGUAGE plpgsql AS $$
DECLARE myvar int := 3; BEGIN RAISE EXCEPTION 'v=%', MyVar; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT r8();`,
					ExpectedErr: `v=3`,
				},
			},
		},
		{
			Name: "a quoted name is a different identifier from an unquoted one",
			Assertions: []ScriptTestAssertion{
				{
					// `MyVar` unquoted folds to `myvar`, so it names the outer variable rather than the
					// quoted inner one.
					Query: `CREATE FUNCTION s1() RETURNS text LANGUAGE plpgsql AS $$
DECLARE myvar text := 'outer';
BEGIN
	DECLARE "MyVar" text := 'inner-quoted';
	BEGIN
		RETURN MyVar;
	END;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT s1();`,
					Expected: []sql.Row{{"outer"}},
				},
				{
					// With only a quoted declaration, an unquoted reference names nothing.
					Query: `CREATE FUNCTION s2() RETURNS int LANGUAGE plpgsql AS $$
DECLARE "MyVar" int := 1; BEGIN RETURN MyVar; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT s2();`,
					ExpectedErr: `myvar`,
				},
			},
		},
		{
			Name: "FOR loop variables fold like any other name",
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION l1() RETURNS int LANGUAGE plpgsql AS $$
DECLARE acc int := 0; BEGIN FOR I IN 1..3 LOOP acc := acc + i; END LOOP; RETURN acc; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT l1();`,
					Expected: []sql.Row{{6}},
				},
				{
					// A quoted loop variable keeps its capitals, so the generated loop condition has to
					// name it in a way that survives folding.
					Query: `CREATE FUNCTION l2() RETURNS int LANGUAGE plpgsql AS $$
DECLARE acc int := 0; BEGIN FOR "I" IN 1..3 LOOP acc := acc + "I"; END LOOP; RETURN acc; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT l2();`,
					Expected: []sql.Row{{6}},
				},
				{
					Query: `CREATE FUNCTION l3() RETURNS int LANGUAGE plpgsql AS $$
DECLARE acc int := 0; BEGIN FOR "I" IN 1..3 LOOP acc := acc + i; END LOOP; RETURN acc; END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT l3();`,
					ExpectedErr: `"i"`,
				},
			},
		},
		{
			// TG_OP is left out deliberately: it is not populated for a trigger whose source node is
			// wrapped, so it resolves to nothing regardless of how it is spelled.
			Name: "trigger records fold like any other name",
			SetUpScript: []string{
				`CREATE TABLE t (id int primary key, v int);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE FUNCTION g() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
	IF new.v < 0 THEN RAISE EXCEPTION 'negative %', new.v; END IF;
	RETURN new;
END; $$;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `CREATE TRIGGER trg BEFORE INSERT ON t FOR EACH ROW EXECUTE FUNCTION g();`,
					Expected: []sql.Row{},
				},
				{
					Query:    `INSERT INTO t VALUES (1, 5);`,
					Expected: []sql.Row{},
				},
				{
					Query:       `INSERT INTO t VALUES (2, -1);`,
					ExpectedErr: `negative -1`,
				},
				{
					Query:    `SELECT id FROM t;`,
					Expected: []sql.Row{{1}},
				},
			},
		},
	})
}
