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
	"fmt"
	"testing"

	"github.com/dolthub/doltgresql/core/id"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestPublicationCatalog(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{Name: "catalog definitions and real owners", SetUpScript: []string{
			`CREATE PUBLICATION empty_pub`,
			`CREATE PUBLICATION all_pub FOR ALL TABLES WITH (publish='insert,delete',publish_via_partition_root=true)`,
			`CREATE ROLE publisher LOGIN PASSWORD 'test'`,
			`GRANT CREATE ON DATABASE postgres TO publisher`,
		}, Assertions: []ScriptTestAssertion{
			{Query: `SELECT pubname,puballtables,pubinsert,pubupdate,pubdelete,pubtruncate,pubviaroot FROM pg_publication ORDER BY pubname`, Expected: []sql.Row{
				{"all_pub", "t", "t", "f", "t", "f", "t"}, {"empty_pub", "f", "t", "t", "t", "t", "f"},
			}},
			{Query: `CREATE PUBLICATION mine WITH (publish='')`, Username: "publisher", Password: "test"},
			{Query: `SELECT p.pubname,r.rolname,pg_get_userbyid(p.pubowner) FROM pg_publication p JOIN pg_roles r ON r.oid=p.pubowner ORDER BY p.pubname`, Expected: []sql.Row{
				{"all_pub", "postgres", "postgres"}, {"empty_pub", "postgres", "postgres"}, {"mine", "publisher", "publisher"},
			}},
			{Query: `DROP PUBLICATION mine`, Username: "publisher", Password: "test"},
			{Query: `SELECT count(*) FROM pg_publication_rel`, Expected: []sql.Row{{0}}},
			{Query: `SELECT count(*) FROM pg_publication_namespace`, Expected: []sql.Row{{0}}},
			{Query: `SELECT count(*) FROM pg_publication_tables`, Expected: []sql.Row{{0}}},
			{Query: `DROP PUBLICATION all_pub,empty_pub`},
			{Query: `SELECT count(*) FROM pg_publication`, Expected: []sql.Row{{0}}},
		}},
		{Name: "all tables membership remains current", SetUpScript: []string{
			`CREATE TABLE first_table (i int, generated_col int GENERATED ALWAYS AS (i+1) STORED)`,
			`CREATE VIEW a_view AS SELECT i FROM first_table`, `CREATE SEQUENCE a_sequence`,
			`CREATE TEMP TABLE temp_table (i int)`,
			`CREATE PUBLICATION all_pub FOR ALL TABLES`, `CREATE PUBLICATION empty_pub`,
		}, Assertions: []ScriptTestAssertion{
			{Query: `SELECT pubname,schemaname,tablename,attnames,rowfilter FROM pg_publication_tables ORDER BY schemaname,tablename`, Expected: []sql.Row{
				{"all_pub", "public", "first_table", "{i,generated_col}", nil},
			}},
			{Query: `CREATE SCHEMA s`}, {Query: `CREATE TABLE s.later_table (a int)`},
			{Query: `ALTER TABLE first_table ADD COLUMN added text`},
			{Query: `SELECT tablename,attnames FROM pg_publication_tables ORDER BY tablename`, Expected: []sql.Row{
				{"first_table", "{i,generated_col,added}"}, {"later_table", "{a}"},
			}},
			{Query: `SELECT pg_relation_is_publishable('first_table'::regclass),pg_relation_is_publishable('a_view'::regclass),pg_relation_is_publishable('a_sequence'::regclass),pg_relation_is_publishable('pg_catalog.pg_class'::regclass),pg_relation_is_publishable(999999::regclass)`, Expected: []sql.Row{{"t", "f", "f", "f", nil}}},
			{Query: fmt.Sprintf("SELECT pg_relation_is_publishable(%d::regclass)", id.Cache().ToOID(id.NewTable("public", "temp_table").AsId())), Expected: []sql.Row{{"f"}}},
			{Query: `DROP TABLE first_table CASCADE`},
			{Query: `SELECT tablename FROM pg_publication_tables`, Expected: []sql.Row{{"later_table"}}},
		}},
		{Name: "staging a same named table does not stage a publication", SetUpScript: []string{
			`CREATE TABLE p (i int)`, `CREATE PUBLICATION p`,
		}, Assertions: []ScriptTestAssertion{
			{Query: `SELECT table_name,staged FROM dolt_status ORDER BY table_name`, Expected: []sql.Row{{"public.p", "f"}, {`publication "p"`, "f"}}},
			{Query: `SELECT dolt_add('p')`, Expected: []sql.Row{{0}}},
			{Query: `SELECT table_name,staged FROM dolt_status ORDER BY table_name`, Expected: []sql.Row{{"public.p", "t"}, {`publication "p"`, "f"}}},
			{Query: `SELECT dolt_add('-A')`, Expected: []sql.Row{{0}}},
			{Query: `SELECT table_name,staged FROM dolt_status ORDER BY table_name`, Expected: []sql.Row{{"public.p", "t"}, {`publication "p"`, "t"}}},
		}},
		{Name: "publication definitions follow branches", SetUpScript: []string{
			`SELECT dolt_commit('--allow-empty','-m','initial')`,
			`SELECT dolt_checkout('-b','pub_branch')`,
			`CREATE PUBLICATION p FOR ALL TABLES WITH (publish='truncate')`,
			`SELECT dolt_commit('-Am','create publication')`,
		}, Assertions: []ScriptTestAssertion{
			{Query: `SELECT pubname,pubinsert,pubtruncate FROM pg_publication`, Expected: []sql.Row{{"p", "f", "t"}}},
			{Query: `SELECT dolt_checkout('main')`, SkipResultsCheck: true},
			{Query: `SELECT count(*) FROM pg_publication`, Expected: []sql.Row{{0}}},
			{Query: `SELECT dolt_merge('pub_branch')`, SkipResultsCheck: true},
			{Query: `SELECT pubname,pubinsert,pubtruncate FROM pg_publication`, Expected: []sql.Row{{"p", "f", "t"}}},
		}},
		{Name: "competing publication definitions require conflict resolution", SetUpScript: []string{
			`CREATE PUBLICATION p WITH (publish='insert')`, `SELECT dolt_commit('-Am','initial')`,
			`SELECT dolt_branch('other_branch')`,
			`DROP PUBLICATION p`, `CREATE PUBLICATION p WITH (publish='delete')`, `SELECT dolt_commit('-Am','ours')`,
			`SELECT dolt_checkout('other_branch')`,
			`DROP PUBLICATION p`, `CREATE PUBLICATION p WITH (publish='update')`, `SELECT dolt_commit('-Am','theirs')`,
			`SELECT dolt_checkout('main')`,
		}, Assertions: []ScriptTestAssertion{
			{Query: `SELECT dolt_merge('other_branch')`, Expected: []sql.Row{{[]any{"", int64(0), int64(1), "conflicts found"}}}},
			{Query: `SELECT * FROM dolt_conflicts`, Expected: []sql.Row{{`publication "p"`, Numeric("1")}}},
			{Query: `SELECT pubupdate,pubdelete FROM pg_publication`, Expected: []sql.Row{{"f", "t"}}},
			{Query: `UPDATE "dolt_conflicts_publication ""p""" SET our_value=their_value`, Expected: []sql.Row{}},
			{Query: `SELECT count(*) FROM dolt_conflicts`, Expected: []sql.Row{{0}}},
			{Query: `SELECT pubupdate,pubdelete FROM pg_publication`, Expected: []sql.Row{{"t", "f"}}},
		}},
		{Name: "independent publication additions merge", SetUpScript: []string{
			`SELECT dolt_commit('--allow-empty','-m','initial')`,
			`SELECT dolt_branch('other_branch')`,
			`CREATE PUBLICATION p`, `SELECT dolt_commit('-Am','first publication')`,
			`SELECT dolt_checkout('other_branch')`,
			`CREATE PUBLICATION q`, `SELECT dolt_commit('-Am','second publication')`,
			`SELECT dolt_checkout('main')`,
		}, Assertions: []ScriptTestAssertion{
			{Query: `SELECT dolt_merge('other_branch')`, SkipResultsCheck: true},
			{Query: `SELECT pubname FROM pg_publication ORDER BY pubname`, Expected: []sql.Row{{"p"}, {"q"}}},
		}},
	})
}
