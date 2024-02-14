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
	"strings"
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
	// The SQL statements to execute as setup, in order. Results are not checked, but statements must not error. 
	// An initial comment can be used to Setup is always run on the primary.
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
	ScriptTestAssertion
}

var replicationTests = []ReplicationTest{
	{
		Name: "simple replication, strings and integers",
		SetUpScript: []string{
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
		},
		Assertions: []ReplicationTestAssertion{
			{
				ReplicationTarget: ReplicationTargetReplica,
				ScriptTestAssertion: ScriptTestAssertion{
					Query: "SELECT * FROM test order by id",
					Expected: []sql.Row{
						{int32(2), "three"},
						{int32(4), "five"},
						{int32(6), "six"},
					},
				},
			},
		},
	},
	{
		Name: "all supported types",
		SetUpScript: []string{
			"/* replica */ drop table if exists test",
			"/* replica */ create table test (id INT primary key, name varchar(100), age INT, height FLOAT, birth_date DATE, birth_timestamp TIMESTAMP)",
			"drop table if exists test",
			"create table test (id INT primary key, name varchar(100), age INT, height FLOAT, birth_date DATE, birth_timestamp TIMESTAMP)",
			"INSERT INTO test VALUES (1, 'one', 1, 1.1, '2021-01-01', '2021-01-01 12:00:00')",
			"INSERT INTO test VALUES (2, 'two', 2, 2.2, '2021-02-02', '2021-02-02 13:00:00')",
			"UPDATE test SET name = 'three' WHERE id = 2",
			"DELETE FROM test WHERE id = 1",
		},
		Assertions: []ReplicationTestAssertion{
			{
				ReplicationTarget: ReplicationTargetReplica,
				ScriptTestAssertion: ScriptTestAssertion{
					Query: "SELECT * FROM test order by id",
					Expected: []sql.Row{
						// TODO: The DATE field should not return time in its output
						{int32(2), "three", int32(2), 2.2, "2021-02-02 00:00:00", "2021-02-02 13:00:00"},
					},
				},
			},
		},
	},
	{
		Name: "simple replication, strings and integers",
		SetUpScript: []string{
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
		},
		Assertions: []ReplicationTestAssertion{
			{
				ReplicationTarget: ReplicationTargetReplica,
				ScriptTestAssertion: ScriptTestAssertion{
					Query: "SELECT * FROM test order by id",
					Expected: []sql.Row{
						{int32(2), "three"},
						{int32(4), "five"},
						{int32(6), "six"},
					},
				},
			},
		},
	},
	{
		Name: "all types",
		Skip: true, // some types don't work yet
		SetUpScript: []string{
			"/* replica */ drop table if exists test",
			"/* replica */ create table test (id INT primary key, name varchar(100), age INT, is_cool BOOLEAN, height FLOAT, birth_date DATE, birth_timestamp TIMESTAMP)",
			"drop table if exists test",
			"create table test (id INT primary key, name varchar(100), age INT, is_cool BOOLEAN, height FLOAT, birth_date DATE, birth_timestamp TIMESTAMP)",
			"INSERT INTO test VALUES (1, 'one', 1, true, 1.1, '2021-01-01', '2021-01-01 12:00:00')",
			"INSERT INTO test VALUES (2, 'two', 2, false, 2.2, '2021-02-02', '2021-02-02 13:00:00')",
			"UPDATE test SET name = 'three' WHERE id = 2",
			"DELETE FROM test WHERE id = 1",
		},
		Assertions: []ReplicationTestAssertion{
			{
				ReplicationTarget: ReplicationTargetReplica,
				ScriptTestAssertion: ScriptTestAssertion{
					Query: "SELECT * FROM test order by id",
					Expected: []sql.Row{
						{int32(2), "three", int32(2), false, 2.2, "2021-02-02", "2021-02-02 13:00:00"},
					},
				},
			},
		},
	},
}

func TestReplication(t *testing.T) {
	for _, test := range replicationTests {
		RunReplicationScript(t, test)
	}
}

const slotName = "doltgres_slot"

// RunScript runs the given script.
func RunReplicationScript(t *testing.T, script ReplicationTest) {
	scriptDatabase := script.Database
	if len(scriptDatabase) == 0 {
		scriptDatabase = "postgres"
	}

	database := "postgres"
	primaryDns := fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/%s?sslmode=disable", 5432, database)
	primaryReplicationDns := fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/%s?replication=database", 5432, database)
	require.NoError(t, logrepl.SetupReplication(primaryDns, slotName))

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
	replicator, err := logrepl.NewLogicalReplicator(primaryReplicationDns, replicationDns)
	require.NoError(t, err)
	
	go func() {
		err := replicator.StartReplication(slotName)
		require.NoError(t, err)
	}()

	// give replication time to begin before running scripts
	time.Sleep(1 * time.Second)

	// TODO: stop replication
	
	ctx = context.Background()
	primaryConn, err := pgx.Connect(ctx, primaryDns)
	require.NoError(t, err)
	
	t.Run(script.Name, func(t *testing.T) {
		runReplicationScript(ctx, t, script, primaryConn, replicaConn)
	})
}

// runScript runs the script given on the postgres connection provided
func runReplicationScript(ctx context.Context, t *testing.T, script ReplicationTest, primaryConn, replicaConn *pgx.Conn) {
	if script.Skip {
		t.Skip("Skip has been set in the script")
	}

	primaryConnections := map[string]*pgx.Conn{
		"a": primaryConn,
	}
	
	// Run the setup
	for _, query := range script.SetUpScript {
		target, client := clientSpecFromQueryComment(query)
		var conn *pgx.Conn
		switch target {
		case "primary":
			conn = primaryConnections[client]
		case "replica":
			conn = replicaConn
		default:
			require.Fail(t, "Invalid target in setup script: ", target)
		}
		log.Println("Running setup query:", query)
		_, err := conn.Exec(ctx, query)
		require.NoError(t, err)
	}
	
	// give replication time to catch up
	time.Sleep(1 * time.Second)
	
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

// clientSpecFromQueryComment returns "replica" if the query is meant to be run on the replica, and "primary" if it's meant
// to be run on the primary, based on the comment in the query. If not comment, the query runs on the primary
func clientSpecFromQueryComment(query string) (string, string) {
	startCommentIdx := strings.Index(query, "/*")
	endCommentIdx := strings.Index(query, "*/")
	if startCommentIdx < 0 || endCommentIdx < 0 {
		return "primary", "a"
	}

	query = query[startCommentIdx+2 : endCommentIdx]
	if strings.Index(query, "replica") >= 0 {
		return "replica", "a"
	}
	
	if i := strings.Index(query, "primary "); i > 0 && i +len("primary ") < len(query) {
		return "primary", query[i + len("primary "):]
	}
	
	return "primary", "a"
}
