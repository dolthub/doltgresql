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

func TestReassignOwned(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("REASSIGN OWNED BY old_role TO new_role"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE TO new_role"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER TO new_role"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER TO new_role"),
		Unimplemented("REASSIGN OWNED BY old_role , old_role TO new_role"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , old_role TO new_role"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , old_role TO new_role"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , old_role TO new_role"),
		Unimplemented("REASSIGN OWNED BY old_role , CURRENT_ROLE TO new_role"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , CURRENT_ROLE TO new_role"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , CURRENT_ROLE TO new_role"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , CURRENT_ROLE TO new_role"),
		Unimplemented("REASSIGN OWNED BY old_role , CURRENT_USER TO new_role"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , CURRENT_USER TO new_role"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , CURRENT_USER TO new_role"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , CURRENT_USER TO new_role"),
		Unimplemented("REASSIGN OWNED BY old_role , SESSION_USER TO new_role"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , SESSION_USER TO new_role"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , SESSION_USER TO new_role"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , SESSION_USER TO new_role"),
		Unimplemented("REASSIGN OWNED BY old_role TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY old_role , old_role TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , old_role TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , old_role TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , old_role TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY old_role , CURRENT_ROLE TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , CURRENT_ROLE TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , CURRENT_ROLE TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , CURRENT_ROLE TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY old_role , CURRENT_USER TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , CURRENT_USER TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , CURRENT_USER TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , CURRENT_USER TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY old_role , SESSION_USER TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , SESSION_USER TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , SESSION_USER TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , SESSION_USER TO CURRENT_ROLE"),
		Unimplemented("REASSIGN OWNED BY old_role TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY old_role , old_role TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , old_role TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , old_role TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , old_role TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY old_role , CURRENT_ROLE TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , CURRENT_ROLE TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , CURRENT_ROLE TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , CURRENT_ROLE TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY old_role , CURRENT_USER TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , CURRENT_USER TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , CURRENT_USER TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , CURRENT_USER TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY old_role , SESSION_USER TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , SESSION_USER TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , SESSION_USER TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , SESSION_USER TO CURRENT_USER"),
		Unimplemented("REASSIGN OWNED BY old_role TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY old_role , old_role TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , old_role TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , old_role TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , old_role TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY old_role , CURRENT_ROLE TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , CURRENT_ROLE TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , CURRENT_ROLE TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , CURRENT_ROLE TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY old_role , CURRENT_USER TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , CURRENT_USER TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , CURRENT_USER TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , CURRENT_USER TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY old_role , SESSION_USER TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_ROLE , SESSION_USER TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY CURRENT_USER , SESSION_USER TO SESSION_USER"),
		Unimplemented("REASSIGN OWNED BY SESSION_USER , SESSION_USER TO SESSION_USER"),
	}
	RunTests(t, tests)
}
