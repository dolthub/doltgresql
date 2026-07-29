package enginetest

import (
	denginetest "github.com/dolthub/dolt/go/libraries/doltcore/sqle/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest/queries"
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

func RunDoltgresTransactionTests(t *testing.T, h denginetest.DoltEnginetestHarness, prepared bool) {
	for _, script := range SequenceTransactionTests {
		func() {
			h := h.NewHarness(t)
			defer h.Close()
			if prepared {
				enginetest.TestTransactionScriptPrepared(t, h, script)
			} else {
				enginetest.TestTransactionScript(t, h, script)
			}
		}()
	}
}

var SequenceTransactionTests = []queries.TransactionTest{
	{
		Name: "two auto increment values in two transactions",
		SetUpScript: []string{
			"CREATE SEQUENCE myseq AS integer START WITH 1 INCREMENT BY 2 NO MINVALUE NO MAXVALUE CACHE 1;",
		},
		Assertions: []queries.ScriptTestAssertion{
			{
				Query: "/* client a */ start transaction",
			},
			{
				Query: "/* client b */ start transaction",
			},
			{
				Query:    "/* client a */ select nextval('myseq')",
				Expected: []sql.Row{{1}},
			},
			{
				Query:    "/* client b */ select nextval('myseq')",
				Expected: []sql.Row{{3}},
			},
		},
	},
}
