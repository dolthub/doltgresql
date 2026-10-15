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

package sql_test

import (
	"context"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/stretchr/testify/require"

	psql "github.com/dolthub/doltgresql/postgres/parser/parser/sql"
	"github.com/dolthub/doltgresql/server/config"
)

func TestVariableStatementsRequireAnEngineRegistry(t *testing.T) {
	parser := psql.NewPostgresParser()
	_, err := parser.ParseSimple("SET application_name TO 'client'")
	require.ErrorContains(t, err, "engine variable registry")

	// A context with no engine variables yields an error, not a panic.
	unbound := sql.NewEmptyContext()
	for _, query := range []string{"SET application_name TO 'client'", "SHOW application_name"} {
		require.NotPanics(t, func() {
			_, _, _, err = parser.Parse(unbound, query, false)
		})
		require.ErrorContains(t, err, "engine variable registry", query)
		require.NotPanics(t, func() {
			_, _, err = parser.ParseOneWithOptions(unbound, query, vitess.ParserOptions{})
		})
		require.ErrorContains(t, err, "engine variable registry", query)
	}

	vars := sqle.NewEngineVariables()
	config.InitWithRegistry(vars.System)
	ctx := sql.NewContextWithVariables(context.Background(), vars)
	statement, _, _, err := parser.Parse(ctx, "SET application_name TO 'client'", false)
	require.NoError(t, err)
	require.NotNil(t, statement)
}
