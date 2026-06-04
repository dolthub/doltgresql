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

// TestReadOnlyDatabaseLoadSkipsLockTimeout is ported from Dolt's
// read_only_db_lock_timeout_test.go. It times `dolt sql -q "show databases"`
// run against a data dir while a server holds the database file locks,
// asserting the fast-fail lock optimization. The doltgres binary provides no
// `sql` CLI subcommand to run against the data dir, so this is skipped.
//
// The full test logic lives in the Dolt source at
// integration-tests/go-sql-server-driver/read_only_db_lock_timeout_test.go.
func TestReadOnlyDatabaseLoadSkipsLockTimeout(t *testing.T) {
	t.Skip("depends on the `dolt sql` CLI subcommand, which the doltgres binary does not provide")
}
