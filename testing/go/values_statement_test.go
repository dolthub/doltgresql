package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestValuesStatement(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "basic values statements",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM (VALUES (1), (2), (3)) sqa;`,
					Expected: []sql.Row{
						{1},
						{2},
						{3},
					},
				},
				{
					Query: `SELECT * FROM (VALUES (1, 2), (3, 4)) sqa;`,
					Expected: []sql.Row{
						{1, 2},
						{3, 4},
					},
				},
				{
					Query: `SELECT i * 10, j * 100 FROM (VALUES (1, 2), (3, 4)) sqa(i, j);`,
					Expected: []sql.Row{
						{10, 200},
						{30, 400},
					},
				},
			},
		},
	})
}
