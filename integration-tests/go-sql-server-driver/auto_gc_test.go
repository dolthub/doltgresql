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
	"math/rand/v2"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	driver "github.com/dolthub/doltgresql/integration-tests/go-sql-server-driver/driver"
)

// TestAutoGC ports Dolt's driver-level automatic-GC test
// (dolt/integration-tests/go-sql-server-driver/auto_gc_test.go) to Doltgres.
//
// TODO: Only the base, single-server variant is covered here: Dolt's ClusterReplication
// and PushToRemotesAPI subtest variants are not ported yet because Doltgres does not
// implement cluster replication or the remotes API yet (see
// tests/sql-server-cluster.yaml, where every test is skipped for that reason).
func TestAutoGC(t *testing.T) {
	t.Parallel()
	t.Run("Enable", func(t *testing.T) {
		t.Parallel()
		for _, sa := range []struct {
			archive bool
			name    string
		}{{true, "Archive"}, {false, "NoArchive"}} {
			t.Run(sa.name, func(t *testing.T) {
				t.Parallel()
				var s AutoGCTest
				s.Enable = true
				s.Archive = sa.archive
				runAutoGCTestUntilGC(t, &s, 3, 16)
			})
		}
	})
	t.Run("Disabled", func(t *testing.T) {
		t.Parallel()
		var s AutoGCTest
		runAutoGCTestDisabled(t, &s, 64, 16)
	})
}

type AutoGCTest struct {
	Enable  bool
	Archive bool

	Server *driver.SqlServer
	DB     *sql.DB

	Ports *DynamicResources

	gcCount        atomic.Int32
	sawDanglingRef atomic.Bool
}

func (s *AutoGCTest) gcVisitor() func(string) {
	return func(line string) {
		if strings.Contains(line, "Successfully completed auto GC") {
			s.gcCount.Add(1)
		}
		if strings.Contains(line, "dangling references requested during GC") {
			s.sawDanglingRef.Store(true)
		}
	}
}

func (s *AutoGCTest) Setup(ctx context.Context, t *testing.T) {
	s.Ports = newPorts(t)

	u, err := driver.NewDoltUser()
	require.NoError(t, err)
	t.Cleanup(func() {
		u.Cleanup()
	})

	rs, err := u.MakeRepoStore()
	require.NoError(t, err)

	_, err = rs.MakeRepo("auto_gc_test")
	require.NoError(t, err)

	archiveFragment := ``
	if s.Archive {
		archiveFragment = `
    archive_level: 1`
	}

	behaviorFragment := fmt.Sprintf(`
behavior:
  auto_gc_behavior:
    enable: %v%v
listener:
  port: {{get_port "server_port"}}
`, s.Enable, archiveFragment)

	err = driver.WithFile{
		Name:     "server.yaml",
		Contents: behaviorFragment,
		Template: s.Ports.ApplyTemplate,
	}.WriteAtDir(rs.Dir)
	require.NoError(t, err)

	server := MakeServer(t, rs, rs.Dir, &driver.Server{
		Name:        "primary",
		Args:        []string{"--config", "server.yaml"},
		DynamicPort: "server_port",
		// Disable the auto-GC load-average throttling so GC runs immediately instead
		// of backing off under CI load (see loadAvgGCScheduler in dolt/go's auto_gc.go).
		Envs: []string{"DOLT_GC_SCHEDULER=NONE"},
	}, s.Ports, driver.WithOutputVisitor(s.gcVisitor()))
	server.DBName = "auto_gc_test"

	db, err := server.DB(driver.Connection{})
	require.NoError(t, err)
	t.Cleanup(func() {
		db.Close()
	})

	s.Server = server
	s.DB = db

	s.CreateDatabase(ctx, t)
}

// CreateDatabase creates the vals table along with a battery of secondary indexes,
// mirroring Dolt's inline `index (...)` column definitions (MySQL syntax, not valid in
// Postgres) as separate CREATE INDEX statements. The extra indexes are load-bearing for
// the test: each insert now writes to many prolly trees instead of one, which is what
// lets a modest number of statements push the store past auto-GC's 128MB size threshold
// within a reasonable test run.
func (s *AutoGCTest) CreateDatabase(ctx context.Context, t *testing.T) {
	conn, err := s.DB.Conn(ctx)
	require.NoError(t, err)
	_, err = conn.ExecContext(ctx, `
create table vals (
    id bigint primary key,
    v1 bigint,
    v2 bigint,
    v3 bigint,
    v4 bigint
)
`)
	require.NoError(t, err)
	for _, cols := range [][]string{
		{"v1"}, {"v2"}, {"v3"}, {"v4"},
		{"v1", "v2"}, {"v1", "v3"}, {"v1", "v4"},
		{"v2", "v3"}, {"v2", "v4"}, {"v2", "v1"},
		{"v3", "v1"}, {"v3", "v2"}, {"v3", "v4"},
		{"v4", "v1"}, {"v4", "v2"}, {"v4", "v3"},
	} {
		_, err = conn.ExecContext(ctx, fmt.Sprintf("create index vals_%s_idx on vals (%s)", strings.Join(cols, "_"), strings.Join(cols, ",")))
		require.NoError(t, err)
	}
	_, err = conn.ExecContext(ctx, "select dolt_commit('-Am', 'create vals table')")
	require.NoError(t, err)
	require.NoError(t, conn.Close())
}

func autoGCInsertStatement(i int) string {
	var vals []string
	for j := i * 1024; j < (i+1)*1024; j++ {
		var vs [4]string
		vs[0] = strconv.Itoa(rand.Int())
		vs[1] = strconv.Itoa(rand.Int())
		vs[2] = strconv.Itoa(rand.Int())
		vs[3] = strconv.Itoa(rand.Int())
		val := "(" + strconv.Itoa(j) + "," + strings.Join(vs[:], ",") + ")"
		vals = append(vals, val)
	}
	return "insert into vals values " + strings.Join(vals, ",")
}

// runAutoGCTestUntilGC runs insert+commit cycles until auto GC has completed
// targetGCCount times, or fails if that doesn't happen within a reasonable number of
// iterations. It fails immediately if the server logs a dangling references message.
func runAutoGCTestUntilGC(t *testing.T, s *AutoGCTest, targetGCCount int, commitEvery int) {
	const maxStatements = 1024
	ctx := t.Context()
	s.Setup(ctx, t)

	for i := range maxStatements {
		require.False(t, s.sawDanglingRef.Load(), "saw dangling references message during auto GC")
		if s.gcCount.Load() >= int32(targetGCCount) {
			t.Logf("reached %d auto GCs after %d statements", targetGCCount, i)
			break
		}

		stmt := autoGCInsertStatement(i)
		conn, err := s.DB.Conn(ctx)
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, stmt)
		require.NoError(t, err)
		if (i+1)%commitEvery == 0 {
			_, err = conn.ExecContext(ctx, "select dolt_commit('-am', 'insert from "+strconv.Itoa(i*1024)+"')")
			require.NoError(t, err)
		}
		require.NoError(t, conn.Close())
	}

	require.False(t, s.sawDanglingRef.Load(), "saw dangling references message during auto GC")
	require.GreaterOrEqual(t, s.gcCount.Load(), int32(targetGCCount),
		"did not reach %d auto GCs within %d statements", targetGCCount, maxStatements)

	t.Logf("auto GC count: %d", s.gcCount.Load())
}

// runAutoGCTestDisabled runs a fixed number of insert+commit cycles against a server
// with auto GC disabled, verifying it never fires.
func runAutoGCTestDisabled(t *testing.T, s *AutoGCTest, numStatements int, commitEvery int) {
	ctx := t.Context()
	s.Setup(ctx, t)

	for i := range numStatements {
		stmt := autoGCInsertStatement(i)
		conn, err := s.DB.Conn(ctx)
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, stmt)
		require.NoError(t, err)
		if (i+1)%commitEvery == 0 {
			_, err = conn.ExecContext(ctx, "select dolt_commit('-am', 'insert from "+strconv.Itoa(i*1024)+"')")
			require.NoError(t, err)
		}
		require.NoError(t, conn.Close())
	}

	require.Zero(t, s.gcCount.Load(), "auto GC should not run when disabled")
}
