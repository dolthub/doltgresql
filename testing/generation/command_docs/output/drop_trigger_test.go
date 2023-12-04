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

func TestDropTrigger(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("DROP TRIGGER name ON table_name"),
		Unimplemented("DROP TRIGGER IF EXISTS name ON table_name"),
		Unimplemented("DROP TRIGGER name ON table_name CASCADE"),
		Unimplemented("DROP TRIGGER IF EXISTS name ON table_name CASCADE"),
		Unimplemented("DROP TRIGGER name ON table_name RESTRICT"),
		Unimplemented("DROP TRIGGER IF EXISTS name ON table_name RESTRICT"),
	}
	RunTests(t, tests)
}
