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
