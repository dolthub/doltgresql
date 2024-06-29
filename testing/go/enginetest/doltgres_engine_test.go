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
	"os"
	"runtime"
	"testing"

	denginetest "github.com/dolthub/dolt/go/libraries/doltcore/sqle/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest/queries"
	"github.com/dolthub/go-mysql-server/enginetest/scriptgen/setup"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/dolt/go/libraries/doltcore/dtestutils"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/utils/config"
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
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestQueries(t, h)
}

func TestSingleQuery(t *testing.T) {
	t.Skip()

	harness := newDoltgresServerHarness(t)
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

	test := queries.QueryTest{
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
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunSchemaOverridesTest(t, harness)
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
		harness := newDoltgresServerHarness(t)
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

func TestAutoIncrementTrackerLockMode(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunAutoIncrementTrackerLockModeTest(t, harness)
}

func TestVersionedQueries(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()

	denginetest.RunVersionedQueriesTest(t, h)
}

func TestAnsiQuotesSqlMode(t *testing.T) {
	t.Skip()
	enginetest.TestAnsiQuotesSqlMode(t, newDoltgresServerHarness(t))
}

func TestAnsiQuotesSqlModePrepared(t *testing.T) {
	t.Skip()
	enginetest.TestAnsiQuotesSqlModePrepared(t, newDoltgresServerHarness(t))
}

// Tests of choosing the correct execution plan independent of result correctness. Mostly useful for confirming that
// the right indexes are being used for joining tables.
func TestQueryPlans(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunQueryTestPlans(t, harness)
}

func TestIntegrationQueryPlans(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t).WithConfigureStats(true)
	defer harness.Close()
	enginetest.TestIntegrationPlans(t, harness)
}

func TestDoltDiffQueryPlans(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t).WithParallelism(2) // want Exchange nodes
	denginetest.RunDoltDiffQueryPlansTest(t, harness)
}

func TestBranchPlans(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunBranchPlanTests(t, harness)
}

func TestQueryErrors(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestQueryErrors(t, h)
}

func TestInfoSchema(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunInfoSchemaTests(t, h)
}

func TestColumnAliases(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestColumnAliases(t, h)
}

func TestOrderByGroupBy(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestOrderByGroupBy(t, h)
}

func TestAmbiguousColumnResolution(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestAmbiguousColumnResolution(t, h)
}

func TestInsertInto(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestInsertInto(t, h)
}

func TestInsertIgnoreInto(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestInsertIgnoreInto(t, h)
}

// TODO: merge this into the above test when we remove old format
func TestInsertDuplicateKeyKeyless(t *testing.T) {
	t.Skip()
	enginetest.TestInsertDuplicateKeyKeyless(t, newDoltgresServerHarness(t))
}

// TODO: merge this into the above test when we remove old format
func TestInsertDuplicateKeyKeylessPrepared(t *testing.T) {
	t.Skip()
	enginetest.TestInsertDuplicateKeyKeylessPrepared(t, newDoltgresServerHarness(t))
}

// TODO: merge this into the above test when we remove old format
func TestIgnoreIntoWithDuplicateUniqueKeyKeyless(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestIgnoreIntoWithDuplicateUniqueKeyKeyless(t, h)
}

// TODO: merge this into the above test when we remove old format
func TestIgnoreIntoWithDuplicateUniqueKeyKeylessPrepared(t *testing.T) {
	t.Skip()
	enginetest.TestIgnoreIntoWithDuplicateUniqueKeyKeylessPrepared(t, newDoltgresServerHarness(t))
}

func TestInsertIntoErrors(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunInsertIntoErrorsTest(t, h)
}

func TestGeneratedColumns(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunGeneratedColumnTests(t, harness)
}

func TestGeneratedColumnPlans(t *testing.T) {
	t.Skip()
	enginetest.TestGeneratedColumnPlans(t, newDoltgresServerHarness(t))
}

func TestSpatialQueries(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestSpatialQueries(t, h)
}

func TestReplaceInto(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestReplaceInto(t, h)
}

func TestReplaceIntoErrors(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestReplaceIntoErrors(t, h)
}

func TestUpdate(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestUpdate(t, h)
}

func TestUpdateIgnore(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestUpdateIgnore(t, h)
}

func TestUpdateErrors(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestUpdateErrors(t, h)
}

func TestDeleteFrom(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestDelete(t, h)
}

func TestDeleteFromErrors(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestDeleteErrors(t, h)
}

func TestSpatialDelete(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestSpatialDelete(t, h)
}

func TestSpatialScripts(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestSpatialScripts(t, h)
}

func TestSpatialScriptsPrepared(t *testing.T) {
	t.Skip()
	enginetest.TestSpatialScriptsPrepared(t, newDoltgresServerHarness(t))
}

func TestSpatialIndexScripts(t *testing.T) {
	t.Skip()
	enginetest.TestSpatialIndexScripts(t, newDoltgresServerHarness(t))
}

func TestSpatialIndexScriptsPrepared(t *testing.T) {
	t.Skip()
	enginetest.TestSpatialIndexScriptsPrepared(t, newDoltgresServerHarness(t))
}

func TestSpatialIndexPlans(t *testing.T) {
	t.Skip()
	enginetest.TestSpatialIndexPlans(t, newDoltgresServerHarness(t))
}

func TestTruncate(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestTruncate(t, h)
}

func TestConvert(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestConvertPrepared(t, h)
}

func TestConvertPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestConvertPrepared(t, h)
}

func TestScripts(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t).WithSkippedQueries(newFormatSkippedScripts)
	defer h.Close()
	enginetest.TestScripts(t, h)
}

func TestJoinOps(t *testing.T) {
	t.Skip()

	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestJoinOps(t, h, enginetest.DefaultJoinOpTests)
}

func TestJoinPlanning(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t).WithConfigureStats(true)
	defer h.Close()
	enginetest.TestJoinPlanning(t, h)
}

func TestJoinQueries(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestJoinQueries(t, h)
}

func TestJoinQueriesPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestJoinQueriesPrepared(t, h)
}

// TestJSONTableQueries runs the canonical test queries against a single threaded index enabled harness.
func TestJSONTableQueries(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestJSONTableQueries(t, h)
}

// TestJSONTableQueriesPrepared runs the canonical test queries against a single threaded index enabled harness.
func TestJSONTableQueriesPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestJSONTableQueriesPrepared(t, h)
}

// TestJSONTableScripts runs the canonical test queries against a single threaded index enabled harness.
func TestJSONTableScripts(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestJSONTableScripts(t, h)
}

// TestJSONTableScriptsPrepared runs the canonical test queries against a single threaded index enabled harness.
func TestJSONTableScriptsPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestJSONTableScriptsPrepared(t, h)
}

func TestUserAuthentication(t *testing.T) {
	t.Skip("Unexpected panic, need to fix")
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestUserAuthentication(t, h)
}

func TestComplexIndexQueries(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestComplexIndexQueries(t, h)
}

func TestCreateTable(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestCreateTable(t, h)
}

func TestRowLimit(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestRowLimit(t, h)
}

func TestBranchDdl(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunBranchDdlTest(t, h)
}

func TestBranchDdlPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunBranchDdlTestPrepared(t, h)
}

func TestPkOrdinalsDDL(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestPkOrdinalsDDL(t, h)
}

func TestPkOrdinalsDML(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestPkOrdinalsDML(t, h)
}

func TestDropTable(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestDropTable(t, h)
}

func TestRenameTable(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestRenameTable(t, h)
}

func TestRenameColumn(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestRenameColumn(t, h)
}

func TestAddColumn(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestAddColumn(t, h)
}

func TestModifyColumn(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestModifyColumn(t, h)
}

func TestDropColumn(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestDropColumn(t, h)
}

func TestCreateDatabase(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestCreateDatabase(t, h)
}

func TestBlobs(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestBlobs(t, h)
}

func TestIndexes(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	defer harness.Close()
	enginetest.TestIndexes(t, harness)
}

func TestIndexPrefix(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunIndexPrefixTest(t, harness)
}

func TestBigBlobs(t *testing.T) {
	t.Skip()

	h := newDoltgresServerHarness(t)
	denginetest.RunBigBlobsTest(t, h)
}

func TestDropDatabase(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDropEngineTest(t, h)
}

func TestCreateForeignKeys(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestCreateForeignKeys(t, h)
}

func TestDropForeignKeys(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestDropForeignKeys(t, h)
}

func TestForeignKeys(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestForeignKeys(t, h)
}

func TestForeignKeyBranches(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunForeignKeyBranchesTest(t, h)
}

func TestForeignKeyBranchesPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunForeignKeyBranchesPreparedTest(t, h)
}

func TestFulltextIndexes(t *testing.T) {
	t.Skip()
	if runtime.GOOS == "windows" && os.Getenv("CI") != "" {
		t.Skip("For some reason, this is flaky only on Windows CI.")
	}
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestFulltextIndexes(t, h)
}

func TestCreateCheckConstraints(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestCreateCheckConstraints(t, h)
}

func TestChecksOnInsert(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestChecksOnInsert(t, h)
}

func TestChecksOnUpdate(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestChecksOnUpdate(t, h)
}

func TestDisallowedCheckConstraints(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestDisallowedCheckConstraints(t, h)
}

func TestDropCheckConstraints(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestDropCheckConstraints(t, h)
}

func TestReadOnly(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestReadOnly(t, h, false /* testStoredProcedures */)
}

func TestViews(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestViews(t, h)
}

func TestBranchViews(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunBranchViewsTest(t, h)
}

func TestBranchViewsPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunBranchViewsPreparedTest(t, h)
}

func TestVersionedViews(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunVersionedViewsTest(t, h)
}

func TestWindowFunctions(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestWindowFunctions(t, h)
}

func TestWindowRowFrames(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestWindowRowFrames(t, h)
}

func TestWindowRangeFrames(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestWindowRangeFrames(t, h)
}

func TestNamedWindows(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestNamedWindows(t, h)
}

func TestNaturalJoin(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestNaturalJoin(t, h)
}

func TestNaturalJoinEqual(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestNaturalJoinEqual(t, h)
}

func TestNaturalJoinDisjoint(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestNaturalJoinEqual(t, h)
}

func TestInnerNestedInNaturalJoins(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestInnerNestedInNaturalJoins(t, h)
}

func TestColumnDefaults(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestColumnDefaults(t, h)
}

func TestOnUpdateExprScripts(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestOnUpdateExprScripts(t, h)
}

func TestAlterTable(t *testing.T) {
	t.Skip()
	// This is a newly added test in GMS that dolt doesn't support yet
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{"ALTER TABLE t42 ADD COLUMN s varchar(20), drop check check1"})
	defer h.Close()
	enginetest.TestAlterTable(t, h)
}

func TestVariables(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunVariableTest(t, h)
}

func TestVariableErrors(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestVariableErrors(t, h)
}

func TestLoadDataPrepared(t *testing.T) {
	t.Skip("feature not supported")
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestLoadDataPrepared(t, h)
}

func TestLoadData(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestLoadData(t, h)
}

func TestLoadDataErrors(t *testing.T) {
	t.Skip()
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestLoadDataErrors(t, h)
}

func TestSelectIntoFile(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestSelectIntoFile(t, h)
}

func TestJsonScripts(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	skippedTests := []string{
		"round-trip into table", // The current Dolt JSON format does not preserve decimals and unsigneds in JSON.
	}
	// TODO: fix this, use a skipping harness
	enginetest.TestJsonScripts(t, h, skippedTests)
}

func TestTriggers(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestTriggers(t, h)
}

func TestRollbackTriggers(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestRollbackTriggers(t, h)
}

func TestStoredProcedures(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunStoredProceduresTest(t, h)
}

func TestDoltStoredProcedures(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltStoredProceduresTest(t, h)
}

func TestDoltStoredProceduresPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltStoredProceduresPreparedTest(t, h)
}

func TestEvents(t *testing.T) {
	t.Skip()
	doltHarness := newDoltgresServerHarness(t)
	defer doltHarness.Close()
	enginetest.TestEvents(t, doltHarness)
}

func TestCallAsOf(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunCallAsOfTest(t, h)
}

func TestLargeJsonObjects(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunLargeJsonObjectsTest(t, harness)
}

func TestTransactions(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunTransactionTests(t, h)
}

func TestBranchTransactions(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunBranchTransactionTest(t, h)
}

func TestMultiDbTransactions(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunMultiDbTransactionsTest(t, h)
}

func TestMultiDbTransactionsPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunMultiDbTransactionsPreparedTest(t, h)
}

func TestConcurrentTransactions(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestConcurrentTransactions(t, h)
}

func TestDoltScripts(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltScriptsTest(t, harness)
}

func TestDoltTempTableScripts(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltTempTableScripts(t, harness)
}

func TestDoltRevisionDbScripts(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltRevisionDbScriptsTest(t, h)
}

func TestDoltRevisionDbScriptsPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltRevisionDbScriptsPreparedTest(t, h)
}

func TestDoltDdlScripts(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltDdlScripts(t, harness)
}

func TestBrokenDdlScripts(t *testing.T) {
	t.Skip()
	for _, script := range denginetest.BrokenDDLScripts {
		t.Skip(script.Name)
	}
}

func TestDescribeTableAsOf(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestScript(t, h, denginetest.DescribeTableAsOfScriptTest)
}

func TestShowCreateTable(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunShowCreateTableTests(t, h)
}

func TestShowCreateTablePrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunShowCreateTablePreparedTests(t, h)
}

func TestViewsWithAsOf(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestScript(t, h, denginetest.ViewsWithAsOfScriptTest)
}

func TestViewsWithAsOfPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestScriptPrepared(t, h, denginetest.ViewsWithAsOfScriptTest)
}

func TestDoltMerge(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltMergeTests(t, h)
}

func TestDoltMergePrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltMergePreparedTests(t, h)
}

func TestDoltRebase(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltRebaseTests(t, h)
}

func TestDoltRebasePrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltRebasePreparedTests(t, h)
}

func TestDoltRevert(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltRevertTests(t, h)
}

func TestDoltRevertPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltRevertPreparedTests(t, h)
}

func TestDoltAutoIncrement(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltAutoIncrementTests(t, h)
}

func TestDoltAutoIncrementPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltAutoIncrementPreparedTests(t, h)
}

func TestDoltConflictsTableNameTable(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltConflictsTableNameTableTests(t, h)
}

// tests new format behavior for keyless merges that create CVs and conflicts
func TestKeylessDoltMergeCVsAndConflicts(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunKelyessDoltMergeCVsAndConflictsTests(t, h)
}

// eventually this will be part of TestDoltMerge
func TestDoltMergeArtifacts(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltMergeArtifacts(t, h)
}

func TestDoltReset(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltResetTest(t, h)
}

func TestDoltGC(t *testing.T) {
	t.Skip()
	for _, script := range denginetest.DoltGC {
		func() {
			h := newDoltgresServerHarness(t)
			defer h.Close()
			enginetest.TestScript(t, h, script)
		}()
	}
}

func TestDoltCheckout(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltCheckoutTests(t, h)
}

func TestDoltCheckoutPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltCheckoutPreparedTests(t, h)
}

func TestDoltBranch(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltBranchTests(t, h)
}

func TestDoltTag(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltTagTests(t, h)
}

func TestDoltRemote(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltRemoteTests(t, h)
}

func TestDoltUndrop(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltUndropTests(t, h)
}

func TestBrokenSystemTableQueries(t *testing.T) {
	t.Skip()

	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.RunQueryTests(t, h, denginetest.BrokenSystemTableQueries)
}

func TestHistorySystemTable(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t).WithParallelism(2)
	denginetest.RunHistorySystemTableTests(t, harness)
}

func TestHistorySystemTablePrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t).WithParallelism(2)
	denginetest.RunHistorySystemTableTestsPrepared(t, harness)
}

func TestBrokenHistorySystemTablePrepared(t *testing.T) {
	t.Skip("test not migrated yet, skipped in Dolt")
}

func TestUnscopedDiffSystemTable(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunUnscopedDiffSystemTableTests(t, h)
}

func TestUnscopedDiffSystemTablePrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunUnscopedDiffSystemTableTestsPrepared(t, h)
}

func TestColumnDiffSystemTable(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunColumnDiffSystemTableTests(t, h)
}

func TestColumnDiffSystemTablePrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunColumnDiffSystemTableTestsPrepared(t, h)
}

func TestStatBranchTests(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunStatBranchTests(t, harness)
}

func TestStatsFunctions(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunStatsFunctionsTest(t, harness)
}

func TestDiffTableFunction(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDiffTableFunctionTests(t, harness)
}

func TestDiffTableFunctionPrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDiffTableFunctionTestsPrepared(t, harness)
}

func TestDiffStatTableFunction(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDiffStatTableFunctionTests(t, harness)
}

func TestDiffStatTableFunctionPrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDiffStatTableFunctionTestsPrepared(t, harness)
}

func TestDiffSummaryTableFunction(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDiffSummaryTableFunctionTests(t, harness)
}

func TestDiffSummaryTableFunctionPrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDiffSummaryTableFunctionTestsPrepared(t, harness)
}

func TestPatchTableFunction(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltPatchTableFunctionTests(t, harness)
}

func TestPatchTableFunctionPrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltPatchTableFunctionTestsPrepared(t, harness)
}

func TestLogTableFunction(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunLogTableFunctionTests(t, harness)
}

func TestLogTableFunctionPrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunLogTableFunctionTestsPrepared(t, harness)
}

func TestDoltReflog(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltReflogTests(t, harness)
}

func TestDoltReflogPrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltReflogTestsPrepared(t, harness)
}

func TestCommitDiffSystemTable(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunCommitDiffSystemTableTests(t, harness)
}

func TestCommitDiffSystemTablePrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunCommitDiffSystemTableTestsPrepared(t, harness)
}

func TestDiffSystemTable(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltDiffSystemTableTests(t, h)
}

func TestDiffSystemTablePrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltDiffSystemTableTestsPrepared(t, h)
}

func TestSchemaDiffTableFunction(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunSchemaDiffTableFunctionTests(t, harness)
}

func TestSchemaDiffTableFunctionPrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunSchemaDiffTableFunctionTestsPrepared(t, harness)
}

func TestDoltDatabaseCollationDiffs(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltDatabaseCollationDiffsTests(t, harness)
}

func TestQueryDiff(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunQueryDiffTests(t, harness)
}

func TestSystemTableIndexes(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunSystemTableIndexesTests(t, harness)
}

func TestSystemTableIndexesPrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunSystemTableIndexesTestsPrepared(t, harness)
}

func TestSystemTableFunctionIndexes(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunSystemTableFunctionIndexesTests(t, harness)
}

func TestSystemTableFunctionIndexesPrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunSystemTableFunctionIndexesTestsPrepared(t, harness)
}

func TestReadOnlyDatabases(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestReadOnlyDatabases(t, h)
}

func TestAddDropPks(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestAddDropPks(t, h)
}

func TestAddAutoIncrementColumn(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunAddAutoIncrementColumnTests(t, h)
}

func TestNullRanges(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestNullRanges(t, h)
}

func TestPersist(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
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
	t.Skip("Port equivalent test from Dolt")
}

func TestDoltCherryPick(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltCherryPickTests(t, harness)
}

func TestDoltCherryPickPrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltCherryPickTestsPrepared(t, harness)
}

func TestDoltCommit(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltCommitTests(t, harness)
}

func TestDoltCommitPrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltCommitTestsPrepared(t, harness)
}

func TestQueriesPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestQueriesPrepared(t, h)
}

func TestStatsHistograms(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunStatsHistogramTests(t, h)
}

// TestStatsIO force a provider reload in-between setup and assertions that
// forces a round trip of the statistics table before inspecting values.
func TestStatsIO(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunStatsIOTests(t, h)
}

func TestJoinStats(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunJoinStatsTests(t, h)
}

func TestStatisticIndexes(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestStatisticIndexFilters(t, h)
}

func TestSpatialQueriesPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)

	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestSpatialQueriesPrepared(t, h)
}

func TestPreparedStatistics(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunPreparedStatisticsTests(t, h)
}

func TestVersionedQueriesPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunVersionedQueriesPreparedTests(t, h)
}

func TestInfoSchemaPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestInfoSchemaPrepared(t, h)
}

func TestUpdateQueriesPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestUpdateQueriesPrepared(t, h)
}

func TestInsertQueriesPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestInsertQueriesPrepared(t, h)
}

func TestReplaceQueriesPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestReplaceQueriesPrepared(t, h)
}

func TestDeleteQueriesPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestDeleteQueriesPrepared(t, h)
}

func TestScriptsPrepared(t *testing.T) {
	t.Skip()
	skipped := newFormatSkippedScripts
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t).WithSkippedQueries(skipped)
	defer h.Close()
	enginetest.TestScriptsPrepared(t, h)
}

func TestInsertScriptsPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestInsertScriptsPrepared(t, h)
}

func TestComplexIndexQueriesPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestComplexIndexQueriesPrepared(t, h)
}

func TestJsonScriptsPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	skippedTests := []string{
		"round-trip into table", // The current Dolt JSON format does not preserve decimals and unsigneds in JSON.
	}
	enginetest.TestJsonScriptsPrepared(t, h, skippedTests)
}

func TestCreateCheckConstraintsScriptsPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestCreateCheckConstraintsScriptsPrepared(t, h)
}

func TestInsertIgnoreScriptsPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestInsertIgnoreScriptsPrepared(t, h)
}

func TestInsertErrorScriptsPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	h = h.WithSkippedQueries([]string{
		"create table bad (vb varbinary(65535))",
		"insert into bad values (repeat('0', 65536))",
	})
	enginetest.TestInsertErrorScriptsPrepared(t, h)
}

func TestViewsPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestViewsPrepared(t, h)
}

func TestVersionedViewsPrepared(t *testing.T) {
	t.Skip()
	t.Skip("not supported for prepareds")
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestVersionedViewsPrepared(t, h)
}

func TestShowTableStatusPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestShowTableStatusPrepared(t, h)
}

func TestPrepared(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestPrepared(t, h)
}

func TestDoltPreparedScripts(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	// TODO
	// DoltPreparedScripts(t, h)
}

func TestPreparedInsert(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestPreparedInsert(t, h)
}

func TestPreparedStatements(t *testing.T) {
	t.Skip()
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestPreparedStatements(t, h)
}

func TestCharsetCollationEngine(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestCharsetCollationEngine(t, h)
}

func TestCharsetCollationWire(t *testing.T) {
	t.Skip("port test from Dolt")
}

func TestDatabaseCollationWire(t *testing.T) {
	t.Skip("port test from Dolt")
}

func TestAddDropPrimaryKeys(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunAddDropPrimaryKeysTests(t, harness)
}

func TestDoltVerifyConstraints(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltVerifyConstraintsTests(t, harness)
}

func TestDoltStorageFormat(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltStorageFormatTests(t, h)
}

func TestDoltStorageFormatPrepared(t *testing.T) {
	t.Skip()
	expectedFormatString := "NEW ( __DOLT__ )"
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestPreparedQuery(t, h, "SELECT dolt_storage_format()", []sql.Row{{expectedFormatString}}, nil)
}

func TestThreeWayMergeWithSchemaChangeScripts(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)

	denginetest.RunThreeWayMergeWithSchemaChangeScripts(t, h)
}

func TestThreeWayMergeWithSchemaChangeScriptsPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)

	denginetest.RunThreeWayMergeWithSchemaChangeScriptsPrepared(t, h)
}

// If CREATE DATABASE has an error within the DatabaseProvider, it should not
// leave behind intermediate filesystem state.
func TestCreateDatabaseErrorCleansUp(t *testing.T) {
	t.Skip("port test from Dolt")
}

// TestStatsAutoRefreshConcurrency tests some common concurrent patterns that stats
// refresh is subject to -- namely reading/writing the stats objects in (1) DML statements
// (2) auto refresh threads, and (3) manual ANALYZE statements.
// todo: the dolt_stat functions should be concurrency tested
func TestStatsAutoRefreshConcurrency(t *testing.T) {
	t.Skip("port test from Dolt")
}

var newFormatSkippedScripts = []string{
	// Different query plans
	"Partial indexes are used and return the expected result",
	"Multiple indexes on the same columns in a different order",
}

func skipPreparedTests(t *testing.T) {
	if skipPrepared {
		t.Skip("skip prepared")
	}
}
