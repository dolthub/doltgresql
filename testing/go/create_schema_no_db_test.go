// Copyright 2025 Dolthub, Inc.
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
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dserver "github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/servercfg"
	"github.com/dolthub/doltgresql/servercfg/cfgdetails"
)

// TestCreateSchemaWithNonExistentDatabase tests that connecting to a non-existent
// database fails with the appropriate error at the connection level.
// This is the PostgreSQL-compliant behavior where database validation happens
// during connection establishment.
//
// NOTE: This test cannot use RunScripts because RunScripts auto-creates the
// database before running assertions. We need to test the connection-level
// failure when connecting to a non-existent database.
//
// See: https://github.com/dolthub/doltgresql/issues/1863
func TestCreateSchemaWithNonExistentDatabase(t *testing.T) {
	port, err := sql.GetEmptyPort()
	require.NoError(t, err)

	// Start server using the same pattern as CreateServerWithPort
	controller, err := dserver.RunInMemory(
		&servercfg.DoltgresConfig{
			DoltgresConfig: cfgdetails.DoltgresConfig{
				ListenerConfig: &cfgdetails.DoltgresListenerConfig{
					PortNumber: &port,
					HostStr:    &serverHost,
				},
				LogLevelStr: &testServerLogLevel,
			},
		}, dserver.NewListener,
	)
	require.NoError(t, err)
	defer func() {
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()

	ctx := context.Background()

	t.Run(
		"connection to non-existent database fails", func(t *testing.T) {
			// Attempt to connect to a database that doesn't exist
			connStr := fmt.Sprintf(
				"postgres://postgres:password@%s:%d/nonexistent_db?sslmode=disable",
				serverHost,
				port,
			)
			_, err := pgx.Connect(ctx, connStr)

			require.Error(t, err, "connection should fail when database doesn't exist")
			assert.Contains(
				t, err.Error(), "does not exist",
				"expected 'does not exist' error, got: %v", err,
			)
		},
	)

	t.Run(
		"connection to existing database succeeds", func(t *testing.T) {
			// Verify that connecting to the default "postgres" database works
			connStr := fmt.Sprintf("postgres://postgres:password@%s:%d/postgres?sslmode=disable", serverHost, port)
			conn, err := pgx.Connect(ctx, connStr)
			require.NoError(t, err)
			defer conn.Close(ctx)

			// Verify we can create a schema on the valid database
			_, err = conn.Exec(ctx, "CREATE SCHEMA test_schema_1863")
			require.NoError(t, err)

			// Verify the schema was created
			rows, err := conn.Query(
				ctx,
				"SELECT schema_name FROM information_schema.schemata WHERE schema_name = 'test_schema_1863'",
			)
			require.NoError(t, err)
			defer rows.Close()

			require.True(t, rows.Next(), "expected to find test_schema_1863")
		},
	)
}
