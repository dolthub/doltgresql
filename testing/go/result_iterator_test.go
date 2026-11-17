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
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// TestStatementEndsItsTransaction checks that a statement's autocommit transaction ends with the statement, however
// its results come out, so the session's next statement sees what other sessions committed in between.
func TestStatementEndsItsTransaction(t *testing.T) {
	RunTransactionTests(t, []ScriptTest{
		{
			Name: "statement that fails while running",
			SetUpScript: []string{
				"CREATE TABLE t (a INT PRIMARY KEY);",
				"INSERT INTO t VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ INSERT INTO t VALUES (1);", ExpectedErr: "duplicate"},
				{Query: "/* client b */ INSERT INTO t VALUES (2);"},
				{Query: "/* client a */ SELECT a FROM t ORDER BY a;", Expected: []sql.Row{{1}, {2}}},
			},
		},
		{
			Name: "single-row lookup with no match",
			SetUpScript: []string{
				"CREATE TABLE t (a INT PRIMARY KEY);",
				"INSERT INTO t VALUES (1);",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "/* client a */ SELECT a FROM t WHERE a = 100;", Expected: []sql.Row{}},
				{Query: "/* client b */ INSERT INTO t VALUES (2);"},
				{Query: "/* client a */ SELECT a FROM t ORDER BY a;", Expected: []sql.Row{{1}, {2}}},
			},
		},
	})
}

// TestStatementEndsItsTransactionSimpleProtocol is TestStatementEndsItsTransaction over the simple query protocol,
// which also reaches the result paths for statements that return at most one row.
func TestStatementEndsItsTransactionSimpleProtocol(t *testing.T) {
	setup := []string{"CREATE TABLE t (a INT PRIMARY KEY);", "INSERT INTO t VALUES (1);"}
	bothRows := SimpleQuery{
		Query:    "SELECT a FROM t ORDER BY a;",
		Expected: []StatementResult{{Tag: "SELECT 2", Rows: [][]string{{"1"}, {"2"}}}},
	}
	RunMessageFlowTests(t, []MessageFlowTest{
		{
			Name:        "statement that fails while running",
			SetUpScript: setup,
			Steps: []FlowStep{
				SimpleQuery{Query: "INSERT INTO t VALUES (1);", ExpectedErr: "duplicate"},
				QueryOnOtherConnection{Query: "INSERT INTO t VALUES (2);"},
				bothRows,
			},
		},
		{
			Name:        "single-row lookup with no match",
			SetUpScript: setup,
			Steps: []FlowStep{
				SimpleQuery{Query: "SELECT a FROM t WHERE a = 100;", Expected: []StatementResult{{Tag: "SELECT 0"}}},
				QueryOnOtherConnection{Query: "INSERT INTO t VALUES (2);"},
				bothRows,
			},
		},
	})
}
