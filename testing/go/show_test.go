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

func TestDescribe(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "describe table",
			SetUpScript: []string{
				`CREATE TABLE t1 (id INT PRIMARY KEY, name TEXT)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `EXPLAIN t1`,
					Expected: []sql.Row{
						{"id", "integer", "NO", "PRI", interface{}(nil), ""},
						{"name", "text", "YES", "", interface{}(nil), ""},
					},
				},
				{
					Query: `DESCRIBE t1`,
					Expected: []sql.Row{
						{"id", "integer", "NO", "PRI", interface{}(nil), ""},
						{"name", "text", "YES", "", interface{}(nil), ""},
					},
				},
				{
					Query: `DESC t1`,
					Expected: []sql.Row{
						{"id", "integer", "NO", "PRI", interface{}(nil), ""},
						{"name", "text", "YES", "", interface{}(nil), ""},
					},
				},
				{
					Query: `DESC public.t1`,
					Expected: []sql.Row{
						{"id", "integer", "NO", "PRI", interface{}(nil), ""},
						{"name", "text", "YES", "", interface{}(nil), ""},
					},
				},
				{
					Query: `DESC postgres.public.t1`,
					Expected: []sql.Row{
						{"id", "integer", "NO", "PRI", interface{}(nil), ""},
						{"name", "text", "YES", "", interface{}(nil), ""},
					},
				},
			},
		},
		{
			Name: "describe table AS OF",
			SetUpScript: []string{
				`CREATE TABLE t1 (id INT PRIMARY KEY, name TEXT)`,
				`call dolt_commit('-Am', 'first commit')`,
				`ALTER TABLE t1 ADD COLUMN age INT`,
				`call dolt_commit('-am', 'second commit')`,
				`ALTER TABLE t1 ADD COLUMN height INT`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `DESC t1`,
					Expected: []sql.Row{
						{"id", "integer", "NO", "PRI", interface{}(nil), ""},
						{"name", "text", "YES", "", interface{}(nil), ""},
						{"age", "integer", "YES", "", interface{}(nil), ""},
						{"height", "integer", "YES", "", interface{}(nil), ""},
					},
				},
				{
					Query: `EXPLAIN public.t1 AS OF 'HEAD'`,
					Expected: []sql.Row{
						{"id", "integer", "NO", "PRI", interface{}(nil), ""},
						{"name", "text", "YES", "", interface{}(nil), ""},
						{"age", "integer", "YES", "", interface{}(nil), ""},
					},
				},
				{
					Query: `DESCRIBE postgres.public.t1 AS OF 'HEAD~'`,
					Expected: []sql.Row{
						{"id", "integer", "NO", "PRI", interface{}(nil), ""},
						{"name", "text", "YES", "", interface{}(nil), ""},
					},
				},
			},
		},
		{
			Name: "describe table in other schema",
			SetUpScript: []string{
				`CREATE TABLE t1 (a INT PRIMARY KEY, b TEXT)`,
				`create schema schema2`,
				`CREATE TABLE schema2.t2 (c INT PRIMARY KEY, d TEXT)`,
				`create schema schema3`,
				`CREATE TABLE schema3.t2 (e INT PRIMARY KEY, f TEXT)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `DESC schema2.t2`,
					Expected: []sql.Row{
						{"c", "integer", "NO", "PRI", interface{}(nil), ""},
						{"d", "text", "YES", "", interface{}(nil), ""},
					},
				},
				{
					Query: `DESC postgres.schema2.t2`,
					Expected: []sql.Row{
						{"c", "integer", "NO", "PRI", interface{}(nil), ""},
						{"d", "text", "YES", "", interface{}(nil), ""},
					},
				},
				{
					Query:       `DESC t2`,
					ExpectedErr: "not found",
				},
				{
					Query: `SET search_path TO 'schema2'`,
				},
				{
					Query: `DESC t2`,
					Expected: []sql.Row{
						{"c", "integer", "NO", "PRI", interface{}(nil), ""},
						{"d", "text", "YES", "", interface{}(nil), ""},
					},
				},
				{
					Query: "SET search_path TO 'schema3,schema2'",
				},
				{
					Query: `DESC t2`,
					Expected: []sql.Row{
						{"e", "integer", "NO", "PRI", interface{}(nil), ""},
						{"f", "text", "YES", "", interface{}(nil), ""},
					},
				},
			},
		},
	})
}

func TestShowTables(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "show tables in single schema",
			SetUpScript: []string{
				`CREATE TABLE t1 (a INT PRIMARY KEY, name TEXT)`,
				`CREATE TABLE t2 (b INT PRIMARY KEY, name TEXT)`,
				`create schema schema2`,
				`create database db2`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SHOW TABLES`,
					Expected: []sql.Row{
						{"t1"},
						{"t2"},
					},
				},
				{
					Query: `SHOW TABLES from public`,
					Expected: []sql.Row{
						{"t1"},
						{"t2"},
					},
				},
				{
					Query:    `SHOW TABLES from schema2`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SHOW TABLES from schema3`,
					ExpectedErr: "not found",
				},
				{
					Query: `SHOW TABLES from postgres.public`,
					Expected: []sql.Row{
						{"t1"},
						{"t2"},
					},
				},
				{
					Query:    `SHOW TABLES from postgres.schema2`,
					Expected: []sql.Row{},
				},
				{
					Query:       `SHOW TABLES from postgres.schema3`,
					ExpectedErr: "not found",
				},
				{
					Query:       `SHOW TABLES from db3`,
					ExpectedErr: "not found",
				},
			},
		},
		{
			Name: "show tables in multiple schemas, dbs",
			SetUpScript: []string{
				`CREATE TABLE t1 (a INT PRIMARY KEY, name TEXT)`,
				`CREATE TABLE t2 (b INT PRIMARY KEY, name TEXT)`,
				`create schema schema2`,
				`CREATE TABLE schema2.t3 (a INT PRIMARY KEY, name TEXT)`,
				`CREATE TABLE schema2.t4 (b INT PRIMARY KEY, name TEXT)`,
				`create database db2`,
				`use db2`,
				`CREATE TABLE t5 (a INT PRIMARY KEY, name TEXT)`,
				`create schema schema3`,
				`CREATE TABLE schema3.t6 (b INT PRIMARY KEY, name TEXT)`,
				`use postgres`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SHOW TABLES`,
					Expected: []sql.Row{
						{"t1"},
						{"t2"},
					},
				},
				{
					Query: `SHOW TABLES from public`,
					Expected: []sql.Row{
						{"t1"},
						{"t2"},
					},
				},
				{
					Query: `SHOW TABLES from schema2`,
					Expected: []sql.Row{
						{"t3"},
						{"t4"},
					},
				},
				{
					Query: `SHOW TABLES from postgres.public`,
					Expected: []sql.Row{
						{"t1"},
						{"t2"},
					},
				},
				{
					Query: `SHOW TABLES from postgres.schema2`,
					Expected: []sql.Row{
						{"t3"},
						{"t4"},
					},
				},
				{
					Query:       `SHOW TABLES from db2`,
					ExpectedErr: "not found",
				},
				{
					Query: `SHOW TABLES from db2.public`,
					Expected: []sql.Row{
						{"t5"},
					},
				},
				{
					Query: `SHOW TABLES from db2.schema3`,
					Expected: []sql.Row{
						{"t6"},
					},
				},
				{
					Query: `SET search_path TO 'schema2'`,
				},
				{
					Query: `SHOW TABLES`,
					Expected: []sql.Row{
						{"t3"},
						{"t4"},
					},
				},
				{
					Query: `SET search_path TO 'schema3'`,
				},
				{
					Query:    `SHOW TABLES`,
					Expected: []sql.Row{},
				},
			},
		},
	})
}

func TestShowCreateTable(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:  "show create table",
			SetUpScript: []string{
				`CREATE TABLE t1 (a INT PRIMARY KEY, name TEXT)`,
				`CREATE TABle t2 (b SERIAL PRIMARY KEY, time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
				`CREATE TABLE t3 (a timestamp PRIMARY KEY, name varchar(100))`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SHOW CREATE TABLE t1`,
					Expected: []sql.Row{
						{"t1", `CREATE TABLE "t1" (
  "a" integer NOT NULL,
  "name" text,
  PRIMARY KEY ("a")
)`},
					},
				},
				{
					Query: `SHOW CREATE TABLE T2`,
					Expected: []sql.Row{
						{"t2", `CREATE TABLE "t2" (
  "b" integer NOT NULL DEFAULT (nextval('public.t2_b_seq')),
  "time" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY ("b")
)`,
						},
					},
				},
				{
					Query: `SHOW CREATE TABLE t3`,
					Expected: []sql.Row{
						{"t3", `CREATE TABLE "t3" (
  "a" timestamp NOT NULL,
  "name" varchar(100),
  PRIMARY KEY ("a")
)`},
					},
				},
				{
					Query:       `SHOW CREATE TABLE dne`,
					ExpectedErr: "not found",
				},
			},
		},
	})
}

func TestShowIndexes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "show indexes",
			SetUpScript: []string{
				`CREATE TABLE t1 (a INT PRIMARY KEY, name TEXT, value int)`,
				`CREATE INDEX idx_name ON t1(name)`,
				`CREATE INDEX idx_value ON t1(value)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SHOW indexes FROM t1`,
					Expected: []sql.Row{
						{"t1", 0, "PRIMARY", 1, "a", nil, 0, nil, nil, "", "BTREE", "", "", "YES", nil},
						{"t1", 1, "idx_name", 1, "name", nil, 0, nil, nil, "YES", "BTREE", "", "", "YES", nil},
						{"t1", 1, "idx_value", 1, "value", nil, 0, nil, nil, "YES", "BTREE", "", "", "YES", nil},
					},
				},
				{
					Query:       `SHOW indexes FROM dne`,
					ExpectedErr: "not found",
				},
			},
		},
	})
}

func TestShowDatabasesAndSchemas(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "show databases",
			SetUpScript: []string{
				`CREATE DATABASE db1`,
				`CREATE DATABASE db2`,
				`CREATE SCHEMA schema1`,
				`CREATE SCHEMA schema2`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SHOW databases`,
					Expected: []sql.Row{
						{"db1"},
						{"db2"},
						{"information_schema"},
						{"postgres"},
					},
				},
				{
					Query: `show SCHEMAS`,
					Expected: []sql.Row{
						{"dolt"},
						{"pg_catalog"},
						{"information_schema"},
						{"schema1"},
						{"schema2"},
						{"public"},
					},
				},
				{
					Query: `show SCHEMAS FROM postgres`,
					Expected: []sql.Row{
						{"dolt"},
						{"pg_catalog"},
						{"information_schema"},
						{"schema1"},
						{"schema2"},
						{"public"},
					},
				},
				{
					Query: `show SCHEMAS FROM db1`,
					Expected: []sql.Row{
						{"dolt"},
						{"pg_catalog"},
						{"information_schema"},
						{"public"},
					},
				},
				{
					Query:       `show SCHEMAS FROM dne`,
					ExpectedErr: "not found",
				},
			},
		},
	})
}

func TestShowSequences(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "show sequences",
			SetUpScript: []string{
				`CREATE SEQUENCE seq1`,
				`CREATE SEQUENCE seq2`,
				`CREATE SCHEMA schema1`,
				`CREATE SEQUENCE schema1.seq3`,
				`CREATE DATABASE db1`,
				`USE db1`,
				`CREATE SEQUENCE seq4`,
				`CREATE SCHEMA schema2`,
				`CREATE SEQUENCE schema2.seq5`,
				`use postgres`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SHOW SEQUENCES`,
					Expected: []sql.Row{
						{"public", "seq1"},
						{"public", "seq2"},
						{"schema1", "seq3"},
					},
				},
				{
					Query: `SHOW SEQUENCES from postgres`,
					Expected: []sql.Row{
						{"public", "seq1"},
						{"public", "seq2"},
						{"schema1", "seq3"},
					},
				},
				{
					Query: `SHOW SEQUENCES FROM db1`,
					Expected: []sql.Row{
						{"public", "seq4"},
						{"schema2", "seq5"},
					},
				},
				{
					Query:       `SHOW SEQUENCES FROM dne`,
					ExpectedErr: "not found",
				},
			},
		},
	})
}
