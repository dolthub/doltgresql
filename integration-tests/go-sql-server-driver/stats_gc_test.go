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

// TestStatsGCConcurrency and TestStatsAnalyzeTableSpeed are ported from Dolt's
// stats_gc_test.go. They exercise the automatic statistics collector
// concurrently with GC. Doltgres' automatic statistics + GC integration is not
// yet verified here, so these are skipped.
//
// The full test logic lives in the Dolt source at
// integration-tests/go-sql-server-driver/stats_gc_test.go.
func TestStatsGCConcurrency(t *testing.T) {
	t.Skip("Dolt statistics + GC integration not yet verified on Doltgres")
}

func TestStatsAnalyzeTableSpeed(t *testing.T) {
	t.Skip("Dolt statistics + GC integration not yet verified on Doltgres")
}
