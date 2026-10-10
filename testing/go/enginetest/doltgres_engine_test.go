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
	"os"
	"runtime"
	"slices"
	"testing"

	denginetest "github.com/dolthub/dolt/go/libraries/doltcore/sqle/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest/queries"
	"github.com/dolthub/go-mysql-server/enginetest/scriptgen/setup"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/dolt/go/libraries/doltcore/dtestutils"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema/typeinfo"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/utils/config"
)

func init() {
	sqle.MinRowsPerPartition = 8
	sqle.MaxRowsPerPartition = 1024
}

func TestQueries(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestQueries(t, h)
}

func TestSingleWriteQuery(t *testing.T) {
	// t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()

	h.Setup(setup.MydbData, setup.AutoincrementData)

	test := queries.WriteQueryTest{
		WriteQuery:          "INSERT INTO auto_increment_tbl (c0) values (44)",
		ExpectedWriteResult: []sql.Row{{types.OkResult{RowsAffected: 1, InsertID: 4}}},
		SelectQuery:         "SELECT * FROM auto_increment_tbl ORDER BY pk",
		ExpectedSelect: []sql.Row{
			{1, 11},
			{2, 22},
			{3, 33},
			{4, 44},
		},
	}

	enginetest.RunWriteQueryTest(t, h, test)
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

	enginetest.TestQuery(t, harness, engine, test)
}

func TestSchemaOverrides(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunSchemaOverridesTest(t, harness)
}

// Convenience test for debugging a single query. Unskip and set to the desired query.
func TestSingleScript(t *testing.T) {
	t.Skip()

	var scripts = []queries.ScriptTest{}

	for _, script := range scripts {
		func() {
			harness := newDoltgresServerHarness(t)
			defer harness.Close()
			// harness.Setup(setup.MydbData, setup.MytableData)

			engine, err := harness.NewEngine(t)
			if err != nil {
				panic(err)
			}
			engine.EngineAnalyzer().Debug = true
			engine.EngineAnalyzer().Verbose = true

			enginetest.TestScriptWithEngine(t, engine, harness, script)
		}()
	}
}

func TestAutoIncrementTrackerLockMode(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunAutoIncrementTrackerLockModeTest(t, harness)
}

func TestVersionedQueries(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		// MySQL-only `cast(... as signed)` syntax — postgres uses `cast(... as integer)`.
		// These views aren't referenced by the assertions, so dropping them from
		// setup is harmless.
		"cast(RIGHT(s, 1) as signed)",
		// AS OF tests on these complex views; they depend on views we can't create.
		"myview4", "myview5",
		// SHOW CREATE TABLE AS OF — doltgres doesn't yet expose the AS OF clause
		// to SHOW CREATE TABLE (it errors with "at or near 'as'").
		"SHOW CREATE TABLE myhistorytable as of",
		// SHOW TABLES AS OF and DESCRIBE AS OF likewise unsupported.
		"SHOW TABLES AS OF",
		"SHOW TABLES FROM mydb AS OF",
		"DESCRIBE myhistorytable AS OF",
	})
	defer h.Close()

	denginetest.RunVersionedQueriesTest(t, h)
}

func TestAnsiQuotesSqlMode(t *testing.T) {
	t.Skip()
	enginetest.TestAnsiQuotesSqlMode(t, newDoltgresServerHarness(t))
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
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"SELECT s as Date, SUM(i) TimeStamp FROM mytable group by 1 order by 2", // ERROR: at or near "timestamp": syntax error
		"SELECT 1 as a, (select a) as b from dual",                              // table not found: dual
	})
	defer h.Close()
	enginetest.TestColumnAliases(t, h)
}

func TestOrderByGroupBy(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"Group by with decimal columns", // syntax error
		"Validation for use of non-aggregated columns with implicit grouping of all rows", // bad error matching
		"group by with any_value()",   // @@ vars not supported
		"group by with strict errors", // @@ vars not supported
	})
	defer h.Close()
	enginetest.TestOrderByGroupBy(t, h)
}

func TestAmbiguousColumnResolution(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestAmbiguousColumnResolution(t, h)
}

func TestInsertInto(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"values and rows", // MySQL VALUES ROW syntax; previously in the skipped SQL logic suite
		"with t (i,f) as (select 4,'fourth row' from dual) insert into mytable select i,f from t",                                                 // WITH unsupported syntax
		"with recursive t (i,f) as (select 4,4 from dual union all select i + 1, i + 1 from t where i < 5) insert into mytable select i,f from t", // WITH unsupported syntax
		"issue 6675: on duplicate rearranged getfield indexes from select source",                                                                 // panic
		"Insert on duplicate key references table in subquery",                                                                                    // bad translation?
		"Insert on duplicate key references table in aliased subquery",                                                                            // bad translation?
		"Insert on duplicate key references table in cte",                                                                                         // CTE not supported
		"insert on duplicate key with incorrect row alias",                                                                                        // column "c" could not be found in any table in scope
		"insert on duplicate key update errors",                                                                                                   // failing
		"INSERT INTO ... SELECT with TEXT types",                                                                                                  // typecasts needed
		"insert duplicate key doesn't prevent other updates, autocommit off",                                                                      // MySQL semantics: postgres aborts a transaction on any error, rejecting further statements
	})
	defer h.Close()
	enginetest.TestInsertInto(t, h)
}

func TestInsertIgnoreInto(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"Test that INSERT IGNORE properly addresses data conversion", // PostgreSQL strict typing rejects MySQL coercions
		"Test that INSERT IGNORE with Non nullable columns works",    // PostgreSQL does not ignore NOT NULL violations
		"Insert Ignore works correctly with ON DUPLICATE UPDATE",     // PostgreSQL strict typing rejects MySQL coercions
		"issue 8611: insert ignore on enum type column",              // enums not supported
	})
	defer h.Close()
	enginetest.TestInsertIgnoreInto(t, h)
}

func TestInsertDuplicateKeyKeyless(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"select c1, c2, c3 from t order by c1, c2, c3", // expects MySQL's NULLs-first ordering
	})
	defer h.Close()
	enginetest.TestInsertDuplicateKeyKeyless(t, h)
}

func TestIgnoreIntoWithDuplicateUniqueKeyKeyless(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"UPDATE IGNORE keyless tables and secondary indexes", // UPDATE IGNORE rewrite doesn't perform the actual update
	})
	defer h.Close()
	enginetest.TestIgnoreIntoWithDuplicateUniqueKeyKeyless(t, h)
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
	// REPLACE INTO is rewritten to INSERT ... ON CONFLICT DO UPDATE in the
	// converter. The VALUES form without a column list still fails because the
	// converter needs the column list to build the SET clause.
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"REPLACE INTO mytable VALUES (1, 'first row')",      // no-column REPLACE; bails to raw syntax
		"REPLACE INTO mytable VALUES (1, 'new row same i')", // no-column REPLACE; bails to raw syntax
		"REPLACE INTO mytable VALUES (999, 'x')",            // no-column REPLACE; bails to raw syntax
	})
	defer h.Close()
	enginetest.TestReplaceInto(t, h)
}

func TestReplaceIntoErrors(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestReplaceIntoErrors(t, h)
}

func TestUpdate(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"UPDATE mytable SET s = _binary 'updated' WHERE i = 3;",         // _binary not supported
		"UPDATE mytable SET s = 'updated' ORDER BY i LIMIT 1 OFFSET 1;", // offset not supported (limit isn't selected in vanilla postgres but is in the cockroach grammar)
		// TODO: Postgres supports update joins but with a very different syntax, and some join types are not supported
		`UPDATE one_pk INNER JOIN two_pk on one_pk.pk = two_pk.pk1 SET two_pk.c1 = two_pk.c1 + 1`,
		"UPDATE mytable INNER JOIN one_pk ON mytable.i = one_pk.c5 SET mytable.i = mytable.i * 10",
		`UPDATE one_pk INNER JOIN two_pk on one_pk.pk = two_pk.pk1 INNER JOIN othertable on othertable.i2 = two_pk.pk2 SET one_pk.c1 = one_pk.c1 + 1`,
		`UPDATE one_pk INNER JOIN (SELECT * FROM two_pk order by pk1, pk2) as t2 on one_pk.pk = t2.pk1 SET one_pk.c1 = t2.c1 + 1 where one_pk.pk < 1`,
		`UPDATE one_pk INNER JOIN two_pk on one_pk.pk = two_pk.pk1 SET one_pk.c1 = one_pk.c1 + 1`,
		`update mytable h join mytable on h.i = mytable.i and h.s <> mytable.s set h.i = mytable.i+1;`,
		`UPDATE othertable CROSS JOIN tabletest set othertable.i2 = othertable.i2 * 10`,                                                                                              // cross join
		`UPDATE tabletest cross join tabletest as t2 set tabletest.i = tabletest.i * 10`,                                                                                             // cross join
		`UPDATE othertable cross join tabletest set tabletest.i = tabletest.i * 10`,                                                                                                  // cross join
		`UPDATE one_pk INNER JOIN two_pk on one_pk.pk = two_pk.pk1 INNER JOIN two_pk a1 on one_pk.pk = two_pk.pk2 SET two_pk.c1 = two_pk.c1 + 1`,                                     // cross join
		`UPDATE othertable INNER JOIN tabletest on othertable.i2=3 and tabletest.i=3 SET othertable.s2 = 'fourth'`,                                                                   // cross join
		`UPDATE tabletest cross join tabletest as t2 set t2.i = t2.i * 10`,                                                                                                           // cross join
		`UPDATE othertable LEFT JOIN tabletest on othertable.i2=3 and tabletest.i=3 SET othertable.s2 = 'fourth'`,                                                                    // left join
		`UPDATE othertable LEFT JOIN tabletest on othertable.i2=3 and tabletest.i=3 SET tabletest.s = 'fourth row', tabletest.i = tabletest.i + 1`,                                   // left join
		`UPDATE othertable LEFT JOIN tabletest t3 on othertable.i2=3 and t3.i=3 SET t3.s = 'fourth row', t3.i = t3.i + 1`,                                                            // left join
		`UPDATE othertable LEFT JOIN tabletest on othertable.i2=3 and tabletest.i=3 LEFT JOIN one_pk on othertable.i2 = one_pk.pk SET one_pk.c1 = one_pk.c1 + 1`,                     // left join
		`UPDATE othertable LEFT JOIN tabletest on othertable.i2=3 and tabletest.i=3 LEFT JOIN one_pk on othertable.i2 = one_pk.pk SET one_pk.c1 = one_pk.c1 + 1 where one_pk.pk > 4`, // left join
		`UPDATE othertable LEFT JOIN tabletest on othertable.i2=3 and tabletest.i=3 LEFT JOIN one_pk on othertable.i2 = 1 and one_pk.pk = 1 SET one_pk.c1 = one_pk.c1 + 1`,           // left join
		`UPDATE othertable RIGHT JOIN tabletest on othertable.i2=3 and tabletest.i=3 SET othertable.s2 = 'fourth'`,                                                                   // right join
		`UPDATE othertable RIGHT JOIN tabletest on othertable.i2=3 and tabletest.i=3 SET othertable.i2 = othertable.i2 + 1`,                                                          // right join
		`UPDATE othertable LEFT JOIN tabletest on othertable.i2=tabletest.i RIGHT JOIN one_pk on othertable.i2 = 1 and one_pk.pk = 1 SET tabletest.s = 'updated';`,                   // right join
		`UPDATE IGNORE one_pk INNER JOIN two_pk on one_pk.pk = two_pk.pk1 SET two_pk.c1 = two_pk.c1 + 1`,
		`UPDATE IGNORE one_pk JOIN one_pk one_pk2 on one_pk.pk = one_pk2.pk SET one_pk.pk = 10`,
		`with t (n) as (select (1) from dual) UPDATE mytable set s = concat('updated ', i) where i in (select n from t)`, // with not supported
		`with recursive t (n) as (select (1) from dual union all select n + 1 from t where n < 2) UPDATE mytable set s = concat('updated ', i) where i in (select n from t)`,
	})
	defer h.Close()
	enginetest.TestUpdate(t, h)
}

func TestUpdateFloatAssignments(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	h.Setup(setup.MydbData)
	enginetest.TestScript(t, h, queries.ScriptTest{
		Name: "floating-point assignments read the original row",
		SetUpScript: []string{
			"CREATE TABLE floattable (i INT PRIMARY KEY, f32 REAL, f64 DOUBLE PRECISION)",
			"INSERT INTO floattable VALUES (2, 1.5, 1.5), (3, 1.5, 1.5)",
			"UPDATE floattable SET f32 = f32 + f32, f64 = f32 * f64 WHERE i = 2;",
			"UPDATE floattable SET f32 = f32 + f32, f64 = (f32 + f32) * f64 WHERE i = 3;",
		},
		Assertions: []queries.ScriptTestAssertion{
			{
				Query: "SELECT * FROM floattable ORDER BY i",
				Expected: []sql.Row{
					// Both assignments use f32's original value of 1.5.
					{int64(2), float32(3.0), float64(2.25)},
					// Doubling must be explicit to produce the MySQL test's 4.5.
					{int64(3), float32(3.0), float64(4.5)},
				},
			},
		},
	})
}

func TestUpdateErrors(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"try updating string that is too long",  // works but error message doesn't match
		"UPDATE mytable SET s = 'hi' LIMIT -1;", // unsupported syntax
	})
	defer h.Close()
	enginetest.TestUpdateErrors(t, h)
}

func TestDeleteFrom(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"DELETE FROM mytable ORDER BY i DESC LIMIT 1 OFFSET 1;", // offset is unsupported syntax
		"with t (n) as (select (1) from dual) delete from mytable where i in (select n from t)",
		"with recursive t (n) as (select (1) from dual union all select n + 1 from t where n < 2) delete from mytable where i in (select n from t)",
	})
	defer h.Close()

	// We've inlined part of engineTest.TestDeleteFrom here because that method tests many queries for join deletions
	// that would be tedious to write out as skips
	h.Setup(setup.MydbData, setup.MytableData, setup.TabletestData)
	t.Run("Delete from single table", func(t *testing.T) {
		for _, tt := range queries.DeleteTests {
			enginetest.RunWriteQueryTest(t, h, tt)
		}
	})
}

func TestDeleteFromErrors(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()

	// These tests are overspecified to mysql-specific errors and include some syntax we don't support, so we redefine
	// the subset we're interested in checking here
	h.Setup(setup.MydbData, setup.MytableData, setup.TabletestData)
	deleteScripts := []queries.ScriptTest{
		{
			Name: "DELETE FROM error cases",
			Assertions: []queries.ScriptTestAssertion{
				{
					Query:          "DELETE FROM invalidtable WHERE x < 1;",
					ExpectedErrStr: "table not found: invalidtable",
				},
				{
					Query:          "DELETE FROM mytable WHERE z = 'dne';",
					ExpectedErrStr: "column \"z\" could not be found in any table in scope",
				},
				{
					Query:          "DELETE FROM mytable LIMIT -1;",
					ExpectedErrStr: "LIMIT must be greater than or equal to 0",
				},
				{
					Query:          "DELETE mytable WHERE i = 1;",
					ExpectedErrStr: "syntax error",
				},
				{
					Query:          "DELETE FROM (SELECT * FROM mytable) mytable WHERE i = 1;",
					ExpectedErrStr: "syntax error",
				},
			},
		},
	}
	for _, tt := range deleteScripts {
		enginetest.TestScript(t, h, tt)
	}
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

func TestSpatialIndexScripts(t *testing.T) {
	t.Skip()
	enginetest.TestSpatialIndexScripts(t, newDoltgresServerHarness(t))
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

func TestAggregationScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"sum() and avg() on non-DECIMAL type column returns the DOUBLE type result", // MySQL-specific: MySQL widens FLOAT to DOUBLE for SUM; Postgres sum(real) stays real
		"sum() and avg() on DECIMAL type column returns the DECIMAL type result",    // MySQL-specific: MySQL truncates AVG's decimal scale; Postgres numeric division keeps full precision
	})
	defer h.Close()
	enginetest.TestAggregationScripts(t, h)
}

func TestAutoIncrementScripts(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestAutoIncrementScripts(t, h)
}

func TestConversionsScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"Handle hex number to binary conversion", // ERROR: can't convert 0x7ED0599B to decimal: exponent is not numeric
	})
	defer h.Close()
	enginetest.TestConversionsScripts(t, h)
}

func TestDatabaseDefinitionsScripts(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestDatabaseDefinitionsScripts(t, h)
}

func TestDeleteScripts(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestDeleteScripts(t, h)
}

func TestDescendingIndexesScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"descending index columns",                 // MySQL index prefix syntax (c(10) DESC)
		"descending index lookups and ordering",    // MySQL NULLs-first ascending order
		"descending unique indexes",                // MySQL REPLACE INTO and ON DUPLICATE KEY UPDATE
		"descending prefix and expression indexes", // MySQL index prefix syntax (s(3) DESC)
		"descending index on a keyless table",      // MySQL NULLs-first ascending order
		"descending indexes backing foreign keys",  // MySQL foreign key error types
		"descending indexes on assorted types",     // MySQL ENUM and DATETIME columns
	})
	defer h.Close()
	enginetest.TestDescendingIndexesScripts(t, h)
}

func TestEnumsAndSetsScripts(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestEnumsAndSetsScripts(t, h)
}

func TestExpressionsScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"(x between y and z), (x between x and z)", // expects MySQL's NULLs-first ordering
		"coalesce with system types",               // unsupported
	})
	defer h.Close()
	enginetest.TestExpressionsScripts(t, h)
}

func TestIndexKeyTypesScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"binary type primary key", // ERROR: blob/text column 'b' used in key specification without a key length
		"varbinary primary key",   // ERROR: blob/text column 'b' used in key specification without a key length
		"varchar primary key",     // literal values longer than the key length returns incorrect results for some queries
	})
	defer h.Close()
	enginetest.TestIndexKeyTypesScripts(t, h)
}

func TestInsertIgnoreRegressionScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"INSERT IGNORE throws an error when json is badly formatted", // error messages don't match
	})
	defer h.Close()
	enginetest.TestInsertIgnoreRegressionScripts(t, h)
}

func TestJoinsScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"filter pushdown through join uppercase name", // syntax error (join without on)
		"join index lookups do not handle filters",    // need a different join syntax (no ON clause not supported in postgres)
	})
	defer h.Close()
	enginetest.TestJoinsScripts(t, h)
}

func TestNameResolutionScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"db1.``.i > 0",               // Multi-db Aliasing: MySQL-only empty-backtick ref
		"join db2.t2 order by",       // Multi-db Aliasing: MySQL implicit-cross-join (no ON)
		"join db2.t2 group by",       // Multi-db Aliasing: MySQL implicit-cross-join (no ON)
		"join db2.t1 b order by a.i", // Multi-db Aliasing: MySQL implicit-cross-join (no ON)
	})
	defer h.Close()
	enginetest.TestNameResolutionScripts(t, h)
}

func TestNumericScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"Ensure proper DECIMAL support (found by fuzzer)",                                                       // unsupported type: SET
		"arithmetic bit operations on int, float and decimal types",                                             // the power operator is not yet supported
		"division and int division operation on negative, small and big value for decimal type column of table", // numeric keys broken
	})
	defer h.Close()
	enginetest.TestNumericScripts(t, h)
}

func TestOrderingScripts(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestOrderingScripts(t, h)
}

func TestPrimaryKeysScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"recreate primary key rebuilds secondary indexes",      // currently no way to drop primary key in doltgres
		"insert into t1 (a, b) values ('1234567890', '12345')", // different error message
		"insert into t2 (a, b) values ('1234567890', '12345')", // different error message
	})
	defer h.Close()
	enginetest.TestPrimaryKeysScripts(t, h)
}

func TestSessionResultsScripts(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestSessionResultsScripts(t, h)
}

func TestSetOperationsScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"WITH RECURSIVE\n" +
			"    rt (foo) AS (\n" +
			"        SELECT 1 as foo\n" +
			"        UNION ALL\n" +
			"        SELECT foo + 1 as foo FROM rt WHERE foo < 5\n" +
			"    ),\n" +
			"        ladder (depth, foo) AS (\n" +
			"        SELECT 1 as depth, NULL as foo from rt\n" +
			"        UNION ALL\n" +
			"        SELECT ladder.depth + 1 as depth, rt.foo\n" +
			"        FROM ladder JOIN rt WHERE ladder.foo = rt.foo\n" +
			"    )\n" +
			"SELECT * FROM ladder;", // syntax error
	})
	defer h.Close()
	enginetest.TestSetOperationsScripts(t, h)
}

func TestStatisticsScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"histogram bucket merging error for implementor buckets", // unsupported "with recursive" syntax
	})
	defer h.Close()
	enginetest.TestStatisticsScripts(t, h)
}

func TestStringFunctionsScripts(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestStringFunctionsScripts(t, h)
}

func TestSubqueriesScripts(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestSubqueriesScripts(t, h)
}

func TestTableDefinitionsScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"CREATE TABLE SELECT Queries",                         // ERROR: TableCopier only accepts CreateTable or TableNode as the destination
		"Show create table with various keys and constraints", // FK adds explicit ON DELETE/UPDATE RESTRICT; CHECK constraints leak backticks; timestamp(6) loses precision
	})
	defer h.Close()
	enginetest.TestTableDefinitionsScripts(t, h)
}

func TestTemporalScripts(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestTemporalScripts(t, h)
}

func TestTransactionsScripts(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestTransactionsScripts(t, h)
}

func TestTupleComparisonsScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"select count(*) from t where (f in (null, cast(0.8 as float)));", // incorrect result, needs a fix
		"mismatched collation using hash in tuples",                       // ERROR: plan is not resolved because of node '*plan.Project'
	})
	defer h.Close()
	enginetest.TestTupleComparisonsScripts(t, h)
}

func TestUpdateJoinsScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"issue 7958, update join uppercase table name validation", // update join syntax not supported
		"Dolt issue 7957, update join matched rows",               // update join syntax not supported
		"update join with update trigger",                         // update join syntax not supported (also catches with-trigger variants by substring)
		"update with left join with some missing rows",            // need to translate update joins
	})
	defer h.Close()
	enginetest.TestUpdateJoinsScripts(t, h)
}

func TestVariablesScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"validate_password_strength and validate_password.length",             // unsupported
		"validate_password_strength and validate_password.number_count",       // unsupported
		"validate_password_strength and validate_password.mixed_case_count",   // unsupported
		"validate_password_strength and validate_password.special_char_count", // unsupported
	})
	defer h.Close()
	enginetest.TestVariablesScripts(t, h)
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

// TestJSONTableQueries runs the canonical test queries against a single threaded index enabled harness.
func TestJSONTableQueries(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestJSONTableQueries(t, h)
}

// TestJSONTableScripts runs the canonical test queries against a single threaded index enabled harness.
func TestJSONTableScripts(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestJSONTableScripts(t, h)
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

func TestPkOrdinalsDDL(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestPkOrdinalsDDL(t, h)
}

func TestPkOrdinalsDML(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestPkOrdinalsDML(t, h)
}

func TestDropTableWarning(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	h.Setup(setup.MydbData)
	enginetest.TestDropTableWarnings(t, h, false)
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
	// MySQL index metadata and error behavior differ.
	skippedScripts := []string{
		"show create table with duplicate primary key", // auto-generated constraint names differ
		"unique key duplicate key update",
		"multiple indexes over same set of columns",
		"secondary index errors",
		"indexes and if exists",
		"Test oversized primary-key lookups",
	}
	h := newDoltgresServerHarness(t)
	defer h.Close()

	h.Setup(setup.MydbData)
	for _, script := range queries.IndexQueries {
		if slices.Contains(skippedScripts, script.Name) {
			t.Run(script.Name, func(t *testing.T) { t.Skip("Doltgres does not yet pass this script") })
			continue
		}

		enginetest.TestScript(t, h, script)
	}
}

func TestIndexedExpressions(t *testing.T) {
	harness := newDoltgresServerHarness(t)
	defer harness.Close()
	enginetest.TestIndexedExpressions(t, harness)
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
	// MySQL foreign-key syntax, metadata, or error expectations differ.
	skippedScripts := []string{
		"Delayed foreign key resolution: update",
		"Delayed foreign key resolution: delete",
		"Delayed foreign key resolution insert",
		"Delayed foreign key still does some validation",
		"Delayed foreign key resolution resetting FOREIGN_KEY_CHECKS",
		"DROP TABLE with FOREIGN_KEY_CHECKS=0",
		"Reordered foreign key columns do match",
		"Self-referential foreign key is not case sensitive",
		"INSERT on DUPLICATE correctly works with FKs",
		"Keyless CASCADE over three tables",
		"Referenced index includes implicit primary key columns",
		"rename foreign key constraints",
		"rename check constraints",
		"foreign key naming",
		"Naming automatically created FK indexes",
		"partial foreign key update",
		"ON UPDATE CASCADE maintains an index over a virtual column",
		"ON UPDATE CASCADE recomputes chained virtual columns",
		"ON DELETE CASCADE maintains an index over a virtual column, self-referential",
		"ON DELETE CASCADE maintains an index over a virtual column between stored columns",
		"CREATE TABLE Type Mismatch",
		"CREATE TABLE Disallow TEXT/BLOB",
		"Test foreign keys with spatial parent columns",
		"ALTER TABLE Single Named FOREIGN KEY",
		"CREATE TABLE Single Named FOREIGN KEY",
		"Inline column REFERENCES creates an enforced foreign key",
		"indexes with prefix lengths are ignored for foreign keys",
		"CREATE TABLE Name Collision",
		"SET DEFAULT not supported",
		"ALTER TABLE Foreign Key Name Collision",
		"ALTER TABLE DROP FOREIGN KEY",
		"RENAME TABLE",
		"RENAME TABLE with autogenerated FK name",
		"RENAME TABLE with primary key indexes",
		"Indexes used by foreign keys can't be dropped",
		"ALTER TABLE RENAME COLUMN",
		"DROP COLUMN parent",
		"DROP COLUMN child",
		"SQL CASCADE",
		"SQL SET NULL",
		"SQL RESTRICT",
		"Multi-table DELETE FROM JOIN with multiple foreign keys",
		"Single-table DELETE FROM JOIN with multiple foreign keys",
		"SQL no reference options",
		"SQL INSERT multiple keys violates only one",
		"Self-referential same column(s)",
		"Self-referential child column follows parent RESTRICT",
		"Self-referential child column follows parent CASCADE",
		"Self-referential child column follows parent SET NULL",
		"Self-referential delete cascade depth limit",
		"INSERT IGNORE INTO works correctly with foreign key violations",
		"ALTER TABLE ADD CONSTRAINT for different database",
		"Creating a foreign key on a table with an unsupported type works", // POINT values are not supported
	}
	h := newDoltgresServerHarness(t)
	defer h.Close()

	for _, suite := range []struct {
		scripts   []queries.ScriptTest
		setupData [][]setup.SetupScript
	}{
		{
			scripts:   queries.ForeignKeyTests,
			setupData: [][]setup.SetupScript{setup.MydbData, setup.Parent_childData},
		},
		{scripts: queries.ForeignKeyTypeTests, setupData: [][]setup.SetupScript{setup.MydbData}},
		{scripts: queries.ForeignKeyResolutionTests, setupData: [][]setup.SetupScript{setup.MydbData}},
	} {
		h.Setup(suite.setupData...)
		for _, script := range suite.scripts {
			if slices.Contains(skippedScripts, script.Name) {
				t.Run(script.Name, func(t *testing.T) { t.Skip("Doltgres does not yet pass this script") })
				continue
			}

			enginetest.TestScript(t, h, script)
		}
	}
}

func TestForeignKeyBranches(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunForeignKeyBranchesTest(t, h)
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
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestReadOnly(t, h, false /* testStoredProcedures */)
}

func TestViews(t *testing.T) {
	// MySQL view metadata, syntax, and errors differ.
	skippedScripts := []string{
		"can't create table with same name as existing view", // Doltgres needs to return a different error message
		"can't create view with same name as existing table", // different error message
		"existing views",
		"multi database view",
		"view of join with projections",
		"view with explicit column list renames literal columns",
		"view with explicit column list supports various literal and expression types",
		"view with numeric column name supports dotted and backtick access",
		"check view with escaped strings",
		"show view",
		"views with defaults",
		"SHOW CREATE VIEW returns stored definition regardless of underlying object state",
	}
	h := newDoltgresServerHarness(t)
	defer h.Close()

	h.Setup(setup.MydbData)
	for _, script := range queries.ViewScripts {
		if slices.Contains(skippedScripts, script.Name) {
			t.Run(script.Name, func(t *testing.T) { t.Skip("Doltgres does not yet pass this script") })
			continue
		}

		enginetest.TestScript(t, h, script)
	}
}

func TestBranchViews(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunBranchViewsTest(t, h)
}

func TestVersionedViews(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunVersionedViewsTest(t, h)
}

func TestWindowFunctions(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"select 1 as a, 'x' as a",     // duplicate derived column names are a MySQL-only error
		"t(a, a)",                     // duplicate derived column names are a MySQL-only error
		"format with window function", // FORMAT(X, D, locale) is MySQL's number formatter
	})
	defer h.Close()
	enginetest.TestWindowFunctions(t, h)
}

func TestWindowRowFrames(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestWindowRowFrames(t, h)
}

func TestWindowRangeFrames(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		// MySQL's unquoted numeric interval literal ("interval 1 DAY") isn't valid Postgres grammar
		// (Postgres requires a quoted quantity, e.g. "interval '1' DAY"). The equivalent quoted-quantity
		// form is exercised elsewhere in this same test (table c), so coverage isn't lost.
		"range between interval 2 DAY preceding and interval 1 DAY preceding",
		"range between interval 1 DAY preceding and interval 1 DAY following",
		"range between interval 1 DAY following and interval 2 DAY following",
		"range interval 1 DAY preceding",
		"range between interval 1 DAY preceding and current row",
		"range between interval 1 DAY preceding and unbounded following",
		"range between unbounded preceding and interval 1 DAY following",

		// Postgres's parser strictly validates interval literal syntax at parse time and correctly
		// rejects "interval 'e' DAY" (not a valid unit/quantity) with a syntax error; MySQL parses
		// malformed interval strings leniently, defaulting to 0. This is a real, intentional dialect
		// difference (correct Postgres behavior), not a bug to fix.
		"range interval 'e' DAY preceding",
	})
	defer h.Close()
	enginetest.TestWindowRangeFrames(t, h)
}

func TestNamedWindows(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestNamedWindows(t, h)
}

func TestNaturalJoin(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestNaturalJoin(t, h)
}

func TestNaturalJoinDisjoint(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestNaturalJoinDisjoint(t, h)
}

func TestInnerNestedInNaturalJoins(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestInnerNestedInNaturalJoins(t, h)
}

func TestColumnDefaults(t *testing.T) {
	// MySQL default syntax and result formatting differ.
	skippedScripts := []string{
		"update columns with default", // broken, see repro in update_test.go
		"preserve now()",              // harness error
		"update join ambiguous default",
		"update join ambiguous generated column",
		"Default expression with function and referenced column",
		"REPLACE INTO with default expression",
		"Add column forward reference to default expression",
		"Add column back reference to default literal",
		"Negative float literal",
		"Column referenced with name change",
		"Column defaults with functions",
		"Invalid literal for column type",
		"Expression contains invalid literal once implicitly converted",
		"BLOB types can define defaults with literals",
		"Other types using NOW/CURRENT_TIMESTAMP literal",
		"Stored procedures are not valid in column default value expressions",
		"Expression contains invalid literal, fails on insertion",
		"Add column after back reference to expression",
		"column default normalization: int column rounds",
		"column default normalization: float column rounds",
		"column default normalization: double quotes",
		"column default normalization: expression string literal",
		"column default normalization: expression int literal",
		"User variables in column defaults are not allowed",
		"System variables in column defaults are not allowed",
		"User variables in generated columns are not allowed",
		"System variables in generated columns are not allowed",
		"User variables in ALTER TABLE ADD COLUMN defaults are not allowed",
		"System variables in ALTER TABLE ADD COLUMN defaults are not allowed",
		"User variables in ALTER TABLE ALTER COLUMN defaults are not allowed",
		"System variables in ALTER TABLE ALTER COLUMN defaults are not allowed",
	}
	h := newDoltgresServerHarness(t)
	defer h.Close()

	h.Setup(setup.MydbData)
	for _, script := range queries.ColumnDefaultTests {
		if slices.Contains(skippedScripts, script.Name) {
			t.Run(script.Name, func(t *testing.T) { t.Skip("Doltgres does not yet pass this script") })
			continue
		}

		enginetest.TestScript(t, h, script)
	}
}

func TestOnUpdateExprScripts(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestOnUpdateExprScripts(t, h)
}

func TestAlterTable(t *testing.T) {
	// MySQL ALTER TABLE syntax, metadata, or error behavior differs.
	skippedScripts := []string{
		"modify set column",
		"ALTER TABLE ... ALTER ADD CHECK / DROP CHECK",
		"ALTER TABLE AUTO INCREMENT no-ops on table with no original auto increment key",
		"Identifier lengths",
		"Prefix index with same columns as another index",
		"Index case-insensitivity",
		"alter column and rename table work within same transaction",
		"alter table comment",
		"preserve enums through alter statements",
		"multi alter with invalid schemas",
		"variety of alter column statements in a single statement",
		"mix of alter column, add and drop constraints in one statement",
		"Error queries",
		"alter table containing column default value expressions",
		"drop check as part of alter block",
		"drop constraint as part of alter block",
		"drop column drops correct check constraint",
		"drop column does not drop when referenced in constraint with other column",
		"drop column prevents foreign key violations",
		"disable keys / enable keys",
		"ALTER TABLE remove AUTO_INCREMENT",
		"add column unique index",
		"add column with inline check constraint definition",
		"multi-alter ddl column errors",
		"ALTER TABLE does not change column collations",
	}
	h := newDoltgresServerHarness(t)
	defer h.Close()

	h.Setup(setup.MydbData, setup.Pk_tablesData)
	for _, script := range queries.AlterTableScripts {
		if slices.Contains(skippedScripts, script.Name) {
			t.Run(script.Name, func(t *testing.T) { t.Skip("Doltgres does not yet pass this script") })
			continue
		}

		enginetest.TestScript(t, h, script)
	}
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
	// MySQL JSON functions, formats, or error expectations differ.
	skippedScripts := []string{
		"types survive round-trip into tables",
		"unsigned tinyint is still unsigned after round-trip into table",
		"JSON_ARRAGG with simple and nested json objects.",
		"JSON -> and ->> operator support",
		"json is ordered correctly",
		"Test consistent JSON object comparisons",
		"json is formatted correctly",
		"json_extract returns missing keys as sql null and handles json null literals correctly",
		"json type value compared with number type value",
		"json bools",
		"Comparisons with JSON values containing non-JSON types",
		"round-trip into table", // The current Dolt JSON format does not preserve decimals and unsigneds in JSON.
		"TextStorage converts to JSON when using dolt wrapper",
		"json_type scripts",
		"large integer values keep precision and ordering in document",
		"json_object preserves types",
		"json_object preserves escaped characters in key and values",
		"json conversion works with escaped characters",
		"json_object with escaped k:v pairs from table",
		"json_value preserves types",
		"JSON_ARRAY properly handles CHAR bind vars", // bind-variable execution is not implemented by the server harness
	}
	h := newDoltgresServerHarness(t)
	defer h.Close()

	h.Setup(setup.MydbData, setup.BlobData)
	for _, script := range queries.JsonScripts {
		if slices.Contains(skippedScripts, script.Name) {
			t.Run(script.Name, func(t *testing.T) { t.Skip("Doltgres does not yet pass this script") })
			continue
		}

		enginetest.TestScript(t, h, script)
	}
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
	// MySQL stored procedure syntax and metadata are not supported.
	skippedScripts := []string{
		"REPEAT with OnceBefore returns first loop evaluation result set",
		"WHILE returns previous loop evaluation result set",
		"Simple SELECT",
		"Multiple SELECTs",
		"IF/ELSE with 1 SELECT at end",
		"IF/ELSE with nested SELECT in branches",
		"REPEAT loop over user variable",
		"WHILE loop over user variable",
		"CASE statements",
		"SELECT with JOIN and table aliases",
		"Nested CALL in IF/ELSE branch",
		"INSERT INTO SELECT doesn't override SELECT",
		"Parameters resolve inside of INSERT",
		"Parameters resolve inside of SELECT UNION",
		"Parameters resolve inside of REPLACE",
		"Parameters resolve inside of INSERT INTO SELECT",
		"Subquery on SET user variable captures parameter",
		"Simple SELECT INTO",
		"Multiple variables in SELECT INTO",
		"SELECT INTO with condition",
		"SELECT INTO with group by, order by and limit",
		"multiple SELECT INTO in begin end block",
		"multiple statement with single SELECT INTO in begin end block",
		"DECLARE variables, proper nesting support",
		"DECLARE multiple variables, same statement",
		"DECLARE variable shadows parameter",
		"DECLARE CONDITION",
		"DECLARE CONDITION nesting priority",
		"FETCH multiple rows",
		"FETCH with multiple opens and closes",
		"issue 7458: proc params as limit values",
		"FETCH captures state at OPEN",
		"FETCH implicitly closes",
		"SQLEXCEPTION declare handler",
		"DECLARE CONTINUE HANDLER",
		"Test cursor continue-handler checksum loops",
		"DECLARE HANDLERs exit according to the block they were declared in",
		"Labeled BEGIN...END",
		"REPEAT runs loop before first evaluation",
		"WHILE runs evaluation before first loop",
		"ITERATE and LEAVE LOOP",
		"ITERATE and LEAVE REPEAT",
		"ITERATE and LEAVE WHILE",
		"Handle setting an uninitialized user variable",
		"Dolt Issue #4980",
		"Conditional expression where body has its own columns",
		"Nested subquery in conditional expression where body has its own columns",
		"Conditional expression with else doesn't have body columns in its scope",
		"HANDLERs ignore variables declared after them",
		"Duplicate parameter names",
		"Duplicate parameter names mixed casing",
		"DECLARE CONDITION wrong positions",
		"DECLARE CONDITION duplicate name",
		"SIGNAL references condition name for MySQL error code",
		"SIGNAL non-existent condition name",
		"Duplicate procedure name",
		"Broken procedure shouldn't break other procedures",
		"DECLARE name duplicate same type",
		"DECLARE name duplicate different type",
		"Variable, condition, and cursor in invalid order",
		"FETCH non-existent cursor",
		"OPEN non-existent cursor",
		"CLOSE non-existent cursor",
		"CLOSE without OPEN",
		"OPEN repeatedly",
		"CLOSE repeatedly",
		"With CTE using variable",
		"With CTE using parameter",
		"Dolt Issue #4480",
		"recursive procedure",
		"multi recursive procedures",
		"user variables are usable within stored procedures",
		"prepare statement inside of stored procedures",
		"stored procedure with exists subquery",
		"stored procedure with subquery set operations",
		"Resolve procedure variable in IS expression",
		"OUT param with SET",
		"OUT param without SET",
		"INOUT param with SET",
		"INOUT param without SET",
		"Nested CALL with INOUT param",
		"Incompatible type for parameter",
		"Incorrect parameter count",
		"use procedure parameter in filter expressions and multiple statements",
		"Call procedures by their qualified name",
		"String literals with escaped chars",
		"Procedures with TimestampFuncExpr",
		"Call a procedure that needs subqueries resolved in an if condition",
		"creating invalid procedure doesn't error until it is called",
		"event must not contain CREATE PROCEDURE",
		"table ddl statements in stored procedures",
		"procedure must not contain CREATE TRIGGER",
		"procedure must not contain CREATE DB",
		"procedure can CREATE VIEW",
		"nested procedure inserts",
		"DROP procedures",
		"SHOW procedures",
		"SHOW non-existent procedures",
	}
	h := newDoltgresServerHarness(t)
	defer h.Close()

	h.Setup(setup.MydbData)
	for _, suite := range []struct {
		name    string
		scripts []queries.ScriptTest
	}{
		{name: "logic tests", scripts: queries.ProcedureLogicTests},
		{name: "call tests", scripts: queries.ProcedureCallTests},
		{name: "create tests", scripts: queries.ProcedureCreateInSubroutineTests},
		{name: "drop tests", scripts: queries.ProcedureDropTests},
		{name: "show status tests", scripts: queries.ProcedureShowStatus},
		{name: "show create tests", scripts: queries.ProcedureShowCreate},
	} {
		t.Run(suite.name, func(t *testing.T) {
			for _, script := range suite.scripts {
				if slices.Contains(skippedScripts, script.Name) {
					t.Run(script.Name, func(t *testing.T) { t.Skip("Doltgres does not yet pass this script") })
					continue
				}

				enginetest.TestScript(t, h, script)
			}
		})
	}
}

func TestDoltStoredProcedures(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltStoredProceduresTest(t, h)
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
	denginetest.RunTransactionTests(t, h, false)
}

func TestTransactionsPrepared(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunTransactionTests(t, h, true)
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

func TestConcurrentTransactions(t *testing.T) {
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
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		// These scripts call dolt_commit_hash_out(@Commit1, ...). Passing a
		// user variable to dolt_commit_hash_out via the converter's
		// current_setting() shim hits an unrelated unsupported-operator
		// path. Unskip once that lands.
		"Show create table as of",
		"Show create table as of with FKs",
	})
	denginetest.RunShowCreateTableTests(t, h)
}

func TestViewsWithAsOf(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestScript(t, h, denginetest.ViewsWithAsOfScriptTest)
}

func TestDoltMerge(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"dolt_preview_merge_conflicts_summary(", // returns schema qualified table names
		"CALL DOLT_MERGE with schema conflicts can be correctly resolved using dolt_conflicts_resolve when autocommit is off", // alter table
		"CALL DOLT_MERGE fails on non-branch revision",                      // MySQL semantics: postgres aborts a transaction on any error, rejecting further statements
		"CALL DOLT_MERGE complains when a merge overrides local changes",    // MySQL semantics: postgres aborts a transaction on any error, rejecting further statements
		"merge conflicts prevent new branch creation",                       // different error message
		"Drop and add primary key on two branches converges to same schema", // alter table
		"insert two tables with the same name and different schema",
		"dropping constraint from one branch drops from both",                                                            // alter table (also catches the no-checkout variant)
		"merge with new triggers defined",                                                                                // triggers
		"try to merge a nullable field into a non-null column",                                                           // alter table
		"merge fulltext with renamed table",                                                                              // alter table
		"select * from dolt_status",                                                                                      // table_name column includes schema name,
		"dolt_merge() (3way) works with no auto increment overlap",                                                       // sequencing doesn't work globally after merge, need to decide product behavior
		"dolt_merge() (3way) with a gap in an auto increment key",                                                        // sequencing doesn't work globally after merge, need to decide product behavior
		"dolt_merge() with a gap in an auto increment key",                                                               // unsupported insert statements (need to call next_val, not insert NULL)
		"Merge does not panic when FK is dropped and re-added on one branch and child has composite PK with mixed types", // different foreign key syntax
		"three-way merge of table with vector index",                                                                     // MySQL VECTOR type has no Postgres equivalent
		"dolt_conflicts_resolve keeps vector index consistent with resolved rows",                                        // MySQL VECTOR type has no Postgres equivalent
	})
	denginetest.RunDoltMergeTests(t, h)
}

func TestDoltRebase(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltRebaseTests(t, h)
}

func TestDoltRevert(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"dolt_revert() respects dolt_ignore",                     // ERROR: INSERT: non-Doltgres type found in destination: text
		"dolt_revert() automatically resolves some conflicts",    // panic: interface conversion: sql.Type is types.VarCharType, not types.StringType
		"select count(*) from dolt_log;",                         // Doltgres creates an additional commit that Dolt doesn't have
		"dolt_revert() --continue: ignored table in working set", // ERROR: ASSIGNMENT_CAST: target is of type boolean but expression is of type integer: 1
	})
	denginetest.RunDoltRevertTests(t, h)
}

func TestDoltAutoIncrement(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltAutoIncrementTests(t, h)
}

func TestDoltConflictsTableNameTable(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"dolt_preview_merge_conflicts_summary(", // returns schema qualified table names
		"Provides a dolt_conflicts_id",          // relies on user vars
		"Updating our cols after schema change", // alter table
	})
	denginetest.RunDoltConflictsTableNameTableTests(t, h)
}

// tests new format behavior for keyless merges that create CVs and conflicts
func TestKeylessDoltMergeCVsAndConflicts(t *testing.T) {
	h := newDoltgresServerHarness(t)
	denginetest.RunKeylessDoltMergeCVsAndConflictsTests(t, h)
}

// eventually this will be part of TestDoltMerge
func TestDoltMergeArtifacts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"conflicts of different schemas can't coexist",                                  // alter table
		"violations with an older commit hash are overwritten if the value is the same", // nothing to commit?
		"regression test for bad column ordering in schema",                             // enum not supported in test harness
		"schema conflicts return an error when autocommit is enabled",                   // problems detecting autocommit for business logic
		"merge error lists all constraint violations when table has multiple violations",
	})
	denginetest.RunDoltMergeArtifacts(t, h)
}

func TestDoltReset(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"select * from dolt_status",                    // table_name column includes schema name
		"SELECT pk, v FROM t AS OF STAGED ORDER BY pk", // AS OF STAGED requires quoting in Postgres
		"SELECT pk FROM t AS OF STAGED ORDER BY pk",    // AS OF STAGED requires quoting in Postgres
	})
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
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"branch last checked out is deleted",
		"Using non-existent refs",
		"read-only databases", // read-only not yet implemented in harness
		"Checkout tables from commit",
		"dolt_checkout with tracking branch and table with same name", // UseLocalFileSystem did not create remote dir
	})
	denginetest.RunDoltCheckoutTests(t, h)
}

func TestDoltBranch(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"Create branch from startpoint",  // missing SET @var syntax
		"Join same table at two commits", // implicit cross-joins (no ON clause) unsupported in postgres
	})

	denginetest.RunDoltBranchTests(t, h)
}

func TestDoltTag(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		// dolt's initialization is different which results in a different user name for the tagger,
		// should fix the harness to match
		"SELECT tag_name, IF(CHAR_LENGTH(tag_hash) < 0, NULL, 'not null'), tagger, email, IF(date IS NULL, NULL, 'not null'), message from dolt_tags",
	})
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
	harness := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"explain",                       // not supported
		"select message from dolt_log",  // doltgres setup has extra commits
		"dolt_history table with enums", // enums
		"can sort by dolt_log.commit",   // more commits
	}).WithParallelism(2)
	denginetest.RunHistorySystemTableTests(t, harness)
}

func TestUnscopedDiffSystemTable(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunUnscopedDiffSystemTableTests(t, h)
}

func TestColumnDiffSystemTable(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunColumnDiffSystemTableTests(t, h)
}

func TestStatBranchTests(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunStatBranchTests(t, harness)
}

func TestDiffTableFunction(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDiffTableFunctionTests(t, harness)
}

func TestDiffStatTableFunction(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDiffStatTableFunctionTests(t, harness)
}

func TestDiffSummaryTableFunction(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDiffSummaryTableFunctionTests(t, harness)
}

func TestPatchTableFunction(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltPatchTableFunctionTests(t, harness)
}

func TestLogTableFunction(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunLogTableFunctionTests(t, harness)
}

func TestDoltReflog(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltReflogTests(t, harness)
}

func TestCommitDiffSystemTable(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunCommitDiffSystemTableTests(t, harness)
}

func TestDiffSystemTable(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltDiffSystemTableTests(t, h)
}

func TestSchemaDiffTableFunction(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunSchemaDiffTableFunctionTests(t, harness)
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

func TestSystemTableFunctionIndexes(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunSystemTableFunctionIndexesTests(t, harness)
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
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestNullRanges(t, h)
}

func TestPersist(t *testing.T) {
	t.Skip()
	ctx := context.Background()
	harness := newDoltgresServerHarness(t)
	defer harness.Close()
	dEnv := dtestutils.CreateTestEnv()
	defer dEnv.DoltDB(ctx).Close()
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

func TestDoltCommit(t *testing.T) {
	harness := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		// These tests set @@autocommit, which we can't translate accurately yet
		"CALL DOLT_COMMIT('-amend') works to update commit message",
		"CALL DOLT_COMMIT('-amend') works to add changes to a commit",
		"CALL DOLT_COMMIT('-amend') works to remove changes from a commit",
		"CALL DOLT_COMMIT('-amend') works to update a merge commit",
		"CALL DOLT_COMMIT('--amend') works on initial commit",
		"DOLT_COMMIT respects foreign_key_checks=0",
	})
	denginetest.RunDoltCommitTests(t, harness)
}

func TestStatsHistograms(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunStatsHistogramTests(t, h)
}

// TestStatsStorage force a provider reload in-between setup and assertions that
// forces a round trip of the statistics table before inspecting values.
func TestStatsStorage(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunStatsStorageTests(t, h)
}

func TestJoinStats(t *testing.T) {
	h := newDoltgresServerHarness(t)
	denginetest.RunJoinStatsTests(t, h)
}

func TestStatisticIndexes(t *testing.T) {
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestStatisticIndexFilters(t, h)
}

func TestCharsetCollationEngine(t *testing.T) {
	// MySQL character set syntax and collation behavior differ.
	skippedScripts := []string{
		"CAST(... AS BINARY) function",
		"Issue #5482",
		"LIKE with a space terminated prefix matches rows with a multibyte character after the prefix",
		"LIKE with a constant prefix keeps rows that sort after the prefix across collations, NOT LIKE, and joins",
		"IN predicate with accent-insensitive collation",
		"CHECK constraint with IN predicate and collation",
		"BINARY() function", // malformed UTF-16 input causes a stack overflow
		"LIKE respects table collations",
		"LIKE respects connection collation",
		"STRCMP() function",
		"LENGTH() function",
		"CHAR_LENGTH() function",
		"CONVERT() USING with malformed multi-byte strings",
		"UPPER() function",
		"LOWER() function",
		"RPAD() function",
		"LPAD() function",
		"SUBSTRING() function",
		"FROM_BASE64() function",
		"TRIM() function",
		"RTRIM() function",
		"LTRIM() function",
		"SET collation handling",
		"invalid utf8 encoding strings", // need to investigate why some strings aren't giving errors, might be a harness error
		"Insert multiple character sets",
		"Sorting differences",
		"Character set introducer with invalid collate",
		"Properly block using not-yet-implemented character sets/collations",
		"Order by behaves differently according to case-sensitivity",
		"Proper index access",
		"Table collation is respected",
		"SET NAMES does not interfere with column charset",
		"SET validates character set and collation variables",
		"setting charset/collation sets the other",
		"ENUM collation handling",
	}
	h := newDoltgresServerHarness(t)
	defer h.Close()

	h.Setup(setup.MydbData)
	for _, script := range queries.CharsetCollationEngineTests {
		if slices.Contains(skippedScripts, script.Name) {
			t.Run(script.Name, func(t *testing.T) { t.Skip("Doltgres does not yet pass this script") })
			continue
		}

		enginetest.TestScript(t, h, script)
	}
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
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltStorageFormatTests(t, h)
}

func TestThreeWayMergeWithSchemaChangeScripts(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)

	denginetest.RunThreeWayMergeWithSchemaChangeScripts(t, h)
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

func TestAdaptiveEncoding(t *testing.T) {
	// The test runs end-to-end, but most subtests currently fail because:
	//  - bytea/text round-trips through the harness lose data for BLOB columns
	//  - the converter does not yet translate MySQL prefix indexes (`col(N)`),
	//    LOAD_FILE(), or INSERT of text literals into bytea columns
	// Unskip once those gaps are filled.
	t.Skip("AdaptiveEncoding runs but fails on bytea round-trip / MySQL-only queries; see comment")

	adaptiveEncoding := typeinfo.UseAdaptiveEncoding
	defer func() { typeinfo.UseAdaptiveEncoding = adaptiveEncoding }()
	typeinfo.UseAdaptiveEncoding = true

	denginetest.RunTestAdaptiveEncoding(t, newDoltgresServerHarness(t), denginetest.AdaptiveEncodingTestType_Blob, denginetest.AdaptiveEncodingTestPurpose_Representation)
	denginetest.RunTestAdaptiveEncoding(t, newDoltgresServerHarness(t), denginetest.AdaptiveEncodingTestType_Blob, denginetest.AdaptiveEncodingTestPurpose_Correctness)
	denginetest.RunTestAdaptiveEncoding(t, newDoltgresServerHarness(t), denginetest.AdaptiveEncodingTestType_Text, denginetest.AdaptiveEncodingTestPurpose_Representation)
	denginetest.RunTestAdaptiveEncoding(t, newDoltgresServerHarness(t), denginetest.AdaptiveEncodingTestType_Text, denginetest.AdaptiveEncodingTestPurpose_Correctness)

	denginetest.RunAdaptiveEncodingScripts(t, newDoltgresServerHarness(t))
}

func TestSchemaOverridesWithAdaptiveEncoding(t *testing.T) {
	// Runs end-to-end but fails on @@dolt_override_schema (user-variable scope)
	// and the MySQL DESCRIBE statement. Unskip once those land.
	t.Skip("SchemaOverrides runs but fails on @@dolt_override_schema / DESCRIBE; see comment")

	adaptiveEncoding := typeinfo.UseAdaptiveEncoding
	defer func() { typeinfo.UseAdaptiveEncoding = adaptiveEncoding }()
	typeinfo.UseAdaptiveEncoding = true
	harness := newDoltgresServerHarness(t)
	denginetest.RunSchemaOverridesTest(t, harness)
}

func TestJsonAdaptiveEncoding(t *testing.T) {
	adaptiveEncoding := typeinfo.UseAdaptiveEncoding
	defer func() { typeinfo.UseAdaptiveEncoding = adaptiveEncoding }()
	typeinfo.UseAdaptiveEncoding = true

	harness := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"json_extract(", // MySQL-specific; postgres uses -> / jsonb_extract_path_text
	})
	denginetest.RunJsonAdaptiveEncodingTests(t, harness)
}

func TestBackupsSystemTable(t *testing.T) {
	t.Skip("port test from Dolt")
}

func TestBranchActivity(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	denginetest.RunBranchActivityTests(t, harness)
}

func TestBranchStatusTableFunction(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	denginetest.RunBranchStatusTableFunctionTests(t, harness)
}

func TestConcurrentCreateDatabaseIfNotExists(t *testing.T) {
	t.Skip("port test from Dolt")
}

func TestDoltBranchesSystemTable(t *testing.T) {
	t.Skip("port test from Dolt")
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltBranchesSystemTableTests(t, h)
}

func TestDoltCommitVerificationScripts(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltCommitVerificationScripts(t, harness)
}

func TestDoltDTableScripts(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltDTableScriptsTest(t, harness)
}

func TestDoltForeignKeyTests(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltForeignKeyTests(t, harness)
}

func TestDoltHelpSystemTable(t *testing.T) {
	harness := newDoltgresServerHarness(t)
	defer harness.Close()
	denginetest.RunDoltHelpSystemTableTests(t, harness)
}

func TestDoltPreviewMergeConflicts(t *testing.T) {
	t.Skip("port test from Dolt")
	h := newDoltgresServerHarness(t)
	denginetest.RunDoltPreviewMergeConflictsTests(t, h)
}

func TestDoltQueryCatalogSystemTable(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	defer harness.Close()
	denginetest.RunDoltQueryCatalogTests(t, harness)
}

func TestDoltRm(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	defer harness.Close()
	denginetest.RunDoltRmTests(t, harness)
}

func TestDoltStash(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	defer harness.Close()
	denginetest.RunDoltStashSystemTableTests(t, harness)
}

func TestDoltTestsSystemTable(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	defer harness.Close()
	denginetest.RunDoltTestsTableTests(t, harness)
}

func TestDoltUserPrivileges(t *testing.T) {
	t.Skip("MySQL-specific user privilege tests, not applicable to doltgresql")
}

func TestDoltWorkspace(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltWorkspaceTests(t, harness)
}

func TestDriverExecution(t *testing.T) {
	t.Skip("port test from Dolt - uses dolt_backup which is not yet supported")
}

func TestJsonDiffTableFunction(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	denginetest.RunJsonDiffTableFunctionTests(t, harness)
}

func TestJsonValueScripts(t *testing.T) {
	harness := newDoltgresServerHarness(t)
	denginetest.RunJsonValueScriptsTest(t, harness)
}

func TestLargeGeometryScripts(t *testing.T) {
	t.Skip("doltgresql does not support spatial types")
}

func TestLegacyCreateTableScripts(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	denginetest.RunLegacyCreateTableScripts(t, harness)
}

func TestLegacyDeleteScripts(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	denginetest.RunLegacyDeleteScripts(t, harness)
}

func TestLegacyDropTableScripts(t *testing.T) {
	harness := newDoltgresServerHarness(t)
	denginetest.RunLegacyDropTableScripts(t, harness)
}

func TestLegacyIndexScripts(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	denginetest.RunLegacyIndexScripts(t, harness)
}

func TestLegacyInsertScripts(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	denginetest.RunLegacyInsertScripts(t, harness)
}

func TestLegacyJoinScripts(t *testing.T) {
	harness := newDoltgresServerHarness(t)
	denginetest.RunLegacyJoinScripts(t, harness)
}

func TestLegacyReplaceScripts(t *testing.T) {
	// TODO: two remaining root causes:
	//  - REPLACE INTO with a multi-row VALUES list (with or without a column list) bails
	//    to raw/untranslated SQL text, which the postgres parser then rejects outright
	//    ("at or near "replace": syntax error") since REPLACE isn't valid postgres syntax.
	//  - is_married is BIGINT (MySQL boolean-as-int convention); postgres's strict typing
	//    rejects assigning a boolean literal to it (ASSIGNMENT_CAST error).
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunLegacyReplaceScripts(t, harness)
}

func TestLegacySelectScripts(t *testing.T) {
	t.Skip("port test from Dolt")
	harness := newDoltgresServerHarness(t)
	denginetest.RunLegacySelectScripts(t, harness)
}

func TestLegacyUpdateScripts(t *testing.T) {
	harness := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		// postgres strict typing rejects MySQL coercions: is_married is BIGINT (MySQL boolean-as-int
		// convention), and postgres's ASSIGNMENT_CAST rejects assigning a boolean literal to it.
		"update one row, all cols, non-primary key where clause",
		"update one row, set columns to existing values",
	})
	denginetest.RunLegacyUpdateScripts(t, harness)
}

func TestNonlocalTable(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		// MySQL auto-names unnamed FKs "<table>_ibfk_<n>"; postgres uses "<table>_<col>_fkey". The generated name
		// therefore differs from the shared (MySQL-authored) expected value in these ported assertions.
		"show create table local_table",
		`insert into local_table values ("fdnfjfjf")`,
		// TODO: DOLT_VERIFY_CONSTRAINTS('--all') re-verifies constraints via a root merge, and the merge machinery
		//  doesn't currently carry the "dolt" namespace schema (where dolt_nonlocal_tables lives under Doltgres's
		//  schema/search-path model) through to the merged root, so FKs referencing a nonlocal table can't be
		//  resolved there. Regular (non-nonlocal) FKs are unaffected.
		"call dolt_verify_constraints",
		"select violation_type from dolt_constraint_violations_local_table",
	})
	denginetest.RunNonlocalTableTests(t, h)
}

func TestNumericErrorScripts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"insert into ui16 values (65535)", // postgres has no unsigned types; ui16 maps to smallint (max 32767), so 65535 legitimately overflows
	})
	defer h.Close()
	enginetest.TestNumericErrorScripts(t, h)
}

func TestSingleMergeScript(t *testing.T) {
	t.Skip("debug harness for a single merge script")
}

func TestSingleTransactionScript(t *testing.T) {
	t.Skip("debug harness for a single transaction script")
}

func TestTimeQueries(t *testing.T) {
	t.Skip("port test from Dolt")
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestTimeQueries(t, h)
}

func TestUpdateIgnore(t *testing.T) {
	t.Skip("MySQL UPDATE IGNORE semantics are not applicable to PostgreSQL")
}

func TestUserPrivileges(t *testing.T) {
	t.Skip("MySQL-specific user privilege tests, not applicable to doltgresql")
}

func TestVectorFunctions(t *testing.T) {
	t.Skip("MySQL-dialect vector tests")
}

func TestVectorIndexes(t *testing.T) {
	t.Skip("MySQL-dialect vector tests")
}

func TestVectorType(t *testing.T) {
	t.Skip("MySQL-dialect vector tests")
}
