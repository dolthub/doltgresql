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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestAlterTable(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Add Foreign Key Constraint",
			SetUpScript: []string{
				"create table child (pk int primary key, c1 int);",
				"insert into child values (1,1), (2,2), (3,3);",
				"create index idx_child_c1 on child (pk, c1);",
				"create table parent (pk int primary key, c1 int, c2 int);",
				"insert into parent values (1, 1, 10);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE parent ADD FOREIGN KEY (c1) REFERENCES child (pk) ON DELETE CASCADE;",
					Expected: []sql.Row{},
				},
				{
					// Test that the FK constraint is working
					Query:       "INSERT INTO parent VALUES (10, 10, 10);",
					ExpectedErr: "Foreign key violation",
				},
				{
					Query:       "ALTER TABLE parent ADD FOREIGN KEY (c2) REFERENCES child (pk);",
					ExpectedErr: "Foreign key violation",
				},
				{
					// Test an FK reference over multiple columns
					Query:       "ALTER TABLE parent ADD FOREIGN KEY (c1, c2) REFERENCES child (pk, c1);",
					ExpectedErr: "Foreign key violation",
				},
				{
					// Unsupported syntax: MATCH PARTIAL
					Query:       "ALTER TABLE parent ADD FOREIGN KEY (c1, c2) REFERENCES child (pk, c1) MATCH PARTIAL;",
					ExpectedErr: "MATCH PARTIAL is not yet supported",
				},
			},
		},
		{
			Name: "Add Unique Constraint",
			SetUpScript: []string{
				"create table t1 (pk int primary key, c1 int);",
				"insert into t1 values (1,1);",
				"create table t2 (pk int primary key, c1 int);",
				"insert into t2 values (1,1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// Add a secondary unique index using create index
					Query:    "CREATE UNIQUE INDEX ON t1(c1);",
					Expected: []sql.Row{},
				},
				{
					// Test that the unique constraint is working
					Query:       "INSERT INTO t1 VALUES (2, 1);",
					ExpectedErr: "unique",
				},
				{
					// Add a secondary unique index using alter table
					Query:    "ALTER TABLE t2 ADD CONSTRAINT uniq1 UNIQUE (c1);",
					Expected: []sql.Row{},
				},
				{
					// Test that the unique constraint is working
					Query:       "INSERT INTO t2 VALUES (2, 1);",
					ExpectedErr: "unique",
				},
			},
		},
		{
			Name: "Add Check Constraint",
			SetUpScript: []string{
				"create table t1 (pk int primary key, c1 int);",
				"insert into t1 values (1,1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// Add a check constraint that is already violated by the existing data
					Query:       "ALTER TABLE t1 ADD CONSTRAINT constraint1 CHECK (c1 > 100);",
					ExpectedErr: "violated",
				},
				{
					// Add a check constraint
					Query:    "ALTER TABLE t1 ADD CONSTRAINT constraint1 CHECK (c1 < 100);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t1 VALUES (2, 2);",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO t1 VALUES (3, 101);",
					ExpectedErr: "violated",
				},
			},
		},
		{
			Name: "Add Check Constraint with IN tuple",
			SetUpScript: []string{
				"create table t1 (pk int primary key, c1 int);",
				"insert into t1 values (1,1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// Add a check constraint that is already violated by the existing data
					Query:       "ALTER TABLE t1 ADD CONSTRAINT constraint1 CHECK (c1 in (100));",
					ExpectedErr: "violated",
				},
				{
					// Add a check constraint
					Query:    "ALTER TABLE t1 ADD CONSTRAINT constraint1 CHECK (c1 in (1,2));",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t1 VALUES (2, 2);",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO t1 VALUES (3, 101);",
					ExpectedErr: "violated",
				},
			},
		},
		{
			Name: "Add Check Constraint and another constraint in same statement",
			SetUpScript: []string{
				"create table t1 (pk int, c1 int);",
				"insert into t1 values (1,1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					// Add a check constraint
					Query:    " ALTER TABLE t1 ADD CONSTRAINT check_a CHECK (c1 IN (1)), ALTER c1 SET NOT NULL;",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO t1 VALUES (2, 2);",
					ExpectedErr: "violated",
				},
				{
					Query:       "INSERT INTO t1 VALUES (1, NULL);",
					ExpectedErr: "non-nullable",
				},
			},
		},
		{
			Name: "Drop Constraint",
			SetUpScript: []string{
				"create table t1 (pk int primary key, c1 int);",
				"ALTER TABLE t1 ADD CONSTRAINT constraint1 CHECK (c1 > 100);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE t1 DROP CONSTRAINT constraint1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO t1 VALUES (1, 1);",
					Expected: []sql.Row{},
				},
				{
					Query:       "ALTER TABLE t1 DROP CONSTRAINT doesnotexist;",
					ExpectedErr: "does not exist",
				},
				{
					Query:    "ALTER TABLE t1 DROP CONSTRAINT IF EXISTS doesnotexist;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Add Primary Key",
			SetUpScript: []string{
				"CREATE TABLE test1 (a INT, b INT);",
				"CREATE TABLE test2 (a INT, b INT, c INT);",
				"CREATE TABLE pkTable1 (a INT PRIMARY KEY);",
				"CREATE TABLE duplicateRows (a INT, b INT);",
				"INSERT INTO duplicateRows VALUES (1, 2), (1, 2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE test1 ADD PRIMARY KEY (a);",
					Expected: []sql.Row{},
				},
				{
					// Test the pk by inserting a duplicate value
					Query:       "INSERT into test1 values (1, 2), (1, 3);",
					ExpectedErr: "duplicate primary key",
				},
				{
					Query:    "ALTER TABLE test2 ADD PRIMARY KEY (a, b);",
					Expected: []sql.Row{},
				},
				{
					// Test the pk by inserting a duplicate value
					Query:       "INSERT into test2 values (1, 2, 3), (1, 2, 4);",
					ExpectedErr: "duplicate primary key",
				},
				{
					Query:       "ALTER TABLE pkTable1 ADD PRIMARY KEY (a);",
					ExpectedErr: "Multiple primary keys defined",
				},
				{
					Query:       "ALTER TABLE duplicateRows ADD PRIMARY KEY (a);",
					ExpectedErr: "duplicate primary key",
				},
				{
					// TODO: This statement fails in analysis, because it can't find a table named
					//       doesNotExist – since IF EXISTS is specified, the analyzer should skip
					//       errors on resolving the table in this case.
					Skip:     true,
					Query:    "ALTER TABLE IF EXISTS doesNotExist ADD PRIMARY KEY (a, b);",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Add Primary Key on text column",
			SetUpScript: []string{
				"CREATE TABLE test1 (a text, b INT);",
				"insert into test1 values ('a', 1), ('b', 2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE test1 ADD PRIMARY KEY (a);",
					Expected: []sql.Row{},
				},
				{
					// Test the pk by inserting a duplicate value
					Query:       "INSERT into test1 values ('a', 3);",
					ExpectedErr: "duplicate primary key",
				},
				{
					Query: "select * from test1;",
					Expected: []sql.Row{
						{"a", 1},
						{"b", 2},
					},
				},
			},
		},
		{
			Name: "Add primary key with generated column",
			SetUpScript: []string{
				`CREATE TABLE t1 (
      id uuid DEFAULT public.gen_random_uuid() NOT NULL,
      data jsonb,
      has_data boolean GENERATED ALWAYS AS ((data IS NOT NULL)) STORED
  );`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            " ALTER TABLE ONLY t1 ADD CONSTRAINT pk PRIMARY KEY (id);",
					SkipResultsCheck: true, // only care if it doesn't error
				},
				{
					Query:            "insert into t1 (id, data) values (default, '{}');",
					SkipResultsCheck: true, // only care if it doesn't error
				},
				{
					Query: "Select has_data from t1;",
					Expected: []sql.Row{
						{"t"},
					},
				},
			},
		},
		{
			Name: "Add Column",
			SetUpScript: []string{
				"CREATE TABLE test1 (a INT, b INT);",
				"INSERT INTO test1 VALUES (1, 1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE test1 ADD COLUMN c INT NOT NULL DEFAULT 42;",
					Expected: []sql.Row{},
				},
				{
					Query:    "select * from test1;",
					Expected: []sql.Row{{1, 1, 42}},
				},
				{
					Query:       "ALTER TABLE test1 ADD COLUMN l non_existing_type;",
					ExpectedErr: `type "non_existing_type" does not exist`,
				},
				{
					Query:    `ALTER TABLE test1 ADD COLUMN m xid;`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Add column with inline check constraint",
			SetUpScript: []string{
				"CREATE TABLE test1 (a INT, b INT);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE test1 ADD COLUMN c INT NOT NULL DEFAULT 42 CONSTRAINT chk1 CHECK (c > 0);",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO test1 VALUES (2, 2, -2);",
					ExpectedErr: `Check constraint "chk1" violated`,
				},
			},
		},
		{
			Name: "Drop Column",
			SetUpScript: []string{
				"CREATE TABLE test1 (a INT, b INT, c INT, d INT);",
				"INSERT INTO test1 VALUES (1, 2, 3, 4);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE test1 DROP COLUMN c;",
					Expected: []sql.Row{},
				},
				{
					Query:    "select * from test1;",
					Expected: []sql.Row{{1, 2, 4}},
				},
				{
					Query:    "ALTER TABLE test1 DROP COLUMN d;",
					Expected: []sql.Row{},
				},
				{
					Query:    "select * from test1;",
					Expected: []sql.Row{{1, 2}},
				},
				{
					// TODO: Skipped until we support conditional execution on existence of column
					Skip:     true,
					Query:    "ALTER TABLE test1 DROP COLUMN IF EXISTS zzz;",
					Expected: []sql.Row{},
				},
				{
					// TODO: Even though we're setting IF EXISTS, this query still fails with an
					//       error about the table not existing.
					Skip:     true,
					Query:    "ALTER TABLE IF EXISTS doesNotExist DROP COLUMN z;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Rename Column",
			SetUpScript: []string{
				"CREATE TABLE test1 (a INT, b INT, c INT, d INT);",
				"INSERT INTO test1 VALUES (1, 2, 3, 4);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE test1 RENAME COLUMN c to jjj;",
					Expected: []sql.Row{},
				},
				{
					Query:    "select * from test1 where jjj=3;",
					Expected: []sql.Row{{1, 2, 3, 4}},
				},
			},
		},
		{
			Name: "Set Column Default",
			SetUpScript: []string{
				"CREATE TABLE test1 (a INT, b INT DEFAULT 42, c INT);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE test1 ALTER COLUMN c SET DEFAULT 43;",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test1 (a) VALUES (1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test1;",
					Expected: []sql.Row{{1, 42, 43}},
				},
				{
					Query:    "ALTER TABLE test1 ALTER COLUMN b DROP DEFAULT;",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test1 (a) VALUES (2);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test1 where a = 2;",
					Expected: []sql.Row{{2, nil, 43}},
				},
				{
					Query:    "ALTER TABLE test1 ALTER COLUMN c SET DEFAULT length('hello world');",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test1 (a) VALUES (3);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test1 where a = 3;",
					Expected: []sql.Row{{3, nil, 11}},
				},
			},
		},
		{
			Name: "Set Column Nullability",
			SetUpScript: []string{
				"CREATE TABLE test1 (a INT, b INT);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE test1 ALTER COLUMN b SET NOT NULL;",
					Expected: []sql.Row{},
				},
				{
					Query:       "INSERT INTO test1 VALUES (1, NULL);",
					ExpectedErr: "column name 'b' is non-nullable",
				},
				{
					Query:    "ALTER TABLE test1 ALTER COLUMN b DROP NOT NULL;",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test1 VALUES (2, NULL);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test1 where a = 2;",
					Expected: []sql.Row{{2, nil}},
				},
				{
					Query:       "ALTER TABLE test1 ALTER COLUMN b SET NOT NULL;",
					ExpectedErr: "'b' is non-nullable but attempted to set a value of null",
				},
			},
		},
		{
			Name: "Alter Column Type",
			SetUpScript: []string{
				"CREATE TABLE test1 (a INT, b smallint);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "INSERT INTO test1 VALUES (1, 32769);",
					ExpectedErr: "smallint out of range",
				},
				{
					Query:    "ALTER TABLE test1 ALTER COLUMN b TYPE INT;",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test1 VALUES (1, 32769);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT * FROM test1;",
					Expected: []sql.Row{{1, 32769}},
				},
				{
					// Attempting to change to a smaller type that doesn't support the values in the
					// column results in an error instead of changing the type.
					Query:       "ALTER TABLE test1 ALTER COLUMN b TYPE smallint;",
					ExpectedErr: "smallint: unhandled type: int32",
				},
			},
		},
		{
			Name: "ALTER COLUMN resolves column default expressions",
			SetUpScript: []string{
				"CREATE TABLE t1 (id VARCHAR PRIMARY KEY, c1 TIMESTAMP DEFAULT CURRENT_TIMESTAMP);",
				"CREATE TABLE t2 (id VARCHAR PRIMARY KEY, c1 VARCHAR(100) DEFAULT concat('f', 'oo'));",
				"CREATE TABLE t3 (id VARCHAR PRIMARY KEY, c1 VARCHAR(20) NOT NULL DEFAULT CONCAT('f', 'oo'));",
				"CREATE TABLE t4 (id VARCHAR PRIMARY KEY, c1 VARCHAR(100) DEFAULT CONCAT('f', 'oo'));",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE t1 ALTER COLUMN c1 SET NOT NULL;",
					Expected: []sql.Row{},
				},
				{
					Query:    "ALTER TABLE t2 ALTER COLUMN c1 TYPE VARCHAR(50);",
					Expected: []sql.Row{},
				},
				{
					Query:    "ALTER TABLE t3 ALTER COLUMN c1 DROP NOT NULL;",
					Expected: []sql.Row{},
				},
				{
					Query:    "ALTER TABLE t4 RENAME COLUMN c1 TO ccc1;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "ALTER TABLE ADD COLUMN with inline FK constraint",
			SetUpScript: []string{
				"create table t (v varchar(100));",
				"create table parent (id int primary key);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE t ADD COLUMN c1 int REFERENCES parent(id);",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT conname AS constraint_name, pg_get_constraintdef(oid) AS constraint_definition FROM pg_constraint WHERE conrelid = 't'::regclass AND contype='f';",
					Expected: []sql.Row{{"t_c1_fkey", "FOREIGN KEY (c1) REFERENCES parent(id)"}},
				},
				{
					Query:       "INSERT INTO t VALUES ('abc', 123);",
					ExpectedErr: "Foreign key violation on fk: `t_c1_fkey`",
				},
			},
		},
		{
			Name: "Rename table",
			SetUpScript: []string{
				"create schema s1",
				"create schema s2",
				"CREATE TABLE t1 (a INT, b INT);",
				"INSERT INTO t1 VALUES (1, 2);",
				"CREATE TABLE t2 (c INT, d INT);",
				"INSERT INTO t2 VALUES (3, 4);",
				"create table s1.t1 (e INT, f INT);",
				"INSERT INTO s1.t1 VALUES (5, 6);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       "ALTER TABLE doesnotexist RENAME TO t3;",
					ExpectedErr: "not found",
				},
				{
					Query: "ALTER TABLE t1 RENAME TO t3;",
				},
				{
					Query: "SELECT * FROM t3;",
					Expected: []sql.Row{
						{1, 2},
					},
				},
				{
					Query: "SELECT * FROM public.t3;",
					Expected: []sql.Row{
						{1, 2},
					},
				},
				{
					Query:       "SELECT * FROM t1;",
					ExpectedErr: "not found",
				},
				{
					Query:       "ALTER TABLE t3 RENAME TO t2;",
					ExpectedErr: "already exists",
				},
				{
					Query: "ALTER TABLE s1.t1 RENAME TO t4;",
					Skip:  true, // schema names not supported yet
				},
				{
					Query: "SELECT * FROM s1.t4;",
					Expected: []sql.Row{
						{5, 6},
					},
					Skip: true, // schema names not supported yet
				},
			},
		},
		{
			Name: "alter table owner",
			SetUpScript: []string{
				"CREATE TABLE t1 (a INT, b INT);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER TABLE t1 OWNER TO new_owner;", // no error is all we expect here
				},
			},
		},
		{
			Name: "alter table add primary key with timestamp column default values",
			SetUpScript: []string{
				`CREATE TABLE t1 (
    id int NOT NULL,
    uid uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);`,
				"INSERT INTO t1 (id, uid) VALUES (1, '00000000-0000-0000-0000-000000000001');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER TABLE ONLY public.t1 ADD CONSTRAINT t1_pkey PRIMARY KEY (id);",
				},
				{
					Query:    "select created_at is not null from t1 where id = 1;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "select updated_at is not null from t1 where id = 1;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "select created_at = updated_at from t1 where id = 1;",
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "alter table add primary key with uuid column default values",
			SetUpScript: []string{
				`CREATE TABLE t1 (
    id int NOT NULL,
    uid uuid default gen_random_uuid() NOT NULL
);`,
				"INSERT INTO t1 (id) VALUES (1);",
				"INSERT INTO t1 (id) VALUES (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER TABLE ONLY public.t1 ADD CONSTRAINT t1_pkey PRIMARY KEY (id);",
				},
				{
					Query:    "select uid is not null from t1 where id = 1;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "select uid is not null from t1 where id = 2;",
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "select (select uid from t1 where id = 2) = (select uid from t1 where id = 1);",
					Skip:     true, // panic in equality function
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			Name: "alter table drop primary key",
			SetUpScript: []string{
				"CREATE TABLE t1 (id int PRIMARY KEY);",
				"INSERT INTO t1 (id) VALUES (1), (2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER TABLE t1 DROP CONSTRAINT t1_pkey;",
					Expected: []sql.Row{},
				},
				{
					// Assert that the constraint is gone
					Query:    "INSERT INTO t1 VALUES (1), (2);",
					Expected: []sql.Row{},
				},
			},
		},
	})
}
