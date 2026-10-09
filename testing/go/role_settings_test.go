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
	"fmt"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

// TestRoleSettings checks ALTER ROLE ... SET and ALTER DATABASE ... SET, along with the catalog tables that show them.
func TestRoleSettings(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "ALTER ROLE SET and RESET",
			SetUpScript: []string{
				"CREATE ROLE settings_user LOGIN PASSWORD 'pw'",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "ALTER ROLE settings_user SET work_mem = 8192"},
				{Query: "ALTER ROLE settings_user SET app.tenant TO 'acme'"},
				{Query: "ALTER USER settings_user SET search_path = public, \"Other\""},
				{
					Query:    "SELECT rolconfig FROM pg_roles WHERE rolname = 'settings_user'",
					Expected: []sql.Row{{`{work_mem=8192,app.tenant=acme,"search_path=public, \"Other\""}`}},
				},
				{
					Query:    "SELECT setdatabase, setconfig FROM pg_db_role_setting s JOIN pg_roles r ON s.setrole = r.oid WHERE r.rolname = 'settings_user'",
					Expected: []sql.Row{{0, `{work_mem=8192,app.tenant=acme,"search_path=public, \"Other\""}`}},
				},
				{
					// Setting an existing parameter replaces its value without changing its position
					Query: "ALTER ROLE settings_user SET WORK_MEM = 16384",
				},
				{
					Query:    "SELECT useconfig FROM pg_user WHERE usename = 'settings_user'",
					Expected: []sql.Row{{`{work_mem=16384,app.tenant=acme,"search_path=public, \"Other\""}`}},
				},
				{Query: "ALTER ROLE settings_user RESET app.tenant"},
				{Query: "ALTER ROLE settings_user SET work_mem TO DEFAULT"},
				{
					Query:    "SELECT rolconfig FROM pg_roles WHERE rolname = 'settings_user'",
					Expected: []sql.Row{{`{"search_path=public, \"Other\""}`}},
				},
				{Query: "ALTER ROLE settings_user RESET ALL"},
				{
					Query:    "SELECT rolconfig FROM pg_roles WHERE rolname = 'settings_user'",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    "SELECT count(*) FROM pg_db_role_setting",
					Expected: []sql.Row{{0}},
				},
			},
		},
		{
			Name: "ALTER DATABASE SET and ALTER ROLE IN DATABASE",
			SetUpScript: []string{
				"CREATE ROLE settings_db_user",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "ALTER DATABASE postgres SET timezone = 'UTC'"},
				{Query: "ALTER ROLE settings_db_user IN DATABASE postgres SET app.scope = 'both'"},
				{Query: "ALTER ROLE ALL SET app.scope = 'global'"},
				{
					Query: "SELECT d.datname, r.rolname, s.setconfig FROM pg_db_role_setting s " +
						"LEFT JOIN pg_database d ON s.setdatabase = d.oid LEFT JOIN pg_roles r ON s.setrole = r.oid " +
						"ORDER BY d.datname NULLS FIRST, r.rolname NULLS FIRST",
					Expected: []sql.Row{
						{nil, nil, "{app.scope=global}"},
						{"postgres", nil, "{TimeZone=UTC}"},
						{"postgres", "settings_db_user", "{app.scope=both}"},
					},
				},
				{
					// ALTER ROLE ALL IN DATABASE is the same target as ALTER DATABASE
					Query: "ALTER ROLE ALL IN DATABASE postgres RESET timezone",
				},
				{Query: "ALTER DATABASE postgres RESET ALL"},
				{Query: "ALTER ROLE ALL RESET ALL"},
				{Query: "ALTER ROLE settings_db_user IN DATABASE postgres RESET ALL"},
				{
					Query:    "SELECT count(*) FROM pg_db_role_setting",
					Expected: []sql.Row{{0}},
				},
			},
		},
		{
			Name: "SET FROM CURRENT",
			SetUpScript: []string{
				"CREATE ROLE settings_current",
				"SET app.color = 'blue'",
				"SET work_mem = 12288",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "ALTER ROLE settings_current SET app.color FROM CURRENT"},
				{Query: "ALTER ROLE settings_current SET work_mem FROM CURRENT"},
				{
					Query:    "SELECT rolconfig FROM pg_roles WHERE rolname = 'settings_current'",
					Expected: []sql.Row{{"{app.color=blue,work_mem=12288}"}},
				},
			},
		},
		{
			Name: "dropping a role removes its settings",
			Assertions: []ScriptTestAssertion{
				{Query: "CREATE ROLE settings_dropped"},
				{Query: "ALTER ROLE settings_dropped SET app.value = 'x'"},
				{Query: "ALTER ROLE settings_dropped IN DATABASE postgres SET app.value = 'y'"},
				{Query: "SELECT count(*) FROM pg_db_role_setting", Expected: []sql.Row{{2}}},
				{Query: "DROP ROLE settings_dropped"},
				{Query: "SELECT count(*) FROM pg_db_role_setting", Expected: []sql.Row{{0}}},
			},
		},
		{
			Name: "dropping a database removes its settings",
			SetUpScript: []string{
				"CREATE ROLE settings_dropped_db_user",
				"CREATE DATABASE settings_dropped_db",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "ALTER DATABASE settings_dropped_db SET app.value = 'x'"},
				{Query: "ALTER ROLE settings_dropped_db_user IN DATABASE settings_dropped_db SET app.value = 'y'"},
				{Query: "ALTER ROLE settings_dropped_db_user SET app.value = 'z'"},
				{Query: "SELECT count(*) FROM pg_db_role_setting", Expected: []sql.Row{{3}}},
				{Query: "DROP DATABASE settings_dropped_db"},
				{
					// Settings that are not specific to the database remain
					Query:    "SELECT setdatabase, setconfig FROM pg_db_role_setting",
					Expected: []sql.Row{{0, "{app.value=z}"}},
				},
				{
					// A new database with the same name does not inherit the old settings
					Query: "CREATE DATABASE settings_dropped_db",
				},
				{Query: "SELECT count(*) FROM pg_db_role_setting", Expected: []sql.Row{{1}}},
				{Query: "DROP DATABASE IF EXISTS settings_dropped_db"},
				{Query: "DROP DATABASE IF EXISTS settings_dropped_db"},
				{Query: "SELECT count(*) FROM pg_db_role_setting", Expected: []sql.Row{{1}}},
			},
		},
		{
			Name: "errors",
			SetUpScript: []string{
				"CREATE ROLE settings_err",
			},
			Assertions: []ScriptTestAssertion{
				{Query: "ALTER ROLE missing_role SET work_mem = 8192", ExpectedErr: `role "missing_role" does not exist`},
				{Query: "ALTER DATABASE missing_db SET work_mem = 8192", ExpectedErr: `database "missing_db" does not exist`},
				{Query: "ALTER ROLE settings_err IN DATABASE missing_db SET work_mem = 8192", ExpectedErr: `database "missing_db" does not exist`},
				{Query: "ALTER ROLE settings_err SET not_a_parameter = 1", ExpectedErr: `unrecognized configuration parameter "not_a_parameter"`},
				{Query: "ALTER ROLE settings_err SET archive_mode = on", ExpectedErr: "cannot be changed now"},
				{Query: "ALTER ROLE settings_err SET role = 'postgres'", ExpectedErr: "cannot be changed now"},
				{Query: "ALTER ROLE settings_err SET app.unset FROM CURRENT", ExpectedErr: `unrecognized configuration parameter "app.unset"`},
				{Query: "SELECT count(*) FROM pg_db_role_setting", Expected: []sql.Row{{0}}},
			},
		},
		{
			Name: "permissions",
			SetUpScript: []string{
				"CREATE ROLE settings_self LOGIN PASSWORD 'pw'",
				"CREATE ROLE settings_other",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "ALTER ROLE settings_self SET app.mine = 'yes'",
					Username: "settings_self",
					Password: "pw",
				},
				{
					Query:    "ALTER ROLE settings_self IN DATABASE postgres SET app.mine = 'in_database'",
					Username: "settings_self",
					Password: "pw",
				},
				{
					Query:       "ALTER ROLE settings_other SET app.theirs = 'no'",
					Username:    "settings_self",
					Password:    "pw",
					ExpectedErr: `permission denied to alter role "settings_other"`,
				},
				{
					Query:       "ALTER ROLE postgres SET app.theirs = 'no'",
					Username:    "settings_self",
					Password:    "pw",
					ExpectedErr: "must be superuser to alter superusers",
				},
				{
					Query:       "ALTER ROLE ALL SET app.theirs = 'no'",
					Username:    "settings_self",
					Password:    "pw",
					ExpectedErr: "permission denied to alter setting",
				},
				{
					Query:       "ALTER DATABASE postgres SET app.theirs = 'no'",
					Username:    "settings_self",
					Password:    "pw",
					ExpectedErr: "must be owner of database postgres",
				},
				{
					Query:       "ALTER ROLE ALL IN DATABASE postgres SET app.theirs = 'no'",
					Username:    "settings_self",
					Password:    "pw",
					ExpectedErr: "must be owner of database postgres",
				},
				{
					Query:    "ALTER ROLE CURRENT_USER SET app.current = 'c'",
					Username: "settings_self",
					Password: "pw",
				},
				{
					Query:    "ALTER USER SESSION_USER SET app.session = 's'",
					Username: "settings_self",
					Password: "pw",
				},
				{
					Query:    "SELECT rolconfig FROM pg_roles WHERE rolname = 'settings_self'",
					Expected: []sql.Row{{"{app.mine=yes,app.current=c,app.session=s}"}},
				},
				{
					Query:    "SELECT count(*) FROM pg_db_role_setting",
					Expected: []sql.Row{{2}},
				},
			},
		},
	})
}

// TestRoleSettingsAppliedAtConnection checks that settings from ALTER ROLE and ALTER DATABASE are applied when a
// session starts, with the more specific settings taking precedence.
func TestRoleSettingsAppliedAtConnection(t *testing.T) {
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() {
		conn.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	exec := func(c *pgx.Conn, statement string) {
		t.Helper()
		_, err := c.Exec(ctx, statement)
		require.NoError(t, err, statement)
	}
	for _, statement := range []string{
		"CREATE DATABASE settings_db",
		"CREATE ROLE settings_login LOGIN PASSWORD 'pw'",
		"CREATE ROLE settings_plain LOGIN PASSWORD 'pw'",
		"GRANT ALL PRIVILEGES ON DATABASE settings_db TO settings_login",
		"GRANT ALL PRIVILEGES ON DATABASE settings_db TO settings_plain",
		"ALTER ROLE ALL SET app.level = 'global'",
		"ALTER ROLE ALL SET app.global_only = 'g'",
		"ALTER DATABASE settings_db SET app.level = 'database'",
		"ALTER DATABASE settings_db SET app.database_only = 'd'",
		"ALTER ROLE settings_login SET app.level = 'role'",
		"ALTER ROLE settings_login SET search_path = other_schema, public",
		"ALTER ROLE settings_login SET work_mem = 16384",
		"ALTER ROLE settings_login IN DATABASE settings_db SET app.level = 'role_in_database'",
	} {
		exec(conn.Default, statement)
	}
	port := conn.Default.Config().Port
	connect := func(user string, database string) *pgx.Conn {
		t.Helper()
		cfg, err := pgx.ParseConfig(fmt.Sprintf("postgres://%s:pw@127.0.0.1:%d/%s", user, port, database))
		require.NoError(t, err)
		// Settings that cannot be applied are reported as warnings, so we fail on any
		cfg.OnNotice = func(_ *pgconn.PgConn, n *pgconn.Notice) { t.Errorf("unexpected notice: %s %s", n.Severity, n.Message) }
		c, err := pgx.ConnectConfig(ctx, cfg)
		require.NoError(t, err)
		return c
	}
	read := func(c *pgx.Conn, name string) string {
		t.Helper()
		var value string
		require.NoError(t, c.QueryRow(ctx, "SELECT current_setting($1)", name).Scan(&value))
		return value
	}

	// The role within the database is the most specific setting
	roleInDatabase := connect("settings_login", "settings_db")
	defer roleInDatabase.Close(ctx)
	require.Equal(t, "role_in_database", read(roleInDatabase, "app.level"))
	require.Equal(t, "d", read(roleInDatabase, "app.database_only"))
	require.Equal(t, "g", read(roleInDatabase, "app.global_only"))
	require.Equal(t, "other_schema, public", read(roleInDatabase, "search_path"))
	require.Equal(t, "16384", read(roleInDatabase, "work_mem"))

	// The search_path from the role is used to resolve unqualified names in new sessions
	admin, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/settings_db", port))
	require.NoError(t, err)
	defer admin.Close(ctx)
	for _, statement := range []string{
		"CREATE SCHEMA other_schema",
		"CREATE TABLE other_schema.path_test (v INT)",
		"INSERT INTO other_schema.path_test VALUES (1)",
		"CREATE TABLE public.path_test (v INT)",
		"INSERT INTO public.path_test VALUES (2)",
		"GRANT ALL PRIVILEGES ON SCHEMA other_schema TO settings_login",
		"GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA other_schema TO settings_login",
	} {
		exec(admin, statement)
	}
	pathSession := connect("settings_login", "settings_db")
	defer pathSession.Close(ctx)
	var schema string
	require.NoError(t, pathSession.QueryRow(ctx, "SELECT current_schema()").Scan(&schema))
	require.Equal(t, "other_schema", schema)
	var searchPath string
	require.NoError(t, pathSession.QueryRow(ctx, "SHOW search_path").Scan(&searchPath))
	require.Equal(t, "other_schema, public", searchPath)
	var v int
	require.NoError(t, pathSession.QueryRow(ctx, "SELECT v FROM path_test").Scan(&v))
	require.Equal(t, 1, v)
	exec(pathSession, "CREATE TABLE path_created (v INT)")
	require.NoError(t, pathSession.QueryRow(ctx, "SELECT schemaname FROM pg_catalog.pg_tables WHERE tablename = 'path_created'").Scan(&schema))
	require.Equal(t, "other_schema", schema)
	// A role without the setting keeps the default search_path
	defaultPath := connect("settings_plain", "settings_db")
	defer defaultPath.Close(ctx)
	require.NoError(t, defaultPath.QueryRow(ctx, "SELECT current_schema()").Scan(&schema))
	require.Equal(t, "public", schema)

	// The role is more specific than the database
	roleOnly := connect("settings_login", "postgres")
	defer roleOnly.Close(ctx)
	require.Equal(t, "role", read(roleOnly, "app.level"))
	require.Equal(t, "other_schema, public", read(roleOnly, "search_path"))
	var missing *string
	require.NoError(t, roleOnly.QueryRow(ctx, "SELECT current_setting('app.database_only', true)").Scan(&missing))
	require.Nil(t, missing)

	// Without any role settings, the database is more specific than every role and database
	databaseOnly := connect("settings_plain", "settings_db")
	defer databaseOnly.Close(ctx)
	require.Equal(t, "database", read(databaseOnly, "app.level"))

	globalOnly := connect("settings_plain", "postgres")
	defer globalOnly.Close(ctx)
	require.Equal(t, "global", read(globalOnly, "app.level"))

	// A session may still change a setting that came from its role
	exec(roleOnly, "SET app.level = 'session'")
	require.Equal(t, "session", read(roleOnly, "app.level"))

	// Settings only apply to new sessions
	exec(conn.Default, "ALTER ROLE settings_plain SET app.level = 'changed'")
	require.Equal(t, "global", read(globalOnly, "app.level"))
	changed := connect("settings_plain", "postgres")
	defer changed.Close(ctx)
	require.Equal(t, "changed", read(changed, "app.level"))

	// The client's startup parameters take precedence over stored settings
	exec(conn.Default, "ALTER ROLE settings_plain SET timezone = 'UTC'")
	clientConfig, err := pgx.ParseConfig(fmt.Sprintf("postgres://settings_plain:pw@127.0.0.1:%d/postgres", port))
	require.NoError(t, err)
	clientConfig.RuntimeParams["timezone"] = "America/New_York"
	client, err := pgx.ConnectConfig(ctx, clientConfig)
	require.NoError(t, err)
	defer client.Close(ctx)
	require.Equal(t, "America/New_York", read(client, "timezone"))
	utc := connect("settings_plain", "postgres")
	defer utc.Close(ctx)
	require.Equal(t, "UTC", read(utc, "timezone"))
}
