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

func TestAlterServer(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("ALTER SERVER name VERSION ' new_version '"),
		Unimplemented("ALTER SERVER name VERSION ' new_version ' OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER SERVER name VERSION ' new_version ' OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER SERVER name VERSION ' new_version ' OPTIONS ( DROP option )"),
		Unimplemented("ALTER SERVER name VERSION ' new_version ' OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER SERVER name VERSION ' new_version ' OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER SERVER name VERSION ' new_version ' OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER SERVER name VERSION ' new_version ' OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER SERVER name VERSION ' new_version ' OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER SERVER name VERSION ' new_version ' OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER SERVER name VERSION ' new_version ' OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER SERVER name VERSION ' new_version ' OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER SERVER name VERSION ' new_version ' OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER SERVER name OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER SERVER name OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER SERVER name OPTIONS ( DROP option )"),
		Unimplemented("ALTER SERVER name OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER SERVER name OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER SERVER name OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER SERVER name OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER SERVER name OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER SERVER name OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER SERVER name OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER SERVER name OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER SERVER name OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER SERVER name OWNER TO new_owner"),
		Unimplemented("ALTER SERVER name OWNER TO CURRENT_ROLE"),
		Unimplemented("ALTER SERVER name OWNER TO CURRENT_USER"),
		Unimplemented("ALTER SERVER name OWNER TO SESSION_USER"),
		Unimplemented("ALTER SERVER name RENAME TO new_name"),
	}

	//RunTests(t, tests)
	RewriteTests(t, tests, "alter_server_test.go")
}
