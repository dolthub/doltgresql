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

func TestHashPart(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_hash_part)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_hash_part,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE mchash (a int, b text, c jsonb)
  PARTITION BY HASH (a part_test_int4_ops, b part_test_text_ops);`,
			},
			{
				Statement: `CREATE TABLE mchash1
  PARTITION OF mchash FOR VALUES WITH (MODULUS 4, REMAINDER 0);`,
			},
			{
				Statement:   `SELECT satisfies_hash_partition(0, 4, 0, NULL);`,
				ErrorString: `could not open relation with OID 0`,
			},
			{
				Statement:   `SELECT satisfies_hash_partition('tenk1'::regclass, 4, 0, NULL);`,
				ErrorString: `"tenk1" is not a hash partitioned table`,
			},
			{
				Statement:   `SELECT satisfies_hash_partition('mchash1'::regclass, 4, 0, NULL);`,
				ErrorString: `"mchash1" is not a hash partitioned table`,
			},
			{
				Statement:   `SELECT satisfies_hash_partition('mchash'::regclass, 0, 0, NULL);`,
				ErrorString: `modulus for hash partition must be an integer value greater than zero`,
			},
			{
				Statement:   `SELECT satisfies_hash_partition('mchash'::regclass, 1, -1, NULL);`,
				ErrorString: `remainder for hash partition must be an integer value greater than or equal to zero`,
			},
			{
				Statement:   `SELECT satisfies_hash_partition('mchash'::regclass, 1, 1, NULL);`,
				ErrorString: `remainder for hash partition must be less than modulus`,
			},
			{
				Statement: `SELECT satisfies_hash_partition('mchash'::regclass, NULL, 0, NULL);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT satisfies_hash_partition('mchash'::regclass, 4, NULL, NULL);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement:   `SELECT satisfies_hash_partition('mchash'::regclass, 4, 0, NULL::int, NULL::text, NULL::json);`,
				ErrorString: `number of partitioning columns (2) does not match number of partition keys provided (3)`,
			},
			{
				Statement:   `SELECT satisfies_hash_partition('mchash'::regclass, 3, 1, NULL::int);`,
				ErrorString: `number of partitioning columns (2) does not match number of partition keys provided (1)`,
			},
			{
				Statement:   `SELECT satisfies_hash_partition('mchash'::regclass, 2, 1, NULL::int, NULL::int);`,
				ErrorString: `column 2 of the partition key has type text, but supplied value is of type integer`,
			},
			{
				Statement: `SELECT satisfies_hash_partition('mchash'::regclass, 4, 0, 0, ''::text);`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT satisfies_hash_partition('mchash'::regclass, 4, 0, 2, ''::text);`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT satisfies_hash_partition('mchash'::regclass, 2, 1,
								variadic array[1,2]::int[]);`,
				ErrorString: `column 2 of the partition key has type "text", but supplied value is of type "integer"`,
			},
			{
				Statement: `CREATE TABLE mcinthash (a int, b int, c jsonb)
  PARTITION BY HASH (a part_test_int4_ops, b part_test_int4_ops);`,
			},
			{
				Statement: `SELECT satisfies_hash_partition('mcinthash'::regclass, 4, 0,
								variadic array[0, 0]);`,
				Results: []sql.Row{{false}},
			},
			{
				Statement: `SELECT satisfies_hash_partition('mcinthash'::regclass, 4, 0,
								variadic array[0, 1]);`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `SELECT satisfies_hash_partition('mcinthash'::regclass, 4, 0,
								variadic array[]::int[]);`,
				ErrorString: `number of partitioning columns (2) does not match number of partition keys provided (0)`,
			},
			{
				Statement: `SELECT satisfies_hash_partition('mcinthash'::regclass, 4, 0,
								variadic array[now(), now()]);`,
				ErrorString: `column 1 of the partition key has type "integer", but supplied value is of type "timestamp with time zone"`,
			},
			{
				Statement: `create table text_hashp (a text) partition by hash (a);`,
			},
			{
				Statement: `create table text_hashp0 partition of text_hashp for values with (modulus 2, remainder 0);`,
			},
			{
				Statement: `create table text_hashp1 partition of text_hashp for values with (modulus 2, remainder 1);`,
			},
			{
				Statement: `select satisfies_hash_partition('text_hashp'::regclass, 2, 0, 'xxx'::text) OR
	   satisfies_hash_partition('text_hashp'::regclass, 2, 1, 'xxx'::text) AS satisfies;`,
				Results: []sql.Row{{true}},
			},
			{
				Statement: `DROP TABLE mchash;`,
			},
			{
				Statement: `DROP TABLE mcinthash;`,
			},
			{
				Statement: `DROP TABLE text_hashp;`,
			},
		},
	})
}
