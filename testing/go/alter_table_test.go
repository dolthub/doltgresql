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
		//{
		//	Name: "Add Foreign Key Constraint",
		//	SetUpScript: []string{
		//		"create table child (pk int primary key, c1 int);",
		//		"insert into child values (1,1), (2,2), (3,3);",
		//		"create index idx_child_c1 on child (pk, c1);",
		//		"create table parent (pk int primary key, c1 int, c2 int);",
		//		"insert into parent values (1, 1, 10);",
		//	},
		//	Assertions: []ScriptTestAssertion{
		//		{
		//			Query:    "ALTER TABLE parent ADD FOREIGN KEY (c1) REFERENCES child (pk) ON DELETE CASCADE;",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			// Test that the FK constraint is working
		//			Query:       "INSERT INTO parent VALUES (10, 10, 10);",
		//			ExpectedErr: "Foreign key violation",
		//		},
		//		{
		//			Query:       "ALTER TABLE parent ADD FOREIGN KEY (c2) REFERENCES child (pk);",
		//			ExpectedErr: "Foreign key violation",
		//		},
		//		{
		//			// Test an FK reference over multiple columns
		//			Query:       "ALTER TABLE parent ADD FOREIGN KEY (c1, c2) REFERENCES child (pk, c1);",
		//			ExpectedErr: "Foreign key violation",
		//		},
		//		{
		//			// Unsupported syntax: MATCH PARTIAL
		//			Query:       "ALTER TABLE parent ADD FOREIGN KEY (c1, c2) REFERENCES child (pk, c1) MATCH PARTIAL;",
		//			ExpectedErr: "MATCH PARTIAL is not yet supported",
		//		},
		//	},
		//},
		//{
		//	Name: "Add Unique Constraint",
		//	SetUpScript: []string{
		//		"create table t1 (pk int primary key, c1 int);",
		//		"insert into t1 values (1,1);",
		//		"create table t2 (pk int primary key, c1 int);",
		//		"insert into t2 values (1,1);",
		//	},
		//	Assertions: []ScriptTestAssertion{
		//		{
		//			// Add a secondary unique index using create index
		//			Query:    "CREATE UNIQUE INDEX ON t1(c1);",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			// Test that the unique constraint is working
		//			Query:       "INSERT INTO t1 VALUES (2, 1);",
		//			ExpectedErr: "unique",
		//		},
		//		{
		//			// Add a secondary unique index using alter table
		//			Query:    "ALTER TABLE t2 ADD CONSTRAINT uniq1 UNIQUE (c1);",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			// Test that the unique constraint is working
		//			Query:       "INSERT INTO t2 VALUES (2, 1);",
		//			ExpectedErr: "unique",
		//		},
		//	},
		//},
		//{
		//	Name: "Add Check Constraint",
		//	SetUpScript: []string{
		//		"create table t1 (pk int primary key, c1 int);",
		//		"insert into t1 values (1,1);",
		//	},
		//	Assertions: []ScriptTestAssertion{
		//		{
		//			// Add a check constraint that is already violated by the existing data
		//			Query:       "ALTER TABLE t1 ADD CONSTRAINT constraint1 CHECK (c1 > 100);",
		//			ExpectedErr: "violated",
		//		},
		//		{
		//			// Add a check constraint
		//			Query:    "ALTER TABLE t1 ADD CONSTRAINT constraint1 CHECK (c1 < 100);",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "INSERT INTO t1 VALUES (2, 2);",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:       "INSERT INTO t1 VALUES (3, 101);",
		//			ExpectedErr: "violated",
		//		},
		//	},
		//},
		//{
		//	Name: "Drop Constraint",
		//	SetUpScript: []string{
		//		"create table t1 (pk int primary key, c1 int);",
		//		"ALTER TABLE t1 ADD CONSTRAINT constraint1 CHECK (c1 > 100);",
		//	},
		//	Assertions: []ScriptTestAssertion{
		//		{
		//			Query:    "ALTER TABLE t1 DROP CONSTRAINT constraint1;",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "INSERT INTO t1 VALUES (1, 1);",
		//			Expected: []sql.Row{},
		//		},
		//	},
		//},
		//{
		//	Name: "Add Primary Key",
		//	SetUpScript: []string{
		//		"CREATE TABLE test1 (a INT, b INT);",
		//		"CREATE TABLE test2 (a INT, b INT, c INT);",
		//		"CREATE TABLE pkTable1 (a INT PRIMARY KEY);",
		//		"CREATE TABLE duplicateRows (a INT, b INT);",
		//		"INSERT INTO duplicateRows VALUES (1, 2), (1, 2);",
		//	},
		//	Assertions: []ScriptTestAssertion{
		//		{
		//			Query:    "ALTER TABLE test1 ADD PRIMARY KEY (a);",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			// Test the pk by inserting a duplicate value
		//			Query:       "INSERT into test1 values (1, 2), (1, 3);",
		//			ExpectedErr: "duplicate primary key",
		//		},
		//		{
		//			Query:    "ALTER TABLE test2 ADD PRIMARY KEY (a, b);",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			// Test the pk by inserting a duplicate value
		//			Query:       "INSERT into test2 values (1, 2, 3), (1, 2, 4);",
		//			ExpectedErr: "duplicate primary key",
		//		},
		//		{
		//			Query:       "ALTER TABLE pkTable1 ADD PRIMARY KEY (a);",
		//			ExpectedErr: "Multiple primary keys defined",
		//		},
		//		{
		//			Query:       "ALTER TABLE duplicateRows ADD PRIMARY KEY (a);",
		//			ExpectedErr: "duplicate primary key",
		//		},
		//		{
		//			// TODO: This statement fails in analysis, because it can't find a table named
		//			//       doesNotExist – since IF EXISTS is specified, the analyzer should skip
		//			//       errors on resolving the table in this case.
		//			Skip:     true,
		//			Query:    "ALTER TABLE IF EXISTS doesNotExist ADD PRIMARY KEY (a, b);",
		//			Expected: []sql.Row{},
		//		},
		//	},
		//},
		//{
		//	Name: "Add Column",
		//	SetUpScript: []string{
		//		"CREATE TABLE test1 (a INT, b INT);",
		//		"INSERT INTO test1 VALUES (1, 1);",
		//	},
		//	Assertions: []ScriptTestAssertion{
		//		{
		//			Query:    "ALTER TABLE test1 ADD COLUMN c INT NOT NULL DEFAULT 42;",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "select * from test1;",
		//			Expected: []sql.Row{{1, 1, 42}},
		//		},
		//	},
		//},
		//{
		//	Name: "Drop Column",
		//	SetUpScript: []string{
		//		"CREATE TABLE test1 (a INT, b INT, c INT, d INT);",
		//		"INSERT INTO test1 VALUES (1, 2, 3, 4);",
		//	},
		//	Assertions: []ScriptTestAssertion{
		//		{
		//			Query:    "ALTER TABLE test1 DROP COLUMN c;",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "select * from test1;",
		//			Expected: []sql.Row{{1, 2, 4}},
		//		},
		//		{
		//			Query:    "ALTER TABLE test1 DROP COLUMN d;",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "select * from test1;",
		//			Expected: []sql.Row{{1, 2}},
		//		},
		//		{
		//			// TODO: Skipped until we support conditional execution on existence of column
		//			Skip:     true,
		//			Query:    "ALTER TABLE test1 DROP COLUMN IF EXISTS zzz;",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			// TODO: Even though we're setting IF EXISTS, this query still fails with an
		//			//       error about the table not existing.
		//			Skip:     true,
		//			Query:    "ALTER TABLE IF EXISTS doesNotExist DROP COLUMN z;",
		//			Expected: []sql.Row{},
		//		},
		//	},
		//},
		//{
		//	Name: "Rename Column",
		//	SetUpScript: []string{
		//		"CREATE TABLE test1 (a INT, b INT, c INT, d INT);",
		//		"INSERT INTO test1 VALUES (1, 2, 3, 4);",
		//	},
		//	Assertions: []ScriptTestAssertion{
		//		{
		//			Query:    "ALTER TABLE test1 RENAME COLUMN c to jjj;",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "select * from test1 where jjj=3;",
		//			Expected: []sql.Row{{1, 2, 3, 4}},
		//		},
		//	},
		//},
		//{
		//	Name: "Set Column Default",
		//	SetUpScript: []string{
		//		"CREATE TABLE test1 (a INT, b INT DEFAULT 42, c INT);",
		//	},
		//	Assertions: []ScriptTestAssertion{
		//		{
		//			Query:    "ALTER TABLE test1 ALTER COLUMN c SET DEFAULT 43;",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "INSERT INTO test1 (a) VALUES (1);",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "SELECT * FROM test1;",
		//			Expected: []sql.Row{{1, 42, 43}},
		//		},
		//		{
		//			Query:    "ALTER TABLE test1 ALTER COLUMN b DROP DEFAULT;",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "INSERT INTO test1 (a) VALUES (2);",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "SELECT * FROM test1 where a = 2;",
		//			Expected: []sql.Row{{2, nil, 43}},
		//		},
		//		{
		//			Query:    "ALTER TABLE test1 ALTER COLUMN c SET DEFAULT length('hello world');",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "INSERT INTO test1 (a) VALUES (3);",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "SELECT * FROM test1 where a = 3;",
		//			Expected: []sql.Row{{3, nil, 11}},
		//		},
		//	},
		//},
		//{
		//	Name: "Set Column Nullability",
		//	SetUpScript: []string{
		//		"CREATE TABLE test1 (a INT, b INT);",
		//	},
		//	Assertions: []ScriptTestAssertion{
		//		{
		//			Query:    "ALTER TABLE test1 ALTER COLUMN b SET NOT NULL;",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:       "INSERT INTO test1 VALUES (1, NULL);",
		//			ExpectedErr: "column name 'b' is non-nullable",
		//		},
		//		{
		//			Query:    "ALTER TABLE test1 ALTER COLUMN b DROP NOT NULL;",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "INSERT INTO test1 VALUES (2, NULL);",
		//			Expected: []sql.Row{},
		//		},
		//		{
		//			Query:    "SELECT * FROM test1 where a = 2;",
		//			Expected: []sql.Row{{2, nil}},
		//		},
		//		{
		//			Query:       "ALTER TABLE test1 ALTER COLUMN b SET NOT NULL;",
		//			ExpectedErr: "'b' is non-nullable but attempted to set a value of null",
		//		},
		//	},
		//},
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
	})
}
