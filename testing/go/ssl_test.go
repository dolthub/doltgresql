package _go

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dserver "github.com/dolthub/doltgresql/server"
)

func TestSSL(t *testing.T) {
	port := GetUnusedPort(t)
	server.DefaultProtocolListenerFunc = dserver.NewLimitedListener
	controller, err := dserver.RunInMemory([]string{fmt.Sprintf("--port=%d", port), "--host=127.0.0.1"})
	require.NoError(t, err)

	defer func() {
		controller.Stop()
		require.NoError(t, controller.WaitForStop())
	}()
	
	ctx := context.Background()
	err = func() error {
		// The connection attempt may be made before the server has grabbed the port, so we'll retry the first
		// connection a few times.
		var conn *pgx.Conn
		var err error
		for i := 0; i < 3; i++ {
			conn, err = pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/?sslmode=require", port))
			if err == nil {
				break
			} else {
				time.Sleep(time.Second)
			}
		}
		if err != nil {
			return err
		}

		defer conn.Close(ctx)
		_, err = conn.Exec(ctx, "CREATE DATABASE postgres;")
		return err
	}()
	require.NoError(t, err)

	conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/postgres?sslmode=require", port))
	require.NoError(t, err)
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "CREATE TABLE test (pk INT8 PRIMARY KEY, v1 int4);")
	require.NoError(t, err)
	_, err = conn.Exec(ctx, "INSERT INTO test VALUES (3645, 37643);")
	require.NoError(t, err)
	rows, err := conn.Query(ctx, "SELECT * FROM test;")
	require.NoError(t, err)
	readRows, err := ReadRows(rows)
	require.NoError(t, err)
	assert.Equal(t, NormalizeRows([]sql.Row{{3645, 37643}}), readRows)
}
