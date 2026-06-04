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

// TestAutoGC is ported from Dolt's auto_gc_test.go. It exercises the server's
// automatic garbage-collection behavior under load. Doltgres currently
// disables automatic GC (servercfg DoltgresAutoGCBehavior.Enable() returns
// false and there is no dolt_auto_gc_enabled / auto_gc_behavior config), so
// this test is skipped until auto-GC is supported.
//
// The full test logic lives in the Dolt source at
// integration-tests/go-sql-server-driver/auto_gc_test.go and should be ported
// once Doltgres supports automatic GC.
func TestAutoGC(t *testing.T) {
	t.Skip("automatic GC is not yet implemented in Doltgres (AutoGCBehavior.Enable() == false)")
}
