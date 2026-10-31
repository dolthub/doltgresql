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

package enginetest

import (
	"strings"

	"github.com/dolthub/go-mysql-server/enginetest/queries"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

// PostgreSQL stores expression trees in pg_index, not hidden table columns.
// These counterparts retain the data operations and independently assert its
// catalog representation and automatic dependency handling.
func nativeIndexScript(script queries.ScriptTest) queries.ScriptTest {
	if !nativePostgresEngine() {
		return script
	}
	script.Assertions = append([]queries.ScriptTestAssertion(nil), script.Assertions...)
	script.SetUpScript = append([]string(nil), script.SetUpScript...)
	ok := []sql.Row{{types.NewOkResult(0)}}
	for i := range script.Assertions {
		a := &script.Assertions[i]
		if strings.HasPrefix(a.ExpectedErrStr, "duplicate unique key given") {
			a.ExpectedErrStr = ""
			a.ExpectedErr = sql.ErrUniqueKeyViolation
		}
		if strings.Contains(a.ExpectedErrStr, "functional index dependency") {
			a.ExpectedErrStr = ""
			a.Expected = ok
		}
		if strings.Contains(a.ExpectedErrStr, "could not be found in any table in scope") && strings.Contains(a.Query, "!hidden!") {
			a.ExpectedErrStr = "does not exist"
		}
	}
	switch script.Name {
	case "choose range over full prefix match":
		// PostgreSQL gathers its own statistics; Dolt histogram injection is not an oracle for its planner.
		script.SetUpScript = append(script.SetUpScript[:2], "ANALYZE xy", "SET enable_seqscan=off")
		script.Assertions = append(script.Assertions, queries.ScriptTestAssertion{Query: "SELECT x FROM xy WHERE x > 1 AND y=1 AND w='a'", Expected: []sql.Row{{2}}})
	case "DDL: RENAME COLUMN blocked by functional index dependency":
		script.Assertions = append(script.Assertions, queries.ScriptTestAssertion{Query: "SELECT pk FROM test WHERE c1_new*10=1000", Expected: []sql.Row{{1}}}, queries.ScriptTestAssertion{Query: "SELECT pg_get_indexdef('idx1'::regclass,1,true)", Expected: []sql.Row{{"(c1_new * 10)"}}})
	case "DDL: DROP COLUMN blocked by functional index dependency", "multiple expressions: DROP COLUMN blocked by dependency for every column referenced by an expression":
		script.Assertions = append(script.Assertions, queries.ScriptTestAssertion{Query: "SELECT count(*) FROM pg_indexes WHERE indexname='idx1' AND tablename='test'", Expected: []sql.Row{{0}}})
	case "!hidden! column name prefix is reserved":
		script.Assertions = []queries.ScriptTestAssertion{
			{Query: "CREATE TABLE t (pk int PRIMARY KEY, `!hidden!idx!0` int)", Expected: ok},
			{Query: "CREATE TABLE t_upper (pk int PRIMARY KEY, `!HIDDEN!idx!0` int)", Expected: ok},
			{Query: "ALTER TABLE t2 RENAME COLUMN c1 TO `!hidden!idx!0`", Expected: ok},
			{Query: `SELECT column_name FROM information_schema.columns WHERE table_name='t2' ORDER BY ordinal_position`, Expected: []sql.Row{{"pk"}, {"!hidden!idx!0"}}},
		}
	case "multiple expressions: SHOW INDEX reports all key parts in the correct order":
		script.Assertions = []queries.ScriptTestAssertion{
			{Query: `SELECT k.n, a.attname FROM pg_index i CROSS JOIN LATERAL unnest(i.indkey) WITH ORDINALITY k(attnum,n) LEFT JOIN pg_attribute a ON a.attrelid=i.indrelid AND a.attnum=k.attnum WHERE i.indexrelid='idx1'::regclass ORDER BY k.n`, Expected: []sql.Row{{1, nil}, {2, "c2"}, {3, nil}}},
			{Query: `SELECT n, pg_get_indexdef('idx1'::regclass,n,true) FROM generate_series(1,3) n`, Expected: []sql.Row{{1, "COALESCE(c1, 0)"}, {2, "c2"}, {3, "COALESCE(c3, 0)"}}},
		}
	case "functional index NULL field reflects the expression's actual nullability, not just the underlying column":
		// Native pg_attribute conservatively leaves expression keys nullable. Test
		// the expression definitions and actual NULL uniqueness semantics instead.
		script.Assertions = []queries.ScriptTestAssertion{
			{Query: `SELECT pg_get_indexdef('idx_not_null'::regclass,1,true), pg_get_indexdef('idx_nullable'::regclass,1,true)`, Expected: []sql.Row{{"COALESCE(b, a)", "COALESCE(b, b)"}}},
			{Query: `INSERT INTO test VALUES (1,'x',NULL),(2,'y',NULL)`, Expected: []sql.Row{{types.NewOkResult(2)}}},
			{Query: `INSERT INTO test VALUES (3,'x',NULL)`, ExpectedErr: sql.ErrUniqueKeyViolation},
			{Query: `SELECT count(*) FROM test`, Expected: []sql.Row{{2}}},
		}
	case "JSON functions":
		script.SetUpScript = []string{`CREATE TABLE events(id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, payload jsonb NOT NULL)`, `INSERT INTO events(payload) VALUES ('{"amount":149,"user_id":42,"event_type":"purchase"}')`}
		script.Assertions = []queries.ScriptTestAssertion{
			{Query: `SELECT id, (payload->>'amount')::numeric(10,2) FROM events`, Expected: []sql.Row{{1, "149.00"}}},
			{Query: `CREATE INDEX idx1 ON events (((payload->>'amount')::numeric(10,2)))`, Expected: ok},
			{Query: `SELECT id FROM events WHERE (payload->>'amount')::numeric(10,2)=149.0`, Expected: []sql.Row{{1}}},
		}
	case "unique key violation prevents insert":
		for i := range script.Assertions {
			if strings.HasPrefix(strings.ToLower(script.Assertions[i].Query), "show create table") {
				script.Assertions[i] = queries.ScriptTestAssertion{Query: `SELECT count(*) FROM pg_index WHERE indrelid='users'::regclass AND indisunique`, Expected: []sql.Row{{2}}}
			}
		}
	case "missing indexes":
		for i := range script.Assertions {
			script.Assertions[i].Query = strings.ReplaceAll(script.Assertions[i].Query, "BIN_TO_UUID(id)", "id")
		}
	}
	return script
}
