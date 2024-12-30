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
			Skip: true, // hard-coded to use tab separated files right now
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 varchar(250));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: fmt.Sprintf("COPY tbl1 FROM '%s' (FORMAT CSV)", filepath.Join(absTestDataDir, "csv-load-basic-cases.sql")),
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
					Query:             fmt.Sprintf("COPY test_info FROM '%s' WITH (HEADER)", filepath.Join(absTestDataDir, "tab-load-with-header.sql")),
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
	})
}
