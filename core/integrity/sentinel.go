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

package integrity

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dolthub/dolt/go/libraries/utils/filesys"
)

// SentinelFileName is the name of the file dropped in a database's .dolt directory once the database
// has passed the startup integrity check, so that subsequent startups can skip re-checking it
const SentinelFileName = ".integrity_check_passed"

// sentinelCheckVersion identifies the semantics of the integrity check that wrote a sentinel. Bump it
// whenever the check learns to detect a new kind of corruption, so that sentinels recorded by older,
// weaker versions of the check don't suppress the new one.
const sentinelCheckVersion = 1

// sentinelPath returns the path of the sentinel file for the database whose .dolt directory is
// |doltDir|.
func sentinelPath(doltDir string) string {
	return filepath.Join(doltDir, SentinelFileName)
}

// HasValidSentinel reports whether the database whose .dolt directory is |doltDir| has already passed
// the current version of the integrity check.
func HasValidSentinel(fs filesys.Filesys, doltDir string) bool {
	if doltDir == "" {
		return false
	}
	data, err := fs.ReadFile(sentinelPath(doltDir))
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == fmt.Sprintf("%d", sentinelCheckVersion)
}

// WriteSentinel records that the database whose .dolt directory is |doltDir| passed the current
// version of the integrity check.
func WriteSentinel(fs filesys.Filesys, doltDir string) error {
	if doltDir == "" {
		return nil
	}
	return fs.WriteFile(sentinelPath(doltDir), []byte(fmt.Sprintf("%d\n", sentinelCheckVersion)), 0644)
}
