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

package _go

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/server/logrepl"
)

type ReplicationTarget byte

// special pseudo-queries for orchestrating replication tests
const (
	createReplicationSlot = "createReplicationSlot"
	dropReplicationSlot   = "dropReplicationSlot"
	stopReplication       = "stopReplication"
	startReplication      = "startReplication"
	waitForCatchup        = "waitForCatchup"
)

type ReplicationTest struct {
	// Name of the script.
	Name string
	// The database to create and use. If not provided, then it defaults to "postgres".
	Database string
	// The SQL statements to execute as setup, in order. Results are not checked, but statements must not error.
	// An initial comment can be used to Setup is always run on the primary.
	SetUpScript []string
	// The set of assertions to make after setup, in order
	Assertions []ScriptTestAssertion
	// When using RunScripts, setting this on one (or more) tests causes RunScripts to ignore all tests that have this
	// set to false (which is the default value). This allows a developer to easily "focus" on a specific test without
	// having to comment out other tests, pull it into a different function, etc. In addition, CI ensures that this is
	// false before passing, meaning this prevents the commented-out situation where the developer forgets to uncomment
	// their code.
	Focus bool
	// Skip is used to completely skip a test including setup
	Skip bool
}

var replicationTests = []ReplicationTest{
	{
		Name: "simple replication, strings and integers",
		SetUpScript: []string{
			dropReplicationSlot,
			createReplicationSlot,
			startReplication,
			"/* replica */ drop table if exists test",
			"/* replica */ create table test (id INT primary key, name varchar(100))",
			"drop table if exists test",
			"CREATE TABLE test (id INT primary key, name varchar(100))",
			"INSERT INTO test VALUES (1, 'one')",
			"INSERT INTO test VALUES (2, 'two')",
			"UPDATE test SET name = 'three' WHERE id = 2",
			"DELETE FROM test WHERE id = 1",
			"INSERT INTO test VALUES (3, 'one')",
			"INSERT INTO test VALUES (4, 'two')",
			"UPDATE test SET name = 'five' WHERE id = 4",
			"DELETE FROM test WHERE id = 3",
			"INSERT INTO test VALUES (5, 'one')",
			"INSERT INTO test VALUES (6, 'two')",
			"UPDATE test SET name = 'six' WHERE id = 6",
			"DELETE FROM test WHERE id = 5",
			waitForCatchup,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "/* replica */ SELECT * FROM test order by id",
				Expected: []sql.Row{
					{int32(2), "three"},
					{int32(4), "five"},
					{int32(6), "six"},
				},
			},
		},
	},
	{
		Name: "stale start",
		SetUpScript: []string{
			// Postgres will not start tracking which WAL locations to send until the replication slot is created, so we have
			// to do that first. Customers have the same constraint: they must import any table data that existed before
			// they create the replication slot.
			dropReplicationSlot,
			createReplicationSlot,
			"/* replica */ drop table if exists test",
			"/* replica */ create table test (id INT primary key, name varchar(100))",
			"drop table if exists test",
			"CREATE TABLE test (id INT primary key, name varchar(100))",
			"INSERT INTO test VALUES (1, 'one')",
			"INSERT INTO test VALUES (2, 'two')",
			"UPDATE test SET name = 'three' WHERE id = 2",
			"DELETE FROM test WHERE id = 1",
			"INSERT INTO test VALUES (3, 'one')",
			"INSERT INTO test VALUES (4, 'two')",
			"UPDATE test SET name = 'five' WHERE id = 4",
			"DELETE FROM test WHERE id = 3",
			"INSERT INTO test VALUES (5, 'one')",
			"INSERT INTO test VALUES (6, 'two')",
			"UPDATE test SET name = 'six' WHERE id = 6",
			"DELETE FROM test WHERE id = 5",
			startReplication,
			waitForCatchup,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "/* replica */ SELECT * FROM test order by id",
				Expected: []sql.Row{
					{int32(2), "three"},
					{int32(4), "five"},
					{int32(6), "six"},
				},
			},
		},
	},
	{
		Name: "stopping and resuming replication",
		SetUpScript: []string{
			dropReplicationSlot,
			createReplicationSlot,
			startReplication,
			"/* replica */ drop table if exists test",
			"/* replica */ create table test (id INT primary key, name varchar(100))",
			"drop table if exists test",
			"CREATE TABLE test (id INT primary key, name varchar(100))",
			"INSERT INTO test VALUES (1, 'one')",
			"INSERT INTO test VALUES (2, 'two')",
			waitForCatchup,
			stopReplication,
			"UPDATE test SET name = 'three' WHERE id = 2",
			"DELETE FROM test WHERE id = 1",
			"INSERT INTO test VALUES (3, 'one')",
			"INSERT INTO test VALUES (4, 'two')",
			"UPDATE test SET name = 'five' WHERE id = 4",
			"DELETE FROM test WHERE id = 3",
			startReplication,
			"INSERT INTO test VALUES (5, 'one')",
			"INSERT INTO test VALUES (6, 'two')",
			"UPDATE test SET name = 'six' WHERE id = 6",
			"DELETE FROM test WHERE id = 5",
			waitForCatchup,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "/* replica */ SELECT * FROM test order by id",
				Expected: []sql.Row{
					{int32(2), "three"},
					{int32(4), "five"},
					{int32(6), "six"},
				},
			},
		},
	},
	{
		Name: "extended stop/start",
		SetUpScript: []string{
			dropReplicationSlot,
			createReplicationSlot,
			"/* replica */ drop table if exists test",
			"/* replica */ create table test (id INT primary key, name varchar(100))",
			"drop table if exists test",
			"CREATE TABLE test (id INT primary key, name varchar(100))",
			"INSERT INTO test VALUES (1, 'one')",
			"INSERT INTO test VALUES (2, 'two')",
			"UPDATE test SET name = 'three' WHERE id = 2",
			"DELETE FROM test WHERE id = 1",
			"INSERT INTO test VALUES (3, 'one')",
			"INSERT INTO test VALUES (4, 'two')",
			"UPDATE test SET name = 'five' WHERE id = 4",
			"DELETE FROM test WHERE id = 3",
			"INSERT INTO test VALUES (5, 'one')",
			startReplication,
			"INSERT INTO test VALUES (6, 'two')",
			"UPDATE test SET name = 'six' WHERE id = 6",
			stopReplication,
			"DELETE FROM test WHERE id = 5",
			"INSERT INTO test VALUES (7, 'one')",
			"INSERT INTO test VALUES (8, 'two')",
			startReplication,
			"UPDATE test SET name = 'nine' WHERE id = 8",
			"DELETE FROM test WHERE id = 7",
			"INSERT INTO test VALUES (9, 'one')",
			stopReplication,
			startReplication,
			"INSERT INTO test VALUES (10, 'two')",
			"UPDATE test SET name = 'eleven' WHERE id = 10",
			stopReplication,
			"DELETE FROM test WHERE id = 9",
			"INSERT INTO test VALUES (11, 'one')",
			"INSERT INTO test VALUES (12, 'two')",
			"UPDATE test SET name = 'thirteen' WHERE id = 12",
			"DELETE FROM test WHERE id = 11",
			startReplication,
			"INSERT INTO test VALUES (13, 'one')",
			"INSERT INTO test VALUES (14, 'two')",
			"UPDATE test SET name = 'fifteen' WHERE id = 14",
			"DELETE FROM test WHERE id = 13",
			waitForCatchup,
			stopReplication,
			// below this point we don't expect to find any values replicated because replication was stopped
			"INSERT INTO test VALUES (15, 'one')",
			"INSERT INTO test VALUES (16, 'two')",
			"UPDATE test SET name = 'seventeen' WHERE id = 16",
			"DELETE FROM test WHERE id = 15",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "/* replica */ SELECT * FROM test order by id",
				Expected: []sql.Row{
					{int32(2), "three"},
					{int32(4), "five"},
					{int32(6), "six"},
					{int32(8), "nine"},
					{int32(10), "eleven"},
					{int32(12), "thirteen"},
					{int32(14), "fifteen"},
				},
			},
		},
	},
	{
		Name: "all supported types",
		SetUpScript: []string{
			dropReplicationSlot,
			createReplicationSlot,
			startReplication,
			"/* replica */ drop table if exists test",
			"/* replica */ create table test (id INT primary key, name varchar(100), u_id uuid, age INT, height FLOAT, birth_date DATE, birth_timestamp TIMESTAMP)",
			"drop table if exists test",
			"create table test (id INT primary key, name varchar(100), u_id uuid, age INT, height FLOAT, birth_date DATE, birth_timestamp TIMESTAMP)",
			"INSERT INTO test VALUES (1, 'one', '5ef34887-e635-4c9c-a994-97b1cb810786', 1, 1.1, '2021-01-01', '2021-01-01 12:00:00')",
			"INSERT INTO test VALUES (2, 'two', '2de55648-76ec-4f66-9fae-bd3d853fb0da', 2, 2.2, '2021-02-02', '2021-02-02 13:00:00')",
			"UPDATE test SET name = 'three' WHERE id = 2",
			"update test set u_id = '3232abe7-560b-4714-a020-2b1a11a1ec65' where id = 2",
			"DELETE FROM test WHERE id = 1",
			waitForCatchup,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "/* replica */ SELECT * FROM test order by id",
				Expected: []sql.Row{
					// TODO: The DATE field should not return time in its output
					{int32(2), "three", "3232abe7-560b-4714-a020-2b1a11a1ec65", int32(2), 2.2, "2021-02-02 00:00:00", "2021-02-02 13:00:00"},
				},
			},
		},
	},
	{
		Name: "concurrent writes",
		SetUpScript: []string{
			dropReplicationSlot,
			createReplicationSlot,
			startReplication,
			"/* replica */ drop table if exists test",
			"/* replica */ create table test (id INT primary key, name varchar(100))",
			"drop table if exists test",
			"CREATE TABLE test (id INT primary key, name varchar(100))",
			"/* primary a */ START TRANSACTION",
			"/* primary a */ INSERT INTO test VALUES (1, 'one')",
			"/* primary a */ INSERT INTO test VALUES (2, 'two')",
			"/* primary b */ START TRANSACTION",
			"/* primary b */ INSERT INTO test VALUES (3, 'one')",
			"/* primary b */ INSERT INTO test VALUES (4, 'two')",
			"/* primary a */ UPDATE test SET name = 'three' WHERE id > 0",
			"/* primary a */ DELETE FROM test WHERE id = 1",
			"/* primary b */ UPDATE test SET name = 'five' WHERE id > 0",
			"/* primary b */ DELETE FROM test WHERE id = 3",
			"/* primary b */ COMMIT",
			"/* primary a */ COMMIT",
			waitForCatchup,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "/* replica */ SELECT * FROM test order by id",
				Expected: []sql.Row{
					{int32(2), "three"},
					{int32(4), "five"},
				},
			},
		},
	},
	{
		Name: "all types",
		Skip: true, // some types don't work yet
		SetUpScript: []string{
			dropReplicationSlot,
			createReplicationSlot,
			startReplication,
			"/* replica */ drop table if exists test",
			"/* replica */ create table test (id INT primary key, name varchar(100), age INT, is_cool BOOLEAN, height FLOAT, birth_date DATE, birth_timestamp TIMESTAMP)",
			"drop table if exists test",
			"create table test (id INT primary key, name varchar(100), age INT, is_cool BOOLEAN, height FLOAT, birth_date DATE, birth_timestamp TIMESTAMP)",
			"INSERT INTO test VALUES (1, 'one', 1, true, 1.1, '2021-01-01', '2021-01-01 12:00:00')",
			"INSERT INTO test VALUES (2, 'two', 2, false, 2.2, '2021-02-02', '2021-02-02 13:00:00')",
			"UPDATE test SET name = 'three' WHERE id = 2",
			"DELETE FROM test WHERE id = 1",
			waitForCatchup,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "/* replica */ SELECT * FROM test order by id",
				Expected: []sql.Row{
					{int32(2), "three", int32(2), false, 2.2, "2021-02-02", "2021-02-02 13:00:00"},
				},
			},
		},
	},
}

func TestReplication(t *testing.T) {
	RunReplicationScripts(t, replicationTests)
}

// RunScripts runs the given collection of scripts.
func RunReplicationScripts(t *testing.T, scripts []ReplicationTest) {
	// First, we'll run through the scripts to check for the Focus variable. If it's true, then append it to the new slice.
	focusScripts := make([]ReplicationTest, 0, len(scripts))
	for _, script := range scripts {
		if script.Focus {
			// If this is running in GitHub Actions, then we'll panic, because someone forgot to disable it before committing
			if _, ok := os.LookupEnv("GITHUB_ACTION"); ok {
				panic(fmt.Sprintf("The script `%s` has Focus set to `true`. GitHub Actions requires that "+
					"all tests are run, which Focus circumvents, leading to this error. Please disable Focus on "+
					"all tests.", script.Name))
			}
			focusScripts = append(focusScripts, script)
		}
	}
	// If we have scripts with Focus set, then we replace the normal script slice with the new slice.
	if len(focusScripts) > 0 {
		scripts = focusScripts
	}

	for _, script := range scripts {
		RunReplicationScript(t, script)
	}
}

const slotName = "doltgres_slot"
const localPostgresPort = 5432

// RunReplicationScript runs the given ReplicationTest.
func RunReplicationScript(t *testing.T, script ReplicationTest) {
	scriptDatabase := script.Database
	if len(scriptDatabase) == 0 {
		scriptDatabase = "postgres"
	}

	database := "postgres"
	// primaryDns is the connection to the actual postgres (not doltgres) database, which is why we use port 5342.
	// If you have postgres running on a different port, you'll need to change this.
	primaryDns := fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/%s?sslmode=disable", localPostgresPort, database)

	ctx, replicaConn, controller := CreateServer(t, scriptDatabase)
	defer func() {
		replicaConn.Close(ctx)
		controller.Stop()
		err := controller.WaitForStop()
		require.NoError(t, err)
	}()

	ctx = context.Background()
	t.Run(script.Name, func(t *testing.T) {
		runReplicationScript(ctx, t, script, replicaConn, primaryDns)
	})
}

func newReplicator(t *testing.T, replicaConn *pgx.Conn, primaryDns string) *logrepl.LogicalReplicator {
	connString := replicaConn.PgConn().Conn().RemoteAddr().String()
	_, port, err := net.SplitHostPort(connString)
	require.NoError(t, err)

	replicaDns := fmt.Sprintf("postgres://postgres:password@127.0.0.1:%s/", port)

	r, err := logrepl.NewLogicalReplicator(primaryDns, replicaDns)
	require.NoError(t, err)
	return r
}

// runReplicationScript runs the script given on the postgres connection provided
func runReplicationScript(
	ctx context.Context,
	t *testing.T,
	script ReplicationTest,
	replicaConn *pgx.Conn,
	primaryDns string,
) {
	r := newReplicator(t, replicaConn, primaryDns)
	defer r.Stop()

	if script.Skip {
		t.Skip("Skip has been set in the script")
	}
	
	// Every replication script should drop and re-create their publication slot, mostly in case it doesn't already exist.
	require.NoError(t, r.DropPublication(slotName))
	require.NoError(t, r.CreatePublication(slotName))

	connections := map[string]*pgx.Conn{
		"replica": replicaConn,
	}

	defer func() {
		for _, conn := range connections {
			if conn != nil {
				conn.Close(ctx)
			}
		}
	}()

	// Run the setup
	for _, query := range script.SetUpScript {
		// handle logic for special pseudo-queries
		if handlePseudoQuery(t, query, r) {
			continue
		}

		conn := connectionForQuery(t, query, connections, primaryDns)
		log.Println("Running setup query:", query)
		_, err := conn.Exec(ctx, query)
		require.NoError(t, err)
	}

	// Run the assertions
	for _, assertion := range script.Assertions {
		t.Run(assertion.Query, func(t *testing.T) {
			if assertion.Skip {
				t.Skip("Skip has been set in the assertion")
			}

			// handle logic for special pseudo-queries
			if handlePseudoQuery(t, assertion.Query, r) {
				return
			}

			conn := connectionForQuery(t, assertion.Query, connections, primaryDns)

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

// connectionForQuery returns the connection to use for the given query
func connectionForQuery(t *testing.T, query string, connections map[string]*pgx.Conn, primaryDns string) *pgx.Conn {
	target, client := clientSpecFromQueryComment(query)
	var conn *pgx.Conn
	switch target {
	case "primary":
		conn = connections[client]
		if conn == nil {
			var err error
			conn, err = pgx.Connect(context.Background(), primaryDns)
			require.NoError(t, err)
			connections[client] = conn
		}
	case "replica":
		conn = connections["replica"]
	default:
		require.Fail(t, "Invalid target in setup script: ", target)
	}
	return conn
}

// handlePseudoQuery handles special pseudo-queries that are used to orchestrate replication tests and returns whether
// one was handled.
func handlePseudoQuery(t *testing.T, query string, r *logrepl.LogicalReplicator) bool {
	switch query {
	case createReplicationSlot:
		require.NoError(t, r.CreateReplicationSlotIfNecessary(slotName))
		return true
	case dropReplicationSlot:
		require.NoError(t, r.DropReplicationSlot(slotName))
		return true
	case startReplication:
		go func() {
			require.NoError(t, r.StartReplication(slotName))
		}()
		require.NoError(t, waitForRunning(r))
		return true
	case stopReplication:
		r.Stop()
		return true
	case waitForCatchup:
		require.NoError(t, waitForCaughtUp(r))
		return true
	}
	return false
}

// clientSpecFromQueryComment returns "replica" if the query is meant to be run on the replica, and "primary" if it's meant
// to be run on the primary, based on the comment in the query. If not comment, the query runs on the primary
func clientSpecFromQueryComment(query string) (string, string) {
	startCommentIdx := strings.Index(query, "/*")
	endCommentIdx := strings.Index(query, "*/")
	if startCommentIdx < 0 || endCommentIdx < 0 {
		return "primary", "a"
	}

	query = query[startCommentIdx+2 : endCommentIdx]
	if strings.Contains(query, "replica") {
		return "replica", "a"
	}

	if i := strings.Index(query, "primary "); i > 0 && i+len("primary ") < len(query) {
		return "primary", query[i+len("primary "):]
	}

	return "primary", "a"
}

func waitForRunning(r *logrepl.LogicalReplicator) error {
	var duration time.Duration
	for {
		if r.Running() {
			break
		}

		duration += 5 * time.Millisecond
		if duration > 2*time.Second {
			return errors.New("Replication did not start")
		}
		time.Sleep(5 * time.Millisecond)
	}

	return nil
}

func waitForCaughtUp(r *logrepl.LogicalReplicator) error {
	log.Println("Waiting for replication to catch up")
	var duration time.Duration
	for {
		if caughtUp, err := r.CaughtUp(); caughtUp {
			log.Println("replication caught up")
			break
		} else if err != nil {
			return err
		}

		log.Println("replication not caught up, waiting")
		duration += 5 * time.Millisecond
		if duration > 2*time.Second {
			return errors.New("Replication did not catch up")
		}
		time.Sleep(20 * time.Millisecond)
	}

	return nil
}
