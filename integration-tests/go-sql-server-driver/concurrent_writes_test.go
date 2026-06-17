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

package main

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	driver "github.com/dolthub/doltgresql/integration-tests/go-sql-server-driver/driver"
)

// TestConcurrentWrites verifies concurrent write behavior and transaction locking in the SQL server driver.
func TestConcurrentWrites(t *testing.T) {
	t.Parallel()
	ports := newPorts(t)
	u, err := driver.NewDoltUser()
	require.NoError(t, err)
	t.Cleanup(func() {
		u.Cleanup()
	})

	rs, err := u.MakeRepoStore()
	require.NoError(t, err)

	_, err = rs.MakeRepo("concurrent_writes_test")
	require.NoError(t, err)

	srvSettings := &driver.Server{
		Args:        []string{},
		DynamicPort: "server_port",
	}
	server := StartServer(t, rs, "concurrent_writes_test", srvSettings, ports)

	db, err := server.DB(driver.Connection{})
	require.NoError(t, err)
	db.SetMaxIdleConns(0)
	defer func() {
		require.NoError(t, db.Close())
	}()
	ctx := t.Context()
	func() {
		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()
		// Create table and initial data.
		_, err = conn.ExecContext(ctx, "CREATE TABLE data (id VARCHAR(64) PRIMARY KEY, worker INT, data TEXT, created_at TIMESTAMP)")
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, "SELECT DOLT_COMMIT('-Am', 'init with table')")
		require.NoError(t, err)
	}()

	eg, ctx := errgroup.WithContext(ctx)
	nextInt := uint32(0)
	const numWriters = 32
	startCh := make(chan struct{})
	readyCh := make(chan struct{})
	for i := range numWriters {
		eg.Go(func() error {
			db, err := server.DB(driver.Connection{})
			if err != nil {
				return err
			}
			defer db.Close()
			db.SetMaxOpenConns(1)
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
			// Each thread writes 16.
			for j := range 16 {
				if ctx.Err() != nil {
					return nil
				}
				key := fmt.Sprintf("main-%d-%d", i, j)
				_, err := conn.ExecContext(ctx, "INSERT INTO data VALUES ($1,$2,$3,$4)", key, i, key, time.Now())
				if err != nil {
					return err
				}
				atomic.AddUint32(&nextInt, 1)
				_, err = conn.ExecContext(ctx, fmt.Sprintf("SELECT DOLT_COMMIT('-Am', 'insert %s')", key))
				if err != nil && !strings.Contains(err.Error(), "nothing to commit") {
					// Technically we can get "nothing to commit" because of how DOLT_COMMIT works.
					// It first commits this transaction to the working set and then commits the
					// merged working set as a dolt commit.
					//
					// If we ever fix DOLT_COMMIT, we should fix this exception here as well.
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
	require.Equal(t, 512, int(nextInt))
	t.Logf("wrote %d", nextInt)
	ctx = t.Context()
	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, conn.Close())
	}()
	var i int
	err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM data").Scan(&i)
	require.NoError(t, err)
	require.Equal(t, int(nextInt), i)
	err = conn.QueryRowContext(ctx, "SELECT COUNT(*) FROM dolt_log").Scan(&i)
	require.NoError(t, err)
	t.Logf("ended with %d commits", i)
}
