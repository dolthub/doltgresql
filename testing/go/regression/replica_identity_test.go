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

func TestReplicaIdentity(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_replica_identity)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_replica_identity,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE test_replica_identity (
       id serial primary key,
       keya text not null,
       keyb text not null,
       nonkey text,
       CONSTRAINT test_replica_identity_unique_defer UNIQUE (keya, keyb) DEFERRABLE,
       CONSTRAINT test_replica_identity_unique_nondefer UNIQUE (keya, keyb)
) ;`,
			},
			{
				Statement: `CREATE TABLE test_replica_identity_othertable (id serial primary key);`,
			},
			{
				Statement: `CREATE INDEX test_replica_identity_keyab ON test_replica_identity (keya, keyb);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX test_replica_identity_keyab_key ON test_replica_identity (keya, keyb);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX test_replica_identity_nonkey ON test_replica_identity (keya, nonkey);`,
			},
			{
				Statement: `CREATE INDEX test_replica_identity_hash ON test_replica_identity USING hash (nonkey);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX test_replica_identity_expr ON test_replica_identity (keya, keyb, (3));`,
			},
			{
				Statement: `CREATE UNIQUE INDEX test_replica_identity_partial ON test_replica_identity (keya, keyb) WHERE keyb != '3';`,
			},
			{
				Statement: `SELECT relreplident FROM pg_class WHERE oid = 'test_replica_identity'::regclass;`,
				Results:   []sql.Row{{`d`}},
			},
			{
				Statement: `SELECT relreplident FROM pg_class WHERE oid = 'pg_class'::regclass;`,
				Results:   []sql.Row{{`n`}},
			},
			{
				Statement: `SELECT relreplident FROM pg_class WHERE oid = 'pg_constraint'::regclass;`,
				Results:   []sql.Row{{`n`}},
			},
			{
				Statement: `----
----
ALTER TABLE test_replica_identity REPLICA IDENTITY USING INDEX test_replica_identity_keyab;`,
				ErrorString: `cannot use non-unique index "test_replica_identity_keyab" as replica identity`,
			},
			{
				Statement:   `ALTER TABLE test_replica_identity REPLICA IDENTITY USING INDEX test_replica_identity_nonkey;`,
				ErrorString: `index "test_replica_identity_nonkey" cannot be used as replica identity because column "nonkey" is nullable`,
			},
			{
				Statement:   `ALTER TABLE test_replica_identity REPLICA IDENTITY USING INDEX test_replica_identity_hash;`,
				ErrorString: `cannot use non-unique index "test_replica_identity_hash" as replica identity`,
			},
			{
				Statement:   `ALTER TABLE test_replica_identity REPLICA IDENTITY USING INDEX test_replica_identity_expr;`,
				ErrorString: `cannot use expression index "test_replica_identity_expr" as replica identity`,
			},
			{
				Statement:   `ALTER TABLE test_replica_identity REPLICA IDENTITY USING INDEX test_replica_identity_partial;`,
				ErrorString: `cannot use partial index "test_replica_identity_partial" as replica identity`,
			},
			{
				Statement:   `ALTER TABLE test_replica_identity REPLICA IDENTITY USING INDEX test_replica_identity_othertable_pkey;`,
				ErrorString: `"test_replica_identity_othertable_pkey" is not an index for table "test_replica_identity"`,
			},
			{
				Statement:   `ALTER TABLE test_replica_identity REPLICA IDENTITY USING INDEX test_replica_identity_unique_defer;`,
				ErrorString: `cannot use non-immediate index "test_replica_identity_unique_defer" as replica identity`,
			},
			{
				Statement: `SELECT relreplident FROM pg_class WHERE oid = 'test_replica_identity'::regclass;`,
				Results:   []sql.Row{{`d`}},
			},
			{
				Statement: `----
----
ALTER TABLE test_replica_identity REPLICA IDENTITY USING INDEX test_replica_identity_pkey;`,
			},
			{
				Statement: `SELECT relreplident FROM pg_class WHERE oid = 'test_replica_identity'::regclass;`,
				Results:   []sql.Row{{`i`}},
			},
			{
				Statement: `\d test_replica_identity
                            Table "public.test_replica_identity"
 Column |  Type   | Collation | Nullable |                      Default                      
--------+---------+-----------+----------+---------------------------------------------------
 id     | integer |           | not null | nextval('test_replica_identity_id_seq'::regclass)
 keya   | text    |           | not null | 
 keyb   | text    |           | not null | 
 nonkey | text    |           |          | 
Indexes:
    "test_replica_identity_pkey" PRIMARY KEY, btree (id) REPLICA IDENTITY
    "test_replica_identity_expr" UNIQUE, btree (keya, keyb, (3))
    "test_replica_identity_hash" hash (nonkey)
    "test_replica_identity_keyab" btree (keya, keyb)
    "test_replica_identity_keyab_key" UNIQUE, btree (keya, keyb)
    "test_replica_identity_nonkey" UNIQUE, btree (keya, nonkey)
    "test_replica_identity_partial" UNIQUE, btree (keya, keyb) WHERE keyb <> '3'::text
    "test_replica_identity_unique_defer" UNIQUE CONSTRAINT, btree (keya, keyb) DEFERRABLE
    "test_replica_identity_unique_nondefer" UNIQUE CONSTRAINT, btree (keya, keyb)
ALTER TABLE test_replica_identity REPLICA IDENTITY USING INDEX test_replica_identity_unique_nondefer;`,
			},
			{
				Statement: `ALTER TABLE test_replica_identity REPLICA IDENTITY USING INDEX test_replica_identity_keyab_key;`,
			},
			{
				Statement: `ALTER TABLE test_replica_identity REPLICA IDENTITY USING INDEX test_replica_identity_keyab_key;`,
			},
			{
				Statement: `SELECT relreplident FROM pg_class WHERE oid = 'test_replica_identity'::regclass;`,
				Results:   []sql.Row{{`i`}},
			},
			{
				Statement: `\d test_replica_identity
                            Table "public.test_replica_identity"
 Column |  Type   | Collation | Nullable |                      Default                      
--------+---------+-----------+----------+---------------------------------------------------
 id     | integer |           | not null | nextval('test_replica_identity_id_seq'::regclass)
 keya   | text    |           | not null | 
 keyb   | text    |           | not null | 
 nonkey | text    |           |          | 
Indexes:
    "test_replica_identity_pkey" PRIMARY KEY, btree (id)
    "test_replica_identity_expr" UNIQUE, btree (keya, keyb, (3))
    "test_replica_identity_hash" hash (nonkey)
    "test_replica_identity_keyab" btree (keya, keyb)
    "test_replica_identity_keyab_key" UNIQUE, btree (keya, keyb) REPLICA IDENTITY
    "test_replica_identity_nonkey" UNIQUE, btree (keya, nonkey)
    "test_replica_identity_partial" UNIQUE, btree (keya, keyb) WHERE keyb <> '3'::text
    "test_replica_identity_unique_defer" UNIQUE CONSTRAINT, btree (keya, keyb) DEFERRABLE
    "test_replica_identity_unique_nondefer" UNIQUE CONSTRAINT, btree (keya, keyb)
SELECT count(*) FROM pg_index WHERE indrelid = 'test_replica_identity'::regclass AND indisreplident;`,
				Results: []sql.Row{{1}},
			},
			{
				Statement: `----
----
ALTER TABLE test_replica_identity REPLICA IDENTITY DEFAULT;`,
			},
			{
				Statement: `SELECT relreplident FROM pg_class WHERE oid = 'test_replica_identity'::regclass;`,
				Results:   []sql.Row{{`d`}},
			},
			{
				Statement: `SELECT count(*) FROM pg_index WHERE indrelid = 'test_replica_identity'::regclass AND indisreplident;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `ALTER TABLE test_replica_identity REPLICA IDENTITY FULL;`,
			},
			{
				Statement: `SELECT relreplident FROM pg_class WHERE oid = 'test_replica_identity'::regclass;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `\d+ test_replica_identity
                                                Table "public.test_replica_identity"
 Column |  Type   | Collation | Nullable |                      Default                      | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------------------------------------------------+----------+--------------+-------------
 id     | integer |           | not null | nextval('test_replica_identity_id_seq'::regclass) | plain    |              | 
 keya   | text    |           | not null |                                                   | extended |              | 
 keyb   | text    |           | not null |                                                   | extended |              | 
 nonkey | text    |           |          |                                                   | extended |              | 
Indexes:
    "test_replica_identity_pkey" PRIMARY KEY, btree (id)
    "test_replica_identity_expr" UNIQUE, btree (keya, keyb, (3))
    "test_replica_identity_hash" hash (nonkey)
    "test_replica_identity_keyab" btree (keya, keyb)
    "test_replica_identity_keyab_key" UNIQUE, btree (keya, keyb)
    "test_replica_identity_nonkey" UNIQUE, btree (keya, nonkey)
    "test_replica_identity_partial" UNIQUE, btree (keya, keyb) WHERE keyb <> '3'::text
    "test_replica_identity_unique_defer" UNIQUE CONSTRAINT, btree (keya, keyb) DEFERRABLE
    "test_replica_identity_unique_nondefer" UNIQUE CONSTRAINT, btree (keya, keyb)
Replica Identity: FULL
ALTER TABLE test_replica_identity REPLICA IDENTITY NOTHING;`,
			},
			{
				Statement: `SELECT relreplident FROM pg_class WHERE oid = 'test_replica_identity'::regclass;`,
				Results:   []sql.Row{{`n`}},
			},
			{
				Statement: `---
---
CREATE TABLE test_replica_identity2 (id int UNIQUE NOT NULL);`,
			},
			{
				Statement: `ALTER TABLE test_replica_identity2 REPLICA IDENTITY USING INDEX test_replica_identity2_id_key;`,
			},
			{
				Statement: `\d test_replica_identity2
       Table "public.test_replica_identity2"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 id     | integer |           | not null | 
Indexes:
    "test_replica_identity2_id_key" UNIQUE CONSTRAINT, btree (id) REPLICA IDENTITY
ALTER TABLE test_replica_identity2 ALTER COLUMN id TYPE bigint;`,
			},
			{
				Statement: `\d test_replica_identity2
      Table "public.test_replica_identity2"
 Column |  Type  | Collation | Nullable | Default 
--------+--------+-----------+----------+---------
 id     | bigint |           | not null | 
Indexes:
    "test_replica_identity2_id_key" UNIQUE CONSTRAINT, btree (id) REPLICA IDENTITY
CREATE TABLE test_replica_identity3 (id int NOT NULL);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX test_replica_identity3_id_key ON test_replica_identity3 (id);`,
			},
			{
				Statement: `ALTER TABLE test_replica_identity3 REPLICA IDENTITY USING INDEX test_replica_identity3_id_key;`,
			},
			{
				Statement: `\d test_replica_identity3
       Table "public.test_replica_identity3"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 id     | integer |           | not null | 
Indexes:
    "test_replica_identity3_id_key" UNIQUE, btree (id) REPLICA IDENTITY
ALTER TABLE test_replica_identity3 ALTER COLUMN id TYPE bigint;`,
			},
			{
				Statement: `\d test_replica_identity3
      Table "public.test_replica_identity3"
 Column |  Type  | Collation | Nullable | Default 
--------+--------+-----------+----------+---------
 id     | bigint |           | not null | 
Indexes:
    "test_replica_identity3_id_key" UNIQUE, btree (id) REPLICA IDENTITY
ALTER TABLE test_replica_identity3 ALTER COLUMN id DROP NOT NULL;`,
				ErrorString: `column "id" is in index used as replica identity`,
			},
			{
				Statement: `CREATE TABLE test_replica_identity4(id integer NOT NULL) PARTITION BY LIST (id);`,
			},
			{
				Statement: `CREATE TABLE test_replica_identity4_1(id integer NOT NULL);`,
			},
			{
				Statement: `ALTER TABLE ONLY test_replica_identity4
  ATTACH PARTITION test_replica_identity4_1 FOR VALUES IN (1);`,
			},
			{
				Statement: `ALTER TABLE ONLY test_replica_identity4
  ADD CONSTRAINT test_replica_identity4_pkey PRIMARY KEY (id);`,
			},
			{
				Statement: `ALTER TABLE ONLY test_replica_identity4
  REPLICA IDENTITY USING INDEX test_replica_identity4_pkey;`,
			},
			{
				Statement: `ALTER TABLE ONLY test_replica_identity4_1
  ADD CONSTRAINT test_replica_identity4_1_pkey PRIMARY KEY (id);`,
			},
			{
				Statement: `\d+ test_replica_identity4
                    Partitioned table "public.test_replica_identity4"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 id     | integer |           | not null |         | plain   |              | 
Partition key: LIST (id)
Indexes:
    "test_replica_identity4_pkey" PRIMARY KEY, btree (id) INVALID REPLICA IDENTITY
Partitions: test_replica_identity4_1 FOR VALUES IN (1)
ALTER INDEX test_replica_identity4_pkey
  ATTACH PARTITION test_replica_identity4_1_pkey;`,
			},
			{
				Statement: `\d+ test_replica_identity4
                    Partitioned table "public.test_replica_identity4"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 id     | integer |           | not null |         | plain   |              | 
Partition key: LIST (id)
Indexes:
    "test_replica_identity4_pkey" PRIMARY KEY, btree (id) REPLICA IDENTITY
Partitions: test_replica_identity4_1 FOR VALUES IN (1)
DROP TABLE test_replica_identity;`,
			},
			{
				Statement: `DROP TABLE test_replica_identity2;`,
			},
			{
				Statement: `DROP TABLE test_replica_identity3;`,
			},
			{
				Statement: `DROP TABLE test_replica_identity4;`,
			},
			{
				Statement: `DROP TABLE test_replica_identity_othertable;`,
			},
		},
	})
}
