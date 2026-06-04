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

// TestCloneNoConjoin is ported from Dolt's clone_no_conjoin_test.go. It clones
// a database through a running sql-server and asserts the clone does not
// trigger a conjoin of table files. This depends on Dolt's remotes/clone
// behavior and on inspecting on-disk table-file/conjoin storage internals,
// neither of which is available from the doltgres test driver yet (the
// remotes API is not implemented in Doltgres).
//
// The full test logic lives in the Dolt source at
// integration-tests/go-sql-server-driver/clone_no_conjoin_test.go.
func TestCloneNoConjoin(t *testing.T) {
	t.Skip("clone/conjoin storage internals and the remotes API are not yet supported in Doltgres")
}
