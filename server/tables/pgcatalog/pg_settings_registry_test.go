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

package pgcatalog

import (
	"context"
	"io"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/server/config"
)

func applicationNameRow(t *testing.T, ctx *sql.Context) sql.Row {
	t.Helper()
	iter, err := (PgSettingsHandler{}).RowIter(ctx, nil)
	require.NoError(t, err)
	defer iter.Close(ctx)
	for {
		row, err := iter.Next(ctx)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if row[0] == "application_name" {
			return row
		}
	}
	t.Fatal("application_name missing from pg_settings")
	return nil
}

func TestPgSettingsUsesEngineAndSessionDefaults(t *testing.T) {
	a, b := sqle.NewEngineVariables(), sqle.NewEngineVariables()
	config.InitWithRegistry(a.System)
	config.InitWithRegistry(b.System)
	aDefinition, _, ok := a.System.GetGlobal("application_name")
	require.True(t, ok)
	aDefinition.SetDefault("a reset")

	aCtx := sql.NewContextWithVariables(context.Background(), a, sql.WithSession(sql.NewBaseSessionWithVariables(a)))
	bCtx := sql.NewContextWithVariables(context.Background(), b, sql.WithSession(sql.NewBaseSessionWithVariables(b)))
	require.NoError(t, aCtx.SetSessionVariable(aCtx, "application_name", "a current"))

	aRow := applicationNameRow(t, aCtx)
	bRow := applicationNameRow(t, bCtx)
	require.Equal(t, "a current", aRow[1])
	require.Equal(t, "psql", aRow[12])    // boot_val
	require.Equal(t, "a reset", aRow[13]) // reset_val
	require.Equal(t, "psql", bRow[1])
	require.Equal(t, "psql", bRow[13])
}
