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

func TestReloptions(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_reloptions)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_reloptions,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE reloptions_test(i INT) WITH (FiLLFaCToR=30,
	autovacuum_enabled = false, autovacuum_analyze_scale_factor = 0.2);`,
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = 'reloptions_test'::regclass;`,
				Results:   []sql.Row{{`{fillfactor=30,autovacuum_enabled=false,autovacuum_analyze_scale_factor=0.2}`}},
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (fillfactor=2);`,
				ErrorString: `value 2 out of bounds for option "fillfactor"`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (fillfactor=110);`,
				ErrorString: `value 110 out of bounds for option "fillfactor"`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (autovacuum_analyze_scale_factor = -10.0);`,
				ErrorString: `value -10.0 out of bounds for option "autovacuum_analyze_scale_factor"`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (autovacuum_analyze_scale_factor = 110.0);`,
				ErrorString: `value 110.0 out of bounds for option "autovacuum_analyze_scale_factor"`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (not_existing_option=2);`,
				ErrorString: `unrecognized parameter "not_existing_option"`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (not_existing_namespace.fillfactor=2);`,
				ErrorString: `unrecognized parameter namespace "not_existing_namespace"`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (fillfactor=-30.1);`,
				ErrorString: `value -30.1 out of bounds for option "fillfactor"`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (fillfactor='string');`,
				ErrorString: `invalid value for integer option "fillfactor": string`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (fillfactor=true);`,
				ErrorString: `invalid value for integer option "fillfactor": true`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (autovacuum_enabled=12);`,
				ErrorString: `invalid value for boolean option "autovacuum_enabled": 12`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (autovacuum_enabled=30.5);`,
				ErrorString: `invalid value for boolean option "autovacuum_enabled": 30.5`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (autovacuum_enabled='string');`,
				ErrorString: `invalid value for boolean option "autovacuum_enabled": string`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (autovacuum_analyze_scale_factor='string');`,
				ErrorString: `invalid value for floating point option "autovacuum_analyze_scale_factor": string`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (autovacuum_analyze_scale_factor=true);`,
				ErrorString: `invalid value for floating point option "autovacuum_analyze_scale_factor": true`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (fillfactor=30, fillfactor=40);`,
				ErrorString: `parameter "fillfactor" specified more than once`,
			},
			{
				Statement:   `CREATE TABLE reloptions_test2(i INT) WITH (fillfactor);`,
				ErrorString: `invalid value for integer option "fillfactor": true`,
			},
			{
				Statement: `ALTER TABLE reloptions_test SET (fillfactor=31,
	autovacuum_analyze_scale_factor = 0.3);`,
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = 'reloptions_test'::regclass;`,
				Results:   []sql.Row{{`{autovacuum_enabled=false,fillfactor=31,autovacuum_analyze_scale_factor=0.3}`}},
			},
			{
				Statement: `ALTER TABLE reloptions_test SET (autovacuum_enabled, fillfactor=32);`,
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = 'reloptions_test'::regclass;`,
				Results:   []sql.Row{{`{autovacuum_analyze_scale_factor=0.3,autovacuum_enabled=true,fillfactor=32}`}},
			},
			{
				Statement: `ALTER TABLE reloptions_test RESET (fillfactor);`,
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = 'reloptions_test'::regclass;`,
				Results:   []sql.Row{{`{autovacuum_analyze_scale_factor=0.3,autovacuum_enabled=true}`}},
			},
			{
				Statement: `ALTER TABLE reloptions_test RESET (autovacuum_enabled,
	autovacuum_analyze_scale_factor);`,
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = 'reloptions_test'::regclass AND
       reloptions IS NULL;`,
				Results: []sql.Row{{``}},
			},
			{
				Statement:   `ALTER TABLE reloptions_test RESET (fillfactor=12);`,
				ErrorString: `RESET must not include values for parameters`,
			},
			{
				Statement: `DROP TABLE reloptions_test;`,
			},
			{
				Statement: `CREATE TEMP TABLE reloptions_test(i INT NOT NULL, j text)
	WITH (vacuum_truncate=false,
	toast.vacuum_truncate=false,
	autovacuum_enabled=false);`,
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = 'reloptions_test'::regclass;`,
				Results:   []sql.Row{{`{vacuum_truncate=false,autovacuum_enabled=false}`}},
			},
			{
				Statement:   `INSERT INTO reloptions_test VALUES (1, NULL), (NULL, NULL);`,
				ErrorString: `null value in column "i" of relation "reloptions_test" violates not-null constraint`,
			},
			{
				Statement: `VACUUM (FREEZE, DISABLE_PAGE_SKIPPING) reloptions_test;`,
			},
			{
				Statement: `SELECT pg_relation_size('reloptions_test') > 0;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid =
	(SELECT reltoastrelid FROM pg_class
	WHERE oid = 'reloptions_test'::regclass);`,
				Results: []sql.Row{{`{vacuum_truncate=false}`}},
			},
			{
				Statement: `ALTER TABLE reloptions_test RESET (vacuum_truncate);`,
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = 'reloptions_test'::regclass;`,
				Results:   []sql.Row{{`{autovacuum_enabled=false}`}},
			},
			{
				Statement:   `INSERT INTO reloptions_test VALUES (1, NULL), (NULL, NULL);`,
				ErrorString: `null value in column "i" of relation "reloptions_test" violates not-null constraint`,
			},
			{
				Statement: `VACUUM (FREEZE, DISABLE_PAGE_SKIPPING) reloptions_test;`,
			},
			{
				Statement: `SELECT pg_relation_size('reloptions_test') = 0;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `DROP TABLE reloptions_test;`,
			},
			{
				Statement: `CREATE TABLE reloptions_test (s VARCHAR)
	WITH (toast.autovacuum_vacuum_cost_delay = 23);`,
			},
			{
				Statement: `SELECT reltoastrelid as toast_oid
	FROM pg_class WHERE oid = 'reloptions_test'::regclass \gset
SELECT reloptions FROM pg_class WHERE oid = :toast_oid;`,
				Results: []sql.Row{{`{autovacuum_vacuum_cost_delay=23}`}},
			},
			{
				Statement: `ALTER TABLE reloptions_test SET (toast.autovacuum_vacuum_cost_delay = 24);`,
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = :toast_oid;`,
				Results:   []sql.Row{{`{autovacuum_vacuum_cost_delay=24}`}},
			},
			{
				Statement: `ALTER TABLE reloptions_test RESET (toast.autovacuum_vacuum_cost_delay);`,
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = :toast_oid;`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `CREATE TABLE reloptions_test2 (i int) WITH (toast.not_existing_option = 42);`,
				ErrorString: `unrecognized parameter "not_existing_option"`,
			},
			{
				Statement: `DROP TABLE reloptions_test;`,
			},
			{
				Statement: `CREATE TABLE reloptions_test (s VARCHAR) WITH
	(toast.autovacuum_vacuum_cost_delay = 23,
	autovacuum_vacuum_cost_delay = 24, fillfactor = 40);`,
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = 'reloptions_test'::regclass;`,
				Results:   []sql.Row{{`{autovacuum_vacuum_cost_delay=24,fillfactor=40}`}},
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = (
	SELECT reltoastrelid FROM pg_class WHERE oid = 'reloptions_test'::regclass);`,
				Results: []sql.Row{{`{autovacuum_vacuum_cost_delay=23}`}},
			},
			{
				Statement: `CREATE INDEX reloptions_test_idx ON reloptions_test (s) WITH (fillfactor=30);`,
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = 'reloptions_test_idx'::regclass;`,
				Results:   []sql.Row{{`{fillfactor=30}`}},
			},
			{
				Statement: `CREATE INDEX reloptions_test_idx ON reloptions_test (s)
	WITH (not_existing_option=2);`,
				ErrorString: `unrecognized parameter "not_existing_option"`,
			},
			{
				Statement: `CREATE INDEX reloptions_test_idx ON reloptions_test (s)
	WITH (not_existing_ns.fillfactor=2);`,
				ErrorString: `unrecognized parameter namespace "not_existing_ns"`,
			},
			{
				Statement:   `CREATE INDEX reloptions_test_idx2 ON reloptions_test (s) WITH (fillfactor=1);`,
				ErrorString: `value 1 out of bounds for option "fillfactor"`,
			},
			{
				Statement:   `CREATE INDEX reloptions_test_idx2 ON reloptions_test (s) WITH (fillfactor=130);`,
				ErrorString: `value 130 out of bounds for option "fillfactor"`,
			},
			{
				Statement: `ALTER INDEX reloptions_test_idx SET (fillfactor=40);`,
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = 'reloptions_test_idx'::regclass;`,
				Results:   []sql.Row{{`{fillfactor=40}`}},
			},
			{
				Statement: `CREATE INDEX reloptions_test_idx3 ON reloptions_test (s);`,
			},
			{
				Statement: `ALTER INDEX reloptions_test_idx3 SET (fillfactor=40);`,
			},
			{
				Statement: `SELECT reloptions FROM pg_class WHERE oid = 'reloptions_test_idx3'::regclass;`,
				Results:   []sql.Row{{`{fillfactor=40}`}},
			},
		},
	})
}
