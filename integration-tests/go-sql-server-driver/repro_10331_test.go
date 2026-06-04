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
	"testing"
)

// TestRegression10331 tests for the checksum error caused by concurrent executions of `dolt sql`
// while dolt sql-server is processing writes.
//
// https://github.com/dolthub/dolt/issues/10331
func TestRegression10331(t *testing.T) {
	t.Parallel()
	t.Skip("depends on the `dolt sql` CLI subcommand running concurrently against the on-disk database; the doltgres binary only runs the server and has no equivalent CLI command, so this regression cannot be reproduced from the doltgres test driver")
}
