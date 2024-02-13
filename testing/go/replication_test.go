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

package _go

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/dolthub/doltgresql/server/logrepl"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ReplicationTarget byte
const (
	ReplicationTargetPrimary ReplicationTarget = iota
	ReplicationTargetReplica
)

type ReplicationTest struct {
	// Name of the script.
	Name string
	// The database to create and use. If not provided, then it defaults to "postgres".
	Database string
	// The SQL statements to execute as setup, in order. Results are not checked, but statements must not error. Setup is always run on the primary.
	SetUpScript []string
	// The set of assertions to make after setup, in order
	Assertions []ReplicationTestAssertion
	// When using RunScripts, setting this on one (or more) tests causes RunScripts to ignore all tests that have this
	// set to false (which is the default value). This allows a developer to easily "focus" on a specific test without
	// having to comment out other tests, pull it into a different function, etc. In addition, CI ensures that this is
	// false before passing, meaning this prevents the commented-out situation where the developer forgets to uncomment
	// their code.
	Focus bool
	// Skip is used to completely skip a test including setup
	Skip bool
}

type ReplicationTestAssertion struct {
	ReplicationTarget ReplicationTarget
	Query       string
	Expected    []sql.Row
	ExpectedErr bool

	BindVars []any

	// SkipResultsCheck is used to skip assertions on the expected rows returned from a query. For now, this is
	// included as some messages do not have a full logical implementation. Skipping the results check allows us to
	// force the test client to not send of those messages.
	SkipResultsCheck bool

	// Skip is used to completely skip a test, not execute its query at all, and record it as a skipped test
	// in the test suite results.
	Skip bool
}

var replicationTests = []ReplicationTest{
	{
		Name: "simple replication",
		SetUpScript: []string{
			"drop table if exists test",
			"CREATE TABLE test (id INT primary key, name varchar(100))",
			"INSERT INTO test VALUES (1, 'one')",
			"INSERT INTO test VALUES (2, 'two')",
			"UPDATE test SET name = 'three' WHERE id = 2",
			"DELETE FROM test WHERE id = 1",
			"INSERT INTO test VALUES (3, 'one')",
			"INSERT INTO test VALUES (4, 'two')",
			"UPDATE test SET name = 'three' WHERE id = 4",
			"DELETE FROM test WHERE id = 3",
			"INSERT INTO test VALUES (5, 'one')",
			"INSERT INTO test VALUES (6, 'two')",
			"UPDATE test SET name = 'three' WHERE id = 5",
			"DELETE FROM test WHERE id = 5",
			"INSERT INTO test VALUES (7, 'one')",
			"INSERT INTO test VALUES (8, 'two')",
			"UPDATE test SET name = 'three' WHERE id = 8",
			"DELETE FROM test WHERE id = 7",
		},
	},
}

func TestReplication(t *testing.T) {
	for _, test := range replicationTests {
		RunReplicationScript(t, test)
	}
}

// RunScript runs the given script.
func RunReplicationScript(t *testing.T, script ReplicationTest) {
	scriptDatabase := script.Database
	if len(scriptDatabase) == 0 {
		scriptDatabase = "postgres"
	}

	database := "postgres"
	require.NoError(t, logrepl.SetupReplication(database))

	ctx, replicaConn, controller := CreateServer(t, scriptDatabase)
	defer func() {
		replicaConn.Close(ctx)
		controller.Stop()
		err := controller.WaitForStop()
		require.NoError(t, err)
	}()

	connString := replicaConn.PgConn().Conn().RemoteAddr().String()
	_, port, err := net.SplitHostPort(connString)
	require.NoError(t, err)

	replicationDns := fmt.Sprintf("postgres://postgres:password@127.0.0.1:%s/", port)
	replicator, err := logrepl.NewLogicalReplicator(replicationDns)
	require.NoError(t, err)
	
	go func() {
		err := replicator.StartReplication(database)
		require.NoError(t, err)
	}()
	
	ctx = context.Background()
	primaryConn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/%s?sslmode=disable", 5432, database))
	require.NoError(t, err)
	
	t.Run(script.Name, func(t *testing.T) {
		runReplicationScript(ctx, t, script, primaryConn, nil)
	})
}

// runScript runs the script given on the postgres connection provided
func runReplicationScript(ctx context.Context, t *testing.T, script ReplicationTest, primaryConn, replicaConn *pgx.Conn) {
	if script.Skip {
		t.Skip("Skip has been set in the script")
	}

	// Run the setup
	for _, query := range script.SetUpScript {
		log.Println("Running setup query:", query)
		_, err := primaryConn.Exec(ctx, query)
		require.NoError(t, err)
		time.Sleep(100 * time.Millisecond)
	}
	
	// Run the assertions
	for _, assertion := range script.Assertions {
		t.Run(assertion.Query, func(t *testing.T) {
			if assertion.Skip {
				t.Skip("Skip has been set in the assertion")
			}
			
			conn := replicaConn
			if assertion.ReplicationTarget == ReplicationTargetPrimary {
				conn = primaryConn
			}
			
			// If we're skipping the results check, then we call Execute, as it uses a simplified message model.
			if assertion.SkipResultsCheck || assertion.ExpectedErr {
				_, err := conn.Exec(ctx, assertion.Query, assertion.BindVars...)
				if assertion.ExpectedErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			} else {
				rows, err := conn.Query(ctx, assertion.Query, assertion.BindVars...)
				require.NoError(t, err)
				readRows, err := ReadRows(rows)
				require.NoError(t, err)
				assert.Equal(t, NormalizeRows(assertion.Expected), readRows)
			}
		})
	}
}

