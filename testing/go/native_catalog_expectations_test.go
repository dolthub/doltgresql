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

import "github.com/dolthub/go-mysql-server/sql"

// OIDs are runtime identities, not portable ordered names. Exercise the same
// equality/range/order lookup shapes with OIDs obtained from this fixture.
func nativeNamespaceIndexScripts() []ScriptTest {
	return []ScriptTest{{Name: "pg_namespace_index_lookups", SetUpScript: []string{
		"CREATE SCHEMA testschema1", "CREATE SCHEMA testschema2", "CREATE SCHEMA testschema3", "CREATE SCHEMA z_schema", "CREATE SCHEMA a_schema",
	}, Assertions: []ScriptTestAssertion{
		{Query: "SELECT nspname FROM pg_namespace WHERE oid='public'::regnamespace", Expected: []sql.Row{{"public"}}},
		{Query: "SELECT nspname FROM pg_namespace WHERE oid='information_schema'::regnamespace", Expected: []sql.Row{{"information_schema"}}},
		{Query: "SELECT nspname FROM pg_namespace WHERE oid='pg_catalog'::regnamespace", Expected: []sql.Row{{"pg_catalog"}}},
		{Query: "SELECT nspname FROM pg_namespace WHERE oid=0", Expected: []sql.Row{}},
		{Query: "SELECT nspname FROM pg_namespace WHERE oid IN ('testschema1'::regnamespace,'testschema3'::regnamespace) ORDER BY nspname", Expected: []sql.Row{{"testschema1"}, {"testschema3"}}},
		{Query: "SELECT nspname FROM pg_namespace WHERE oid > ('testschema2'::regnamespace::oid::bigint-1) AND oid < ('testschema2'::regnamespace::oid::bigint+1)", Expected: []sql.Row{{"testschema2"}}},
		{Query: "SELECT nspname FROM pg_namespace WHERE oid >= 'testschema2'::regnamespace AND oid <= 'testschema2'::regnamespace", Expected: []sql.Row{{"testschema2"}}},
		{Query: "SELECT nspname FROM pg_namespace WHERE oid BETWEEN 'testschema2'::regnamespace AND 'testschema2'::regnamespace", Expected: []sql.Row{{"testschema2"}}},
		{Query: "SELECT nspname FROM pg_namespace WHERE nspname > 'testschema1' AND nspname < 'testschema3'", Expected: []sql.Row{{"testschema2"}}},
		{Query: "SELECT nspname FROM pg_namespace WHERE nspname >= 'testschema1' AND nspname <= 'testschema3' ORDER BY nspname", Expected: []sql.Row{{"testschema1"}, {"testschema2"}, {"testschema3"}}},
		{Query: "SELECT nspname FROM pg_namespace WHERE nspname > 'testschema1' AND nspname <= 'testschema3' ORDER BY nspname", Expected: []sql.Row{{"testschema2"}, {"testschema3"}}},
		{Query: "SELECT nspname FROM pg_namespace WHERE nspname >= 'testschema1' AND nspname < 'testschema3' ORDER BY nspname", Expected: []sql.Row{{"testschema1"}, {"testschema2"}}},
		{Query: "SELECT nspname FROM pg_namespace WHERE nspname IN ('a_schema','z_schema') ORDER BY nspname DESC LIMIT 1", Expected: []sql.Row{{"z_schema"}}},
		{Query: "SELECT nspname FROM pg_namespace WHERE nspname='absent_schema'", Expected: []sql.Row{}},
		{Query: "SELECT count(*)=count(DISTINCT oid) FROM pg_namespace", Expected: []sql.Row{{"t"}}},
		{Query: "EXPLAIN (FORMAT JSON, COSTS OFF) SELECT nspname FROM pg_namespace WHERE oid='public'::regnamespace", ExpectedPlanRelations: []string{"pg_namespace"}, PostgresFork: &ScriptTestAssertion{ExpectedCatalogScanAliases: []string{"pg_namespace"}}},
		{Query: "EXPLAIN (FORMAT JSON, COSTS OFF) SELECT oid FROM pg_namespace WHERE nspname='testschema2'", ExpectedPlanRelations: []string{"pg_namespace"}, PostgresFork: &ScriptTestAssertion{ExpectedCatalogScanAliases: []string{"pg_namespace"}}},
	}}}
}
