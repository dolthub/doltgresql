package _go

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestIdentityExpressions(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "login identity expressions",
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT current_user, current_role, user, session_user, current_user()", Expected: []sql.Row{{"postgres", "postgres", "postgres", "postgres", "postgres"}}},
			{Query: "CREATE ROLE identity_reader LOGIN PASSWORD 'identity_password'"},
			{Query: "SELECT current_user, current_role, user, session_user", Username: "identity_reader", Password: "identity_password", Expected: []sql.Row{{"identity_reader", "identity_reader", "identity_reader", "identity_reader"}}},
			{Query: "DROP ROLE identity_reader"},
		},
	}})
}

func TestSetRoleWireAndTransactionScopes(t *testing.T) {
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() {
		conn.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	for _, statement := range []string{
		"CREATE ROLE role_login LOGIN PASSWORD 'role_password' NOINHERIT",
		"CREATE ROLE role_middle",
		"CREATE ROLE role_target",
		"CREATE ROLE role_unrelated",
		"CREATE SCHEMA role_login",
		"CREATE SCHEMA role_target",
		"GRANT USAGE ON SCHEMA role_login, role_target TO PUBLIC",
		"GRANT role_target TO role_middle",
		"GRANT role_middle TO role_login",
	} {
		_, err := conn.Default.Exec(ctx, statement)
		require.NoError(t, err, statement)
	}
	config := conn.Default.Config()
	reader, err := pgx.Connect(ctx, fmt.Sprintf("postgres://role_login:role_password@127.0.0.1:%d/postgres", config.Port))
	require.NoError(t, err)
	defer reader.Close(ctx)

	checkIdentity := func(wantCurrent string) {
		t.Helper()
		var current, session string
		require.NoError(t, reader.QueryRow(ctx, "SELECT current_user, session_user").Scan(&current, &session))
		require.Equal(t, wantCurrent, current)
		require.Equal(t, "role_login", session)
	}
	checkIdentity("role_login")
	var currentSchema string
	require.NoError(t, reader.QueryRow(ctx, "SELECT current_schema()").Scan(&currentSchema))
	require.Equal(t, "role_login", currentSchema)
	var roleSetting string
	require.NoError(t, reader.QueryRow(ctx, "SELECT current_setting('role')").Scan(&roleSetting))
	require.Equal(t, "none", roleSetting)
	_, err = reader.Prepare(ctx, "role_identity", "SELECT current_user, session_user")
	require.NoError(t, err)
	for _, statement := range []string{"SET ROLE role_target", "SET ROLE TO role_target", "SET ROLE = role_target"} {
		tag, execErr := reader.Exec(ctx, statement)
		err = execErr
		require.NoError(t, err, statement)
		require.Equal(t, "SET", tag.String())
		checkIdentity("role_target")
		require.NoError(t, reader.QueryRow(ctx, "SELECT current_schema()").Scan(&currentSchema))
		require.Equal(t, "role_target", currentSchema)
		require.NoError(t, reader.QueryRow(ctx, "SHOW role").Scan(&roleSetting))
		require.Equal(t, "role_target", roleSetting)
	}
	var current, session string
	require.NoError(t, reader.QueryRow(ctx, "role_identity").Scan(&current, &session))
	require.Equal(t, "role_target", current)
	require.Equal(t, "role_login", session)
	_, err = reader.Exec(ctx, "CREATE ROLE role_forbidden")
	require.Error(t, err)
	_, err = reader.Exec(ctx, "SET ROLE role_unrelated")
	require.ErrorContains(t, err, `permission denied to set role "role_unrelated"`)
	var pgErr *pgconn.PgError
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "42501", pgErr.Code)
	_, err = reader.Exec(ctx, "SET ROLE role_missing")
	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, "42704", pgErr.Code)
	checkIdentity("role_target")
	_, err = reader.Exec(ctx, "SET ROLE NONE")
	require.NoError(t, err)
	checkIdentity("role_login")
	_, err = reader.Exec(ctx, "SET ROLE role_middle")
	require.NoError(t, err)
	tag, err := reader.Exec(ctx, "RESET ROLE")
	require.NoError(t, err)
	require.Equal(t, "RESET", tag.String())
	checkIdentity("role_login")
	require.NoError(t, reader.QueryRow(ctx, "SELECT set_config('role', 'role_target', false)").Scan(&roleSetting))
	require.Equal(t, "role_target", roleSetting)
	checkIdentity("role_target")
	_, err = reader.Exec(ctx, "SELECT set_config('role', 'role_unrelated', false)")
	require.Error(t, err)
	checkIdentity("role_target")
	tag, err = reader.Exec(ctx, "SET ROLE DEFAULT")
	require.NoError(t, err)
	require.Equal(t, "SET", tag.String())
	checkIdentity("role_login")
	_, err = reader.Exec(ctx, "SET ROLE TO DEFAULT")
	require.NoError(t, err)
	checkIdentity("role_login")

	_, err = reader.Exec(ctx, "BEGIN")
	require.NoError(t, err)
	_, err = reader.Exec(ctx, "SET LOCAL ROLE role_target")
	require.NoError(t, err)
	checkIdentity("role_target")
	_, err = reader.Exec(ctx, "SAVEPOINT role_sp")
	require.NoError(t, err)
	_, err = reader.Exec(ctx, "SET LOCAL ROLE role_middle")
	require.NoError(t, err)
	checkIdentity("role_middle")
	_, err = reader.Exec(ctx, "ROLLBACK TO SAVEPOINT role_sp")
	require.NoError(t, err)
	checkIdentity("role_target")
	_, err = reader.Exec(ctx, "COMMIT")
	require.NoError(t, err)
	checkIdentity("role_login")
	_, err = reader.Exec(ctx, "BEGIN")
	require.NoError(t, err)
	_, err = reader.Exec(ctx, "SET ROLE role_target")
	require.NoError(t, err)
	checkIdentity("role_target")
	_, err = reader.Exec(ctx, "ROLLBACK")
	require.NoError(t, err)
	checkIdentity("role_login")
	_, err = reader.Exec(ctx, "BEGIN")
	require.NoError(t, err)
	require.NoError(t, reader.QueryRow(ctx, "SELECT set_config('role', 'role_target', true)").Scan(&roleSetting))
	require.Equal(t, "role_target", roleSetting)
	checkIdentity("role_target")
	_, err = reader.Exec(ctx, "COMMIT")
	require.NoError(t, err)
	checkIdentity("role_login")

	_, err = reader.Exec(ctx, "SET ROLE role_target")
	require.NoError(t, err)
	_, err = conn.Default.Exec(ctx, "REVOKE role_target FROM role_middle")
	require.NoError(t, err)
	_, err = reader.Exec(ctx, "SET ROLE role_target")
	require.Error(t, err)
	if !strings.Contains(err.Error(), "permission denied to set role") {
		t.Fatal(err)
	}
	checkIdentity("role_target") // A failed switch does not change an existing selection.
	_, err = reader.Exec(ctx, "SET ROLE NONE")
	require.NoError(t, err)
	checkIdentity("role_login")
	var postgresCurrent string
	require.NoError(t, conn.Default.QueryRow(ctx, "SELECT current_user").Scan(&postgresCurrent))
	require.Equal(t, "postgres", postgresCurrent)

	_, err = conn.Default.Exec(ctx, "CREATE TABLE role_protected (id INT PRIMARY KEY)")
	require.NoError(t, err)
	_, err = conn.Default.Exec(ctx, "INSERT INTO role_protected VALUES (7)")
	require.NoError(t, err)
	_, err = conn.Default.Prepare(ctx, "protected_query", "SELECT id FROM role_protected")
	require.NoError(t, err)
	_, err = conn.Default.Exec(ctx, "SET ROLE role_target")
	require.NoError(t, err)
	var id int
	err = conn.Default.QueryRow(ctx, "protected_query").Scan(&id)
	require.Error(t, err)
	_, err = conn.Default.Exec(ctx, "RESET ROLE")
	require.NoError(t, err)
	_, err = conn.Default.Exec(ctx, "GRANT SELECT ON role_protected TO role_target")
	require.NoError(t, err)
	_, err = conn.Default.Exec(ctx, "SET ROLE role_target")
	require.NoError(t, err)
	require.NoError(t, conn.Default.QueryRow(ctx, "SELECT id FROM role_protected").Scan(&id))
	require.NoError(t, conn.Default.QueryRow(ctx, "protected_query").Scan(&id))
	require.Equal(t, 7, id)
	_, err = conn.Default.Exec(ctx, "RESET ROLE")
	require.NoError(t, err)
	_, err = conn.Default.Exec(ctx, "REVOKE SELECT ON role_protected FROM role_target")
	require.NoError(t, err)
	_, err = conn.Default.Exec(ctx, "SET ROLE role_target")
	require.NoError(t, err)
	err = conn.Default.QueryRow(ctx, "protected_query").Scan(&id)
	require.Error(t, err)
	_, err = conn.Default.Exec(ctx, "RESET ROLE")
	require.NoError(t, err)
}

func TestSetRoleResetAfterConcurrentDrop(t *testing.T) {
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() {
		conn.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	_, err := conn.Default.Exec(ctx, "CREATE ROLE role_deleted_selected")
	require.NoError(t, err)
	config := conn.Default.Config()
	other, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/postgres", config.Port))
	require.NoError(t, err)
	defer other.Close(ctx)

	for _, reset := range []string{"RESET ROLE", "SET ROLE NONE"} {
		_, err = conn.Default.Exec(ctx, "SET ROLE role_deleted_selected")
		require.NoError(t, err)
		_, err = other.Exec(ctx, "DROP ROLE role_deleted_selected")
		require.NoError(t, err)
		_, err = other.Exec(ctx, "CREATE ROLE role_deleted_selected")
		require.NoError(t, err)
		var current string
		err = conn.Default.QueryRow(ctx, "SELECT current_user").Scan(&current)
		require.ErrorContains(t, err, "no longer exists")
		_, err = conn.Default.Exec(ctx, reset)
		require.NoError(t, err, reset)
		require.NoError(t, conn.Default.QueryRow(ctx, "SELECT current_user").Scan(&current))
		require.Equal(t, "postgres", current)
	}
}

func TestIdentityPreparedAndIndependentConnections(t *testing.T) {
	ctx, conn, controller := CreateServer(t, "postgres")
	defer func() {
		conn.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	_, err := conn.Default.Exec(ctx, "CREATE ROLE identity_reader LOGIN PASSWORD 'identity_password'")
	require.NoError(t, err)
	config := conn.Default.Config()
	reader, err := pgx.Connect(ctx, fmt.Sprintf("postgres://identity_reader:identity_password@127.0.0.1:%d/postgres", config.Port))
	require.NoError(t, err)
	defer reader.Close(ctx)

	_, err = reader.Prepare(ctx, "identity_query", "SELECT current_user, session_user")
	require.NoError(t, err)
	var current, session string
	require.NoError(t, reader.QueryRow(ctx, "identity_query").Scan(&current, &session))
	require.Equal(t, "identity_reader", current)
	require.Equal(t, "identity_reader", session)
	require.NoError(t, conn.Default.QueryRow(ctx, "SELECT current_user").Scan(&current))
	require.Equal(t, "postgres", current)

	require.NoError(t, reader.QueryRow(ctx, "identity_query").Scan(&current, &session))
	require.Equal(t, "identity_reader", current)
	require.Equal(t, "identity_reader", session)
}
