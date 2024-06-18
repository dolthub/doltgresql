// Copyright 2020 Dolthub, Inc.
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
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	denginetest "github.com/dolthub/dolt/go/libraries/doltcore/sqle/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest/queries"
	"github.com/dolthub/go-mysql-server/enginetest/scriptgen/setup"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/memo"
	"github.com/dolthub/go-mysql-server/sql/mysql_db"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
	gmstypes "github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/dolt/go/libraries/doltcore/dtestutils"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/statspro"
	"github.com/dolthub/dolt/go/libraries/utils/config"
	"github.com/dolthub/dolt/go/store/datas"
	"github.com/dolthub/dolt/go/store/types"
)

var skipPrepared bool

// SkipPreparedsCount is used by the "ci-check-repo CI workflow
// as a reminder to consider prepareds when adding a new
// enginetest suite.
const SkipPreparedsCount = 83

const skipPreparedFlag = "DOLT_SKIP_PREPARED_ENGINETESTS"

func init() {
	sqle.MinRowsPerPartition = 8
	sqle.MaxRowsPerPartition = 1024

	if v := os.Getenv(skipPreparedFlag); v != "" {
		skipPrepared = true
	}
}

func TestQueries(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestQueries(t, h)
}

func TestSingleQuery(t *testing.T) {
	t.Skip()

	harness := newDoltHarness(t)
	harness.Setup(setup.SimpleSetup...)
	engine, err := harness.NewEngine(t)
	if err != nil {
		panic(err)
	}

	setupQueries := []string{
		// "create table t1 (pk int primary key, c int);",
		// "insert into t1 values (1,2), (3,4)",
		// "call dolt_add('.')",
		// "set @Commit1 = dolt_commit('-am', 'initial table');",
		// "insert into t1 values (5,6), (7,8)",
		// "set @Commit2 = dolt_commit('-am', 'two more rows');",
	}

	for _, q := range setupQueries {
		enginetest.RunQueryWithContext(t, engine, harness, nil, q)
	}

	// engine.EngineAnalyzer().Debug = true
	// engine.EngineAnalyzer().Verbose = true

	var test queries.QueryTest
	test = queries.QueryTest{
		Query: `show create table mytable`,
		Expected: []sql.Row{
			{"mytable",
				"CREATE TABLE `mytable` (\n" +
						"  `i` bigint NOT NULL,\n" +
						"  `s` varchar(20) NOT NULL COMMENT 'column s',\n" +
						"  PRIMARY KEY (`i`),\n" +
						"  KEY `idx_si` (`s`,`i`),\n" +
						"  KEY `mytable_i_s` (`i`,`s`),\n" +
						"  UNIQUE KEY `mytable_s` (`s`)\n" +
						") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_bin"},
		},
	}

	enginetest.TestQueryWithEngine(t, harness, engine, test)
}

func TestSchemaOverrides(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunSchemaOverridesTest(t, harness)
}

func newDoltEnginetestHarness(t *testing.T) denginetest.DoltEnginetestHarness {
	
}

// Convenience test for debugging a single query. Unskip and set to the desired query.
func TestSingleScript(t *testing.T) {
	t.Skip()

	var scripts = []queries.ScriptTest{
		{
			Name:        "",
			SetUpScript: []string{},
			Assertions:  []queries.ScriptTestAssertion{},
		},
	}

	for _, script := range scripts {
		harness := newDoltHarness(t)
		harness.Setup(setup.MydbData)

		engine, err := harness.NewEngine(t)
		if err != nil {
			panic(err)
		}
		// engine.EngineAnalyzer().Debug = true
		// engine.EngineAnalyzer().Verbose = true

		enginetest.TestScriptWithEngine(t, engine, harness, script)
	}
}

func newUpdateResult(matched, updated int) gmstypes.OkResult {
	return gmstypes.OkResult{
		RowsAffected: uint64(updated),
		Info:         plan.UpdateInfo{Matched: matched, Updated: updated},
	}
}

func TestAutoIncrementTrackerLockMode(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunAutoIncrementTrackerLockModeTest(t, harness)
}

// testAutoIncrementTrackerWithLockMode tests that interleaved inserts don't cause deadlocks, regardless of the value of innodb_autoinc_lock_mode.
// In a real use case, these interleaved operations would be happening in different sessions on different threads.
// In order to make the test behave predictably, we manually interleave the two iterators.
func testAutoIncrementTrackerWithLockMode(t *testing.T, harness denginetest.DoltEnginetestHarness, lockMode int64) {
	err := sql.SystemVariables.AssignValues(map[string]interface{}{"innodb_autoinc_lock_mode": lockMode})
	require.NoError(t, err)

	setupScripts := []setup.SetupScript{[]string{
		"CREATE TABLE test1 (pk int NOT NULL PRIMARY KEY AUTO_INCREMENT,c0 int,index t1_c_index (c0));",
		"CREATE TABLE test2 (pk int NOT NULL PRIMARY KEY AUTO_INCREMENT,c0 int,index t2_c_index (c0));",
		"CREATE TABLE timestamps (pk int NOT NULL PRIMARY KEY AUTO_INCREMENT, t int);",
		"CREATE TRIGGER t1 AFTER INSERT ON test1 FOR EACH ROW INSERT INTO timestamps VALUES (0, 1);",
		"CREATE TRIGGER t2 AFTER INSERT ON test2 FOR EACH ROW INSERT INTO timestamps VALUES (0, 2);",
		"CREATE VIEW bin AS SELECT 0 AS v UNION ALL SELECT 1;",
		"CREATE VIEW sequence5bit AS SELECT b1.v + 2*b2.v + 4*b3.v + 8*b4.v + 16*b5.v AS v from bin b1, bin b2, bin b3, bin b4, bin b5;",
	}}

	harness = harness.NewHarness(t)
	defer harness.Close()
	harness.Setup(setup.MydbData, setupScripts)
	e := mustNewEngine(t, harness)

	defer e.Close()
	ctx := enginetest.NewContext(harness)

	// Confirm that the system variable was correctly set.
	_, iter, err := e.Query(ctx, "select @@innodb_autoinc_lock_mode")
	require.NoError(t, err)
	rows, err := sql.RowIterToRows(ctx, iter)
	require.NoError(t, err)
	assert.Equal(t, rows, []sql.Row{{lockMode}})

	// Ordinarily QueryEngine.query manages transactions.
	// Since we can't use that for this test, we manually start a new transaction.
	ts := ctx.Session.(sql.TransactionSession)
	tx, err := ts.StartTransaction(ctx, sql.ReadWrite)
	require.NoError(t, err)
	ctx.SetTransaction(tx)

	getTriggerIter := func(query string) sql.RowIter {
		root, err := e.AnalyzeQuery(ctx, query)
		require.NoError(t, err)

		var triggerNode *plan.TriggerExecutor
		transform.Node(root, func(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
			if triggerNode != nil {
				return n, transform.SameTree, nil
			}
			if t, ok := n.(*plan.TriggerExecutor); ok {
				triggerNode = t
			}
			return n, transform.NewTree, nil
		})
		iter, err := e.EngineAnalyzer().ExecBuilder.Build(ctx, triggerNode, nil)
		require.NoError(t, err)
		return iter
	}

	iter1 := getTriggerIter("INSERT INTO test1 (c0) select v from sequence5bit;")
	iter2 := getTriggerIter("INSERT INTO test2 (c0) select v from sequence5bit;")

	// Alternate the iterators until they're exhausted.
	var err1 error
	var err2 error
	for err1 != io.EOF || err2 != io.EOF {
		if err1 != io.EOF {
			var row1 sql.Row
			require.NoError(t, err1)
			row1, err1 = iter1.Next(ctx)
			_ = row1
		}
		if err2 != io.EOF {
			require.NoError(t, err2)
			_, err2 = iter2.Next(ctx)
		}
	}
	err = iter1.Close(ctx)
	require.NoError(t, err)
	err = iter2.Close(ctx)
	require.NoError(t, err)

	dsess.DSessFromSess(ctx.Session).CommitTransaction(ctx, ctx.GetTransaction())

	// Verify that the inserts are seen by the engine.
	{
		_, iter, err := e.Query(ctx, "select count(*) from timestamps")
		require.NoError(t, err)
		rows, err := sql.RowIterToRows(ctx, iter)
		require.NoError(t, err)
		assert.Equal(t, rows, []sql.Row{{int64(64)}})
	}

	// Verify that the insert operations are actually interleaved by inspecting the order that values were added to `timestamps`
	{
		_, iter, err := e.Query(ctx, "select (select min(pk) from timestamps where t = 1) < (select max(pk) from timestamps where t = 2)")
		require.NoError(t, err)
		rows, err := sql.RowIterToRows(ctx, iter)
		require.NoError(t, err)
		assert.Equal(t, rows, []sql.Row{{true}})
	}

	{
		_, iter, err := e.Query(ctx, "select (select min(pk) from timestamps where t = 2) < (select max(pk) from timestamps where t = 1)")
		require.NoError(t, err)
		rows, err := sql.RowIterToRows(ctx, iter)
		require.NoError(t, err)
		assert.Equal(t, rows, []sql.Row{{true}})
	}
}

func TestVersionedQueries(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	defer h.Close()

	denginetest.RunVersionedQueriesTest(t, h)
}

func TestAnsiQuotesSqlMode(t *testing.T) {
	enginetest.TestAnsiQuotesSqlMode(t, newDoltHarness(t))
}

func TestAnsiQuotesSqlModePrepared(t *testing.T) {
	enginetest.TestAnsiQuotesSqlModePrepared(t, newDoltHarness(t))
}

// Tests of choosing the correct execution plan independent of result correctness. Mostly useful for confirming that
// the right indexes are being used for joining tables.
func TestQueryPlans(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunQueryTestPlans(t, harness)
}

func TestIntegrationQueryPlans(t *testing.T) {
	harness := newDoltEnginetestHarness(t).WithConfigureStats(true)
	defer harness.Close()
	enginetest.TestIntegrationPlans(t, harness)
}

func TestDoltDiffQueryPlans(t *testing.T) {
	if !types.IsFormat_DOLT(types.Format_Default) {
		t.Skip("only new format support system table indexing")
	}

	harness := newDoltEnginetestHarness(t).WithParallelism(2) // want Exchange nodes
	denginetest.RunDoltDiffQueryPlansTest(t, harness)
}

func TestBranchPlans(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunBranchPlanTests(t, harness)
}

func TestQueryErrors(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestQueryErrors(t, h)
}

func TestInfoSchema(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunInfoSchemaTests(t, h)
}

func TestColumnAliases(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestColumnAliases(t, h)
}

func TestOrderByGroupBy(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestOrderByGroupBy(t, h)
}

func TestAmbiguousColumnResolution(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestAmbiguousColumnResolution(t, h)
}

func TestInsertInto(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestInsertInto(t, h)
}

func TestInsertIgnoreInto(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestInsertIgnoreInto(t, h)
}

// TODO: merge this into the above test when we remove old format
func TestInsertDuplicateKeyKeyless(t *testing.T) {
	if !types.IsFormat_DOLT(types.Format_Default) {
		t.Skip()
	}
	enginetest.TestInsertDuplicateKeyKeyless(t, newDoltHarness(t))
}

// TODO: merge this into the above test when we remove old format
func TestInsertDuplicateKeyKeylessPrepared(t *testing.T) {
	if !types.IsFormat_DOLT(types.Format_Default) {
		t.Skip()
	}
	enginetest.TestInsertDuplicateKeyKeylessPrepared(t, newDoltHarness(t))
}

// TODO: merge this into the above test when we remove old format
func TestIgnoreIntoWithDuplicateUniqueKeyKeyless(t *testing.T) {
	if !types.IsFormat_DOLT(types.Format_Default) {
		t.Skip()
	}
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestIgnoreIntoWithDuplicateUniqueKeyKeyless(t, h)
}

// TODO: merge this into the above test when we remove old format
func TestIgnoreIntoWithDuplicateUniqueKeyKeylessPrepared(t *testing.T) {
	if !types.IsFormat_DOLT(types.Format_Default) {
		t.Skip()
	}
	enginetest.TestIgnoreIntoWithDuplicateUniqueKeyKeylessPrepared(t, newDoltHarness(t))
}

func TestInsertIntoErrors(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunInsertIntoErrorsTest(t, h)
}

func TestGeneratedColumns(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunGeneratedColumnTests(t, harness)
}

func TestGeneratedColumnPlans(t *testing.T) {
	enginetest.TestGeneratedColumnPlans(t, newDoltHarness(t))
}

func TestSpatialQueries(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestSpatialQueries(t, h)
}

func TestReplaceInto(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestReplaceInto(t, h)
}

func TestReplaceIntoErrors(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestReplaceIntoErrors(t, h)
}

func TestUpdate(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestUpdate(t, h)
}

func TestUpdateIgnore(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestUpdateIgnore(t, h)
}

func TestUpdateErrors(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestUpdateErrors(t, h)
}

func TestDeleteFrom(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestDelete(t, h)
}

func TestDeleteFromErrors(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestDeleteErrors(t, h)
}

func TestSpatialDelete(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestSpatialDelete(t, h)
}

func TestSpatialScripts(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestSpatialScripts(t, h)
}

func TestSpatialScriptsPrepared(t *testing.T) {
	enginetest.TestSpatialScriptsPrepared(t, newDoltHarness(t))
}

func TestSpatialIndexScripts(t *testing.T) {
	skipOldFormat(t)
	enginetest.TestSpatialIndexScripts(t, newDoltHarness(t))
}

func TestSpatialIndexScriptsPrepared(t *testing.T) {
	skipOldFormat(t)
	enginetest.TestSpatialIndexScriptsPrepared(t, newDoltHarness(t))
}

func TestSpatialIndexPlans(t *testing.T) {
	skipOldFormat(t)
	enginetest.TestSpatialIndexPlans(t, newDoltHarness(t))
}

func TestTruncate(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestTruncate(t, h)
}

func TestConvert(t *testing.T) {
	if types.IsFormat_LD(types.Format_Default) {
		t.Skip("noms format has outdated type enforcement")
	}
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestConvertPrepared(t, h)
}

func TestConvertPrepared(t *testing.T) {
	if types.IsFormat_LD(types.Format_Default) {
		t.Skip("noms format has outdated type enforcement")
	}
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestConvertPrepared(t, h)
}

func TestScripts(t *testing.T) {
	var skipped []string
	if types.IsFormat_DOLT(types.Format_Default) {
		skipped = append(skipped, newFormatSkippedScripts...)
	}
	h := newDoltHarness(t).WithSkippedQueries(skipped)
	defer h.Close()
	enginetest.TestScripts(t, h)
}

// TestDoltUserPrivileges tests Dolt-specific code that needs to handle user privilege checking
func TestDoltUserPrivileges(t *testing.T) {
	harness := newDoltHarness(t)
	defer harness.Close()
	for _, script := range DoltUserPrivTests {
		t.Run(script.Name, func(t *testing.T) {
			harness.Setup(setup.MydbData)
			engine, err := harness.NewEngine(t)
			require.NoError(t, err)
			defer engine.Close()

			ctx := enginetest.NewContextWithClient(harness, sql.Client{
				User:    "root",
				Address: "localhost",
			})

			engine.EngineAnalyzer().Catalog.MySQLDb.AddRootAccount()
			engine.EngineAnalyzer().Catalog.MySQLDb.SetPersister(&mysql_db.NoopPersister{})

			for _, statement := range script.SetUpScript {
				if sh, ok := interface{}(harness).(enginetest.SkippingHarness); ok {
					if sh.SkipQueryTest(statement) {
						t.Skip()
					}
				}
				enginetest.RunQueryWithContext(t, engine, harness, ctx, statement)
			}
			for _, assertion := range script.Assertions {
				if sh, ok := interface{}(harness).(enginetest.SkippingHarness); ok {
					if sh.SkipQueryTest(assertion.Query) {
						t.Skipf("Skipping query %s", assertion.Query)
					}
				}

				user := assertion.User
				host := assertion.Host
				if user == "" {
					user = "root"
				}
				if host == "" {
					host = "localhost"
				}
				ctx := enginetest.NewContextWithClient(harness, sql.Client{
					User:    user,
					Address: host,
				})

				if assertion.ExpectedErr != nil {
					t.Run(assertion.Query, func(t *testing.T) {
						enginetest.AssertErrWithCtx(t, engine, harness, ctx, assertion.Query, assertion.ExpectedErr)
					})
				} else if assertion.ExpectedErrStr != "" {
					t.Run(assertion.Query, func(t *testing.T) {
						enginetest.AssertErrWithCtx(t, engine, harness, ctx, assertion.Query, nil, assertion.ExpectedErrStr)
					})
				} else {
					t.Run(assertion.Query, func(t *testing.T) {
						enginetest.TestQueryWithContext(t, ctx, engine, harness, assertion.Query, assertion.Expected, nil, nil)
					})
				}
			}
		})
	}
}

func TestJoinOps(t *testing.T) {
	if types.IsFormat_LD(types.Format_Default) {
		t.Skip("DOLT_LD keyless indexes are not sorted")
	}

	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestJoinOps(t, h, enginetest.DefaultJoinOpTests)
}

func TestJoinPlanning(t *testing.T) {
	if types.IsFormat_LD(types.Format_Default) {
		t.Skip("DOLT_LD keyless indexes are not sorted")
	}
	h := newDoltEnginetestHarness(t).WithConfigureStats(true)
	defer h.Close()
	enginetest.TestJoinPlanning(t, h)
}

func TestJoinQueries(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestJoinQueries(t, h)
}

func TestJoinQueriesPrepared(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestJoinQueriesPrepared(t, h)
}

// TestJSONTableQueries runs the canonical test queries against a single threaded index enabled harness.
func TestJSONTableQueries(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestJSONTableQueries(t, h)
}

// TestJSONTableQueriesPrepared runs the canonical test queries against a single threaded index enabled harness.
func TestJSONTableQueriesPrepared(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestJSONTableQueriesPrepared(t, h)
}

// TestJSONTableScripts runs the canonical test queries against a single threaded index enabled harness.
func TestJSONTableScripts(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestJSONTableScripts(t, h)
}

// TestJSONTableScriptsPrepared runs the canonical test queries against a single threaded index enabled harness.
func TestJSONTableScriptsPrepared(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestJSONTableScriptsPrepared(t, h)
}

func TestUserPrivileges(t *testing.T) {
	h := newDoltHarness(t)
	h.setupTestProcedures = true
	h.configureStats = true
	defer h.Close()
	enginetest.TestUserPrivileges(t, h)
}

func TestUserAuthentication(t *testing.T) {
	t.Skip("Unexpected panic, need to fix")
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestUserAuthentication(t, h)
}

func TestComplexIndexQueries(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestComplexIndexQueries(t, h)
}

func TestCreateTable(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestCreateTable(t, h)
}

func TestRowLimit(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestRowLimit(t, h)
}

func TestBranchDdl(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunBranchDdlTest(t, h)
}

func TestBranchDdlPrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunBranchDdlTestPrepared(t, h)
}

func TestPkOrdinalsDDL(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestPkOrdinalsDDL(t, h)
}

func TestPkOrdinalsDML(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestPkOrdinalsDML(t, h)
}

func TestDropTable(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestDropTable(t, h)
}

func TestRenameTable(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestRenameTable(t, h)
}

func TestRenameColumn(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestRenameColumn(t, h)
}

func TestAddColumn(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestAddColumn(t, h)
}

func TestModifyColumn(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestModifyColumn(t, h)
}

func TestDropColumn(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestDropColumn(t, h)
}

func TestCreateDatabase(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestCreateDatabase(t, h)
}

func TestBlobs(t *testing.T) {
	skipOldFormat(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestBlobs(t, h)
}

func TestIndexes(t *testing.T) {
	harness := newDoltHarness(t)
	defer harness.Close()
	enginetest.TestIndexes(t, harness)
}

func TestIndexPrefix(t *testing.T) {
	skipOldFormat(t)
	harness := newDoltHarness(t)
	denginetest.RunIndexPrefixTest(t, harness)
}

func TestBigBlobs(t *testing.T) {
	skipOldFormat(t)

	h := newDoltHarness(t)
	denginetest.RunBigBlobsTest(t, h)
}

func TestDropDatabase(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDropEngineTest(t, h)
}

func TestCreateForeignKeys(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestCreateForeignKeys(t, h)
}

func TestDropForeignKeys(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestDropForeignKeys(t, h)
}

func TestForeignKeys(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestForeignKeys(t, h)
}

func TestForeignKeyBranches(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunForeignKeyBranchesTest(t, h)
}

func TestForeignKeyBranchesPrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunForeignKeyBranchesPreparedTest(t, h)
}

func TestFulltextIndexes(t *testing.T) {
	if !types.IsFormat_DOLT(types.Format_Default) {
		t.Skip("FULLTEXT is not supported on the old format")
	}
	if runtime.GOOS == "windows" && os.Getenv("CI") != "" {
		t.Skip("For some reason, this is flaky only on Windows CI.")
	}
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestFulltextIndexes(t, h)
}

func TestCreateCheckConstraints(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestCreateCheckConstraints(t, h)
}

func TestChecksOnInsert(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestChecksOnInsert(t, h)
}

func TestChecksOnUpdate(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestChecksOnUpdate(t, h)
}

func TestDisallowedCheckConstraints(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestDisallowedCheckConstraints(t, h)
}

func TestDropCheckConstraints(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestDropCheckConstraints(t, h)
}

func TestReadOnly(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestReadOnly(t, h, false /* testStoredProcedures */)
}

func TestViews(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestViews(t, h)
}

func TestBranchViews(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunBranchViewsTest(t, h)
}

func TestBranchViewsPrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunBranchViewsPreparedTest(t, h)
}

func TestVersionedViews(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunVersionedViewsTest(t, h)
}

func TestWindowFunctions(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestWindowFunctions(t, h)
}

func TestWindowRowFrames(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestWindowRowFrames(t, h)
}

func TestWindowRangeFrames(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestWindowRangeFrames(t, h)
}

func TestNamedWindows(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestNamedWindows(t, h)
}

func TestNaturalJoin(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestNaturalJoin(t, h)
}

func TestNaturalJoinEqual(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestNaturalJoinEqual(t, h)
}

func TestNaturalJoinDisjoint(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestNaturalJoinEqual(t, h)
}

func TestInnerNestedInNaturalJoins(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestInnerNestedInNaturalJoins(t, h)
}

func TestColumnDefaults(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestColumnDefaults(t, h)
}

func TestOnUpdateExprScripts(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestOnUpdateExprScripts(t, h)
}

func TestAlterTable(t *testing.T) {
	// This is a newly added test in GMS that dolt doesn't support yet
	h := newDoltHarness(t).WithSkippedQueries([]string{"ALTER TABLE t42 ADD COLUMN s varchar(20), drop check check1"})
	defer h.Close()
	enginetest.TestAlterTable(t, h)
}

func TestVariables(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunVariableTest(t, h)
}

func TestVariableErrors(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestVariableErrors(t, h)
}

func TestLoadDataPrepared(t *testing.T) {
	t.Skip("feature not supported")
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestLoadDataPrepared(t, h)
}

func TestLoadData(t *testing.T) {
	t.Skip()
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestLoadData(t, h)
}

func TestLoadDataErrors(t *testing.T) {
	t.Skip()
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestLoadDataErrors(t, h)
}

func TestSelectIntoFile(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestSelectIntoFile(t, h)
}

func TestJsonScripts(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	skippedTests := []string{
		"round-trip into table", // The current Dolt JSON format does not preserve decimals and unsigneds in JSON.
	}
	// TODO: fix this, use a skipping harness
	enginetest.TestJsonScripts(t, h, skippedTests)
}

func TestTriggers(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestTriggers(t, h)
}

func TestRollbackTriggers(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestRollbackTriggers(t, h)
}

func TestStoredProcedures(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunStoredProceduresTest(t, h)
}

func TestDoltStoredProcedures(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltStoredProceduresTest(t, h)
}

func TestDoltStoredProceduresPrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltStoredProceduresPreparedTest(t, h)
}

func TestEvents(t *testing.T) {
	doltHarness := newDoltHarness(t)
	defer doltHarness.Close()
	enginetest.TestEvents(t, doltHarness)
}

func TestCallAsOf(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunCallAsOfTest(t, h)
}

func TestLargeJsonObjects(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunLargeJsonObjectsTest(t, harness)
}

func SkipByDefaultInCI(t *testing.T) {
	if os.Getenv("CI") != "" && os.Getenv("DOLT_TEST_RUN_NON_RACE_TESTS") == "" {
		t.Skip()
	}
}

func TestTransactions(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunTransactionTests(t, h)
}

func TestBranchTransactions(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunBranchTransactionTest(t, h)
}

func TestMultiDbTransactions(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunMultiDbTransactionsTest(t, h)
}

func TestMultiDbTransactionsPrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunMultiDbTransactionsPreparedTest(t, h)
}

func TestConcurrentTransactions(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestConcurrentTransactions(t, h)
}

func TestDoltScripts(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDoltScriptsTest(t, harness)
}

func TestDoltTempTableScripts(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDoltTempTableScripts(t, harness)
}

func TestDoltRevisionDbScripts(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltRevisionDbScriptsTest(t, h)
}

func TestDoltRevisionDbScriptsPrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltRevisionDbScriptsPreparedTest(t, h)
}

func TestDoltDdlScripts(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDoltDdlScripts(t, harness)
}

func TestBrokenDdlScripts(t *testing.T) {
	for _, script := range BrokenDDLScripts {
		t.Skip(script.Name)
	}
}

func TestDescribeTableAsOf(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestScript(t, h, DescribeTableAsOfScriptTest)
}

func TestShowCreateTable(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunShowCreateTableTests(t, h)
}

func TestShowCreateTablePrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunShowCreateTablePreparedTests(t, h)
}

func TestViewsWithAsOf(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestScript(t, h, ViewsWithAsOfScriptTest)
}

func TestViewsWithAsOfPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestScriptPrepared(t, h, ViewsWithAsOfScriptTest)
}

func TestDoltMerge(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltMergeTests(t, h)
}

func TestDoltMergePrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltMergePreparedTests(t, h)
}

func TestDoltRebase(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltRebaseTests(t, h)
}

func TestDoltRebasePrepared(t *testing.T) {
	h := newDoltHarness(t)
	denginetest.RunDoltRebasePreparedTests(t, h)
}

func TestDoltRevert(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltRevertTests(t, h)
}

func TestDoltRevertPrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltRevertPreparedTests(t, h)
}

func TestDoltAutoIncrement(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltAutoIncrementTests(t, h)
}

func TestDoltAutoIncrementPrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltAutoIncrementPreparedTests(t, h)
}

func TestDoltConflictsTableNameTable(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltConflictsTableNameTableTests(t, h)
}

// tests new format behavior for keyless merges that create CVs and conflicts
func TestKeylessDoltMergeCVsAndConflicts(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunKelyessDoltMergeCVsAndConflictsTests(t, h)
}

// eventually this will be part of TestDoltMerge
func TestDoltMergeArtifacts(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltMergeArtifacts(t, h)
}

func TestDoltReset(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltResetTest(t, h)
}

func TestDoltGC(t *testing.T) {
	t.SkipNow()
	for _, script := range DoltGC {
		func() {
			h := newDoltHarness(t)
			defer h.Close()
			enginetest.TestScript(t, h, script)
		}()
	}
}

func TestDoltCheckout(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltCheckoutTests(t, h)
}

func TestDoltCheckoutPrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltCheckoutPreparedTests(t, h)
}

func TestDoltBranch(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltBranchTests(t, h)
}

func TestDoltTag(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltTagTests(t, h)
}

func TestDoltRemote(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltRemoteTests(t, h)
}

func TestDoltUndrop(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltUndropTests(t, h)
}

type testCommitClock struct {
	unixNano int64
}

func (tcc *testCommitClock) Now() time.Time {
	now := time.Unix(0, tcc.unixNano)
	tcc.unixNano += int64(time.Hour)
	return now
}

func installTestCommitClock(tcc *testCommitClock) func() {
	oldNowFunc := datas.CommitterDate
	oldCommitLoc := datas.CommitLoc
	datas.CommitterDate = tcc.Now
	datas.CommitLoc = time.UTC
	return func() {
		datas.CommitterDate = oldNowFunc
		datas.CommitLoc = oldCommitLoc
	}
}

func TestBrokenSystemTableQueries(t *testing.T) {
	t.Skip()

	h := newDoltHarness(t)
	defer h.Close()
	enginetest.RunQueryTests(t, h, BrokenSystemTableQueries)
}

func TestHistorySystemTable(t *testing.T) {
	harness := newDoltEnginetestHarness(t).WithParallelism(2)
	denginetest.RunHistorySystemTableTests(t, harness)
}

func TestHistorySystemTablePrepared(t *testing.T) {
	harness := newDoltEnginetestHarness(t).WithParallelism(2)
	denginetest.RunHistorySystemTableTestsPrepared(t, harness)
}

func TestBrokenHistorySystemTablePrepared(t *testing.T) {
	t.Skip()
	harness := newDoltHarness(t)
	defer harness.Close()
	harness.Setup(setup.MydbData)
	for _, test := range BrokenHistorySystemTableScriptTests {
		harness.engine = nil
		t.Run(test.Name, func(t *testing.T) {
			enginetest.TestScriptPrepared(t, harness, test)
		})
	}
}

func TestUnscopedDiffSystemTable(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunUnscopedDiffSystemTableTests(t, h)
}

func TestUnscopedDiffSystemTablePrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunUnscopedDiffSystemTableTestsPrepared(t, h)
}

func TestColumnDiffSystemTable(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunColumnDiffSystemTableTests(t, h)
}

func TestColumnDiffSystemTablePrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunColumnDiffSystemTableTestsPrepared(t, h)
}

func TestStatBranchTests(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunStatBranchTests(t, harness)
}

func TestStatsFunctions(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunStatsFunctionsTest(t, harness)
}

func TestDiffTableFunction(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDiffTableFunctionTests(t, harness)
}

func TestDiffTableFunctionPrepared(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDiffTableFunctionTestsPrepared(t, harness)
}

func TestDiffStatTableFunction(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDiffStatTableFunctionTests(t, harness)
}

func TestDiffStatTableFunctionPrepared(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDiffStatTableFunctionTestsPrepared(t, harness)
}

func TestDiffSummaryTableFunction(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDiffSummaryTableFunctionTests(t, harness)
}

func TestDiffSummaryTableFunctionPrepared(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDiffSummaryTableFunctionTestsPrepared(t, harness)
}

func TestPatchTableFunction(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDoltPatchTableFunctionTests(t, harness)
}

func TestPatchTableFunctionPrepared(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDoltPatchTableFunctionTestsPrepared(t, harness)
}

func TestLogTableFunction(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunLogTableFunctionTests(t, harness)
}

func TestLogTableFunctionPrepared(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunLogTableFunctionTestsPrepared(t, harness)
}

func TestDoltReflog(t *testing.T) {
	for _, script := range DoltReflogTestScripts {
		h := newDoltHarnessForLocalFilesystem(t)
		h.SkipSetupCommit()
		enginetest.TestScript(t, h, script)
		h.Close()
	}
}

func TestDoltReflogPrepared(t *testing.T) {
	for _, script := range DoltReflogTestScripts {
		h := newDoltHarnessForLocalFilesystem(t)
		h.SkipSetupCommit()
		enginetest.TestScriptPrepared(t, h, script)
		h.Close()
	}
}

func TestCommitDiffSystemTable(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunCommitDiffSystemTableTests(t, harness)
}

func TestCommitDiffSystemTablePrepared(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunCommitDiffSystemTableTestsPrepared(t, harness)
}

func TestDiffSystemTable(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltDiffSystemTableTests(t, h)
}

func TestDiffSystemTablePrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltDiffSystemTableTestsPrepared(t, h)
}

func TestSchemaDiffTableFunction(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunSchemaDiffTableFunctionTests(t, harness)
}

func TestSchemaDiffTableFunctionPrepared(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunSchemaDiffTableFunctionTestsPrepared(t, harness)
}

func TestDoltDatabaseCollationDiffs(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDoltDatabaseCollationDiffsTests(t, harness)
}

func TestQueryDiff(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunQueryDiffTests(t, harness)
}

func mustNewEngine(t *testing.T, h enginetest.Harness) enginetest.QueryEngine {
	e, err := h.NewEngine(t)
	if err != nil {
		require.NoError(t, err)
	}
	return e
}

var biasedCosters = []memo.Coster{
	memo.NewInnerBiasedCoster(),
	memo.NewLookupBiasedCoster(),
	memo.NewHashBiasedCoster(),
	memo.NewMergeBiasedCoster(),
}

func TestSystemTableIndexes(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunSystemTableIndexesTests(t, harness)
}

func TestSystemTableIndexesPrepared(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunSystemTableIndexesTestsPrepared(t, harness)
}

func TestSystemTableFunctionIndexes(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunSystemTableFunctionIndexesTests(t, harness)
}

func TestSystemTableFunctionIndexesPrepared(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunSystemTableFunctionIndexesTestsPrepared(t, harness)
}

func TestReadOnlyDatabases(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestReadOnlyDatabases(t, h)
}

func TestAddDropPks(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestAddDropPks(t, h)
}

func TestAddAutoIncrementColumn(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunAddAutoIncrementColumnTests(t, h)
}

func TestNullRanges(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestNullRanges(t, h)
}

func TestPersist(t *testing.T) {
	harness := newDoltHarness(t)
	defer harness.Close()
	dEnv := dtestutils.CreateTestEnv()
	defer dEnv.DoltDB.Close()
	localConf, ok := dEnv.Config.GetConfig(env.LocalConfig)
	require.True(t, ok)
	globals := config.NewPrefixConfig(localConf, env.SqlServerGlobalsPrefix)
	newPersistableSession := func(ctx *sql.Context) sql.PersistableSession {
		session := ctx.Session.(*dsess.DoltSession).WithGlobals(globals)
		err := session.RemoveAllPersistedGlobals()
		require.NoError(t, err)
		return session
	}

	enginetest.TestPersist(t, harness, newPersistableSession)
}

func TestTypesOverWire(t *testing.T) {
	harness := newDoltHarness(t)
	defer harness.Close()
	enginetest.TestTypesOverWire(t, harness, newSessionBuilder(harness))
}

func TestDoltCherryPick(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDoltCherryPickTests(t, harness)
}

func TestDoltCherryPickPrepared(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDoltCherryPickTestsPrepared(t, harness)
}

func TestDoltCommit(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDoltCommitTests(t, harness)
}

func TestDoltCommitPrepared(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDoltCommitTestsPrepared(t, harness)
}

func TestQueriesPrepared(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestQueriesPrepared(t, h)
}

func TestStatsHistograms(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunStatsHistogramTests(t, h)
}

// TestStatsIO force a provider reload in-between setup and assertions that
// forces a round trip of the statistics table before inspecting values.
func TestStatsIO(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunStatsIOTests(t, h)
}

func TestJoinStats(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunJoinStatsTests(t, h)
}

func TestStatisticIndexes(t *testing.T) {
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestStatisticIndexFilters(t, h)
}

func TestSpatialQueriesPrepared(t *testing.T) {
	skipPreparedTests(t)

	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestSpatialQueriesPrepared(t, h)
}

func TestPreparedStatistics(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunPreparedStatisticsTests(t, h)
}

func TestVersionedQueriesPrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunVersionedQueriesPreparedTests(t, h)
}

func TestInfoSchemaPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestInfoSchemaPrepared(t, h)
}

func TestUpdateQueriesPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestUpdateQueriesPrepared(t, h)
}

func TestInsertQueriesPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestInsertQueriesPrepared(t, h)
}

func TestReplaceQueriesPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestReplaceQueriesPrepared(t, h)
}

func TestDeleteQueriesPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestDeleteQueriesPrepared(t, h)
}

func TestScriptsPrepared(t *testing.T) {
	var skipped []string
	if types.IsFormat_DOLT(types.Format_Default) {
		skipped = append(skipped, newFormatSkippedScripts...)
	}
	skipPreparedTests(t)
	h := newDoltHarness(t).WithSkippedQueries(skipped)
	defer h.Close()
	enginetest.TestScriptsPrepared(t, h)
}

func TestInsertScriptsPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestInsertScriptsPrepared(t, h)
}

func TestComplexIndexQueriesPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestComplexIndexQueriesPrepared(t, h)
}

func TestJsonScriptsPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	skippedTests := []string{
		"round-trip into table", // The current Dolt JSON format does not preserve decimals and unsigneds in JSON.
	}
	enginetest.TestJsonScriptsPrepared(t, h, skippedTests)
}

func TestCreateCheckConstraintsScriptsPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestCreateCheckConstraintsScriptsPrepared(t, h)
}

func TestInsertIgnoreScriptsPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestInsertIgnoreScriptsPrepared(t, h)
}

func TestInsertErrorScriptsPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltEnginetestHarness(t)
	defer h.Close()
	h = h.WithSkippedQueries([]string{
		"create table bad (vb varbinary(65535))",
		"insert into bad values (repeat('0', 65536))",
	})
	enginetest.TestInsertErrorScriptsPrepared(t, h)
}

func TestViewsPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestViewsPrepared(t, h)
}

func TestVersionedViewsPrepared(t *testing.T) {
	t.Skip("not supported for prepareds")
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestVersionedViewsPrepared(t, h)
}

func TestShowTableStatusPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestShowTableStatusPrepared(t, h)
}

func TestPrepared(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestPrepared(t, h)
}

func TestDoltPreparedScripts(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	// TODO
	// DoltPreparedScripts(t, h)
}

func TestPreparedInsert(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestPreparedInsert(t, h)
}

func TestPreparedStatements(t *testing.T) {
	skipPreparedTests(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestPreparedStatements(t, h)
}

func TestCharsetCollationEngine(t *testing.T) {
	skipOldFormat(t)
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestCharsetCollationEngine(t, h)
}

func TestCharsetCollationWire(t *testing.T) {
	skipOldFormat(t)
	harness := newDoltHarness(t)
	defer harness.Close()
	enginetest.TestCharsetCollationWire(t, harness, newSessionBuilder(harness))
}

func TestDatabaseCollationWire(t *testing.T) {
	skipOldFormat(t)
	harness := newDoltHarness(t)
	defer harness.Close()
	enginetest.TestDatabaseCollationWire(t, harness, newSessionBuilder(harness))
}

func TestAddDropPrimaryKeys(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunAddDropPrimaryKeysTests(t, harness)
}

func TestDoltVerifyConstraints(t *testing.T) {
	harness := newDoltEnginetestHarness(t)
	denginetest.RunDoltVerifyConstraintsTests(t, harness)
}

func TestDoltStorageFormat(t *testing.T) {
	h := newDoltEnginetestHarness(t)
	denginetest.RunDoltStorageFormatTests(t, h)
}

func TestDoltStorageFormatPrepared(t *testing.T) {
	var expectedFormatString string
	if types.IsFormat_DOLT(types.Format_Default) {
		expectedFormatString = "NEW ( __DOLT__ )"
	} else {
		expectedFormatString = fmt.Sprintf("OLD ( %s )", types.Format_Default.VersionString())
	}
	h := newDoltHarness(t)
	defer h.Close()
	enginetest.TestPreparedQuery(t, h, "SELECT dolt_storage_format()", []sql.Row{{expectedFormatString}}, nil)
}

func TestThreeWayMergeWithSchemaChangeScripts(t *testing.T) {
	h := newDoltEnginetestHarness(t)

	denginetest.RunThreeWayMergeWithSchemaChangeScripts(t, h)
}

func TestThreeWayMergeWithSchemaChangeScriptsPrepared(t *testing.T) {
	h := newDoltEnginetestHarness(t)

	denginetest.RunThreeWayMergeWithSchemaChangeScriptsPrepared(t, h)
}

// If CREATE DATABASE has an error within the DatabaseProvider, it should not
// leave behind intermediate filesystem state.
func TestCreateDatabaseErrorCleansUp(t *testing.T) {
	dh := newDoltHarness(t)
	require.NotNil(t, dh)
	e, err := dh.NewEngine(t)
	require.NoError(t, err)
	require.NotNil(t, e)

	doltDatabaseProvider := dh.provider.(*sqle.DoltDatabaseProvider)
	doltDatabaseProvider.InitDatabaseHooks = append(doltDatabaseProvider.InitDatabaseHooks,
		func(_ *sql.Context, _ *sqle.DoltDatabaseProvider, name string, _ *env.DoltEnv, _ dsess.SqlDatabase) error {
			if name == "cannot_create" {
				return fmt.Errorf("there was an error initializing this database. abort!")
			}
			return nil
		})

	err = dh.provider.CreateDatabase(enginetest.NewContext(dh), "can_create")
	require.NoError(t, err)

	err = dh.provider.CreateDatabase(enginetest.NewContext(dh), "cannot_create")
	require.Error(t, err)

	fs := dh.multiRepoEnv.FileSystem()
	exists, _ := fs.Exists("cannot_create")
	require.False(t, exists)
	exists, isDir := fs.Exists("can_create")
	require.True(t, exists)
	require.True(t, isDir)
}

// TestStatsAutoRefreshConcurrency tests some common concurrent patterns that stats
// refresh is subject to -- namely reading/writing the stats objects in (1) DML statements
// (2) auto refresh threads, and (3) manual ANALYZE statements.
// todo: the dolt_stat functions should be concurrency tested
func TestStatsAutoRefreshConcurrency(t *testing.T) {
	// create engine
	harness := newDoltHarness(t)
	harness.Setup(setup.MydbData)
	engine := mustNewEngine(t, harness)
	defer engine.Close()

	enginetest.RunQueryWithContext(t, engine, harness, nil, `create table xy (x int primary key, y int, z int, key (z), key (y,z), key (y,z,x))`)
	enginetest.RunQueryWithContext(t, engine, harness, nil, `create table uv (u int primary key, v int, w int, key (w), key (w,u), key (u,w,v))`)

	sqlDb, _ := harness.provider.BaseDatabase(harness.NewContext(), "mydb")

	// Setting an interval of 0 and a threshold of 0 will result
	// in the stats being updated after every operation
	intervalSec := time.Duration(0)
	thresholdf64 := 0.
	bThreads := sql.NewBackgroundThreads()
	branches := []string{"main"}
	statsProv := engine.EngineAnalyzer().Catalog.StatsProvider.(*statspro.Provider)

	// it is important to use new sessions for this test, to avoid working root conflicts
	readCtx := enginetest.NewSession(harness)
	writeCtx := enginetest.NewSession(harness)
	newCtx := func(context.Context) (*sql.Context, error) {
		return enginetest.NewSession(harness), nil
	}

	err := statsProv.InitAutoRefreshWithParams(newCtx, sqlDb.Name(), bThreads, intervalSec, thresholdf64, branches)
	require.NoError(t, err)

	execQ := func(ctx *sql.Context, q string, id int, tag string) {
		_, iter, err := engine.Query(ctx, q)
		require.NoError(t, err)
		_, err = sql.RowIterToRows(ctx, iter)
		//fmt.Printf("%s %d\n", tag, id)
		require.NoError(t, err)
	}

	iters := 1_000
	{
		// 3 threads to test auto-refresh/DML concurrency safety
		// - auto refresh (read + write)
		// - write (write only)
		// - read (read only)

		wg := sync.WaitGroup{}
		wg.Add(2)

		go func() {
			for i := 0; i < iters; i++ {
				q := "select count(*) from xy a join xy b on a.x = b.x"
				execQ(readCtx, q, i, "read")
				q = "select count(*) from uv a join uv b on a.u = b.u"
				execQ(readCtx, q, i, "read")
			}
			wg.Done()
		}()

		go func() {
			for i := 0; i < iters; i++ {
				q := fmt.Sprintf("insert into xy values (%d,%d,%d)", i, i, i)
				execQ(writeCtx, q, i, "write")
				q = fmt.Sprintf("insert into uv values (%d,%d,%d)", i, i, i)
				execQ(writeCtx, q, i, "write")
			}
			wg.Done()
		}()

		wg.Wait()
	}

	{
		// 3 threads to test auto-refresh/manual ANALYZE concurrency
		// - auto refresh (read + write)
		// - add (read + write)
		// - drop (write only)

		wg := sync.WaitGroup{}
		wg.Add(2)

		analyzeAddCtx := enginetest.NewSession(harness)
		analyzeDropCtx := enginetest.NewSession(harness)

		// hammer the provider with concurrent stat updates
		go func() {
			for i := 0; i < iters; i++ {
				execQ(analyzeAddCtx, "analyze table xy,uv", i, "analyze create")
			}
			wg.Done()
		}()

		go func() {
			for i := 0; i < iters; i++ {
				execQ(analyzeDropCtx, "analyze table xy drop histogram on (y,z)", i, "analyze drop yz")
				execQ(analyzeDropCtx, "analyze table uv drop histogram on (w,u)", i, "analyze drop wu")
			}
			wg.Done()
		}()

		wg.Wait()
	}
}

var newFormatSkippedScripts = []string{
	// Different query plans
	"Partial indexes are used and return the expected result",
	"Multiple indexes on the same columns in a different order",
}

func skipOldFormat(t *testing.T) {
	if !types.IsFormat_DOLT(types.Format_Default) {
		t.Skip()
	}
}

func skipPreparedTests(t *testing.T) {
	if skipPrepared {
		t.Skip("skip prepared")
	}
}

func newSessionBuilder(harness *DoltHarness) server.SessionBuilder {
	return func(ctx context.Context, conn *mysql.Conn, host string) (sql.Session, error) {
		newCtx := harness.NewSession()
		return newCtx.Session, nil
	}
}
