package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestApplicationSettingsWire(t *testing.T) {
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() {
		conn.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	c := conn.Default
	read := func(name string) string {
		t.Helper()
		var value string
		require.NoError(t, c.QueryRow(ctx, "SELECT current_setting($1)", name).Scan(&value))
		return value
	}
	var missing *string
	var shown string
	require.NoError(t, c.QueryRow(ctx, "SELECT current_setting('app.tenant', true)").Scan(&missing))
	require.Nil(t, missing)
	var pgErr *pgconn.PgError
	err := c.QueryRow(ctx, "SELECT current_setting('app.never')").Scan(&shown)
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "42704", pgErr.Code)
	_, err = c.Exec(ctx, "SET app.tenant = 'session-a'")
	require.NoError(t, err)
	require.Equal(t, "session-a", read("app.tenant"))
	_, err = c.Exec(ctx, "SET app.tenant = 42")
	require.NoError(t, err)
	require.Equal(t, "42", read("app.tenant"))
	_, err = c.Exec(ctx, "SET app.tenant = unquoted")
	require.NoError(t, err)
	require.Equal(t, "unquoted", read("app.tenant"))
	_, err = c.Exec(ctx, "SET app.tenant = 'session-a'")
	require.NoError(t, err)
	require.Equal(t, "session-a", read("app.tenant"))
	_, err = c.Prepare(ctx, "read_tenant", "SELECT current_setting('app.tenant')")
	require.NoError(t, err)
	readPrepared := func() string {
		t.Helper()
		var value string
		require.NoError(t, c.QueryRow(ctx, "read_tenant").Scan(&value))
		return value
	}
	require.Equal(t, "session-a", readPrepared())
	require.NoError(t, c.QueryRow(ctx, "SHOW app.tenant").Scan(&shown))
	require.Equal(t, "session-a", shown)
	_, err = c.Exec(ctx, "BEGIN")
	require.NoError(t, err)
	_, err = c.Exec(ctx, "SET LOCAL app.tenant = 'local-a'")
	require.NoError(t, err)
	require.Equal(t, "local-a", read("app.tenant"))
	require.Equal(t, "local-a", readPrepared())
	_, err = c.Exec(ctx, "SAVEPOINT request")
	require.NoError(t, err)
	var returned string
	require.NoError(t, c.QueryRow(ctx, "SELECT set_config('app.tenant', 'local-b', true)").Scan(&returned))
	require.Equal(t, "local-b", returned)
	require.Equal(t, "local-b", read("app.tenant"))
	_, err = c.Exec(ctx, "ROLLBACK TO request")
	require.NoError(t, err)
	require.Equal(t, "local-a", read("app.tenant"))
	_, err = c.Exec(ctx, "COMMIT")
	require.NoError(t, err)
	require.Equal(t, "session-a", read("app.tenant"))
	require.Equal(t, "session-a", readPrepared())
	tag, err := c.Exec(ctx, "RESET app.tenant")
	require.NoError(t, err)
	require.Equal(t, "RESET", tag.String())
	require.Equal(t, "", read("app.tenant"))
	_, err = c.Exec(ctx, "SET app.tenant = 'session-b'")
	require.NoError(t, err)
	require.Equal(t, "session-b", read("app.tenant"))
	require.Equal(t, "session-b", readPrepared())
	_, err = c.Exec(ctx, "SET search_path = public")
	require.NoError(t, err)
	require.Equal(t, "public", read("search_path"))
	_, err = c.Exec(ctx, "BEGIN")
	require.NoError(t, err)
	require.NoError(t, c.QueryRow(ctx, "SELECT set_config('search_path', 'other', true)").Scan(&returned))
	require.Equal(t, "other", returned)
	require.Equal(t, "other", read("search_path"))
	_, err = c.Exec(ctx, "SAVEPOINT settings")
	require.NoError(t, err)
	_, err = c.Exec(ctx, "SET LOCAL row_security = off")
	require.NoError(t, err)
	require.Equal(t, "off", read("row_security"))
	_, err = c.Exec(ctx, "ROLLBACK TO settings")
	require.NoError(t, err)
	require.Equal(t, "on", read("row_security"))
	_, err = c.Exec(ctx, "COMMIT")
	require.NoError(t, err)
	require.Equal(t, "public", read("search_path"))
	require.NoError(t, c.QueryRow(ctx, "SELECT set_config('search_path', NULL, false)").Scan(&returned))
	require.Equal(t, `"$user", public`, returned)
	_, err = c.Exec(ctx, "SET search_path = public")
	require.NoError(t, err)
	tag, err = c.Exec(ctx, "RESET ALL")
	require.NoError(t, err)
	require.Equal(t, "RESET", tag.String())
	require.Equal(t, `"$user", public`, read("search_path"))
	require.Equal(t, "", read("app.tenant"))
	err = c.QueryRow(ctx, "SELECT set_config('bad..name', 'x', false)").Scan(&returned)
	pgErr = nil
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "42602", pgErr.Code)
	_, err = c.Exec(ctx, "SET app.tenant = 'session-c'")
	require.NoError(t, err)
	_, err = c.Exec(ctx, "DISCARD ALL")
	require.NoError(t, err)
	require.NoError(t, c.QueryRow(ctx, "SELECT current_setting('app.tenant', true)").Scan(&missing))
	require.Nil(t, missing)
	require.Equal(t, `"$user", public`, read("search_path"))
}

// TestApplicationSettingsExecution checks that the scoped values read by SQL
// identity functions are also used for table resolution and SHOW.
func TestApplicationSettingsExecution(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "built-in settings reach execution and restore at transaction boundaries",
		SetUpScript: []string{
			"CREATE SCHEMA settings_schema",
			"CREATE TABLE settings_schema.settings_table (v int)",
			"INSERT INTO settings_schema.settings_table VALUES (42)",
			"SET search_path = public",
			"SET row_security = on",
		},
		Assertions: []ScriptTestAssertion{
			{Query: "BEGIN"},
			{Query: "SAVEPOINT settings"},
			{Query: "SET LOCAL search_path = settings_schema"},
			{Query: "SELECT v FROM settings_table", Expected: []sql.Row{{42}}},
			{Query: "SHOW search_path", Expected: []sql.Row{{"settings_schema"}}},
			{Query: "SET row_security = off"},
			{Query: "COMMIT"},
			{Query: "SHOW search_path", Expected: []sql.Row{{"public"}}},
			{Query: "SELECT current_setting('row_security')", Expected: []sql.Row{{"off"}}},
			{Query: "BEGIN"},
			{Query: "SET search_path = settings_schema"},
			{Query: "SELECT v FROM settings_table", Expected: []sql.Row{{42}}},
			{Query: "ROLLBACK"},
			{Query: "SHOW search_path", Expected: []sql.Row{{"public"}}},
			{Query: "SELECT v FROM settings_table", ExpectedErr: "table not found: settings_table"},
		},
	}})
}
