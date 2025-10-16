// Copyright 2025 Dolthub, Inc.
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

func TestCreateProcedureLanguagePlpgsql(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Simple example",
			SetUpScript: []string{
				`CREATE TABLE test (v1 INT8);`,
				`CREATE PROCEDURE example(input INT8) AS $$
				BEGIN
					INSERT INTO test VALUES (input);
				END;
				$$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CALL example(1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL example('2');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1},
						{2},
					},
				},
			},
		},
		{
			Name: "WHILE Label",
			SetUpScript: []string{
				`CREATE TABLE test (v1 INT8);`,
				`CREATE PROCEDURE interpreted_while_label(input INT4) AS $$
DECLARE
	counter INT4;
BEGIN
	<<while_label>>
	WHILE input < 1000 LOOP
		input := input + 1;
		counter := counter + 1;
		IF counter >= 10 THEN
			EXIT while_label;
		END IF;
	END LOOP;
	INSERT INTO test VALUES (input);
END;
$$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CALL interpreted_while_label(42);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test;",
					Expected: []sql.Row{{52}},
				},
			},
		},
		{
			Name: "Overloading",
			SetUpScript: []string{
				`CREATE TABLE test (v1 TEXT);`,
				`CREATE PROCEDURE interpreted_overload(input TEXT) AS $$
DECLARE
	var1 TEXT;
BEGIN
	IF length(input) > 3 THEN
		var1 := input || '_long';
	ELSE
		var1 := input;
	END IF;
	INSERT INTO test VALUES (var1);
END;
$$ LANGUAGE plpgsql;`,
				`CREATE PROCEDURE interpreted_overload(input INT4) AS $$
DECLARE
	var1 INT4;
BEGIN
	IF input > 3 THEN
		var1 := -input;
	ELSE
		var1 := input;
	END IF;
	INSERT INTO test VALUES (var1::text);
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CALL interpreted_overload('abc');",
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL interpreted_overload('abcd');",
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL interpreted_overload(3);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL interpreted_overload(4);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{"abc"},
						{"abcd_long"},
						{"3"},
						{"-4"},
					},
				},
			},
		},
		{
			Name: "Branching",
			SetUpScript: []string{
				`CREATE TABLE test(v1 INT4, v2 INT4);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE PROCEDURE interpreted_branch(input INT4) AS $$
BEGIN
	DELETE FROM test WHERE v1 = 1;
	INSERT INTO test VALUES (1, input + 100);
END;
$$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL interpreted_branch(4);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test;",
					Expected: []sql.Row{{1, 104}},
				},
				{
					Query:    "DELETE FROM test WHERE v1 = 1;",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT dolt_checkout('-b', 'other')`,
					Expected: []sql.Row{{`{0,"Switched to branch 'other'"}`}},
				},
				{
					Query: `CREATE OR REPLACE PROCEDURE interpreted_branch(input INT4) AS $$
BEGIN
	DELETE FROM test WHERE v1 = 2;
	INSERT INTO test VALUES (2, input + 1000);
END;
$$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'updated func')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "CALL interpreted_branch(56);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test;",
					Expected: []sql.Row{{2, 1056}},
				},
				{
					Query:    "DELETE FROM test WHERE v1 = 2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT dolt_checkout('main')",
					Expected: []sql.Row{{`{0,"Switched to branch 'main'"}`}},
				},
				{
					Query:    "CALL interpreted_branch(57);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test;",
					Expected: []sql.Row{{1, 157}},
				},
			},
		},
		{
			Name: "Merging No Conflict",
			SetUpScript: []string{
				`CREATE TABLE test(v1 INT4, v2 INT4);`,
				`INSERT INTO test VALUES (1, 77);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `CREATE PROCEDURE interpreted_merging(input TEXT) AS $$
BEGIN
	DELETE FROM test WHERE v1 = 2;
	INSERT INTO test VALUES (2, input::int4 + 100);
END;
$$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL interpreted_merging('12');",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test;",
					Expected: []sql.Row{{1, 77}, {2, 112}},
				},
				{
					Query:       "CALL interpreted_merging(55);",
					ExpectedErr: "does not exist",
				},
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT dolt_checkout('-b', 'other')`,
					Expected: []sql.Row{{`{0,"Switched to branch 'other'"}`}},
				},
				{
					Query: `CREATE PROCEDURE interpreted_merging(input INT4) AS $$
BEGIN
	DELETE FROM test WHERE v1 = 3;
	INSERT INTO test VALUES (3, input::int4 + 1000);
END;
$$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'another func')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT dolt_checkout('main')",
					Expected: []sql.Row{{`{0,"Switched to branch 'main'"}`}},
				},
				{
					Query:       "CALL interpreted_merging(55);",
					ExpectedErr: "does not exist",
				},
				{
					Query: `CREATE OR REPLACE PROCEDURE interpreted_merging(input TEXT) AS $$
BEGIN
	DELETE FROM test WHERE v1 = 2;
	INSERT INTO test VALUES (2, input::int4 + 10000);
END;
$$ LANGUAGE plpgsql;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'updated table')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "SELECT length(dolt_merge('other')::text) = 57;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "CALL interpreted_merging('33');",
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL interpreted_merging(77);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test;",
					Expected: []sql.Row{{1, 77}, {2, 10033}, {3, 1077}},
				},
			},
		},
		{
			Name: `Procedure updates "definition" with custom body`,
			SetUpScript: []string{
				`CREATE TABLE test (v1 TEXT);`,
				`CREATE PROCEDURE interpreted_example(input TEXT) AS $$ BEGIN INSERT INTO test VALUES ('1' || input); END; $$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'initial')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT dolt_checkout('-b', 'other')`,
					Expected: []sql.Row{{`{0,"Switched to branch 'other'"}`}},
				},
				{
					Query:    "CREATE OR REPLACE PROCEDURE interpreted_example(input TEXT) AS $$ BEGIN INSERT INTO test VALUES ('3' || input); END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'other')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT dolt_checkout('main')`,
					Expected: []sql.Row{{`{0,"Switched to branch 'main'"}`}},
				},
				{
					Query:    "CREATE OR REPLACE PROCEDURE interpreted_example(input TEXT) AS $$ BEGIN INSERT INTO test VALUES ('2' || input); END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{
						{"public.interpreted_example(text)", "f", "modified"},
					},
				},
				{
					Query:    `SELECT dolt_add('.');`,
					Expected: []sql.Row{{"{0}"}},
				},
				{
					Query:    "SELECT length(dolt_commit('-m', 'next')::text) = 34;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query: `SELECT dolt_merge('other');`,
					Expected: []sql.Row{
						{`{0,1,"conflicts found"}`},
					},
				},
				{
					Query: `SELECT * FROM dolt_conflicts;`,
					Expected: []sql.Row{
						{"public.interpreted_example(text)", 1},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"BEGIN INSERT INTO test VALUES ('1' || input); END;", "BEGIN INSERT INTO test VALUES ('2' || input); END;", "modified", "BEGIN INSERT INTO test VALUES ('3' || input); END;", "modified", "definition"},
					},
				},
				{
					Query:    `UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'BEGIN INSERT INTO test VALUES (''7'' || input); END;' WHERE dolt_conflict_id = 'definition';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DELETE FROM "dolt_conflicts_interpreted_example(text)" WHERE dolt_conflict_id = 'definition';`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM "dolt_conflicts_interpreted_example(text)";`,
					ExpectedErr: `table not found`,
				},
				{
					Query:    "CALL interpreted_example('12');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{"712"},
					},
				},
			},
		},
	})
}
