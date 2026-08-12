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
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file tests Postgres's implicit transaction behavior, which depends on the exact sequence of protocol
// messages exchanged with the server. The full ruleset is documented in
// https://www.postgresql.org/docs/current/protocol-flow.html, and is summarized as:
//
// Simple query protocol:
//   - All statements in a single Query message run in a single implicit transaction block, which is committed
//     after the last statement, or rolled back if any statement fails. Statements after a failing statement are
//     never executed.
//   - Explicit transaction control statements within the message change this: COMMIT / ROLLBACK close the
//     current (implicit or explicit) transaction block, and any following statements run in a new implicit
//     transaction block. A BEGIN within an implicit transaction block converts it into an explicit (regular)
//     transaction block, retroactively including the statements already executed, and the block remains open
//     after the end of the message until an explicit COMMIT / ROLLBACK.
//   - If the session is already inside a transaction block, the statements of a Query message simply continue it.
//
// Extended query protocol:
//   - All statements executed between Sync messages run in a single implicit transaction block, which Sync
//     closes: committed if everything succeeded, rolled back if any message failed. After an error the server
//     skips all messages until Sync. A Flush delivers pending responses but does not close the transaction.
//   - Sync does NOT close an explicit transaction block opened with BEGIN; the ReadyForQuery response carries
//     the transaction status ('I' idle, 'T' in transaction, 'E' in failed transaction).

// TestImplicitTransactionsSimpleProtocol tests the implicit transaction behavior of multi-statement Query
// messages in the simple query protocol.
func TestImplicitTransactionsSimpleProtocol(t *testing.T) {
	setup := []string{"CREATE TABLE mytable (i BIGINT PRIMARY KEY);"}
	RunMessageFlowTests(t, []MessageFlowTest{
		{
			Name:        "multiple statements commit as a single implicit transaction",
			SetUpScript: setup,
			Steps: []FlowStep{
				SimpleQuery{
					Query: "INSERT INTO mytable VALUES (1); INSERT INTO mytable VALUES (2); SELECT * FROM mytable ORDER BY i;",
					Expected: []StatementResult{
						{Tag: "INSERT 0 1"},
						{Tag: "INSERT 0 1"},
						{Tag: "SELECT 2", Rows: [][]string{{"1"}, {"2"}}},
					},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}, {"2"}},
				},
			},
		},
		{
			Name:        "error rolls back the entire implicit transaction",
			SetUpScript: setup,
			Steps: []FlowStep{
				SimpleQuery{
					// The error rolls back the first INSERT, and the second INSERT is never executed
					Query:       "INSERT INTO mytable VALUES (1); SELECT 1/0; INSERT INTO mytable VALUES (2);",
					Expected:    []StatementResult{{Tag: "INSERT 0 1"}},
					ExpectedErr: "division by zero",
				},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
				// The implicit transaction was rolled back at the error, so the session returns to idle (not a
				// failed transaction block) and normal processing resumes with the next Query message
				SimpleQuery{
					Query:    "INSERT INTO mytable VALUES (3);",
					Expected: []StatementResult{{Tag: "INSERT 0 1"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"3"}},
				},
			},
		},
		{
			Name:        "explicit COMMIT inside the message commits preceding statements only",
			SetUpScript: setup,
			Steps: []FlowStep{
				// This is the example from the Postgres docs: the first INSERT is committed by the explicit
				// COMMIT, while the second INSERT runs in a new implicit transaction block that is rolled back
				// by the divide-by-zero error
				SimpleQuery{
					Query: "BEGIN; INSERT INTO mytable VALUES (1); COMMIT; INSERT INTO mytable VALUES (2); SELECT 1/0;",
					Expected: []StatementResult{
						{Tag: "BEGIN"},
						{Tag: "INSERT 0 1"},
						{Tag: "COMMIT"},
						{Tag: "INSERT 0 1"},
					},
					ExpectedErr: "division by zero",
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
			},
		},
		{
			Name:        "COMMIT closes an implicit transaction block and starts a new one",
			SetUpScript: setup,
			Steps: []FlowStep{
				// A COMMIT without a preceding BEGIN closes the implicit transaction block (with a warning,
				// which we don't check), committing the first INSERT. The remaining statements run in a new
				// implicit transaction block, which the error rolls back.
				SimpleQuery{
					Query: "INSERT INTO mytable VALUES (1); COMMIT; INSERT INTO mytable VALUES (2); SELECT 1/0;",
					Expected: []StatementResult{
						{Tag: "INSERT 0 1"},
						{Tag: "COMMIT"},
						{Tag: "INSERT 0 1"},
					},
					ExpectedErr: "division by zero",
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
			},
		},
		{
			Name:        "ROLLBACK closes an implicit transaction block and discards its work",
			SetUpScript: setup,
			Steps: []FlowStep{
				// The ROLLBACK discards the first INSERT; the second INSERT runs in a new implicit transaction
				// block that commits at the end of the message
				SimpleQuery{
					Query: "INSERT INTO mytable VALUES (1); ROLLBACK; INSERT INTO mytable VALUES (2);",
					Expected: []StatementResult{
						{Tag: "INSERT 0 1"},
						{Tag: "ROLLBACK"},
						{Tag: "INSERT 0 1"},
					},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"2"}},
				},
			},
		},
		{
			Name:        "BEGIN converts an implicit transaction block into an explicit one",
			SetUpScript: setup,
			Steps: []FlowStep{
				// The BEGIN retroactively includes the first INSERT in the new explicit transaction block, which
				// remains open after the end of the message
				SimpleQuery{
					Query: "INSERT INTO mytable VALUES (1); BEGIN; INSERT INTO mytable VALUES (2);",
					Expected: []StatementResult{
						{Tag: "INSERT 0 1"},
						{Tag: "BEGIN"},
						{Tag: "INSERT 0 1"},
					},
					ExpectedReadyStatus: 'T',
				},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
				SimpleQuery{
					Query:    "ROLLBACK;",
					Expected: []StatementResult{{Tag: "ROLLBACK"}},
				},
				// Both INSERTs are discarded, including the one executed before the BEGIN
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
			},
		},
		{
			Name:        "transaction block left open by a Query message continues across messages",
			SetUpScript: setup,
			Steps: []FlowStep{
				SimpleQuery{
					Query:               "BEGIN; INSERT INTO mytable VALUES (1);",
					Expected:            []StatementResult{{Tag: "BEGIN"}, {Tag: "INSERT 0 1"}},
					ExpectedReadyStatus: 'T',
				},
				// The insert is visible within the transaction, but not to other connections
				SimpleQuery{
					Query:               "SELECT * FROM mytable ORDER BY i;",
					Expected:            []StatementResult{{Tag: "SELECT 1", Rows: [][]string{{"1"}}}},
					ExpectedReadyStatus: 'T',
				},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
				SimpleQuery{
					Query:               "INSERT INTO mytable VALUES (2);",
					Expected:            []StatementResult{{Tag: "INSERT 0 1"}},
					ExpectedReadyStatus: 'T',
				},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
				SimpleQuery{
					Query:    "COMMIT;",
					Expected: []StatementResult{{Tag: "COMMIT"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}, {"2"}},
				},
			},
		},
		{
			Name:        "COMMIT of a block opened in an earlier message starts an implicit block for remaining statements",
			SetUpScript: setup,
			Steps: []FlowStep{
				SimpleQuery{
					Query:               "BEGIN;",
					Expected:            []StatementResult{{Tag: "BEGIN"}},
					ExpectedReadyStatus: 'T',
				},
				// The first INSERT continues the open transaction block and is committed by the COMMIT; the
				// second INSERT runs in a new implicit transaction block that the error rolls back
				SimpleQuery{
					Query: "INSERT INTO mytable VALUES (1); COMMIT; INSERT INTO mytable VALUES (2); SELECT 1/0;",
					Expected: []StatementResult{
						{Tag: "INSERT 0 1"},
						{Tag: "COMMIT"},
						{Tag: "INSERT 0 1"},
					},
					ExpectedErr: "division by zero",
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
			},
		},
		{
			Name:        "error in an explicit transaction block leaves the session in a failed transaction",
			SetUpScript: setup,
			Steps: []FlowStep{
				// Unlike an implicit transaction block, an explicit one is not rolled back by an error: the
				// session is left in a failed transaction block, and the ROLLBACK after the error is never
				// executed
				SimpleQuery{
					Query:               "BEGIN; SELECT 1/0; ROLLBACK;",
					Expected:            []StatementResult{{Tag: "BEGIN"}},
					ExpectedErr:         "division by zero",
					ExpectedReadyStatus: 'E',
				},
				SimpleQuery{
					Query:               "SELECT 1;",
					ExpectedErr:         "current transaction is aborted",
					ExpectedReadyStatus: 'E',
				},
				SimpleQuery{
					Query:    "ROLLBACK;",
					Expected: []StatementResult{{Tag: "ROLLBACK"}},
				},
			},
		},
		{
			Name:        "syntax error anywhere in the message prevents any statement from executing",
			SetUpScript: setup,
			Steps: []FlowStep{
				// The entire query string is parsed before any of it executes, so the misspelled SELCT in the
				// last statement prevents even the first (explicitly committed) INSERT from running
				SimpleQuery{
					Query:       "BEGIN; INSERT INTO mytable VALUES (1); COMMIT; INSERT INTO mytable VALUES (2); SELCT 1/0;",
					ExpectedErr: "syntax error",
				},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
			},
		},
		{
			Name:        "ROLLBACK TO SAVEPOINT recovers a failed transaction block",
			SetUpScript: setup,
			Steps: []FlowStep{
				SimpleQuery{
					Query:               "BEGIN; INSERT INTO mytable VALUES (1); SAVEPOINT sp1;",
					Expected:            []StatementResult{{Tag: "BEGIN"}, {Tag: "INSERT 0 1"}, {Tag: "SAVEPOINT"}},
					ExpectedReadyStatus: 'T',
				},
				SimpleQuery{
					Query:               "SELECT 1/0;",
					ExpectedErr:         "division by zero",
					ExpectedReadyStatus: 'E',
				},
				// Rolling back to the savepoint recovers the failed transaction block, and the work done before
				// the savepoint is preserved
				SimpleQuery{
					Query:               "ROLLBACK TO sp1;",
					Expected:            []StatementResult{{Tag: "ROLLBACK"}},
					ExpectedReadyStatus: 'T',
				},
				SimpleQuery{
					Query:               "SELECT * FROM mytable ORDER BY i;",
					Expected:            []StatementResult{{Tag: "SELECT 1", Rows: [][]string{{"1"}}}},
					ExpectedReadyStatus: 'T',
				},
				SimpleQuery{
					Query:    "COMMIT;",
					Expected: []StatementResult{{Tag: "COMMIT"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
			},
		},
		{
			Name:        "savepoints are not allowed in an implicit transaction block",
			SetUpScript: setup,
			Steps: []FlowStep{
				SimpleQuery{
					Query:       "INSERT INTO mytable VALUES (1); SAVEPOINT sp1;",
					Expected:    []StatementResult{{Tag: "INSERT 0 1"}},
					ExpectedErr: "SAVEPOINT can only be used in transaction blocks",
				},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
			},
		},
		{
			Name:        "error on the final statement rolls back the entire implicit transaction",
			SetUpScript: setup,
			Steps: []FlowStep{
				SimpleQuery{
					Query:       "INSERT INTO mytable VALUES (1); INSERT INTO mytable VALUES (2); SELECT 1/0;",
					Expected:    []StatementResult{{Tag: "INSERT 0 1"}, {Tag: "INSERT 0 1"}},
					ExpectedErr: "division by zero",
				},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
			},
		},
		{
			Name:        "error part way through the final statement's results rolls back the entire implicit transaction",
			SetUpScript: setup,
			Steps: []FlowStep{
				// Unlike SELECT 1/0, which fails before returning any rows, this SELECT scans the table and
				// fails part way through its result set, exercising the engine's streaming result path
				SimpleQuery{
					Query:       "INSERT INTO mytable VALUES (1); INSERT INTO mytable VALUES (2); SELECT i / (i - 2) FROM mytable ORDER BY i;",
					Expected:    []StatementResult{{Tag: "INSERT 0 1"}, {Tag: "INSERT 0 1"}},
					ExpectedErr: "division by zero",
				},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
			},
		},
		{
			Name:        "explicit COMMIT as the final statement commits the implicit transaction",
			SetUpScript: setup,
			Steps: []FlowStep{
				SimpleQuery{
					Query: "INSERT INTO mytable VALUES (1); INSERT INTO mytable VALUES (2); COMMIT;",
					Expected: []StatementResult{
						{Tag: "INSERT 0 1"},
						{Tag: "INSERT 0 1"},
						{Tag: "COMMIT"},
					},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}, {"2"}},
				},
			},
		},
		{
			Name:        "explicit ROLLBACK as the final statement discards the implicit transaction",
			SetUpScript: setup,
			Steps: []FlowStep{
				SimpleQuery{
					Query: "INSERT INTO mytable VALUES (1); INSERT INTO mytable VALUES (2); ROLLBACK;",
					Expected: []StatementResult{
						{Tag: "INSERT 0 1"},
						{Tag: "INSERT 0 1"},
						{Tag: "ROLLBACK"},
					},
				},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
			},
		},
		{
			Name:        "error on the final statement of an explicit transaction block preserves the block",
			SetUpScript: setup,
			Steps: []FlowStep{
				SimpleQuery{
					Query:               "BEGIN; INSERT INTO mytable VALUES (1); SAVEPOINT sp1; SELECT 1/0;",
					Expected:            []StatementResult{{Tag: "BEGIN"}, {Tag: "INSERT 0 1"}, {Tag: "SAVEPOINT"}},
					ExpectedErr:         "division by zero",
					ExpectedReadyStatus: 'E',
				},
				// The error must not roll back the explicit transaction block: rolling back to the savepoint
				// recovers it, with the work done before the savepoint intact
				SimpleQuery{
					Query:               "ROLLBACK TO sp1;",
					Expected:            []StatementResult{{Tag: "ROLLBACK"}},
					ExpectedReadyStatus: 'T',
				},
				SimpleQuery{
					Query:               "SELECT * FROM mytable ORDER BY i;",
					Expected:            []StatementResult{{Tag: "SELECT 1", Rows: [][]string{{"1"}}}},
					ExpectedReadyStatus: 'T',
				},
				SimpleQuery{
					Query:    "COMMIT;",
					Expected: []StatementResult{{Tag: "COMMIT"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
			},
		},
		{
			Name:        "statement handled outside the engine as the final statement still commits the block",
			SetUpScript: setup,
			Steps: []FlowStep{
				// DEALLOCATE is handled by the connection handler directly rather than being passed to the
				// engine, so the implicit transaction block must be committed by the end of the message instead
				// of by the final statement's execution
				SimpleQuery{
					Query:    "INSERT INTO mytable VALUES (1); DEALLOCATE ALL;",
					Expected: []StatementResult{{Tag: "INSERT 0 1"}, {}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
			},
		},
		{
			Name:        "errors do not poison the session outside of explicit transaction blocks",
			SetUpScript: setup,
			Steps: []FlowStep{
				// A parse error in a single-statement message
				SimpleQuery{
					Query:       "SELCT 1;",
					ExpectedErr: "syntax error",
				},
				SimpleQuery{
					Query:    "INSERT INTO mytable VALUES (1);",
					Expected: []StatementResult{{Tag: "INSERT 0 1"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
				// An analysis error in a single-statement message
				SimpleQuery{
					Query:       "SELECT * FROM doesnotexist;",
					ExpectedErr: "table not found",
				},
				SimpleQuery{
					Query:    "INSERT INTO mytable VALUES (2);",
					Expected: []StatementResult{{Tag: "INSERT 0 1"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}, {"2"}},
				},
				// Several errors in a row, of different kinds
				SimpleQuery{
					Query:       "SELECT 1/0;",
					ExpectedErr: "division by zero",
				},
				SimpleQuery{
					Query:       "INSERT INTO mytable VALUES (1);",
					ExpectedErr: "duplicate primary key",
				},
				SimpleQuery{
					Query:       "SELCT 1;",
					ExpectedErr: "syntax error",
				},
				SimpleQuery{
					Query:    "INSERT INTO mytable VALUES (3);",
					Expected: []StatementResult{{Tag: "INSERT 0 1"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}, {"2"}, {"3"}},
				},
				// An error in a multi-statement message
				SimpleQuery{
					Query:       "INSERT INTO mytable VALUES (4); SELECT 1/0;",
					Expected:    []StatementResult{{Tag: "INSERT 0 1"}},
					ExpectedErr: "division by zero",
				},
				SimpleQuery{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: []StatementResult{{Tag: "SELECT 3", Rows: [][]string{{"1"}, {"2"}, {"3"}}}},
				},
			},
		},
		{
			Name:        "errors in a transaction block do not poison the session after the block is ended",
			SetUpScript: setup,
			Steps: []FlowStep{
				// This is the shape of many Postgres regression test scripts: a transaction block around a few
				// statements, ended by a ROLLBACK. An error inside the block fails the block, but the ROLLBACK
				// must fully restore the session for the statements that follow.
				SimpleQuery{
					Query:               "begin;",
					Expected:            []StatementResult{{Tag: "BEGIN"}},
					ExpectedReadyStatus: 'T',
				},
				SimpleQuery{
					Query:               "SELECT * FROM doesnotexist;",
					ExpectedErr:         "table not found",
					ExpectedReadyStatus: 'E',
				},
				SimpleQuery{
					Query:               "SELECT 1;",
					ExpectedErr:         "current transaction is aborted",
					ExpectedReadyStatus: 'E',
				},
				SimpleQuery{
					Query:    "rollback;",
					Expected: []StatementResult{{Tag: "ROLLBACK"}},
				},
				SimpleQuery{
					Query:    "INSERT INTO mytable VALUES (1);",
					Expected: []StatementResult{{Tag: "INSERT 0 1"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
				// The same, but opened with BEGIN WORK and ended with COMMIT (which rolls back)
				SimpleQuery{
					Query:               "begin work;",
					Expected:            []StatementResult{{Tag: "BEGIN"}},
					ExpectedReadyStatus: 'T',
				},
				SimpleQuery{
					Query:               "SELECT 1/0;",
					ExpectedErr:         "division by zero",
					ExpectedReadyStatus: 'E',
				},
				SimpleQuery{
					Query:    "commit;",
					Expected: []StatementResult{{Tag: "ROLLBACK"}},
				},
				SimpleQuery{
					Query:    "INSERT INTO mytable VALUES (2);",
					Expected: []StatementResult{{Tag: "INSERT 0 1"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}, {"2"}},
				},
			},
		},
		{
			Name:        "single-statement Query messages commit immediately and are visible to other connections",
			SetUpScript: setup,
			Steps: []FlowStep{
				SimpleQuery{
					Query:    "INSERT INTO mytable VALUES (1);",
					Expected: []StatementResult{{Tag: "INSERT 0 1"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
				SimpleQuery{
					Query:    "UPDATE mytable SET i = 2;",
					Expected: []StatementResult{{Tag: "UPDATE 1"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"2"}},
				},
				SimpleQuery{
					Query:    "DELETE FROM mytable;",
					Expected: []StatementResult{{Tag: "DELETE 1"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
			},
		},
	})
}

// TestImplicitTransactionsExtendedProtocol tests the implicit transaction behavior of the extended query
// protocol, where all statements executed between Sync messages share an implicit transaction.
func TestImplicitTransactionsExtendedProtocol(t *testing.T) {
	setup := []string{"CREATE TABLE mytable (i BIGINT PRIMARY KEY);"}
	RunMessageFlowTests(t, []MessageFlowTest{
		{
			Name:        "statements in a batch commit as a single implicit transaction at Sync",
			SetUpScript: setup,
			Steps: []FlowStep{
				Parse{Name: "ins1", Query: "INSERT INTO mytable VALUES (1)"},
				Bind{PreparedStatement: "ins1"},
				Execute{Tag: "INSERT 0 1"},
				Parse{Name: "ins2", Query: "INSERT INTO mytable VALUES (2)"},
				Bind{PreparedStatement: "ins2"},
				Execute{Tag: "INSERT 0 1"},
				Parse{Name: "sel", Query: "SELECT * FROM mytable ORDER BY i"},
				Describe{ObjectType: 'S', Name: "sel"},
				Bind{PreparedStatement: "sel"},
				Execute{Tag: "SELECT 2", Rows: [][]string{{"1"}, {"2"}}},
				Sync{},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}, {"2"}},
				},
			},
		},
		{
			Name:        "work is not visible to other connections until Sync commits it",
			SetUpScript: setup,
			Steps: []FlowStep{
				Parse{Name: "ins1", Query: "INSERT INTO mytable VALUES (1)"},
				Bind{PreparedStatement: "ins1"},
				Execute{Tag: "INSERT 0 1"},
				// Flush forces the server to deliver the responses above, proving the INSERT has executed, but
				// unlike Sync it does not close the implicit transaction
				Flush{},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
				Sync{},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
			},
		},
		{
			Name:        "error rolls back the batch at Sync and later messages are skipped",
			SetUpScript: setup,
			Steps: []FlowStep{
				Parse{Name: "ins1", Query: "INSERT INTO mytable VALUES (1)"},
				Bind{PreparedStatement: "ins1"},
				Execute{Tag: "INSERT 0 1"},
				Parse{Name: "boom", Query: "SELECT 1/0"},
				Bind{PreparedStatement: "boom"},
				Execute{ExpectedErr: "division by zero"},
				// After the error, the server discards all messages until the Sync, so these produce no
				// responses and the INSERT of 2 is never executed
				Parse{Name: "ins2", Query: "INSERT INTO mytable VALUES (2)"},
				Bind{PreparedStatement: "ins2"},
				Execute{},
				// Sync rolls back the implicit transaction and returns the session to idle
				Sync{},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
				// Normal processing resumes after the Sync
				Parse{Name: "ins3", Query: "INSERT INTO mytable VALUES (3)"},
				Bind{PreparedStatement: "ins3"},
				Execute{Tag: "INSERT 0 1"},
				Sync{},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"3"}},
				},
			},
		},
		{
			Name:        "parse error rolls back earlier statements in the batch",
			SetUpScript: setup,
			Steps: []FlowStep{
				Parse{Name: "ins1", Query: "INSERT INTO mytable VALUES (1)"},
				Bind{PreparedStatement: "ins1"},
				Execute{Tag: "INSERT 0 1"},
				// An error from any extended-query message (not just Execute) fails the implicit transaction
				Parse{Name: "bad", Query: "SELCT 1", ExpectedErr: "syntax error"},
				Sync{},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
			},
		},
		{
			Name:        "Sync does not close an explicit transaction block",
			SetUpScript: setup,
			Steps: []FlowStep{
				Parse{Name: "begin", Query: "BEGIN"},
				Bind{PreparedStatement: "begin"},
				Execute{Tag: "BEGIN"},
				Sync{ExpectedReadyStatus: 'T'},
				Parse{Name: "ins1", Query: "INSERT INTO mytable VALUES (1)"},
				Bind{PreparedStatement: "ins1"},
				Execute{Tag: "INSERT 0 1"},
				// The transaction was opened explicitly, so Sync leaves it open and uncommitted
				Sync{ExpectedReadyStatus: 'T'},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
				Parse{Name: "commit", Query: "COMMIT"},
				Bind{PreparedStatement: "commit"},
				Execute{Tag: "COMMIT"},
				// Flush is required to avoid a race (execute only sends bytes to the server without waiting on a response)
				Flush{},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
				Sync{},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
			},
		},
		{
			Name:        "error inside an explicit transaction block leaves a failed transaction at Sync",
			SetUpScript: setup,
			Steps: []FlowStep{
				Parse{Name: "begin", Query: "BEGIN"},
				Bind{PreparedStatement: "begin"},
				Execute{Tag: "BEGIN"},
				Parse{Name: "ins1", Query: "INSERT INTO mytable VALUES (1)"},
				Bind{PreparedStatement: "ins1"},
				Execute{Tag: "INSERT 0 1"},
				Parse{Name: "boom", Query: "SELECT 1/0"},
				Bind{PreparedStatement: "boom"},
				Execute{ExpectedErr: "division by zero"},
				// Sync does not roll back the explicit transaction block; the session is left in a failed
				// transaction
				Sync{ExpectedReadyStatus: 'E'},
				Parse{Name: "sel", Query: "SELECT 1", ExpectedErr: "current transaction is aborted"},
				Sync{ExpectedReadyStatus: 'E'},
				Parse{Name: "rb", Query: "ROLLBACK"},
				Bind{PreparedStatement: "rb"},
				Execute{Tag: "ROLLBACK"},
				Sync{},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
			},
		},
		{
			Name:        "error on the final Execute of a batch rolls back the whole batch",
			SetUpScript: setup,
			Steps: []FlowStep{
				Parse{Name: "ins1", Query: "INSERT INTO mytable VALUES (1)"},
				Bind{PreparedStatement: "ins1"},
				Execute{Tag: "INSERT 0 1"},
				Parse{Name: "ins2", Query: "INSERT INTO mytable VALUES (2)"},
				Bind{PreparedStatement: "ins2"},
				Execute{Tag: "INSERT 0 1"},
				Parse{Name: "boom", Query: "SELECT 1/0"},
				Bind{PreparedStatement: "boom"},
				Execute{ExpectedErr: "division by zero"},
				Sync{},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
			},
		},
		{
			Name:        "explicit COMMIT as the final statement of a batch commits before Sync",
			SetUpScript: setup,
			Steps: []FlowStep{
				Parse{Name: "ins1", Query: "INSERT INTO mytable VALUES (1)"},
				Bind{PreparedStatement: "ins1"},
				Execute{Tag: "INSERT 0 1"},
				Parse{Name: "commit", Query: "COMMIT"},
				Bind{PreparedStatement: "commit"},
				Execute{Tag: "COMMIT"},
				// Flush proves the server has executed the messages above; the COMMIT takes effect when it
				// executes, so the work is visible to other connections even before the Sync
				Flush{},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
				Sync{},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
			},
		},
		{
			Name:        "explicit ROLLBACK as the final statement of a batch discards it",
			SetUpScript: setup,
			Steps: []FlowStep{
				Parse{Name: "ins1", Query: "INSERT INTO mytable VALUES (1)"},
				Bind{PreparedStatement: "ins1"},
				Execute{Tag: "INSERT 0 1"},
				Parse{Name: "rb", Query: "ROLLBACK"},
				Bind{PreparedStatement: "rb"},
				Execute{Tag: "ROLLBACK"},
				Sync{},
				QueryOnOtherConnection{
					Query:    "SELECT count(*) FROM mytable;",
					Expected: [][]string{{"0"}},
				},
			},
		},
		{
			Name:        "a batch handled entirely outside the engine does not disable autocommit for later statements",
			SetUpScript: setup,
			// TODO: DEALLOCATE cannot currently be executed via the extended query protocol: the engine errors
			//  at Bind with "Unknown prepared statement handler () given to EXECUTE". This is unrelated to
			//  transaction handling; unskip this test when that is fixed.
			Skip: true,
			Steps: []FlowStep{
				// DEALLOCATE is handled by the connection handler without ever starting an engine transaction,
				// so this batch's implicit transaction block has nothing to commit at Sync
				Parse{Name: "da", Query: "DEALLOCATE ALL"},
				Bind{PreparedStatement: "da"},
				Execute{},
				Sync{},
				// The session must return to normal autocommit behavior afterwards: a single-statement Query
				// message commits immediately and is visible to other connections
				SimpleQuery{
					Query:    "INSERT INTO mytable VALUES (1);",
					Expected: []StatementResult{{Tag: "INSERT 0 1"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
			},
		},
		{
			Name:        "DDL committed by the engine mid-batch does not disable autocommit for later statements",
			SetUpScript: setup,
			Steps: []FlowStep{
				// The engine currently commits DDL statements as soon as they execute (they cannot be part of a
				// transaction block), which ends the batch's implicit transaction early
				Parse{Name: "ct", Query: "CREATE TABLE other (j BIGINT)"},
				Bind{PreparedStatement: "ct"},
				Execute{Tag: "CREATE TABLE"},
				Sync{},
				// The session must return to normal autocommit behavior afterwards: a single-statement Query
				// message commits immediately and is visible to other connections
				SimpleQuery{
					Query:    "INSERT INTO mytable VALUES (1);",
					Expected: []StatementResult{{Tag: "INSERT 0 1"}},
				},
				QueryOnOtherConnection{
					Query:    "SELECT * FROM mytable ORDER BY i;",
					Expected: [][]string{{"1"}},
				},
			},
		},
	})
}

// MessageFlowTest is a test that drives the wire protocol message-by-message, unlike ScriptTest, which is
// limited to the query patterns that pgx generates. This allows testing behavior that depends on the exact
// sequence of protocol messages, such as implicit transaction commit and rollback. The test starts its own
// server, runs SetUpScript over a normal (pgx) connection, and then runs each of the Steps in order: protocol
// steps are exchanged over a dedicated raw wire connection, while QueryOnOtherConnection steps run on the pgx
// connection so that tests can observe which effects have been committed.
type MessageFlowTest struct {
	// Name of the test.
	Name string
	// The SQL statements to execute as setup, in order, over a separate connection. Results are not checked,
	// but statements must not error.
	SetUpScript []string
	// The steps to run, in order. See the FlowStep implementations for the available steps.
	Steps []FlowStep
	// Focus behaves as it does on ScriptTest: when set on one or more tests, only those tests run.
	Focus bool
	// Skip is used to completely skip a test, including its setup.
	Skip bool
}

// FlowStep is a single step in a MessageFlowTest.
type FlowStep interface {
	// runStep executes the step, failing the test if the server's responses don't match the step's
	// expectations.
	runStep(r *messageFlowRunner)
	// describe returns a short human-readable description of the step, used to name its subtest.
	describe() string
}

// StatementResult is the expected result of a single statement within a SimpleQuery message.
type StatementResult struct {
	// Tag is the expected CommandComplete tag (e.g. "INSERT 0 1"). An empty string skips the tag check.
	Tag string
	// Rows are the expected rows, in text (wire) format, with NULL values rendered as "NULL". Nil means the
	// statement returns no rows.
	Rows [][]string
}

// SimpleQuery sends a single simple-protocol Query ('Q') message, which may contain multiple semicolon-separated
// statements, and reads the server's responses through the trailing ReadyForQuery.
type SimpleQuery struct {
	// Query is the full query string to send, possibly containing multiple statements.
	Query string
	// Expected contains the expected result of each statement that completes, in order. When an error is
	// expected, only the statements before the error produce results.
	Expected []StatementResult
	// ExpectedErr, when non-empty, asserts that an ErrorResponse whose message contains this string is received.
	ExpectedErr string
	// ExpectedReadyStatus is the transaction status expected in the trailing ReadyForQuery message: 'I' (idle),
	// 'T' (in transaction block), or 'E' (in failed transaction block). The zero value defaults to 'I'.
	ExpectedReadyStatus byte
}

// Parse sends an extended-protocol Parse ('P') message. Its response is validated by the next Sync or Flush step.
type Parse struct {
	// Name is the destination prepared statement name; empty means the unnamed statement.
	Name string
	// Query is the statement to parse.
	Query string
	// ExpectedErr, when non-empty, asserts that this message draws an ErrorResponse containing this string.
	// After an error, the server skips all messages until Sync, so later steps in the batch must not have
	// expectations of their own.
	ExpectedErr string
}

// Bind sends an extended-protocol Bind ('B') message. Its response is validated by the next Sync or Flush step.
type Bind struct {
	// PreparedStatement is the name of the prepared statement to bind; empty means the unnamed statement.
	PreparedStatement string
	// Portal is the destination portal name; empty means the unnamed portal.
	Portal string
	// Parameters are the parameter values, in text format.
	Parameters []string
	// ExpectedErr, when non-empty, asserts that this message draws an ErrorResponse containing this string.
	ExpectedErr string
}

// Describe sends an extended-protocol Describe ('D') message. Its response is validated (loosely: only the
// response message types are checked) by the next Sync or Flush step.
type Describe struct {
	// ObjectType is 'S' to describe a prepared statement or 'P' to describe a portal. The zero value defaults
	// to 'S'.
	ObjectType byte
	// Name is the name of the prepared statement or portal to describe.
	Name string
	// ExpectedErr, when non-empty, asserts that this message draws an ErrorResponse containing this string.
	ExpectedErr string
}

// Execute sends an extended-protocol Execute ('E') message. Its response is validated by the next Sync or
// Flush step.
type Execute struct {
	// Portal is the name of the portal to execute; empty means the unnamed portal.
	Portal string
	// Tag is the expected CommandComplete tag. An empty string skips the tag check.
	Tag string
	// Rows are the expected rows, in text (wire) format, with NULL values rendered as "NULL". Nil means the
	// statement returns no rows.
	Rows [][]string
	// ExpectedErr, when non-empty, asserts that this message draws an ErrorResponse containing this string.
	ExpectedErr string
}

// Sync sends an extended-protocol Sync ('S') message, then reads and validates the responses to every extended
// protocol step sent since the last Sync or Flush, through the trailing ReadyForQuery. Sync closes an implicit
// transaction: committing it if all preceding steps succeeded, and rolling it back otherwise.
type Sync struct {
	// ExpectedReadyStatus is the transaction status expected in the trailing ReadyForQuery message: 'I' (idle),
	// 'T' (in transaction block), or 'E' (in failed transaction block). The zero value defaults to 'I'.
	ExpectedReadyStatus byte
}

// Flush sends an extended-protocol Flush ('H') message, then reads and validates the responses to every extended
// protocol step sent since the last Sync or Flush. Unlike Sync, Flush does not close the implicit transaction,
// and no ReadyForQuery is sent. Steps validated at a Flush must not expect errors, since error recovery only
// happens at a Sync.
type Flush struct{}

// QueryOnOtherConnection runs a query on a separate connection (in autocommit mode) and checks its results.
// This is how tests observe which effects have been committed by the message-flow connection: work in an open
// or rolled-back transaction is invisible here, while committed work is visible.
type QueryOnOtherConnection struct {
	// Query is the query to run.
	Query string
	// Expected are the expected rows, with each value rendered as a string via fmt.Sprintf("%v") and NULL
	// values rendered as "NULL". Nil means no rows are expected.
	Expected [][]string
}

// RunMessageFlowTests runs the given collection of message flow tests.
func RunMessageFlowTests(t *testing.T, tests []MessageFlowTest) {
	// First, we'll run through the tests to check for the Focus variable. If it's true, then append it to the new slice.
	focusTests := make([]MessageFlowTest, 0, len(tests))
	for _, test := range tests {
		if test.Focus {
			// If this is running in GitHub Actions, then we'll panic, because someone forgot to disable it before committing
			if _, ok := os.LookupEnv("GITHUB_ACTION"); ok {
				panic(fmt.Sprintf("The message flow test `%s` has Focus set to `true`. GitHub Actions requires "+
					"that all tests are run, which Focus circumvents, leading to this error. Please disable Focus "+
					"on all tests.", test.Name))
			}
			focusTests = append(focusTests, test)
		}
	}
	// If we have tests with Focus set, then we replace the normal test slice with the new slice.
	if len(focusTests) > 0 {
		tests = focusTests
	}

	for _, test := range tests {
		RunMessageFlowTest(t, test)
	}
}

// RunMessageFlowTest runs the given message flow test.
func RunMessageFlowTest(t *testing.T, test MessageFlowTest) {
	t.Run(test.Name, func(t *testing.T) {
		if test.Skip {
			t.Skip("Skip has been set in the test")
		}

		port, err := sql.GetEmptyPort()
		require.NoError(t, err)
		ctx, checkConn, controller := CreateServerWithPort(t, "postgres", port)
		defer func() {
			checkConn.Close(ctx)
			controller.Stop()
			require.NoError(t, controller.WaitForStop())
		}()
		for _, query := range test.SetUpScript {
			_, err := checkConn.Exec(ctx, query)
			require.NoError(t, err, "error running setup query: %s", query)
		}

		flowConn := NewRawWireConnection(t, "localhost", port, "", "", 10*time.Second)
		defer flowConn.Close()

		runner := &messageFlowRunner{
			t:         t,
			ctx:       ctx,
			flowConn:  flowConn,
			checkConn: checkConn,
		}
		for i, step := range test.Steps {
			runner.stepIdx = i
			// Each step runs as its own subtest, so that a failure identifies the exact step that failed. Any
			// failure aborts the rest of the flow, since the protocol conversation is in an unknown state and
			// later steps would just fail confusingly.
			if !runner.runSubtest(fmt.Sprintf("step %d %s", i, step.describe()), func() {
				step.runStep(runner)
			}) {
				t.Fatalf("aborting message flow after step %d failed", i)
			}
		}
		require.Empty(t, runner.pending, "test ended with extended-query messages awaiting a Sync or Flush")
	})
}

// messageFlowRunner holds the state needed to run the steps of a MessageFlowTest.
type messageFlowRunner struct {
	t   *testing.T
	ctx context.Context
	// flowConn is the raw wire connection that protocol steps are exchanged over.
	flowConn *RawWireConnection
	// checkConn is a normal (pgx) connection used by QueryOnOtherConnection steps.
	checkConn *Connection
	// stepIdx is the index of the step currently being run, used in failure messages.
	stepIdx int
	// pending contains the extended-protocol steps that have been sent but whose responses have not yet been
	// read, which happens at the next Sync or Flush step.
	pending []pendingFlowStep
}

// pendingFlowStep is an extended-protocol step whose response has not been read yet, along with its step index
// for failure messages.
type pendingFlowStep struct {
	idx  int
	step FlowStep
}

// runSubtest runs the given function as a named subtest of the runner's current test, temporarily pointing the
// runner's assertions at the subtest, and returns whether the subtest passed.
func (r *messageFlowRunner) runSubtest(name string, f func()) bool {
	parent := r.t
	defer func() { r.t = parent }()
	return parent.Run(name, func(t *testing.T) {
		r.t = t
		defer func() { r.t = parent }()
		f()
	})
}

// send sends the given message to the server, failing the test on a connection error.
func (r *messageFlowRunner) send(msg pgproto3.FrontendMessage) {
	require.NoError(r.t, r.flowConn.Send(r.t, msg), "step %d: error sending %T", r.stepIdx, msg)
}

// receiveNext returns the next backend message, skipping asynchronous messages that these tests don't verify
// (e.g. the NoticeResponse warning that Postgres sends for a COMMIT outside of a transaction block).
func (r *messageFlowRunner) receiveNext() pgproto3.BackendMessage {
	for {
		msg, err := r.flowConn.Receive(r.t)
		require.NoError(r.t, err, "step %d: error receiving message from server", r.stepIdx)
		switch msg.(type) {
		case *pgproto3.NoticeResponse, *pgproto3.ParameterStatus, *pgproto3.NotificationResponse:
			continue
		}
		return msg
	}
}

var _ FlowStep = SimpleQuery{}
var _ FlowStep = Parse{}
var _ FlowStep = Bind{}
var _ FlowStep = Describe{}
var _ FlowStep = Execute{}
var _ FlowStep = Sync{}
var _ FlowStep = Flush{}
var _ FlowStep = QueryOnOtherConnection{}

// describe implements FlowStep.
func (s SimpleQuery) describe() string {
	return "Query " + summarizeQuery(s.Query)
}

// describe implements FlowStep.
func (s Parse) describe() string {
	name := s.Name
	if name == "" {
		name = "(unnamed)"
	}
	return fmt.Sprintf("Parse %s %s", name, summarizeQuery(s.Query))
}

// describe implements FlowStep.
func (s Bind) describe() string {
	name := s.PreparedStatement
	if name == "" {
		name = "(unnamed)"
	}
	return "Bind " + name
}

// describe implements FlowStep.
func (s Describe) describe() string {
	objectType := "statement"
	if s.ObjectType == 'P' {
		objectType = "portal"
	}
	name := s.Name
	if name == "" {
		name = "(unnamed)"
	}
	return fmt.Sprintf("Describe %s %s", objectType, name)
}

// describe implements FlowStep.
func (s Execute) describe() string {
	name := s.Portal
	if name == "" {
		name = "(unnamed)"
	}
	return "Execute " + name
}

// describe implements FlowStep.
func (s Sync) describe() string {
	return "Sync"
}

// describe implements FlowStep.
func (s Flush) describe() string {
	return "Flush"
}

// describe implements FlowStep.
func (s QueryOnOtherConnection) describe() string {
	return "OtherConnection " + summarizeQuery(s.Query)
}

// summarizeQuery shortens the given query string for use in a subtest name.
func summarizeQuery(query string) string {
	if len(query) > 60 {
		return query[:57] + "..."
	}
	return query
}

// runStep implements FlowStep.
func (s SimpleQuery) runStep(r *messageFlowRunner) {
	t := r.t
	require.Empty(t, r.pending,
		"step %d: SimpleQuery sent while extended-query messages are awaiting a Sync or Flush step", r.stepIdx)
	r.send(&pgproto3.Query{String: s.Query})

	var results []StatementResult
	var current *StatementResult
	errMsg := ""
	for {
		msg := r.receiveNext()
		switch m := msg.(type) {
		case *pgproto3.RowDescription:
			// Marks the start of a row-returning statement's results; the field descriptions themselves aren't
			// checked by these tests
			if current == nil {
				current = &StatementResult{}
			}
		case *pgproto3.DataRow:
			if current == nil {
				current = &StatementResult{}
			}
			current.Rows = append(current.Rows, textRowValues(m))
		case *pgproto3.CommandComplete:
			if current == nil {
				current = &StatementResult{}
			}
			current.Tag = string(m.CommandTag)
			results = append(results, *current)
			current = nil
		case *pgproto3.EmptyQueryResponse:
			current = nil
		case *pgproto3.ErrorResponse:
			require.Empty(t, errMsg, "step %d: received more than one ErrorResponse: %s, then %s",
				r.stepIdx, errMsg, m.Message)
			errMsg = m.Message
		case *pgproto3.ReadyForQuery:
			if s.ExpectedErr != "" {
				if assert.NotEmpty(t, errMsg, "step %d: expected an error containing %q, but no ErrorResponse "+
					"was received", r.stepIdx, s.ExpectedErr) {
					assert.Contains(t, errMsg, s.ExpectedErr, "step %d: wrong error message", r.stepIdx)
				}
			} else {
				assert.Empty(t, errMsg, "step %d: unexpected ErrorResponse: %s", r.stepIdx, errMsg)
			}
			assertStatementResults(t, r.stepIdx, s.Expected, results)
			assertReadyStatus(t, r.stepIdx, s.ExpectedReadyStatus, m.TxStatus)
			assert.NoError(t, r.flowConn.EmptyReceiveBuffer(), "step %d", r.stepIdx)
			return
		default:
			t.Fatalf("step %d: unexpected message %T while reading simple query results: %v", r.stepIdx, msg, msg)
		}
	}
}

// runStep implements FlowStep.
func (s Parse) runStep(r *messageFlowRunner) {
	r.send(&pgproto3.Parse{Name: s.Name, Query: s.Query})
	r.pending = append(r.pending, pendingFlowStep{idx: r.stepIdx, step: s})
}

// runStep implements FlowStep.
func (s Bind) runStep(r *messageFlowRunner) {
	params := make([][]byte, len(s.Parameters))
	for i, param := range s.Parameters {
		params[i] = []byte(param)
	}
	r.send(&pgproto3.Bind{
		DestinationPortal: s.Portal,
		PreparedStatement: s.PreparedStatement,
		Parameters:        params,
	})
	r.pending = append(r.pending, pendingFlowStep{idx: r.stepIdx, step: s})
}

// runStep implements FlowStep.
func (s Describe) runStep(r *messageFlowRunner) {
	if s.ObjectType == 0 {
		s.ObjectType = 'S'
	}
	r.send(&pgproto3.Describe{ObjectType: s.ObjectType, Name: s.Name})
	r.pending = append(r.pending, pendingFlowStep{idx: r.stepIdx, step: s})
}

// runStep implements FlowStep.
func (s Execute) runStep(r *messageFlowRunner) {
	r.send(&pgproto3.Execute{Portal: s.Portal})
	r.pending = append(r.pending, pendingFlowStep{idx: r.stepIdx, step: s})
}

// runStep implements FlowStep.
func (s Sync) runStep(r *messageFlowRunner) {
	r.send(&pgproto3.Sync{})
	r.validatePendingResponses()
	r.pending = nil
	msg := r.receiveNext()
	rfq, ok := msg.(*pgproto3.ReadyForQuery)
	if !ok {
		r.t.Fatalf("step %d: expected ReadyForQuery after Sync, but received %T: %v", r.stepIdx, msg, msg)
	}
	assertReadyStatus(r.t, r.stepIdx, s.ExpectedReadyStatus, rfq.TxStatus)
	assert.NoError(r.t, r.flowConn.EmptyReceiveBuffer(), "step %d", r.stepIdx)
}

// runStep implements FlowStep.
func (s Flush) runStep(r *messageFlowRunner) {
	r.send(&pgproto3.Flush{})
	errored := r.validatePendingResponses()
	require.False(r.t, errored, "step %d: an ErrorResponse was received before a Flush; expected errors must be "+
		"resolved by a Sync step, since the server skips messages until Sync after an error", r.stepIdx)
	r.pending = nil
}

// runStep implements FlowStep.
func (s QueryOnOtherConnection) runStep(r *messageFlowRunner) {
	t := r.t
	rows, err := r.checkConn.Query(r.ctx, s.Query)
	require.NoError(t, err, "step %d: error running query on other connection: %s", r.stepIdx, s.Query)
	defer rows.Close()
	actual := make([][]string, 0)
	for rows.Next() {
		vals, err := rows.Values()
		require.NoError(t, err, "step %d: error reading rows on other connection", r.stepIdx)
		strRow := make([]string, len(vals))
		for i, val := range vals {
			if val == nil {
				strRow[i] = "NULL"
			} else {
				strRow[i] = fmt.Sprintf("%v", val)
			}
		}
		actual = append(actual, strRow)
	}
	require.NoError(t, rows.Err(), "step %d: error reading rows on other connection", r.stepIdx)
	assertRowsEqual(t, r.stepIdx, s.Expected, actual)
}

// validatePendingResponses reads and validates the responses to every pending extended-protocol step, in order.
// Returns whether an ErrorResponse was received: once one is, the server discards all further messages until a
// Sync, so the remaining pending steps produce no responses and their expectations are skipped.
func (r *messageFlowRunner) validatePendingResponses() bool {
	errored := false
	for _, p := range r.pending {
		if errored {
			continue
		}
		// Each pending message's responses are validated in their own nested subtest, so that a failure
		// identifies the exact message whose response was wrong. A failure aborts the enclosing Sync or Flush
		// step, since the response stream is now in an unknown position.
		if !r.runSubtest(fmt.Sprintf("step %d %s", p.idx, p.step.describe()), func() {
			switch s := p.step.(type) {
			case Parse:
				errored = r.expectAck(p.idx, s.ExpectedErr, "ParseComplete", func(msg pgproto3.BackendMessage) bool {
					_, ok := msg.(*pgproto3.ParseComplete)
					return ok
				})
			case Bind:
				errored = r.expectAck(p.idx, s.ExpectedErr, "BindComplete", func(msg pgproto3.BackendMessage) bool {
					_, ok := msg.(*pgproto3.BindComplete)
					return ok
				})
			case Describe:
				errored = r.expectDescribeResponse(p.idx, s)
			case Execute:
				errored = r.expectExecuteResponse(p.idx, s)
			default:
				r.t.Fatalf("step %d: unhandled pending step type %T", p.idx, p.step)
			}
		}) {
			r.t.FailNow()
		}
	}
	return errored
}

// expectAck reads the next message and asserts that it's the expected acknowledgement (or the expected error).
// Returns whether an ErrorResponse was received.
func (r *messageFlowRunner) expectAck(stepIdx int, expectedErr string, expectedDesc string,
	matches func(pgproto3.BackendMessage) bool) bool {
	msg := r.receiveNext()
	if errResp, ok := msg.(*pgproto3.ErrorResponse); ok {
		if expectedErr == "" {
			r.t.Fatalf("step %d: unexpected error: %s", stepIdx, errResp.Message)
		}
		assert.Contains(r.t, errResp.Message, expectedErr, "step %d: wrong error message", stepIdx)
		return true
	}
	if expectedErr != "" {
		r.t.Fatalf("step %d: expected an error containing %q, but received %T: %v", stepIdx, expectedErr, msg, msg)
	}
	if !matches(msg) {
		r.t.Fatalf("step %d: expected %s, but received %T: %v", stepIdx, expectedDesc, msg, msg)
	}
	return false
}

// expectDescribeResponse reads and loosely validates the response to a Describe step: only the types of the
// response messages are checked, not their contents. Returns whether an ErrorResponse was received.
func (r *messageFlowRunner) expectDescribeResponse(stepIdx int, s Describe) bool {
	msg := r.receiveNext()
	if errResp, ok := msg.(*pgproto3.ErrorResponse); ok {
		if s.ExpectedErr == "" {
			r.t.Fatalf("step %d: unexpected error: %s", stepIdx, errResp.Message)
		}
		assert.Contains(r.t, errResp.Message, s.ExpectedErr, "step %d: wrong error message", stepIdx)
		return true
	}
	if s.ExpectedErr != "" {
		r.t.Fatalf("step %d: expected an error containing %q, but received %T: %v", stepIdx, s.ExpectedErr, msg, msg)
	}
	// Describing a statement returns a ParameterDescription before the row description
	if s.ObjectType == 'S' {
		if _, ok := msg.(*pgproto3.ParameterDescription); !ok {
			r.t.Fatalf("step %d: expected ParameterDescription, but received %T: %v", stepIdx, msg, msg)
		}
		msg = r.receiveNext()
	}
	switch msg.(type) {
	case *pgproto3.RowDescription, *pgproto3.NoData:
	default:
		r.t.Fatalf("step %d: expected RowDescription or NoData, but received %T: %v", stepIdx, msg, msg)
	}
	return false
}

// expectExecuteResponse reads and validates the response to an Execute step. Returns whether an ErrorResponse
// was received.
func (r *messageFlowRunner) expectExecuteResponse(stepIdx int, s Execute) bool {
	var rows [][]string
	for {
		msg := r.receiveNext()
		switch m := msg.(type) {
		case *pgproto3.RowDescription:
			// Postgres never sends a RowDescription in response to Execute (only Describe does that), but we
			// tolerate it here since these tests are focused on transaction semantics
		case *pgproto3.DataRow:
			rows = append(rows, textRowValues(m))
		case *pgproto3.CommandComplete:
			if s.ExpectedErr != "" {
				r.t.Fatalf("step %d: expected an error containing %q, but the Execute completed with tag %q",
					stepIdx, s.ExpectedErr, string(m.CommandTag))
			}
			if s.Tag != "" {
				assert.Equal(r.t, s.Tag, string(m.CommandTag), "step %d: wrong command tag", stepIdx)
			}
			assertRowsEqual(r.t, stepIdx, s.Rows, rows)
			return false
		case *pgproto3.EmptyQueryResponse, *pgproto3.PortalSuspended:
			if s.ExpectedErr != "" {
				r.t.Fatalf("step %d: expected an error containing %q, but received %T", stepIdx, s.ExpectedErr, msg)
			}
			assertRowsEqual(r.t, stepIdx, s.Rows, rows)
			return false
		case *pgproto3.ErrorResponse:
			if s.ExpectedErr == "" {
				r.t.Fatalf("step %d: unexpected error: %s", stepIdx, m.Message)
			}
			assert.Contains(r.t, m.Message, s.ExpectedErr, "step %d: wrong error message", stepIdx)
			return true
		default:
			r.t.Fatalf("step %d: unexpected message %T while reading Execute results: %v", stepIdx, msg, msg)
		}
	}
}

// assertStatementResults asserts that the actual per-statement results match the expected ones.
func assertStatementResults(t *testing.T, stepIdx int, expected, actual []StatementResult) {
	if !assert.Equal(t, len(expected), len(actual),
		"step %d: wrong number of statement results: expected %v, got %v", stepIdx, expected, actual) {
		return
	}
	for i := range expected {
		if expected[i].Tag != "" {
			assert.Equal(t, expected[i].Tag, actual[i].Tag, "step %d: wrong command tag for statement %d", stepIdx, i)
		}
		assertRowsEqual(t, stepIdx, expected[i].Rows, actual[i].Rows)
	}
}

// assertRowsEqual asserts that the actual rows match the expected ones, treating nil and empty as equivalent.
func assertRowsEqual(t *testing.T, stepIdx int, expected, actual [][]string) {
	if expected == nil {
		expected = [][]string{}
	}
	if actual == nil {
		actual = [][]string{}
	}
	assert.Equal(t, expected, actual, "step %d: wrong rows", stepIdx)
}

// assertReadyStatus asserts that a ReadyForQuery message's transaction status matches the expected one, with the
// zero value defaulting to 'I' (idle).
func assertReadyStatus(t *testing.T, stepIdx int, expected, actual byte) {
	if expected == 0 {
		expected = 'I'
	}
	assert.Equal(t, string(expected), string(actual),
		"step %d: wrong transaction status in ReadyForQuery", stepIdx)
}

// textRowValues converts a DataRow's values (which must be in text format) into strings, copying them out of the
// connection's shared read buffer. NULL values are rendered as "NULL".
func textRowValues(row *pgproto3.DataRow) []string {
	vals := make([]string, len(row.Values))
	for i, val := range row.Values {
		if val == nil {
			vals[i] = "NULL"
		} else {
			vals[i] = string(val)
		}
	}
	return vals
}
