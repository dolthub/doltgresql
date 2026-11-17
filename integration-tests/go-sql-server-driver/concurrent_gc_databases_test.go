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
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	driver "github.com/dolthub/doltgresql/integration-tests/go-sql-server-driver/driver"
)

func TestConcurrentGCAgainstDifferentDatabases(t *testing.T) {
	t.Parallel()
	ports := newPorts(t)

	u, err := driver.NewDoltUser()
	require.NoError(t, err)
	t.Cleanup(func() { u.Cleanup() })
	rs, err := u.MakeRepoStore()
	require.NoError(t, err)
	for _, name := range []string{"gc_first", "gc_second"} {
		_, err := rs.MakeRepo(name)
		require.NoError(t, err)
	}

	server := StartServer(t, rs, "gc_first", &driver.Server{
		Args:          []string{},
		DynamicPort:   "server_port",
		Envs:          []string{"DOLT_GC_SAFEPOINT_CONTROLLER_CHOICE=session_aware"},
		LogNotMatches: []string{"recovered panic", "caught panic"},
	}, ports)
	connect := func(name string) *sql.DB {
		db, err := server.DB(driver.Connection{Database: name})
		require.NoError(t, err)
		t.Cleanup(func() { db.Close() })
		return db
	}
	firstDB, secondDB := connect("gc_first"), connect("gc_second")
	ctx, cancel := context.WithTimeout(t.Context(), 45*time.Second)
	defer cancel()

	for _, db := range []*sql.DB{firstDB, secondDB} {
		_, err = db.ExecContext(ctx, "CREATE TABLE vals (id INT PRIMARY KEY, val INT)")
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, "INSERT INTO vals VALUES (1, 1)")
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, "SELECT DOLT_COMMIT('-Am', 'initial value')")
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, "UPDATE vals SET val = 2 WHERE id = 1")
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, "SELECT DOLT_COMMIT('-Am', 'updated value')")
		require.NoError(t, err)
	}

	// An open command in a third session keeps both GCs at their safepoints.
	// This makes overlap observable even when the databases are small.
	//
	// Dolt's version of this test releases the blocker with KILL QUERY.
	// Doltgres does not implement pg_cancel_backend or CancelRequest, so the
	// blocker here is a bounded sleep which must outlast the checks below.
	sleeper, err := firstDB.Conn(ctx)
	require.NoError(t, err)
	defer sleeper.Close()
	var sleeperID int
	require.NoError(t, sleeper.QueryRowContext(ctx, "SELECT pg_backend_pid()").Scan(&sleeperID))
	sleepDone := make(chan error, 1)
	go func() {
		_, err := sleeper.ExecContext(ctx, "SELECT pg_sleep(10)")
		sleepDone <- err
	}()

	const activeSleep = "SELECT COUNT(*) FROM pg_stat_activity WHERE pid = $1::int4 AND LOWER(query) LIKE 'select pg_sleep%'"
	for {
		var count int
		require.NoError(t, firstDB.QueryRowContext(ctx, activeSleep, sleeperID).Scan(&count))
		if count == 1 {
			break
		}
		select {
		case err := <-sleepDone:
			t.Fatalf("blocking command ended before the GCs started: %v", err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(25 * time.Millisecond):
		}
	}

	type gcResult struct {
		name string
		err  error
	}
	gcDone := make(chan gcResult, 2)
	for name, db := range map[string]*sql.DB{"gc_first": firstDB, "gc_second": secondDB} {
		go func() {
			_, err := db.ExecContext(ctx, "SELECT DOLT_GC('--full')")
			gcDone <- gcResult{name: name, err: err}
		}()
	}

	// Both statements must still be executing before we release the blocker.
	// A controller that rejects the second waiter fails this check.
	const activeGC = "SELECT COUNT(*) FROM pg_stat_activity WHERE LOWER(query) LIKE 'select dolt_gc%'"
	for {
		var count int
		require.NoError(t, firstDB.QueryRowContext(ctx, activeGC).Scan(&count))
		if count == 2 {
			break
		}
		select {
		case result := <-gcDone:
			t.Fatalf("%s GC returned before both GCs reached the safepoint: %v", result.name, result.err)
		case err := <-sleepDone:
			t.Fatalf("blocking command ended before both GCs started: %v", err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(25 * time.Millisecond):
		}
	}
	// Seeing both SQL statements once is insufficient: one could still be
	// entering BeginGC. Keep the blocking command open long enough to verify
	// that neither GC is rejected or completes while the other is active.
	select {
	case result := <-gcDone:
		t.Fatalf("%s GC returned while the blocking command was still open: %v", result.name, result.err)
	case err := <-sleepDone:
		t.Fatalf("blocking command ended before the GCs were observed waiting: %v", err)
	case <-time.After(time.Second):
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}

	select {
	case err := <-sleepDone:
		require.NoError(t, err)
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	for range 2 {
		select {
		case result := <-gcDone:
			require.NoErrorf(t, result.err, "%s GC failed", result.name)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	for _, db := range []*sql.DB{firstDB, secondDB} {
		var val int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT val FROM vals WHERE id = 1").Scan(&val))
		require.Equal(t, 2, val)
	}
}
