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
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/dolthub/dolt/go/libraries/doltcore/dbfactory"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb/durable"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/doltcore/ref"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/dolt/go/store/chunks"
	"github.com/dolthub/dolt/go/store/datas"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/nbs"
	"github.com/dolthub/dolt/go/store/val"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	dserver "github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/servercfg"
	"github.com/dolthub/doltgresql/servercfg/cfgdetails"
)

const testDbName = "corruption_test"

// TestReportAndRepairEndToEnd creates a doltgres database with out-of-band adaptive values,
// simulates the GC bug by filtering their blob chunks out of the chunk journal, then exercises
// report mode and repair mode, and finally restarts a server against the repaired database to
// verify it is fully readable again.
func TestReportAndRepairEndToEnd(t *testing.T) {
	ctx := context.Background()
	dataDir, err := os.MkdirTemp(os.TempDir(), "admin_corruption_test")
	require.NoError(t, err)
	defer os.RemoveAll(dataDir)

	createCorruptedDatabase(t, ctx, dataDir)

	// ------- report mode -------
	sqlCtx := sql.NewContext(ctx)
	mrEnv, err := loadMultiEnv(ctx, dataDir)
	require.NoError(t, err)
	dEnv := mrEnv.GetEnv(testDbName)
	require.NotNil(t, dEnv)

	dbRep, err := reportDatabase(sqlCtx, testDbName, dEnv.DoltDB(ctx), "")
	require.NoError(t, err)
	require.Empty(t, dbRep.Errors)
	require.Len(t, dbRep.Branches, 2)

	for _, br := range dbRep.Branches {
		require.Len(t, br.Tables, 2, "t1 and t3 have adaptive columns")
		require.EqualValues(t, 1, br.SkippedTables, "t2 has no adaptive columns")
		byName := make(map[string]*TableStats)
		for _, ts := range br.Tables {
			byName[ts.Table] = ts
		}

		t1 := byName["public.t1"]
		require.NotNil(t, t1)
		require.Equal(t, []string{"big"}, t1.AdaptiveColumns)
		require.EqualValues(t, 2, t1.MissingValues, "branch %s", br.Branch)
		require.EqualValues(t, 2, t1.RowsWithMissing, "branch %s", br.Branch)
		require.EqualValues(t, 0, t1.MissingKeyValues)
		switch br.Branch {
		case "main":
			require.EqualValues(t, 12, t1.RowsScanned)
			require.EqualValues(t, 12, t1.AdaptiveValues)
			require.EqualValues(t, 7, t1.OutOfBandValues)
		case "b2":
			require.EqualValues(t, 10, t1.RowsScanned)
			require.EqualValues(t, 10, t1.AdaptiveValues)
			require.EqualValues(t, 5, t1.OutOfBandValues)
		default:
			t.Fatalf("unexpected branch %s", br.Branch)
		}

		t3 := byName["public.t3"]
		require.NotNil(t, t3)
		require.Equal(t, []string{"big"}, t3.AdaptiveKeyColumns)
		require.EqualValues(t, 3, t3.RowsScanned)
		require.EqualValues(t, 3, t3.AdaptiveValues)
		require.EqualValues(t, 3, t3.OutOfBandValues)
		require.EqualValues(t, 0, t3.MissingValues)
		require.EqualValues(t, 1, t3.MissingKeyValues, "branch %s", br.Branch)
		require.EqualValues(t, 1, t3.RowsWithMissing)
	}

	// -branch restricts the scan to a single branch
	dbRep, err = reportDatabase(sqlCtx, testDbName, dEnv.DoltDB(ctx), "b2")
	require.NoError(t, err)
	require.Empty(t, dbRep.Errors)
	require.Len(t, dbRep.Branches, 1)
	require.Equal(t, "b2", dbRep.Branches[0].Branch)

	// a branch that doesn't exist is surfaced as a database error
	dbRep, err = reportDatabase(sqlCtx, testDbName, dEnv.DoltDB(ctx), "no_such_branch")
	require.NoError(t, err)
	require.Len(t, dbRep.Errors, 1)
	require.Contains(t, dbRep.Errors[0], "no_such_branch")
	require.Empty(t, dbRep.Branches)

	require.NoError(t, mrEnv.Close(ctx))

	// ------- repair mode -------
	mrEnv, err = loadMultiEnv(ctx, dataDir)
	require.NoError(t, err)
	dEnv = mrEnv.GetEnv(testDbName)
	require.NotNil(t, dEnv)

	dbRep, err = repairDatabase(sqlCtx, testDbName, dEnv.DoltDB(ctx), "")
	require.NoError(t, err)
	// t3's key corruption is unrepairable and must be surfaced as errors; nothing else may error.
	require.NotEmpty(t, dbRep.Errors)
	for _, e := range dbRep.Errors {
		require.Contains(t, e, "public.t3")
		require.Contains(t, e, "primary key")
	}
	var repaired uint64
	for _, br := range dbRep.Branches {
		require.GreaterOrEqual(t, br.CommitsScanned, uint64(1))
		for _, ts := range br.Tables {
			repaired += ts.RepairedValues
		}
	}
	require.EqualValues(t, 4, repaired, "expected 2 values repaired at each branch head lineage (t1 in commit 1 and commit 2)")
	require.NoError(t, mrEnv.Close(ctx))

	// ------- verify: a fresh report over the repaired database finds nothing missing, and a
	// full history walk (a second repair run) changes nothing -------
	mrEnv, err = loadMultiEnv(ctx, dataDir)
	require.NoError(t, err)
	dEnv = mrEnv.GetEnv(testDbName)
	dbRep, err = repairDatabase(sqlCtx, testDbName, dEnv.DoltDB(ctx), "")
	require.NoError(t, err)
	for _, e := range dbRep.Errors {
		require.Contains(t, e, "public.t3", "only t3's unrepairable key corruption may remain")
	}
	for _, br := range dbRep.Branches {
		for _, ts := range br.Tables {
			require.EqualValues(t, 0, ts.MissingValues, "branch %s table %s still missing values after repair", br.Branch, ts.Table)
			if ts.Table == "public.t3" {
				require.EqualValues(t, 1, ts.MissingKeyValues, "t3's key corruption is unrepairable and must persist")
				require.EqualValues(t, 0, ts.RepairedValues)
			} else {
				require.EqualValues(t, 0, ts.MissingKeyValues)
			}
		}
	}
	verifyT1Index(t, sqlCtx, dEnv.DoltDB(ctx))
	require.NoError(t, mrEnv.Close(ctx))

	// ------- verify: the repaired database is fully readable through a server again -------
	port := freePort(t)
	controller := startServer(t, ctx, dataDir, port)
	conn := connectDb(t, ctx, port, testDbName)

	var nullCount, bigCount int
	require.NoError(t, conn.QueryRow(ctx, "SELECT count(*) FROM t1 WHERE big IS NULL").Scan(&nullCount))
	require.Equal(t, 2, nullCount, "the two corrupted values should now be NULL")
	require.NoError(t, conn.QueryRow(ctx, "SELECT count(*) FROM t1 WHERE length(big) = 20000").Scan(&bigCount))
	require.Equal(t, 5, bigCount, "the healthy out-of-band values should be intact")

	// reading every value exercises the out-of-band resolution path for every row
	rows, err := conn.Query(ctx, "SELECT id, big FROM t1 ORDER BY id")
	require.NoError(t, err)
	seen := 0
	for rows.Next() {
		var id int
		var big *string
		require.NoError(t, rows.Scan(&id, &big))
		if id == 1 || id == 2 {
			require.Nil(t, big, "row %d should have been NULLed", id)
		} else {
			require.NotNil(t, big, "row %d should be readable", id)
		}
		seen++
	}
	require.NoError(t, rows.Err())
	require.Equal(t, 12, seen)

	// the other branch must be readable too
	var b2Nulls int
	require.NoError(t, conn.QueryRow(ctx, "SELECT count(*) FROM t1 AS OF 'b2' WHERE big IS NULL").Scan(&b2Nulls))
	require.Equal(t, 2, b2Nulls)

	// history was rewritten but preserved: both commits still present on main
	var commits int
	require.NoError(t, conn.QueryRow(ctx, "SELECT count(*) FROM dolt_log WHERE message LIKE 'commit %'").Scan(&commits))
	require.Equal(t, 2, commits)

	require.NoError(t, conn.Close(ctx))
	stopServer(t, controller)
}

// createCorruptedDatabase creates a doltgres database under |dataDir| with adaptive-encoded
// values on two branches and two commits, then removes a subset of the out-of-band blob chunks
// from the chunk journal, simulating the GC bug this tool exists to repair. The corrupted
// values are t1.big for ids 1 and 2 (repairable value corruption, covered by a secondary
// index) and one of t3's text primary keys (unrepairable key corruption).
func createCorruptedDatabase(t *testing.T, ctx context.Context, dataDir string) {
	port := freePort(t)
	controller := startServer(t, ctx, dataDir, port)

	conn := connectDb(t, ctx, port, "postgres")
	_, err := conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", testDbName))
	require.NoError(t, err)
	require.NoError(t, conn.Close(ctx))

	conn = connectDb(t, ctx, port, testDbName)
	stmts := []string{
		"CREATE TABLE t1 (id int primary key, big text)",
		"CREATE INDEX t1_big_idx ON t1 (big)",
		// ids 1-5: unique out-of-band values; ids 101-105: small inline values
		"INSERT INTO t1 SELECT i, rpad(i::text, 20000, 'x') FROM generate_series(1, 5) AS g(i)",
		"INSERT INTO t1 SELECT i, 'small-' || i::text FROM generate_series(101, 105) AS g(i)",
		"CREATE TABLE t2 (id int primary key, n int)",
		"INSERT INTO t2 VALUES (1, 1), (2, 2)",
		// t3 has an adaptive-encoded primary key column: corruption in key fields cannot be
		// repaired (keys cannot be NULLed), only reported.
		"CREATE TABLE t3 (big text primary key, n int)",
		"INSERT INTO t3 SELECT rpad(i::text, 20000, 'z'), i FROM generate_series(1, 3) AS g(i)",
		"SELECT dolt_commit('-Am', 'commit 1')",
		"SELECT dolt_branch('b2')",
		// ids 6-7: out-of-band values only present on main's second commit
		"INSERT INTO t1 SELECT i, rpad(i::text, 20000, 'y') FROM generate_series(6, 7) AS g(i)",
		"SELECT dolt_commit('-Am', 'commit 2')",
	}
	for _, stmt := range stmts {
		_, err = conn.Exec(ctx, stmt)
		require.NoError(t, err, "statement: %s", stmt)
	}
	require.NoError(t, conn.Close(ctx))
	stopServer(t, controller)

	// Simulate the GC bug: remove the blob chunks for t1 rows 1 and 2, plus one of t3's
	// primary key values.
	addrs, t3KeyAddrs := collectAddresses(t, ctx, dataDir)
	require.GreaterOrEqual(t, len(addrs), 7)
	ids := make([]int, 0, len(addrs))
	for id := range addrs {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	require.Equal(t, []int{1, 2, 3, 4, 5, 6, 7}, ids, "expected exactly rows 1-7 to be out-of-band")
	require.Len(t, t3KeyAddrs, 3, "expected all three t3 keys to be out-of-band")
	doomed := []hash.Hash{addrs[1], addrs[2], t3KeyAddrs[0]}
	removeChunksFromJournal(t, filepath.Join(dataDir, testDbName), doomed)
}

// TestGenerateCorruptedDataDir is a manual-testing helper, not a regression test: it builds a
// corrupted data directory at $ADMIN_TEST_GEN_DIR for exercising the compiled admin binary by
// hand (e.g. `go build ./cmd/admin && ./admin report -data-dir <dir>`). Skipped unless the
// environment variable is set.
func TestGenerateCorruptedDataDir(t *testing.T) {
	dataDir := os.Getenv("ADMIN_TEST_GEN_DIR")
	if dataDir == "" {
		t.Skip("set ADMIN_TEST_GEN_DIR to generate a corrupted data dir for manual testing")
	}
	require.NoError(t, os.MkdirAll(dataDir, 0755))
	createCorruptedDatabase(t, context.Background(), dataDir)
}

func freePort(t *testing.T) int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	require.NoError(t, l.Close())
	return port
}

func startServer(t *testing.T, ctx context.Context, dataDir string, port int) *svcs.Controller {
	fs, err := filesys.LocalFilesysWithWorkingDir(dataDir)
	require.NoError(t, err)
	dEnv := env.Load(ctx, env.GetCurrentUserHomeDir, fs, doltdb.LocalDirDoltDB, dserver.Version)

	host := "127.0.0.1"
	logLevel := "warn"
	controller, err := dserver.RunOnDisk(ctx, &servercfg.DoltgresConfig{
		DoltgresConfig: cfgdetails.DoltgresConfig{
			ListenerConfig: &cfgdetails.DoltgresListenerConfig{
				PortNumber: &port,
				HostStr:    &host,
			},
			LogLevelStr: &logLevel,
		},
	}, dEnv)
	require.NoError(t, err)
	auth.ClearDatabase()
	return controller
}

func stopServer(t *testing.T, controller *svcs.Controller) {
	controller.Stop()
	require.NoError(t, controller.WaitForStop())
	// The server's databases live in the process-wide singleton cache and keep their file locks
	// until closed. A separate admin process would not see them; this in-process test must
	// drop them explicitly, or subsequent opens are forced into read-only mode.
	require.NoError(t, dbfactory.CloseAllLocalDatabases())
}

func connectDb(t *testing.T, ctx context.Context, port int, database string) *pgx.Conn {
	url := fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/%s", port, database)
	var conn *pgx.Conn
	var err error
	for i := 0; i < 5; i++ {
		conn, err = pgx.Connect(ctx, url)
		if err == nil {
			return conn
		}
		time.Sleep(time.Second)
	}
	require.NoError(t, err)
	return nil
}

// collectAddresses opens the test database offline and returns the out-of-band address of
// t1.big for every row (keyed by row id), plus the out-of-band addresses of t3's primary key
// values, all read from the HEAD of main.
func collectAddresses(t *testing.T, ctx context.Context, dataDir string) (map[int]hash.Hash, []hash.Hash) {
	mrEnv, err := loadMultiEnv(ctx, dataDir)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, mrEnv.Close(ctx))
	}()
	dEnv := mrEnv.GetEnv(testDbName)
	require.NotNil(t, dEnv, "database %s not found", testDbName)

	sqlCtx := sql.NewContext(ctx)
	ddb := dEnv.DoltDB(ctx)
	cm, err := ddb.ResolveCommitRef(sqlCtx, ref.NewBranchRef("main"))
	require.NoError(t, err)
	root, err := cm.GetRootValue(sqlCtx)
	require.NoError(t, err)

	t1Addrs := make(map[int]hash.Hash)
	var t3KeyAddrs []hash.Hash
	err = root.IterTables(sqlCtx, func(name doltdb.TableName, tbl *doltdb.Table, sch schema.Schema) (bool, error) {
		rows, err := tbl.GetRowData(sqlCtx)
		require.NoError(t, err)
		m, err := durable.ProllyMapFromIndex(rows)
		require.NoError(t, err)
		kd, vd := m.Descriptors()
		iter, err := m.IterAll(sqlCtx)
		require.NoError(t, err)

		switch name.Name {
		case "t1":
			valIdxs := adaptiveFieldIndexes(vd)
			require.Len(t, valIdxs, 1, "t1 should have exactly one adaptive value column")
			for {
				k, v, err := iter.Next(sqlCtx)
				if err != nil {
					break
				}
				id, ok := kd.GetInt32(0, k)
				require.True(t, ok)
				field := val.AdaptiveValue(v.GetField(valIdxs[0]))
				if field.IsNull() || !field.IsOutOfBand() {
					continue
				}
				addr, err := field.OutOfBandAddr()
				require.NoError(t, err)
				t1Addrs[int(id)] = addr
			}
		case "t3":
			keyIdxs := adaptiveFieldIndexes(kd)
			require.Len(t, keyIdxs, 1, "t3 should have exactly one adaptive key column")
			for {
				k, _, err := iter.Next(sqlCtx)
				if err != nil {
					break
				}
				field := val.AdaptiveValue(k.GetField(keyIdxs[0]))
				if field.IsNull() || !field.IsOutOfBand() {
					continue
				}
				addr, err := field.OutOfBandAddr()
				require.NoError(t, err)
				t3KeyAddrs = append(t3KeyAddrs, addr)
			}
		}
		return false, nil
	})
	require.NoError(t, err)
	return t1Addrs, t3KeyAddrs
}

// verifyT1Index scans t1's secondary index at the HEAD of main, checking that no index key
// still references a missing chunk and that the two repaired entries are now NULL.
func verifyT1Index(t *testing.T, sqlCtx *sql.Context, ddb *doltdb.DoltDB) {
	cm, err := ddb.ResolveCommitRef(sqlCtx, ref.NewBranchRef("main"))
	require.NoError(t, err)
	root, err := cm.GetRootValue(sqlCtx)
	require.NoError(t, err)
	tbl, ok, err := root.GetTable(sqlCtx, doltdb.TableName{Schema: "public", Name: "t1"})
	require.NoError(t, err)
	require.True(t, ok)

	idxData, err := tbl.GetIndexRowData(sqlCtx, "t1_big_idx")
	require.NoError(t, err)
	m, err := durable.ProllyMapFromIndex(idxData)
	require.NoError(t, err)

	scanner := newDbScanner(datas.ChunkStoreFromDatabase(doltdb.ExposeDatabaseFromDoltDB(ddb)))
	sr, err := scanner.scanMap(sqlCtx, m)
	require.NoError(t, err)
	require.EqualValues(t, 12, sr.RowsScanned, "index should have one entry per row")
	require.EqualValues(t, 0, sr.missingValues())
	require.EqualValues(t, 0, sr.missingKeyValues(), "repaired index keys must not reference missing chunks")

	iter, err := m.IterAll(sqlCtx)
	require.NoError(t, err)
	nulls := 0
	for {
		k, _, err := iter.Next(sqlCtx)
		if err != nil {
			break
		}
		if len(k.GetField(0)) == 0 {
			nulls++
		}
	}
	require.Equal(t, 2, nulls, "the two repaired rows should have NULL index entries")
}

// removeChunksFromJournal rewrites the database's chunk journal without the given chunks,
// simulating the GC bug that dropped out-of-band adaptive value chunks.
func removeChunksFromJournal(t *testing.T, dbDir string, doomed []hash.Hash) {
	nomsDir := filepath.Join(dbDir, ".dolt", "noms")
	journalPath := filepath.Join(nomsDir, chunks.JournalFileID)
	_, err := os.Stat(journalPath)
	require.NoError(t, err, "chunk journal not found at %s", journalPath)

	hashStrs := make([]string, len(doomed))
	for i, h := range doomed {
		hashStrs[i] = h.String()
	}
	result, exitCode := nbs.JournalFilter(journalPath, "", strings.Join(hashStrs, ","))
	require.Equal(t, 0, exitCode, "journal filter failed")
	require.EqualValues(t, len(doomed), result.FilteredRecords, "expected chunks to be filtered")

	require.NoError(t, os.Rename(journalPath+".filtered", journalPath))
	// The journal index is a cache derived from the journal contents; it must not survive the rewrite.
	_ = os.Remove(filepath.Join(nomsDir, "journal.idx"))
}
