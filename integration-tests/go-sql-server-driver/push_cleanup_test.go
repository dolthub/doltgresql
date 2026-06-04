// Copyright 2025 Dolthub, Inc.
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

import (
	"testing"
)

// A simple test to ensure that temptf is cleaned up when we push.
//
// This is inconvenient to do from somewhere like `bats` because
// the failure mode we are looking for is before process shutdown,
// i.e., when there is a long running server.
func TestPushTemptfCleanup(t *testing.T) {
	t.Parallel()
	t.Skip("depends on dolt_push to a remote and on inspecting the on-disk .dolt/temptf directory layout, which are Dolt storage internals not accessible from the doltgres test driver")
}
