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
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTargetExpectationSelection(t *testing.T) {
	for _, target := range []serverTestTarget{targetDoltgres, targetPostgres, targetPostgresFork} {
		t.Run(string(target), func(t *testing.T) {
			cases := []ScriptTestAssertion{{Query: "INSERT shared", Postgres: &ScriptTestAssertion{Query: "INSERT native"}, PostgresFork: &ScriptTestAssertion{Query: "INSERT fork"}}}
			variants := expandAssertionsForTarget(cases, target)
			require.Len(t, variants, 3)
			active := 0
			for _, v := range variants {
				if v.variantEnabled {
					active++
					require.Equal(t, string(target), v.variant)
				}
			}
			require.Equal(t, 1, active)
			require.Empty(t, cases[0].variant, "selection must not mutate reusable fixtures")
			inherited := expandAssertionsForTarget([]ScriptTestAssertion{{Query: "SELECT 1", Postgres: &ScriptTestAssertion{Expected: []sql.Row{{1}}}}}, target)
			require.Equal(t, "SELECT 1", inherited[1].Query)
			require.Equal(t, target != targetDoltgres, inherited[1].variantEnabled)
			forkOnly := expandAssertionsForTarget([]ScriptTestAssertion{{Query: "SELECT dolt_commit()", PostgresFork: &ScriptTestAssertion{}}}, target)
			if target == targetPostgres {
				require.False(t, forkOnly[0].variantEnabled)
				require.False(t, forkOnly[1].variantEnabled)
			}
		})
	}
	target, err := parseTestTarget("")
	require.NoError(t, err)
	require.Equal(t, targetDoltgres, target)
	_, err = parseTestTarget("postgre")
	require.Error(t, err)
	skipped := expandAssertionsForTarget([]ScriptTestAssertion{{Skip: true, Postgres: &ScriptTestAssertion{}}}, targetPostgres)
	require.True(t, skipped[1].Skip)
	require.NoError(t, checkPlanRelations(`[{"Plan":{"Node Type":"Sort","Plans":[{"Node Type":"Index Scan","Relation Name":"pg_class"}]}}]`, []string{"pg_class"}))
	require.Error(t, checkPlanRelations(`[{"Plan":{"Node Type":"Result"}}]`, []string{"pg_class"}))
	require.Error(t, checkPlanRelations(`[]`, nil))
	require.NoError(t, checkCatalogScanAliases(`[{"Plan":{"Node Type":"Custom Scan","Custom Plan Provider":"DoltCatalogScan","Alias":"c"}}]`, []string{"c"}))
	require.Error(t, checkCatalogScanAliases(`[{"Plan":{"Node Type":"Seq Scan","Alias":"c"}}]`, []string{"c"}))
	require.Error(t, checkCatalogScanAliases(`[{"Plan":{"Node Type":"Custom Scan","Custom Plan Provider":"DoltCatalogScan","Alias":"c"}}]`, []string{"other"}))

}

func TestTargetExpectationExecutesOnce(t *testing.T) {
	RunScripts(t, []ScriptTest{{Name: "only the selected mutation executes", SetUpScript: []string{"CREATE TABLE target_once (id int PRIMARY KEY)"}, Assertions: []ScriptTestAssertion{
		{Query: "INSERT INTO target_once VALUES (1)", Postgres: &ScriptTestAssertion{Query: "INSERT INTO target_once VALUES (1)"}, PostgresFork: &ScriptTestAssertion{Query: "INSERT INTO target_once VALUES (1)"}},
		{Query: "SELECT id FROM target_once", Expected: []sql.Row{{1}}},
	}}})
	RunTransactionTests(t, []ScriptTest{{Name: "transaction counterparts execute once", SetUpScript: []string{"CREATE TABLE target_once (id int PRIMARY KEY)"}, Assertions: []ScriptTestAssertion{
		{Query: "/* client A */ BEGIN"},
		{Query: "/* client A */ INSERT INTO target_once VALUES (1)", Postgres: &ScriptTestAssertion{Query: "/* client A */ INSERT INTO target_once VALUES (1)"}},
		{Query: "/* client A */ COMMIT"},
		{Query: "/* client B */ SELECT id FROM target_once", Expected: []sql.Row{{1}}},
	}}})
}
