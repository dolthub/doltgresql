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

// TestNoTableFilesRead is ported from Dolt's no_table_files_read_test.go. It
// asserts that certain operations do not read table files, by inspecting the
// on-disk storage layout. The doltgres test driver cannot inspect Dolt
// table-file storage internals, so this is skipped.
//
// The full test logic lives in the Dolt source at
// integration-tests/go-sql-server-driver/no_table_files_read_test.go.
func TestNoTableFilesRead(t *testing.T) {
	t.Skip("depends on Dolt table-file storage internals not accessible from the doltgres test driver")
}
