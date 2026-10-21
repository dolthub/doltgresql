package _go

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExternalServerCleanup(t *testing.T) {
	dir, err := os.MkdirTemp("", "doltgres-cleanup-")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "state"), []byte("old state"), 0600))
	s := &externalServer{dir: dir, cmd: exec.Command("not-started")}
	require.NoError(t, s.stop())
	require.NoError(t, s.cleanup())
	require.NoError(t, s.cleanup())
	_, err = os.Stat(dir)
	require.True(t, os.IsNotExist(err))
}

func TestExternalPostgresIsolation(t *testing.T) {
	if !externalPostgres() {
		t.Skip("PostgreSQL external backend only")
	}
	for i := 0; i < 2; i++ {
		ctx, conn, controller := CreateServer(t, "isolated")
		var count int
		require.NoError(t, conn.Default.QueryRow(ctx, "SELECT count(*) FROM pg_dolt_ref").Scan(&count))
		require.Zero(t, count)
		_, err := conn.Exec(ctx, "CREATE TABLE reused (id int PRIMARY KEY)")
		require.NoError(t, err)
		_, err = conn.Exec(ctx, "INSERT INTO reused VALUES (1)")
		require.NoError(t, err)
		_, err = conn.Exec(ctx, "SELECT dolt_commit('-Am','first')")
		require.NoError(t, err)
		conn.Close(ctx)
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}
}
