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
