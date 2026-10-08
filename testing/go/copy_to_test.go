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

func TestCopyTo(t *testing.T) {
	// The setup script used by most of the tests below, exercising NULLs, empty strings, and values that need
	// escaping or quoting in the output.
	setup := []string{
		"CREATE TABLE tbl1 (pk int primary key, c1 varchar(100), c2 int);",
		`INSERT INTO tbl1 VALUES (1, 'foo', 10), (2, NULL, 20), (3, 'back\slash', NULL), (4, '', 40),` +
			` (5, 'a,b', 50), (6, 'say "hi"', 60), (7, E'multi\nline', 70);`,
	}

	RunScripts(t, []ScriptTest{
		{
			Name:        "tab delimited to stdout",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:            "COPY tbl1 TO STDOUT;",
					CopyToStdOutFile: "copy-to-basic.txt",
				},
			},
		},
		{
			Name:        "tab delimited with header to stdout",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:            "COPY tbl1 TO STDOUT WITH (HEADER);",
					CopyToStdOutFile: "copy-to-header.txt",
				},
			},
		},
		{
			Name:        "tab delimited with column names",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:            "COPY tbl1 (c1, pk) TO STDOUT;",
					CopyToStdOutFile: "copy-to-columns.txt",
				},
			},
		},
		{
			Name:        "csv with header to stdout",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:            "COPY tbl1 TO STDOUT (FORMAT CSV, HEADER);",
					CopyToStdOutFile: "copy-to-basic.csv",
				},
				{
					Query:            "COPY tbl1 TO STDOUT (FORMAT 'csv', HEADER);",
					CopyToStdOutFile: "copy-to-basic.csv",
				},
			},
		},
		{
			Name:        "psv with header to stdout",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:            "COPY tbl1 TO STDOUT (FORMAT CSV, HEADER, DELIMITER '|');",
					CopyToStdOutFile: "copy-to-header.psv",
				},
			},
		},
		{
			Name:        "query to stdout",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:            "COPY (SELECT pk, c1 FROM tbl1 WHERE pk < 3 ORDER BY pk DESC) TO STDOUT;",
					CopyToStdOutFile: "copy-to-query.txt",
				},
				{
					Query:            "COPY (SELECT 1 AS x UNION ALL SELECT 2 ORDER BY x) TO STDOUT;",
					CopyToStdOutFile: "copy-to-union.txt",
				},
				{
					Query:            "COPY (SELECT pk FROM tbl1 WHERE pk > 100) TO STDOUT;",
					CopyToStdOutFile: "copy-to-empty.txt",
				},
			},
		},
		{
			Name: "boolean values to stdout",
			SetUpScript: []string{
				"CREATE TABLE tbl2 (pk int primary key, b boolean);",
				"INSERT INTO tbl2 VALUES (1, true), (2, false), (3, NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "COPY tbl2 TO STDOUT;",
					CopyToStdOutFile: "copy-to-bool.txt",
				},
			},
		},
		{
			Name: "schema qualified table to stdout",
			SetUpScript: []string{
				"CREATE SCHEMA s1;",
				"CREATE TABLE s1.tbl2 (pk int primary key, b boolean);",
				"INSERT INTO s1.tbl2 VALUES (1, true), (2, false), (3, NULL);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "COPY s1.tbl2 TO STDOUT;",
					CopyToStdOutFile: "copy-to-bool.txt",
				},
			},
		},
		{
			Name: "binary to stdout",
			SetUpScript: []string{
				"CREATE TABLE tbl3 (pk int primary key, c1 text, b boolean);",
				`INSERT INTO tbl3 VALUES (1, 'foo', true), (2, NULL, false), (3, '', NULL), (4, 'héllo', true);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "COPY tbl3 TO STDOUT (FORMAT BINARY);",
					CopyToStdOutFile: "copy-to-basic.bin",
				},
				{
					Query:            "COPY tbl3 TO STDOUT BINARY;",
					CopyToStdOutFile: "copy-to-basic.bin",
				},
				{
					Query:            "COPY (SELECT * FROM tbl3 ORDER BY pk) TO STDOUT (FORMAT BINARY);",
					CopyToStdOutFile: "copy-to-basic.bin",
				},
				{
					// DuckDB's postgres extension sends the format name as a quoted identifier
					Query:            `COPY (SELECT * FROM tbl3 ORDER BY pk) TO STDOUT (FORMAT "binary");`,
					CopyToStdOutFile: "copy-to-basic.bin",
				},
			},
		},
		{
			Name:        "csv round trip",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:            "CREATE TABLE tbl1_copy (pk int primary key, c1 varchar(100), c2 int);",
					SkipResultsCheck: true,
				},
				{
					Query:                   "COPY tbl1 TO STDOUT (FORMAT CSV, HEADER);",
					CopyRoundTripStdInQuery: "COPY tbl1_copy FROM STDIN (FORMAT CSV, HEADER);",
				},
				{
					Query: "SELECT count(*) FROM tbl1 t1 JOIN tbl1_copy t2 ON t1.pk = t2.pk WHERE t1.c1 IS NOT DISTINCT FROM t2.c1 AND t1.c2 IS NOT DISTINCT FROM t2.c2;",
					Expected: []sql.Row{
						{7},
					},
				},
			},
		},
		{
			Name:        "tab delimited round trip",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:            "CREATE TABLE tbl1_copy (pk int primary key, c1 varchar(100), c2 int);",
					SkipResultsCheck: true,
				},
				{
					Query:                   "COPY tbl1 TO STDOUT;",
					CopyRoundTripStdInQuery: "COPY tbl1_copy FROM STDIN;",
				},
				{
					Query: "SELECT count(*) FROM tbl1 t1 JOIN tbl1_copy t2 ON t1.pk = t2.pk WHERE t1.c1 IS NOT DISTINCT FROM t2.c1 AND t1.c2 IS NOT DISTINCT FROM t2.c2;",
					Expected: []sql.Row{
						{7},
					},
				},
			},
		},
		{
			Name: "tab delimited escape characters round trip",
			SetUpScript: []string{
				"CREATE TABLE esc (pk int primary key, c1 text);",
				`INSERT INTO esc VALUES
					(1, E'tab\tseparated'),
					(2, E'new\nline'),
					(3, E'carriage\rreturn'),
					(4, E'back\\slash'),
					(5, E'\b\f\v'),
					(6, E'\\N'),
					(7, E'ends with backslash\\'),
					(8, E'\\.'),
					(9, E'mixed\t\\.\nescapes\\\r'),
					(10, 'pipe|delimiter');`,
				"CREATE TABLE esc_copy (pk int primary key, c1 text);",
				"CREATE TABLE esc_copy2 (pk int primary key, c1 text);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:                   "COPY esc TO STDOUT;",
					CopyRoundTripStdInQuery: "COPY esc_copy FROM STDIN;",
				},
				{
					Query: "SELECT count(*) FROM esc t1 JOIN esc_copy t2 ON t1.pk = t2.pk WHERE t1.c1 = t2.c1;",
					Expected: []sql.Row{
						{10},
					},
				},
				{
					// A custom delimiter that appears inside a value must be escaped and unescaped correctly
					Query:                   "COPY esc TO STDOUT (FORMAT TEXT, DELIMITER '|');",
					CopyRoundTripStdInQuery: "COPY esc_copy2 FROM STDIN (FORMAT TEXT, DELIMITER '|');",
				},
				{
					Query: "SELECT count(*) FROM esc t1 JOIN esc_copy2 t2 ON t1.pk = t2.pk WHERE t1.c1 = t2.c1;",
					Expected: []sql.Row{
						{10},
					},
				},
			},
		},
		{
			Name: "single column round trip with NULL",
			SetUpScript: []string{
				"CREATE TABLE single (c1 text);",
				`INSERT INTO single VALUES ('foo'), (NULL), ('');`,
				"CREATE TABLE single_copy_text (c1 text);",
				"CREATE TABLE single_copy_csv (c1 text);",
				"CREATE TABLE single_copy_bin (c1 text);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:                   "COPY single TO STDOUT;",
					CopyRoundTripStdInQuery: "COPY single_copy_text FROM STDIN;",
				},
				{
					Query: "SELECT count(*), count(c1), count(CASE WHEN c1 = '' THEN 1 END) FROM single_copy_text;",
					Expected: []sql.Row{
						{3, 2, 1},
					},
				},
				{
					Query:                   "COPY single TO STDOUT (FORMAT CSV);",
					CopyRoundTripStdInQuery: "COPY single_copy_csv FROM STDIN (FORMAT CSV);",
				},
				{
					Query: "SELECT count(*), count(c1), count(CASE WHEN c1 = '' THEN 1 END) FROM single_copy_csv;",
					Expected: []sql.Row{
						{3, 2, 1},
					},
				},
				{
					Query:                   "COPY single TO STDOUT (FORMAT BINARY);",
					CopyRoundTripStdInQuery: "COPY single_copy_bin FROM STDIN (FORMAT BINARY);",
				},
				{
					Query: "SELECT count(*), count(c1), count(CASE WHEN c1 = '' THEN 1 END) FROM single_copy_bin;",
					Expected: []sql.Row{
						{3, 2, 1},
					},
				},
			},
		},
		{
			Name:        "binary round trip",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:            "CREATE TABLE tbl1_copy (pk int primary key, c1 varchar(100), c2 int);",
					SkipResultsCheck: true,
				},
				{
					// Unlike the text format, the binary format round-trips embedded newlines, so all rows survive
					Query:                   "COPY tbl1 TO STDOUT (FORMAT BINARY);",
					CopyRoundTripStdInQuery: "COPY tbl1_copy FROM STDIN (FORMAT BINARY);",
				},
				{
					Query: "SELECT count(*) FROM tbl1 t1 JOIN tbl1_copy t2 ON t1.pk = t2.pk WHERE t1.c1 IS NOT DISTINCT FROM t2.c1 AND t1.c2 IS NOT DISTINCT FROM t2.c2;",
					Expected: []sql.Row{
						{7},
					},
				},
			},
		},
		{
			Name: "binary round trip over multiple chunks",
			SetUpScript: []string{
				"CREATE TABLE big (pk int primary key, c1 text);",
				// Roughly 220KB of data, so that the client splits the COPY FROM STDIN stream into multiple
				// CopyData chunks and tuples land across chunk boundaries
				"INSERT INTO big SELECT i, repeat('x', 200) || i FROM generate_series(1, 1000) g(i);",
				"CREATE TABLE big_copy (pk int primary key, c1 text);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:                   "COPY big TO STDOUT (FORMAT BINARY);",
					CopyRoundTripStdInQuery: "COPY big_copy FROM STDIN (FORMAT BINARY);",
				},
				{
					Query: "SELECT count(*) FROM big t1 JOIN big_copy t2 ON t1.pk = t2.pk WHERE t1.c1 = t2.c1;",
					Expected: []sql.Row{
						{1000},
					},
				},
			},
		},
		{
			Name: "binary round trip with various types",
			SetUpScript: []string{
				`CREATE TABLE typed (pk int primary key, i2 int2, i8 int8, f4 float4, f8 float8, n numeric(10,2),
					d date, ts timestamp, u uuid, by bytea, b boolean);`,
				`INSERT INTO typed VALUES
					(1, 32767, 9223372036854775807, 1.5, -2.25, 12345.67, '2025-01-01', '2025-01-01 12:34:56',
					 '1077f506-a6fc-4cb2-aed2-9dea9351ed9c', '\xdeadbeef', true),
					(2, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL),
					(3, -32768, -9223372036854775808, 0.0, 1e10, -0.01, '1999-12-31', '1999-12-31 23:59:59',
					 '428d0815-d95b-4cfc-89af-9fca38585dcc', '\x00', false);`,
				`CREATE TABLE typed_copy (pk int primary key, i2 int2, i8 int8, f4 float4, f8 float8, n numeric(10,2),
					d date, ts timestamp, u uuid, by bytea, b boolean);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:                   "COPY typed TO STDOUT (FORMAT BINARY);",
					CopyRoundTripStdInQuery: "COPY typed_copy FROM STDIN (FORMAT BINARY);",
				},
				{
					Query: `SELECT count(*) FROM typed t1 JOIN typed_copy t2 ON t1.pk = t2.pk
						WHERE t1.i2 IS NOT DISTINCT FROM t2.i2 AND t1.i8 IS NOT DISTINCT FROM t2.i8
						AND t1.f4 IS NOT DISTINCT FROM t2.f4 AND t1.f8 IS NOT DISTINCT FROM t2.f8
						AND t1.n IS NOT DISTINCT FROM t2.n AND t1.d IS NOT DISTINCT FROM t2.d
						AND t1.ts IS NOT DISTINCT FROM t2.ts AND t1.u IS NOT DISTINCT FROM t2.u
						AND t1.by IS NOT DISTINCT FROM t2.by AND t1.b IS NOT DISTINCT FROM t2.b;`,
					Expected: []sql.Row{
						{3},
					},
				},
			},
		},
		{
			Name:        "errors",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:       "COPY tbl2 TO STDOUT;",
					ExpectedErr: "table not found: tbl2",
				},
				{
					Query:       "COPY tbl1 (pk, c3) TO STDOUT;",
					ExpectedErr: "column \"c3\" could not be found",
				},
				{
					Query:       "COPY tbl1 TO STDOUT (FORMAT BINARY, HEADER);",
					ExpectedErr: "cannot specify HEADER in BINARY mode",
				},
				{
					Query:       "COPY tbl1 TO STDOUT (FORMAT BINARY, DELIMITER '|');",
					ExpectedErr: "cannot specify DELIMITER in BINARY mode",
				},
				{
					Query:       `COPY tbl1 TO STDOUT (FORMAT "nonsense");`,
					ExpectedErr: `COPY format "nonsense" not recognized`,
				},
				{
					// Writing files on the server is deliberately unsupported, as a security measure
					Query:       "COPY tbl1 TO '/tmp/copy-to-out.csv' (FORMAT CSV);",
					ExpectedErr: "COPY TO a server-side file is not supported",
				},
			},
		},
	})
}
