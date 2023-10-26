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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dserver "github.com/dolthub/doltgresql/server"
)

func TestSmokeTests(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Commit and diff across branches",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"INSERT INTO test VALUES (1, 1), (2, 2);",
				"CALL DOLT_ADD('-A');",
				"CALL DOLT_COMMIT('-m', 'initial commit');",
				"CALL DOLT_BRANCH('other');",
				"UPDATE test SET v1 = 3;",
				"CALL DOLT_ADD('-A');",
				"CALL DOLT_COMMIT('-m', 'commit main');",
				"CALL DOLT_CHECKOUT('other');",
				"UPDATE test SET v1 = 4 WHERE pk = 2;",
				"CALL DOLT_ADD('-A');",
				"CALL DOLT_COMMIT('-m', 'commit other');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "CALL DOLT_CHECKOUT('main');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 3},
						{2, 3},
					},
				},
				{
					Query:            "CALL DOLT_CHECKOUT('other');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 1},
						{2, 4},
					},
				},
				{
					Query: "SELECT from_pk, to_pk, from_v1, to_v1 FROM dolt_diff_test;",
					Expected: []sql.Row{
						{2, 2, 2, 4},
						{nil, 1, nil, 1},
						{nil, 2, nil, 2},
					},
				},
			},
		},
		{
			Name: "Unsupported MySQL statements",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SHOW CREATE TABLE;",
					ExpectedErr: true,
				},
			},
		},
	})
}

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

func TestSSL(t *testing.T) {
	port := GetUnusedPort(t)
	server.DefaultProtocolListenerFunc = dserver.NewLimitedListener
	code, _ := dserver.RunInMemory([]string{fmt.Sprintf("--port=%d", port), "--host=127.0.0.1"})
	require.Equal(t, 0, *code)

	ctx := context.Background()
	err := func() error {
		// The connection attempt may be made before the server has grabbed the port, so we'll retry the first
		// connection a few times.
		var conn *pgx.Conn
		var err error
		for i := 0; i < 3; i++ {
			conn, err = pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/?sslmode=require", port))
			if err == nil {
				break
			} else {
				time.Sleep(time.Second)
			}
		}
		if err != nil {
			return err
		}

		defer conn.Close(ctx)
		_, err = conn.Exec(ctx, "CREATE DATABASE postgres;")
		return err
	}()
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/postgres?sslmode=require", port))
	require.NoError(t, err)
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "CREATE TABLE test (pk INT8 PRIMARY KEY, v1 int4);")
	require.NoError(t, err)
	_, err = conn.Exec(ctx, "INSERT INTO test VALUES (3645, 37643);")
	require.NoError(t, err)
	rows, err := conn.Query(ctx, "SELECT * FROM test;")
	require.NoError(t, err)
	assert.Equal(t, NormalizeRows([]sql.Row{{3645, 37643}}), ReadRows(t, rows))
}
