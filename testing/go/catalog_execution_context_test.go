package _go

import (
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

// Runtime type handles describe immutable revisions, including when explicitly
// supplied in extended-protocol Parse messages. This tests the fork's native
// composite wire protocol; it is not a Doltgres composite-support assertion.
func TestPostgresCatalogTypedParameters(t *testing.T) {
	if os.Getenv("DOLTGRES_TEST_SERVER_KIND") != "postgres" {
		t.Skip("canonical runtime type handles are PostgreSQL-fork machinery")
	}
	ctx, connection, controller := CreateServer(t, "postgres")
	defer func() { connection.Close(ctx); controller.Stop(); require.NoError(t, controller.WaitForStop()) }()
	conn := connection.Default
	exec := func(query string) { _, err := connection.Exec(ctx, query); require.NoError(t, err) }
	for _, query := range []string{
		"CREATE TABLE t(id int PRIMARY KEY,v text)",
		"SELECT dolt_commit('-Am','base')",
		"SELECT dolt_checkout('-b','other')",
		"ALTER TABLE t ADD COLUMN extra int",
		"SELECT dolt_commit('-Am','other')",
		"SELECT dolt_checkout('main')",
	} {
		exec(query)
	}
	var mainType uint32
	require.NoError(t, conn.QueryRow(ctx, "SELECT 't'::regtype::oid").Scan(&mainType))
	description, err := conn.PgConn().Prepare(ctx, "old_type", "SELECT ($1).v", []uint32{mainType})
	require.NoError(t, err)
	require.Equal(t, []uint32{mainType}, description.ParamOIDs)
	executeOld := func(value string, expected string) {
		result := conn.PgConn().ExecPrepared(ctx, "old_type", [][]byte{[]byte(value)}, []int16{0}, []int16{0}).Read()
		require.NoError(t, result.Err)
		require.Equal(t, [][][]byte{{[]byte(expected)}}, result.Rows)
	}
	executeOld("(1,main)", "main")
	exec("SELECT dolt_checkout('other')")
	executeOld("(2,retained)", "retained")
	var otherType uint32
	require.NoError(t, conn.QueryRow(ctx, "SELECT 't'::regtype::oid").Scan(&otherType))
	require.NotEqual(t, mainType, otherType)
	result := conn.PgConn().ExecParams(ctx, "SELECT ($1).extra", [][]byte{[]byte("(3,other,7)")}, []uint32{otherType}, []int16{0}, []int16{0}).Read()
	require.NoError(t, result.Err)
	require.Equal(t, [][][]byte{{[]byte("7")}}, result.Rows)
	result = conn.PgConn().ExecPrepared(ctx, "old_type", [][]byte{[]byte("(3,other,7)")}, []int16{0}, []int16{0}).Read()
	var pgerr *pgconn.PgError
	require.ErrorAs(t, result.Err, &pgerr)
	require.Equal(t, "22P02", pgerr.Code)
	require.Contains(t, pgerr.Detail, "Too many columns")
	executeOld("(4,recovered)", "recovered")
}
