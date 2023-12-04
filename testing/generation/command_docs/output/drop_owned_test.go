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

func TestDropOwned(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("DROP OWNED BY name"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE"),
		Unimplemented("DROP OWNED BY CURRENT_USER"),
		Unimplemented("DROP OWNED BY SESSION_USER"),
		Unimplemented("DROP OWNED BY name , name"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE , name"),
		Unimplemented("DROP OWNED BY CURRENT_USER , name"),
		Unimplemented("DROP OWNED BY SESSION_USER , name"),
		Unimplemented("DROP OWNED BY name , CURRENT_ROLE"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE , CURRENT_ROLE"),
		Unimplemented("DROP OWNED BY CURRENT_USER , CURRENT_ROLE"),
		Unimplemented("DROP OWNED BY SESSION_USER , CURRENT_ROLE"),
		Unimplemented("DROP OWNED BY name , CURRENT_USER"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE , CURRENT_USER"),
		Unimplemented("DROP OWNED BY CURRENT_USER , CURRENT_USER"),
		Unimplemented("DROP OWNED BY SESSION_USER , CURRENT_USER"),
		Unimplemented("DROP OWNED BY name , SESSION_USER"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE , SESSION_USER"),
		Unimplemented("DROP OWNED BY CURRENT_USER , SESSION_USER"),
		Unimplemented("DROP OWNED BY SESSION_USER , SESSION_USER"),
		Unimplemented("DROP OWNED BY name CASCADE"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE CASCADE"),
		Unimplemented("DROP OWNED BY CURRENT_USER CASCADE"),
		Unimplemented("DROP OWNED BY SESSION_USER CASCADE"),
		Unimplemented("DROP OWNED BY name , name CASCADE"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE , name CASCADE"),
		Unimplemented("DROP OWNED BY CURRENT_USER , name CASCADE"),
		Unimplemented("DROP OWNED BY SESSION_USER , name CASCADE"),
		Unimplemented("DROP OWNED BY name , CURRENT_ROLE CASCADE"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE , CURRENT_ROLE CASCADE"),
		Unimplemented("DROP OWNED BY CURRENT_USER , CURRENT_ROLE CASCADE"),
		Unimplemented("DROP OWNED BY SESSION_USER , CURRENT_ROLE CASCADE"),
		Unimplemented("DROP OWNED BY name , CURRENT_USER CASCADE"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE , CURRENT_USER CASCADE"),
		Unimplemented("DROP OWNED BY CURRENT_USER , CURRENT_USER CASCADE"),
		Unimplemented("DROP OWNED BY SESSION_USER , CURRENT_USER CASCADE"),
		Unimplemented("DROP OWNED BY name , SESSION_USER CASCADE"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE , SESSION_USER CASCADE"),
		Unimplemented("DROP OWNED BY CURRENT_USER , SESSION_USER CASCADE"),
		Unimplemented("DROP OWNED BY SESSION_USER , SESSION_USER CASCADE"),
		Unimplemented("DROP OWNED BY name RESTRICT"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE RESTRICT"),
		Unimplemented("DROP OWNED BY CURRENT_USER RESTRICT"),
		Unimplemented("DROP OWNED BY SESSION_USER RESTRICT"),
		Unimplemented("DROP OWNED BY name , name RESTRICT"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE , name RESTRICT"),
		Unimplemented("DROP OWNED BY CURRENT_USER , name RESTRICT"),
		Unimplemented("DROP OWNED BY SESSION_USER , name RESTRICT"),
		Unimplemented("DROP OWNED BY name , CURRENT_ROLE RESTRICT"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE , CURRENT_ROLE RESTRICT"),
		Unimplemented("DROP OWNED BY CURRENT_USER , CURRENT_ROLE RESTRICT"),
		Unimplemented("DROP OWNED BY SESSION_USER , CURRENT_ROLE RESTRICT"),
		Unimplemented("DROP OWNED BY name , CURRENT_USER RESTRICT"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE , CURRENT_USER RESTRICT"),
		Unimplemented("DROP OWNED BY CURRENT_USER , CURRENT_USER RESTRICT"),
		Unimplemented("DROP OWNED BY SESSION_USER , CURRENT_USER RESTRICT"),
		Unimplemented("DROP OWNED BY name , SESSION_USER RESTRICT"),
		Unimplemented("DROP OWNED BY CURRENT_ROLE , SESSION_USER RESTRICT"),
		Unimplemented("DROP OWNED BY CURRENT_USER , SESSION_USER RESTRICT"),
		Unimplemented("DROP OWNED BY SESSION_USER , SESSION_USER RESTRICT"),
	}
	RunTests(t, tests)
}
