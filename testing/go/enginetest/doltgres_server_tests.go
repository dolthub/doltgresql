package enginetest

import (
	"os"
	"testing"

	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/servercfg"
	"github.com/dolthub/doltgresql/servercfg/cfgdetails"
	dtest "github.com/dolthub/doltgresql/testing/go"
)

// Ptr is a helper function that returns a pointer to the value passed in. This is necessary to e.g. get a pointer to
// a const value without assigning to an intermediate variable.
func Ptr[T any](v T) *T {
	return &v
}

// startServer will start sql-server with given host, unix socket file path and whether to use specific port, which is defined randomly.
func startServer(t *testing.T, host string, unixSocketPath string) (*svcs.Controller, *servercfg.DoltgresConfig) {
	port, err := sql.GetEmptyPort()
	require.NoError(t, err)

	doltgresConfig := servercfg.DoltgresConfig{
		DoltgresConfig: cfgdetails.DoltgresConfig{
			LogLevelStr: Ptr("warn"),
			ListenerConfig: &cfgdetails.DoltgresListenerConfig{
				//HostStr:    Ptr("localhost"),
				PortNumber: &port,
				//Socket:     Ptr(unixSocketPath),
			},
		},
	}
	if os.Getenv("DOLTGRES_TEST_SERVER_BINARY") != "" {
		ctx, initial, ctrl := dtest.CreateServerWithPort(t, "postgres", port)
		initial.Close(ctx)
		return ctrl, &doltgresConfig
	}
	ctrl, err := server.RunInMemory(&doltgresConfig, server.NewListener)
	require.NoError(t, err)

	return ctrl, &doltgresConfig
}
