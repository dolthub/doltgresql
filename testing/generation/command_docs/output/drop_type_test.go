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

func TestDropType(t *testing.T) {
	tests := []QueryParses{
		Converts("DROP TYPE name"),
		Converts("DROP TYPE IF EXISTS name"),
		Parses("DROP TYPE name , name"),
		Parses("DROP TYPE IF EXISTS name , name"),
		Converts("DROP TYPE name CASCADE"),
		Converts("DROP TYPE IF EXISTS name CASCADE"),
		Parses("DROP TYPE name , name CASCADE"),
		Parses("DROP TYPE IF EXISTS name , name CASCADE"),
		Converts("DROP TYPE name RESTRICT"),
		Converts("DROP TYPE IF EXISTS name RESTRICT"),
		Parses("DROP TYPE name , name RESTRICT"),
		Parses("DROP TYPE IF EXISTS name , name RESTRICT"),
	}
	RunTests(t, tests)
}
