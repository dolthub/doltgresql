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

import "testing"

// TestSQLServerInfoFile is ported from Dolt's sqlserver_info_test.go. It
// asserts the behavior of the .dolt/sql-server.info lock file by racing the
// `dolt sql` and `dolt sql-server` CLI subcommands against a running server.
// The doltgres binary only runs the server and provides no `sql` CLI
// subcommand, so this cannot be reproduced and is skipped.
//
// The full test logic lives in the Dolt source at
// integration-tests/go-sql-server-driver/sqlserver_info_test.go.
func TestSQLServerInfoFile(t *testing.T) {
	t.Skip("depends on the `dolt sql` / `dolt sql-server` CLI subcommands, which the doltgres binary does not provide")
}
