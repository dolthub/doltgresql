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

package config

import (
	"context"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"
)

func TestPostgresParameterPrecedence(t *testing.T) {
	vars := sqle.NewEngineVariables()
	ctx := sql.NewContextWithVariables(context.Background(), vars)
	require.NoError(t, vars.System.SetGlobal(ctx, dsess.ShowSystemTables, 1))
	_, set, ok := vars.System.GetGlobal(dsess.ShowSystemTables)
	require.True(t, ok)

	InitWithRegistry(vars.System)

	// Dolt's definitions and values are left alone.
	require.NotEmpty(t, doltVariableNames())
	for name := range doltVariableNames() {
		def, _, ok := vars.System.GetGlobal(name)
		require.True(t, ok, name)
		require.IsType(t, &sql.MysqlSystemVariable{}, def, name)
	}
	_, got, ok := vars.System.GetGlobal(dsess.ShowSystemTables)
	require.True(t, ok)
	require.Equal(t, set, got)

	// PostgreSQL replaces the MySQL built-ins it shares a name with.
	for _, name := range []string{"lc_messages", "log_error_verbosity", "max_connections", "port", "transaction_isolation", "transaction_read_only"} {
		def, _, ok := vars.System.GetGlobal(name)
		require.True(t, ok, name)
		require.IsType(t, &Parameter{}, def, name)
	}
}

func TestPostgresParametersSkipDoltNames(t *testing.T) {
	vars := sqle.NewEngineVariables()
	port, _, ok := vars.System.GetGlobal("port")
	require.True(t, ok)

	addPostgresParameters(vars.System, map[string]struct{}{"port": {}})

	got, _, ok := vars.System.GetGlobal("port")
	require.True(t, ok)
	require.Same(t, port, got)
	app, _, ok := vars.System.GetGlobal("application_name")
	require.True(t, ok)
	require.IsType(t, &Parameter{}, app)
}
