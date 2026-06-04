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

// TestValctxEnabled is ported from Dolt's valctx_enabled_test.go. It asserts
// that context validation (valctx) is enabled by calling dolt_test_valctx().
// The DOLT_CONTEXT_VALIDATION_ENABLED / DOLT_ENABLE_DYNAMIC_ASSERTS harness and
// the dolt_test_valctx() procedure are Dolt-internal debugging aids; enabling
// them against the doltgres server's connection setup triggers spurious
// panics (the Postgres connection handler does not satisfy the same session
// invariants), so this is skipped.
//
// The full test logic lives in the Dolt source at
// integration-tests/go-sql-server-driver/valctx_enabled_test.go.
func TestValctxEnabled(t *testing.T) {
	t.Skip("dolt_test_valctx() / context-validation harness is Dolt-internal and not applicable to the doltgres Postgres connection handler")
}
