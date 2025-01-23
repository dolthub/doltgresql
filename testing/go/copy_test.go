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
	"fmt"
	"path/filepath"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	absTestDataDir, err := filepath.Abs("testdata")
	require.NoError(t, err)

	RunScripts(t, []ScriptTest{
		{
			Name: "tab delimited with header",
			SetUpScript: []string{
				"CREATE TABLE test (pk int primary key);",
				"INSERT INTO test VALUES (0), (1);",
				"CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:             "COPY test_info FROM STDIN WITH (HEADER);",
					CopyFromStdInFile: "tab-load-with-header.sql",
				},
				{
					Query: "SELECT * FROM test_info order by 1;",
					Expected: []sql.Row{
						{4, "string for 4", 1},
						{5, "string for 5", 0},
						{6, "string for 6", 0},
					},
				},
			},
		},
		{
			Name: "tab delimited with header and column names",
			SetUpScript: []string{
				"CREATE TABLE test (pk int primary key);",
				"INSERT INTO test VALUES (0), (1);",
				"CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:             "COPY test_info (id, info, test_pk) FROM STDIN WITH (HEADER);",
					CopyFromStdInFile: "tab-load-with-header.sql",
				},
				{
					Query: "SELECT * FROM test_info order by 1;",
					Expected: []sql.Row{
						{4, "string for 4", 1},
						{5, "string for 5", 0},
						{6, "string for 6", 0},
					},
				},
			},
		},
		{
			Name: "tab delimited with quoted column names",
			SetUpScript: []string{
				`CREATE TABLE Regions (
   "Id" SERIAL UNIQUE NOT NULL,
   "Code" VARCHAR(4) UNIQUE NOT NULL,
   "Capital" VARCHAR(10) NOT NULL,
   "Name" VARCHAR(255) UNIQUE NOT NULL
);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:             "COPY regions (\"Id\", \"Code\", \"Capital\", \"Name\") FROM stdin;\n",
					CopyFromStdInFile: "tab-load-with-quoted-column-names.sql",
				},
			},
		},
		{
			Name: "basic csv",
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:             "COPY tbl1 FROM STDIN (FORMAT CSV)",
					CopyFromStdInFile: "csv-load-basic-cases.sql",
				},
				{
					Query: "select * from tbl1 where pk = 6 order by pk;",
					Expected: []sql.Row{
						{6, `foo
\\.
bar`, "baz"},
					},
				},
				{
					Query: "select * from tbl1 where pk = 9;",
					Expected: []sql.Row{
						{9, nil, "''"},
					},
				},
			},
		},
		{
			Name: "csv with header",
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:             " COPY tbl1 FROM STDIN (FORMAT CSV, HEADER TRUE);",
					CopyFromStdInFile: "csv-load-with-header.sql",
				},
				{
					Query: "select * from tbl1 where pk = 6 order by pk;",
					Expected: []sql.Row{
						{6, `foo
\\.
bar`, "baz"},
					},
				},
			},
		},
		{
			Name: "generated column",
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250), c3 int generated always as (pk + 10) stored);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:             "COPY tbl1 (pk, c1, c2) FROM STDIN (FORMAT CSV)",
					CopyFromStdInFile: "csv-load-basic-cases.sql",
				},
				{
					Query: "select * from tbl1 where pk = 6 order by pk;",
					Expected: []sql.Row{
						{6, `foo
\\.
bar`, "baz", 16},
					},
				},
				{
					Query: "select * from tbl1 where pk = 9;",
					Expected: []sql.Row{
						{9, nil, "''", 19},
					},
				},
			},
		},
		{
			Name: "load multiple chunks",
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:             "COPY tbl1 FROM STDIN (FORMAT CSV);",
					CopyFromStdInFile: "csv-load-multi-chunk.sql",
				},
				{
					Query: "select * from tbl1 where pk = 99 order by pk;",
					Expected: []sql.Row{
						{99, "foo", "barbazbashbarbazbashbarbazbashbarbazbashbarbazbashbarbazbashbarbazbashbarbazbashbarbazbashbarbazbashbarbazbash"},
					},
				},
			},
		},
		{
			Name: "load psv with headers",
			SetUpScript: []string{
				"CREATE TABLE test (pk int primary key);",
				"INSERT INTO test VALUES (0), (1);",
				"CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:             "COPY test_info FROM STDIN (FORMAT CSV, HEADER TRUE, DELIMITER '|');",
					CopyFromStdInFile: "psv-load.sql",
				},
				{
					Query: "SELECT * FROM test_info order by 1;",
					Expected: []sql.Row{
						{4, "string for 4", 1},
						{5, "string for 5", 0},
						{6, "string for 6", 0},
					},
				},
			},
		},
		{
			Name: "csv from file",
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            fmt.Sprintf("COPY tbl1 FROM '%s' (FORMAT CSV)", filepath.Join(absTestDataDir, "csv-load-basic-cases.sql")),
					SkipResultsCheck: true,
				},
				{
					Query: "select * from tbl1 where pk = 6 order by pk;",
					Expected: []sql.Row{
						{6, `foo
\\.
bar`, "baz"},
					},
				},
				{
					Query: "select * from tbl1 where pk = 9;",
					Expected: []sql.Row{
						{9, nil, "''"},
					},
				},
			},
		},
		{
			Name: "csv from file with column names",
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            fmt.Sprintf("COPY tbl1 (pk, c1, c2) FROM '%s' (FORMAT CSV)", filepath.Join(absTestDataDir, "csv-load-basic-cases.sql")),
					SkipResultsCheck: true,
				},
				{
					Query: "select * from tbl1 where pk = 6 order by pk;",
					Expected: []sql.Row{
						{6, `foo
\\.
bar`, "baz"},
					},
				},
				{
					Query: "select * from tbl1 where pk = 9;",
					Expected: []sql.Row{
						{9, nil, "''"},
					},
				},
			},
		},
		{
			Name: "tab delimited with header from file",
			SetUpScript: []string{
				"CREATE TABLE test (pk int primary key);",
				"INSERT INTO test VALUES (0), (1);",
				"CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: fmt.Sprintf("COPY test_info FROM '%s' WITH (HEADER)", filepath.Join(absTestDataDir, "tab-load-with-header.sql")),
				},
				{
					Query: "SELECT * FROM test_info order by 1;",
					Expected: []sql.Row{
						{4, "string for 4", 1},
						{5, "string for 5", 0},
						{6, "string for 6", 0},
					},
				},
			},
		},
		{
			Name: "tab delimited with uuid values",
			SetUpScript: []string{
				`CREATE TABLE public.uuid_table (
    id uuid NOT NULL,
    name character varying NOT NULL,
    second_uuid uuid DEFAULT '428d0815-d95b-4cfc-89af-9fca38585dcc'::uuid);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:             fmt.Sprintf("COPY uuid_table (id, name, second_uuid) FROM STDIN"),
					CopyFromStdInFile: "uuid-table.sql",
				},
				{
					Query: "SELECT * FROM uuid_table order by id;",
					Expected: []sql.Row{
						{"1077f506-a6fc-4cb2-aed2-9dea9351ed9c", "Company A", "428d0815-d95b-4cfc-89af-9fca38585dcc"},
						{"5e080b3a-361f-4e16-b7a4-70d4f175e283", "Company B", "428d0815-d95b-4cfc-89af-9fca38585dcc"},
					},
				},
			},
		},
		{
			Name: "file not found",
			SetUpScript: []string{
				"CREATE TABLE test (pk int primary key);",
				"INSERT INTO test VALUES (0), (1);",
				"CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       fmt.Sprintf("COPY test_info FROM '%s' WITH (HEADER)", filepath.Join(absTestDataDir, "file-not-found.sql")),
					ExpectedErr: "file", // exact error message varies by platform
				},
			},
		},
		{
			Name: "wrong columns",
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       fmt.Sprintf("COPY tbl1 (pk, c1) FROM '%s' (FORMAT CSV)", filepath.Join(absTestDataDir, "csv-load-basic-cases.sql")),
					ExpectedErr: "extra data after last expected column",
				},
				{
					Query:       fmt.Sprintf("COPY tbl1 (pk, c1, c3) FROM '%s' (FORMAT CSV)", filepath.Join(absTestDataDir, "csv-load-basic-cases.sql")),
					ExpectedErr: "Unknown column",
				},
			},
		},
		{
			Name: "table not found",
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       fmt.Sprintf("COPY tbl2 (pk, c1) FROM '%s' (FORMAT CSV)", filepath.Join(absTestDataDir, "csv-load-basic-cases.sql")),
					ExpectedErr: "table not found: tbl2",
				},
			},
		},
		{
			Name: "read only table",
			Assertions: []ScriptTestAssertion{
				{
					Query:       fmt.Sprintf("COPY dolt_log FROM '%s' (FORMAT CSV)", filepath.Join(absTestDataDir, "csv-load-basic-cases.sql")),
					ExpectedErr: "table doesn't support INSERT INTO",
				},
			},
		},
		{
			Name: "bad data rows",
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       fmt.Sprintf("COPY tbl1 (pk, c1, c2) FROM '%s' (FORMAT CSV)", filepath.Join(absTestDataDir, "missing-columns.sql")),
					ExpectedErr: "record on line 2: wrong number of fields",
				},
				{
					Query:    "select count(*) from tbl1;",
					Expected: []sql.Row{{0}},
				},
				{
					Query:       fmt.Sprintf("COPY tbl1 (pk, c1, c2) FROM '%s' (FORMAT CSV)", filepath.Join(absTestDataDir, "too-many-columns.sql")),
					ExpectedErr: "record on line 6: wrong number of fields",
				},
				{
					Query:    "select count(*) from tbl1;",
					Expected: []sql.Row{{0}},
				},
				{
					Query:       fmt.Sprintf("COPY tbl1 (pk, c1, c2) FROM '%s' (FORMAT CSV)", filepath.Join(absTestDataDir, "wrong-types.sql")),
					ExpectedErr: "invalid input syntax for type int4",
				},
				{
					Query:    "select count(*) from tbl1;",
					Expected: []sql.Row{{0}},
				},
			},
		},
	})
}
