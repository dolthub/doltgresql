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

func TestSameTypes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Integer types",
			SetUpScript: []string{
				"CREATE TABLE test1 (v1 SMALLINT, v2 INTEGER, v3 BIGINT);",
				"CREATE TABLE test2 (v1 INT2, v2 INT4, v3 INT8);",
				"INSERT INTO test1 VALUES (1, 2, 3), (4, 5, 6);",
				"INSERT INTO test2 VALUES (1, 2, 3), (4, 5, 6);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test1 ORDER BY 1;",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, 6},
					},
				},
				{
					Query: "SELECT * FROM test2 ORDER BY 1;",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, 6},
					},
				},
			},
		},
		{
			Name: "Arbitrary precision types",
			SetUpScript: []string{
				"CREATE TABLE test (v1 DECIMAL(10, 1), v2 NUMERIC(11, 2));",
				"INSERT INTO test VALUES (14854.5, 2504.25), (566821525.5, 735134574.75);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY 1;",
					Expected: []sql.Row{
						{14854.5, 2504.25},
						{566821525.5, 735134574.75},
					},
				},
			},
		},
		{
			Name: "Floating point types",
			SetUpScript: []string{
				"CREATE TABLE test1 (v1 REAL, v2 DOUBLE PRECISION);",
				"CREATE TABLE test2 (v1 FLOAT4, v2 FLOAT8);",
				"INSERT INTO test1 VALUES (10.125, 20.4), (40.875, 81.6);",
				"INSERT INTO test2 VALUES (10.125, 20.4), (40.875, 81.6);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test1 ORDER BY 1;",
					Expected: []sql.Row{
						{10.125, 20.4},
						{40.875, 81.6},
					},
				},
				{
					Query: "SELECT * FROM test2 ORDER BY 1;",
					Expected: []sql.Row{
						{10.125, 20.4},
						{40.875, 81.6},
					},
				},
			},
		},
		{
			// TIME has the same name, but operates a bit differently, so it's not included as a "same type"
			Name: "Date and time types",
			SetUpScript: []string{
				"CREATE TABLE test (v1 TIMESTAMP, v2 DATE);",
				"INSERT INTO test VALUES ('1986-08-02 17:04:22', '2023-09-03');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY 1;",
					Expected: []sql.Row{
						{"1986-08-02 17:04:22", "2023-09-03 00:00:00"},
					},
				},
			},
		},
		{
			// ENUM exists, but features too many differences to incorporate as a "same type"
			// BLOB exists, but functions as a BYTEA, which operates differently than a BINARY/VARBINARY in MySQL
			Name: "Text types",
			SetUpScript: []string{
				"CREATE TABLE test (v1 CHARACTER VARYING(255), v2 CHARACTER(3), v3 TEXT);",
				"INSERT INTO test VALUES ('abc', 'def', 'ghi'), ('jkl', 'mno', 'pqr');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY 1;",
					Expected: []sql.Row{
						{"abc", "def", "ghi"},
						{"jkl", "mno", "pqr"},
					},
				},
			},
		},
		{
			Name: "JSON type",
			SetUpScript: []string{
				"CREATE TABLE test (v1 INT, v2 JSON);",
				`INSERT INTO test VALUES (1, '{"key1": {"key": "value"}}'), (2, '{"key1": "value1", "key2": "value2"}'), (3, '{"key1": {"key": [2,3]}}');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY 1;",
					Expected: []sql.Row{
						{1, `{"key1":{"key":"value"}}`},
						{2, `{"key1":"value1","key2":"value2"}`},
						{3, `{"key1":{"key":[2,3]}}`},
					},
				},
			},
		},
	})
}
