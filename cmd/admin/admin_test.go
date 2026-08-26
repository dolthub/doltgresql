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
	"testing"
	"time"

	"github.com/dolthub/dolt/go/gen/fb/serial"
	"github.com/dolthub/dolt/go/libraries/doltcore/dbfactory"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/doltcore/ref"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/val"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core/integrity"
	dserver "github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/servercfg"
	"github.com/dolthub/doltgresql/servercfg/cfgdetails"
)

const testDbName = "corruption_test"

// TestReportAndRepairEndToEnd creates a doltgres database with out-of-band adaptive values, simulates
// the historical serializer bug by rewriting the entire commit history with value_address_offsets
// stripped from every leaf node, then exercises report mode and repair mode. Since the repair
// re-serializes the corrupt leaves with the current serializer using their original tuple bytes, a
// successful repair restores the exact pre-corruption chunks, and therefore the exact pre-corruption
// commit hashes.
func TestReportAndRepairEndToEnd(t *testing.T) {
	ctx := context.Background()
	dataDir, err := os.MkdirTemp(os.TempDir(), "admin_corruption_test")
	require.NoError(t, err)
	defer os.RemoveAll(dataDir)

	createTestDatabase(t, ctx, dataDir)

	// ------- baseline: freshly written data is healthy -------
	db, cleanup := openTestDatabase(t, ctx, dataDir)
	sctx := db.sctx

	baseline := scanAllBranches(t, sctx, db)
	assertTableStats(t, baseline, "main", "t1", expectStats{rows: 1007, adaptive: 1007, oob: 7})
	assertTableStats(t, baseline, "main", "t4", expectStats{rows: 3, adaptive: 3, oob: 2})
	assertTableStats(t, baseline, "main", "t5", expectStats{rows: 2, adaptive: 2, oob: 2})
	assertTableStats(t, baseline, "b2", "t1", expectStats{rows: 1005, adaptive: 1005, oob: 5})
	assertTableStats(t, baseline, "b3", "t1", expectStats{rows: 1008, adaptive: 1008, oob: 8})
	assertKeyStats(t, baseline, "main", "t3", 3, 3)
	assertNotScanned(t, baseline, "main", "t2")

	origHeads := branchHeadHashes(t, sctx, db)
	require.Len(t, origHeads, 3)
	origOOBAddrs := tableOOBAddrs(t, sctx, db, "main", "t1")
	require.Len(t, origOOBAddrs, 7)

	// t1 is large enough that its map is a multi-level tree, so repair exercises internal node
	// rewrites as well as leaf rewrites.
	require.Greater(t, mapHeight(t, sctx, db, "main", "t1"), 1)

	// ------- corrupt every branch and commit -------
	corrupter := newRepairer(db, testing.Verbose())
	corrupter.transformLeaf = corruptLeafTransform
	summary, err := corrupter.repairDatabase(sctx)
	require.NoError(t, err)
	// main has three user commits (t2-only, initial data, extra t1 rows) plus repository
	// initialization commits; b2 shares all but the last, and b3 adds one divergent commit.
	// Only the three commits containing out-of-band adaptive data are rewritten.
	require.GreaterOrEqual(t, summary.CommitsExamined, 5)
	require.Equal(t, 3, summary.CommitsRewritten, "only commits with out-of-band data are rewritten")
	require.Equal(t, 3, summary.BranchesUpdated)
	// main's working set has uncommitted changes that must be preserved through the branch ref
	// update; b2's and b3's working sets match their heads and are handled by the branch ref
	// update itself.
	require.Equal(t, 1, summary.WorkingSetsFixed)
	require.NotZero(t, summary.LeafChunksRewritten)
	require.NotZero(t, summary.InternalChunksRewritten)
	cleanup(t)

	// ------- report mode sees the corruption -------
	db, cleanup = openTestDatabase(t, ctx, dataDir)
	sctx = db.sctx

	corrupted := scanAllBranches(t, sctx, db)
	assertTableStats(t, corrupted, "main", "t1", expectStats{rows: 1007, adaptive: 1007, oob: 7, corruptVals: 7, corruptRows: 7})
	assertTableStats(t, corrupted, "main", "t4", expectStats{rows: 3, adaptive: 3, oob: 2, corruptVals: 2, corruptRows: 2})
	assertTableStats(t, corrupted, "main", "t5", expectStats{rows: 2, adaptive: 2, oob: 2, corruptVals: 2, corruptRows: 2})
	assertTableStats(t, corrupted, "b2", "t1", expectStats{rows: 1005, adaptive: 1005, oob: 5, corruptVals: 5, corruptRows: 5})
	assertTableStats(t, corrupted, "b3", "t1", expectStats{rows: 1008, adaptive: 1008, oob: 8, corruptVals: 8, corruptRows: 8})
	assertKeyStats(t, corrupted, "main", "t3", 3, 3)

	// The corrupt commits are new commits; the clean initial commit keeps its hash.
	corruptHeads := branchHeadHashes(t, sctx, db)
	require.NotEqual(t, origHeads["main"], corruptHeads["main"])
	require.NotEqual(t, origHeads["b2"], corruptHeads["b2"])
	require.Equal(t, rootCommitHash(t, sctx, db, "main"), rootCommitHash(t, sctx, db, "b2"))

	// The reachability walk (the same walk used by push, clone, and GC) no longer reaches the
	// out-of-band chunks: this is exactly why the corruption loses data.
	walked := walkedAddrs(t, sctx, db, "main", "t1")
	for addr := range origOOBAddrs {
		require.False(t, walked.Has(addr), "corrupt tree should not reach out-of-band chunk %s", addr)
	}

	// The startup integrity check refuses the corrupt database.
	checkErr := integrity.CheckDatabase(sctx, testDbName, db.ddb)
	var corruptionErr *integrity.CorruptionError
	require.ErrorAs(t, checkErr, &corruptionErr)
	require.Equal(t, testDbName, corruptionErr.Database)
	require.NotZero(t, corruptionErr.Stats.CorruptValues)
	require.Contains(t, checkErr.Error(), "BACKUP")
	cleanup(t)

	// ------- repair mode via the CLI entry point -------
	reportPath := filepath.Join(dataDir, "report.html")
	err = run([]string{"repair", "-dir", dataDir, "-out", reportPath})
	require.NoError(t, err)
	html, err := os.ReadFile(reportPath)
	require.NoError(t, err)
	require.Contains(t, string(html), "Repair summary")
	require.Contains(t, string(html), "Post-repair verification scan")

	// ------- verify the repair -------
	db, cleanup = openTestDatabase(t, ctx, dataDir)
	defer cleanup(t)
	sctx = db.sctx

	repaired := scanAllBranches(t, sctx, db)
	assertTableStats(t, repaired, "main", "t1", expectStats{rows: 1007, adaptive: 1007, oob: 7})
	assertTableStats(t, repaired, "main", "t4", expectStats{rows: 3, adaptive: 3, oob: 2})
	assertTableStats(t, repaired, "main", "t5", expectStats{rows: 2, adaptive: 2, oob: 2})
	assertTableStats(t, repaired, "b2", "t1", expectStats{rows: 1005, adaptive: 1005, oob: 5})
	assertTableStats(t, repaired, "b3", "t1", expectStats{rows: 1008, adaptive: 1008, oob: 8})
	assertKeyStats(t, repaired, "main", "t3", 3, 3)

	// Repair rebuilds the exact chunks the fixed serializer originally wrote, so the rewritten
	// commit history converges back to the original commit hashes.
	repairedHeads := branchHeadHashes(t, sctx, db)
	require.Equal(t, origHeads, repairedHeads)

	// The startup integrity check accepts the repaired database.
	require.NoError(t, integrity.CheckDatabase(sctx, testDbName, db.ddb))

	// The reachability walk reaches every out-of-band chunk again.
	walked = walkedAddrs(t, sctx, db, "main", "t1")
	for addr := range origOOBAddrs {
		require.True(t, walked.Has(addr), "repaired tree should reach out-of-band chunk %s", addr)
	}
}

// TestRepairedDatabaseIsServable restarts a server against a corrupted-then-repaired database and
// reads all the adaptive values back through SQL.
func TestRepairedDatabaseIsServable(t *testing.T) {
	ctx := context.Background()
	dataDir, err := os.MkdirTemp(os.TempDir(), "admin_servable_test")
	require.NoError(t, err)
	defer os.RemoveAll(dataDir)

	createTestDatabase(t, ctx, dataDir)

	db, cleanup := openTestDatabase(t, ctx, dataDir)
	sctx := db.sctx
	corrupter := newRepairer(db, testing.Verbose())
	corrupter.transformLeaf = corruptLeafTransform
	_, err = corrupter.repairDatabase(sctx)
	require.NoError(t, err)
	cleanup(t)

	require.NoError(t, run([]string{"repair", "-dir", dataDir, "-out", filepath.Join(dataDir, "report.html")}))

	port := freePort(t)
	controller := startServer(t, ctx, dataDir, port)
	defer stopServer(t, controller)

	conn := connectDb(t, ctx, port, testDbName)
	defer conn.Close(ctx)

	// 7 committed out-of-band values plus 2 uncommitted ones preserved in the working set
	var count int
	require.NoError(t, conn.QueryRow(ctx, "SELECT count(*) FROM t1 WHERE length(big) = 20000").Scan(&count))
	require.Equal(t, 9, count)
	require.NoError(t, conn.QueryRow(ctx, "SELECT count(*) FROM t3 WHERE length(big) = 20000").Scan(&count))
	require.Equal(t, 3, count)
	require.NoError(t, conn.QueryRow(ctx, "SELECT count(*) FROM t4 WHERE length(j::text) > 20000").Scan(&count))
	require.Equal(t, 2, count)

	var b2Count int
	require.NoError(t, conn.QueryRow(ctx, "SELECT count(*) FROM t1 AS OF 'b2' WHERE length(big) = 20000").Scan(&b2Count))
	require.Equal(t, 5, b2Count)
}

// TestGenerateCorruptedDataDir is a manual-testing helper, not a regression test: it builds a corrupted
// data directory at $ADMIN_TEST_GEN_DIR for exercising the compiled admin binary by hand (e.g.
// `go build ./cmd/admin && ./admin report -dir <dir>`). Skipped unless the environment variable is set.
func TestGenerateCorruptedDataDir(t *testing.T) {
	dataDir := os.Getenv("ADMIN_TEST_GEN_DIR")
	if dataDir == "" {
		t.Skip("set ADMIN_TEST_GEN_DIR to generate a corrupted data dir for manual testing")
	}
	if !filepath.IsAbs(dataDir) {
		// `go test` runs the test binary with the package directory as its working directory,
		// so a relative path would resolve somewhere inside the repository, not relative to the
		// caller's shell.
		t.Fatalf("ADMIN_TEST_GEN_DIR must be an absolute path, got %q", dataDir)
	}
	ctx := context.Background()
	require.NoError(t, os.MkdirAll(dataDir, 0755))
	createTestDatabase(t, ctx, dataDir)

	db, cleanup := openTestDatabase(t, ctx, dataDir)
	defer cleanup(t)
	corrupter := newRepairer(db, true)
	corrupter.transformLeaf = corruptLeafTransform
	_, err := corrupter.repairDatabase(db.sctx)
	require.NoError(t, err)
}

// corruptLeafTransform simulates the bug in old releases: it re-serializes leaf nodes with their
// value_address_offsets field omitted. This matches the behavior of releases where the tuple
// descriptor's AddressFieldCount did not count adaptive-encoded fields.
func corruptLeafTransform(r *repairer, pm *serial.ProllyTreeNode, kd, vd *val.TupleDesc) (serial.Message, bool, error) {
	if pm.ValueAddressOffsetsLength() == 0 {
		return nil, false, nil
	}
	msg, err := reserializeLeaf(pm, strippedDescriptor(vd), r.db.ns.Pool())
	if err != nil {
		return nil, false, err
	}
	return msg, true, nil
}

// strippedDescriptor returns a copy of the descriptor with all address and adaptive encodings replaced
// by a plain encoding, so the serializer records no value address offsets, mirroring the historical bug.
// Extended encodings are also replaced: they would require type handlers, which the serializer never
// consults, and which plain descriptor construction rejects.
func strippedDescriptor(vd *val.TupleDesc) *val.TupleDesc {
	types := make([]val.Type, len(vd.Types))
	for i, typ := range vd.Types {
		if val.IsAdaptiveEncoding(typ.Enc) || val.IsAddrEncoding(typ.Enc) || val.IsExtendedEncoding(typ.Enc) {
			typ.Enc = val.StringEnc
		}
		types[i] = typ
	}
	return val.NewTupleDescriptor(types...)
}

// ------- database construction -------

// createTestDatabase creates a doltgres database with:
//   - an initial commit containing only t2 (no adaptive columns), which stays clean throughout
//   - t1: int primary key with a text column; 5 out-of-band + 1000 inline values at the second commit,
//     2 more out-of-band values in a third commit on main only
//   - t3: text primary key (out-of-band adaptive values in key tuples, which are reported but untracked)
//   - t4: jsonb column with 2 out-of-band + 1 inline value
//   - branch b2 at the second commit; branch b3 diverging from it with 3 more out-of-band values
func createTestDatabase(t *testing.T, ctx context.Context, dataDir string) {
	port := freePort(t)
	controller := startServer(t, ctx, dataDir, port)

	conn := connectDb(t, ctx, port, "postgres")
	_, err := conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s", testDbName))
	require.NoError(t, err)
	require.NoError(t, conn.Close(ctx))

	conn = connectDb(t, ctx, port, testDbName)
	stmts := []string{
		"CREATE TABLE t2 (id int primary key, n int)",
		"INSERT INTO t2 VALUES (1, 1), (2, 2)",
		"SELECT dolt_commit('-Am', 'commit 0: no adaptive data')",

		"CREATE TABLE t1 (id int primary key, big text)",
		// ids 1-5: out-of-band values; ids 101-1100: small inline values (enough rows that the
		// table's prolly tree has internal nodes)
		"INSERT INTO t1 SELECT i, rpad(i::text, 20000, 'x') FROM generate_series(1, 5) AS g(i)",
		"INSERT INTO t1 SELECT i, 'small-' || i::text FROM generate_series(101, 1100) AS g(i)",
		// t3 has an adaptive-encoded primary key column: out-of-band key values cannot be recorded
		// in the message format, so they are reported but not repaired.
		"CREATE TABLE t3 (big text primary key, n int)",
		"INSERT INTO t3 SELECT rpad(i::text, 20000, 'z'), i FROM generate_series(1, 3) AS g(i)",
		"CREATE TABLE t4 (id int primary key, j jsonb)",
		`INSERT INTO t4 SELECT i, ('{"k": "' || rpad(i::text, 20000, 'j') || '"}')::jsonb FROM generate_series(1, 2) AS g(i)`,
		`INSERT INTO t4 VALUES (3, '{"k": "small"}')`,
		// t5 has a user-defined type column: deserializing its schema requires resolving the type
		// through a real session (a naked sql.Context panics in DSessFromSess).
		"CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy')",
		"CREATE TABLE t5 (id int primary key, m mood, big text)",
		"INSERT INTO t5 SELECT i, 'happy', rpad(i::text, 20000, 'e') FROM generate_series(1, 2) AS g(i)",
		"SELECT dolt_commit('-Am', 'commit 1: initial data')",
		"SELECT dolt_branch('b2')",

		// b3 diverges from commit 1 with its own out-of-band values (ids 11-13)
		"SELECT dolt_checkout('-b', 'b3')",
		"INSERT INTO t1 SELECT i, rpad(i::text, 20000, 'v') FROM generate_series(11, 13) AS g(i)",
		"SELECT dolt_commit('-Am', 'commit on b3: divergent out-of-band values')",
		"SELECT dolt_checkout('main')",

		// ids 6-7: out-of-band values only present at main's head
		"INSERT INTO t1 SELECT i, rpad(i::text, 20000, 'y') FROM generate_series(6, 7) AS g(i)",
		"SELECT dolt_commit('-Am', 'commit 2: more out-of-band values')",

		// ids 8-9: uncommitted out-of-band values, present only in main's working set
		"INSERT INTO t1 SELECT i, rpad(i::text, 20000, 'w') FROM generate_series(8, 9) AS g(i)",
	}
	for _, stmt := range stmts {
		_, err = conn.Exec(ctx, stmt)
		require.NoError(t, err, "statement: %s", stmt)
	}
	require.NoError(t, conn.Close(ctx))
	stopServer(t, controller)
}

// ------- offline helpers -------

// openTestDatabase opens the test database offline and returns it along with a cleanup function.
func openTestDatabase(t *testing.T, ctx context.Context, dataDir string) (*database, func(*testing.T)) {
	dbs, cleanup, err := openDatabases(ctx, dataDir)
	require.NoError(t, err)
	for _, db := range dbs {
		if db.name == testDbName {
			return db, func(t *testing.T) {
				cleanup()
				require.NoError(t, dbfactory.CloseAllLocalDatabases())
			}
		}
	}
	cleanup()
	t.Fatalf("database %s not found in %s", testDbName, dataDir)
	return nil, nil
}

// scanAllBranches scans every branch head with a fresh scanner and indexes the results by branch and
// table name.
func scanAllBranches(t *testing.T, sctx *sql.Context, db *database) map[string]map[string]*tableReport {
	branches, err := scanBranchHeads(sctx, db, integrity.NewScanner(db.cs), testing.Verbose())
	require.NoError(t, err)
	result := make(map[string]map[string]*tableReport)
	for _, br := range branches {
		tables := make(map[string]*tableReport)
		for _, tr := range br.Tables {
			require.Empty(t, tr.Err, "table %s.%s on branch %s", tr.Schema, tr.Table, br.Branch)
			tables[tr.Table] = tr
		}
		result[br.Branch] = tables
	}
	return result
}

type expectStats struct {
	rows        uint64
	adaptive    uint64
	oob         uint64
	corruptVals uint64
	corruptRows uint64
}

func assertTableStats(t *testing.T, scans map[string]map[string]*tableReport, branch, table string, expect expectStats) {
	t.Helper()
	tr := scans[branch][table]
	require.NotNil(t, tr, "table %s on branch %s", table, branch)
	require.NotNil(t, tr.Stats, "table %s on branch %s was not scanned", table, branch)
	require.Equal(t, expect.rows, tr.Stats.Rows, "rows of %s on %s", table, branch)
	require.Equal(t, expect.adaptive, tr.Stats.AdaptiveValues, "adaptive values of %s on %s", table, branch)
	require.Equal(t, expect.oob, tr.Stats.OutOfBandValues, "out-of-band values of %s on %s", table, branch)
	require.Equal(t, expect.corruptVals, tr.Stats.CorruptValues, "corrupt values of %s on %s", table, branch)
	require.Equal(t, expect.corruptRows, tr.Stats.CorruptRows, "corrupt rows of %s on %s", table, branch)
	require.Zero(t, tr.Stats.UnexpectedOffsets, "unexpected offsets of %s on %s", table, branch)
	require.Zero(t, tr.Stats.MissingChunks, "missing chunks of %s on %s", table, branch)
	if expect.corruptVals > 0 {
		require.NotZero(t, tr.Stats.CorruptChunks, "corrupt chunks of %s on %s", table, branch)
	} else {
		require.Zero(t, tr.Stats.CorruptChunks, "corrupt chunks of %s on %s", table, branch)
	}
}

// assertKeyStats verifies the key-side adaptive statistics of a table whose primary key contains
// adaptive-encoded columns.
func assertKeyStats(t *testing.T, scans map[string]map[string]*tableReport, branch, table string, keyAdaptive, keyOOB uint64) {
	t.Helper()
	tr := scans[branch][table]
	require.NotNil(t, tr, "table %s on branch %s", table, branch)
	require.NotNil(t, tr.Stats, "table %s on branch %s was not scanned", table, branch)
	require.True(t, tr.KeyImpacted)
	require.Equal(t, keyAdaptive, tr.Stats.KeyAdaptiveValues, "key adaptive values of %s on %s", table, branch)
	require.Equal(t, keyOOB, tr.Stats.KeyOutOfBandValues, "key out-of-band values of %s on %s", table, branch)
	require.Zero(t, tr.Stats.CorruptValues)
	require.Zero(t, tr.Stats.MissingChunks, "missing chunks of %s on %s", table, branch)
}

func assertNotScanned(t *testing.T, scans map[string]map[string]*tableReport, branch, table string) {
	t.Helper()
	tr := scans[branch][table]
	require.NotNil(t, tr, "table %s on branch %s", table, branch)
	require.False(t, tr.Impacted)
	require.False(t, tr.KeyImpacted)
	require.Nil(t, tr.Stats)
}

// branchHeadHashes returns the head commit hash of every branch.
func branchHeadHashes(t *testing.T, sctx *sql.Context, db *database) map[string]hash.Hash {
	branches, err := db.ddb.GetBranches(sctx)
	require.NoError(t, err)
	heads := make(map[string]hash.Hash)
	for _, branchRef := range branches {
		cm, err := db.ddb.ResolveCommitRef(sctx, branchRef)
		require.NoError(t, err)
		h, err := cm.HashOf()
		require.NoError(t, err)
		heads[branchRef.GetPath()] = h
	}
	return heads
}

// rootCommitHash returns the hash of the initial (parentless) commit reachable from the given branch.
func rootCommitHash(t *testing.T, sctx *sql.Context, db *database, branch string) hash.Hash {
	cm, err := db.ddb.ResolveCommitRef(sctx, ref.NewBranchRef(branch))
	require.NoError(t, err)
	for cm.NumParents() > 0 {
		optParents, err := db.ddb.ResolveAllParents(sctx, cm)
		require.NoError(t, err)
		parent, ok := optParents[0].ToCommit()
		require.True(t, ok)
		cm = parent
	}
	h, err := cm.HashOf()
	require.NoError(t, err)
	return h
}

// tableAtBranchHead returns the tableInfo of the named table at the head of the given branch.
func tableAtBranchHead(t *testing.T, sctx *sql.Context, db *database, branch, table string) *integrity.TableInfo {
	cm, err := db.ddb.ResolveCommitRef(sctx, ref.NewBranchRef(branch))
	require.NoError(t, err)
	root, err := cm.GetRootValue(sctx)
	require.NoError(t, err)
	tables, err := integrity.TablesForRoot(sctx, root, db.ns)
	require.NoError(t, err)
	for _, ti := range tables {
		if ti.Name.Name == table {
			return ti
		}
	}
	t.Fatalf("table %s not found on branch %s", table, branch)
	return nil
}

// tableOOBAddrs returns the set of out-of-band chunk addresses referenced from the value tuples of the
// named table at the head of the given branch.
func tableOOBAddrs(t *testing.T, sctx *sql.Context, db *database, branch, table string) hash.HashSet {
	ti := tableAtBranchHead(t, sctx, db, branch, table)
	m, err := ti.RowMap(sctx)
	require.NoError(t, err)

	addrs := hash.NewHashSet()
	var walk func(addr hash.Hash)
	walk = func(addr hash.Hash) {
		msg, err := integrity.GetTreeNodeMessage(sctx, db.cs, addr)
		require.NoError(t, err)
		var pm serial.ProllyTreeNode
		require.NoError(t, serial.InitProllyTreeNodeRoot(&pm, msg, serial.MessagePrefixSz))
		if pm.TreeLevel() > 0 {
			for _, child := range integrity.ChildAddresses(&pm) {
				walk(child)
			}
			return
		}
		la, err := integrity.AnalyzeLeaf(&pm, ti.KeyDesc, ti.ValDesc)
		require.NoError(t, err)
		for _, a := range la.ValueOOBAddrs {
			addrs.Insert(a)
		}
	}
	walk(m.Node().HashOf())
	return addrs
}

// walkedAddrs returns every chunk address reachable from the named table's row map via the standard
// tree address walk: the same walk used by push, clone, and garbage collection.
func walkedAddrs(t *testing.T, sctx *sql.Context, db *database, branch, table string) hash.HashSet {
	ti := tableAtBranchHead(t, sctx, db, branch, table)
	m, err := ti.RowMap(sctx)
	require.NoError(t, err)

	addrs := hash.NewHashSet()
	err = tree.WalkAddresses(sctx, m.Node(), db.ns, func(_ context.Context, addr hash.Hash) error {
		addrs.Insert(addr)
		return nil
	})
	require.NoError(t, err)
	return addrs
}

// mapHeight returns the height of the named table's row map.
func mapHeight(t *testing.T, sctx *sql.Context, db *database, branch, table string) int {
	ti := tableAtBranchHead(t, sctx, db, branch, table)
	m, err := ti.RowMap(sctx)
	require.NoError(t, err)
	return m.Height()
}

// ------- server lifecycle -------

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
	// until closed. A separate admin process would not see them; this in-process test must drop
	// them explicitly, or subsequent opens are forced into read-only mode.
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
