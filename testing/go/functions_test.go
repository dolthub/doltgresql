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

func TestAggregateFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "bool_and",
			SetUpScript: []string{
				`CREATE TABLE t1 (pk INT primary key, v1 BOOLEAN, v2 BOOLEAN);`,
				`INSERT INTO t1 VALUES (1, true, false), (2, true, true), (3, true, true);`,
				`CREATE TABLE t2 (v1 BOOLEAN);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT bool_and(v1), bool_and(v2) FROM t1;`,
					Expected: []sql.Row{{"t", "f"}},
				},
				{
					Query: `SELECT bool_and(v1 and v2) FROM t1;`,
					Expected: []sql.Row{
						{"f"},
					},
				},
				{
					Query: `SELECT bool_and(v1 and v2) FROM t1 where v1 and v2;`,
					Expected: []sql.Row{
						{"t"},
					},
				},
				{
					Query: `SELECT bool_and(v1) FROM t1 where pk > 10;`,
					Expected: []sql.Row{
						{nil},
					},
				},
				{
					Skip:  true, // building a values-derived table's type fails here, postgres is more permissive
					Query: `SELECT bool_and(a) FROM (VALUES(true),(false),(null)) r(a);`,
					Expected: []sql.Row{
						{"f"},
					},
				},
				{
					Query: `SELECT bool_and(a) FROM (VALUES(true),(false),(null::bool)) r(a);`,
					Expected: []sql.Row{
						{"f"},
					},
				},
				{
					Query: `SELECT bool_and(a) FROM (VALUES(null::bool),(true),(null::bool)) r(a);`,
					Expected: []sql.Row{
						{"t"},
					},
				},
				{
					Query: `SELECT bool_and(v1) FROM t2`,
					Expected: []sql.Row{
						{nil},
					},
				},
			},
		},
		{
			Name: "bool_or",
			SetUpScript: []string{
				`CREATE TABLE t1 (pk INT primary key, v1 BOOLEAN, v2 BOOLEAN);`,
				`INSERT INTO t1 VALUES (1, false, false), (2, true, true), (3, true, false);`,
				`CREATE TABLE t2 (v1 BOOLEAN);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT bool_or(v1), bool_or(v2) FROM t1;`,
					Expected: []sql.Row{{"t", "t"}},
				},
				{
					Query:    `SELECT bool_or(v1), bool_or(v2) FROM t1 where pk <> 2;`,
					Expected: []sql.Row{{"t", "f"}},
				},
				{
					Query: `SELECT bool_or(v1 and v2) FROM t1;`,
					Expected: []sql.Row{
						{"t"},
					},
				},
				{
					Query: `SELECT bool_or(v1) FROM t1 where pk > 10;`,
					Expected: []sql.Row{
						{nil},
					},
				},
				{
					Query: `SELECT bool_or(a) FROM (VALUES(true),(false),(null::bool)) r(a);`,
					Expected: []sql.Row{
						{"t"},
					},
				},
				{
					Query: `SELECT bool_or(a) FROM (VALUES(null::bool),(false),(null::bool)) r(a);`,
					Expected: []sql.Row{
						{"f"},
					},
				},
				{
					Query: `SELECT bool_or(v1) FROM t2`,
					Expected: []sql.Row{
						{nil},
					},
				},
			},
		},
		{
			Name: "array_agg",
			SetUpScript: []string{
				`CREATE TABLE t1 (pk INT primary key, t timestamp, v varchar, f float[]);`,
				`INSERT INTO t1 VALUES 
                   (1, '2023-01-01 00:00:00', 'a', '{1.0, 2.0}'),
                   (2, '2023-01-02 00:00:00', 'b', '{3.0, 4.0}'),
                   (3, '2023-01-03 00:00:00', 'c', '{5.0, 6.0}');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT array_agg(pk) FROM t1;`,
					Expected: []sql.Row{
						{"{1,2,3}"},
					},
				},
				{
					Query: `SELECT array_agg(t) FROM t1;`,
					Expected: []sql.Row{
						{`{"2023-01-01 00:00:00","2023-01-02 00:00:00","2023-01-03 00:00:00"}`},
					},
				},
				{
					Query: `SELECT array_agg(v) FROM t1;`,
					Expected: []sql.Row{
						{"{a,b,c}"},
					},
				},
				{
					Skip:  true, // Higher-level arrays don't work because they panic during output
					Query: `SELECT array_agg(f) FROM t1;`,
					Expected: []sql.Row{
						{"{{1.0,2.0},{3.0,4.0},{5.0,6.0}}"},
					},
				},
			},
		},
		{
			Name: "array_agg with order by",
			SetUpScript: []string{
				`CREATE TABLE test_data (
					id INT PRIMARY KEY, 
					name VARCHAR(50), 
					age INT, 
					score FLOAT, 
					created_at TIMESTAMP, 
					category CHAR(1),
					nullable_field VARCHAR(20)
				);`,
				`INSERT INTO test_data VALUES 
					(1, 'Alice', 25, 95.5, '2023-01-03 10:00:00', 'A', 'value1'),
					(2, 'Bob', 30, 87.2, '2023-01-01 09:30:00', 'B', NULL),
					(3, 'Charlie', 22, 92.8, '2023-01-02 11:15:00', 'A', 'value2'),
					(4, 'Diana', 28, 88.9, '2023-01-04 08:45:00', 'C', NULL),
					(5, 'Eve', 35, 94.1, '2023-01-05 14:20:00', 'B', 'value3'),
					(6, 'Frank', 26, 89.3, '2023-01-06 16:30:00', 'A', 'value4');`,
			},
			Assertions: []ScriptTestAssertion{
				// Basic ORDER BY ASC
				{
					Query: `SELECT array_agg(name ORDER BY age ASC) FROM test_data;`,
					Expected: []sql.Row{
						{"{Charlie,Alice,Frank,Diana,Bob,Eve}"},
					},
				},
				// Basic ORDER BY DESC
				{
					Query: `SELECT array_agg(name ORDER BY age DESC) FROM test_data;`,
					Expected: []sql.Row{
						{"{Eve,Bob,Diana,Frank,Alice,Charlie}"},
					},
				},
				// ORDER BY with integers
				{
					Query: `SELECT array_agg(id ORDER BY age) FROM test_data;`,
					Expected: []sql.Row{
						{"{3,1,6,4,2,5}"},
					},
				},
				// ORDER BY with floats
				{
					Query: `SELECT array_agg(name ORDER BY score DESC) FROM test_data;`,
					Expected: []sql.Row{
						{"{Alice,Eve,Charlie,Frank,Diana,Bob}"},
					},
				},
				// ORDER BY with timestamps
				{
					Query: `SELECT array_agg(name ORDER BY created_at ASC) FROM test_data;`,
					Expected: []sql.Row{
						{"{Bob,Charlie,Alice,Diana,Eve,Frank}"},
					},
				},
				// ORDER BY with VARCHAR/CHAR
				{
					Query: `SELECT array_agg(age ORDER BY name) FROM test_data;`,
					Expected: []sql.Row{
						{"{25,30,22,28,35,26}"},
					},
				},
				// Multiple columns in ORDER BY
				{
					Query: `SELECT array_agg(name ORDER BY category ASC, age DESC) FROM test_data;`,
					Expected: []sql.Row{
						{"{Frank,Alice,Charlie,Eve,Bob,Diana}"},
					},
				},
				// ORDER BY with mixed ASC/DESC
				{
					Query: `SELECT array_agg(id ORDER BY category ASC, score DESC) FROM test_data;`,
					Expected: []sql.Row{
						{"{1,3,6,5,2,4}"},
					},
				},
				// ORDER BY with expression
				{
					Query: `SELECT array_agg(name ORDER BY age * 2) FROM test_data;`,
					Expected: []sql.Row{
						{"{Charlie,Alice,Frank,Diana,Bob,Eve}"},
					},
				},
				// ORDER BY with string concatenation
				{
					Query: `SELECT array_agg(age ORDER BY category || name) FROM test_data;`,
					Expected: []sql.Row{
						{"{25,22,26,30,35,28}"},
					},
				},
				// ORDER BY with CASE expression
				{
					Query: `SELECT array_agg(name ORDER BY CASE WHEN age > 27 THEN 1 ELSE 0 END, age) FROM test_data;`,
					Expected: []sql.Row{
						{"{Charlie,Alice,Frank,Diana,Bob,Eve}"},
					},
				},
				// ORDER BY with NULL values (NULLS FIRST behavior)
				{
					Query: `SELECT array_agg(name ORDER BY nullable_field) FROM test_data;`,
					Expected: []sql.Row{
						{"{Bob,Diana,Alice,Charlie,Eve,Frank}"},
					},
				},
				// ORDER BY with GROUP BY
				{
					Query: `SELECT category, array_agg(name ORDER BY age) FROM test_data GROUP BY category ORDER BY category;`,
					Expected: []sql.Row{
						{"A", "{Charlie,Alice,Frank}"},
						{"B", "{Bob,Eve}"},
						{"C", "{Diana}"},
					},
				},
				// ORDER BY with subquery correlation
				{
					Query: `SELECT category, array_agg(name ORDER BY (SELECT COUNT(*) FROM test_data t2 WHERE t2.category = test_data.category AND t2.age < test_data.age)) FROM test_data GROUP BY category ORDER BY category;`,
					Expected: []sql.Row{
						{"A", "{Charlie,Alice,Frank}"},
						{"B", "{Bob,Eve}"},
						{"C", "{Diana}"},
					},
				},
				// ORDER BY with COALESCE for NULL handling
				{
					Query: `SELECT array_agg(name ORDER BY COALESCE(nullable_field, 'zzz')) FROM test_data;`,
					Expected: []sql.Row{
						{"{Alice,Charlie,Eve,Frank,Bob,Diana}"},
					},
				},
				// Complex ORDER BY with multiple expressions
				{
					Query: `SELECT array_agg(name ORDER BY LENGTH(name) DESC, name ASC) FROM test_data;`,
					Expected: []sql.Row{
						{"{Charlie,Alice,Diana,Frank,Bob,Eve}"},
					},
				},
				// ORDER BY with aggregated values in grouped context
				{
					Query: `SELECT category, array_agg(name ORDER BY score - (SELECT AVG(score) FROM test_data t2 WHERE t2.category = test_data.category)) FROM test_data GROUP BY category ORDER BY category;`,
					Expected: []sql.Row{
						{"A", "{Frank,Charlie,Alice}"},
						{"B", "{Bob,Eve}"},
						{"C", "{Diana}"},
					},
				},
				// ORDER BY with date functions
				{
					Query: `SELECT array_agg(name ORDER BY EXTRACT(hour FROM created_at)) FROM test_data;`,
					Expected: []sql.Row{
						{"{Diana,Bob,Alice,Charlie,Eve,Frank}"},
					},
				},
				// Empty result set
				{
					Query: `SELECT array_agg(name ORDER BY age) FROM test_data WHERE age > 100;`,
					Expected: []sql.Row{
						{nil},
					},
				},
				// ORDER BY with boolean expression
				{
					Query: `SELECT array_agg(name ORDER BY age > 27, age) FROM test_data;`,
					Expected: []sql.Row{
						{"{Charlie,Alice,Frank,Diana,Bob,Eve}"},
					},
				},
			},
		},
		{
			Name: "array agg with case statement",
			SetUpScript: []string{
				"CREATE TABLE t1 (pk INT primary key, v1 INT, v2 INT);",
				"INSERT INTO t1 VALUES (1, 10, 20), (2, 30, 40), (3, 50, 60);",
				"CREATE TABLE t2 (pk INT primary key, v1 INT, v2 TEXT);",
				"INSERT INTO t2 VALUES (1, 10, 'a'), (2, 20, 'b'), (3, 30, 'c');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT array_agg(CASE WHEN v1 > 20 THEN v1 ELSE NULL END) FROM t1;`,
					Expected: []sql.Row{
						{"{NULL,30,50}"},
					},
				},
				{
					Query: `SELECT array_agg(CASE WHEN v1 >= 20 THEN v2 ELSE NULL END) FROM t2;`,
					Expected: []sql.Row{
						{"{NULL,b,c}"},
					},
				},
				{
					Query: `SELECT array_agg(CASE WHEN v1 > 20 THEN v1::text ELSE v2 END) FROM t2;`,
					Expected: []sql.Row{
						{"{a,b,30}"},
					},
				},
				{
					// Panic on type mixing, the logic for mixed types is hard-coded in GMS plan builder, needs
					// to be configurable. Postgres rejects this plan because of the type differences
					Skip:  true,
					Query: `SELECT array_agg(CASE WHEN v1 > 20 THEN v1 ELSE v2 END) FROM t2;`,
					Expected: []sql.Row{
						{"{a,b,30}"},
					},
				},
			},
		},
	})
}

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
					Query:            `SELECT round(cbrt(v1)::numeric, 10), round(cbrt(v2)::numeric, 10), round(cbrt(v3)::numeric, 10) FROM test ORDER BY pk;`,
					ExpectedColNames: []string{"round", "round", "round"},
					Expected: []sql.Row{
						{Numeric("-1.0000000000"), Numeric("-1.2599210499"), Numeric("-1.4422495703")},
						{Numeric("1.9129311828"), Numeric("2.2239800906"), Numeric("2.3513346877")},
						{Numeric("2.6684016487"), Numeric("-2.8438669799"), Numeric("3.0723168257")},
					},
				},
				{
					Query:       `SELECT cbrt(v4) FROM test ORDER BY pk;`,
					ExpectedErr: "function cbrt(varchar(255)) does not exist",
				},
				{
					Query:            `SELECT cbrt('64');`,
					ExpectedColNames: []string{"cbrt"},
					Expected: []sql.Row{
						{4.0},
					},
				},
				{
					Query:            `SELECT round(cbrt('64'));`,
					ExpectedColNames: []string{"round"},
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
					Query:            `SELECT gcd(v1, 10), gcd(v2, 20) FROM test ORDER BY pk;`,
					ExpectedColNames: []string{"gcd", "gcd"},
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
					Query:            `SELECT gcd(36, '48');`,
					ExpectedColNames: []string{"gcd"},
					Expected: []sql.Row{
						{12},
					},
				},
				{
					Query:            `SELECT gcd('36', 48);`,
					ExpectedColNames: []string{"gcd"},
					Expected: []sql.Row{
						{12},
					},
				},
				{
					Query:            `SELECT gcd(1, 0), gcd(0, 1), gcd(0, 0);`,
					ExpectedColNames: []string{"gcd", "gcd", "gcd"},
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
					Query:            `SELECT lcm(v1, 10), lcm(v2, 20) FROM test ORDER BY pk;`,
					ExpectedColNames: []string{"lcm", "lcm"},
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
					Query:            `SELECT to_regclass('testing');`,
					ExpectedColNames: []string{"to_regclass"},
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
				{
					// When the relation is from a schema on the search path, it is not qualified with the schema name
					Query: `SELECT to_regclass(('public.testing'::regclass)::text);`,
					Expected: []sql.Row{
						{"testing"},
					},
				},
				{
					// Clear out the current search_path setting to test fully qualified relation names
					Query:    `SET search_path = '';`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT to_regclass(('public.testing'::regclass)::text);`,
					Expected: []sql.Row{
						{"public.testing"},
					},
				},
			},
		},
		{
			Name: "to_regproc",
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT to_regproc('acos');`,
					ExpectedColNames: []string{"to_regproc"},
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
		{
			Name: "to_regtype",
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT to_regtype('integer');`,
					ExpectedColNames: []string{"to_regtype"},
					Expected: []sql.Row{
						{"integer"},
					},
				},
				{
					Query: `SELECT to_regtype('integer[]');`,
					Expected: []sql.Row{
						{"integer[]"},
					},
				},
				{
					Query: `SELECT to_regtype('int4');`,
					Expected: []sql.Row{
						{"integer"},
					},
				},
				{
					Query: `SELECT to_regtype('varchar');`,
					Expected: []sql.Row{
						{"character varying"},
					},
				},
				{
					Query: `SELECT to_regtype('pg_catalog.varchar');`,
					Expected: []sql.Row{
						{"character varying"},
					},
				},
				{
					Query: `SELECT to_regtype('varchar(10)');`,
					Expected: []sql.Row{
						{"character varying"},
					},
				},
				{
					Query: `SELECT to_regtype('char');`,
					Expected: []sql.Row{
						{"character"},
					},
				},
				{
					Query: `SELECT to_regtype('pg_catalog.char');`,
					Expected: []sql.Row{
						{`"char"`},
					},
				},
				{
					Query: `SELECT to_regtype('char(10)');`,
					Expected: []sql.Row{
						{"character"},
					},
				},
				{
					Query: `SELECT to_regtype('"char"');`,
					Expected: []sql.Row{
						{`"char"`},
					},
				},
				{
					Query: `SELECT to_regtype('pg_catalog."char"');`,
					Expected: []sql.Row{
						{`"char"`},
					},
				},
				{
					Query: `SELECT to_regtype('otherschema.char');`,
					Expected: []sql.Row{
						{nil},
					},
				},
				{
					Query: `SELECT to_regtype('timestamp');`,
					Expected: []sql.Row{
						{"timestamp without time zone"},
					},
				},
				{
					Query: `SELECT to_regtype('timestamp without time zone');`,
					Expected: []sql.Row{
						{"timestamp without time zone"},
					},
				},
				{
					Query: `SELECT to_regtype('integer"');`,
					Expected: []sql.Row{
						{nil},
					},
				},
				{
					Query: `SELECT to_regtype(('integer'::regtype)::text);`,
					Expected: []sql.Row{
						{"integer"},
					},
				},
				{
					Query: `SELECT to_regtype((('integer'::regtype)::oid)::text);`,
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
					Query:            `SELECT current_database();`,
					ExpectedColNames: []string{"current_database"},
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
					Skip:             true, // TODO: current_catalog currently returns current_database column name
					Query:            `SELECT current_catalog;`,
					ExpectedColNames: []string{"current_catalog"},
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
					Skip:             true, // TODO: current_schema currently returns column name in quotes
					Query:            `SELECT current_schema();`,
					ExpectedColNames: []string{"\"current_schema\""},
					Expected: []sql.Row{
						{"public"},
					},
				},
				{
					Query: `SELECT current_schema();`,
					Expected: []sql.Row{
						{"public"},
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
					Query:    `SET SEARCH_PATH TO public, test_schema;`,
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
				{
					Query:    `SET SEARCH_PATH TO test_schema, public;`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT current_schema();`,
					Expected: []sql.Row{
						{"test_schema"},
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
					Query:            `SELECT current_schemas(true);`,
					ExpectedColNames: []string{"current_schemas"},
					Expected: []sql.Row{
						{"{pg_catalog,public}"},
					},
				},
				{ // TODO: Not sure why Postgres does not display "$user" here
					Query: `SELECT current_schemas(false);`,
					Expected: []sql.Row{
						{"{public}"},
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
					Query:            `SELECT version();`,
					ExpectedColNames: []string{"version"},
					Expected: []sql.Row{
						{"PostgreSQL 15.5"},
					},
				},
			},
		},
		{
			Name: "col_description",
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT col_description(100, 1);`,
					ExpectedColNames: []string{"col_description"},
					Expected: []sql.Row{
						{""},
					},
				},
				{
					Query:       `SELECT col_description('not_a_table'::regclass, 1);`,
					ExpectedErr: `relation "not_a_table" does not exist`,
				},
				{
					Query:    `CREATE TABLE test_table (id INT);`,
					Expected: []sql.Row{},
				},
				{
					Skip:     true, // TODO: Implement column comments
					Query:    `COMMENT ON COLUMN test_table.id IS 'This is col id';`,
					Expected: []sql.Row{},
				},
				{
					Skip:             true, // TODO: Implement column object comments
					Query:            `SELECT col_description('test_table'::regclass, 1);`,
					ExpectedColNames: []string{"col_description"},
					Expected: []sql.Row{
						{"This is col id"},
					},
				},
			},
		},
		{
			Name: "obj_description",
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT obj_description(1003);`,
					ExpectedColNames: []string{"obj_description"},
					Expected: []sql.Row{
						{""},
					},
				},
				{
					Query:            `SELECT obj_description(100, 'pg_class');`,
					ExpectedColNames: []string{"obj_description"},
					Expected: []sql.Row{
						{""},
					},
				},
				{
					Query:       `SELECT obj_description('does-not-exist'::regproc, 'pg_class');`,
					ExpectedErr: `function "does-not-exist" does not exist`,
				},
				{
					Skip:             true, // TODO: Implement database object comments
					Query:            `SELECT obj_description('sinh'::regproc, 'pg_proc');`,
					ExpectedColNames: []string{"col_description"},
					Expected: []sql.Row{
						{"hyperbolic sine"},
					},
				},
			},
		},
		{
			Name: "shobj_description",
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT shobj_description(100, 'pg_class');`,
					ExpectedColNames: []string{"shobj_description"},
					Expected: []sql.Row{
						{""},
					},
				},
				{
					Query:       `SELECT shobj_description('does-not-exist'::regproc, 'pg_class');`,
					ExpectedErr: `function "does-not-exist" does not exist`,
				},
				{
					Skip:     true, // TODO: Implement tablespaces
					Query:    `CREATE TABLESPACE tblspc_2 LOCATION '/';`,
					Expected: []sql.Row{},
				},
				{
					Skip:     true, // TODO: Implement shared database object comments
					Query:    `COMMENT ON TABLESPACE tblspc_2 IS 'Store a few of the things';`,
					Expected: []sql.Row{},
				},
				{
					Skip: true, // TODO: Implement shared database object comments
					Query: `SELECT shobj_description(
                 (SELECT oid FROM pg_tablespace WHERE spcname = 'tblspc_2'),
                 'pg_tablespace');`,
					ExpectedColNames: []string{"shobj_description"},
					Expected: []sql.Row{
						{"Store a few of the things"},
					},
				},
			},
		},
		{
			Name: "format_type",
			Assertions: []ScriptTestAssertion{
				// Without typemod
				{
					Query:            `SELECT format_type('integer'::regtype, null);`,
					ExpectedColNames: []string{"format_type"},
					Expected: []sql.Row{
						{"integer"},
					},
				},
				{
					Query: `SELECT format_type('character varying'::regtype, null);`,
					Expected: []sql.Row{
						{"character varying"},
					},
				},
				{
					Query: `SELECT format_type('varchar'::regtype, null);`,
					Expected: []sql.Row{
						{"character varying"},
					},
				},
				{
					Query: `SELECT format_type('date'::regtype, null);`,
					Expected: []sql.Row{
						{"date"},
					},
				},
				{
					Query: `SELECT format_type('timestamptz'::regtype, null);`,
					Expected: []sql.Row{
						{"timestamp with time zone"},
					},
				},
				{
					Query: `SELECT format_type('bool'::regtype, null);`,
					Expected: []sql.Row{
						{"boolean"},
					},
				},
				{
					Query: `SELECT format_type(1007, null);`,
					Expected: []sql.Row{
						{"integer[]"},
					},
				},
				{
					Query: `SELECT format_type('"char"'::regtype, null);`,
					Expected: []sql.Row{
						{`"char"`},
					},
				},
				{
					Query: `SELECT format_type('"char"[]'::regtype, null);`,
					Expected: []sql.Row{
						{"\"char\"[]"},
					},
				},
				{
					Query: `SELECT format_type(1002, null);`,
					Expected: []sql.Row{
						{"\"char\"[]"},
					},
				},
				{
					Query: `SELECT format_type('real[]'::regtype, null);`,
					Expected: []sql.Row{
						{"real[]"},
					},
				},
				// With typemod
				{
					Query: `SELECT format_type('character varying'::regtype, 100);`,
					Expected: []sql.Row{
						{"character varying(96)"},
					},
				},
				{
					Query: `SELECT format_type('text'::regtype, 0);`,
					Expected: []sql.Row{
						{"text(0)"},
					},
				},
				{
					Query: `SELECT format_type('text'::regtype, 4);`,
					Expected: []sql.Row{
						{"text(4)"},
					},
				},
				{
					Query: `SELECT format_type('text'::regtype, -1);`,
					Expected: []sql.Row{
						{"text"},
					},
				},
				{
					Query: `SELECT format_type('name'::regtype, 0);`,
					Expected: []sql.Row{
						{"name(0)"},
					},
				},
				{
					Query: `SELECT format_type('bpchar'::regtype, -1);`,
					Expected: []sql.Row{
						{"bpchar"},
					},
				},
				{
					Query: `SELECT format_type('bpchar'::regtype, 10);`,
					Expected: []sql.Row{
						{"character(6)"},
					},
				},
				{
					Query: `SELECT format_type('bpchar'::regtype, 10);`,
					Expected: []sql.Row{
						{"character(6)"},
					},
				},
				{
					Query: `SELECT format_type('character'::regtype, 4);`,
					Expected: []sql.Row{
						{"character"},
					},
				},
				{
					Query: `SELECT format_type('varchar'::regtype, 0);`,
					Expected: []sql.Row{
						{"character varying"},
					},
				},
				{
					Query: `SELECT format_type('"char"'::regtype, 0);`,
					Expected: []sql.Row{
						{"\"char\"(0)"},
					},
				},
				{
					Query: `SELECT format_type('numeric'::regtype, 12);`,
					Expected: []sql.Row{
						{"numeric(0,8)"},
					},
				},
				// OID does not exist
				{
					Query:            `SELECT format_type(874938247, 20);`,
					ExpectedColNames: []string{"format_type"},
					Expected: []sql.Row{
						{"???"},
					},
				},
				{
					Query:            `SELECT format_type(874938247, null);`,
					ExpectedColNames: []string{"format_type"},
					Expected: []sql.Row{
						{"???"},
					},
				},
			},
		},
		{
			Name: "pg_get_constraintdef",
			SetUpScript: []string{
				`CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);`,
				`CREATE TABLE testing2 (pk INT primary key, pktesting INT REFERENCES testing(pk), v1 TEXT);`,
				`CREATE TABLE testing3 (pk1 INT, pk2 INT, PRIMARY KEY (pk1, pk2));`,
				// TODO: Uncomment when check constraints supported
				// `ALTER TABLE testing2 ADD CONSTRAINT v1_check CHECK (v1 != '');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT pg_get_constraintdef(845743985);`,
					ExpectedColNames: []string{"pg_get_constraintdef"},
					Expected:         []sql.Row{{""}},
				},
				{
					Query:            `SELECT pg_get_constraintdef(oid) FROM pg_catalog.pg_constraint WHERE conrelid='testing'::regclass;`,
					ExpectedColNames: []string{"pg_get_constraintdef"},
					Expected: []sql.Row{
						{"PRIMARY KEY (pk)"},
						{"UNIQUE (v1)"},
					},
				},
				{
					Query: `SELECT pg_get_constraintdef(oid) FROM pg_catalog.pg_constraint WHERE conrelid='testing2'::regclass;`,
					Expected: []sql.Row{
						{"FOREIGN KEY (pktesting) REFERENCES testing(pk)"},
						{"PRIMARY KEY (pk)"},
					},
				},
				{
					Query: `SELECT pg_get_constraintdef(oid) FROM pg_catalog.pg_constraint WHERE conrelid='testing3'::regclass;`,
					Expected: []sql.Row{
						{"PRIMARY KEY (pk1, pk2)"},
					},
				},
				{
					Query: `SELECT pg_get_constraintdef(oid, true) FROM pg_catalog.pg_constraint WHERE conrelid='testing3'::regclass;`,
					Expected: []sql.Row{
						{"PRIMARY KEY (pk1, pk2)"},
					},
				},
				{
					Query: `SELECT pg_get_constraintdef(oid, false) FROM pg_catalog.pg_constraint WHERE conrelid='testing3'::regclass;`,
					Expected: []sql.Row{
						{"PRIMARY KEY (pk1, pk2)"},
					},
				},
			},
		},
		{
			Name: "pg_get_ruledef",
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT pg_get_ruledef(845743985);`,
					ExpectedColNames: []string{"pg_get_ruledef"},
					Expected:         []sql.Row{{nil}},
				},
				{
					Query:            `SELECT pg_get_ruledef(845743985, true);`,
					ExpectedColNames: []string{"pg_get_ruledef"},
					Expected:         []sql.Row{{nil}},
				},
				{
					Query:            `SELECT pg_get_ruledef(845743985, false);`,
					ExpectedColNames: []string{"pg_get_ruledef"},
					Expected:         []sql.Row{{nil}},
				},
			},
		},
		{
			Name: "pg_get_expr",
			SetUpScript: []string{
				`CREATE TABLE testing (id INT primary key);`,
				`CREATE TABLE temperature (celsius SMALLINT NOT NULL, fahrenheit SMALLINT NOT NULL GENERATED ALWAYS AS ((celsius * 9/5) + 32) STORED);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Skip:             true, // TODO: pg_attrdef.adbin not implemented
					Query:            `SELECT pg_get_expr(adbin, adrelid) FROM pg_catalog.pg_attrdef WHERE adrelid = 'temperature'::regclass;`,
					ExpectedColNames: []string{"pg_get_expr"},
					Expected:         []sql.Row{{"(celsius * 9 / 5 + 32)"}},
				},
				{
					Query:            `SELECT indexrelid, pg_get_expr(indpred, indrelid) FROM pg_catalog.pg_index WHERE indrelid='testing'::regclass;`,
					ExpectedColNames: []string{"indexrelid", "pg_get_expr"},
					Expected:         []sql.Row{{3757635986, nil}},
				},
				{
					Query:    `SELECT indexrelid, pg_get_expr(indpred, indrelid, true) FROM pg_catalog.pg_index WHERE indrelid='testing'::regclass;`,
					Expected: []sql.Row{{3757635986, nil}},
				},
				{
					Query:    `SELECT indexrelid, pg_get_expr(indpred, indrelid, NULL) FROM pg_catalog.pg_index WHERE indrelid='testing'::regclass;`,
					Expected: []sql.Row{{3757635986, nil}},
				},
			},
		},
		{
			Name: "pg_get_serial_sequence",
			SetUpScript: []string{
				`create table t0 (id INTEGER NOT NULL PRIMARY KEY);`,
				`create table t1 (id SERIAL PRIMARY KEY);`,
				`create sequence t2_id_seq START 1 INCREMENT 3;`,
				`create table t2 (id INTEGER NOT NULL DEFAULT nextval('t2_id_seq'));`,
				// TODO: ALTER SEQUENCE OWNED BY is not supported yet. When the sequence is created
				//       explicitly, separate from the column, the owner must be updated before
				//       pg_get_serial_sequence() will identify it.
				// `ALTER SEQUENCE t2_id_seq OWNED BY t2.id;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT pg_get_serial_sequence('doesnotexist.t1', 'id');`,
					ExpectedErr: "does not exist",
				},
				{
					Query:       `SELECT pg_get_serial_sequence('doesnotexist', 'id');`,
					ExpectedErr: "does not exist",
				},
				{
					Query:       `SELECT pg_get_serial_sequence('t0', 'doesnotexist');`,
					ExpectedErr: "does not exist",
				},
				{
					// No sequence for column returns null
					Query:            `SELECT pg_get_serial_sequence('t0', 'id');`,
					ExpectedColNames: []string{"pg_get_serial_sequence"},
					Expected:         []sql.Row{{nil}},
				},
				{
					Query:            `SELECT pg_get_serial_sequence('public.t1', 'id');`,
					ExpectedColNames: []string{"pg_get_serial_sequence"},
					Expected:         []sql.Row{{"public.t1_id_seq"}},
				},
				{
					// Test with no schema specified
					Query:            `SELECT pg_get_serial_sequence('t1', 'id');`,
					ExpectedColNames: []string{"pg_get_serial_sequence"},
					Expected:         []sql.Row{{"public.t1_id_seq"}},
				},
				{
					// TODO: This test shouldn't pass until we're able to use
					//       ALTER SEQUENCE OWNED BY to set the owning column.
					Skip:             true,
					Query:            `SELECT pg_get_serial_sequence('t2', 'id');`,
					ExpectedColNames: []string{"pg_get_serial_sequence"},
					Expected:         []sql.Row{{"public.t2_id_seq"}},
				},
			},
		},
		{
			Name: "current_setting function",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SET timezone TO '+00:00';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT current_setting('timezone')",
					Expected: []sql.Row{{"+00:00"}},
				},
				{
					Query:       "SELECT current_setting('wrong_input')",
					ExpectedErr: `unrecognized configuration parameter "wrong_input"`,
				},
				{
					Query:    "SELECT current_setting('wrong_input', true)",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:       "SELECT current_setting('wrong_input', false)",
					ExpectedErr: `unrecognized configuration parameter "wrong_input"`,
				},
			},
		},
	})
}

func TestJsonFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "json_build_array",
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT json_build_array(1, 2, 3);`,
					ExpectedColNames: []string{"json_build_array"},
					Expected:         []sql.Row{{`[1,2,3]`}},
				},
				{
					Query:            `SELECT json_build_array(1, '2', 3);`,
					ExpectedColNames: []string{"json_build_array"},
					Expected:         []sql.Row{{`[1,"2",3]`}},
				},
				{
					Query:            `SELECT json_build_array();`,
					Skip:             true, // variadic functions can't handle 0 arguments right now
					ExpectedColNames: []string{"json_build_array"},
					Expected:         []sql.Row{{`[]`}},
				},
			},
		},
		{
			Name: "json_build_object",
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT json_build_object('a', 2, 'b', 4);`,
					ExpectedColNames: []string{"json_build_object"},
					Expected:         []sql.Row{{`{"a":2,"b":4}`}},
				},
				{
					Query:       `SELECT json_build_object('a', 2, 'b');`,
					ExpectedErr: "even number",
				},
				{
					Query:            `SELECT json_build_object(1, 2, 'b', 3);`,
					ExpectedColNames: []string{"json_build_object"},
					Expected:         []sql.Row{{`{"1":2,"b":3}`}},
				},
			},
		},
		{
			Name: "jsonb_build_array",
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT jsonb_build_array(1, 2, 3);`,
					ExpectedColNames: []string{"jsonb_build_array"},
					Expected:         []sql.Row{{`[1, 2, 3]`}},
				},
				{
					Query:            `SELECT jsonb_build_array(1, '2', 3);`,
					ExpectedColNames: []string{"jsonb_build_array"},
					Expected:         []sql.Row{{`[1, "2", 3]`}},
				},
				{
					Query:            `SELECT jsonb_build_array();`,
					Skip:             true, // variadic functions can't handle 0 arguments right now
					ExpectedColNames: []string{"jsonb_build_array"},
					Expected:         []sql.Row{{`[]`}},
				},
			},
		},
		{
			Name: "jsonb_build_object",
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT jsonb_build_object('a', 2, 'b', 4);`,
					ExpectedColNames: []string{"jsonb_build_object"},
					Expected:         []sql.Row{{`{"a": 2, "b": 4}`}},
				},
				{
					Query:       `SELECT jsonb_build_object('a', 2, 'b');`,
					ExpectedErr: "even number",
				},
				{
					Query:            `SELECT jsonb_build_object(1, 2, 'b', 3);`,
					ExpectedColNames: []string{"jsonb_build_object"},
					Expected:         []sql.Row{{`{"1": 2, "b": 3}`}},
				},
			},
		},
	})
}

func TestArrayFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "unnest",
			SetUpScript: []string{
				`CREATE TABLE testing (id INT primary key, val1 smallint[]);`,
				`INSERT INTO testing VALUES (1, '{}'), (2, '{1}'), (3, '{1, 2}');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT unnest(val1) FROM testing WHERE id=1;`,
					ExpectedColNames: []string{"unnest"},
					Expected:         []sql.Row{},
				},
				{
					Query:            `SELECT unnest(val1) FROM testing WHERE id=2;`,
					ExpectedColNames: []string{"unnest"},
					Expected:         []sql.Row{{1}},
				},
				{
					Query:    `SELECT unnest(val1) FROM testing WHERE id=3;`,
					Expected: []sql.Row{{1}, {2}},
				},
				{
					// TODO: change in Eval in file table_function.go in gms breaks regression test???
					Skip:     true,
					Query:    `select * from unnest(array[1,2,3]);`,
					Expected: []sql.Row{{1}, {2}, {3}},
				},
			},
		},
		{
			Name:        "array_to_string",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT array_to_string(ARRAY[1, 2, 3, NULL, 5], ',', '*')`,
					Expected: []sql.Row{{"1,2,3,*,5"}},
				},
				{
					Query:    `SELECT array_to_string(ARRAY[1, 2, 3, NULL, 5], ',')`,
					Expected: []sql.Row{{"1,2,3,5"}},
				},
				{
					Query:    `SELECT array_to_string(ARRAY[37.89, 1.2], '_');`,
					Expected: []sql.Row{{"37.89_1.2"}},
				},
				{
					Skip:     true, // TODO: we currently return "37_1"
					Query:    `SELECT array_to_string(ARRAY[37.89::int4, 1.2::int4], '_');`,
					Expected: []sql.Row{{"38_1"}},
				},
			},
		},
		{
			Name: "array_upper",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `select array_upper(ARRAY[1,2,3,4], 1);`,
					Expected: []sql.Row{{4}},
				},
				{
					Skip:     true, // TODO: multi-dimensional is not supported yet
					Query:    `select array_upper(ARRAY[1,2,3,4], 2);`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
		{
			Name: "array_cat",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `select array_cat(ARRAY[1,2,3], ARRAY[4,5]);`,
					Expected: []sql.Row{{"{1,2,3,4,5}"}},
				},
				{
					Query:    `select array_cat(NULL, ARRAY[4,5]);`,
					Expected: []sql.Row{{"{4,5}"}},
				},
				{
					Query:    `select array_cat(ARRAY[1,2,3], NULL);`,
					Expected: []sql.Row{{"{1,2,3}"}},
				},
				{
					Query:    `select array_cat(NULL, NULL);`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
		{
			Name: "array_length",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `select array_length(ARRAY[1,2,3,4,5], 1);`,
					Expected: []sql.Row{{5}},
				},
			},
		},
		{
			Name: "array_position and array_positions",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT array_position(ARRAY[1,2,3,4,5], 4);`,
					Expected: []sql.Row{{4}},
				},
				{
					Query:    `SELECT array_position(ARRAY[1,4,2,3,4,5,4], 4, 3);`,
					Expected: []sql.Row{{5}},
				},
				{
					Query:    `select array_position(NULL, 1);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `select array_position(ARRAY[1,4,2,3,4,5,4], NULL);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `select array_position(NULL, NULL);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `SELECT array_positions(ARRAY[1,2,3,4,5,6,1,2,3,4,5,6], 4);`,
					Expected: []sql.Row{{"{4,10}"}},
				},
				{
					Query:    `select array_positions(NULL, 1);`,
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    `select array_positions(ARRAY[1,4,2,3,4,5,4], NULL);`,
					Expected: []sql.Row{{"{}"}},
				},
				{
					Query:    `select array_positions(NULL, NULL);`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
	})
}

func TestSchemaVisibilityInquiryFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Skip:        true, // TODO: not supported
			Name:        "pg_function_is_visible",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT pg_function_is_visible(1342177280);`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT pg_function_is_visible(22);`, // invalid
					Expected: []sql.Row{{"f"}},
				},
			},
		},
		{
			Name: "pg_table_is_visible",
			SetUpScript: []string{
				"CREATE SCHEMA myschema;",
				"SET search_path TO myschema;",
				"CREATE TABLE mytable (id int, name text);",
				"INSERT INTO mytable VALUES (1,'desk'), (2,'chair');",
				"CREATE VIEW myview AS SELECT name FROM mytable;",
				"CREATE SCHEMA testschema;",
				"SET search_path TO testschema;",
				`CREATE TABLE test_table (pk INT primary key, v1 INT UNIQUE);`,
				"INSERT INTO test_table VALUES (1,5), (2,7);",
				"CREATE INDEX test_index ON test_table(v1);",
				"CREATE SEQUENCE test_seq START 39;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT c.oid, c.relname AS table_name, n.nspname AS table_schema FROM pg_catalog.pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE (n.nspname='myschema' OR n.nspname='testschema') AND left(relname, 5) <> 'dolt_' order by relname;`,
					Expected: []sql.Row{
						{3905781870, "mytable", "myschema"},
						{3983475213, "myview", "myschema"},
						{3057657334, "test_index", "testschema"},
						{1539973141, "test_seq", "testschema"},
						{1952237395, "test_table", "testschema"},
						{3508950454, "test_table_pkey", "testschema"},
						{521883837, "v1", "testschema"},
					},
				},
				{
					Query:    `SHOW search_path;`,
					Expected: []sql.Row{{"testschema"}},
				},
				{
					Query:    `select pg_table_is_visible(3057657334);`, // index from testschema
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `select pg_table_is_visible(1952237395);`, // table from testschema
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `select pg_table_is_visible(1539973141);`, // sequence from testschema
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `select pg_table_is_visible(3983475213);`, // view from myschema
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SET search_path = 'myschema';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SHOW search_path;`,
					Expected: []sql.Row{{"myschema"}},
				},
				{
					Query:    `select pg_table_is_visible(3983475213);`, // view from myschema
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `select pg_table_is_visible(3905781870);`, // table from myschema
					Expected: []sql.Row{{"t"}},
				},
			},
		},
		{
			Name: "pg_type_is_visible",
			SetUpScript: []string{
				"CREATE SCHEMA myschema;",
				"SET search_path TO myschema;",
				"CREATE DOMAIN mydomain AS text;",
				"CREATE TYPE myenum AS ENUM ('a', 'b', 'c');",
				"CREATE SCHEMA testschema;",
				"SET search_path TO testschema;",
				"CREATE DOMAIN test_domain AS int;",
				"CREATE TYPE test_enum AS ENUM ('x', 'y', 'z');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT t.oid, t.typname, n.nspname FROM pg_catalog.pg_type t JOIN pg_namespace n ON n.oid = t.typnamespace WHERE n.nspname='myschema' OR n.nspname='testschema' ORDER BY t.typname;`,
					Expected: []sql.Row{
						{4183151504, "_mydomain", "myschema"},
						{2179370508, "_myenum", "myschema"},
						{2560281972, "_test_domain", "testschema"},
						{2798112218, "_test_enum", "testschema"},
						{340132571, "mydomain", "myschema"},
						{1684884017, "myenum", "myschema"},
						{2272253470, "test_domain", "testschema"},
						{1117094145, "test_enum", "testschema"},
					},
				},
				{
					Query:    `SHOW search_path;`,
					Expected: []sql.Row{{"testschema"}},
				},
				{
					Query:    `SELECT pg_type_is_visible(2272253470);`, // test_domain from testschema
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT pg_type_is_visible(1117094145);`, // test_enum from testschema
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT pg_type_is_visible(340132571);`, // mydomain from myschema
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT pg_type_is_visible(1684884017);`, // myenum from myschema
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SET search_path = 'myschema';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SHOW search_path;`,
					Expected: []sql.Row{{"myschema"}},
				},
				{
					Query:    `SELECT pg_type_is_visible(340132571);`, // mydomain from myschema
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT pg_type_is_visible(1684884017);`, // myenum from myschema
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT pg_type_is_visible(2272253470);`, // test_domain from testschema
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT pg_type_is_visible(1117094145);`, // test_enum from testschema
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SELECT pg_type_is_visible(999999);`, // non-existent type oid
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:    `SET search_path = 'pg_catalog';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT pg_type_is_visible('text'::regtype::oid);`, // built-in type should be visible in pg_catalog
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT pg_type_is_visible('int4'::regtype::oid);`, // built-in type should be visible in pg_catalog
					Expected: []sql.Row{{"t"}},
				},
			},
		},
	})
}

func TestSystemCatalogInformationFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:        "pg_encoding_to_char",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT pg_encoding_to_char(encoding) FROM pg_database WHERE datname = 'postgres';`,
					Expected: []sql.Row{
						{"UTF8"},
					},
				},
			},
		},
		{
			Name:        "pg_get_functiondef",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					// TODO: not supported yet
					Query: `SELECT pg_get_functiondef(22)`,
					Expected: []sql.Row{
						{""},
					},
				},
			},
		},
		{
			Name:        "pg_get_function_result",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					// TODO: not supported yet
					Query: `SELECT pg_get_function_result(22)`,
					Expected: []sql.Row{
						{""},
					},
				},
			},
		},
		{
			Name:        "pg_get_triggerdef",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					// TODO: triggers are not supported yet
					Query: `SELECT pg_get_triggerdef(22)`,
					Expected: []sql.Row{
						{""},
					},
				},
			},
		},
		{
			Name:        "pg_get_userbyid",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					// TODO: users and roles are not supported yet
					Query: `SELECT pg_get_userbyid(22)`,
					Expected: []sql.Row{
						{"postgres"},
					},
				},
			},
		},
		{
			Name: "pg_get_viewdef",
			SetUpScript: []string{
				"CREATE TABLE test (id int, name text)",
				"INSERT INTO test VALUES (1,'desk'), (2,'chair')",
				"CREATE VIEW test_view AS SELECT name FROM test",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT c.oid, c.relname AS table_name, n.nspname AS table_schema FROM pg_catalog.pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace WHERE (n.nspname='myschema' OR n.nspname='public') and left(relname, 5) <> 'dolt_';`,
					Expected: []sql.Row{
						{2707638987, "test_view", "public"},
						{1397286223, "test", "public"},
					},
				},
				{
					Query:    `select pg_get_viewdef(2707638987);`,
					Expected: []sql.Row{{"SELECT name FROM test"}},
				},
			},
		},
	})
}

func TestDateAndTimeFunction(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:        "extract from date",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT EXTRACT(CENTURY FROM DATE '2022-02-02');`,
					Expected: []sql.Row{{Numeric("21")}},
				},
				{
					Query:    `SELECT EXTRACT(CENTURY FROM DATE '0002-12-31 BC');`,
					Expected: []sql.Row{{Numeric("-1")}},
				},
				{
					Query:    `SELECT EXTRACT(DAY FROM DATE '2022-02-02');`,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT EXTRACT(DECADE FROM DATE '2022-02-02');`,
					Expected: []sql.Row{{Numeric("202")}},
				},
				{
					Query:    `SELECT EXTRACT(DOW FROM DATE '2022-02-02');`,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT EXTRACT(DOY FROM DATE '2022-02-02');`,
					Expected: []sql.Row{{Numeric("33")}},
				},
				{
					Query:    `SELECT EXTRACT(EPOCH FROM DATE '2022-02-02');`,
					Expected: []sql.Row{{Numeric("1643760000")}},
				},
				{
					Query:       `SELECT EXTRACT(HOUR FROM DATE '2022-02-02');`,
					ExpectedErr: `unit "hour" not supported for type date`,
				},
				{
					Query:    `SELECT EXTRACT(ISODOW FROM DATE '2022-02-02');`,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT EXTRACT(ISOYEAR FROM DATE '2006-01-01');`,
					Expected: []sql.Row{{Numeric("2005")}},
				},
				{
					Query:    `SELECT EXTRACT(ISOYEAR FROM DATE '2006-01-02');`,
					Expected: []sql.Row{{Numeric("2006")}},
				},
				{
					Skip:     true, // TODO: not supported yet
					Query:    `SELECT extract(julian from date '2021-06-23');`,
					Expected: []sql.Row{{Numeric("2459389")}},
				},
				{
					Query:       `SELECT EXTRACT(MICROSECONDS FROM DATE '2022-02-02');`,
					ExpectedErr: `unit "microseconds" not supported for type date`,
				},
				{
					Query:    `SELECT EXTRACT(MILLENNIUM FROM DATE '2022-02-02');`,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:       `SELECT EXTRACT(MILLISECONDS FROM DATE '2022-02-02');`,
					ExpectedErr: `unit "milliseconds" not supported for type date`,
				},
				{
					Query:       `SELECT EXTRACT(MINUTE FROM DATE '2022-02-02');`,
					ExpectedErr: `unit "minute" not supported for type date`,
				},
				{
					Query:    `SELECT EXTRACT(MONTH FROM DATE '2022-02-02');`,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT EXTRACT(QUARTER FROM DATE '2022-02-02');`,
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:       `SELECT EXTRACT(SECOND FROM DATE '2022-02-02');`,
					ExpectedErr: `unit "second" not supported for type date`,
				},
				{
					Query:       `SELECT EXTRACT(TIMEZONE FROM DATE '2022-02-02');`,
					ExpectedErr: `unit "timezone" not supported for type date`,
				},
				{
					Query:       `SELECT EXTRACT(TIMEZONE_HOUR FROM DATE '2022-02-02');`,
					ExpectedErr: `unit "timezone_hour" not supported for type date`,
				},
				{
					Query:       `SELECT EXTRACT(TIMEZONE_MINUTE FROM DATE '2022-02-02');`,
					ExpectedErr: `unit "timezone_minute" not supported for type date`,
				},
				{
					Query:    `SELECT EXTRACT(WEEK FROM DATE '2022-02-02');`,
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    `SELECT EXTRACT(YEAR FROM DATE '2022-02-02');`,
					Expected: []sql.Row{{Numeric("2022")}},
				},
			},
		},
		{
			Name:        "extract from time without time zone",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT EXTRACT(CENTURY FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "century" not supported for type time without time zone`,
				},
				{
					Query:       `SELECT EXTRACT(DAY FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "day" not supported for type time without time zone`,
				},
				{
					Query:       `SELECT EXTRACT(DECADE FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "decade" not supported for type time without time zone`,
				},
				{
					Query:       `SELECT EXTRACT(DOW FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "dow" not supported for type time without time zone`,
				},
				{
					Query:       `SELECT EXTRACT(DOY FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "doy" not supported for type time without time zone`,
				},
				{
					Query:    `SELECT EXTRACT(EPOCH FROM TIME '17:12:28.5');`,
					Expected: []sql.Row{{Numeric("61948.500000")}},
				},
				{
					Query:    `SELECT EXTRACT(HOUR FROM TIME '17:12:28.5');`,
					Expected: []sql.Row{{Numeric("17")}},
				},
				{
					Query:       `SELECT EXTRACT(ISODOW FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "isodow" not supported for type time without time zone`,
				},
				{
					Query:       `SELECT EXTRACT(ISOYEAR FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "isoyear" not supported for type time without time zone`,
				},
				{
					Query:       `SELECT EXTRACT(JULIAN FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "julian" not supported for type time without time zone`,
				},
				{
					Query:    `SELECT EXTRACT(MICROSECONDS FROM TIME '17:12:28.5');`,
					Expected: []sql.Row{{Numeric("28500000")}},
				},
				{
					Query:       `SELECT EXTRACT(MILLENNIUM FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "millennium" not supported for type time without time zone`,
				},
				{
					Query:    `SELECT EXTRACT(MILLISECONDS FROM TIME '17:12:28.5');`,
					Expected: []sql.Row{{Numeric("28500.000")}},
				},
				{
					Query:    `SELECT EXTRACT(MINUTE FROM TIME '17:12:28.5');`,
					Expected: []sql.Row{{Numeric("12")}},
				},
				{
					Query:       `SELECT EXTRACT(MONTH FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "month" not supported for type time without time zone`,
				},
				{
					Query:       `SELECT EXTRACT(QUARTER FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "quarter" not supported for type time without time zone`,
				},
				{
					Query:    `SELECT EXTRACT(SECOND FROM TIME '17:12:28.5');`,
					Expected: []sql.Row{{Numeric("28.500000")}},
				},
				{
					Query:       `SELECT EXTRACT(TIMEZONE FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "timezone" not supported for type time without time zone`,
				},
				{
					Query:       `SELECT EXTRACT(TIMEZONE_HOUR FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "timezone_hour" not supported for type time without time zone`,
				},
				{
					Query:       `SELECT EXTRACT(TIMEZONE_MINUTE FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "timezone_minute" not supported for type time without time zone`,
				},
				{
					Query:       `SELECT EXTRACT(WEEK FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "week" not supported for type time without time zone`,
				},
				{
					Query:       `SELECT EXTRACT(YEAR FROM TIME '17:12:28.5');`,
					ExpectedErr: `unit "year" not supported for type time without time zone`,
				},
			},
		},
		{
			Name:        "extract from time with time zone",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT EXTRACT(CENTURY FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					ExpectedErr: `unit "century" not supported for type time`,
				},
				{
					Query:       `SELECT EXTRACT(DAY FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					ExpectedErr: `unit "day" not supported for type time`,
				},
				{
					Query:       `SELECT EXTRACT(DECADE FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					ExpectedErr: `unit "decade" not supported for type time`,
				},
				{
					Query:       `SELECT EXTRACT(DOW FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					ExpectedErr: `unit "dow" not supported for type time`,
				},
				{
					Query:       `SELECT EXTRACT(DOY FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					ExpectedErr: `unit "doy" not supported for type time`,
				},
				{
					Query:    `SELECT EXTRACT(EPOCH FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					Expected: []sql.Row{{Numeric("72748.500000")}},
				},
				{
					Query:    `SELECT EXTRACT(HOUR FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					Expected: []sql.Row{{Numeric("17")}},
				},
				{
					Query:       `SELECT EXTRACT(ISODOW FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					ExpectedErr: `unit "isodow" not supported for type time`,
				},
				{
					Query:       `SELECT EXTRACT(ISOYEAR FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					ExpectedErr: `unit "isoyear" not supported for type time`,
				},
				{
					Query:       `SELECT EXTRACT(JULIAN FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					ExpectedErr: `unit "julian" not supported for type time`,
				},
				{
					Query:    `SELECT EXTRACT(MICROSECONDS FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					Expected: []sql.Row{{Numeric("28500000")}},
				},
				{
					Query:       `SELECT EXTRACT(MILLENNIUM FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					ExpectedErr: `unit "millennium" not supported for type time`,
				},
				{
					Query:    `SELECT EXTRACT(MILLISECONDS FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					Expected: []sql.Row{{Numeric("28500.000")}},
				},
				{
					Query:    `SELECT EXTRACT(MINUTE FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					Expected: []sql.Row{{Numeric("12")}},
				},
				{
					Query:       `SELECT EXTRACT(MONTH FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					ExpectedErr: `unit "month" not supported for type time`,
				},
				{
					Query:       `SELECT EXTRACT(QUARTER FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					ExpectedErr: `unit "quarter" not supported for type time`,
				},
				{
					Query:    `SELECT EXTRACT(SECOND FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					Expected: []sql.Row{{Numeric("28.500000")}},
				},
				{
					Query:    `SELECT EXTRACT(TIMEZONE FROM TIME WITH TIME ZONE '17:12:28.5+03');`,
					Expected: []sql.Row{{Numeric("10800")}},
				},
				{
					Query:    `SELECT EXTRACT(TIMEZONE_HOUR FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					Expected: []sql.Row{{Numeric("-3")}},
				},
				{
					Query:    `SELECT EXTRACT(TIMEZONE_MINUTE FROM TIME WITH TIME ZONE '17:12:28.5-03:45');`,
					Expected: []sql.Row{{Numeric("-45")}},
				},
				{
					Query:       `SELECT EXTRACT(WEEK FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					ExpectedErr: `unit "week" not supported for type time`,
				},
				{
					Query:       `SELECT EXTRACT(YEAR FROM TIME WITH TIME ZONE '17:12:28.5-03');`,
					ExpectedErr: `unit "year" not supported for type time`,
				},
			},
		},
		{
			Name:        "extract from timestamp without time zone",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT EXTRACT(CENTURY FROM TIMESTAMP '2000-12-16 12:21:13');`,
					Expected: []sql.Row{{Numeric("20")}},
				},
				{
					Query:    `SELECT EXTRACT(CENTURY FROM TIMESTAMP '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("21")}},
				},
				{
					Query:    `SELECT EXTRACT(DAY FROM TIMESTAMP '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("16")}},
				},
				{
					Query:    `SELECT EXTRACT(DECADE FROM TIMESTAMP '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("200")}},
				},
				{
					Query:    `SELECT EXTRACT(DOW FROM TIMESTAMP '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    `SELECT EXTRACT(DOY FROM TIMESTAMP '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("47")}},
				},
				{
					Query:    `SELECT EXTRACT(EPOCH FROM TIMESTAMP '2001-02-16 20:38:40.12');`,
					Expected: []sql.Row{{Numeric("982355920.120000")}},
				},
				{
					Query:    `SELECT EXTRACT(HOUR FROM TIMESTAMP '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("20")}},
				},
				{
					Query:    `SELECT EXTRACT(ISODOW FROM TIMESTAMP '2001-02-18 20:38:40');`,
					Expected: []sql.Row{{Numeric("7")}},
				},
				{
					Query:    `SELECT EXTRACT(ISOYEAR FROM TIMESTAMP '2001-02-18 20:38:40');`,
					Expected: []sql.Row{{Numeric("2001")}},
				},
				{
					Skip:     true, // TODO: not supported yet
					Query:    `SELECT EXTRACT(JULIAN FROM TIMESTAMP '2001-02-18 20:38:40');`,
					Expected: []sql.Row{{Numeric("2451959.86018518518518518519")}},
				},
				{
					Query:    `SELECT EXTRACT(MICROSECONDS FROM TIMESTAMP '2001-02-18 20:38:40');`,
					Expected: []sql.Row{{Numeric("40000000")}},
				},
				{
					Query:    `SELECT EXTRACT(MILLENNIUM FROM TIMESTAMP '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT EXTRACT(MILLENNIUM FROM TIMESTAMP '2000-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT EXTRACT(MILLISECONDS FROM TIMESTAMP '2000-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("40000.000")}},
				},
				{
					Query:    `SELECT EXTRACT(MINUTE FROM TIMESTAMP '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("38")}},
				},
				{
					Query:    `SELECT EXTRACT(MONTH FROM TIMESTAMP '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT EXTRACT(QUARTER FROM TIMESTAMP '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    `SELECT EXTRACT(SECOND FROM TIMESTAMP '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("40.000000")}},
				},
				{
					Query:       `SELECT EXTRACT(TIMEZONE FROM TIMESTAMP '2001-02-16 20:38:40');`,
					ExpectedErr: `unit "timezone" not supported for type timestamp without time zone`,
				},
				{
					Query:       `SELECT EXTRACT(TIMEZONE_HOUR FROM TIMESTAMP '2001-02-16 20:38:40');`,
					ExpectedErr: `unit "timezone_hour" not supported for type timestamp without time zone`,
				},
				{
					Query:       `SELECT EXTRACT(TIMEZONE_MINUTE FROM TIMESTAMP '2001-02-16 20:38:40');`,
					ExpectedErr: `unit "timezone_minute" not supported for type timestamp without time zone`,
				},
				{
					Query:    `SELECT EXTRACT(WEEK FROM TIMESTAMP '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("7")}},
				},
				{
					Query:    `SELECT EXTRACT(YEAR FROM TIMESTAMP '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{Numeric("2001")}},
				},
			},
		},
		{
			// The TIMESTAMPTZ value gets converted to Local timezone / server timezone,
			// so set the server timezone to UTC. GitHub CI runs on UTC time zone.
			Name:        "extract from timestamp with time zone",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SET TIMEZONE TO 'UTC';`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT EXTRACT(CENTURY FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("21")}},
				},
				{
					Query:    `SELECT EXTRACT(DAY FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("16")}},
				},
				{
					Query:    `SELECT EXTRACT(DECADE FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("200")}},
				},
				{
					Query:    `SELECT EXTRACT(DOW FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    `SELECT EXTRACT(DOY FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("47")}},
				},
				{
					Query:    `SELECT EXTRACT(EPOCH FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("982345120.120000")}},
				},
				{
					Query:    `SELECT EXTRACT(HOUR FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("17")}},
				},
				{
					Query:    `SELECT EXTRACT(ISODOW FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    `SELECT EXTRACT(ISOYEAR FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("2001")}},
				},
				{
					Skip:     true, // TODO: not supported yet
					Query:    `SELECT EXTRACT(JULIAN FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("2451957.73518657407407407407")}},
				},
				{
					Skip:     true, // TODO: not supported yet
					Query:    `SELECT extract(julian from '2021-06-23 7:00:00-04'::timestamptz at time zone 'UTC+12');`,
					Expected: []sql.Row{{Numeric("2459388.95833333333333333333")}},
				},
				{
					Skip:     true, // TODO: not supported yet
					Query:    `SELECT extract(julian from '2021-06-23 8:00:00-04'::timestamptz at time zone 'UTC+12');`,
					Expected: []sql.Row{{Numeric("2459389.0000000000000000000000000000")}},
				},
				{
					Query:    `SELECT EXTRACT(MICROSECONDS FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("40120000")}},
				},
				{
					Query:    `SELECT EXTRACT(MILLENNIUM FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT EXTRACT(MILLISECONDS FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("40120.000")}},
				},
				{
					Query:    `SELECT EXTRACT(MINUTE FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("38")}},
				},
				{
					Query:    `SELECT EXTRACT(MONTH FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `SELECT EXTRACT(QUARTER FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    `SELECT EXTRACT(SECOND FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("40.120000")}},
				},
				{
					Query:    `SELECT EXTRACT(TIMEZONE FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("-28800")}},
				},
				{
					Query:    `SELECT EXTRACT(TIMEZONE_HOUR FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("-8")}},
				},
				{
					Query:    `SELECT EXTRACT(TIMEZONE_MINUTE FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05:45');`,
					Expected: []sql.Row{{Numeric("0")}},
				},
				{
					Query:    `SELECT EXTRACT(WEEK FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("7")}},
				},
				{
					Query:    `SELECT EXTRACT(YEAR FROM TIMESTAMP WITH TIME ZONE '2001-02-16 12:38:40.12-05');`,
					Expected: []sql.Row{{Numeric("2001")}},
				},
				{
					Query:    `SET TIMEZONE TO DEFAULT;`,
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name:        "extract from interval",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT EXTRACT(CENTURY FROM INTERVAL '2001 years');`,
					Expected: []sql.Row{{Numeric("20")}},
				},
				{
					Query:    `SELECT EXTRACT(DAY FROM INTERVAL '40 days 1 minute');`,
					Expected: []sql.Row{{Numeric("40")}},
				},
				{
					Query:    `select extract(decades from interval '1000 months');`,
					Expected: []sql.Row{{Numeric("8")}},
				},
				{
					Query:    `SELECT EXTRACT(EPOCH FROM INTERVAL '5 days 3 hours');`,
					Expected: []sql.Row{{Numeric("442800.000000")}},
				},
				{
					Query:    `select extract(epoch from interval '10 months 10 seconds');`,
					Expected: []sql.Row{{Numeric("25920010.000000")}},
				},
				{
					Query:    `select extract(hours from interval '10 months 65 minutes 10 seconds');`,
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    `select extract(microsecond from interval '10 months 65 minutes 10 seconds');`,
					Expected: []sql.Row{{Numeric("10000000")}},
				},
				{
					Query:    `SELECT EXTRACT(MILLENNIUM FROM INTERVAL '2001 years');`,
					Expected: []sql.Row{{Numeric("2")}},
				},
				{
					Query:    `select extract(millenniums from interval '3000 years 65 minutes 10 seconds');`,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `select extract(millisecond from interval '10 months 65 minutes 10 seconds');`,
					Expected: []sql.Row{{Numeric("10000.000")}},
				},
				{
					Query:    `select extract(minutes from interval '10 months 65 minutes 10 seconds');`,
					Expected: []sql.Row{{Numeric("5")}},
				},
				{
					Query:    `SELECT EXTRACT(MONTH FROM INTERVAL '2 years 3 months');`,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `SELECT EXTRACT(MONTH FROM INTERVAL '2 years 13 months');`,
					Expected: []sql.Row{{Numeric("1")}},
				},
				{
					Query:    `select extract(months from interval '20 months 65 minutes 10 seconds');`,
					Expected: []sql.Row{{Numeric("8")}},
				},
				{
					Query:    `select extract(quarter from interval '20 months 65 minutes 10 seconds');`,
					Expected: []sql.Row{{Numeric("3")}},
				},
				{
					Query:    `select extract(seconds from interval '65 minutes 10 seconds 5 millisecond');`,
					Expected: []sql.Row{{Numeric("10.005000")}},
				},
				{
					Query:    `select extract(years from interval '20 months 65 minutes 10 seconds');`,
					Expected: []sql.Row{{Numeric("1")}},
				},
			},
		},
		{
			Name:        "age",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT age(timestamp '2001-04-10', timestamp '1957-06-13');`,
					Expected: []sql.Row{{"43 years 9 mons 27 days"}},
				},
				{
					Query:    `SELECT age(timestamp '1957-06-13', timestamp '2001-04-10');`,
					Expected: []sql.Row{{"-43 years -9 mons -27 days"}},
				},
				{
					Query:    `SELECT age(timestamp '2001-06-13', timestamp '2001-04-10');`,
					Expected: []sql.Row{{"2 mons 3 days"}},
				},
				{
					Query:    `SELECT age(timestamp '2001-04-10', timestamp '2001-06-13');`,
					Expected: []sql.Row{{"-2 mons -3 days"}},
				},
				{
					Query:    `SELECT age(timestamp '2001-04-10 12:23:33', timestamp '1957-06-13 13:23:34.4');`,
					Expected: []sql.Row{{"43 years 9 mons 26 days 22:59:58.6"}},
				},
				{
					Query:    `SELECT age(timestamp '1957-06-13 13:23:34.4', timestamp '2001-04-10 12:23:33');`,
					Expected: []sql.Row{{"-43 years -9 mons -26 days -22:59:58.6"}},
				},
				{
					Skip:     true, // TODO: current_date should return timestamp, not text
					Query:    `SELECT age(current_date);`,
					Expected: []sql.Row{{"00:00:00"}},
				},
				{
					Query:    `SELECT age(current_date::timestamp);`,
					Expected: []sql.Row{{"00:00:00"}},
				},
			},
		},
		{
			Name:        "timezone",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `select timezone(interval '2 minutes', timestamp with time zone '2001-02-16 20:38:40.12-05');`,
					Expected: []sql.Row{{"2001-02-17 01:40:40.12"}},
				},
				{
					Query:    `select timezone('UTC', timestamp with time zone '2001-02-16 20:38:40.12-05');`,
					Expected: []sql.Row{{"2001-02-17 01:38:40.12"}},
				},
				{
					Query:    `select timezone('-04:45', time with time zone '20:38:40.12-05');`,
					Expected: []sql.Row{{"06:23:40.12+04:45"}},
				},
				{
					Query:    `select timezone(interval '2 hours 2 minutes', time with time zone '20:38:40.12-05');`,
					Expected: []sql.Row{{"03:40:40.12+02:02"}},
				},
				{
					Query:    `select timezone('-04:45', timestamp '2001-02-16 20:38:40.12');`,
					Expected: []sql.Row{{"2001-02-16 07:53:40.12-08"}},
				},
				{
					Query:    `select timezone('-04:45:44', timestamp '2001-02-16 20:38:40.12');`,
					Expected: []sql.Row{{"2001-02-16 07:52:56.12-08"}},
				},
				{
					Query:    `select timezone(interval '2 hours 2 minutes', timestamp '2001-02-16 20:38:40.12');`,
					Expected: []sql.Row{{"2001-02-16 10:36:40.12-08"}},
				},
				{
					Query:    `select '2024-08-22 14:47:57 -07' at time zone 'utc';`,
					Expected: []sql.Row{{"2024-08-22 21:47:57"}},
				},
				{
					Query:    `select round(extract(epoch from '2024-08-22 13:47:57-07' at time zone 'UTC')) as startup_time;`,
					Expected: []sql.Row{{Numeric("1724359677")}},
				},
				{
					Query:    `select timestamptz '2024-08-22 13:47:57-07' at time zone 'utc';`,
					Expected: []sql.Row{{"2024-08-22 20:47:57"}},
				},
				{
					Query:    `select timestamp '2024-08-22 13:47:57-07';`,
					Expected: []sql.Row{{"2024-08-22 13:47:57"}},
				},
				{
					Query:    `select timestamp '2024-08-22 13:47:57-07' at time zone 'utc';`,
					Expected: []sql.Row{{"2024-08-22 06:47:57-07"}},
				},
			},
		},
		{
			Name: "date_part",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `select date_part('month', date '2001-02-16');`,
					Expected: []sql.Row{{float64(2)}},
				},
				{
					Query:    `select date_part('minute', time without time zone '20:38:40');`,
					Expected: []sql.Row{{float64(38)}},
				},
				{
					Query:    `select date_part('second', time with time zone '20:38:40 UTC');`,
					Expected: []sql.Row{{float64(40)}},
				},
				{
					Query:    `select date_part('year', timestamp without time zone '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{float64(2001)}},
				},
				{
					Query:    `select date_part('day', timestamp with time zone '2001-02-16 20:38:40 UTC');`,
					Expected: []sql.Row{{float64(16)}},
				},
				{
					Query:    `select date_part('month', interval '2 years 3 months');`,
					Expected: []sql.Row{{float64(3)}},
				},
			},
		},
		{
			Name: "date_trunc",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `select date_trunc('hour', timestamp '2001-02-16 20:38:40');`,
					Expected: []sql.Row{{"2001-02-16 20:00:00"}},
				},
				{
					Query: `SET timezone to '+06:30';`,
				},
				{
					Query:    `select date_trunc('day', timestamp with time zone '2001-02-16 20:38:40 UTC');`,
					Expected: []sql.Row{{"2001-02-16 00:00:00-06:30"}},
				},
				{
					Query:    `select date_trunc('day', timestamp with time zone '2001-02-16 20:38:40 UTC', '-07:00');`,
					Expected: []sql.Row{{"2001-02-16 10:30:00-06:30"}},
				},
				{
					Query: `SET timezone to '+06:30';`,
				},
				{
					Query:    `select date_trunc('hour', interval '2 days 10 hours 30 minutes');`,
					Expected: []sql.Row{{"2 days 10:00:00"}},
				},
			},
		},
		{
			Name: "to_date",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT to_date('1 4 1902', 'Q MM YYYY');`,
					Expected: []sql.Row{{"1902-04-01"}},
				},
				{
					Query:    `SELECT to_date('3 4 21 01', 'W MM CC YY');`,
					Expected: []sql.Row{{"2001-04-15"}},
				},
				{
					Query:    `SELECT to_date('2458872', 'J');`,
					Expected: []sql.Row{{"2020-01-23"}},
				},
				{
					Query:    `SELECT to_date('44-02-01 BC','YYYY-MM-DD BC');`,
					Expected: []sql.Row{{"0044-02-01 BC"}},
				},
				{
					Query:    `SELECT to_date('-44-02-01','YYYY-MM-DD');`,
					Expected: []sql.Row{{"0044-02-01 BC"}},
				},
				{
					Query:    `SELECT to_date('2011x 12x 18', 'YYYYxMMxDD');`,
					Expected: []sql.Row{{"2011-12-18"}},
				},
				{
					Query:    `SELECT to_date('2015 365', 'YYYY DDD');`,
					Expected: []sql.Row{{"2015-12-31"}},
				},
			},
		},
		{
			Name: "to_timestamp",
			Assertions: []ScriptTestAssertion{
				{
					Query: `SET timezone to '+06:30';`,
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18 23:38:15', 'YYYY-MM-DD HH24:MI:SS');`,
					Expected: []sql.Row{{"2011-12-18 23:38:15-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2000-01-01 12:30:45', 'YYYY-MM-DD HH24:MI:SS');`,
					Expected: []sql.Row{{"2000-01-01 12:30:45-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('0097/Feb/16 --> 08:14:30', 'YYYY/Mon/DD --> HH:MI:SS');`,
					Expected: []sql.Row{{"0097-02-16 08:14:30-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('97/2/16 8:14:30', 'FMYYYY/FMMM/FMDD FMHH:FMMI:FMSS');`,
					Expected: []sql.Row{{"0097-02-16 08:14:30-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011$03!18 23_38_15', 'YYYY-MM-DD HH24:MI:SS');`,
					Expected: []sql.Row{{"2011-03-18 23:38:15-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('1985 January 12', 'YYYY FMMonth DD');`,
					Expected: []sql.Row{{"1985-01-12 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('1985 FMMonth 12', 'YYYY "FMMonth" DD');`,
					Expected: []sql.Row{{"1985-01-12 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('1985 \\ 12', 'YYYY \\\\ DD');`,
					Expected: []sql.Row{{"1985-01-12 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('My birthday-> Year: 1976, Month: May, Day: 16', '"My birthday-> Year:" YYYY, "Month:" FMMonth, "Day:" DD');`,
					Expected: []sql.Row{{"1976-05-16 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('1,582nd VIII 21', 'Y,YYYth FMRM DD');`,
					Expected: []sql.Row{{"1582-08-21 00:00:00-06:30"}},
				},
				{
					Query: `SELECT to_timestamp('15 "text between quote marks" 98 54 45',
				  E'HH24 "\\"text between quote marks\\"" YY MI SS');`,
					Expected: []sql.Row{{"1998-01-01 15:54:45-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('05121445482000', 'MMDDHH24MISSYYYY');`,
					Expected: []sql.Row{{"2000-05-12 14:45:48-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2000January09Sunday', 'YYYYFMMonthDDFMDay');`,
					Expected: []sql.Row{{"2000-01-09 00:00:00-06:30"}},
				},
				{
					Query:       `SELECT to_timestamp('97/Feb/16', 'YYMonDD');`,
					ExpectedErr: "invalid value",
				},
				{
					Query:    `SELECT to_timestamp('97/Feb/16', 'YY:Mon:DD');`,
					Expected: []sql.Row{{"1997-02-16 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('97/Feb/16', 'FXYY:Mon:DD');`,
					Expected: []sql.Row{{"1997-02-16 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('97/Feb/16', 'FXYY/Mon/DD');`,
					Expected: []sql.Row{{"1997-02-16 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('19971116', 'YYYYMMDD');`,
					Expected: []sql.Row{{"1997-11-16 00:00:00-06:30"}},
				},
				{
					// TODO: this test passes but time library parsing does not allow year length to be more than 4
					//  the using pgx library for tests relies on it.
					// https://github.com/jackc/pgx/blob/master/pgtype/timestamptz.go#L312
					Skip:     true,
					Query:    `SELECT to_timestamp('20000-1116', 'FXYYYY-MMDD');`,
					Expected: []sql.Row{{"20000-11-16 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('1997 AD 11 16', 'YYYY BC MM DD');`,
					Expected: []sql.Row{{"1997-11-16 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('1997 BC 11 16', 'YYYY BC MM DD');`,
					Expected: []sql.Row{{"1997-11-16 00:00:00-06:30 BC"}},
				},
				{
					Query:    `SELECT to_timestamp('1997 A.D. 11 16', 'YYYY B.C. MM DD');`,
					Expected: []sql.Row{{"1997-11-16 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('1997 B.C. 11 16', 'YYYY B.C. MM DD');`,
					Expected: []sql.Row{{"1997-11-16 00:00:00-06:30 BC"}},
				},
				{
					Query:    `SELECT to_timestamp('9-1116', 'Y-MMDD');`,
					Expected: []sql.Row{{"2009-11-16 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('95-1116', 'YY-MMDD');`,
					Expected: []sql.Row{{"1995-11-16 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('995-1116', 'YYY-MMDD');`,
					Expected: []sql.Row{{"1995-11-16 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2005426', 'YYYYWWD');`,
					Expected: []sql.Row{{"2005-10-15 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2005300', 'YYYYDDD');`,
					Expected: []sql.Row{{"2005-10-27 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2005527', 'IYYYIWID');`,
					Expected: []sql.Row{{"2006-01-01 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('005527', 'IYYIWID');`,
					Expected: []sql.Row{{"2006-01-01 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('05527', 'IYIWID');`,
					Expected: []sql.Row{{"2006-01-01 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('5527', 'IIWID');`,
					Expected: []sql.Row{{"2006-01-01 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2005364', 'IYYYIDDD');`,
					Expected: []sql.Row{{"2006-01-01 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('20050302', 'YYYYMMDD');`,
					Expected: []sql.Row{{"2005-03-02 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2005 03 02', 'YYYYMMDD');`,
					Expected: []sql.Row{{"2005-03-02 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp(' 2005 03 02', 'YYYYMMDD');`,
					Expected: []sql.Row{{"2005-03-02 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('  20050302', 'YYYYMMDD');`,
					Expected: []sql.Row{{"2005-03-02 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18 11:38 AM', 'YYYY-MM-DD HH12:MI PM');`,
					Expected: []sql.Row{{"2011-12-18 11:38:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18 11:38 PM', 'YYYY-MM-DD HH12:MI PM');`,
					Expected: []sql.Row{{"2011-12-18 23:38:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18 11:38 A.M.', 'YYYY-MM-DD HH12:MI P.M.');`,
					Expected: []sql.Row{{"2011-12-18 11:38:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18 11:38 P.M.', 'YYYY-MM-DD HH12:MI P.M.');`,
					Expected: []sql.Row{{"2011-12-18 23:38:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18 11:38 +05', 'YYYY-MM-DD HH12:MI TZH');`,
					Expected: []sql.Row{{"2011-12-18 00:08:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18 11:38 -05', 'YYYY-MM-DD HH12:MI TZH');`,
					Expected: []sql.Row{{"2011-12-18 10:08:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18 11:38 +05:20', 'YYYY-MM-DD HH12:MI TZH:TZM');`,
					Expected: []sql.Row{{"2011-12-17 23:48:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18 11:38 -05:20', 'YYYY-MM-DD HH12:MI TZH:TZM');`,
					Expected: []sql.Row{{"2011-12-18 10:28:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18 11:38 20', 'YYYY-MM-DD HH12:MI TZM');`,
					Expected: []sql.Row{{"2011-12-18 04:48:00-06:30"}},
				},
				{
					Skip:        true, // TODO: support formatting TZ
					Query:       `SELECT to_timestamp('2011-12-18 11:38 PST', 'YYYY-MM-DD HH12:MI TZ');`,
					ExpectedErr: `formatting field "TZ" is only supported in to_char`,
				},
				{
					Query:    `SELECT to_timestamp('2018-11-02 12:34:56.025', 'YYYY-MM-DD HH24:MI:SS.MS');`,
					Expected: []sql.Row{{"2018-11-02 12:34:56.025-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('44-02-01 11:12:13 BC','YYYY-MM-DD HH24:MI:SS BC');`,
					Expected: []sql.Row{{"0044-02-01 11:12:13-06:30 BC"}},
				},
				{
					Query:    `SELECT to_timestamp('-44-02-01 11:12:13','YYYY-MM-DD HH24:MI:SS');`,
					Expected: []sql.Row{{"0044-02-01 11:12:13-06:30 BC"}},
				},
				{
					Query:    `SELECT to_timestamp('-44-02-01 11:12:13 BC','YYYY-MM-DD HH24:MI:SS BC');`,
					Expected: []sql.Row{{"0044-02-01 11:12:13-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18 23:38:15', 'YYYY-MM-DD  HH24:MI:SS');`,
					Expected: []sql.Row{{"2011-12-18 23:38:15-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18  23:38:15', 'YYYY-MM-DD  HH24:MI:SS');`,
					Expected: []sql.Row{{"2011-12-18 23:38:15-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18   23:38:15', 'YYYY-MM-DD  HH24:MI:SS');`,
					Expected: []sql.Row{{"2011-12-18 23:38:15-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18  23:38:15', 'YYYY-MM-DD HH24:MI:SS');`,
					Expected: []sql.Row{{"2011-12-18 23:38:15-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18  23:38:15', 'YYYY-MM-DD  HH24:MI:SS');`,
					Expected: []sql.Row{{"2011-12-18 23:38:15-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2011-12-18  23:38:15', 'YYYY-MM-DD   HH24:MI:SS');`,
					Expected: []sql.Row{{"2011-12-18 23:38:15-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2000+   JUN', 'YYYY/MON');`,
					Expected: []sql.Row{{"2000-06-01 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('  2000 +JUN', 'YYYY/MON');`,
					Expected: []sql.Row{{"2000-06-01 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp(' 2000 +JUN', 'YYYY//MON');`,
					Expected: []sql.Row{{"2000-06-01 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2000  +JUN', 'YYYY//MON');`,
					Expected: []sql.Row{{"2000-06-01 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2000 + JUN', 'YYYY MON');`,
					Expected: []sql.Row{{"2000-06-01 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2000 ++ JUN', 'YYYY  MON');`,
					Expected: []sql.Row{{"2000-06-01 00:00:00-06:30"}},
				},
				{
					Query:       `SELECT to_timestamp('2000 + - JUN', 'YYYY  MON');`,
					ExpectedErr: `invalid value "-" for "MON"`,
				},
				{
					Query:    `SELECT to_timestamp('2000 + + JUN', 'YYYY   MON');`,
					Expected: []sql.Row{{"2000-06-01 00:00:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2000 -10', 'YYYY TZH');`,
					Expected: []sql.Row{{"2000-01-01 03:30:00-06:30"}},
				},
				{
					Query:    `SELECT to_timestamp('2000 -10', 'YYYY  TZH');`,
					Expected: []sql.Row{{"1999-12-31 07:30:00-06:30"}},
				},
				{
					Query:       `SELECT to_timestamp('2005527', 'YYYYIWID');`,
					ExpectedErr: "invalid combination of date conventions",
				},
				{
					Query:       `SELECT to_timestamp('19971', 'YYYYMMDD');`,
					ExpectedErr: `source string too short for "MM" formatting field`,
				},
				{
					Query:       `SELECT to_timestamp('19971)24', 'YYYYMMDD');`,
					ExpectedErr: `invalid value "1)" for "MM"`,
				},
				{
					Query:       `SELECT to_timestamp('Friday 1-January-1999', 'DY DD MON YYYY');`,
					ExpectedErr: `invalid value "da" for "DD"`,
				},
				{
					Query:       `SELECT to_timestamp('Fri 1-January-1999', 'DY DD MON YYYY');`,
					ExpectedErr: `invalid value "uary" for "YYYY"`,
				},
				{
					Query:    `SELECT to_timestamp('Fri 1-Jan-1999', 'DY DD MON YYYY');`,
					Expected: []sql.Row{{"1999-01-01 00:00:00-06:30"}},
				},
				{
					Query:       `SELECT to_timestamp('1997-11-Jan-16', 'YYYY-MM-Mon-DD');`,
					ExpectedErr: `conflicting values for "Mon" field in formatting string`,
				},
				{
					Query:       `SELECT to_timestamp('199711xy', 'YYYYMMDD');`,
					ExpectedErr: `invalid value "xy" for "DD"`,
				},
				{
					Query:       `SELECT to_timestamp('10000000000', 'FMYYYY');`,
					ExpectedErr: `value for "YYYY" in source string is out of range`,
				},
				{
					Query:       `SELECT to_timestamp('2016-06-13 25:00:00', 'YYYY-MM-DD HH24:MI:SS');`,
					ExpectedErr: `date/time field value out of range`,
				},
				{
					Query:       `SELECT to_timestamp('2016-06-13 15:60:00', 'YYYY-MM-DD HH24:MI:SS');`,
					ExpectedErr: `date/time field value out of range`,
				},
				{
					Query:       `SELECT to_timestamp('2016-06-13 15:50:60', 'YYYY-MM-DD HH24:MI:SS');`,
					ExpectedErr: `date/time field value out of range`,
				},
				{
					Query:    `SELECT to_timestamp('2016-06-13 15:50:55', 'YYYY-MM-DD HH24:MI:SS');`,
					Expected: []sql.Row{{"2016-06-13 15:50:55-06:30"}},
				},
				{
					Query:       `SELECT to_timestamp('2016-06-13 15:50:55', 'YYYY-MM-DD HH:MI:SS');`,
					ExpectedErr: `hour "15" is invalid for the 12-hour clock`,
				},
				{
					Query:       `SELECT to_timestamp('2016-13-01 15:50:55', 'YYYY-MM-DD HH24:MI:SS');`,
					ExpectedErr: `date/time field value out of range`,
				},
				{
					Query:       `SELECT to_timestamp('2016-02-30 15:50:55', 'YYYY-MM-DD HH24:MI:SS');`,
					ExpectedErr: `date/time field value out of range`,
				},
				{
					Query:    `SELECT to_timestamp('2016-02-29 15:50:55', 'YYYY-MM-DD HH24:MI:SS');`,
					Expected: []sql.Row{{"2016-02-29 15:50:55-06:30"}},
				},
				{
					Query:       `SELECT to_timestamp('2015-02-29 15:50:55', 'YYYY-MM-DD HH24:MI:SS');`,
					ExpectedErr: `date/time field value out of range`,
				},
				{
					Query:    `SELECT to_timestamp('2015-02-11 86000', 'YYYY-MM-DD SSSS');`,
					Expected: []sql.Row{{"2015-02-11 23:53:20-06:30"}},
				},
				{
					Query:       `SELECT to_timestamp('2015-02-11 86400', 'YYYY-MM-DD SSSS');`,
					ExpectedErr: `date/time field value out of range`,
				},
				{
					Query:    `SELECT to_timestamp('2015-02-11 86000', 'YYYY-MM-DD SSSSS');`,
					Expected: []sql.Row{{"2015-02-11 23:53:20-06:30"}},
				},
				{
					Query:       `SELECT to_timestamp('2015-02-11 86400', 'YYYY-MM-DD SSSSS');`,
					ExpectedErr: `date/time field value out of range`,
				},
				{
					Query: `SET timezone to '+06:30';`,
				},
			},
		},
	})
}

func TestStringFunction(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:        "use name type for text type input",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT ascii('name'::name)`,
					Expected: []sql.Row{{110}},
				},
				{
					Query:    "SELECT bit_length('name'::name);",
					Expected: []sql.Row{{32}},
				},
				{
					Query:    "SELECT btrim(' name  '::name);",
					Expected: []sql.Row{{"name"}},
				},
				{
					Query:    "SELECT initcap('name'::name);",
					Expected: []sql.Row{{"Name"}},
				},
				{
					Query:    "SELECT left('name'::name, 2);",
					Expected: []sql.Row{{"na"}},
				},
				{
					Query:    "SELECT length('name'::name);",
					Expected: []sql.Row{{4}},
				},
				{
					Query:    "SELECT lower('naMe'::name);",
					Expected: []sql.Row{{"name"}},
				},
				{
					Query:    "SELECT lpad('name'::name, 7, '*');",
					Expected: []sql.Row{{"***name"}},
				},
			},
		},
		{
			Name:        "quote_ident",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `select quote_ident('hi"bye');`,
					Expected: []sql.Row{{`"hi""bye"`}},
				},
				{
					Query:    `select quote_ident('hi""bye');`,
					Expected: []sql.Row{{`"hi""""bye"`}},
				},
				{
					Query:    `select quote_ident('hi"""bye');`,
					Expected: []sql.Row{{`"hi""""""bye"`}},
				},
				{
					Query:    `select quote_ident('hi"b"ye');`,
					Expected: []sql.Row{{`"hi""b""ye"`}},
				},
			},
		},
		{
			Name:        "translate",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `select translate('12345', '143', 'ax');`,
					Expected: []sql.Row{{`a2x5`}},
				},
				{
					Query:    `select translate('12345', '143', 'axs');`,
					Expected: []sql.Row{{`a2sx5`}},
				},
				{
					Query:    `select translate('12345', '143', 'axsl');`,
					Expected: []sql.Row{{`a2sx5`}},
				},
				{
					Query:    `select translate('こんにちは', 'ん', 'a');`,
					Expected: []sql.Row{{`こaにちは`}},
				},
			},
		},
		{
			Name:        "substring with integer arg",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT substr('hello', 2)`,
					Expected: []sql.Row{{"ello"}},
				},
				{
					Query:    `SELECT substring('hello', 2)`,
					Expected: []sql.Row{{"ello"}},
				},
			},
		},
		{
			Name:        "substring with integer args",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT substr('hello', 2, 3)`,
					Expected: []sql.Row{{"ell"}},
				},
				{
					Query:    `SELECT substring('hello', 2, 3)`,
					Expected: []sql.Row{{"ell"}},
				},
			},
		},
		{
			Name:        "substring with integer args, expanded form",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT substr('hello' from 2 for 3)`,
					ExpectedErr: "syntax error",
				},
				{
					Query:    `SELECT substring('hello' from 2 for 3)`,
					Expected: []sql.Row{{"ell"}},
				},
				{
					Query:       `SELECT substr('hello' from 2)`,
					ExpectedErr: "syntax error",
				},
				{
					Query:    `SELECT substring('hello' from 2)`,
					Expected: []sql.Row{{"ello"}},
				},
				{
					Query:       `SELECT substr('hello' for 3)`,
					ExpectedErr: "syntax error",
				},
				{
					Query:    `SELECT substring('hello' for 3)`,
					Skip:     true, // ERROR: function substring(unknown, bigint, integer) does not exist
					Expected: []sql.Row{{"hel"}},
				},
			},
		},
		{
			Name:        "substring with regex",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT substring('hello', 'l+')",
					Expected: []sql.Row{{"ll"}},
				},
				{
					Query:    "SELECT substring('hello' FROM 'l+')",
					Expected: []sql.Row{{"ll"}},
				},
				{
					Query:    `SELECT substring('hello.' similar 'hello#.' escape '#')`,
					Skip:     true, // syntax error
					Expected: []sql.Row{{"hello."}},
				},
				{
					Query:    `SELECT substring('Thomas' similar '%#"o_a#"_' escape '#')`,
					Skip:     true, // syntax error
					Expected: []sql.Row{{"oma"}},
				},
			},
		},
		{
			Name: "string_agg",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT primary key, v1 INT, v2 TEXT);",
				"INSERT INTO test VALUES (1, 1, 'a'), (2, 2, 'b'), (3, 3, 'c'), (4, 4, 'd'), (5, 5, 'e');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT string_agg(v1::text, ',') FROM test;`,
					Expected: []sql.Row{
						{"1,2,3,4,5"},
					},
				},
				{
					Query: `SELECT string_agg(v2, '|') FROM test;`,
					Expected: []sql.Row{
						{"a|b|c|d|e"},
					},
				},
				{
					Query: `SELECT STRING_AGG(concat(v1::text, v2), ' * ') FROM test;`,
					Expected: []sql.Row{
						{"1a * 2b * 3c * 4d * 5e"},
					},
				},
				{
					Skip:  true, // can't use expressions for separator because GROUP_CONCAT can't at the moment
					Query: `SELECT STRING_agg(concat(v1::text, v2), CONCAT(' *', ' ') ORDER BY V1 DESC) FROM test;`,
					Expected: []sql.Row{
						{"5e * 4d * 3c * 2b * 1a"},
					},
				},
				{
					Query:       `SELECT STRING_AGG(v2, '*', v1) FROM test;`,
					ExpectedErr: "string_agg requires two arguments",
				},
				{
					Query:       `SELECT STRING_AGG(v2) FROM test;`,
					ExpectedErr: "string_agg requires two arguments",
				},
			},
		},
	})
}

func TestFormatFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:        "test to_char",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'YYYY-MM-DD HH24:MI:SS.MS');`,
					Expected: []sql.Row{
						{"2021-09-15 21:43:56.123"},
					},
				},
				{
					Query: `SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'HH HH12 HH24 hh hh12 hh24 H h hH Hh');`,
					Expected: []sql.Row{
						{"09 09 21 09 09 21 H h hH Hh"},
					},
				},
				{
					Query: `SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'MI mi M m');`,
					Expected: []sql.Row{
						{"43 43 M m"},
					},
				},
				{
					Query: `SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'SS ss S s MS ms Ms mS US us Us uS');`,
					Expected: []sql.Row{
						{"56 56 S s 123 123 Ms mS 123457 123457 Us uS"},
					},
				},
				{
					Query: `SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'Y,YYY y,yyy YYYY yyyy YYY yyy YY yy Y y');`,
					Expected: []sql.Row{
						{"2,021 2,021 2021 2021 021 021 21 21 1 1"},
					},
				},
				{
					Query: `SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'MONTH Month month MON Mon mon MM mm Mm mM');`,
					Expected: []sql.Row{
						{"SEPTEMBER September september SEP Sep sep 09 09 Mm mM"},
					},
				},
				{
					Query: `SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'DAY Day day DDD ddd DY Dy dy DD dd D d');`,
					Expected: []sql.Row{
						{"WEDNESDAY Wednesday wednesday 258 258 WED Wed wed 15 15 4 4"},
					},
				},
				{
					Query: `SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'DAY Day day DDD ddd DY Dy dy DD dd D d');`,
					Expected: []sql.Row{
						{"WEDNESDAY Wednesday wednesday 258 258 WED Wed wed 15 15 4 4"},
					},
				},
				{
					Query: `SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'IW iw');`,
					Expected: []sql.Row{
						{"37 37"},
					},
				},
				{
					Query: `SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'AM PM am pm A.M. P.M. a.m. p.m.');`,
					Expected: []sql.Row{
						{"PM PM pm pm P.M. P.M. p.m. p.m."},
					},
				},
				{
					Query: `SELECT to_char(timestamp '2021-09-15 21:43:56.123456789', 'Q q');`,
					Expected: []sql.Row{
						{"3 3"},
					},
				},
			},
		},
	})
}

func TestUnknownFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:        "unknown functions",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT unknown_func();`,
					ExpectedErr: `function: 'unknown_func' not found`,
				},
			},
		},
		{
			Name: "Unsupported group_concat syntax",
			SetUpScript: []string{
				"CREATE TABLE x (pk int)",
				"INSERT INTO x VALUES (1),(2),(3),(4),(NULL)",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `SELECT group_concat(pk ORDER BY pk) FROM x;`,
					ExpectedErr: "is not yet supported", // error message is kind of nonsensical, we just want to make sure there isn't a panic
				},
			},
		},
	})
}

func TestSelectFromFunctions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name:        "select * FROM functions",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT array_to_string(ARRAY[1, 2, 3, NULL, 5], ',', '*')`,
					Expected: []sql.Row{{"1,2,3,*,5"}},
				},
				{
					Query:    `SELECT * FROM array_to_string(ARRAY[1, 2, 3, NULL, 5], ',', '*')`,
					Expected: []sql.Row{{"1,2,3,*,5"}},
				},
				{
					Query:    `SELECT * FROM array_to_string(ARRAY[37.89, 1.2], '_');`,
					Expected: []sql.Row{{"37.89_1.2"}},
				},
				{
					Query:    `SELECT * FROM format_type('text'::regtype, 4);`,
					Expected: []sql.Row{{"text(4)"}},
				},
				{
					Query:    `SELECT * from format_type(874938247, 20);`,
					Expected: []sql.Row{{"???"}},
				},
				{
					Query: `SELECT * FROM to_char(timestamp '2021-09-15 21:43:56.123456789', 'IW iw');`,
					Expected: []sql.Row{
						{"37 37"},
					},
				},
				{
					Query: `SELECT * from format_type('text'::regtype, -1);`,
					Expected: []sql.Row{
						{"text"},
					},
				},
				{
					Query:    `SELECT "left" FROM left('name'::name, 2);`,
					Expected: []sql.Row{{"na"}},
				},
				{
					Query:    "SELECT length FROM length('name'::name);",
					Expected: []sql.Row{{4}},
				},
				{
					Query:    "SELECT lower FROM lower('naMe'::name);",
					Expected: []sql.Row{{"name"}},
				},
				{
					Query:    "SELECT * FROM lpad('name'::name, 7, '*');",
					Expected: []sql.Row{{"***name"}},
				},
			},
		},
		{
			Name: "test select  from dolt_ functions",
			Skip: true, // need a way for single-row functions to declare a schema like table functions do, maybe just by modeling them as table functions in the first place
			SetUpScript: []string{
				"CREATE TABLE test (pk INT primary key, v1 INT, v2 TEXT);",
				"INSERT INTO test VALUES (1, 1, 'a'), (2, 2, 'b'), (3, 3, 'c'), (4, 4, 'd'), (5, 5, 'e');",
				"call dolt_commit('-Am', 'first table');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `select * from dolt_branch('newBranch')`,
					Expected: []sql.Row{{0}},
				},
				{
					Query:    `select status from dolt_checkout('newBranch')`,
					Expected: []sql.Row{{0}},
				},
				{
					Query: `insert into test values (6, 6, 'f')`,
				},
				{
					Query:    `select length(commit_hash) > 0 from (select commit_hash from dolt_commit('-Am', 'added f') as result)`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    "select dolt_checkout('main')",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    `select fast_forward, conflicts from dolt_merge('newBranch')`,
					Expected: []sql.Row{{"t", 0}},
				},
			},
		},
	})
}

func TestSetReturningFunctions(t *testing.T) {
	RunScripts(
		t,
		[]ScriptTest{
			{
				Name: "generate_series",
				Assertions: []ScriptTestAssertion{
					{
						Query:    `SELECT generate_series(1,3)`,
						Expected: []sql.Row{{1}, {2}, {3}},
					},
					{
						Query:    `SELECT generate_series(1,6,2)`,
						Expected: []sql.Row{{1}, {3}, {5}},
					},
					{
						Query:       `SELECT generate_series(1,6,0)`,
						ExpectedErr: "step size cannot equal zero",
					},
					{
						Query:    `SELECT generate_series(6,1,-2)`,
						Expected: []sql.Row{{6}, {4}, {2}},
					},
					{
						Query:    `SELECT generate_series(1.5,6,2)`,
						Expected: []sql.Row{{Numeric("1.5")}, {Numeric("3.5")}, {Numeric("5.5")}},
					},
					{
						Query:       `SELECT generate_series(1,6,0)`,
						ExpectedErr: "step size cannot equal zero",
					},
					{
						Query:    `SELECT generate_series(6,2.2,-2)`,
						Expected: []sql.Row{{Numeric("6")}, {Numeric("4")}},
					},
					{
						Query: `SELECT generate_series('2008-03-01 00:00'::timestamp,'2008-03-02 12:00', '10 hours');`,
						Expected: []sql.Row{
							{"2008-03-01 00:00:00"},
							{"2008-03-01 10:00:00"},
							{"2008-03-01 20:00:00"},
							{"2008-03-02 06:00:00"},
						},
					},
					{
						Query: `SELECT generate_series('2008-03-02 12:00'::timestamp,'2008-03-01 00:00'::timestamp, '-10 hours');`,
						Expected: []sql.Row{
							{"2008-03-02 12:00:00"},
							{"2008-03-02 02:00:00"},
							{"2008-03-01 16:00:00"},
							{"2008-03-01 06:00:00"},
						},
					},
				},
			},
			{
				Name: "generate_series as table function",
				Assertions: []ScriptTestAssertion{
					{
						Query:    `SELECT * FROM generate_series(1,3)`,
						Expected: []sql.Row{{1}, {2}, {3}},
					},
					{
						Query:    `SELECT * FROM generate_series(1,6,2)`,
						Expected: []sql.Row{{1}, {3}, {5}},
					},
					{
						Query: `SELECT * FROM generate_series('2008-03-02 12:00'::timestamp,'2008-03-01 00:00'::timestamp, '-10 hours'::interval);`,
						Expected: []sql.Row{
							{"2008-03-02 12:00:00"},
							{"2008-03-02 02:00:00"},
							{"2008-03-01 16:00:00"},
							{"2008-03-01 06:00:00"},
						},
					},
					{
						Skip:  true, // TODO: cannot cast unknown to interval, but this should work
						Query: `SELECT * FROM generate_series('2008-03-02 12:00'::timestamp,'2008-03-01 00:00'::timestamp, '-10 hours');`,
						Expected: []sql.Row{
							{"2008-03-02 12:00:00"},
							{"2008-03-02 02:00:00"},
							{"2008-03-01 16:00:00"},
							{"2008-03-01 06:00:00"},
						},
					},
				},
			},
			{
				Name: "nested generate_series",
				// Nested SRF expressions cause an infinite loop, skipped in regression tests.
				// Challenging to fix with the current expression eval architecture and very marginal as a use case.
				Skip: true,
				Assertions: []ScriptTestAssertion{
					{
						Query: `SELECT generate_series(1, generate_series(1, 3))`,
						Expected: []sql.Row{
							{1}, {1}, {2}, {1}, {2}, {3},
						},
					},
				},
			},
			{
				Name: "limit, offset, sort",
				Assertions: []ScriptTestAssertion{
					{
						Query: `SELECT a, generate_series(1,2) FROM (VALUES(1),(2),(3)) r(a) LIMIT 2 OFFSET 2;`,
						Expected: []sql.Row{
							{2, 1},
							{2, 2},
						},
					},
					{
						Query: `SELECT a, generate_series(1,2) FROM (VALUES(1),(2),(3)) r(a) ORDER BY 1;`,
						Expected: []sql.Row{
							{1, 1},
							{1, 2},
							{2, 1},
							{2, 2},
							{3, 1},
							{3, 2},
						},
					},
				},
			},
			{
				Name: "generate_series with table",
				SetUpScript: []string{
					"CREATE TABLE t1 (pk INT primary key, v1 INT);",
					"INSERT INTO t1 VALUES (1, 1), (2, 2), (3, 3);",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: `SELECT generate_series(1,3), pk from t1`,
						Expected: []sql.Row{
							{1, 1},
							{2, 1},
							{3, 1},
							{1, 2},
							{2, 2},
							{3, 2},
							{1, 3},
							{2, 3},
							{3, 3},
						},
					},
					{
						Query: `SELECT generate_series(1,3) + pk, pk from t1`,
						Expected: []sql.Row{
							{2, 1},
							{3, 1},
							{4, 1},
							{3, 2},
							{4, 2},
							{5, 2},
							{4, 3},
							{5, 3},
							{6, 3},
						},
					},
				},
			},
			{
				Name: "set returning function as table function: generate_series",
				Skip: true, // select * from functions does not work yet
				Assertions: []ScriptTestAssertion{
					{
						Query:    `select * from generate_series(1,3)`,
						Expected: []sql.Row{{1}, {2}, {3}},
					},
					{
						Query:    `select sum(null::int4) from generate_series(1,3);`,
						Expected: []sql.Row{{nil}},
					},
					{
						Query: `SELECT * from generate_series('2008-03-01 00:00'::timestamp,'2008-03-02 12:00', '10 hours');`,
						Expected: []sql.Row{
							{"2008-03-01 00:00:00"},
							{"2008-03-01 10:00:00"},
							{"2008-03-01 20:00:00"},
							{"2008-03-02 06:00:00"},
						},
					},
				},
			},
			{
				Name: "generate_subscripts",
				SetUpScript: []string{
					"CREATE TABLE t1 (pk INT primary key, v1 INT[]);",
					"INSERT INTO t1 VALUES (1, ARRAY[1, 2, 3]), (2, ARRAY[4, 5]), (3, NULL);",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "select generate_subscripts(v1, 1) from t1 where pk = 1",
						Expected: []sql.Row{
							{1}, {2}, {3},
						},
					},
					{
						Query: "select generate_subscripts(v1, 1) + 100 from t1 where pk = 1",
						Expected: []sql.Row{
							{101}, {102}, {103},
						},
					},
					{
						Query:    "select generate_subscripts(v1, 1) from t1 where pk = 3",
						Expected: []sql.Row{},
					},
					{
						Query: "select generate_subscripts(v1, 1), v1 from t1",
						Expected: []sql.Row{
							{1, "{1,2,3}"},
							{2, "{1,2,3}"},
							{3, "{1,2,3}"},
							{1, "{4,5}"},
							{2, "{4,5}"},
						},
					},
				},
			},
			{
				Name: "generate_subscripts with join",
				SetUpScript: []string{
					"CREATE TABLE t1 (a INT[]);",
					"CREATE TABLE t2 (b int[]);",
					"INSERT INTO t1 VALUES (ARRAY[1]), (ARRAY[1, 2, 3])",
					"INSERT INTO t2 VALUES (ARRAY[9,10])",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "select generate_subscripts(a, 1), a, generate_subscripts(b, 1), b from t1, t2;",
						Expected: []sql.Row{
							{1, "{1}", 1, "{9,10}"},
							{nil, "{1}", 2, "{9,10}"},
							{1, "{1,2,3}", 1, "{9,10}"},
							{2, "{1,2,3}", 2, "{9,10}"},
							{3, "{1,2,3}", nil, "{9,10}"},
						},
					},
				},
			},
			{
				Name: "generate_subscripts and generate_series combined",
				SetUpScript: []string{
					"CREATE TABLE t1 (a INT[]);",
					"INSERT INTO t1 VALUES (ARRAY[1, 2, 3]), (ARRAY[4, 5]);",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: "select generate_subscripts(a, 1), a, generate_series(1,4) from t1;",
						Expected: []sql.Row{
							{1, "{1,2,3}", 1},
							{2, "{1,2,3}", 2},
							{3, "{1,2,3}", 3},
							{nil, "{1,2,3}", 4},
							{1, "{4,5}", 1},
							{2, "{4,5}", 2},
							{nil, "{4,5}", 3},
							{nil, "{4,5}", 4},
						},
					},
				},
			},
			{
				Name: "set generation with other func calls",
				SetUpScript: []string{
					"CREATE sequence test_seq START WITH 1 INCREMENT BY 3;",
				},
				Assertions: []ScriptTestAssertion{
					{
						Query: `SELECT generate_series(1, 5), nextval('test_seq')`,
						Expected: []sql.Row{
							{1, 1},
							{2, 4},
							{3, 7},
							{4, 10},
							{5, 13},
						},
					},
				},
			},
		},
	)
}
