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

// These tests are ported from Dolt's large_values_test.go. They exercise
// replication of large out-of-band values, a diversity of column types, and
// very wide tables between a primary and a standby server. They depend on
// Dolt cluster replication, which is not yet implemented in Doltgres, so they
// are skipped. The YAML suites sql-server-large-values-*.yaml,
// sql-server-type-diversity.yaml, and sql-server-wide-*-table.yaml cover the
// same scenarios (also skipped) in the YAML test-definition format.
//
// The full test logic lives in the Dolt source at
// integration-tests/go-sql-server-driver/large_values_test.go.

func TestLargeOutOfBandValues(t *testing.T) {
	t.Skip("exercises cluster replication of large values; cluster replication not yet implemented in Doltgres")
}

func TestTypeDiversity(t *testing.T) {
	t.Skip("exercises cluster replication across a diversity of types; cluster replication not yet implemented in Doltgres")
}

func TestWideTable(t *testing.T) {
	t.Skip("exercises cluster replication of very wide tables; cluster replication not yet implemented in Doltgres")
}
