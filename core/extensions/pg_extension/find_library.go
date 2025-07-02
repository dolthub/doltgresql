// Copyright 2025 Dolthub, Inc.
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

package pg_extension

import (
	"bytes"
	"os/exec"
	"strings"
)

// PostgresDirectories returns the installation directories of a local Postgres instance.
func PostgresDirectories() (libDir string, extensionDir string, err error) {
	var buffer bytes.Buffer
	cmd := exec.Command("pg_config", "--pkglibdir")
	cmd.Stdout = &buffer
	if err := cmd.Run(); err != nil {
		return "", "", err
	}
	libDir = strings.TrimSpace(buffer.String())
	buffer.Reset()
	cmd = exec.Command("pg_config", "--sharedir")
	cmd.Stdout = &buffer
	if err := cmd.Run(); err != nil {
		return "", "", err
	}
	extensionDir = strings.TrimSpace(buffer.String()) + "/extension"
	return libDir, extensionDir, nil
}
