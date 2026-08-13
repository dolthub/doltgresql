package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestDiscard(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Test discard",
			SetUpScript: []string{
				`CREATE temporary TABLE test (a INT)`,
				`insert into test values (1)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select * from test",
					Expected: []sql.Row{
						{1},
					},
				},
				{
					Query:    "DISCARD ALL",
					Expected: []sql.Row{},
				},
				{
					Query:       "select * from test",
					ExpectedErr: "table not found",
				},
			},
		},
		{
			Name: "Test discard errors",
			SetUpScript: []string{
				`CREATE temporary TABLE test (a INT)`,
				`insert into test values (1)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "DISCARD SEQUENCES",
					ExpectedErr: "unimplemented",
				},
				{
					Query: "select * from test",
					Expected: []sql.Row{
						{1},
					},
				},
			},
		},
		{
			Name: "Test discard in transaction",
			SetUpScript: []string{
				`CREATE temporary TABLE test (a INT)`,
				`insert into test values (1)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN",
				},
				{
					Query:       "DISCARD ALL",
					ExpectedErr: "DISCARD ALL cannot run inside a transaction block",
					Skip:        true, // not yet implemented
				},
			},
		},
	})
}

// TestBeginIsolationLevel asserts that BEGIN statements accept any transaction isolation level clause.
func TestBeginIsolationLevel(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "BEGIN with any isolation level clause is accepted as a no-op",
			SetUpScript: []string{
				`CREATE TABLE test (a INT PRIMARY KEY)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "BEGIN TRANSACTION ISOLATION LEVEL READ COMMITTED",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test VALUES (1)",
					Expected: []sql.Row{},
				},
				{
					Query:    "COMMIT",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "BEGIN ISOLATION LEVEL READ UNCOMMITTED",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "ROLLBACK",
					Expected: []sql.Row{},
				},
				{
					Query:    "BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "COMMIT",
					Expected: []sql.Row{},
				},
				{
					Query:    "BEGIN ISOLATION LEVEL REPEATABLE READ, READ WRITE",
					Expected: []sql.Row{},
				},
				{
					Query:    "COMMIT",
					Expected: []sql.Row{},
				},
			},
		},
		{
			// Postgres treats a BEGIN issued while already inside a transaction as a no-op.
			Name: "A duplicate BEGIN does not change the active transaction's characteristics",
			SetUpScript: []string{
				`CREATE TABLE test (a INT PRIMARY KEY)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "BEGIN ISOLATION LEVEL SERIALIZABLE, READ WRITE",
					Expected: []sql.Row{},
				},
				{
					Query:    "BEGIN ISOLATION LEVEL REPEATABLE READ, READ ONLY",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test VALUES (1)",
					Expected: []sql.Row{},
				},
				{
					Query:    "COMMIT",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "ROLLBACK clears the in-transaction flag so a following BEGIN is honored",
			SetUpScript: []string{
				`CREATE TABLE test_rollback (a INT)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "BEGIN READ WRITE",
					Expected: []sql.Row{},
				},
				{
					Query:    "BEGIN READ ONLY",
					Expected: []sql.Row{},
				},
				{
					Query:    "ROLLBACK",
					Expected: []sql.Row{},
				},
				{
					Query:    "BEGIN READ ONLY",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO test_rollback VALUES (1)",
					ExpectedErr: "READ ONLY",
				},
			},
		},
	})
}

func TestRollback(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Test rollback transaction",
			SetUpScript: []string{
				`BEGIN`,
				`CREATE temporary TABLE test (a INT)`,
				`insert into test values (1)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select * from test",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "ROLLBACK",
					Expected: []sql.Row{},
				},
				{
					Query:       "select * from test",
					ExpectedErr: "table not found",
					Skip:        true, // temp table should be dropped after ROLLBACK
				},
				{
					Query:    "create temp table test (b int)",
					Expected: []sql.Row{},
					Skip:     true, // temp table should be dropped after ROLLBACK
				},
			},
		},
	})
}

func TestSessionStateAfterQueryError(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Test failed query does not pin the session to a stale root",
			SetUpScript: []string{
				`CREATE TABLE test (a INT PRIMARY KEY)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT * FROM doesnotexist",
					ExpectedErr: "table not found",
				},
				{
					Username: "postgres",
					Password: "password",
					Query:    "INSERT INTO test VALUES (1)",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "Test failed root object lookup does not pin the session to a stale root",
			SetUpScript: []string{
				`CREATE TABLE test (a INT PRIMARY KEY)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SELECT doesnotexist()",
					ExpectedErr: "'doesnotexist' not found",
				},
				{
					Username: "postgres",
					Password: "password",
					Query:    "CREATE SEQUENCE seq",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('seq')",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "Test failed query inside a transaction aborts the transaction",
			SetUpScript: []string{
				`CREATE TABLE test (a INT PRIMARY KEY)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "START TRANSACTION",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT * FROM doesnotexist",
					ExpectedErr: "table not found",
				},
				{
					Username: "postgres",
					Password: "password",
					Query:    "INSERT INTO test VALUES (1)",
					Expected: []sql.Row{},
				},
				{ // The failed transaction rejects all statements until it is ended
					Query:       "SELECT * FROM test",
					ExpectedErr: "current transaction is aborted",
				},
				{ // COMMIT ends the failed transaction by rolling it back
					Query:       "COMMIT",
					ExpectedTag: "ROLLBACK",
				},
				{ // With the failed transaction ended, the other session's write is visible
					Query:    "SELECT * FROM test",
					Expected: []sql.Row{{1}},
				},
			},
		},
	})
}
