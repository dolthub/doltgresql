// Copyright 2026 Dolthub, Inc.
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
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	driver "github.com/dolthub/doltgresql/integration-tests/go-sql-server-driver/driver"
)

// TestConcurrentDropDatabase is a regression test for #10692.
//
// A lock-ordering bug between *DatabaseProvider and *DoltSession meant that
// DROP DATABASE during concurrency could cause the DatabaseProvider to
// deadlock and the server databases to become unavailable.
func TestConcurrentDropDatabase(t *testing.T) {
	t.Parallel()
	ports := newPorts(t)
	u, err := driver.NewDoltUser()
	require.NoError(t, err)
	t.Cleanup(func() {
		u.Cleanup()
	})

	rs, err := u.MakeRepoStore()
	require.NoError(t, err)

	_, err = rs.MakeRepo("concurrent_drop_database_test")
	require.NoError(t, err)

	srvSettings := &driver.Server{
		Args:        []string{},
		DynamicPort: "server_port",
	}
	server := StartServer(t, rs, "concurrent_drop_database_test", srvSettings, ports)

	db, err := server.DB(driver.Connection{})
	require.NoError(t, err)
	db.SetMaxIdleConns(0)
	defer func() {
		require.NoError(t, db.Close())
	}()
	ctx := t.Context()

	eg, ctx := errgroup.WithContext(ctx)
	var numcreates int32 = 0
	const numWriters = 8
	const numDatabasesPerWriter = 12
	startCh := make(chan struct{})
	readyCh := make(chan struct{})
	for i := range numWriters {
		eg.Go(func() error {
			conn, err := db.Conn(ctx)
			if err != nil {
				return err
			}
			defer conn.Close()
			select {
			case readyCh <- struct{}{}:
			case <-ctx.Done():
				return nil
			}
			select {
			case <-startCh:
			case <-ctx.Done():
				return nil
			}
			for j := range numDatabasesPerWriter {
				if ctx.Err() != nil {
					return context.Cause(ctx)
				}
				database := fmt.Sprintf("db%08d%08d", i, j)
				_, err := conn.ExecContext(ctx, "CREATE DATABASE "+database)
				if err != nil {
					return err
				}
				atomic.AddInt32(&numcreates, 1)
				_, err = conn.ExecContext(ctx, "DROP DATABASE "+database)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}
	for range numWriters {
		select {
		case <-readyCh:
		case <-ctx.Done():
			// This will fail.
			require.NoError(t, eg.Wait())
			t.FailNow()
		}
	}
	close(startCh)
	require.NoError(t, eg.Wait())
	ctx = t.Context()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()
	rows, err := conn.QueryContext(ctx, "SELECT datname FROM pg_database")
	require.NoError(t, err)
	defer rows.Close()
	var databases []string
	for rows.Next() {
		var db string
		err = rows.Scan(&db)
		require.NoError(t, err)
		databases = append(databases, db)
	}
	require.NoError(t, rows.Err())
	// The provider must still be functional, and none of the created-and-dropped
	// databases should remain. The test database must still be present.
	require.Contains(t, databases, "concurrent_drop_database_test")
	for _, name := range databases {
		require.False(t, strings.HasPrefix(name, "db0000"),
			"created-and-dropped database %q should not remain", name)
	}
	t.Logf("created and dropped %d databases", numcreates)
}
