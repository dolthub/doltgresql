// Copyright 2024 Dolthub, Inc.
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

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestNamespace(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_namespace)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_namespace,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT pg_catalog.set_config('search_path', ' ', false);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `CREATE SCHEMA test_ns_schema_1
       CREATE UNIQUE INDEX abc_a_idx ON abc (a)
       CREATE VIEW abc_view AS
              SELECT a+1 AS a, b+1 AS b FROM abc
       CREATE TABLE abc (
              a serial,
              b int UNIQUE
       );`,
			},
			{
				Statement: `SET search_path to public;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET search_path to public, test_ns_schema_1;`,
			},
			{
				Statement: `CREATE SCHEMA test_ns_schema_2
       CREATE VIEW abc_view AS SELECT c FROM abc;`,
				ErrorString: `column "c" does not exist`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SHOW search_path;`,
				Results:   []sql.Row{{`public`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET search_path to public, test_ns_schema_1;`,
			},
			{
				Statement: `CREATE SCHEMA test_ns_schema_2
       CREATE VIEW abc_view AS SELECT a FROM abc;`,
			},
			{
				Statement: `SHOW search_path;`,
				Results:   []sql.Row{{`public, test_ns_schema_1`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SHOW search_path;`,
				Results:   []sql.Row{{`public, test_ns_schema_1`}},
			},
			{
				Statement: `DROP SCHEMA test_ns_schema_2 CASCADE;`,
			},
			{
				Statement: `SELECT COUNT(*) FROM pg_class WHERE relnamespace =
    (SELECT oid FROM pg_namespace WHERE nspname = 'test_ns_schema_1');`,
				Results: []sql.Row{{5}},
			},
			{
				Statement: `INSERT INTO test_ns_schema_1.abc DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO test_ns_schema_1.abc DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO test_ns_schema_1.abc DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT * FROM test_ns_schema_1.abc;`,
				Results:   []sql.Row{{1, ``}, {2, ``}, {3, ``}},
			},
			{
				Statement: `SELECT * FROM test_ns_schema_1.abc_view;`,
				Results:   []sql.Row{{2, ``}, {3, ``}, {4, ``}},
			},
			{
				Statement: `ALTER SCHEMA test_ns_schema_1 RENAME TO test_ns_schema_renamed;`,
			},
			{
				Statement: `SELECT COUNT(*) FROM pg_class WHERE relnamespace =
    (SELECT oid FROM pg_namespace WHERE nspname = 'test_ns_schema_1');`,
				Results: []sql.Row{{0}},
			},
			{
				Statement:   `CREATE SCHEMA test_ns_schema_renamed; -- fail, already exists`,
				ErrorString: `schema "test_ns_schema_renamed" already exists`,
			},
			{
				Statement: `CREATE SCHEMA IF NOT EXISTS test_ns_schema_renamed; -- ok with notice`,
			},
			{
				Statement: `CREATE SCHEMA IF NOT EXISTS test_ns_schema_renamed -- fail, disallowed
       CREATE TABLE abc (
              a serial,
              b int UNIQUE
       );`,
				ErrorString: `CREATE SCHEMA IF NOT EXISTS cannot include schema elements`,
			},
			{
				Statement: `DROP SCHEMA test_ns_schema_renamed CASCADE;`,
			},
			{
				Statement: `SELECT COUNT(*) FROM pg_class WHERE relnamespace =
    (SELECT oid FROM pg_namespace WHERE nspname = 'test_ns_schema_renamed');`,
				Results: []sql.Row{{0}},
			},
		},
	})
}
