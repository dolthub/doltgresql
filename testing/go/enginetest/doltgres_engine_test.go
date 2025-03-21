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
	"regexp"
	"runtime"
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

	enginetest.TestQueryWithEngine(t, harness, engine, test)
}

func TestSchemaOverrides(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunSchemaOverridesTest(t, harness)
}

type doltCommitValidator struct{}

var _ enginetest.CustomValueValidator = &doltCommitValidator{}

// TODO: this custom validator is supposed to match only a commit hash, but we extend it to match the formatting
//
//	characters present in the Doltgres response for some calls. We can remove this when we support the syntax
//	`select * from dolt_commit(...)`
var hashRegex = regexp.MustCompile(`^\{?([0-9a-v]{32}).*$`)

// Validate returns true if the value is a valid commit hash.
func (dcv *doltCommitValidator) Validate(val interface{}) (bool, error) {
	hash, ok := val.(string)
	if !ok {
		return false, nil
	}
	return hashRegex.MatchString(hash), nil
}

// CommitHash returns the commit hash from the value, if it is a valid commit hash.
func (dcv *doltCommitValidator) CommitHash(val interface{}) (bool, string) {
	hash, ok := val.(string)
	if !ok {
		return false, ""
	}

	matches := hashRegex.FindStringSubmatch(hash)
	if len(matches) == 0 {
		return false, ""
	}
	return true, matches[1]
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
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"SELECT s as coL1, SUM(i) coL2 FROM mytable group by 1 order by 2",      // incorrect result
		"SELECT s as Date, SUM(i) TimeStamp FROM mytable group by 1 order by 2", // ERROR: at or near "timestamp": syntax error
		"select \"foo\" as dummy, (select dummy)",                               // Unhandled OID 705
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
		`insert into keyless (c0, c1) select a.c0, a.c1 from (select 1, 1) as a(c0, c1) join keyless on a.c0 = keyless.c0`,                        // missing result element, needs investigation
		`INSERT INTO mytable (i,s) SELECT i * 2, concat(s,s) from mytable order by 1 desc limit 1`,                                                // invalid type: bigint
		"with t (i,f) as (select 4,'fourth row' from dual) insert into mytable select i,f from t",                                                 // WITH unsupported syntax
		"with recursive t (i,f) as (select 4,4 from dual union all select i + 1, i + 1 from t where i < 5) insert into mytable select i,f from t", // WITH unsupported syntax
		"issue 6675: on duplicate rearranged getfield indexes from select source",                                                                 // panic
		"issue 4857: insert cte column alias with table alias qualify panic",                                                                      // WITH unsupported syntax
		"Insert on duplicate key references table in subquery",                                                                                    // bad translation?
		"Insert on duplicate key references table in aliased subquery",                                                                            // bad translation?
		"Insert on duplicate key references table in cte",                                                                                         // CTE not supported
		"insert on duplicate key with incorrect row alias",                                                                                        // column "c" could not be found in any table in scope
		"insert on duplicate key update errors",                                                                                                   // failing
		"Insert on duplicate key references table in subquery with join",                                                                          // untranslated
		"INSERT INTO ... SELECT with TEXT types",                                                                                                  // typecasts needed
	})
	defer h.Close()
	enginetest.TestInsertInto(t, h)
}

func TestInsertIgnoreInto(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestInsertIgnoreInto(t, h)
}

func TestInsertDuplicateKeyKeyless(t *testing.T) {
	t.Skip()
	enginetest.TestInsertDuplicateKeyKeyless(t, newDoltgresServerHarness(t))
}

func TestInsertDuplicateKeyKeylessPrepared(t *testing.T) {
	t.Skip()
	enginetest.TestInsertDuplicateKeyKeylessPrepared(t, newDoltgresServerHarness(t))
}

func TestIgnoreIntoWithDuplicateUniqueKeyKeyless(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestIgnoreIntoWithDuplicateUniqueKeyKeyless(t, h)
}

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
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"UPDATE mytable SET s = _binary 'updated' WHERE i = 3;",         // _binary not supported
		"UPDATE mytable SET s = 'updated' ORDER BY i LIMIT 1 OFFSET 1;", // offset not supported (limit isn't selected in vanilla postgres but is in the cockroach grammar)
		// TODO: Postgres supports update joins but with a very different syntax, and some join types are not supported
		`UPDATE one_pk INNER JOIN two_pk on one_pk.pk = two_pk.pk1 SET two_pk.c1 = two_pk.c1 + 1`,
		"UPDATE mytable INNER JOIN one_pk ON mytable.i = one_pk.c5 SET mytable.i = mytable.i * 10",
		`UPDATE one_pk INNER JOIN two_pk on one_pk.pk = two_pk.pk1 SET two_pk.c1 = two_pk.c1 + 1 WHERE one_pk.c5 < 10`,
		`UPDATE one_pk INNER JOIN two_pk on one_pk.pk = two_pk.pk1 INNER JOIN othertable on othertable.i2 = two_pk.pk2 SET one_pk.c1 = one_pk.c1 + 1`,
		`UPDATE one_pk INNER JOIN (SELECT * FROM two_pk order by pk1, pk2) as t2 on one_pk.pk = t2.pk1 SET one_pk.c1 = t2.c1 + 1 where one_pk.pk < 1`,
		`UPDATE one_pk INNER JOIN two_pk on one_pk.pk = two_pk.pk1 SET one_pk.c1 = one_pk.c1 + 1`,
		`UPDATE one_pk INNER JOIN two_pk on one_pk.pk = two_pk.pk1 SET one_pk.c1 = one_pk.c1 + 1, one_pk.c2 = one_pk.c2 + 1 ORDER BY one_pk.pk`,
		`UPDATE one_pk INNER JOIN two_pk on one_pk.pk = two_pk.pk1 SET one_pk.c1 = one_pk.c1 + 1, two_pk.c1 = two_pk.c2 + 1`,
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

func TestUpdateErrors(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		`UPDATE keyless INNER JOIN one_pk on keyless.c0 = one_pk.pk SET keyless.c0 = keyless.c0 + 1`,
		"try updating string that is too long",  // works but error message doesn't match
		"UPDATE mytable SET s = 'hi' LIMIT -1;", // unsupported syntax
	})
	defer h.Close()
	enginetest.TestUpdateErrors(t, h)
}

func TestDeleteFrom(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"DELETE FROM mytable ORDER BY i DESC LIMIT 1 OFFSET 1;", // offset is unsupported syntax
		"DELETE FROM mytable WHERE (i,s) = (1, 'first row');",   // type error, needs investigation
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
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"filter pushdown through join uppercase name",             // syntax error (join without on)
		"issue 7958, update join uppercase table name validation", // update join syntax not supported
		"Dolt issue 7957, update join matched rows",               // update join syntax not supported
		"update join with update trigger different value",         // update join syntax not supported
		"update join with update trigger same value",              // update join syntax not supported
		"update join with update trigger",                         // update join syntax not supported
		"update join with update trigger if condition",            // update join syntax not supported
		"missing indexes",                       // unsupported test harness setup
		"table x intersect table y order by i;", // type coercion rules for unions among different schemas need to change for this to work, mysql is much more lenient
		"table x intersect table y order by 1;", // type coercion rules for unions among different schemas need to change for this to work, mysql is much more lenient
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
		"with recursive cte as ((select * from xy order by y asc limit 1 offset 1) union (select * from xy order by y asc limit 1 offset 2)) select * from cte", // invalid type: bigint
		"CrossDB Queries", // needs harness work to properly qualify the names
		"SELECT rand(10) FROM tab1 GROUP BY tab1.col1", // different rand() behavior
		"Nested Subquery projections (NTC)",            // ERROR: blob/text column 'id' used in key specification without a key length
		"CREATE TABLE SELECT Queries",                  // ERROR: TableCopier only accepts CreateTable or TableNode as the destination
		// "Simple Update Join test that manipulates two tables",
		// "Partial indexes are used and return the expected result",
		// "Multiple indexes on the same columns in a different order",
		"Ensure proper DECIMAL support (found by fuzzer)", // unsupported type: SET
		// "Ensure scale is not rounded when inserting to DECIMAL type through float64",
		"Show create table with various keys and constraints",                                                     // error in harness query converter
		"show create table with duplicate primary key",                                                            // error in harness query converter
		"recreate primary key rebuilds secondary indexes",                                                         // currently no way to drop primary key in doltgres
		"Handle hex number to binary conversion",                                                                  // ERROR: can't convert 0x7ED0599B to decimal: exponent is not numeric
		"join index lookups do not handle filters",                                                                // need a different join syntax (no ON clause not supported in postgres)
		"select count(*) from numbers group by val having count(*) < val;",                                        // ERROR: unable to find field with index 1 in row of 1 columns
		"using having and group by clauses in subquery ",                                                          // lots of index errors, something very broken
		"can't create view with same name as existing table",                                                      // error message wrong
		"arithmetic bit operations on int, float and decimal types",                                               // the power operator is not yet supported
		"INSERT IGNORE throws an error when json is badly formatted",                                              // error messages don't match
		"identical expressions over different windows should produce different results",                           // ERROR: integer: unhandled type: float64
		"windows without ORDER BY should be treated as RANGE BETWEEN UNBOUNDED PRECEDING AND UNBOUNDED FOLLOWING", // ERROR: integer: unhandled type: float64
		"decimal literals should be parsed correctly",                                                             // ERROR: text: unhandled type: decimal.Decimal (error in harness)
		"division and int division operation on negative, small and big value for decimal type column of table",   // numeric keys broken
		"different cases of function name should result in the same outcome",                                      // ERROR: blob/text column 'b' used in key specification without a key length
		"Multi-db Aliasing",                                                                  // need harness support for qualified table names
		"Complex Filter Index Scan #2",                                                       // panic in index lookup, needs investigation
		"Complex Filter Index Scan #3",                                                       // panic in index lookup, needs investigation
		"update columns with default",                                                        // broken, see repro in update_test.go
		"select * from t0 where i > 0.1 or i >= 0.1 order by i;",                             // incorrect result, needs a fix
		"int secondary index with float filter",                                              // panic
		"select count(*) from t where (f in (null, cast(0.8 as float)));",                    // incorrect result, needs a fix
		"update with left join with some missing rows",                                       // need to translate update joins
		"SELECT - col2 AS col0 FROM tab2 GROUP BY col0, col2 HAVING NOT + + col2 <= - col0;", // incorrect result
		"SELECT -col2 AS col0 FROM tab2 GROUP BY col0, col2 HAVING NOT col2 <= - col0;",      // incorrect result
		"SELECT -col2 AS col0 FROM tab2 GROUP BY col0, col2 HAVING col2 > -col0;",            // incorrect result
		"select col2-100 as col0 from tab2 group by col0 having col0 > 0;",                   // incorrect result
		"complicated range tree",                                                             // panic in index lookup, needs investigation
		"preserve now()",                                                                     // harness error
		"binary type primary key",                                                            // ERROR: blob/text column 'b' used in key specification without a key length
		"varbinary primary key",                                                              // ERROR: blob/text column 'b' used in key specification without a key length
		"insert into t1 (a, b) values ('1234567890', '12345')",                               // different error message
		"insert into t2 (a, b) values ('1234567890', '12345')",                               // different error message
		"invalid utf8 encoding strings",                                                      // need to investigate why some strings aren't giving errors, might be a harness error
		"mismatched collation using hash in tuples",                                          // ERROR: plan is not resolved because of node '*plan.Project'
		"validate_password_strength and validate_password.length",                            // unsupported
		"validate_password_strength and validate_password.number_count",                      // unsupported
		"validate_password_strength and validate_password.mixed_case_count",                  // unsupported
		"validate_password_strength and validate_password.special_char_count",                // unsupported
		"preserve enums through alter statements",                                            // enum types unsupported
		"coalesce with system types",                                                         // unsupported
		"multi enum return types",                                                            // enum types unsupported
		"enum cast to int and string",                                                        // enum types unsupported
		"select * from vt where v = cast('def' as char(6));",                                 // incorrect result
		"select * from vt where v < cast('def' as char(6));",                                 // incorrect result
		"select * from vt where v >= cast('def' as char(6));",                                // incorrect result
	})
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
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"keyless table merge with constraint violation on duplicate rows",                                                     // alter table
		"CALL DOLT_MERGE without conflicts correctly works with autocommit off with commit flag",                              // datetime support
		"CALL DOLT_MERGE without conflicts correctly works with autocommit off and no commit flag",                            // datetime support
		"CALL DOLT_MERGE with conflicts can be correctly resolved when autocommit is off",                                     // datetime support
		"CALL DOLT_MERGE with schema conflicts can be correctly resolved using dolt_conflicts_resolve when autocommit is off", // alter table
		"merge conflicts prevent new branch creation",                                                                         // different error message
		"select message from dolt_log where date ",                                                                            // datetime support
		"DOLT_MERGE(--abort) clears staged",
		"CALL DOLT_MERGE complains when a merge overrides local changes",
		"Drop and add primary key on two branches converges to same schema",  // alter table
		"Constraint violations are persisted",                                // foreign key support
		"left adds a unique key constraint and resolves existing violations", // alter table
		"insert two tables with the same name and different schema",
		"merge with new triggers defined",                                                                 // triggers
		"add multiple columns, then set and unset a value. No conflicts expected.",                        // alter table
		"dropping constraint from one branch drops from both",                                             // alter table
		"dropping constraint from one branch drops from both, no checkout",                                // alter table
		"merge constraint with valid data on different branches",                                          // alter table
		"resolving a deleted and modified row handles constraint checks",                                  // alter table
		"resolving a modified/modified row still checks nullness constraint",                              // alter table
		"Merge errors if the primary key types have changed (even if the new type has the same NomsKind)", // alter table
		"parent index is longer than child index",
		"parallel column updates (repro issue #4547)",
		"try to merge a nullable field into a non-null column",        // alter table
		"merge fulltext with renamed table",                           // alter table
		"merge when schemas are equal, but column tags are different", // alter table
		"merge with float column default",                             // alter table
		"merge with float 1.23 column default",                        // alter table
		"merge with decimal 1.23 column default",                      // alter table
		"merge with different types",                                  // alter table
		"select * from dolt_status",                                   // table_name column includes schema name,
		"dolt_merge() (3way) works with no auto increment overlap",    // sequencing doesn't work globally after merge, need to decide product behavior
		"dolt_merge() (3way) with a gap in an auto increment key",     // sequencing doesn't work globally after merge, need to decide product behavior
		"dolt_merge() with a gap in an auto increment key",            // unsupported insert statements (need to call next_val, not insert NULL)
	})
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
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"dolt_revert() respects dolt_ignore",                  // ERROR: INSERT: non-Doltgres type found in destination: text
		"dolt_revert() automatically resolves some conflicts", // panic: interface conversion: sql.Type is types.VarCharType, not types.StringType
	})
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
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"Provides a dolt_conflicts_id",                                                     // relies on user vars
		"dolt_conflicts_id is unique across merges",                                        // relies on user vars
		"Updates on our columns get applied to the source table - compound / inverted pks", // broken, not clear why
		"Updates on our columns get applied to the source table - keyless",                 // type issue
		"Updating our cols after schema change",                                            // alter table
	})
	denginetest.RunDoltConflictsTableNameTableTests(t, h)
}

// tests new format behavior for keyless merges that create CVs and conflicts
func TestKeylessDoltMergeCVsAndConflicts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"Keyless merge with foreign keys documents violations", // foreign keys
		"unique key violation for keyless table",               // alter table
	})
	denginetest.RunKeylessDoltMergeCVsAndConflictsTests(t, h)
}

// eventually this will be part of TestDoltMerge
func TestDoltMergeArtifacts(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"conflicts of different schemas can't coexist",                                            // alter table
		"violations with an older commit hash are overwritten if the value is the same",           // nothing to commit?
		"right adds a unique key constraint and resolves existing violations.",                    // alter table
		"unique key violation should be thrown even if a PK column is used in the unique index",   // alter table ADD UNIQUE syntax
		"unique key violation should be thrown even if a PK column is used in the unique index 2", // alter table ADD UNIQUE syntax
		"unique key violations should not be thrown for keys with null values",                    // alter table ADD UNIQUE syntax
		"regression test for bad column ordering in schema",                                       // enum not supported in test harness
		"schema conflicts return an error when autocommit is enabled",                             // problems detecting autocommit for business logic
		"Multiple foreign key violations for a given row not supported",                           // foreign keys
		"divergent type change causes schema conflict",                                            // alter table
	})
	denginetest.RunDoltMergeArtifacts(t, h)
}

func TestDoltReset(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"CALL DOLT_RESET('--hard') should reset the merge state after uncommitted merge", // problem with autocommit detection
		"select * from dolt_status", // table_name column includes schema name
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
		"dolt_checkout and base name resolution", // needs db-qualified table names
		"branch last checked out is deleted",
		"Using non-existent refs",
		"read-only databases", // read-only not yet implemented in harness
		"Checkout tables from commit",
		"dolt_checkout with new branch forcefully",                        // string primary key ordering broken
		"dolt_checkout with new branch forcefully with dirty working set", // string primary key ordering broken
	})
	denginetest.RunDoltCheckoutTests(t, h)
}

func TestDoltCheckoutPrepared(t *testing.T) {
	t.Skip("need to implement prepared queries in harness")
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"dolt_checkout and base name resolution", // needs db-qualified table names
		"branch last checked out is deleted",
		"Using non-existent refs",
		"read-only databases", // read-only not yet implemented in harness
	})
	denginetest.RunDoltCheckoutPreparedTests(t, h)
}

func TestDoltBranch(t *testing.T) {
	h := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		"Create branch from startpoint",  // missing SET @var syntax
		"Join same table at two commits", // needs different branch-qualified DB syntax
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
		"explain",                                       // not supported
		"select message from dolt_log;",                 // more commits
		"primary key table: rename table",               // DDL
		"primary key table: non-pk column type changes", // DDL
		"dolt_history table with AS OF",                 // AS OF
		"dolt_history table with AS OF",                 // AS OF
		"dolt_history table with enums",                 // enums
		"can sort by dolt_log.commit",                   // more commits
	}).WithParallelism(2)
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

func TestDoltCherryPickPrepared(t *testing.T) {
	t.Skip()
	harness := newDoltgresServerHarness(t)
	denginetest.RunDoltCherryPickTestsPrepared(t, harness)
}

func TestDoltCommit(t *testing.T) {
	harness := newDoltgresServerHarness(t).WithSkippedQueries([]string{
		// These tests set @@autocommit, which we can't translate accurately yet
		"CALL DOLT_COMMIT('-amend') works to update commit message",
		"CALL DOLT_COMMIT('-amend') works to add changes to a commit",
		"CALL DOLT_COMMIT('-amend') works to remove changes from a commit",
		"CALL DOLT_COMMIT('-amend') works to update a merge commit",
	})
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

// TestStatsStorage force a provider reload in-between setup and assertions that
// forces a round trip of the statistics table before inspecting values.
func TestStatsStorage(t *testing.T) {
	t.Skip()
	h := newDoltgresServerHarness(t)
	denginetest.RunStatsStorageTests(t, h)
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
	skipPreparedTests(t)
	h := newDoltgresServerHarness(t)
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

func skipPreparedTests(t *testing.T) {
	if skipPrepared {
		t.Skip("skip prepared")
	}
}
