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

func TestCreateViewStatements(t *testing.T) {
	RunScripts(t, createViewStmts)
}

var createViewStmts = []ScriptTest{
	{
		Name: "basic create view statements",
		SetUpScript: []string{
			"create table t1 (pk int);",
			"insert into t1 values (1), (2), (3), (1);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "create view v as select * from t1 order by pk;",
				Expected: []sql.Row{},
			},
			{
				Query:    "select * from v order by pk;",
				Expected: []sql.Row{{1}, {1}, {2}, {3}},
			},
		},
	},
	{
		Name: "create view from view",
		SetUpScript: []string{
			"create table t1 (pk int);",
			"insert into t1 values (1), (2), (3), (1);",
			"create view v as select * from t1 where pk > 1;",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "create view v1 as select * from v order by pk;",
				Expected: []sql.Row{},
			},
			{
				Query:    "select * from v1 order by pk;",
				Expected: []sql.Row{{2}, {3}},
			},
		},
	},
	{
		Name: "view with expression name",
		SetUpScript: []string{
			"create view v as select 2+2",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "SELECT * from v;",
				Expected: []sql.Row{{4}},
			},
		},
	},
	{
		Name: "view with column names",
		SetUpScript: []string{
			`CREATE TABLE xy (x int primary key, y int);`,
			`insert into xy values (1, 4), (4, 9)`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "create view v_today(today) as select 2",
				Expected: []sql.Row{},
			},
			{
				Query:    "CREATE VIEW xyv (u,v) AS SELECT * from xy",
				Expected: []sql.Row{},
			},
			{
				Query:    "SELECT v from xyv;",
				Expected: []sql.Row{{4}, {9}},
			},
			{
				Query:    "SELECT today from v_today;",
				Expected: []sql.Row{{2}},
			},
		},
	},
	{
		Skip: true, // TODO: getting subquery alias not supported error
		Name: "nested view",
		SetUpScript: []string{
			"create table t1 (pk int);",
			"insert into t1 values (1), (2), (3), (4);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "create view unionView as (select * from t1 order by pk desc limit 1) union all (select * from t1 order by pk limit 1)",
				Expected: []sql.Row{},
			},
			{
				Query:    "select * from unionView order by pk;",
				Expected: []sql.Row{{1}, {4}},
			},
		},
	},
	{
		Name: "cast (postgres-specific syntax)",
		SetUpScript: []string{
			"create table t1 (pk int);",
			"insert into t1 values (1), (2), (3), (4);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "CREATE VIEW v AS SELECT pk::INT2 FROM t1 ORDER BY pk;",
				Expected: []sql.Row{},
			},
			{
				Query:    "select * from v order by pk;",
				Expected: []sql.Row{{1}, {2}, {3}, {4}},
			},
			{
				Query:    "CREATE VIEW v_text AS SELECT pk::int2, (pk)::text AS pk_text FROM t1;",
				Expected: []sql.Row{},
			},
			{
				Query:    "select pk_text from v_text order by pk;",
				Expected: []sql.Row{{"1"}, {"2"}, {"3"}, {"4"}},
			},
		},
	},
	{
		Name: "not yet supported create view queries",
		Assertions: []ScriptTestAssertion{
			{
				Query:       "CREATE TEMPORARY VIEW v AS SELECT 1;",
				Skip: true,
			},
			{
				Query:       "CREATE RECURSIVE VIEW v AS SELECT 1;",
				Skip: true,
			},
			{
				Query:       "CREATE VIEW v AS SELECT 1 WITH LOCAL CHECK OPTION;",
				Skip: true,
			},
			{
				Query:       "CREATE VIEW v WITH check_option = 'local' AS SELECT 1;",
				Skip: true,
			},
			{
				Query:       "CREATE VIEW v WITH security_barrier = true AS SELECT 1;",
				Skip: true,
			},
		},
	},
}
