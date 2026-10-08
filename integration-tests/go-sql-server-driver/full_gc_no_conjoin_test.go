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
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	driver "github.com/dolthub/doltgresql/integration-tests/go-sql-server-driver/driver"
)

// This test asserts that running a full gc into the old gen at the
// point that it is at its conjoin limit will not generate a conjoin
// in the middle of the GC.
//
// It first generates empty commits and runs GC in a loop. After every
// GC, it checks how many table files are in the oldgen. When it sees
// them get conjoined, it assumes max table files is the observed high
// water mark.
//
// It then creates exactly that many table files in oldgen by creating
// empty commits and running gc however many times it needs to.
//
// Then it needs to create one more commit and run a
// dolt_gc(--full). Currently the sql-server logs whenever it begins
// a conjoin, and that determination is made synchronously on the
// write path. So, no real attempt is made to get a theoretically
// started conjoin to win the race against the newgen collection, for
// example.
//
// As described above, this test is tightly coupled with
// GenerationalNBS, NomsBlockStore and the file persisters, with
// conjoin strategy behavior. Doltgres uses the same Dolt storage
// engine, so the conjoin log lines (pkg=store.noms) and the on-disk
// oldgen table-file layout are identical.
//
// At the end of the test, there should be one file in the oldgen, the
// results of the --full. There should be exactly one "beginning
// conjoin of database" in the server logs and it should correspond to
// when we measured the high water mark. After shutting down the
// server should happily start again.
//
// NOTE: a successful online dolt_gc() invalidates every open connection
// to the doltgres server, so the helpers below open a fresh connection
// for each commit/gc iteration (db.SetMaxIdleConns(0) ensures each
// connection is a new one rather than a pooled, now-invalid connection).
func TestFullGCNoOldgenConjoin(t *testing.T) {
	t.Parallel()
	ports := newPorts(t)
	u, err := driver.NewDoltUser()
	require.NoError(t, err)
	t.Cleanup(func() {
		u.Cleanup()
	})

	dbname := "full_gc_no_oldgen_conjoin_test"

	rs, err := u.MakeRepoStore()
	require.NoError(t, err)
	repo, err := rs.MakeRepo(dbname)
	require.NoError(t, err)
	srvSettings := &driver.Server{
		Args:        []string{},
		DynamicPort: "server",
	}
	var conjoinStarted, conjoinFinished atomic.Bool
	upstreamLenOnConjoin := 0
	server := MakeServer(t, rs, rs.Dir, srvSettings, ports, driver.WithOutputVisitor(func(out string) {
		// Matching a line like:
		// time="..." level=info msg="beginning conjoin of database" database=full_gc_no_oldgen_conjoin_test generation=old pkg=store.noms upstream_len=257
		// Extracting its upstream_len to see exactly when the conjoin was triggered.
		if upstreamLenOnConjoin <= 0 && strings.Contains(out, "beginning conjoin of database") {
			if i := strings.Index(out, "upstream_len="); i != -1 {
				i += len("upstream_len=")
				if _, err := fmt.Sscanf(out[i:], "%d", &upstreamLenOnConjoin); err != nil {
					upstreamLenOnConjoin = -1
				}
			}
			conjoinStarted.Store(true)
		}
		// Matching the line:
		// time="..." level=info msg="conjoin completed successfully" database=full_gc_no_oldgen_conjoin_test generation=old new_upstream_len=2 pkg=store.noms
		// So we can block on further operations until conjoin is done.
		if strings.Contains(out, "conjoin completed successfully") {
			conjoinFinished.Store(true)
		}
	}))
	server.DBName = dbname

	oldgendir := filepath.Join(repo.Dir, "/.dolt/noms/oldgen")

	CommitAndGCUntilConjoin(t, server, &conjoinStarted, oldgendir)
	require.Greater(t, upstreamLenOnConjoin, 0)
	require.Eventually(t, func() bool {
		return conjoinFinished.Load()
	}, 5*time.Second, 32*time.Millisecond)

	CreateUpToNumFiles(t, server, oldgendir, upstreamLenOnConjoin)
	cnt := CountTableFiles(t, oldgendir)
	t.Logf("now there are %d", cnt)

	RunGCFull(t, server, oldgendir)

	require.NoError(t, server.GracefulStop())
	output := server.Output.String()
	assert.Equal(t, 1, strings.Count(output, "beginning conjoin of database"))
	// The line for triggering a conjoin on policy but not proceeding because
	// conjoin is dynamically disabled looks like:
	// time="..." level=info msg="conjoin dynamically disabled. not conjoining." database=full_gc_no_oldgen_conjoin_test generation=old pkg=store.noms
	assert.Equal(t, 1, strings.Count(output, "conjoin dynamically disabled"))
	cnt = CountTableFiles(t, oldgendir)
	assert.Equal(t, 1, cnt)

	newServer := MakeServer(t, rs, rs.Dir, srvSettings, ports)
	newServer.DBName = dbname
	db, err := newServer.DB(driver.Connection{})
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, db.PingContext(t.Context()))
}

// isTableFileName reports whether n is the name of a Dolt/Noms table file: a
// 32-character string in Noms' base32 hash alphabet. This mirrors
// hash.MaybeParse from Dolt's store/hash package without importing it.
func isTableFileName(n string) bool {
	const alphabet = "0123456789abcdefghijklmnopqrstuv"
	if len(n) != 32 {
		return false
	}
	for _, c := range n {
		if !strings.ContainsRune(alphabet, c) {
			return false
		}
	}
	return true
}

func CountTableFiles(t *testing.T, dir string) int {
	var count int
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		n := d.Name()
		// Archive table files carry a .darc suffix.
		n = strings.TrimSuffix(n, ".darc")
		if isTableFileName(n) {
			count += 1
		}
		return nil
	})
	require.NoError(t, err)
	return count
}

func CommitAndGCUntilConjoin(t *testing.T, srv *driver.SqlServer, conjoinStarted *atomic.Bool, path string) {
	ctx := t.Context()
	db, err := srv.DB(driver.Connection{})
	require.NoError(t, err)
	defer db.Close()
	// Ensure each db.Conn() hands out a new connection rather than a pooled
	// one that a prior online GC has invalidated.
	db.SetMaxIdleConns(0)

	for {
		func() {
			conn, err := db.Conn(ctx)
			require.NoError(t, err)
			defer conn.Close()
			_, err = conn.ExecContext(ctx, "SELECT DOLT_COMMIT('-A', '--allow-empty', '-m', 'creating a new commit')")
			require.NoError(t, err)
			_, err = conn.ExecContext(ctx, "SELECT DOLT_GC()")
			require.NoError(t, err)
		}()
		if conjoinStarted.Load() {
			return
		}
	}
}

func CreateUpToNumFiles(t *testing.T, srv *driver.SqlServer, path string, numFiles int) {
	ctx := t.Context()
	db, err := srv.DB(driver.Connection{})
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxIdleConns(0)

	for {
		cnt := CountTableFiles(t, path)
		if cnt == numFiles {
			return
		}
		func() {
			conn, err := db.Conn(ctx)
			require.NoError(t, err)
			defer conn.Close()
			_, err = conn.ExecContext(ctx, "SELECT DOLT_COMMIT('-A', '--allow-empty', '-m', 'creating a new commit')")
			require.NoError(t, err)
			_, err = conn.ExecContext(ctx, "SELECT DOLT_GC()")
			require.NoError(t, err)
		}()
	}
}

func RunGCFull(t *testing.T, srv *driver.SqlServer, path string) {
	ctx := t.Context()
	db, err := srv.DB(driver.Connection{})
	require.NoError(t, err)
	defer db.Close()
	db.SetMaxIdleConns(0)

	conn, err := db.Conn(ctx)
	require.NoError(t, err)
	defer conn.Close()
	_ = CountTableFiles(t, path)
	_, err = conn.ExecContext(ctx, "SELECT DOLT_GC('--full')")
	require.NoError(t, err)
	cnt := CountTableFiles(t, path)
	require.Equal(t, 1, cnt)
}
