package enginetest

import (
	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/doltgresql/server"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
	"time"

	"github.com/dolthub/doltgresql/servercfg"
	"github.com/dolthub/doltgresql/servercfg/cfgdetails"
)

// Ptr is a helper function that returns a pointer to the value passed in. This is necessary to e.g. get a pointer to
// a const value without assigning to an intermediate variable.
func Ptr[T any](v T) *T {
	return &v
}

// startServer will start sql-server with given host, unix socket file path and whether to use specific port, which is defined randomly.
func startServer(t *testing.T, host string, unixSocketPath string) (*svcs.Controller, *servercfg.DoltgresConfig) {
	rand.Seed(time.Now().UnixNano())
	port := 15403 + rand.Intn(25)

	doltgresConfig := servercfg.DoltgresConfig{
		DoltgresConfig: cfgdetails.DoltgresConfig{
			LogLevelStr: Ptr("debug"),
			ListenerConfig: &cfgdetails.DoltgresListenerConfig{
				//HostStr:    Ptr("localhost"),
				PortNumber: &port,
				//Socket:     Ptr(unixSocketPath),
			},
		},
	}
	ctrl, err := server.RunInMemory(&doltgresConfig, server.NewListener)
	require.NoError(t, err)

	return ctrl, &doltgresConfig
}
