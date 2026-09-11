// Copyright 2024 Dolthub, Inc.
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

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	driver "github.com/dolthub/doltgresql/integration-tests/go-sql-server-driver/driver"
)

// newPorts returns a DynamicResources bound to the global port pool for use in
// standalone Go tests.
func newPorts(t *testing.T) *DynamicResources {
	ports := &DynamicResources{}
	ports.global = &GlobalPorts
	ports.t = t
	return ports
}

// StartServer is the doltgres convenience wrapper used by standalone Go tests
// to start a server for the database |dbName| in the store |rs|. The doltgres
// server runs from the store directory (the data-dir) and serves |dbName| as a
// database. The returned server's DBName is set to |dbName| so connections
// target it by default.
func StartServer(t *testing.T, rs driver.RepoStore, dbName string, s *driver.Server, ports *DynamicResources) *driver.SqlServer {
	server := MakeServer(t, rs, rs.Dir, s, ports)
	if server != nil {
		server.DBName = dbName
	}
	return server
}

// RunServerUntilEndOfTest runs a server until the end of the test. Because we
// do not return the server for doing things like making connections to it,
// this is only useful for asserting the behavior of other commands which
// interact with the server.
func RunServerUntilEndOfTest(t *testing.T, rs driver.RepoStore, s *driver.Server, ports *DynamicResources) {
	server := MakeServer(t, rs, rs.Dir, s, ports)
	require.NotNil(t, server)
	db, err := server.DB(driver.Connection{})
	require.NoError(t, err)
	require.NoError(t, db.Close())
}

// GC tests observe informational server logs to detect completion, so generating
// a listener configuration must not suppress the default or configured log level.
func TestPrepareDoltgresServerArgsLogging(t *testing.T) {
	for _, level := range []string{"", "info", "warn", "trace"} {
		name := level
		if name == "" {
			name = "default"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			var args []string
			if level != "" {
				require.NoError(t, os.WriteFile(filepath.Join(dir, "server.yaml"), []byte("log_level: "+level+"\n"), 0600))
				args = []string{"--config", "server.yaml"}
			}
			prepareDoltgresServerArgs(t, dir, "test", 5432, args)
			contents, err := os.ReadFile(filepath.Join(dir, ".generated-test-config.yaml"))
			require.NoError(t, err)
			var config map[string]any
			require.NoError(t, yaml.Unmarshal(contents, &config))
			if level == "" {
				require.NotContains(t, config, "log_level", "preserve the server's default logging")
			} else {
				require.Equal(t, level, config["log_level"])
			}
		})
	}
}
