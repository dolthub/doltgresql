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

func TestDropExtension(t *testing.T) {
	tests := []QueryParses{
		Converts("DROP EXTENSION name"),
		Converts("DROP EXTENSION IF EXISTS name"),
		Converts("DROP EXTENSION name , name"),
		Converts("DROP EXTENSION IF EXISTS name , name"),
		Converts("DROP EXTENSION name CASCADE"),
		Converts("DROP EXTENSION IF EXISTS name CASCADE"),
		Converts("DROP EXTENSION name , name CASCADE"),
		Converts("DROP EXTENSION IF EXISTS name , name CASCADE"),
		Converts("DROP EXTENSION name RESTRICT"),
		Converts("DROP EXTENSION IF EXISTS name RESTRICT"),
		Converts("DROP EXTENSION name , name RESTRICT"),
		Converts("DROP EXTENSION IF EXISTS name , name RESTRICT"),
	}
	RunTests(t, tests)
}
