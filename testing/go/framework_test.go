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
	"net"
	"sync"
	"testing"
	"time"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dserver "github.com/dolthub/doltgresql/server"
)

// ScriptTest defines a consistent structure for testing queries.
type ScriptTest struct {
	// Name of the script.
	Name string
	// The database to create and use. If not provided, then it defaults to "postgres".
	Database string
	// The SQL statements to execute as setup, in order. Results are not checked, but statements must not error.
	SetUpScript []string
	// The set of assertions to make after setup, in order
	Assertions []ScriptTestAssertion
}

// ScriptTestAssertion are the assertions upon which the script executes its main "testing" logic.
type ScriptTestAssertion struct {
	Query       string
	Expected    []sql.Row
	ExpectedErr bool

	// SkipResultsCheck is used to skip assertions on the expected rows returned from a query. For now, this is
	// included as some messages do not have a full logical implementation. Skipping the results check allows us to
	// force the test client to not send of those messages.
	SkipResultsCheck bool

	// Skip is used to completely skip a test, not execute its query at all, and record it as a skipped test
	// in the test suite results.
	Skip bool
}

// RunScript runs the given script.
func RunScript(t *testing.T, script ScriptTest) {
	scriptDatabase := script.Database
	if len(scriptDatabase) == 0 {
		scriptDatabase = "postgres"
	}
	ctx, conn, serverClosed := CreateServer(t, scriptDatabase)
	defer func() {
		conn.Close(ctx)
		serverClosed.Wait()
	}()

	t.Run(script.Name, func(t *testing.T) {
		// Run the setup
		for _, query := range script.SetUpScript {
			_, err := conn.Exec(ctx, query)
			require.NoError(t, err)
		}
		// Run the assertions
		for _, assertion := range script.Assertions {
			t.Run(assertion.Query, func(t *testing.T) {
				if assertion.Skip {
					t.Skip("Skip has been set in the assertion")
				}
				// If we're skipping the results check, then we call Execute, as it uses a simplified message model.
				// The more complicated model is only partially implemented, and therefore won't work for all queries.
				if assertion.SkipResultsCheck || assertion.ExpectedErr {
					_, err := conn.Exec(ctx, assertion.Query)
					if assertion.ExpectedErr {
						require.Error(t, err)
					} else {
						require.NoError(t, err)
					}
				} else {
					rows, err := conn.Query(ctx, assertion.Query)
					require.NoError(t, err)
					defer rows.Close()
					assert.Equal(t, NormalizeRows(assertion.Expected), ReadRows(t, rows))
				}
			})
		}
	})
}

// RunScripts runs the given collection of scripts.
func RunScripts(t *testing.T, scripts []ScriptTest) {
	for _, script := range scripts {
		RunScript(t, script)
	}
}

// CreateServer creates a server with the given database, returning a connection to the server. The server will close
// when the connection is closed (or loses its connection to the server). The accompanying WaitGroup may be used to wait
// until the server has closed.
func CreateServer(t *testing.T, database string) (context.Context, *pgx.Conn, *sync.WaitGroup) {
	require.NotEmpty(t, database)
	port := GetUnusedPort(t)
	server.DefaultProtocolListenerFunc = dserver.NewLimitedListener
	code, serverClosed := dserver.RunInMemory([]string{fmt.Sprintf("--port=%d", port), "--host=127.0.0.1"})
	require.Equal(t, 0, *code)

	ctx := context.Background()
	err := func() error {
		// The connection attempt may be made before the server has grabbed the port, so we'll retry the first
		// connection a few times.
		var conn *pgx.Conn
		var err error
		for i := 0; i < 3; i++ {
			conn, err = pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/", port))
			if err == nil {
				break
			} else {
				time.Sleep(time.Second)
			}
		}
		if err != nil {
			return err
		}

		defer conn.Close(ctx)
		_, err = conn.Exec(ctx, fmt.Sprintf("CREATE DATABASE %s;", database))
		return err
	}()
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/%s", port, database))
	require.NoError(t, err)
	return ctx, conn, serverClosed
}

// ReadRows reads all of the given rows into a slice. This also normalizes all of the rows. Does not call Close() on the rows.
func ReadRows(t *testing.T, rows pgx.Rows) []sql.Row {
	var slice []sql.Row
	for rows.Next() {
		row, err := rows.Values()
		require.NoError(t, err)
		slice = append(slice, row)
	}
	return NormalizeRows(slice)
}

// NormalizeRow normalizes each value's type, as the tests only want to compare values. Returns a new row.
func NormalizeRow(row sql.Row) sql.Row {
	newRow := make(sql.Row, len(row))
	for i := range row {
		switch val := row[i].(type) {
		case int:
			newRow[i] = int64(val)
		case int8:
			newRow[i] = int64(val)
		case int16:
			newRow[i] = int64(val)
		case int32:
			newRow[i] = int64(val)
		case uint:
			newRow[i] = int64(val)
		case uint8:
			newRow[i] = int64(val)
		case uint16:
			newRow[i] = int64(val)
		case uint32:
			newRow[i] = int64(val)
		case uint64:
			// PostgreSQL does not support an uint64 type, so we can always convert this to an int64 safely.
			newRow[i] = int64(val)
		case float32:
			newRow[i] = float64(val)
		default:
			newRow[i] = val
		}
	}
	return newRow
}

// NormalizeRows normalizes each value's type within each row, as the tests only want to compare values. Returns a new
// set of rows in the same order.
func NormalizeRows(rows []sql.Row) []sql.Row {
	newRows := make([]sql.Row, len(rows))
	for i := range rows {
		newRows[i] = NormalizeRow(rows[i])
	}
	return newRows
}

// GetUnusedPort returns an unused port.
func GetUnusedPort(t *testing.T) int {
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := listener.Addr().(*net.TCPAddr).Port
	require.NoError(t, listener.Close())
	return port
}
