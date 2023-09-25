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

package main

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestBasicConnection(t *testing.T) {
	port := getEmptyPort(t)
	go RunMainInMemory([]string{fmt.Sprintf("--port=%d", port)})

	ctx := context.Background()
	t.Run("Create Database", func(t *testing.T) {
		conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@localhost:%d/", port))
		require.NoError(t, err)
		defer conn.Close(ctx)

		func() {
			_, err := conn.Exec(ctx, "CREATE DATABASE postgres;")
			require.NoError(t, err)
		}()
	})

	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@localhost:%d/postgres", port))
	require.NoError(t, err)
	defer conn.Close(ctx)

	t.Run("Create Table", func(t *testing.T) {
		//TODO: fix CHAR
		_, err := conn.Exec(ctx, "CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 VARCHAR(13), v2 VARCHAR(11), v3 TEXT);")
		require.NoError(t, err)
	})
	t.Run("Insert 1", func(t *testing.T) {
		_, err := conn.Exec(ctx, "INSERT INTO test VALUES (1, 'hey1', 'heythere2', 'hellofellow3');")
		require.NoError(t, err)
	})
	t.Run("Insert 2", func(t *testing.T) {
		_, err := conn.Exec(ctx, "INSERT INTO test VALUES (2, 'hey44', 'heythere55', 'hellofellow66');")
		require.NoError(t, err)
	})
	t.Run("Select Rows", func(t *testing.T) {
		rows, err := conn.Query(ctx, "SELECT * FROM test;")
		require.NoError(t, err)
		defer rows.Close()

		expected := [][]interface{}{
			{int64(1), "hey1", "heythere2", "hellofellow3"},
			{int64(2), "hey44", "heythere55", "hellofellow66"},
		}
		i := int(0)
		for ; rows.Next(); i++ {
			row, err := rows.Values()
			require.NoError(t, err)
			require.ElementsMatch(t, expected[i], row)
		}
		require.Equal(t, int(2), i)
	})
}

func getEmptyPort(t *testing.T) int {
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	return port
}
