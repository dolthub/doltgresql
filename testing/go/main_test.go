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
			Name: "String types",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 VARCHAR(13), v2 CHAR(11), v3 TEXT);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (1, 'hey1', 'heythere2', 'hellofellow3');",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test VALUES (2, 'hey44', 'heythere55', 'hellofellow66');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, "hey1", "heythere2", "hellofellow3"},
						{2, "hey44", "heythere55", "hellofellow66"},
					},
				},
			},
		},
		{
			Name: "Int types",
			SetUpScript: []string{
				"CREATE TABLE test (v1 SMALLINT, v2 INTEGER, v3 BIGINT);",
				"INSERT INTO test VALUES (1, 2, 3), (4, 5, 6);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY 1;",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, 6},
					},
				},
			},
		},
		{
			Name: "Float types",
			SetUpScript: []string{
				"CREATE TABLE test (v1 REAL, v2 DOUBLE PRECISION);",
				"INSERT INTO test VALUES (1.5, 2.25), (10.125, 15.0625);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY 1;",
					Expected: []sql.Row{
						{1.5, 2.25},
						{10.125, 15.0625},
					},
				},
			},
		},
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
