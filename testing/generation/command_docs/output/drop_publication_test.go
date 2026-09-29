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

func TestDropPublication(t *testing.T) {
	tests := []QueryParses{
		Converts("DROP PUBLICATION name"),
		Converts("DROP PUBLICATION IF EXISTS name"),
		Converts("DROP PUBLICATION name , name"),
		Converts("DROP PUBLICATION IF EXISTS name , name"),
		Converts("DROP PUBLICATION name CASCADE"),
		Converts("DROP PUBLICATION IF EXISTS name CASCADE"),
		Converts("DROP PUBLICATION name , name CASCADE"),
		Converts("DROP PUBLICATION IF EXISTS name , name CASCADE"),
		Converts("DROP PUBLICATION name RESTRICT"),
		Converts("DROP PUBLICATION IF EXISTS name RESTRICT"),
		Converts("DROP PUBLICATION name , name RESTRICT"),
		Converts("DROP PUBLICATION IF EXISTS name , name RESTRICT"),
	}
	RunTests(t, tests)
}
