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

	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

func TestServerConfigVariableStatement(t *testing.T) {
	scriptDatabase := "postgres"
	var ctx context.Context
	var conn *Connection
	port, err := sql.GetEmptyPort()
	require.NoError(t, err)

	if runOnPostgres {
		ctx = context.Background()
		pgxConn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/%s?sslmode=disable", 5432, scriptDatabase))
		require.NoError(t, err)
		conn = &Connection{
			Default: pgxConn,
			Current: pgxConn,
		}
		require.NoError(t, pgxConn.Ping(ctx))
		defer func() {
			conn.Close(ctx)
		}()
	} else {
		var controller *svcs.Controller
		ctx, conn, controller = CreateServerWithPort(t, scriptDatabase, port)
		defer func() {
			conn.Close(ctx)
			controller.Stop()
			err := controller.WaitForStop()
			require.NoError(t, err)
		}()
	}

	t.Run("show 'port' configuration variable", func(t *testing.T) {
		runScript(t, ctx, ScriptTest{
			Name:        "set 'port' configuration variable",
			SetUpScript: []string{},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SHOW port",
					Expected: []sql.Row{{port}},
				},
				{
					Query:       "SET port TO '5432'",
					ExpectedErr: "is a read only variable",
				},
				{
					Query:    "SELECT current_setting('port')",
					Expected: []sql.Row{{fmt.Sprintf("%v", port)}},
				},
			},
		}, conn, true)
	})
}
