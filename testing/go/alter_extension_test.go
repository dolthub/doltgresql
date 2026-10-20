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

package _go

import (
	"sync"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/server/extensions"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

func TestAlterExtension(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "update an extension that does not exist",
			Assertions: []ScriptTestAssertion{
				{
					Query:           `ALTER EXTENSION doltgres_no_such_extension UPDATE;`,
					ExpectedErr:     `extension "doltgres_no_such_extension" does not exist`,
					ExpectedErrCode: "42704",
				},
			},
		},
		{
			Name: "update to the installed version",
			SetUpScript: []string{
				`CREATE EXTENSION "uuid-ossp";`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `ALTER EXTENSION "uuid-ossp" UPDATE;`,
					Expected:        []sql.Row{},
					ExpectedTag:     "ALTER EXTENSION",
					ExpectedNotices: []ExpectedNotice{{Severity: "NOTICE", Message: `version "1.1" of extension "uuid-ossp" is already installed`}},
				},
				{
					Query:           `ALTER EXTENSION "uuid-ossp" UPDATE TO '1.1';`,
					Expected:        []sql.Row{},
					ExpectedNotices: []ExpectedNotice{{Severity: "NOTICE", Message: `version "1.1" of extension "uuid-ossp" is already installed`}},
				},
				{
					Query:           `ALTER EXTENSION "uuid-ossp" UPDATE TO "1.1";`,
					Expected:        []sql.Row{},
					ExpectedNotices: []ExpectedNotice{{Severity: "NOTICE", Message: `version "1.1" of extension "uuid-ossp" is already installed`}},
				},
				{
					Query:           `ALTER EXTENSION "uuid-ossp" UPDATE TO '9.9';`,
					ExpectedErr:     `extension "uuid-ossp" has no update path from version "1.1" to version "9.9"`,
					ExpectedErrCode: "22023",
				},
				{
					Query:    `SELECT uuid_nil();`,
					Expected: []sql.Row{{"00000000-0000-0000-0000-000000000000"}},
				},
				{
					Query:    `SELECT extversion FROM pg_extension WHERE extname = 'uuid-ossp';`,
					Expected: []sql.Row{{"1.1"}},
				},
				{
					Query:    `DROP EXTENSION "uuid-ossp";`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT count(*) FROM pg_extension WHERE extname = 'uuid-ossp';`,
					Expected: []sql.Row{{0}},
				},
			},
		},
		{
			Name: "update keeps the values of the extension's types",
			SetUpScript: []string{
				`CREATE EXTENSION vector;`,
				`CREATE TABLE t (id INT PRIMARY KEY, v vector(3));`,
				`INSERT INTO t VALUES (1, '[1,2,3]'), (2, '[3,2,1]');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:           `ALTER EXTENSION vector UPDATE;`,
					Expected:        []sql.Row{},
					ExpectedNotices: []ExpectedNotice{{Severity: "NOTICE", Message: `version "0.8.6" of extension "vector" is already installed`}},
				},
				{
					Query:    `SELECT id, v FROM t ORDER BY v <-> '[3,2,0]' LIMIT 1;`,
					Expected: []sql.Row{{2, "[3,2,1]"}},
				},
				{
					Query:           `ALTER EXTENSION vector UPDATE TO '0.9.0';`,
					ExpectedErr:     `extension "vector" has no update path from version "0.8.6" to version "0.9.0"`,
					ExpectedErrCode: "22023",
				},
			},
		},
	})
}

// updateTestRoutines are the routines that the update test extension declares across its simulated releases.
var updateTestRoutines = map[string]extdef.Routine{}

// registerUpdateTestExtension registers the extension that TestAlterExtensionUpdatesDeclaration changes between its
// simulated releases, returning its declaration.
var registerUpdateTestExtension = sync.OnceValue(func() *extdef.Extension {
	for _, name := range []string{"dgupdate_kept", "dgupdate_removed", "dgupdate_added"} {
		updateTestRoutines[name] = extdef.Routine{
			Name:    name,
			Symbol:  name,
			Returns: "text",
			Impl: func(ctx *sql.Context, args ...any) (any, error) {
				return name, nil
			},
		}
	}
	ext := &extdef.Extension{
		Name:    "doltgres_test_update",
		Control: extdef.Control{DefaultVersion: "1.0", Comment: "test extension whose declaration changes between releases"},
		Routines: []extdef.Routine{
			updateTestRoutines["dgupdate_kept"], updateTestRoutines["dgupdate_removed"], updateTestRoutines["dgupdate_added"],
		},
	}
	extensions.Register(ext)
	return ext
})

func TestAlterExtensionUpdatesDeclaration(t *testing.T) {
	ext := registerUpdateTestExtension()
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() {
		conn.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	c := conn.Default
	call := func(name string) (string, string) {
		t.Helper()
		var result string
		var pgErr *pgconn.PgError
		if err := c.QueryRow(ctx, "SELECT "+name+"();").Scan(&result); err != nil {
			require.ErrorAs(t, err, &pgErr)
			return "", pgErr.Code
		}
		return result, ""
	}

	ext.Routines = []extdef.Routine{updateTestRoutines["dgupdate_kept"], updateTestRoutines["dgupdate_removed"]}
	_, err := c.Exec(ctx, "CREATE EXTENSION doltgres_test_update;")
	require.NoError(t, err)
	_, code := call("dgupdate_added")
	require.Equal(t, "42883", code)

	ext.Routines = []extdef.Routine{updateTestRoutines["dgupdate_kept"], updateTestRoutines["dgupdate_added"]}
	result, _ := call("dgupdate_removed")
	require.Equal(t, "dgupdate_removed", result)
	_, err = c.Exec(ctx, "ALTER EXTENSION doltgres_test_update UPDATE;")
	require.NoError(t, err)
	result, _ = call("dgupdate_kept")
	require.Equal(t, "dgupdate_kept", result)
	result, _ = call("dgupdate_added")
	require.Equal(t, "dgupdate_added", result)
	_, code = call("dgupdate_removed")
	require.Equal(t, "42883", code)

	ext.Routines = []extdef.Routine{updateTestRoutines["dgupdate_kept"]}
	_, err = c.Exec(ctx, "DROP EXTENSION doltgres_test_update;")
	require.NoError(t, err)
	_, err = c.Exec(ctx, "CREATE EXTENSION doltgres_test_update;")
	require.NoError(t, err)
	ext.Routines = []extdef.Routine{updateTestRoutines["dgupdate_kept"], updateTestRoutines["dgupdate_added"]}
	_, err = c.Exec(ctx, "DROP EXTENSION doltgres_test_update;")
	require.NoError(t, err)
	_, code = call("dgupdate_kept")
	require.Equal(t, "42883", code)
}
