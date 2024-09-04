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

func TestMvcc(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_mvcc)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_mvcc,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL enable_seqscan = false;`,
			},
			{
				Statement: `SET LOCAL enable_indexonlyscan = false;`,
			},
			{
				Statement: `SET LOCAL enable_bitmapscan = false;`,
			},
			{
				Statement: `CREATE TABLE clean_aborted_self(key int, data text);`,
			},
			{
				Statement: `CREATE INDEX clean_aborted_self_key ON clean_aborted_self(key);`,
			},
			{
				Statement: `INSERT INTO clean_aborted_self (key, data) VALUES (-1, 'just to allocate metapage');`,
			},
			{
				Statement: `SELECT pg_relation_size('clean_aborted_self_key') AS clean_aborted_self_key_before \gset
DO $$
BEGIN
    -- iterate often enough to see index growth even on larger-than-default page sizes
    FOR i IN 1..100 LOOP
        BEGIN
	    -- perform index scan over all the inserted keys to get them to be seen as dead
            IF EXISTS(SELECT * FROM clean_aborted_self WHERE key > 0 AND key < 100) THEN
	        RAISE data_corrupted USING MESSAGE = 'these rows should not exist';`,
			},
			{
				Statement: `            END IF;`,
			},
			{
				Statement: `            INSERT INTO clean_aborted_self SELECT g.i, 'rolling back in a sec' FROM generate_series(1, 100) g(i);`,
			},
			{
				Statement: `	    -- just some error that's not normally thrown
	    RAISE reading_sql_data_not_permitted USING MESSAGE = 'round and round again';`,
			},
			{
				Statement: `	EXCEPTION WHEN reading_sql_data_not_permitted THEN END;`,
			},
			{
				Statement: `    END LOOP;`,
			},
			{
				Statement: `END;$$;`,
			},
			{
				Statement: `SELECT :clean_aborted_self_key_before AS size_before, pg_relation_size('clean_aborted_self_key') size_after
WHERE :clean_aborted_self_key_before != pg_relation_size('clean_aborted_self_key');`,
				Results: []sql.Row{},
			},
			{
				Statement: `ROLLBACK;`,
			},
		},
	})
}
