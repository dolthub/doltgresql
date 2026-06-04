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

// TestGCConjoinsOldgen is ported from Dolt's gc_oldgen_conjoin_test.go. It
// asserts that GC conjoins oldgen table files, which requires inspecting
// on-disk storage internals not accessible from the doltgres test driver.
//
// The full test logic lives in the Dolt source at
// integration-tests/go-sql-server-driver/gc_oldgen_conjoin_test.go.
func TestGCConjoinsOldgen(t *testing.T) {
	t.Skip("depends on Dolt oldgen conjoin storage internals not accessible from the doltgres test driver")
}
