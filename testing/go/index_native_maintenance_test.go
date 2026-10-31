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
	"github.com/dolthub/go-mysql-server/sql"
	"testing"
)

func TestNativeExpressionIndexMaintenance(t *testing.T) {
	if testTarget(t) == targetDoltgres {
		t.Skip("PostgreSQL executor and partial-index counterpart")
	}
	for _, mode := range []string{"SET enable_bitmapscan=off", "SET enable_indexscan=off", "SET enable_indexonlyscan=off"} {
		RunScripts(t, []ScriptTest{{
			Name: mode,
			SetUpScript: []string{
				"CREATE TABLE t(id int PRIMARY KEY, name varchar(20), n int, live bool)",
				"INSERT INTO t VALUES (1,'Alpha',2,true),(2,'ALPHA',2,false),(3,NULL,3,true),(4,'beta',4,true)",
				"CREATE UNIQUE INDEX expr ON t(lower(name), (n+1) DESC) INCLUDE(id) WHERE live",
				"CREATE INDEX pattern ON t(name varchar_pattern_ops)",
				"SET enable_seqscan=off", mode,
			},
			Assertions: []ScriptTestAssertion{
				{Query: "EXPLAIN (FORMAT JSON) SELECT id FROM t WHERE live AND lower(name)='alpha' AND n+1=3", ExpectedPlanNodes: []map[string]any{{"Index Name": "expr"}}},
				{Query: "SELECT id FROM t WHERE live AND lower(name)='alpha' AND n+1=3", Expected: []sql.Row{{1}}},
				{Query: "SELECT id FROM t WHERE live AND lower(name) IS NULL", Expected: []sql.Row{{3}}},
				{Query: "UPDATE t SET live=true WHERE id=2", ExpectedErrCode: "23505"},
				{Query: "UPDATE t SET live=false WHERE id=1", Expected: []sql.Row{}},
				{Query: "UPDATE t SET live=true WHERE id=2", Expected: []sql.Row{}},
				{Query: "SELECT id FROM t WHERE live AND lower(name)='alpha'", Expected: []sql.Row{{2}}},
				{Query: "SELECT id FROM t WHERE name LIKE 'Al%'", Expected: []sql.Row{{1}}},
				{Query: "DELETE FROM t WHERE id=2", Expected: []sql.Row{}},
				{Query: "SELECT id FROM t WHERE live AND lower(name)='alpha'", Expected: []sql.Row{}},
				{Query: "INSERT INTO t VALUES (5,'ALpha',2,true)", Expected: []sql.Row{}},
				{Query: "SELECT id FROM t WHERE live AND lower(name) IN ('alpha','beta') ORDER BY id", Expected: []sql.Row{{4}, {5}}},
			},
		}})
	}
}
