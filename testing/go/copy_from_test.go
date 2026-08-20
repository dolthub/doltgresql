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
			Name: "timestamp columns",
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk timestamp primary key, ts timestamp);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:             "COPY tbl1 FROM STDIN WITH (HEADER)",
					CopyFromStdInFile: "tab-load-with-timestamp-col.sql",
				},
				{
					Query: "select * from tbl1 order by pk;",
					Expected: []sql.Row{
						{"2020-12-19 19:00:00", "2021-04-04 20:00:00"},
						{"2020-12-19 21:36:32.188", "2020-12-19 19:00:00"},
						{"2021-04-04 20:00:00", "2020-12-19 21:36:32.188"},
					},
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
					Query:             "COPY uuid_table (id, name, second_uuid) FROM STDIN",
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
			Name: "binary from stdin",
			SetUpScript: []string{
				"CREATE TABLE tbl3 (pk int primary key, c1 text, b boolean);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:             "COPY tbl3 FROM STDIN (FORMAT BINARY);",
					CopyFromStdInFile: "copy-to-basic.bin",
				},
				{
					Query: "SELECT * FROM tbl3 ORDER BY pk;",
					Expected: []sql.Row{
						{1, "foo", "t"},
						{2, nil, "f"},
						{3, "", nil},
						{4, "héllo", "t"},
					},
				},
			},
		},
		{
			Name: "binary load multiple chunks",
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk int primary key, c1 text);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// binary-load-multi-chunk.bin is ~230KB, so the client splits it into multiple CopyData
					// chunks and tuples land across chunk boundaries. It holds 2000 rows of (pk, c1) where c1
					// is 'x' repeated (pk % 211) times, except that every 100th row is NULL.
					Query:             "COPY tbl1 FROM STDIN (FORMAT BINARY);",
					CopyFromStdInFile: "binary-load-multi-chunk.bin",
				},
				{
					Query: "SELECT count(*), count(c1), sum(length(c1)) FROM tbl1;",
					Expected: []sql.Row{
						{2000, 1980, 202536},
					},
				},
				{
					// pk = 211 is an empty string, which must stay distinct from NULL
					Query: "SELECT pk, length(c1) FROM tbl1 WHERE pk IN (99, 211, 300, 1999) ORDER BY pk;",
					Expected: []sql.Row{
						{99, 99},
						{211, 0},
						{300, nil},
						{1999, 100},
					},
				},
			},
		},
		{
			Name: "binary from file",
			SetUpScript: []string{
				"CREATE TABLE tbl3 (pk int primary key, c1 text, b boolean);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       fmt.Sprintf("COPY tbl3 FROM '%s' (FORMAT BINARY);", filepath.Join(absTestDataDir, "copy-to-basic.bin")),
					ExpectedTag: "COPY 4",
				},
				{
					Query: "SELECT * FROM tbl3 ORDER BY pk;",
					Expected: []sql.Row{
						{1, "foo", "t"},
						{2, nil, "f"},
						{3, "", nil},
						{4, "héllo", "t"},
					},
				},
			},
		},
		{
			Name: "malformed binary load does not poison the session",
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk int primary key, c1 text);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// binary-load-malformed.bin holds 575 valid rows, then a malformed tuple placed exactly at
					// the client's CopyData chunk boundary, then 5 more valid rows that arrive in a later chunk.
					// The bad load must be rejected, all of its work rolled back, and the trailing chunk discarded.
					Query:             "COPY tbl1 FROM STDIN (FORMAT BINARY);",
					CopyFromStdInFile: "binary-load-malformed.bin",
					ExpectedErr:       "row field count 5, expected 2",
				},
				{
					Query: "SELECT count(*) FROM tbl1;",
					Expected: []sql.Row{
						{0},
					},
				},
				{
					// The same connection stays usable for regular statements
					Query:    "INSERT INTO tbl1 VALUES (100, 'still works');",
					Expected: []sql.Row{},
				},
				{
					// And for a subsequent, valid COPY FROM
					Query:             "COPY tbl1 FROM STDIN (FORMAT BINARY);",
					CopyFromStdInFile: "binary-load-2col.bin",
				},
				{
					Query: "SELECT * FROM tbl1 ORDER BY pk;",
					Expected: []sql.Row{
						{1, "one"},
						{2, nil},
						{3, "three"},
						{100, "still works"},
					},
				},
			},
		},
		{
			Name: "binary load failing after a successful chunk does not poison the session",
			SetUpScript: []string{
				"CREATE TABLE tbl1 (pk int primary key, c1 text);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// binary-load-malformed-late.bin fills the client's first CopyData chunk with 575 valid rows,
					// with the malformed tuple arriving in the second chunk. The rows loaded by the first chunk
					// must be rolled back along with the rest of the failed operation.
					Query:             "COPY tbl1 FROM STDIN (FORMAT BINARY);",
					CopyFromStdInFile: "binary-load-malformed-late.bin",
					ExpectedErr:       "row field count 5, expected 2",
				},
				{
					Query: "SELECT count(*) FROM tbl1;",
					Expected: []sql.Row{
						{0},
					},
				},
				{
					Query:             "COPY tbl1 FROM STDIN (FORMAT BINARY);",
					CopyFromStdInFile: "binary-load-2col.bin",
				},
				{
					Query: "SELECT count(*) FROM tbl1;",
					Expected: []sql.Row{
						{3},
					},
				},
			},
		},
		{
			Name: "binary load missing its trailer does not poison the session",
			SetUpScript: []string{
				"CREATE TABLE tbl3 (pk int primary key, c1 text, b boolean);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// The data is valid except that it ends without the file trailer, so the error only
					// surfaces when the load is finalized. Its rows must still be rolled back.
					Query:             "COPY tbl3 FROM STDIN (FORMAT BINARY);",
					CopyFromStdInFile: "copy-from-missing-trailer.bin",
					ExpectedErr:       "missing file trailer",
				},
				{
					Query: "SELECT count(*) FROM tbl3;",
					Expected: []sql.Row{
						{0},
					},
				},
				{
					Query:             "COPY tbl3 FROM STDIN (FORMAT BINARY);",
					CopyFromStdInFile: "copy-to-basic.bin",
				},
				{
					Query: "SELECT * FROM tbl3 ORDER BY pk;",
					Expected: []sql.Row{
						{1, "foo", "t"},
						{2, nil, "f"},
						{3, "", nil},
						{4, "héllo", "t"},
					},
				},
			},
		},
		{
			Name: "binary errors",
			SetUpScript: []string{
				"CREATE TABLE tbl3 (pk int primary key, c1 text, b boolean);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "COPY tbl3 FROM STDIN (FORMAT BINARY, HEADER);",
					ExpectedErr: "cannot specify HEADER in BINARY mode",
				},
				{
					Query:       "COPY tbl3 FROM STDIN (FORMAT BINARY, DELIMITER '|');",
					ExpectedErr: "cannot specify DELIMITER in BINARY mode",
				},
				{
					// A text file is not valid binary COPY input
					Query:       fmt.Sprintf("COPY tbl3 FROM '%s' (FORMAT BINARY);", filepath.Join(absTestDataDir, "copy-to-basic.txt")),
					ExpectedErr: "COPY file signature not recognized",
				},
				{
					// Binary data that ends without the file trailer indicates truncation
					Query:       fmt.Sprintf("COPY tbl3 FROM '%s' (FORMAT BINARY);", filepath.Join(absTestDataDir, "copy-from-missing-trailer.bin")),
					ExpectedErr: "missing file trailer",
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
