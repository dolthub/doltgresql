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
	"context"
	goerrors "errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	dserver "github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/server/initialization"
	"github.com/dolthub/doltgresql/servercfg"
)

// externalServerBinaryEnvVar is the environment variable that, when set, redirects the test framework to run tests
// against an external Doltgres process rather than an in-process server. Its value names the binary to execute to
// start the server (either a path, or a name to be found on the PATH).
const externalServerBinaryEnvVar = "DOLTGRES_TEST_SERVER_BINARY"

// externalServerStartTimeout is how long to wait for an external server to begin accepting connections.
const externalServerStartTimeout = 30 * time.Second

// externalServerStopTimeout is how long to wait for an external server to exit after being interrupted, before it is
// forcibly killed.
const externalServerStopTimeout = 10 * time.Second

// externalServerBinary returns the binary named by |externalServerBinaryEnvVar|, and whether it was set.
// externalPostgres selects PostgreSQL process startup rather than Doltgres CLI flags.
func externalPostgres() bool {
	return os.Getenv("DOLTGRES_TEST_SERVER_KIND") == "postgres"
}

func externalServerBinary() (string, bool) {
	binary, ok := os.LookupEnv(externalServerBinaryEnvVar)
	return binary, ok && len(binary) > 0
}

// externalServer is a Doltgres server running in a separate process, used in place of an in-process server when
// |externalServerBinaryEnvVar| is set.
type externalServer struct {
	cmd     *exec.Cmd
	dir     string
	port    int
	exited  chan struct{}
	exitErr error
}

// createExternalServer starts an external Doltgres server on the given port using the given configuration, returning a
// connection to |database|. The returned [svcs.Controller] stops the process (and then cleans up after it) when
// stopped, mirroring the behavior of the in-process servers.
func createExternalServer(t *testing.T, binary string, database string, port int, config *servercfg.DoltgresConfig) (context.Context, *Connection, *svcs.Controller) {
	require.NotEmpty(t, database)
	config = testServerConfig(config, port)
	initializeTestProcess(config)
	server, err := newExternalServer(t, binary, config, port)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, server.stop())
		require.NoError(t, server.cleanup())
	})

	controller := svcs.NewController()
	require.NoError(t, controller.Register(svcs.AnonService{
		InitF: func(ctx context.Context) error {
			return server.start()
		},
		StopF: func(svcs.RunState) error {
			return goerrors.Join(server.stop(), server.cleanup())
		},
	}))
	go func() {
		_ = controller.Start(context.Background())
	}()
	require.NoError(t, controller.WaitForStart())
	// Callers stop the controller themselves, but a panicking test skips their deferred stops, which would otherwise
	// leave the process running after the test binary exits. Stopping an already stopped controller is a no-op.
	t.Cleanup(controller.Stop)
	if err = server.waitForListener(); err != nil {
		controller.Stop()
		require.NoError(t, err)
	}
	fmt.Printf("port is %d\n", port)

	ctx := context.Background()
	connection := newTestDatabaseConnection(t, ctx, database, serverHost, port)
	return ctx, connection, controller
}

// initializeTestProcess performs the same one-time initialization that an in-process server performs on startup. The
// test framework relies on this initialization outside of the server (such as the type and function registries used
// when normalizing rows), so it must still occur within this process when the server runs externally.
func initializeTestProcess(config *servercfg.DoltgresConfig) {
	ctx := context.Background()
	dEnv := env.Load(ctx, env.GetCurrentUserHomeDir, filesys.EmptyInMemFS(""), doltdb.InMemDoltDB, dserver.Version)
	initialization.Initialize(dEnv, config)
}

// newExternalServer prepares (but does not start) an external server process. A new temporary directory is created to
// hold the server's configuration and data, so that each server begins from a clean slate.
func newExternalServer(t *testing.T, binary string, config *servercfg.DoltgresConfig, port int) (*externalServer, error) {
	// t.Name() contains "/" for subtests, which os.MkdirTemp rejects as a path separator in the pattern.
	safeName := strings.ReplaceAll(t.Name(), "/", "_")
	dir, err := os.MkdirTemp(os.TempDir(), safeName)
	if err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			_ = os.RemoveAll(dir)
		}
	}()
	dataDir := filepath.Join(dir, "databases")
	if err = os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	configCopy := *config
	configCopy.DataDirStr = &dataDir
	configData, err := yaml.Marshal(configCopy.DoltgresConfig)
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(dir, "config.yaml")
	if err = os.WriteFile(configPath, configData, 0644); err != nil {
		return nil, err
	}

	cmd := exec.Command(binary, "--config", configPath, "--data-dir", dataDir)
	if externalPostgres() {
		path, err := exec.LookPath(binary)
		if err != nil {
			return nil, err
		}
		initdb := exec.Command(filepath.Join(filepath.Dir(path), "initdb"), "-D", dataDir, "-U", "postgres", "--auth=trust", "--encoding=UTF8", "--locale=C", "--no-sync")
		if output, err := initdb.CombinedOutput(); err != nil {
			return nil, errors.Errorf("initdb failed: %s: %w", output, err)
		}
		cmd = exec.Command(path, "-D", dataDir, "-p", strconv.Itoa(port), "-h", serverHost, "-k", dir, "-c", "default_table_access_method=dolt_prolly_table", "-c", "max_prepared_transactions=10")
	}
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	success = true
	return &externalServer{
		cmd:    cmd,
		dir:    dir,
		port:   port,
		exited: make(chan struct{}),
	}, nil
}

// start launches the server process.
func (s *externalServer) start() error {
	if err := s.cmd.Start(); err != nil {
		return errors.Errorf("unable to start external server `%s`: %w", s.cmd.Path, err)
	}
	go func() {
		s.exitErr = s.cmd.Wait()
		close(s.exited)
	}()
	return nil
}

// waitForListener blocks until the server process is accepting connections on its port, returning an error if the
// process exits or does not begin listening in time.
func (s *externalServer) waitForListener() error {
	address := net.JoinHostPort(serverHost, strconv.Itoa(s.port))
	deadline := time.Now().Add(externalServerStartTimeout)
	for time.Now().Before(deadline) {
		select {
		case <-s.exited:
			return errors.Errorf("external server exited before accepting connections: %v", s.exitErr)
		default:
		}
		conn, err := net.DialTimeout("tcp", address, time.Second)
		if err == nil {
			return conn.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.Errorf("external server did not accept connections on %s within %v", address, externalServerStartTimeout)
}

// stop interrupts the server process and waits for it to exit, killing it if it does not exit in time.
func (s *externalServer) stop() error {
	if s.cmd.Process == nil {
		return nil
	}
	select {
	case <-s.exited:
		return nil
	default:
	}
	// Interrupt is not supported on all platforms (such as Windows), in which case we fall back to killing the process.
	if err := s.cmd.Process.Signal(os.Interrupt); err != nil {
		_ = s.cmd.Process.Kill()
	}
	select {
	case <-s.exited:
		return nil
	case <-time.After(externalServerStopTimeout):
		_ = s.cmd.Process.Kill()
		<-s.exited
		return errors.Errorf("external server did not exit within %v of being interrupted and was killed", externalServerStopTimeout)
	}
}

// cleanup is called after the external server process has stopped, and is responsible for removing any state that
// would otherwise allow one test to affect another. This is only executed when |externalServerBinaryEnvVar| is set.
func (s *externalServer) cleanup() error {
	return os.RemoveAll(s.dir)
}
