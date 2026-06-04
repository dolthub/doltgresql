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

// TestEventsUndrop is ported from Dolt's events_undrop_test.go. It verifies
// that MySQL-style scheduled EVENTs continue to fire after a database is
// dropped and then undropped. Doltgres does not enable the event scheduler
// (servercfg EventSchedulerStatus() returns "OFF") and does not support the
// MySQL CREATE EVENT syntax, so this test is skipped.
//
// The full test logic lives in the Dolt source at
// integration-tests/go-sql-server-driver/events_undrop_test.go.
func TestEventsUndrop(t *testing.T) {
	t.Skip("MySQL event scheduler is not enabled in Doltgres (EventSchedulerStatus is OFF)")
}
