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

func TestConflictsRootObject(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:        `Function delete "definition" conflict without modification`,
			SetUpScript: []string{`CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"112"},
					},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"312"},
					},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS INT2 AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{212},
					},
				},
				{
					Query: `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{
						{"pg_catalog.interpreted_example(text)", "f", "modified"},
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
						{"pg_catalog.interpreted_example(text)", 1},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"BEGIN RETURN '1' || input; END;", "BEGIN RETURN '2' || input; END;", "modified", "BEGIN RETURN '3' || input; END;", "modified", "definition"},
					},
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
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{212},
					},
				},
			},
		},
		{
			Name:        `Function update "definition" with custom body`,
			SetUpScript: []string{`CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"112"},
					},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"312"},
					},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"212"},
					},
				},
				{
					Query: `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{
						{"pg_catalog.interpreted_example(text)", "f", "modified"},
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
						{"pg_catalog.interpreted_example(text)", 1},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"BEGIN RETURN '1' || input; END;", "BEGIN RETURN '2' || input; END;", "modified", "BEGIN RETURN '3' || input; END;", "modified", "definition"},
					},
				},
				{
					Query:    `UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'BEGIN RETURN ''7'' || input; END;' WHERE dolt_conflict_id = 'definition';`,
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
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"712"},
					},
				},
			},
		},
		{
			Name:        `Function update "definition" with "theirs" body`,
			SetUpScript: []string{`CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"112"},
					},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"312"},
					},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"212"},
					},
				},
				{
					Query: `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{
						{"pg_catalog.interpreted_example(text)", "f", "modified"},
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
						{"pg_catalog.interpreted_example(text)", 1},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"BEGIN RETURN '1' || input; END;", "BEGIN RETURN '2' || input; END;", "modified", "BEGIN RETURN '3' || input; END;", "modified", "definition"},
					},
				},
				{
					Query:    `UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = their_value WHERE dolt_conflict_id = 'definition';`,
					Expected: []sql.Row{},
				},
				{ // Updating to "theirs" will delete the conflicts table since it was the only conflict
					Query:       `DELETE FROM "dolt_conflicts_interpreted_example(text)" WHERE dolt_conflict_id = 'definition';`,
					ExpectedErr: `table not found`,
				},
				{
					Query:       `SELECT * FROM "dolt_conflicts_interpreted_example(text)";`,
					ExpectedErr: `table not found`,
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"312"},
					},
				},
			},
		},
		{
			Name:        `Function update "definition" with "ancestor" body`,
			SetUpScript: []string{`CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"112"},
					},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"312"},
					},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"212"},
					},
				},
				{
					Query: `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{
						{"pg_catalog.interpreted_example(text)", "f", "modified"},
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
						{"pg_catalog.interpreted_example(text)", 1},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"BEGIN RETURN '1' || input; END;", "BEGIN RETURN '2' || input; END;", "modified", "BEGIN RETURN '3' || input; END;", "modified", "definition"},
					},
				},
				{
					Query:    `UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = base_value WHERE dolt_conflict_id = 'definition';`,
					Expected: []sql.Row{},
				},
				{ // Updating to "ancestor" will delete the conflicts table since it was the only conflict
					Query:       `DELETE FROM "dolt_conflicts_interpreted_example(text)" WHERE dolt_conflict_id = 'definition';`,
					ExpectedErr: `table not found`,
				},
				{
					Query:       `SELECT * FROM "dolt_conflicts_interpreted_example(text)";`,
					ExpectedErr: `table not found`,
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"112"},
					},
				},
			},
		},
		{
			Name:        `Function update "return_type" with custom type`,
			SetUpScript: []string{`CREATE FUNCTION interpreted_example(input TEXT) RETURNS INT4 AS $$ BEGIN RETURN input || ''; END; $$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{12},
					},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS INT8 AS $$ BEGIN RETURN input || ''; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{12},
					},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS FLOAT AS $$ BEGIN RETURN input || ''; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{12.0},
					},
				},
				{
					Query: `SELECT * FROM dolt_status;`,
					Expected: []sql.Row{
						{"pg_catalog.interpreted_example(text)", "f", "modified"},
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
						{"pg_catalog.interpreted_example(text)", 1},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"int4", "float8", "modified", "int8", "modified", "return_type"},
					},
				},
				{
					Query:    `UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'int2' WHERE dolt_conflict_id = 'return_type';`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"int4", "int2", "modified", "int8", "modified", "return_type"},
					},
				},
				{
					Query:    `DELETE FROM "dolt_conflicts_interpreted_example(text)" WHERE dolt_conflict_id = 'return_type';`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM "dolt_conflicts_interpreted_example(text)";`,
					ExpectedErr: `table not found`,
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{12},
					},
				},
				{
					Query:       "SELECT interpreted_example('123456');",
					ExpectedErr: `int2`,
				},
			},
		},
		{
			Name:        `Function deleted "ours" updated "theirs", chose "ours"`,
			Skip:        true, // TODO: there's a bug dealing with root objects being incorrectly treated as tables
			SetUpScript: []string{`CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;`},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
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
					Query:    "DROP FUNCTION interpreted_example(input TEXT);",
					Expected: []sql.Row{},
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
						{"pg_catalog.interpreted_example(text)", 1},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"ancestor", nil, "deleted", "theirs", "modified", "root_object"},
					},
				},
				{
					Query:    `UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'ours';`,
					Expected: []sql.Row{},
				},
				{ // Updating to "ours" will delete the conflicts table since we chose our deleted root object
					Query:       `DELETE FROM "dolt_conflicts_interpreted_example(text)";`,
					ExpectedErr: `table not found`,
				},
				{
					Query:       "SELECT interpreted_example('12');",
					ExpectedErr: `'interpreted_example' not found`,
				},
			},
		},
		{
			Name:        `Function deleted "ours" updated "theirs", chose "theirs"`,
			Skip:        true, // TODO: there's a bug dealing with root objects being incorrectly treated as tables
			SetUpScript: []string{`CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;`},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
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
					Query:    "DROP FUNCTION interpreted_example(input TEXT);",
					Expected: []sql.Row{},
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
						{"pg_catalog.interpreted_example(text)", 1},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"ancestor", nil, "deleted", "theirs", "modified", "root_object"},
					},
				},
				{
					Query:    `UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'theirs';`,
					Expected: []sql.Row{},
				},
				{ // Updating to "theirs" will delete the conflicts table since we chose the other root object to keep
					Query:       `DELETE FROM "dolt_conflicts_interpreted_example(text)";`,
					ExpectedErr: `table not found`,
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"312"},
					},
				},
			},
		},
		{
			Name:        `Function deleted "ours" updated "theirs", chose "ancestor"`,
			Skip:        true, // TODO: there's a bug dealing with root objects being incorrectly treated as tables
			SetUpScript: []string{`CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;`},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '3' || input; END; $$ LANGUAGE plpgsql;",
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
					Query:    "DROP FUNCTION interpreted_example(input TEXT);",
					Expected: []sql.Row{},
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
						{"pg_catalog.interpreted_example(text)", 1},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"ancestor", nil, "deleted", "theirs", "modified", "root_object"},
					},
				},
				{
					Query:    `UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'ancestor';`,
					Expected: []sql.Row{},
				},
				{ // Updating to "ancestor" will delete the conflicts table since we effectively rolled back the merge
					Query:       `DELETE FROM "dolt_conflicts_interpreted_example(text)";`,
					ExpectedErr: `table not found`,
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"112"},
					},
				},
			},
		},
		{
			Name:        `Function deleted "theirs" updated "ours", chose "ours"`,
			Skip:        true, // TODO: there's a bug dealing with root objects being incorrectly treated as tables
			SetUpScript: []string{`CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;`},
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
					Query:    "DROP FUNCTION interpreted_example(input TEXT);",
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
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
						{"pg_catalog.interpreted_example(text)", 1},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"ancestor", nil, "deleted", "theirs", "modified", "root_object"},
					},
				},
				{
					Query:    `UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'ours';`,
					Expected: []sql.Row{},
				},
				{ // Updating to "ours" will delete the conflicts table since we're keeping our original root object
					Query:       `DELETE FROM "dolt_conflicts_interpreted_example(text)";`,
					ExpectedErr: `table not found`,
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"212"},
					},
				},
			},
		},
		{
			Name:        `Function deleted "theirs" updated "ours", chose "theirs"`,
			Skip:        true, // TODO: there's a bug dealing with root objects being incorrectly treated as tables
			SetUpScript: []string{`CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;`},
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
					Query:    "DROP FUNCTION interpreted_example(input TEXT);",
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
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
						{"pg_catalog.interpreted_example(text)", 1},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"ancestor", nil, "deleted", "theirs", "modified", "root_object"},
					},
				},
				{
					Query:    `UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'theirs';`,
					Expected: []sql.Row{},
				},
				{ // Updating to "theirs" will delete the conflicts table since we're choosing the deleted root object
					Query:       `DELETE FROM "dolt_conflicts_interpreted_example(text)";`,
					ExpectedErr: `table not found`,
				},
				{
					Query:       "SELECT interpreted_example('12');",
					ExpectedErr: `'interpreted_example' not found`,
				},
			},
		},
		{
			Name:        `Function deleted "theirs" updated "ours", chose "ancestor"`,
			Skip:        true, // TODO: there's a bug dealing with root objects being incorrectly treated as tables
			SetUpScript: []string{`CREATE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '1' || input; END; $$ LANGUAGE plpgsql;`},
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
					Query:    "DROP FUNCTION interpreted_example(input TEXT);",
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS TEXT AS $$ BEGIN RETURN '2' || input; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
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
						{"pg_catalog.interpreted_example(text)", 1},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"ancestor", nil, "deleted", "theirs", "modified", "root_object"},
					},
				},
				{
					Query:    `UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = 'ancestor';`,
					Expected: []sql.Row{},
				},
				{ // Updating to "ancestor" will delete the conflicts table since we effectively rolled back the merge
					Query:       `DELETE FROM "dolt_conflicts_interpreted_example(text)";`,
					ExpectedErr: `table not found`,
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{"112"},
					},
				},
			},
		},
		{
			Name:        `Function update multiple conflicts`,
			SetUpScript: []string{`CREATE FUNCTION interpreted_example(input TEXT) RETURNS INT4 AS $$ BEGIN RETURN input || '1'; END; $$ LANGUAGE plpgsql;`},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS INT8 AS $$ BEGIN RETURN input || '3'; END; $$ LANGUAGE plpgsql;",
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS FLOAT AS $$ BEGIN RETURN input || '2'; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
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
						{"pg_catalog.interpreted_example(text)", 2},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"int4", "float8", "modified", "int8", "modified", "return_type"},
						{"BEGIN RETURN input || '1'; END;", "BEGIN RETURN input || '2'; END;", "modified", "BEGIN RETURN input || '3'; END;", "modified", "definition"},
					},
				},
				{
					Query:    `UPDATE "dolt_conflicts_interpreted_example(text)" SET our_value = their_value;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM "dolt_conflicts_interpreted_example(text)";`,
					ExpectedErr: `table not found`,
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{123},
					},
				},
				{
					Query: "SELECT interpreted_example('123456789012');",
					Expected: []sql.Row{
						{1234567890123},
					},
				},
			},
		},
		{
			Name:        `Function delete multiple conflicts`,
			SetUpScript: []string{`CREATE FUNCTION interpreted_example(input TEXT) RETURNS INT4 AS $$ BEGIN RETURN input || '1'; END; $$ LANGUAGE plpgsql;`},
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS INT8 AS $$ BEGIN RETURN input || '3'; END; $$ LANGUAGE plpgsql;",
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
					Query:    "CREATE OR REPLACE FUNCTION interpreted_example(input TEXT) RETURNS FLOAT AS $$ BEGIN RETURN input || '2'; END; $$ LANGUAGE plpgsql;",
					Expected: []sql.Row{},
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
						{"pg_catalog.interpreted_example(text)", 2},
					},
				},
				{
					Query: `SELECT base_value, our_value, our_diff_type, their_value, their_diff_type, dolt_conflict_id FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{
						{"int4", "float8", "modified", "int8", "modified", "return_type"},
						{"BEGIN RETURN input || '1'; END;", "BEGIN RETURN input || '2'; END;", "modified", "BEGIN RETURN input || '3'; END;", "modified", "definition"},
					},
				},
				{
					Query:    `DELETE FROM "dolt_conflicts_interpreted_example(text)";`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SELECT * FROM "dolt_conflicts_interpreted_example(text)";`,
					ExpectedErr: `table not found`,
				},
				{
					Query: "SELECT interpreted_example('12');",
					Expected: []sql.Row{
						{122.0},
					},
				},
			},
		},
	})
}
