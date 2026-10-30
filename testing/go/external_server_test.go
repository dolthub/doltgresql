package _go

import (
	"github.com/jackc/pgx/v5/pgproto3"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
	t.Run("long,comma-name-"+strings.Repeat("x", 90), func(t *testing.T) {
		for i := 0; i < 2; i++ {
			ctx, conn, controller := CreateServer(t, "isolated")
			var sharedMemory string
			require.NoError(t, conn.Default.QueryRow(ctx, "SHOW dynamic_shared_memory_type").Scan(&sharedMemory))
			require.Equal(t, "mmap", sharedMemory)
			var count int
			require.NoError(t, conn.Default.QueryRow(ctx, "SELECT count(*) FROM pg_class WHERE relname='reused'").Scan(&count))
			require.Zero(t, count) // Prior server data must not leak; initial branch refs are intentional.
			_, err := conn.Exec(ctx, "CREATE TABLE reused (id int PRIMARY KEY)")
			require.NoError(t, err)
			_, err = conn.Exec(ctx, "INSERT INTO reused VALUES (1)")
			require.NoError(t, err)
			if testTarget(t) != targetPostgres {
				_, err = conn.Exec(ctx, "SELECT dolt_commit('-Am','first')")
				require.NoError(t, err)
			}
			conn.Close(ctx)
			controller.Stop()
			require.NoError(t, controller.WaitForStop())
		}
	})
}

// A live listener can still reject startup while PostgreSQL is recovering.
func TestExternalPostgresReadiness(t *testing.T) {
	t.Setenv("DOLTGRES_TEST_SERVER_KIND", "postgres")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	require.NoError(t, listener.(*net.TCPListener).SetDeadline(time.Now().Add(5*time.Second)))
	completed := make(chan error, 1)
	go func() {
		for attempt := 0; attempt < 2; attempt++ {
			conn, err := listener.Accept()
			if err != nil {
				completed <- err
				return
			}
			_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
			backend := pgproto3.NewBackend(conn, conn)
			_, err = backend.ReceiveStartupMessage()
			if err != nil {
				conn.Close()
				completed <- err
				return
			}
			if attempt == 0 {
				backend.Send(&pgproto3.ErrorResponse{Severity: "FATAL", Code: "57P03", Message: "the database system is starting up"})
				err = backend.Flush()
				conn.Close()
				if err != nil {
					completed <- err
					return
				}
				continue
			}
			backend.Send(&pgproto3.AuthenticationOk{})
			backend.Send(&pgproto3.ParameterStatus{Name: "server_version", Value: "18.6"})
			backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
			if err = backend.Flush(); err == nil {
				_, err = backend.Receive() // pgx Ping's simple query
			}
			if err == nil {
				backend.Send(&pgproto3.CommandComplete{CommandTag: []byte("SELECT 1")})
				backend.Send(&pgproto3.ReadyForQuery{TxStatus: 'I'})
				err = backend.Flush()
			}
			conn.Close()
			completed <- err
		}
	}()
	server := &externalServer{port: listener.Addr().(*net.TCPAddr).Port, exited: make(chan struct{})}
	require.NoError(t, server.waitForListener())
	require.NoError(t, <-completed)
}
