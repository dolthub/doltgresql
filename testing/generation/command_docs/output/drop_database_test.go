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

func TestDropDatabase(t *testing.T) {
	tests := []QueryParses{
		Converts("DROP DATABASE name"),
		Converts("DROP DATABASE IF EXISTS name"),
		Unimplemented("DROP DATABASE name ( FORCE )"),
		Unimplemented("DROP DATABASE IF EXISTS name ( FORCE )"),
		Unimplemented("DROP DATABASE name WITH ( FORCE )"),
		Unimplemented("DROP DATABASE IF EXISTS name WITH ( FORCE )"),
		Unimplemented("DROP DATABASE name ( FORCE , FORCE )"),
		Unimplemented("DROP DATABASE IF EXISTS name ( FORCE , FORCE )"),
		Unimplemented("DROP DATABASE name WITH ( FORCE , FORCE )"),
		Unimplemented("DROP DATABASE IF EXISTS name WITH ( FORCE , FORCE )"),
	}
	RunTests(t, tests)
}
