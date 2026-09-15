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
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/require"
)

// extendedCopy runs one COPY FROM STDIN exchange through Parse, Bind, Execute, and Sync.
type extendedCopy struct {
	query       string
	input       CopyInput
	expectedTag string
	expectedErr string
}

// fatalCopyMessage sends an illegal frontend message during COPY and verifies fatal protocol shutdown.
type fatalCopyMessage struct {
	extended   bool
	unexpected pgproto3.FrontendMessage
}

// describe returns a concise label for the fatal COPY protocol step.
func (s fatalCopyMessage) describe() string {
	return fmt.Sprintf("Fatal COPY message %T", s.unexpected)
}

// runStep verifies PostgreSQL-compatible error ordering and connection closure for an illegal COPY message.
func (s fatalCopyMessage) runStep(r *messageFlowRunner) {
	require.Empty(r.t, r.pending)
	if s.extended {
		r.send(&pgproto3.Parse{Query: "COPY test3 FROM STDIN"})
		require.IsType(r.t, &pgproto3.ParseComplete{}, r.receiveNext())
		r.send(&pgproto3.Bind{})
		require.IsType(r.t, &pgproto3.BindComplete{}, r.receiveNext())
		r.send(&pgproto3.Execute{})
	} else {
		r.send(&pgproto3.Query{String: "COPY test3 FROM STDIN"})
	}
	require.IsType(r.t, &pgproto3.CopyInResponse{}, r.receiveNext())
	r.send(&pgproto3.CopyData{Data: []byte("1\n")})
	r.send(s.unexpected)

	encoded, err := s.unexpected.Encode(nil)
	require.NoError(r.t, err)
	require.NotEmpty(r.t, encoded)
	first, ok := r.receiveNext().(*pgproto3.ErrorResponse)
	require.True(r.t, ok, "expected ERROR response first")
	require.Equal(r.t, "ERROR", first.Severity)
	require.Equal(r.t, "08P01", first.Code)
	require.Equal(r.t, fmt.Sprintf("unexpected message type 0x%02x during COPY from stdin", encoded[0]), first.Message)
	second, ok := r.receiveNext().(*pgproto3.ErrorResponse)
	require.True(r.t, ok, "expected FATAL response second")
	require.Equal(r.t, "FATAL", second.Severity)
	require.Equal(r.t, "08P01", second.Code)
	require.Equal(r.t, "terminating connection because protocol synchronization was lost", second.Message)
	message, err := r.flowConn.Receive(r.t)
	require.Error(r.t, err)
	require.Nil(r.t, message)
}

// describe returns a concise label for the extended COPY step.
func (s extendedCopy) describe() string {
	return "Extended COPY " + s.query
}

// runStep executes and validates an extended-protocol COPY exchange.
func (s extendedCopy) runStep(r *messageFlowRunner) {
	require.Empty(r.t, r.pending)
	r.send(&pgproto3.Parse{Query: s.query})
	require.IsType(r.t, &pgproto3.ParseComplete{}, r.receiveNext())
	r.send(&pgproto3.Bind{})
	require.IsType(r.t, &pgproto3.BindComplete{}, r.receiveNext())
	r.send(&pgproto3.Execute{})
	require.IsType(r.t, &pgproto3.CopyInResponse{}, r.receiveNext())
	for _, message := range s.input.BeforeData {
		r.send(message)
	}
	for _, chunk := range s.input.Chunks {
		r.send(&pgproto3.CopyData{Data: chunk})
	}
	if s.input.FailMessage == "" {
		r.send(&pgproto3.CopyDone{})
	} else {
		r.send(&pgproto3.CopyFail{Message: s.input.FailMessage})
	}
	result := r.receiveNext()
	if s.expectedErr != "" {
		errResponse, ok := result.(*pgproto3.ErrorResponse)
		require.True(r.t, ok, "expected ErrorResponse, received %T", result)
		require.Equal(r.t, s.expectedErr, errResponse.Message)
		require.Equal(r.t, "ERROR", errResponse.Severity)
		require.Equal(r.t, "57014", errResponse.Code)
	} else {
		complete, ok := result.(*pgproto3.CommandComplete)
		require.True(r.t, ok, "expected CommandComplete, received %T", result)
		require.Equal(r.t, s.expectedTag, string(complete.CommandTag))
	}
	r.send(&pgproto3.Sync{})
	ready, ok := r.receiveNext().(*pgproto3.ReadyForQuery)
	require.True(r.t, ok, "expected ReadyForQuery after Sync")
	require.Equal(r.t, byte('I'), ready.TxStatus)
}

// TestCopyFromStdinInMultiStatementSimpleQuery verifies that COPY pauses and resumes a compound simple query.
func TestCopyFromStdinInMultiStatementSimpleQuery(t *testing.T) {
	RunMessageFlowTests(t, []MessageFlowTest{
		{
			Name:        "empty copy initializes and completes the transfer",
			SetUpScript: []string{"CREATE TABLE test3 (c int);"},
			Steps: []FlowStep{
				SimpleQuery{
					Query:      "COPY test3 FROM STDIN; SELECT count(*) FROM test3;",
					CopyInputs: []CopyInput{{}},
					Expected: []StatementResult{
						{Tag: "COPY 0"},
						{Tag: "SELECT 1", Rows: [][]string{{"0"}}},
					},
				},
			},
		},
		{
			Name:        "copy fail before data aborts cleanly",
			SetUpScript: []string{"CREATE TABLE test3 (c int);"},
			Steps: []FlowStep{
				SimpleQuery{
					Query:               "COPY test3 FROM STDIN; SELECT 1;",
					CopyInputs:          []CopyInput{{FailMessage: "client aborted empty copy"}},
					ExpectedErr:         "client aborted empty copy",
					ExpectedErrExact:    "COPY from stdin failed: client aborted empty copy",
					ExpectedErrCode:     "57014",
					ExpectedErrSeverity: "ERROR",
				},
				QueryOnOtherConnection{Query: "SELECT * FROM test3;", Expected: nil},
			},
		},
		{
			Name:        "flush and sync are ignored during copy",
			SetUpScript: []string{"CREATE TABLE test3 (c int);"},
			Steps: []FlowStep{
				SimpleQuery{
					Query: "COPY test3 FROM STDIN; SELECT count(*) FROM test3;",
					CopyInputs: []CopyInput{{
						BeforeData: []pgproto3.FrontendMessage{&pgproto3.Flush{}, &pgproto3.Sync{}},
						Chunks:     [][]byte{[]byte("1\n")},
					}},
					Expected: []StatementResult{
						{Tag: "COPY 1"},
						{Tag: "SELECT 1", Rows: [][]string{{"1"}}},
					},
				},
			},
		},
		{
			Name:        "multiple copy inputs preserve statement order",
			SetUpScript: []string{"CREATE TABLE test3 (c int);"},
			Steps: []FlowStep{
				SimpleQuery{
					Query: "SELECT 0; COPY test3 FROM STDIN; COPY test3 FROM STDIN; SELECT 1;",
					CopyInputs: []CopyInput{
						{Chunks: [][]byte{[]byte("1\n")}},
						{Chunks: [][]byte{[]byte("2\n")}},
					},
					Expected: []StatementResult{
						{Tag: "SELECT 1", Rows: [][]string{{"0"}}},
						{Tag: "COPY 1"},
						{Tag: "COPY 1"},
						{Tag: "SELECT 1", Rows: [][]string{{"1"}}},
					},
				},
				SimpleQuery{
					Query:    "SELECT * FROM test3 ORDER BY c;",
					Expected: []StatementResult{{Tag: "SELECT 2", Rows: [][]string{{"1"}, {"2"}}}},
				},
				SimpleQuery{
					Query:    "DROP TABLE test3;",
					Expected: []StatementResult{{Tag: "DROP TABLE"}},
				},
			},
		},
		{
			Name:        "copy first rolls back when a later statement fails",
			SetUpScript: []string{"CREATE TABLE test3 (c int);"},
			Steps: []FlowStep{
				SimpleQuery{
					Query:       "COPY test3 FROM STDIN; SELECT * FROM missing_table;",
					CopyInputs:  []CopyInput{{Chunks: [][]byte{[]byte("1\n")}}},
					Expected:    []StatementResult{{Tag: "COPY 1"}},
					ExpectedErr: "missing_table",
				},
				QueryOnOtherConnection{Query: "SELECT * FROM test3;", Expected: nil},
			},
		},
		{
			Name:        "copy fail rolls back compound query",
			SetUpScript: []string{"CREATE TABLE test3 (c int);"},
			Steps: []FlowStep{
				SimpleQuery{
					Query: "INSERT INTO test3 VALUES (0); COPY test3 FROM STDIN; SELECT 1;",
					CopyInputs: []CopyInput{{
						Chunks:      [][]byte{[]byte("1\n"), []byte("2\n")},
						FailMessage: "client aborted copy",
					}},
					Expected:        []StatementResult{{Tag: "INSERT 0 1"}},
					ExpectedErr:     "client aborted copy",
					ExpectedErrCode: "57014",
				},
				QueryOnOtherConnection{Query: "SELECT * FROM test3;", Expected: nil},
			},
		},
		{
			Name:        "copy fail marks explicit transaction failed",
			SetUpScript: []string{"CREATE TABLE test3 (c int);"},
			Steps: []FlowStep{
				SimpleQuery{
					Query:               "BEGIN; COPY test3 FROM STDIN; SELECT 1;",
					CopyInputs:          []CopyInput{{Chunks: [][]byte{[]byte("1\n")}, FailMessage: "client aborted copy"}},
					Expected:            []StatementResult{{Tag: "BEGIN"}},
					ExpectedErr:         "client aborted copy",
					ExpectedErrCode:     "57014",
					ExpectedReadyStatus: 'E',
				},
				SimpleQuery{Query: "ROLLBACK;", Expected: []StatementResult{{Tag: "ROLLBACK"}}},
				QueryOnOtherConnection{Query: "SELECT * FROM test3;", Expected: nil},
			},
		},
		{
			Name:        "explicit transaction commits copy and surrounding statements",
			SetUpScript: []string{"CREATE TABLE test3 (c int);"},
			Steps: []FlowStep{
				SimpleQuery{
					Query:      "BEGIN; INSERT INTO test3 VALUES (0); COPY test3 FROM STDIN; INSERT INTO test3 VALUES (2); COMMIT;",
					CopyInputs: []CopyInput{{Chunks: [][]byte{[]byte("1\n")}}},
					Expected: []StatementResult{
						{Tag: "BEGIN"},
						{Tag: "INSERT 0 1"},
						{Tag: "COPY 1"},
						{Tag: "INSERT 0 1"},
						{Tag: "COMMIT"},
					},
				},
				QueryOnOtherConnection{Query: "SELECT * FROM test3 ORDER BY c;", Expected: [][]string{{"0"}, {"1"}, {"2"}}},
			},
		},
		{
			Name:        "explicit transaction rolls back copy and surrounding statements",
			SetUpScript: []string{"CREATE TABLE test3 (c int);"},
			Steps: []FlowStep{
				SimpleQuery{
					Query:      "BEGIN; INSERT INTO test3 VALUES (0); COPY test3 FROM STDIN; INSERT INTO test3 VALUES (2); ROLLBACK;",
					CopyInputs: []CopyInput{{Chunks: [][]byte{[]byte("1\n")}}},
					Expected: []StatementResult{
						{Tag: "BEGIN"},
						{Tag: "INSERT 0 1"},
						{Tag: "COPY 1"},
						{Tag: "INSERT 0 1"},
						{Tag: "ROLLBACK"},
					},
				},
				QueryOnOtherConnection{Query: "SELECT * FROM test3;", Expected: nil},
			},
		},
		{
			Name:        "copy compound query continues an existing explicit transaction",
			SetUpScript: []string{"CREATE TABLE test3 (c int);"},
			Steps: []FlowStep{
				SimpleQuery{Query: "BEGIN;", Expected: []StatementResult{{Tag: "BEGIN"}}, ExpectedReadyStatus: 'T'},
				SimpleQuery{
					Query:               "COPY test3 FROM STDIN; INSERT INTO test3 VALUES (2);",
					CopyInputs:          []CopyInput{{Chunks: [][]byte{[]byte("1\n")}}},
					Expected:            []StatementResult{{Tag: "COPY 1"}, {Tag: "INSERT 0 1"}},
					ExpectedReadyStatus: 'T',
				},
				QueryOnOtherConnection{Query: "SELECT * FROM test3;", Expected: nil},
				SimpleQuery{Query: "ROLLBACK;", Expected: []StatementResult{{Tag: "ROLLBACK"}}},
				QueryOnOtherConnection{Query: "SELECT * FROM test3;", Expected: nil},
			},
		},
	})
}

// TestCopyFromStdinExtendedProtocol verifies COPY success and failure return to the initiating extended batch.
func TestCopyFromStdinExtendedProtocol(t *testing.T) {
	RunMessageFlowTests(t, []MessageFlowTest{
		{
			Name:        "extended copy completes at sync",
			SetUpScript: []string{"CREATE TABLE test3 (c int);"},
			Steps: []FlowStep{
				extendedCopy{query: "COPY test3 FROM STDIN", input: CopyInput{Chunks: [][]byte{[]byte("1\n")}}, expectedTag: "COPY 1"},
				QueryOnOtherConnection{Query: "SELECT * FROM test3;", Expected: [][]string{{"1"}}},
			},
		},
		{
			Name:        "extended copy failure discards until sync",
			SetUpScript: []string{"CREATE TABLE test3 (c int);"},
			Steps: []FlowStep{
				extendedCopy{query: "COPY test3 FROM STDIN", input: CopyInput{FailMessage: "abort extended copy"}, expectedErr: "COPY from stdin failed: abort extended copy"},
				QueryOnOtherConnection{Query: "SELECT * FROM test3;", Expected: nil},
			},
		},
	})
}

// TestCopyUnexpectedMessageFatal verifies illegal COPY messages fatally desynchronize simple and extended origins.
func TestCopyUnexpectedMessageFatal(t *testing.T) {
	for _, extended := range []bool{false, true} {
		origin := "simple"
		if extended {
			origin = "extended"
		}
		for _, unexpected := range []pgproto3.FrontendMessage{
			&pgproto3.Query{String: "SELECT 2"},
			&pgproto3.Parse{Query: "SELECT 2"},
		} {
			RunMessageFlowTest(t, MessageFlowTest{
				Name:        fmt.Sprintf("%s origin rejects %T", origin, unexpected),
				SetUpScript: []string{"CREATE TABLE test3 (c int);"},
				Steps: []FlowStep{
					fatalCopyMessage{extended: extended, unexpected: unexpected},
					QueryOnOtherConnection{Query: "SELECT * FROM test3;", Expected: nil},
				},
			})
		}
	}
}

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
