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
	"math/big"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestForeignKeys(t *testing.T) {
	RunScripts(
		t,
		[]ScriptTest{
			{
				Name: "simple foreign key",
				SetUpScript: []string{
					`CREATE TABLE parent (a INT PRIMARY KEY, b int)`,
					`CREATE TABLE child (a INT PRIMARY KEY, b INT, FOREIGN KEY (b) REFERENCES parent(a))`,
					`INSERT INTO parent VALUES (1, 1)`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "INSERT INTO child VALUES (1, 1)",
					},
					{
						Query: "INSERT INTO child VALUES (2, 1)",
					},
					{
						Query:       "INSERT INTO child VALUES (2, 2)",
						ExpectedErr: "Foreign key violation",
					},
				},
			},
			{
				Name: "named constraint",
				SetUpScript: []string{
					`CREATE TABLE parent (a INT PRIMARY KEY, b int)`,
					`CREATE TABLE child (a INT PRIMARY KEY, b INT)`,
					`INSERT INTO parent VALUES (1, 1)`,
					`ALTER TABLE child ADD CONSTRAINT fk123 FOREIGN KEY (b) REFERENCES parent(a)`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "INSERT INTO child VALUES (1, 1)",
					},
					{
						Query: "INSERT INTO child VALUES (2, 1)",
					},
					{
						Query:       "INSERT INTO child VALUES (2, 2)",
						ExpectedErr: "fk123",
					},
				},
			},
			{
				Name: "unnamed constraint",
				SetUpScript: []string{
					`CREATE TABLE parent (a INT PRIMARY KEY, b int)`,
					`CREATE TABLE child (a INT PRIMARY KEY, b INT)`,
					`INSERT INTO parent VALUES (1, 1)`,
					`ALTER TABLE child ADD FOREIGN KEY (b) REFERENCES parent(a)`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "INSERT INTO child VALUES (1, 1)",
					},
					{
						Query: "INSERT INTO child VALUES (2, 1)",
					},
					{
						Query:       "INSERT INTO child VALUES (2, 2)",
						ExpectedErr: "child_b_fkey",
					},
				},
			},
			{
				Name: "text foreign key",
				SetUpScript: []string{
					`CREATE TABLE parent (a text PRIMARY KEY, b int)`,
					`CREATE TABLE child (a INT PRIMARY KEY, b text, FOREIGN KEY (b) REFERENCES parent(a))`,
					`INSERT INTO parent VALUES ('a', 1)`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "INSERT INTO child VALUES (1, 'a')",
					},
					{
						Query: "INSERT INTO child VALUES (2, 'a')",
					},
					{
						Query:       "INSERT INTO child VALUES (3, 'b')",
						ExpectedErr: "Foreign key violation",
					},
				},
			},
			{
				Name: "type compatibility",
				SetUpScript: []string{
					`create table parent (i2 int2, i4 int4, i8 int8, f float, d double precision, v varchar, vl varchar(100), t text, j json, ts timestamp);`,
					"alter table parent add constraint u1 unique (i2);",
					"alter table parent add constraint u2 unique (i4);",
					"alter table parent add constraint u3 unique (i8);",
					"alter table parent add constraint u4 unique (d);",
					"alter table parent add constraint u5 unique (f);",
					"alter table parent add constraint u6 unique (v);",
					"alter table parent add constraint u7 unique (vl);",
					"alter table parent add constraint u8 unique (t);",
					"alter table parent add constraint u9 unique (ts);",
					`create table child (i2 int2, i4 int4, i8 int8, f float, d double precision, v varchar, vl varchar(100), t text, j json, ts timestamp);`,
					"insert into parent values (1, 1, 1, 1.0, 1.0, 'a', 'a', 'a', '{\"a\": 1}', '2021-01-01 00:00:00');",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "alter table child add constraint fi2i2 foreign key (i2) references parent(i2)",
					},
					{
						Query: "alter table child add constraint fi2i4 foreign key (i2) references parent(i4)",
					},
					{
						Query: "alter table child add constraint fi2i8 foreign key (i2) references parent(i8);",
					},
					{
						Query: "alter table child add constraint fi2f foreign key (i2) references parent(f);",
					},
					{
						Query: "alter table child add constraint fi2d foreign key (i2) references parent(d);",
					},
					{
						Query:       "alter table child add constraint fi2v foreign key (i2) references parent(v);",
						ExpectedErr: "incompatible types",
					},
					{
						Query:       "alter table child add constraint fi2vl foreign key (i2) references parent(vl);",
						ExpectedErr: "incompatible types",
					},
					{
						Query:       "alter table child add constraint fi2t foreign key (i2) references parent(t);",
						ExpectedErr: "incompatible types",
					},
					{
						Query:       "alter table child add constraint fi2ts foreign key (i2) references parent(ts);",
						ExpectedErr: "incompatible types",
					},
					{
						Query: "alter table child add constraint fi4i2 foreign key (i4) references parent(i2);",
					},
					{
						Query: "alter table child add constraint fi4i4 foreign key (i4) references parent(i4);",
					},
					{
						Query: "alter table child add constraint fi4i8 foreign key (i4) references parent(i8);",
					},
					{
						Query: "alter table child add constraint fi4f foreign key (i4) references parent(f);",
					},
					{
						Query: "alter table child add constraint fi8i2 foreign key (i8) references parent(i2);",
					},
					{
						Query: "alter table child add constraint fi8i4 foreign key (i8) references parent(i4);",
					},
					{
						Query: "alter table child add constraint fi8d foreign key (i8) references parent(d);",
					},
					{
						Query:       "alter table child add constraint fi8t foreign key (i8) references parent(t);",
						ExpectedErr: "incompatible types",
					},
					{
						Skip:        true, // this isn't allowed in postgres, but works with our constraints currently
						Query:       "alter table child add constraint ffi2 foreign key (f) references parent(i2);",
						ExpectedErr: "incompatible types",
					},
					{
						Skip:        true, // this isn't allowed in postgres, but works with our constraints currently
						Query:       "alter table child add constraint ffi4 foreign key (f) references parent(i4);",
						ExpectedErr: "incompatible types",
					},
					{
						Skip:        true, // this isn't allowed in postgres, but works with our constraints currently
						Query:       "alter table child add constraint ffi8 foreign key (f) references parent(i8);",
						ExpectedErr: "incompatible types",
					},
					{
						Query: "alter table child add constraint ffd foreign key (f) references parent(d);",
					},
					{
						Query: "alter table child add constraint fdf foreign key (d) references parent(f);",
					},
					{
						Query:       "alter table child add constraint fft foreign key (f) references parent(t);",
						ExpectedErr: "incompatible types",
					},
					{
						Query:       "alter table child add constraint ffv foreign key (f) references parent(v);",
						ExpectedErr: "incompatible types",
					},
					{
						Query: "alter table child add constraint fvv foreign key (v) references parent(v);",
					},
					{
						Query: "alter table child add constraint fvvl foreign key (v) references parent(vl);",
					},
					{
						Query:       "alter table child add constraint fvi8 foreign key (v) references parent(i8);",
						ExpectedErr: "incompatible types",
					},
					{
						Query:       "alter table child add constraint fvf foreign key (v) references parent(f);",
						ExpectedErr: "incompatible types",
					},
					{
						Query:       "alter table child add constraint fvts foreign key (v) references parent(ts);",
						ExpectedErr: "incompatible types",
					},
					{
						Query:       "alter table child add constraint fvj foreign key (v) references parent(j);",
						ExpectedErr: "incompatible types",
					},
					{
						Query: "alter table child add constraint fvt foreign key (v) references parent(t);",
					},
					{
						Query: "alter table child add constraint fvllv foreign key (vl) references parent(vl);",
					},
					{
						Query: "alter table child add constraint fvlv foreign key (vl) references parent(v);",
					},
					{
						Query: "alter table child add constraint fvlt foreign key (vl) references parent(t);",
					},
					{
						Query: "alter table child add constraint ftt foreign key (t) references parent(t);",
					},
					{
						Query: "alter table child add constraint ftv foreign key (t) references parent(v);",
					},
					{
						Query: "alter table child add constraint ftvl foreign key (t) references parent(vl);",
					},
					{
						Query:       "alter table child add constraint fti8 foreign key (t) references parent(i8);",
						ExpectedErr: "incompatible types",
					},
					{
						Query: "alter table child add constraint ftsts foreign key (ts) references parent(ts);",
					},
					{
						Query:       "alter table child add constraint ftst foreign key (ts) references parent(t);",
						ExpectedErr: "incompatible types",
					},
					{
						Query:       "alter table child add constraint ftsi8 foreign key (ts) references parent(i8);",
						ExpectedErr: "incompatible types",
					},
					{
						Query: "insert into child values (1, 1, 1, 1.0, 1.0, 'a', 'a', 'a', '{\"a\": 1}', '2021-01-01 00:00:00');",
					},
					{
						Query:       "insert into child values (1, 2, 1, 1.0, 1.0, 'a', 'a', 'a', '{\"a\": 1}', '2021-01-01 00:00:00');",
						ExpectedErr: "Foreign key",
					},
					{
						Query:       "insert into child values (1, 1, 1, 2.0, 1.0, 'a', 'a', 'a', '{\"a\": 1}', '2021-01-01 00:00:00');",
						ExpectedErr: "Foreign key",
					},
					{
						Query:       "insert into child values (1, 1, 1, 1.0, 1.0, 'a', 'a', 'b', '{\"a\": 1}', '2021-01-01 00:00:00');",
						ExpectedErr: "Foreign key",
					},
					{
						Query:       "insert into child values (1, 1, 1, 1.0, 1.0, 'a', 'a', 'a', '{\"a\": 1}', '2021-01-01 00:00:01');",
						ExpectedErr: "Foreign key",
					},
				},
			},
			{
				Name: "type conversion: text to varchar",
				SetUpScript: []string{
					`CREATE TABLE parent (a INT PRIMARY KEY, b varchar(100))`,
					`CREATE TABLE child (c INT PRIMARY KEY, d text)`,
					`INSERT INTO parent VALUES (1, 'abc'), (2, 'def')`,
					`alter table parent add constraint ub unique (b)`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "alter table child add constraint fk foreign key (d) references parent(b)",
					},
					{
						Query: "insert into child values (1, 'abc')",
					},
					{
						Query:       "insert into child values (2, 'xyz')",
						ExpectedErr: "Foreign key",
					},
					{
						Query: "delete from parent where b = 'def'",
					},
					{
						Query:       "delete from parent where b = 'abc'",
						ExpectedErr: "Foreign key",
					},
				},
			},
			{
				Name: "type conversion: integer to double",
				SetUpScript: []string{
					`CREATE TABLE parent (a INT PRIMARY KEY, b double precision)`,
					`CREATE TABLE child (c INT PRIMARY KEY, d int)`,
					`INSERT INTO parent VALUES (1, 1), (3, 3)`,
					`alter table parent add constraint ub unique (b)`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "alter table child add constraint fk foreign key (d) references parent(b)",
					},
					{
						Query: "select * from parent where b = 1.0",
						Expected: []sql.Row{
							{1, 1.0},
						},
					},
					{
						Query: "insert into child values (1, 1)",
					},
					{
						Query: "insert into child values (2, 1)",
					},
					{
						Query:       "insert into child values (2, 2)",
						ExpectedErr: "Foreign key",
					},
					{
						Query: "delete from parent where b = 3.0",
					},
					{
						Query:       "delete from parent where b = 1.0",
						ExpectedErr: "Foreign key",
					},
				},
			},
			{
				Name: "type conversion: value out of bounds, child larger",
				SetUpScript: []string{
					`CREATE TABLE parent (a INT PRIMARY KEY, b int2)`,
					`CREATE TABLE child (c INT PRIMARY KEY, d int8)`,
					`INSERT INTO parent VALUES (1, 1), (3, 3)`,
					`alter table parent add constraint ub unique (b)`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "alter table child add constraint fk foreign key (d) references parent(b)",
					},
					{
						Query: "insert into child values (1, 1)",
					},
					{
						Query:       "insert into child values (2, 2)",
						ExpectedErr: "Foreign key",
					},
					{
						Query:       "insert into child values (2, 65536)", // above maximum int2
						ExpectedErr: "Foreign key",
					},
					{
						Query: "delete from parent where b = 3",
					},
					{
						Query:       "delete from parent where b = 1",
						ExpectedErr: "Foreign key",
					},
				},
			},
			{
				Name: "type conversion: value out of bound, parent larger",
				SetUpScript: []string{
					`CREATE TABLE parent (a INT PRIMARY KEY, b int8)`,
					`CREATE TABLE child (c INT PRIMARY KEY, d int2)`,
					`INSERT INTO parent VALUES (1, 1), (65536, 65536)`, // above maximum int2
					`alter table parent add constraint ub unique (b)`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "alter table child add constraint fk foreign key (d) references parent(b)",
					},
					{
						Query: "insert into child values (1, 1)",
					},
					{
						Query:       "insert into child values (2, 2)",
						ExpectedErr: "Foreign key",
					},
					{
						Query:       "insert into child values (2, 65536)",
						ExpectedErr: "out of range",
					},
					{
						Query: "delete from parent where b = 65536",
					},
					{
						Query:       "delete from parent where b = 1",
						ExpectedErr: "Foreign key",
					},
				},
			},
			{
				Name: "foreign key with dolt_add, dolt_commit",
				SetUpScript: []string{
					"create table test (pk int, \"value\" int, primary key(pk));",
					"CREATE TABLE test_info (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references test(pk))",
					"INSERT INTO test VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_ADD('.')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT * FROM dolt.status",
						Expected: []sql.Row{
							{"public.test", "t", "new table"},
							{"public.test_info", "t", "new table"},
						},
					},
					{
						Query:            "SELECT dolt_commit('-am', 'new tables')",
						SkipResultsCheck: true,
					},
					{
						Query:    "SELECT * FROM dolt.status",
						Expected: []sql.Row{},
					},
					{
						Query:    "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'test_info')",
						Expected: []sql.Row{},
					},
					{
						Query:    "INSERT INTO test_info VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:       "INSERT INTO test_info VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM test_info",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "foreign key with explicit schema",
				SetUpScript: []string{
					"create table parent (pk int, \"value\" int, primary key(pk));",
					"CREATE TABLE child (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references public.parent(pk))",
					"INSERT INTO parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_ADD('.')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT * FROM dolt.status",
						Expected: []sql.Row{
							{"public.child", "t", "new table"},
							{"public.parent", "t", "new table"},
						},
					},
					{
						Query:            "SELECT dolt_commit('-am', 'new tables')",
						SkipResultsCheck: true,
					},
					{
						Query:    "SELECT * FROM dolt.status",
						Expected: []sql.Row{},
					},
					{
						Query:    "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'child')",
						Expected: []sql.Row{},
					},
					{
						Query:    "INSERT INTO child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:       "INSERT INTO child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "foreign key in another schema with search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"select dolt_commit('-Am', 'create schemas')",
					"set search_path to 'parent, child'",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references parent(pk))",
					"INSERT INTO parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_ADD('.')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT * FROM dolt_status",
						Expected: []sql.Row{
							{"child.child", "t", "new table"},
							{"fake.parent", "t", "new table"},
							{"parent.parent", "t", "new table"},
						},
					},
					{
						Query:            "SELECT dolt_commit('-am', 'new tables')",
						SkipResultsCheck: true,
					},
					{
						Query:    "SELECT * FROM dolt_status",
						Expected: []sql.Row{},
					},
					{
						Query:    "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'child')",
						Expected: []sql.Row{},
					},
					{
						Query:    "INSERT INTO child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:       "INSERT INTO child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child.child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "foreign key in another schema with search path, parent table not on search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"select dolt_commit('-Am', 'create schemas')",
					"set search_path to 'child, fake'",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references parent.parent(pk))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_ADD('.')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT * FROM dolt.status",
						Expected: []sql.Row{
							{"child.child", "t", "new table"},
							{"fake.parent", "t", "new table"},
							{"parent.parent", "t", "new table"},
						},
					},
					{
						Query:            "SELECT dolt_commit('-am', 'new tables')",
						SkipResultsCheck: true,
					},
					{
						Query:    "SELECT * FROM dolt_status",
						Expected: []sql.Row{},
					},
					{
						Query:    "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'child')",
						Expected: []sql.Row{},
					},
					{
						Query:    "INSERT INTO child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:       "INSERT INTO child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child.child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "foreign key in another schema, no search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"select dolt_commit('-Am', 'create schemas')",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id), foreign key (test_pk) references parent.parent(pk))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_ADD('.')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT * FROM dolt_status",
						Expected: []sql.Row{
							{"child.child", "t", "new table"},
							{"fake.parent", "t", "new table"},
							{"parent.parent", "t", "new table"},
						},
					},
					{
						Query:            "SELECT dolt_commit('-am', 'new tables')",
						SkipResultsCheck: true,
					},
					{
						Query:    "SELECT * FROM dolt.status",
						Expected: []sql.Row{},
					},
					{
						Query:    "SELECT * FROM dolt_schema_diff('HEAD', 'WORKING', 'child')",
						Expected: []sql.Row{},
					},
					{
						Query:    "INSERT INTO child.child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:       "INSERT INTO child.child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child.child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "add foreign key in another schema on search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"select dolt_commit('-Am', 'create schemas')",
					"set search_path to 'child, parent'",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_COMMIT('-Am', 'new tables')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "INSERT INTO child.child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:            "ALTER TABLE child ADD FOREIGN KEY (test_pk) REFERENCES parent(pk)",
						SkipResultsCheck: true,
					},
					{
						Query:       "INSERT INTO child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "add foreign key in another schema, parent table not on search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"select dolt_commit('-Am', 'create schemas')",
					"set search_path to 'child, fake'",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_COMMIT('-Am', 'new tables')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "INSERT INTO child.child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:            "ALTER TABLE child ADD FOREIGN KEY (test_pk) REFERENCES parent.parent(pk)",
						SkipResultsCheck: true,
					},
					{
						Query:       "INSERT INTO child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "add foreign key in another schema, no search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"select dolt_commit('-Am', 'create schemas')",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_COMMIT('-Am', 'new tables')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "INSERT INTO child.child VALUES (2, 'two', 2)",
						Expected: []sql.Row{},
					},
					{
						Query:            "ALTER TABLE child.child ADD FOREIGN KEY (test_pk) REFERENCES parent.parent(pk)",
						SkipResultsCheck: true,
					},
					{
						Query:       "INSERT INTO child.child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query: "SELECT * FROM child.child",
						Expected: []sql.Row{
							{2, "two", 2},
						},
					},
				},
			},
			{
				Name: "drop foreign key in another schema, on search path",
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"select dolt_commit('-Am', 'create schemas')",
					"set search_path to 'child, parent'",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_COMMIT('-Am', 'new tables')",
					"INSERT INTO child.child VALUES (2, 'two', 2)",
					"ALTER TABLE child.child ADD CONSTRAINT fk1 FOREIGN KEY (test_pk) REFERENCES parent(pk)",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:       "INSERT INTO child.child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query:            "alter table child DROP constraint fk1;",
						SkipResultsCheck: true,
					},
					{
						Query:    "INSERT INTO child.child VALUES (3, 'three', 3)",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "drop foreign key in another schema, no search path",
				Skip: true, // not getting the explicit schema name passed to the node
				SetUpScript: []string{
					"create schema parent",
					"create schema child",
					"create schema fake",
					"select dolt_commit('-Am', 'create schemas')",
					`create table parent.parent (pk int, val int, primary key(pk));`,
					`create table fake.parent (pk int, val int, primary key(pk));`,
					"CREATE TABLE child.child (id int, info varchar(255), test_pk int, primary key(id))",
					"INSERT INTO parent.parent VALUES (0, 0), (1, 1), (2,2)",
					"SELECT DOLT_COMMIT('-Am', 'new tables')",
					"INSERT INTO child.child VALUES (2, 'two', 2)",
					"ALTER TABLE child.child ADD FOREIGN KEY (test_pk) REFERENCES parent.parent(pk)",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:       "INSERT INTO child.child VALUES (3, 'three', 3)",
						ExpectedErr: "Foreign key violation",
					},
					{
						Query:            "alter table child.child DROP constraint child_ibfk_1",
						SkipResultsCheck: true,
					},
					{
						Query:    "INSERT INTO child.child VALUES (3, 'three', 3)",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "foreign key default naming",
				SetUpScript: []string{
					"CREATE TABLE webhooks (id varchar not null, id2 int8, primary key (id));",
					"CREATE UNIQUE INDEX idx1 on webhooks(id, id2);",
					"CREATE TABLE t33 (id varchar not null, webhook_id_fk varchar not null, webhook_id2_fk int8, foreign key (webhook_id_fk) references webhooks(id), foreign key (webhook_id_fk, webhook_id2_fk) references webhooks(id, id2), primary key (id));",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass  AND contype = 'f';",
						Expected: []sql.Row{
							{"t33_webhook_id_fk_fkey"},
							{"t33_webhook_id_fk_webhook_id2_fk_fkey"},
						},
					},
					{
						Query:    "ALTER TABLE t33 DROP CONSTRAINT t33_webhook_id_fk_fkey;",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "foreign key default naming, name collision ",
				SetUpScript: []string{
					"CREATE TABLE parent (id varchar not null primary key);",
					"CREATE TABLE child (id varchar primary key, constraint t33_webhook_id_fk_fkey foreign key (id) references parent(id));",
					"CREATE TABLE webhooks (id varchar not null, id2 int8, primary key (id));",
					"CREATE TABLE t33 (id varchar not null, webhook_id_fk varchar not null, foreign key (webhook_id_fk) references webhooks(id), primary key (id));",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass  AND contype = 'f';",
						Expected: []sql.Row{
							{"t33_webhook_id_fk_fkey1"},
						},
					},
					{
						Query:    "ALTER TABLE t33 DROP CONSTRAINT t33_webhook_id_fk_fkey1;",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "foreign key default naming, in column definition",
				SetUpScript: []string{
					"CREATE TABLE webhooks (id varchar not null, primary key (id));",
					"CREATE TABLE t33 (id varchar not null, webhook_id_fk varchar not null references webhooks(id), primary key (id));",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass  AND contype = 'f';",
						Expected: []sql.Row{
							{"t33_webhook_id_fk_fkey"},
						},
					},
					{
						Query:    "ALTER TABLE t33 DROP CONSTRAINT t33_webhook_id_fk_fkey;",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "foreign key custom naming",
				SetUpScript: []string{
					"CREATE TABLE webhooks (id VARCHAR NOT NULL, PRIMARY KEY (id));",
					"CREATE TABLE t33 (id VARCHAR NOT NULL, webhook_id_fk VARCHAR NOT NULL, CONSTRAINT foo1 FOREIGN KEY (webhook_id_fk) REFERENCES webhooks(id), PRIMARY KEY (id));",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass AND contype = 'f';",
						Expected: []sql.Row{{"foo1"}},
					},
					{
						Query:    "ALTER TABLE t33 DROP CONSTRAINT foo1;",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "foreign key default naming, added through alter table",
				SetUpScript: []string{
					"CREATE TABLE webhooks (id varchar not null, primary key (id));",
					"CREATE TABLE t33 (id varchar not null, webhook_id_fk varchar not null, primary key (id));",
					"ALTER TABLE t33 ADD FOREIGN KEY (webhook_id_fk) REFERENCES webhooks(id);",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "SELECT conname AS constraint_name FROM pg_constraint WHERE conrelid = 't33'::regclass  AND contype = 'f';",
						Expected: []sql.Row{
							{"t33_webhook_id_fk_fkey"},
						},
					},
					{
						Query:    "ALTER TABLE t33 DROP CONSTRAINT t33_webhook_id_fk_fkey;",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "ON DELETE ... SET DEFAULT",
				SetUpScript: []string{
					"CREATE TABLE public.hn_stories (title text NOT NULL, website_url text);",
					"CREATE TABLE public.websites (url text primary key, title text);",
					"INSERT into public.websites VALUES ('http://www.dolthub.com', 'foo1'), ('http://www.google.com', 'foo2');",
					"INSERT into public.hn_stories VALUES ('test1', 'http://www.dolthub.com'), ('test2', 'http://www.google.com');",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: `ALTER TABLE ONLY public.hn_stories
				ADD CONSTRAINT hn_stories_website_url_fkey FOREIGN KEY (website_url) REFERENCES public.websites(url) ON UPDATE CASCADE ON DELETE SET DEFAULT;`,
						Expected: []sql.Row{},
					},
					{
						Query: "DELETE FROM public.websites WHERE title = 'foo1';",
					},
					{
						Query:    "SELECT * FROM public.hn_stories where title = 'test1';",
						Expected: []sql.Row{{"test1", nil}},
					},
					{
						Query:    "ALTER TABLE hn_stories ALTER COLUMN website_url SET DEFAULT (title);",
						Expected: []sql.Row{},
					},
					{
						Query: "DELETE FROM public.websites WHERE title = 'foo2';",
					},
					{
						Query:    "SELECT * FROM public.hn_stories where title = 'test2';",
						Expected: []sql.Row{{"test2", "test2"}},
					},
				},
			},
			{
				Name: "ON UPDATE ... SET DEFAULT",
				SetUpScript: []string{
					"CREATE TABLE public.hn_stories (title text NOT NULL, website_url text);",
					"CREATE TABLE public.websites (url text primary key, title text);",
					"INSERT into public.websites VALUES ('http://www.dolthub.com', 'foo1'), ('http://www.google.com', 'foo2');",
					"INSERT into public.hn_stories VALUES ('test1', 'http://www.dolthub.com'), ('test2', 'http://www.google.com');",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: `ALTER TABLE ONLY public.hn_stories
				ADD CONSTRAINT hn_stories_website_url_fkey FOREIGN KEY (website_url) REFERENCES public.websites(url) ON UPDATE SET DEFAULT;`,
						Expected: []sql.Row{},
					},
					{
						Query: "UPDATE public.websites SET url = 'http://fake.com' WHERE title = 'foo1';",
					},
					{
						Query:    "SELECT * FROM public.hn_stories where title = 'test1';",
						Expected: []sql.Row{{"test1", nil}},
					},
					{
						Query:    "ALTER TABLE hn_stories ALTER COLUMN website_url SET DEFAULT (title);",
						Expected: []sql.Row{},
					},
					{
						Query: "UPDATE public.websites SET url = 'http://doltdb.com' WHERE title = 'foo2';",
					},
					{
						Query:    "SELECT * FROM public.hn_stories where title = 'test2';",
						Expected: []sql.Row{{"test2", "test2"}},
					},
				},
			},
			{
				Name: "merging",
				SetUpScript: []string{
					`CREATE TABLE "evaluation_job_config" (
	"tenant_id" varchar(256) NOT NULL,
	"id" varchar(256) NOT NULL,
	"project_id" varchar(256) NOT NULL,
	"job_filters" jsonb,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "evaluation_job_config_tenant_id_project_id_id_pk" PRIMARY KEY("tenant_id","project_id","id")
);`,
					`CREATE TABLE "evaluation_job_config_evaluator_relations" (
	"tenant_id" varchar(256) NOT NULL,
	"id" varchar(256) NOT NULL,
	"project_id" varchar(256) NOT NULL,
	"evaluation_job_config_id" text NOT NULL,
	"evaluator_id" text NOT NULL,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "eval_job_cfg_evaluator_rel_pk" PRIMARY KEY("tenant_id","project_id","id")
);`,
					`CREATE TABLE "agent" (
	"tenant_id" varchar(256) NOT NULL,
	"id" varchar(256) NOT NULL,
	"project_id" varchar(256) NOT NULL,
	"name" varchar(256) NOT NULL,
	"description" text,
	"default_sub_agent_id" varchar(256),
	"context_config_id" varchar(256),
	"models" jsonb,
	"status_updates" jsonb,
	"prompt" text,
	"stop_when" jsonb,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "agent_tenant_id_project_id_id_pk" PRIMARY KEY("tenant_id","project_id","id")
);`,
					`CREATE TABLE "projects" (
	"tenant_id" varchar(256) NOT NULL,
	"id" varchar(256) NOT NULL,
	"name" varchar(256) NOT NULL,
	"description" text,
	"models" jsonb,
	"stop_when" jsonb,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "projects_tenant_id_id_pk" PRIMARY KEY("tenant_id","id")
);`,
					`ALTER TABLE "evaluation_job_config" ADD CONSTRAINT "evaluation_job_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "public"."projects"("tenant_id","id") ON DELETE cascade ON UPDATE no action;`,
					`ALTER TABLE "evaluation_job_config_evaluator_relations" ADD CONSTRAINT "eval_job_cfg_evaluator_rel_job_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_job_config_id") REFERENCES "public"."evaluation_job_config"("tenant_id","project_id","id") ON DELETE cascade ON UPDATE no action;`,
					`INSERT INTO projects VALUES ('tenant1', 'project1', 'Project One', 'First project', '{"model": "gpt-4"}', '{"condition": "complete"}', now(), now());`,
					`INSERT INTO evaluation_job_config VALUES ('tenant1', 'jobconfig1', 'project1', '{"filter": "all"}', now(), now());`,
					`INSERT INTO evaluation_job_config_evaluator_relations VALUES ('tenant1', 'rel1', 'project1', 'jobconfig1', 'evaluator1', now(), now());`,
					`INSERT INTO agent VALUES ('tenant1', 'agent1', 'project1', 'Agent One', 'First agent', null, null, '{"model": "gpt-4"}', '{}', 'You are an agent.', '{}', now(), now());`,
					`SELECT DOLT_COMMIT('-Am', 'initial tables')`,
					`SELECT DOLT_BRANCH('feature')`,
					`CREATE TABLE "triggers" (
	"tenant_id" varchar(256) NOT NULL,
	"id" varchar(256) NOT NULL,
	"project_id" varchar(256) NOT NULL,
	"agent_id" varchar(256) NOT NULL,
	"name" varchar(256) NOT NULL,
	"description" text,
	"enabled" boolean DEFAULT true NOT NULL,
	"input_schema" jsonb,
	"output_transform" jsonb,
	"message_template" text NOT NULL,
	"authentication" jsonb,
	"signing_secret" text,
	"created_at" timestamp DEFAULT now() NOT NULL,
	"updated_at" timestamp DEFAULT now() NOT NULL,
	CONSTRAINT "triggers_tenant_id_project_id_agent_id_id_pk" PRIMARY KEY("tenant_id","project_id","agent_id","id")
);`,
					`ALTER TABLE "triggers" ADD CONSTRAINT "triggers_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "public"."agent"("tenant_id","project_id","id") ON DELETE cascade ON UPDATE no action;`,
					`select DOLT_COMMIT('-Am', 'add triggers table')`,
					`select dolt_checkout('feature')`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "insert into agent VALUES ('tenant1', 'agent2', 'project1', 'Agent Two', 'Second agent', null, null, '{\"model\": \"gpt-4\"}', '{}', 'You are another agent.', '{}', now(), now());",
					},
					{
						Query:            "select dolt_commit('-Am', 'add second agent')",
						SkipResultsCheck: true,
					},
					{
						Query:    "select strpos(dolt_merge('main')::text, 'merge successful') > 1;",
						Expected: []sql.Row{{"t"}},
					},
				},
			},
			{
				Name: "merge with constraint violations",
				SetUpScript: []string{
					"CREATE TABLE parent (a INT PRIMARY KEY, b INT UNIQUE);",
					"CREATE TABLE child (c INT PRIMARY KEY, d INT);",
					"alter table child add constraint fk foreign key (d) references parent(b);",
					"INSERT INTO parent VALUES (1, 1), (2, 2), (3, 3);",
					"INSERT INTO child VALUES (1, 1), (2, 2);",
					"SELECT DOLT_COMMIT('-Am', 'initial commit')",
					"SELECT DOLT_BRANCH('feature')",
					"insert into child VALUES (3, 3);",
					"SELECT DOLT_COMMIT('-Am', 'new child')",
					"select dolt_checkout('feature')",
					"delete from parent where b = 3;",
					"SELECT DOLT_COMMIT('-Am', 'delete from parent')",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:       "select dolt_merge('main')",
						ExpectedErr: "constraint violations",
					},
					{
						Query: "set dolt_force_transaction_commit = 1;",
					},
					{
						Query:            "select dolt_merge('main')",
						SkipResultsCheck: true,
					},
					{
						Query: "select * from dolt_constraint_violations order by 1",
						Expected: []sql.Row{
							{"child", pgtype.Numeric{Int: big.NewInt(1), Valid: true}},
						},
					},
					{
						Query: "select violation_type, c, d, violation_info from dolt_constraint_violations_child order by 1",
						Expected: []sql.Row{
							{"foreign key", 3, 3, "{\"Columns\":[\"d\"],\"ForeignKey\":\"fk\",\"Index\":\"fk\",\"OnDelete\":\"RESTRICT\",\"OnUpdate\":\"RESTRICT\",\"ReferencedColumns\":[\"b\"],\"ReferencedIndex\":\"b\",\"ReferencedTable\":\"parent\",\"Table\":\"child\"}"},
						},
					},
				},
			},
			{
				Name: "foreign keys in drizzle migration and merge",
				SetUpScript: []string{
					`CREATE SCHEMA "drizzle";`,
					`CREATE SEQUENCE drizzle."__drizzle_migrations_id_seq" AS int4;`,
					`CREATE TABLE "__drizzle_migrations" (
  "id" integer NOT NULL DEFAULT (nextval('drizzle.__drizzle_migrations_id_seq')),
  "hash" text NOT NULL,
  "created_at" bigint,
  PRIMARY KEY ("id")
);`,
					`INSERT INTO "__drizzle_migrations" ("hash","created_at") VALUES ('d3445cf0eaeb405a6b4b9c8386188aece144d40ba89b9616175ca0f69229cc51',1767821157311);`,
					`CREATE TABLE "projects" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "models" jsonb,
  "stop_when" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","id")
);`,
					`CREATE TABLE "agent" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "default_sub_agent_id" varchar(256),
  "context_config_id" varchar(256),
  "models" jsonb,
  "status_updates" jsonb,
  "prompt" text,
  "stop_when" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "agent_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "artifact_components" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "props" jsonb,
  "render" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "artifact_components_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "context_configs" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "headers_schema" jsonb,
  "context_variables" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "context_configs_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "credential_references" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "type" varchar(256) NOT NULL,
  "credential_store_id" varchar(256) NOT NULL,
  "retrieval_params" jsonb,
  "tool_id" varchar(256),
  "user_id" varchar(256),
  "created_by" varchar(256),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "credential_references_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE UNIQUE INDEX "credential_references_id_unique" ON "credential_references" ("id");`,
					`CREATE UNIQUE INDEX "credential_references_tool_user_unique" ON "credential_references" ("tool_id", "user_id");`,
					`CREATE TABLE "data_components" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "props" jsonb,
  "render" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "data_components_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "dataset" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "dataset_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "dataset_item" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "dataset_id" text NOT NULL,
  "input" jsonb NOT NULL,
  "expected_output" jsonb,
  "simulation_agent" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "dataset_item_dataset_fk" FOREIGN KEY ("tenant_id","project_id","dataset_id") REFERENCES "dataset" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "dataset_run_config" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "dataset_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "dataset_run_config_dataset_fk" FOREIGN KEY ("tenant_id","project_id","dataset_id") REFERENCES "dataset" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "dataset_run_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "dataset_run_config_agent_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "dataset_run_config_id" text NOT NULL,
  "agent_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "dataset_run_config_agent_relations_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "dataset_run_config_agent_relations_dataset_run_config_fk" FOREIGN KEY ("tenant_id","project_id","dataset_run_config_id") REFERENCES "dataset_run_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "evaluation_job_config" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "job_filters" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "evaluation_job_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "evaluator" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "prompt" text NOT NULL,
  "schema" jsonb NOT NULL,
  "model" jsonb NOT NULL,
  "pass_criteria" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "evaluator_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "evaluation_job_config_evaluator_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "evaluation_job_config_id" text NOT NULL,
  "evaluator_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "eval_job_cfg_evaluator_rel_evaluator_fk" FOREIGN KEY ("tenant_id","project_id","evaluator_id") REFERENCES "evaluator" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "eval_job_cfg_evaluator_rel_job_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_job_config_id") REFERENCES "evaluation_job_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "evaluation_run_config" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "is_active" boolean NOT NULL DEFAULT 'true',
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "evaluation_run_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "evaluation_suite_config" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "filters" jsonb,
  "sample_rate" double precision,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "evaluation_suite_config_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "evaluation_run_config_evaluation_suite_config_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "evaluation_run_config_id" text NOT NULL,
  "evaluation_suite_config_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "eval_run_cfg_eval_suite_rel_run_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_run_config_id") REFERENCES "evaluation_run_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "eval_run_cfg_eval_suite_rel_suite_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_suite_config_id") REFERENCES "evaluation_suite_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "evaluation_suite_config_evaluator_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "evaluation_suite_config_id" text NOT NULL,
  "evaluator_id" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "eval_suite_cfg_evaluator_rel_evaluator_fk" FOREIGN KEY ("tenant_id","project_id","evaluator_id") REFERENCES "evaluator" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "eval_suite_cfg_evaluator_rel_suite_cfg_fk" FOREIGN KEY ("tenant_id","project_id","evaluation_suite_config_id") REFERENCES "evaluation_suite_config" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "external_agents" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "base_url" text NOT NULL,
  "credential_reference_id" varchar(256),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "external_agents_credential_reference_fk" FOREIGN KEY ("credential_reference_id") REFERENCES "credential_references" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
  CONSTRAINT "external_agents_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "functions" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "input_schema" jsonb,
  "execute_code" text NOT NULL,
  "dependencies" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "functions_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "function_tools" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "function_id" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "function_tools_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "function_tools_function_fk" FOREIGN KEY ("tenant_id","project_id","function_id") REFERENCES "functions" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "sub_agents" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "prompt" text,
  "conversation_history_config" jsonb DEFAULT '{"mode":"full","limit":50,"maxOutputTokens":4000,"includeInternal":false,"messageTypes":["chat","tool-result"]}'::JSONB,
  "models" jsonb,
  "stop_when" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agents_agents_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "tools" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "config" jsonb NOT NULL,
  "credential_reference_id" varchar(256),
  "credential_scope" varchar(50) NOT NULL DEFAULT 'project',
  "headers" jsonb,
  "image_url" text,
  "capabilities" jsonb,
  "last_error" text,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "tools_credential_reference_fk" FOREIGN KEY ("credential_reference_id") REFERENCES "credential_references" ("id") ON DELETE SET NULL ON UPDATE NO ACTION,
  CONSTRAINT "tools_project_fk" FOREIGN KEY ("tenant_id","project_id") REFERENCES "projects" ("tenant_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "sub_agent_artifact_components" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "artifact_component_id" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","sub_agent_id","id"),
  CONSTRAINT "sub_agent_artifact_components_artifact_component_fk" FOREIGN KEY ("tenant_id","project_id","artifact_component_id") REFERENCES "artifact_components" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_artifact_components_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "sub_agent_data_components" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "data_component_id" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","id"),
  CONSTRAINT "sub_agent_data_components_data_component_fk" FOREIGN KEY ("tenant_id","project_id","data_component_id") REFERENCES "data_components" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_data_components_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "sub_agent_external_agent_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "external_agent_id" varchar(256) NOT NULL,
  "headers" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_external_agent_relations_external_agent_fk" FOREIGN KEY ("tenant_id","project_id","external_agent_id") REFERENCES "external_agents" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_external_agent_relations_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "sub_agent_function_tool_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "function_tool_id" varchar(256) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_function_tool_relations_function_tool_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","function_tool_id") REFERENCES "function_tools" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_function_tool_relations_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "sub_agent_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "source_sub_agent_id" varchar(256) NOT NULL,
  "target_sub_agent_id" varchar(256),
  "relation_type" varchar(256),
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_relations_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "sub_agent_team_agent_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "target_agent_id" varchar(256) NOT NULL,
  "headers" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_team_agent_relations_sub_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_team_agent_relations_target_agent_fk" FOREIGN KEY ("tenant_id","project_id","target_agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`CREATE TABLE "sub_agent_tool_relations" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "sub_agent_id" varchar(256) NOT NULL,
  "tool_id" varchar(256) NOT NULL,
  "selected_tools" jsonb,
  "headers" jsonb,
  "tool_policies" jsonb,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "sub_agent_tool_relations_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id","sub_agent_id") REFERENCES "sub_agents" ("tenant_id","project_id","agent_id","id") ON DELETE CASCADE ON UPDATE NO ACTION,
  CONSTRAINT "sub_agent_tool_relations_tool_fk" FOREIGN KEY ("tenant_id","project_id","tool_id") REFERENCES "tools" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`SELECT DOLT_COMMIT('-Am', 'Applied database migrations');`,
					`SELECT DOLT_BRANCH('default_my-weather-project_main');`,
					`INSERT INTO "__drizzle_migrations" ("hash","created_at") VALUES ('634b9140001f10d551fe0d81ca19050f3cc8af8da1ab6c9b6e93d99f33e5fc84',1768766675586);`,
					`CREATE TABLE "triggers" (
  "tenant_id" varchar(256) NOT NULL,
  "id" varchar(256) NOT NULL,
  "project_id" varchar(256) NOT NULL,
  "agent_id" varchar(256) NOT NULL,
  "name" varchar(256) NOT NULL,
  "description" text,
  "enabled" boolean NOT NULL DEFAULT 'true',
  "input_schema" jsonb,
  "output_transform" jsonb,
  "message_template" text NOT NULL,
  "authentication" jsonb,
  "signing_secret" text,
  "created_at" timestamp NOT NULL DEFAULT (now()),
  "updated_at" timestamp NOT NULL DEFAULT (now()),
  PRIMARY KEY ("tenant_id","project_id","agent_id","id"),
  CONSTRAINT "triggers_agent_fk" FOREIGN KEY ("tenant_id","project_id","agent_id") REFERENCES "agent" ("tenant_id","project_id","id") ON DELETE CASCADE ON UPDATE NO ACTION
);`,
					`SELECT DOLT_COMMIT('-Am', 'Applied database migrations');`,
					`SELECT DOLT_CHECKOUT('default_my-weather-project_main');`,
					`INSERT INTO "projects" ("tenant_id","id","name","description","models","stop_when","created_at","updated_at") VALUES ('default','my-weather-project','Weather Project','Project containing sample agent framework using ','{"base": {"model": "openai/gpt-4o-mini"}}',NULL,'2026-01-22 16:19:32.74','2026-01-22 16:19:32.74');`,
					`INSERT INTO "agent" ("tenant_id","id","project_id","name","description","default_sub_agent_id","context_config_id","models","status_updates","prompt","stop_when","created_at","updated_at") VALUES ('default','weather-agent','my-weather-project','Weather agent',NULL,'weather-assistant',NULL,NULL,NULL,NULL,NULL,'2026-01-22 16:19:32.782','2026-01-22 16:19:32.862');`,
					`INSERT INTO "data_components" ("tenant_id","id","project_id","name","description","props","render","created_at","updated_at") VALUES ('default','weather-forecast','my-weather-project','WeatherForecast','A hourly forecast for the weather at a given location','{"type": "object", "required": ["forecast"], "properties": {"forecast": {"type": "array", "items": {"type": "object", "required": ["time", "temperature", "code"], "properties": {"code": {"type": "number", "description": "Weather code at given time"}, "time": {"type": "string", "description": "The time of current item E.g. 12PM, 1PM"}, "temperature": {"type": "number", "description": "The temperature at given time in Farenheit"}}, "additionalProperties": false}, "description": "The hourly forecast for the weather at a given location"}}, "additionalProperties": false}',NULL,'2026-01-22 16:19:32.773665','2026-01-22 16:19:32.773665');`,
					`INSERT INTO "sub_agents" ("tenant_id","id","project_id","agent_id","name","description","prompt","conversation_history_config","models","stop_when","created_at","updated_at") VALUES ('default','geocoder-agent','my-weather-project','weather-agent','Geocoder agent','Specialized agent for converting addresses and location names into geographic coordinates. This agent handles all location-related queries and provides accurate latitude/longitude data for weather lookups.','You are a geocoding specialist that converts addresses, place names, and location descriptions
 into precise geographic coordinates. You help users find the exact location they''re asking about
 and provide the coordinates needed for weather forecasting.

 When users provide:
 - Street addresses
 - City names
 - Landmarks
 - Postal codes
 - General location descriptions

 You should use your geocoding tools to find the most accurate coordinates and provide clear
 information about the location found.','{"mode": "full", "limit": 50, "messageTypes": ["chat", "tool-result"], "includeInternal": false, "maxOutputTokens": 4000}',NULL,NULL,'2026-01-22 16:19:32.848333','2026-01-22 16:19:32.848333');`,
					`INSERT INTO "sub_agents" ("tenant_id","id","project_id","agent_id","name","description","prompt","conversation_history_config","models","stop_when","created_at","updated_at") VALUES ('default','weather-assistant','my-weather-project','weather-agent','Weather assistant','Main weather assistant that coordinates between geocoding and forecasting services to provide comprehensive weather information. This assistant handles user queries and delegates tasks to specialized sub-agents as needed.','You are a helpful weather assistant that provides comprehensive weather information
 for any location worldwide. You coordinate with specialized agents to:

 1. Convert location names/addresses to coordinates (via geocoder)
 2. Retrieve detailed weather forecasts (via weather forecaster)
 3. Present weather information in a clear, user-friendly format

 When users ask about weather:
 - If they provide a location name or address, delegate to the geocoder first
 - Once you have coordinates, delegate to the weather forecaster
 - Present the final weather information in an organized, easy-to-understand format
 - Include relevant details like temperature, conditions, precipitation, wind, etc.
 - Provide helpful context and recommendations when appropriate

 You have access to weather forecast data components that can enhance your responses
 with structured weather information.','{"mode": "full", "limit": 50, "messageTypes": ["chat", "tool-result"], "includeInternal": false, "maxOutputTokens": 4000}',NULL,NULL,'2026-01-22 16:19:32.851804','2026-01-22 16:19:32.851804');`,
					`INSERT INTO "sub_agents" ("tenant_id","id","project_id","agent_id","name","description","prompt","conversation_history_config","models","stop_when","created_at","updated_at") VALUES ('default','weather-forecaster','my-weather-project','weather-agent','Weather forecaster','Specialized agent for retrieving detailed weather forecasts and current conditions. This agent focuses on providing accurate, up-to-date weather information using geographic coordinates.','You are a weather forecasting specialist that provides detailed weather information
 including current conditions, forecasts, and weather-related insights.

 You work with precise geographic coordinates to deliver:
 - Current weather conditions
 - Short-term and long-term forecasts
 - Temperature, humidity, wind, and precipitation data
 - Weather alerts and advisories
 - Seasonal and climate information

 Always provide clear, actionable weather information that helps users plan their activities.','{"mode": "full", "limit": 50, "messageTypes": ["chat", "tool-result"], "includeInternal": false, "maxOutputTokens": 4000}',NULL,NULL,'2026-01-22 16:19:32.844618','2026-01-22 16:19:32.844618');`,
					`INSERT INTO "tools" ("tenant_id","id","project_id","name","description","config","credential_reference_id","credential_scope","headers","image_url","capabilities","last_error","created_at","updated_at") VALUES ('default','fUI2riwrBVJ6MepT8rjx0','my-weather-project','Forecast weather',NULL,'{"mcp": {"server": {"url": "https://weather-mcp-hazel.vercel.app/mcp"}}, "type": "mcp"}',NULL,'project',NULL,NULL,NULL,NULL,'2026-01-22 16:19:32.748','2026-01-22 16:19:32.748');`,
					`INSERT INTO "tools" ("tenant_id","id","project_id","name","description","config","credential_reference_id","credential_scope","headers","image_url","capabilities","last_error","created_at","updated_at") VALUES ('default','fdxgfv9HL7SXlfynPx8hf','my-weather-project','Geocode address',NULL,'{"mcp": {"server": {"url": "https://weather-mcp-hazel.vercel.app/mcp"}}, "type": "mcp"}',NULL,'project',NULL,NULL,NULL,NULL,'2026-01-22 16:19:32.75','2026-01-22 16:19:32.75');`,
					`INSERT INTO "sub_agent_relations" ("tenant_id","id","project_id","agent_id","source_sub_agent_id","target_sub_agent_id","relation_type","created_at","updated_at") VALUES ('default','0y59hwkkyzml4dq4t1sx8','my-weather-project','weather-agent','weather-assistant','weather-forecaster','delegate','2026-01-22 16:19:32.92219','2026-01-22 16:19:32.92219');`,
					`INSERT INTO "sub_agent_relations" ("tenant_id","id","project_id","agent_id","source_sub_agent_id","target_sub_agent_id","relation_type","created_at","updated_at") VALUES ('default','7ye45uc4j5442ihgqwn6d','my-weather-project','weather-agent','weather-assistant','geocoder-agent','delegate','2026-01-22 16:19:32.925527','2026-01-22 16:19:32.925527');`,
					`INSERT INTO "sub_agent_data_components" ("tenant_id","id","project_id","agent_id","sub_agent_id","data_component_id","created_at") VALUES ('default','689yd78rj16p9880bndfo','my-weather-project','weather-agent','weather-assistant','weather-forecast','2026-01-22 16:19:32.907332');`,
					`INSERT INTO "sub_agent_tool_relations" ("tenant_id","id","project_id","agent_id","sub_agent_id","tool_id","selected_tools","headers","tool_policies","created_at","updated_at") VALUES ('default','4kws0lm8bqi1mkzwbvmz4','my-weather-project','weather-agent','weather-forecaster','fUI2riwrBVJ6MepT8rjx0',NULL,NULL,NULL,'2026-01-22 16:19:32.888','2026-01-22 16:19:32.888');`,
					`INSERT INTO "sub_agent_tool_relations" ("tenant_id","id","project_id","agent_id","sub_agent_id","tool_id","selected_tools","headers","tool_policies","created_at","updated_at") VALUES ('default','ttz1a9tnso0sxim79iphr','my-weather-project','weather-agent','geocoder-agent','fdxgfv9HL7SXlfynPx8hf',NULL,NULL,NULL,'2026-01-22 16:19:32.889','2026-01-22 16:19:32.889');`,
					`SELECT DOLT_COMMIT('-Am', '//Update /manage/tenants/default/project-full/my-weather-project via API');`,
					`UPDATE "tools" SET "updated_at"='2026-01-22 16:19:50.912' WHERE "tenant_id"='default' AND "id"='fUI2riwrBVJ6MepT8rjx0' AND "project_id"='my-weather-project';`,
					`UPDATE "tools" SET "updated_at"='2026-01-22 16:19:50.967' WHERE "tenant_id"='default' AND "id"='fdxgfv9HL7SXlfynPx8hf' AND "project_id"='my-weather-project';`,
					`SELECT DOLT_COMMIT('-Am', 'GET /manage/tenants/default/projects/my-weather-project/tools via API');`,
					`INSERT INTO "evaluator" ("tenant_id","id","project_id","name","description","prompt","schema","model","pass_criteria","created_at","updated_at") VALUES ('default','ubqho5lsm6h7bd3ra8loz','my-weather-project','test','test','test','{"type": "object", "required": ["test"], "properties": {"test": {"type": "string", "description": "test"}}, "additionalProperties": false}','{"model": "anthropic/claude-opus-4-5"}',NULL,'2026-01-22 16:20:07.188','2026-01-22 16:20:07.188');`,
					`SELECT DOLT_COMMIT('-Am', 'Create /manage/tenants/default/projects/my-weather-project/evals/evaluators via API');`,
					`UPDATE "tools" SET "updated_at"='2026-01-22 16:20:11.438' WHERE "tenant_id"='default' AND "id"='fUI2riwrBVJ6MepT8rjx0' AND "project_id"='my-weather-project';`,
					`UPDATE "tools" SET "updated_at"='2026-01-22 16:20:11.448' WHERE "tenant_id"='default' AND "id"='fdxgfv9HL7SXlfynPx8hf' AND "project_id"='my-weather-project';`,
					`SELECT DOLT_COMMIT('-Am', 'GET /manage/tenants/default/projects/my-weather-project/tools via API');`,
					`UPDATE "tools" SET "updated_at"='2026-01-22 16:20:17.821' WHERE "tenant_id"='default' AND "id"='fUI2riwrBVJ6MepT8rjx0' AND "project_id"='my-weather-project';`,
					`UPDATE "tools" SET "updated_at"='2026-01-22 16:20:18.082' WHERE "tenant_id"='default' AND "id"='fdxgfv9HL7SXlfynPx8hf' AND "project_id"='my-weather-project';`,
					`SELECT DOLT_COMMIT('-Am', 'GET /manage/tenants/default/projects/my-weather-project/tools via API');`,
					`INSERT INTO "evaluation_job_config" ("tenant_id","id","project_id","job_filters","created_at","updated_at") VALUES ('default','tj06kzjt8ltlyixgfzeao','my-weather-project','{"dateRange": {"endDate": "2026-01-23T04:59:59.999Z", "startDate": "2026-01-21T05:00:00.000Z"}}','2026-01-22 16:20:55.774','2026-01-22 16:20:55.774');`,
					`INSERT INTO "evaluation_job_config_evaluator_relations" ("tenant_id","id","project_id","evaluation_job_config_id","evaluator_id","created_at","updated_at") VALUES ('default','5qk0w692h5ij1sxtohdua','my-weather-project','tj06kzjt8ltlyixgfzeao','ubqho5lsm6h7bd3ra8loz','2026-01-22 16:20:55.781','2026-01-22 16:20:55.781');`,
					`SELECT DOLT_COMMIT('-Am', 'Create /manage/tenants/default/projects/my-weather-project/evals/evaluation-job-configs via API');`,
					`INSERT INTO "evaluation_suite_config" ("tenant_id","id","project_id","filters","sample_rate","created_at","updated_at") VALUES ('default','j5gvgluqzwzhjhycrsnpf','my-weather-project','{"agentIds": ["weather-agent"]}',NULL,'2026-01-22 16:21:19.974','2026-01-22 16:21:19.974');`,
					`INSERT INTO "evaluation_suite_config_evaluator_relations" ("tenant_id","id","project_id","evaluation_suite_config_id","evaluator_id","created_at","updated_at") VALUES ('default','tz51dzynx71gits265e9d','my-weather-project','j5gvgluqzwzhjhycrsnpf','ubqho5lsm6h7bd3ra8loz','2026-01-22 16:21:19.982','2026-01-22 16:21:19.982');`,
					`SELECT DOLT_COMMIT('-Am', 'Create /manage/tenants/default/projects/my-weather-project/evals/evaluation-suite-configs via API');`,
					`INSERT INTO "evaluation_run_config" ("tenant_id","id","project_id","name","description","is_active","created_at","updated_at") VALUES ('default','74pgwrprmea2o7e6avbh7','my-weather-project','test','test',true,'2026-01-22 16:21:20.104','2026-01-22 16:21:20.104');`,
					`INSERT INTO "evaluation_run_config_evaluation_suite_config_relations" ("tenant_id","id","project_id","evaluation_run_config_id","evaluation_suite_config_id","created_at","updated_at") VALUES ('default','plb31qfzw9803g6hbjhef','my-weather-project','74pgwrprmea2o7e6avbh7','j5gvgluqzwzhjhycrsnpf','2026-01-22 16:21:20.111','2026-01-22 16:21:20.111');`,
					`SELECT DOLT_COMMIT('-Am', 'Create /manage/tenants/default/projects/my-weather-project/evals/evaluation-run-configs via API');`,
					`UPDATE "tools" SET "updated_at"='2026-01-22 16:21:23.521' WHERE "tenant_id"='default' AND "id"='fUI2riwrBVJ6MepT8rjx0' AND "project_id"='my-weather-project';`,
					`UPDATE "tools" SET "updated_at"='2026-01-22 16:21:23.771' WHERE "tenant_id"='default' AND "id"='fdxgfv9HL7SXlfynPx8hf' AND "project_id"='my-weather-project';`,
					`SELECT DOLT_COMMIT('-Am', 'GET /manage/tenants/default/projects/my-weather-project/tools via API');`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
						Expected: []sql.Row{{"t"}},
					},
				},
			},
			{
				Name: "Merge foreign keys across types, text -> varchar",
				SetUpScript: []string{
					`CREATE TABLE table1 (
            table1_col1 VARCHAR(256),
            table1_col2 VARCHAR(256),
            table1_col3 VARCHAR(256),
            PRIMARY KEY (table1_col1, table1_col3, table1_col2)
        );`,
					`CREATE TABLE table2 (
            table2_col1 VARCHAR(256),
            table2_col2 VARCHAR(256),
            table2_col3 VARCHAR(256),
            table2_col4 TEXT,
            PRIMARY KEY (table2_col1, table2_col3, table2_col2),
            CONSTRAINT table2_fk FOREIGN KEY (table2_col1, table2_col3, table2_col4) REFERENCES table1 (table1_col1, table1_col3, table1_col2) ON DELETE CASCADE ON UPDATE NO ACTION
        );`,
					`SELECT DOLT_COMMIT('-Am', '1');`,
					`SELECT DOLT_BRANCH('other_branch');`,
					`CREATE TABLE table3 (table3_col1 VARCHAR(256));`,
					`SELECT DOLT_COMMIT('-Am', '2');`,
					`SELECT DOLT_CHECKOUT('other_branch');`,
					`INSERT INTO table1 (table1_col1, table1_col2, table1_col3) VALUES ('abc','def','ghi');`,
					`SELECT DOLT_COMMIT('-Am', '3');`,
					`INSERT INTO table2 (table2_col1, table2_col2, table2_col3, table2_col4) VALUES ('abc','jkl','ghi','def');`,
					`SELECT DOLT_COMMIT('-Am', '4');`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
						Expected: []sql.Row{{"t"}},
					},
				},
			},
			{
				Name: "Merge foreign keys across types, varchar -> text",
				SetUpScript: []string{
					`CREATE TABLE table1 (
            table1_col1 VARCHAR(256),
            table1_col2 text,
            table1_col3 text,
            PRIMARY KEY (table1_col1, table1_col3, table1_col2)
        );`,
					`CREATE TABLE table2 (
            table2_col1 VARCHAR(256),
            table2_col2 VARCHAR(256),
            table2_col3 VARCHAR(256),
            table2_col4 VARCHAR(256),
            PRIMARY KEY (table2_col1, table2_col3, table2_col2),
            CONSTRAINT table2_fk FOREIGN KEY (table2_col1, table2_col3, table2_col4) REFERENCES table1 (table1_col1, table1_col3, table1_col2) ON DELETE CASCADE ON UPDATE NO ACTION
        );`,
					`SELECT DOLT_COMMIT('-Am', '1');`,
					`SELECT DOLT_BRANCH('other_branch');`,
					`CREATE TABLE table3 (table3_col1 VARCHAR(256));`,
					`SELECT DOLT_COMMIT('-Am', '2');`,
					`SELECT DOLT_CHECKOUT('other_branch');`,
					`INSERT INTO table1 (table1_col1, table1_col2, table1_col3) VALUES ('abc','def','ghi');`,
					`SELECT DOLT_COMMIT('-Am', '3');`,
					`INSERT INTO table2 (table2_col1, table2_col2, table2_col3, table2_col4) VALUES ('abc','jkl','ghi','def');`,
					`SELECT DOLT_COMMIT('-Am', '4');`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
						Expected: []sql.Row{{"t"}},
					},
				},
			},
			{
				Name: "Merge foreign keys across types, varchar -> text, violation",
				SetUpScript: []string{
					`CREATE TABLE parent (a TEXT PRIMARY KEY);`,
					`CREATE TABLE child (b INT PRIMARY KEY, c varchar(255), CONSTRAINT fk FOREIGN KEY (c) REFERENCES parent(a) ON DELETE CASCADE ON UPDATE NO ACTION);`,
					`INSERT INTO parent (a) VALUES ('abc'), ('def');`,
					`INSERT INTO child (b, c) VALUES (1, 'abc'), (2, 'def');`,
					`SELECT DOLT_COMMIT('-Am', '1');`,
					`SELECT DOLT_BRANCH('other_branch');`,
					`DELETE FROM child WHERE b=1;`,
					`DELETE FROM parent WHERE a='abc';`,
					`SELECT DOLT_COMMIT('-Am', '2');`,
					`SELECT DOLT_CHECKOUT('other_branch');`,
					`INSERT INTO child (b, c) VALUES (3, 'abc');`,
					`SELECT DOLT_COMMIT('-Am', '3');`,
					`set dolt_force_transaction_commit=1;`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
						Expected: []sql.Row{{"f"}},
					},
					{
						Query:    "select violation_type, b, c from dolt_constraint_violations_child;",
						Expected: []sql.Row{{"foreign key", 3, "abc"}},
					},
				},
			},
			{
				Name: "Merge foreign keys across types, varchar -> text 2 column key, violation",
				SetUpScript: []string{
					`CREATE TABLE parent (
            a VARCHAR(256),
            b text,
            c VARCHAR(256),
            PRIMARY KEY (b, a)
        );`,
					`CREATE INDEX idx_parent_on_a_b ON parent (a, b);`,
					`CREATE TABLE child (
            d VARCHAR(256),
            e VARCHAR(256),
            f varchar(256),
            PRIMARY KEY (e, d),
            CONSTRAINT child_fk FOREIGN KEY (d, f) REFERENCES parent (a, b) ON DELETE CASCADE ON UPDATE NO ACTION
        );`,
					`INSERT INTO parent VALUES ('abc','def', 'xyz');`,
					`INSERT INTO child VALUES ('abc','123','def');`,
					`SELECT DOLT_COMMIT('-Am', '1');`,
					`SELECT DOLT_BRANCH('other_branch');`,
					`INSERT INTO child VALUES ('abc','www','def');`,
					`SELECT DOLT_COMMIT('-Am', '2');`,
					`CREATE TABLE table3 (table3_col1 VARCHAR(256));`,
					`SELECT DOLT_COMMIT('-Am', '3');`,
					`SELECT DOLT_CHECKOUT('other_branch');`,
					`delete from child where e='def';`,
					`delete from parent where a='abc';`,
					`SELECT DOLT_COMMIT('-Am', '4');`,
					`set dolt_force_transaction_commit=1;`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
						Expected: []sql.Row{{"f"}},
					},
					{
						Query:    "select violation_type, d, e, f from dolt_constraint_violations_child;",
						Expected: []sql.Row{{"foreign key", "abc", "www", "def"}},
					},
				},
			},
			{
				Name: "Merge foreign keys across types, varchar -> text 2 column key, no violation",
				SetUpScript: []string{
					`CREATE TABLE parent (
            a VARCHAR(256),
            b text,
            c VARCHAR(256),
            PRIMARY KEY (b, a)
        );`,
					`CREATE INDEX idx_parent_on_a_b ON parent (a, b);`,
					`CREATE TABLE child (
            d VARCHAR(256),
            e VARCHAR(256),
            f varchar(256),
            PRIMARY KEY (e, d),
            CONSTRAINT child_fk FOREIGN KEY (d, f) REFERENCES parent (a, b) ON DELETE CASCADE ON UPDATE NO ACTION
        );`,
					`INSERT INTO parent VALUES ('abc','def', 'xyz');`,
					`INSERT INTO child VALUES ('abc','123','def');`,
					`SELECT DOLT_COMMIT('-Am', '1');`,
					`SELECT DOLT_BRANCH('other_branch');`,
					`INSERT INTO child VALUES ('abc','www','def');`,
					`SELECT DOLT_COMMIT('-Am', '2');`,
					`CREATE TABLE table3 (table3_col1 VARCHAR(256));`,
					`SELECT DOLT_COMMIT('-Am', '3');`,
					`SELECT DOLT_CHECKOUT('other_branch');`,
					`INSERT INTO child VALUES ('abc','xyz','def');`,
					`SELECT DOLT_COMMIT('-Am', '4');`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
						Expected: []sql.Row{{"t"}},
					},
					{
						Query:    "select violation_type, d, e, f from dolt_constraint_violations_child;",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "Merge foreign keys across types, varchar -> text 2 column key, parent primary index, violation",
				SetUpScript: []string{
					`CREATE TABLE parent (
            a VARCHAR(256),
            b text,
            c VARCHAR(256),
            PRIMARY KEY (b, a)
        );`,
					`CREATE TABLE child (
            d VARCHAR(256),
            e VARCHAR(256),
            f varchar(256),
            PRIMARY KEY (e, d),
            CONSTRAINT child_fk FOREIGN KEY (f, d) REFERENCES parent (b, a) ON DELETE CASCADE ON UPDATE NO ACTION
        );`,
					`INSERT INTO parent VALUES ('abc','def', 'xyz');`,
					`INSERT INTO child VALUES ('abc','123','def');`,
					`SELECT DOLT_COMMIT('-Am', '1');`,
					`SELECT DOLT_BRANCH('other_branch');`,
					`INSERT INTO child VALUES ('abc','www','def');`,
					`SELECT DOLT_COMMIT('-Am', '2');`,
					`CREATE TABLE table3 (table3_col1 VARCHAR(256));`,
					`SELECT DOLT_COMMIT('-Am', '3');`,
					`SELECT DOLT_CHECKOUT('other_branch');`,
					`delete from child where e='def';`,
					`delete from parent where a='abc';`,
					`SELECT DOLT_COMMIT('-Am', '4');`,
					`set dolt_force_transaction_commit=1;`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
						Expected: []sql.Row{{"f"}},
					},
					{
						Query:    "select violation_type, d, e, f from dolt_constraint_violations_child;",
						Expected: []sql.Row{{"foreign key", "abc", "www", "def"}},
					},
				},
			},
			{
				Name: "Merge foreign keys across types, varchar -> text 2 column key, parent primary index, no violation",
				SetUpScript: []string{
					`CREATE TABLE parent (
            a VARCHAR(256),
            b text,
            c VARCHAR(256),
            PRIMARY KEY (b, a)
        );`,
					`CREATE TABLE child (
            d VARCHAR(256),
            e VARCHAR(256),
            f varchar(256),
            PRIMARY KEY (e, d),
            CONSTRAINT child_fk FOREIGN KEY (f, d) REFERENCES parent (b, a) ON DELETE CASCADE ON UPDATE NO ACTION
        );`,
					`INSERT INTO parent VALUES ('abc','def', 'xyz');`,
					`INSERT INTO child VALUES ('abc','123','def');`,
					`SELECT DOLT_COMMIT('-Am', '1');`,
					`SELECT DOLT_BRANCH('other_branch');`,
					`INSERT INTO child VALUES ('abc','www','def');`,
					`SELECT DOLT_COMMIT('-Am', '2');`,
					`CREATE TABLE table3 (table3_col1 VARCHAR(256));`,
					`SELECT DOLT_COMMIT('-Am', '3');`,
					`SELECT DOLT_CHECKOUT('other_branch');`,
					`INSERT INTO child VALUES ('abc','xyz','def');`,
					`SELECT DOLT_COMMIT('-Am', '4');`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
						Expected: []sql.Row{{"t"}},
					},
					{
						Query:    "select violation_type, d, e, f from dolt_constraint_violations_child;",
						Expected: []sql.Row{},
					},
				},
			},
			{
				Name: "Merge foreign keys across types, varchar -> text 2 column key, child primary index, violation",
				SetUpScript: []string{
					`CREATE TABLE parent (
            a VARCHAR(256),
            b text,
            c VARCHAR(256),
            PRIMARY KEY (b, a)
        );`,
					`CREATE TABLE child (
            d VARCHAR(256),
            e VARCHAR(256),
            f varchar(256),
            PRIMARY KEY (e, d)
        );`,
					`INSERT INTO parent VALUES ('abc','def', 'xyz');`,
					`INSERT INTO child VALUES ('abc','123', 'def');`,
					`SELECT DOLT_COMMIT('-Am', '1');`,
					`SELECT DOLT_BRANCH('other_branch');`,
					`ALTER TABLE child ADD CONSTRAINT child_fk FOREIGN KEY (f, d) REFERENCES parent (b, a);`,
					`SELECT DOLT_COMMIT('-Am', '2');`,
					`CREATE TABLE table3 (table3_col1 VARCHAR(256));`,
					`SELECT DOLT_COMMIT('-Am', '3');`,
					`SELECT DOLT_CHECKOUT('other_branch');`,
					`INSERT INTO child VALUES ('abc','def','xxx');`,
					`SELECT DOLT_COMMIT('-Am', '4');`,
					`set dolt_force_transaction_commit=1;`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
						Expected: []sql.Row{{"f"}},
					},
					{
						Query:    "select violation_type, d, e, f from dolt_constraint_violations_child;",
						Expected: []sql.Row{{"foreign key", "abc", "def", "xxx"}},
					},
				},
			},
			{
				Name: "Merge foreign keys across types, varchar -> text 2 column key, child secondary index, violation",
				SetUpScript: []string{
					`CREATE TABLE parent (
            a VARCHAR(256),
            b text,
            c VARCHAR(256),
            PRIMARY KEY (b, a)
        );`,
					`CREATE TABLE child (
            d VARCHAR(256),
            e VARCHAR(256),
            f varchar(256),
            g VARCHAR(256),
            PRIMARY KEY (e, d)
        );`,
					`INSERT INTO parent VALUES ('abc','def', 'xyz');`,
					`INSERT INTO child VALUES ('abc','123', 'abc', 'def');`,
					`SELECT DOLT_COMMIT('-Am', '1');`,
					`SELECT DOLT_BRANCH('other_branch');`,
					`ALTER TABLE child ADD CONSTRAINT child_fk FOREIGN KEY (g, f) REFERENCES parent (b, a);`,
					`SELECT DOLT_COMMIT('-Am', '2');`,
					`CREATE TABLE table3 (table3_col1 VARCHAR(256));`,
					`SELECT DOLT_COMMIT('-Am', '3');`,
					`SELECT DOLT_CHECKOUT('other_branch');`,
					`INSERT INTO child VALUES ('xyz','123', 'def', 'abc');`,
					`SELECT DOLT_COMMIT('-Am', '4');`,
					`set dolt_force_transaction_commit=1;`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
						Expected: []sql.Row{{"f"}},
					},
					{
						Query:    "select violation_type, d, e, f, g from dolt_constraint_violations_child;",
						Expected: []sql.Row{{"foreign key", "xyz", "123", "def", "abc"}},
					},
				},
			},
			{
				Name: "Merge foreign keys across types, varchar -> text 3 column key, violation",
				SetUpScript: []string{
					`CREATE TABLE table1 (
            table1_col1 VARCHAR(256),
            table1_col2 text,
            table1_col3 text,
            PRIMARY KEY (table1_col1, table1_col3, table1_col2)
        );`,
					`CREATE TABLE table2 (
            table2_col1 VARCHAR(256),
            table2_col2 VARCHAR(256),
            table2_col3 VARCHAR(256),
            table2_col4 VARCHAR(256),
            PRIMARY KEY (table2_col1, table2_col3, table2_col2),
            CONSTRAINT table2_fk FOREIGN KEY (table2_col1, table2_col3, table2_col4) REFERENCES table1 (table1_col1, table1_col3, table1_col2) ON DELETE CASCADE ON UPDATE NO ACTION
        );`,
					`INSERT INTO table1 (table1_col1, table1_col2, table1_col3) VALUES ('abc','def','ghi');`,
					`INSERT INTO table2 (table2_col1, table2_col2, table2_col3, table2_col4) VALUES ('abc','jkl','ghi','def');`,
					`SELECT DOLT_COMMIT('-Am', '1');`,
					`SELECT DOLT_BRANCH('other_branch');`,
					`INSERT INTO table2 (table2_col1, table2_col2, table2_col3, table2_col4) VALUES ('abc','xyz','ghi','def');`,
					`SELECT DOLT_COMMIT('-Am', '2');`,
					`CREATE TABLE table3 (table3_col1 VARCHAR(256));`,
					`SELECT DOLT_COMMIT('-Am', '3');`,
					`SELECT DOLT_CHECKOUT('other_branch');`,
					`delete from table2 where table2_col2='jkl';`,
					`delete from table1 where table1_col2='def';`,
					`SELECT DOLT_COMMIT('-Am', '4');`,
					`set dolt_force_transaction_commit=1;`,
				},
				Assertions: []ScriptTestAssertion{
					{
						Query:    "select strpos(dolt_merge('main')::text, 'merge successful') > 0;",
						Expected: []sql.Row{{"f"}},
					},
					{
						Query:    "select violation_type, table2_col1, table2_col2, table2_col3, table2_col4 from dolt_constraint_violations_table2;",
						Expected: []sql.Row{{"foreign key", "abc", "xyz", "ghi", "def"}},
					},
				},
			},
		},
	)
}
