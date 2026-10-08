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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	driver "github.com/dolthub/doltgresql/integration-tests/go-sql-server-driver/driver"
)

// Archive levels accepted by dolt_gc's --archive-level option. These mirror
// chunks.GCArchiveLevel in Dolt (NoArchive=0, SimpleArchive=1); we use plain
// ints here to avoid depending on Dolt's internal store/chunks package.
const (
	noArchive     = 0
	simpleArchive = 1
)

func TestGCIncremental(t *testing.T) {
	t.Parallel()

	archiveLevels := []int{noArchive, simpleArchive}

	// smallFileSize is chosen so that every leaf chunk gets put in its own table file. In this test, we know
	// exactly how many table files to expect after GC.
	smallFileSize := 1

	// mediumFileSize is chosen so that multiple chunks are put into each table file, and ensures that we correctly handle
	// an in-progress table file at the end of the mark-and-sweep.
	mediumFileSize := 5_000

	// largeFileSize is arbitrarily large and tests the case where all leaf chunks fit in a single table file.
	largeFileSize := 1_000_0000

	fileSizes := []int{smallFileSize, mediumFileSize, largeFileSize}

	for _, archiveLevel := range archiveLevels {
		t.Run(fmt.Sprintf("archiveLevel=%d", archiveLevel), func(t *testing.T) {
			t.Parallel()
			for _, fileSize := range fileSizes {
				t.Run(fmt.Sprintf("fileSize=%d", fileSize), func(t *testing.T) {
					t.Parallel()
					for _, full := range []bool{true, false} {
						t.Run(fmt.Sprintf("full=%v", full), func(t *testing.T) {
							t.Parallel()

							runGCIncrementalTest(t, archiveLevel, fileSize, full)
						})
					}
				})
			}
		})
	}
}

func runGCIncrementalTest(t *testing.T, archiveLevel int, fileSize int, full bool) {
	ports := newPorts(t)

	u, err := driver.NewDoltUser()
	require.NoError(t, err)
	t.Cleanup(func() { u.Cleanup() })

	rs, err := u.MakeRepoStore()
	require.NoError(t, err)

	repoName := fmt.Sprintf("incremental_gc_test_archivelevel_%d_filesize_%d_full_%v", archiveLevel, fileSize, full)
	repo, err := rs.MakeRepo(repoName)
	require.NoError(t, err)

	ctx := context.Background()

	server := StartServer(t, rs, repoName, &driver.Server{
		Args:        []string{},
		DynamicPort: "server",
	}, ports)

	// Create the database and run GC
	func() {
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		populateDB(t, conn)
		// dolt_gc takes its options as string arguments; build the statement
		// with literals rather than bind parameters so the string-typed args
		// are passed exactly as the procedure expects.
		gcSQL := fmt.Sprintf("select dolt_gc('--archive-level','%d','--incremental-file-size','%d')", archiveLevel, fileSize)
		if full {
			gcSQL = fmt.Sprintf("select dolt_gc('--archive-level','%d','--incremental-file-size','%d','--full')", archiveLevel, fileSize)
		}
		_, err = conn.ExecContext(ctx, gcSQL)
		require.NoError(t, err)
	}()

	// Verify that the DB is in a valid state after GC. Doltgres does not expose
	// an fsck command (Dolt's `dolt fsck` CLI / dolt_fsck() are unavailable
	// here), so as a stand-in we reconnect (the previous connection is
	// invalidated by the online GC) and confirm the database is still
	// queryable and the committed data survived.
	requireDatabaseQueryable(t, server)

	repoSize, err := GetRepoSize(repo.Dir)
	require.NoError(t, err)
	// Both newgen and oldgen should contain multiple chunk files
	require.Greater(t, repoSize.NewGenC, 1)
	require.Greater(t, repoSize.OldGenC, 1)
}

// requireDatabaseQueryable opens a fresh connection to the server and confirms
// the committed table is readable. A successful online GC invalidates prior
// connections, so callers must not reuse a connection held across a GC.
func requireDatabaseQueryable(t *testing.T, server *driver.SqlServer) {
	ctx := context.Background()
	db, err := server.DB(driver.Connection{})
	require.NoError(t, err)
	defer db.Close()
	var n int
	err = db.QueryRowContext(ctx, "select count(*) from vals").Scan(&n)
	require.NoError(t, err)
	require.Greater(t, n, 0)
}

func populateDB(t *testing.T, conn *sql.Conn) {
	ctx := context.Background()

	// Create some chunks that will be collected into oldgen because they're referenced by a commit
	_, err := conn.ExecContext(ctx, "create table vals (id bigint primary key, val bigint)")
	require.NoError(t, err)
	vals := []string{}
	for i := 0; i <= 1024; i++ {
		vals = append(vals, fmt.Sprintf("(%d,0)", i))
	}
	_, err = conn.ExecContext(ctx, "insert into vals values "+strings.Join(vals, ","))
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, "select dolt_commit('-Am', 'create vals table')")
	require.NoError(t, err)

	// Create some chunks that will be collected into newgen because they aren't referenced by a commit
	_, err = conn.ExecContext(ctx, "create table vals2 (id bigint primary key, val bigint)")
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, "insert into vals2 values "+strings.Join(vals, ","))
	require.NoError(t, err)
}

type RepoSize struct {
	Journal int64
	NewGen  int64
	NewGenC int
	OldGen  int64
	OldGenC int
}

func GetRepoSize(dir string) (RepoSize, error) {
	skipFile := func(filename string) bool {
		return filename == "manifest" || filename == "LOCK"
	}
	var ret RepoSize
	entries, err := os.ReadDir(filepath.Join(dir, ".dolt/noms"))
	if err != nil {
		return ret, err
	}
	for _, e := range entries {
		if skipFile(e.Name()) {
			continue
		}
		stat, err := e.Info()
		if err != nil {
			return ret, err
		}
		if stat.IsDir() {
			continue
		}
		ret.NewGen += stat.Size()
		ret.NewGenC += 1
		if e.Name() == "vvvvvvvvvvvvvvvvvvvvvvvvvvvvvvvv" {
			ret.Journal += stat.Size()
		}
	}
	entries, err = os.ReadDir(filepath.Join(dir, ".dolt/noms/oldgen"))
	if err != nil {
		return ret, err
	}
	for _, e := range entries {
		if skipFile(e.Name()) {
			continue
		}
		stat, err := e.Info()
		if err != nil {
			return ret, err
		}
		if stat.IsDir() {
			continue
		}
		ret.OldGen += stat.Size()
		ret.OldGenC += 1
	}
	return ret, nil
}

func TestResumableGC(t *testing.T) {
	t.Parallel()
	ports := newPorts(t)

	u, err := driver.NewDoltUser()
	require.NoError(t, err)
	t.Cleanup(func() { u.Cleanup() })

	rs, err := u.MakeRepoStore()
	require.NoError(t, err)

	repoName := "resumable_gc_test"
	repo, err := rs.MakeRepo(repoName)
	require.NoError(t, err)

	ctx := context.Background()

	server := StartServer(t, rs, repoName, &driver.Server{
		Args:        []string{},
		DynamicPort: "server",
		Envs:        []string{"DOLT_TEST_ABORT_GC_AFTER_INCREMENTAL_FILE_WRITE=true"},
	}, ports)

	// Create the database and run GC
	func() {
		db, err := server.DB(driver.Connection{})
		require.NoError(t, err)
		defer db.Close()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()

		populateDB(t, conn)

		// Test 1: Oldgen is not amended when --full is set
		_, err = conn.ExecContext(ctx, "select dolt_gc('--archive-level','0','--incremental-file-size','1','--full')")
		require.Errorf(t, err, "GC aborting after writing incremental table file")
		requireFileAddedToManifest(t, filepath.Join(repo.Dir, ".dolt/noms/oldgen"), false)

		// Test 2: Oldgen is amended when --full is not set
		_, err = conn.ExecContext(ctx, "select dolt_gc('--archive-level','0','--incremental-file-size','1')")
		require.Errorf(t, err, "GC aborting after writing incremental table file")
		requireFileAddedToManifest(t, filepath.Join(repo.Dir, ".dolt/noms/oldgen"), true)

		_, err = conn.ExecContext(ctx, "select dolt_gc('--archive-level','0','--incremental-file-size','1')")

		// Test 3: Newgen is not amended even when --full is not set (but there are no changes to oldgen)
		_, err = conn.ExecContext(ctx, "insert into vals2 values (1025, 0)")
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, "select dolt_gc('--archive-level','0','--incremental-file-size','1')")
		require.Errorf(t, err, "GC aborting after writing incremental table file")
		requireFileAddedToManifest(t, filepath.Join(repo.Dir, ".dolt/noms/oldgen"), false)
	}()
}

func requireFileAddedToManifest(t *testing.T, path string, expectedResult bool) {
	// Verify that the manifest contains the new chunk file

	manifest, err := os.Open(filepath.Join(path, "manifest"))
	if os.IsNotExist(err) {
		// A nonexistent manifest contains no files
		require.False(t, expectedResult)
		return
	}
	require.NoError(t, err)

	entries, err := os.ReadDir(path)
	require.NoError(t, err)
	var tablefileName string
	for _, e := range entries {
		if e.Name() == "manifest" || e.Name() == "LOCK" {
			continue
		}
		if e.IsDir() {
			continue
		}

		tablefileName = e.Name()
		break
	}
	require.NotEmpty(t, tablefileName)

	manifestBuffer := make([]byte, 200)
	_, err = manifest.Read(manifestBuffer)
	require.NoError(t, err)

	if expectedResult {
		require.Contains(t, string(manifestBuffer), tablefileName)
	} else {
		require.NotContains(t, string(manifestBuffer), tablefileName)
	}
}
