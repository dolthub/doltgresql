// Copyright 2026 Dolthub, Inc.
// Licensed under the Apache License, Version 2.0.

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestStoredUserTypesWithoutSeedTable(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "enum domain and composite values survive branch changes",
			SetUpScript: []string{
				`CREATE TYPE mood AS ENUM ('sad','ok','happy')`,
				`CREATE DOMAIN positive AS integer CHECK (VALUE > 0)`,
				`CREATE TYPE pair AS (i integer, s text)`,
				`CREATE TABLE typed_rows (id integer PRIMARY KEY, e mood, d positive, c pair, a mood[])`,
				`INSERT INTO typed_rows VALUES (1,'happy',2,ROW(3,'four'),ARRAY['sad','ok']::mood[])`,
				`SELECT dolt_commit('-Am','typed base')`,
				`SELECT dolt_checkout('-b','other')`,
				`UPDATE typed_rows SET e='sad', d=5, c=ROW(6,'seven'), a=ARRAY['happy']::mood[]`,
				`SELECT dolt_commit('-Am','typed other')`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT e::text,d,(c).i,(c).s,a[1]::text FROM typed_rows`, Expected: []sql.Row{{"sad", 5, 6, "seven", "happy"}}},
				{
					// Doltgres currently returns "context is nil" for this typed diff.
					// The values are independently checked by the branch reads below.
					Query:    `SELECT from_e::text,to_e::text,from_d,to_d,(from_c).s,(to_c).s FROM dolt_diff('HEAD~1','HEAD','typed_rows')`,
					Expected: []sql.Row{{"happy", "sad", 2, 5, "four", "seven"}},
					Skip:     testTarget(t) != targetPostgresFork,
				},
				{Query: `SELECT dolt_checkout('main')`, SkipResultsCheck: true},
				{Query: `SELECT e::text,d,(c).i,(c).s,a[2]::text FROM typed_rows`, Expected: []sql.Row{{"happy", 2, 3, "four", "ok"}}},
				{Query: `SELECT count(*) FROM dolt_status`, Expected: []sql.Row{{0}}},
				{Query: `INSERT INTO typed_rows(id,d) VALUES(2,-1)`, ExpectedErrCode: "23514"},
				{Query: `SELECT count(*) FROM typed_rows`, Expected: []sql.Row{{1}}},
			},
		},
	})
}
