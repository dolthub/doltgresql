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
	//TODO: fix me
	return
	//port := getEmptyPort(t)
	//go RunMain([]string{fmt.Sprintf("--port=%d", port)})
	//port := 5431
	port := 5432

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@localhost:%d/postgres", port))
	require.NoError(t, err)
	defer conn.Close(ctx)

	func() {
		//rows, err := conn.Query(ctx, "CREATE DATABASE testdb;")
		rows, err := conn.Query(ctx, "SELECT * FROM test;")
		require.NoError(t, err)
		defer rows.Close()
		for rows.Next() {
			row, err := rows.Values()
			require.NoError(t, err)
			row = row
		}
	}()
}

func getEmptyPort(t *testing.T) int {
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	return port
}
