// Copyright 2023 Dolthub, Inc.
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

package output

import "testing"

func TestRollback(t *testing.T) {
	tests := []QueryParses{
		Converts("ROLLBACK"),
		Converts("ROLLBACK WORK"),
		Converts("ROLLBACK TRANSACTION"),
		Converts("ROLLBACK AND CHAIN"),
		Converts("ROLLBACK WORK AND CHAIN"),
		Converts("ROLLBACK TRANSACTION AND CHAIN"),
		Converts("ROLLBACK AND NO CHAIN"),
		Converts("ROLLBACK WORK AND NO CHAIN"),
		Converts("ROLLBACK TRANSACTION AND NO CHAIN"),
	}
	RunTests(t, tests)
}
