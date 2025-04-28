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
					"call dolt_commit('-Am', 'create schemas')",
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
					"call dolt_commit('-Am', 'create schemas')",
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
					"call dolt_commit('-Am', 'create schemas')",
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
					"call dolt_commit('-Am', 'create schemas')",
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
					"call dolt_commit('-Am', 'create schemas')",
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
					"call dolt_commit('-Am', 'create schemas')",
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
					"call dolt_commit('-Am', 'create schemas')",
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
					"call dolt_commit('-Am', 'create schemas')",
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
		},
	)
}
