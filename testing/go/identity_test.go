package _go

import (
	"fmt"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
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
