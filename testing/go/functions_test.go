// Copyright 2023 Dolthub, Inc.
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

// https://www.postgresql.org/docs/15/functions-math.html
func TestFunctionsMath(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "cbrt",
			SetUpScript: []string{
				`CREATE TABLE test (pk INT primary key, v1 INT, v2 FLOAT4, v3 FLOAT8, v4 VARCHAR(255));`,
				`INSERT INTO test VALUES (1, -1, -2, -3, '-5'), (2, 7, 11, 13, '17'), (3, 19, -23, 29, '-31');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT cbrt(v1), cbrt(v2), cbrt(v3) FROM test ORDER BY pk;`,
					Skip:  true, // Our values are slightly different
					Expected: []sql.Row{
						{-1.0, -1.259921049894873, -1.4422495703074083},
						{1.9129311827723892, 2.2239800905693157, 2.3513346877207573},
						{2.668401648721945, -2.8438669798515654, 3.0723168256858475},
					},
				},
				{
					Query: `SELECT round(cbrt(v1)::numeric, 10), round(cbrt(v2)::numeric, 10), round(cbrt(v3)::numeric, 10) FROM test ORDER BY pk;`,
					Cols:  []string{"round", "round", "round"},
					Expected: []sql.Row{
						{-1.0000000000, -1.2599210499, -1.4422495703},
						{1.9129311828, 2.2239800906, 2.3513346877},
						{2.6684016487, -2.8438669799, 3.0723168257},
					},
				},
				{
					Query:       `SELECT cbrt(v4) FROM test ORDER BY pk;`,
					ExpectedErr: "function cbrt(varchar(255)) does not exist",
				},
				{
					Query: `SELECT cbrt('64');`,
					Cols:  []string{"cbrt"},
					Expected: []sql.Row{
						{4.0},
					},
				},
				{
					Query: `SELECT round(cbrt('64'));`,
					Cols:  []string{"round"},
					Expected: []sql.Row{
						{4.0},
					},
				},
			},
		},
		{
			Name: "gcd",
			SetUpScript: []string{
				`CREATE TABLE test (pk INT primary key, v1 INT4, v2 INT8, v3 FLOAT8, v4 VARCHAR(255));`,
				`INSERT INTO test VALUES (1, -2, -4, -6, '-8'), (2, 10, 12, 14.14, '16.16'), (3, 18, -20, 22.22, '-24.24');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT gcd(v1, 10), gcd(v2, 20) FROM test ORDER BY pk;`,
					Cols:  []string{"gcd", "gcd"},
					Expected: []sql.Row{
						{2, 4},
						{10, 4},
						{2, 20},
					},
				},
				{
					Query:       `SELECT gcd(v3, 10) FROM test ORDER BY pk;`,
					ExpectedErr: "function gcd(double precision, integer)",
				},
				{
					Query:       `SELECT gcd(v4, 10) FROM test ORDER BY pk;`,
					ExpectedErr: "function gcd(varchar(255), integer) does not exist",
				},
				{
					Query: `SELECT gcd(36, '48');`,
					Cols:  []string{"gcd"},
					Expected: []sql.Row{
						{12},
					},
				},
				{
					Query: `SELECT gcd('36', 48);`,
					Cols:  []string{"gcd"},
					Expected: []sql.Row{
						{12},
					},
				},
				{
					Query: `SELECT gcd(1, 0), gcd(0, 1), gcd(0, 0);`,
					Cols:  []string{"gcd", "gcd", "gcd"},
					Expected: []sql.Row{
						{1, 1, 0},
					},
				},
			},
		},
		{
			Name: "lcm",
			SetUpScript: []string{
				`CREATE TABLE test (pk INT primary key, v1 INT4, v2 INT8, v3 FLOAT8, v4 VARCHAR(255));`,
				`INSERT INTO test VALUES (1, -2, -4, -6, '-8'), (2, 10, 12, 14.14, '16.16'), (3, 18, -20, 22.22, '-24.24');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT lcm(v1, 10), lcm(v2, 20) FROM test ORDER BY pk;`,
					Cols:  []string{"lcm", "lcm"},
					Expected: []sql.Row{
						{10, 20},
						{10, 60},
						{90, 20},
					},
				},
				{
					Query:       `SELECT lcm(v3, 10) FROM test ORDER BY pk;`,
					ExpectedErr: "function lcm(double precision, integer)",
				},
				{
					Query:       `SELECT lcm(v4, 10) FROM test ORDER BY pk;`,
					ExpectedErr: "function lcm(varchar(255), integer) does not exist",
				},
				{
					Query: `SELECT lcm(36, '48');`,
					Expected: []sql.Row{
						{144},
					},
				},
				{
					Query: `SELECT lcm('36', 48);`,
					Expected: []sql.Row{
						{144},
					},
				},
				{
					Query: `SELECT lcm(1, 0), lcm(0, 1), lcm(0, 0);`,
					Expected: []sql.Row{
						{0, 0, 0},
					},
				},
			},
		},
	})
}

func TestFunctionsOID(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "to_regclass",
			SetUpScript: []string{
				`CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);`,
				`CREATE TABLE "Testing2" (pk INT primary key, v1 INT);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT to_regclass('testing');`,
					Cols:  []string{"to_regclass"},
					Expected: []sql.Row{
						{"testing"},
					},
				},
				{
					Query: `SELECT to_regclass('Testing2');`,
					Expected: []sql.Row{
						{nil},
					},
				},
				{
					Query: `SELECT to_regclass('"Testing2"');`,
					Expected: []sql.Row{
						{"Testing2"},
					},
				},
				{
					Query: `SELECT to_regclass(('testing'::regclass)::text);`,
					Expected: []sql.Row{
						{"testing"},
					},
				},
				{
					Query: `SELECT to_regclass((('testing'::regclass)::oid)::text);`,
					Expected: []sql.Row{
						{nil},
					},
				},
			},
		},
		{
			Name: "to_regproc",
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT to_regproc('acos');`,
					Cols:  []string{"to_regproc"},
					Expected: []sql.Row{
						{"acos"},
					},
				},
				{
					Query: `SELECT to_regproc('acos"');`,
					Expected: []sql.Row{
						{nil},
					},
				},
				{
					Query: `SELECT to_regproc(('acos'::regproc)::text);`,
					Expected: []sql.Row{
						{"acos"},
					},
				},
				{
					Query: `SELECT to_regproc((('acos'::regproc)::oid)::text);`,
					Expected: []sql.Row{
						{nil},
					},
				},
			},
		},
	})
}

func TestSystemInformationFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:     "current_database",
			Database: "test",
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT current_database();`,
					Cols:  []string{"current_database"},
					Expected: []sql.Row{
						{"test"},
					},
				},
				{
					Query:       `SELECT current_database;`,
					ExpectedErr: `column "current_database" could not be found in any table in scope`,
				},
				// TODO: Implement table function for current_database
				{
					Query: `SELECT * FROM current_database();`,
					Skip:  true,
					Expected: []sql.Row{
						{"test"},
					},
				},
				{
					Query:       `SELECT * FROM current_database;`,
					ExpectedErr: "table not found: current_database",
				},
			},
		},
		{
			Name:     "current_catalog",
			Database: "test",
			Assertions: []ScriptTestAssertion{
				{
					Skip:  true, // TODO: current_catalog currently returns current_database column name
					Query: `SELECT current_catalog;`,
					Cols:  []string{"current_catalog"},
					Expected: []sql.Row{
						{"test"},
					},
				},
				{
					Query: `SELECT current_catalog;`,
					Expected: []sql.Row{
						{"test"},
					},
				},
				{
					Query:       `SELECT current_catalog();`,
					ExpectedErr: `ERROR: at or near "(": syntax error (SQLSTATE XX000)`,
				},
				// // TODO: Implement table function for current_catalog
				{
					Query: `SELECT * FROM current_catalog;`,
					Skip:  true,
					Expected: []sql.Row{
						{"test"},
					},
				},
				{
					Query:       `SELECT * FROM current_catalog();`,
					ExpectedErr: `ERROR: at or near "(": syntax error (SQLSTATE XX000)`,
				},
			},
		},
		{
			Name: "current_schema",
			Assertions: []ScriptTestAssertion{
				{
					Skip:  true, // TODO: current_schema currently returns column name in quotes
					Query: `SELECT current_schema();`,
					Cols:  []string{"\"current_schema\""},
					Expected: []sql.Row{
						{"postgres"},
					},
				},
				{
					Query: `SELECT current_schema();`,
					Expected: []sql.Row{
						{"postgres"},
					},
				},
				{
					Query:    "CREATE SCHEMA test_schema;",
					Expected: []sql.Row{},
				},
				{
					Query:    `SET SEARCH_PATH TO test_schema;`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT current_schema();`,
					Expected: []sql.Row{
						{"test_schema"},
					},
				},
				{
					Query:    `SET SEARCH_PATH TO public;`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT current_schema();`,
					Expected: []sql.Row{
						{"public"},
					},
				},
				{
					Query: `SELECT current_schema;`,
					Expected: []sql.Row{
						{"public"},
					},
				},
				// TODO: Implement table function for current_schema
				{
					Query: `SELECT * FROM current_schema();`,
					Skip:  true,
					Expected: []sql.Row{
						{"public"},
					},
				},
				{
					Query: `SELECT * FROM current_schema;`,
					Skip:  true,
					Expected: []sql.Row{
						{"public"},
					},
				},
			},
		},
		{
			Name: "current_schemas",
			Assertions: []ScriptTestAssertion{
				{ // TODO: Not sure why Postgres does not display "$user", which is postgres here
					Query: `SELECT current_schemas(true);`,
					Cols:  []string{"current_schemas"},
					Expected: []sql.Row{
						{"{pg_catalog,postgres,public}"},
					},
				},
				{ // TODO: Not sure why Postgres does not display "$user" here
					Query: `SELECT current_schemas(false);`,
					Expected: []sql.Row{
						{"{postgres,public}"},
					},
				},
				{
					Query:    "CREATE SCHEMA test_schema;",
					Expected: []sql.Row{},
				},
				{
					Query:    `SET SEARCH_PATH TO test_schema;`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT current_schemas(true);`,
					Expected: []sql.Row{
						{"{pg_catalog,test_schema}"},
					},
				},
				{
					Query: `SELECT current_schemas(false);`,
					Expected: []sql.Row{
						{"{test_schema}"},
					},
				},
				{
					Query:    `SET SEARCH_PATH TO public, test_schema;`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT current_schemas(true);`,
					Expected: []sql.Row{
						{"{pg_catalog,public,test_schema}"},
					},
				},
				{
					Query: `SELECT current_schemas(false);`,
					Expected: []sql.Row{
						{"{public,test_schema}"},
					},
				},
			},
		},
		{
			Name: "version",
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT version();`,
					Cols:  []string{"version"},
					Expected: []sql.Row{
						{"PostgreSQL 15.5"},
					},
				},
			},
		},
	})
}
