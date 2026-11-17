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
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"
)

func TestPostgresParametersBelongToTheirEngine(t *testing.T) {
	a, b := sqle.NewEngineVariables(), sqle.NewEngineVariables()
	InitWithRegistry(a.System)
	InitWithRegistry(b.System)

	aName, _, ok := a.System.GetGlobal("application_name")
	require.True(t, ok)
	bName, _, ok := b.System.GetGlobal("application_name")
	require.True(t, ok)
	require.NotSame(t, aName, bName)
	require.NotSame(t, postgresConfigParameters()["application_name"], aName)

	aName.SetDefault("engine_a")
	require.Equal(t, "engine_a", aName.GetDefault())
	require.Equal(t, "psql", bName.GetDefault())
	require.Equal(t, "psql", postgresConfigParameters()["application_name"].GetDefault())

	aMax, _, ok := a.System.GetGlobal("max_connections")
	require.True(t, ok)
	bMax, _, ok := b.System.GetGlobal("max_connections")
	require.True(t, ok)
	require.NotSame(t, aMax, bMax)
}

func TestPostgresRegistrationDoesNotOverwriteDoltEngine(t *testing.T) {
	mysql, postgres := sqle.NewEngineVariables(), sqle.NewEngineVariables()
	before, _, ok := mysql.System.GetGlobal("max_connections")
	require.True(t, ok)
	_, isMysql := before.(*sql.MysqlSystemVariable)
	require.True(t, isMysql)

	InitWithRegistry(postgres.System)
	pgMax, _, ok := postgres.System.GetGlobal("max_connections")
	require.True(t, ok)
	_, isPostgres := pgMax.(*Parameter)
	require.True(t, isPostgres)
	after, _, ok := mysql.System.GetGlobal("max_connections")
	require.True(t, ok)
	require.Same(t, before, after)
	require.False(t, IsValidPostgresConfigParameterInRegistry(mysql.System, "max_connections"))
}

func TestRegressionTimezoneOverrideIsPerEngine(t *testing.T) {
	t.Setenv("REGRESSION_TESTING", "1")
	a := sqle.NewEngineVariables()
	InitWithRegistry(a.System)
	aTimezone, _, ok := a.System.GetGlobal("timezone")
	require.True(t, ok)
	require.Equal(t, "America/Los_Angeles", aTimezone.GetDefault())
	require.NotEqual(t, aTimezone.GetDefault(), postgresConfigParameters()["timezone"].GetDefault())
}
