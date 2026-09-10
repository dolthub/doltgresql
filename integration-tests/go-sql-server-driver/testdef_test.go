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

package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestPrepareDoltgresServerArgsLogLevel(t *testing.T) {
	for _, level := range []string{"", "trace", "warn"} {
		t.Run("log_level="+level, func(t *testing.T) {
			cwd := t.TempDir()
			var args []string
			if level != "" {
				require.NoError(t, os.WriteFile(filepath.Join(cwd, "server.yaml"), []byte("log_level: "+level+"\n"), 0600))
				args = []string{"--config", "server.yaml"}
			}
			prepareDoltgresServerArgs(t, cwd, "logging", 5433, args)
			contents, err := os.ReadFile(filepath.Join(cwd, ".generated-logging-config.yaml"))
			require.NoError(t, err)
			var cfg map[string]any
			require.NoError(t, yaml.Unmarshal(contents, &cfg))
			require.Equal(t, "warn", cfg["log_level"])
		})
	}
}
