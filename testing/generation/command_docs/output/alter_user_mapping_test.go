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

func TestAlterUserMapping(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("ALTER USER MAPPING FOR user_name SERVER server_name OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR USER SERVER server_name OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR SESSION_USER SERVER server_name OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR user_name SERVER server_name OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR USER SERVER server_name OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR SESSION_USER SERVER server_name OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR user_name SERVER server_name OPTIONS ( DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR USER SERVER server_name OPTIONS ( DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR SESSION_USER SERVER server_name OPTIONS ( DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR user_name SERVER server_name OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR USER SERVER server_name OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR SESSION_USER SERVER server_name OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR user_name SERVER server_name OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR USER SERVER server_name OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR SESSION_USER SERVER server_name OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR user_name SERVER server_name OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR USER SERVER server_name OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR SESSION_USER SERVER server_name OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR user_name SERVER server_name OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR USER SERVER server_name OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR SESSION_USER SERVER server_name OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR user_name SERVER server_name OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR USER SERVER server_name OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR SESSION_USER SERVER server_name OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR user_name SERVER server_name OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR USER SERVER server_name OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR SESSION_USER SERVER server_name OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER USER MAPPING FOR user_name SERVER server_name OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR USER SERVER server_name OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR SESSION_USER SERVER server_name OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR user_name SERVER server_name OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR USER SERVER server_name OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR SESSION_USER SERVER server_name OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR user_name SERVER server_name OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR USER SERVER server_name OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR SESSION_USER SERVER server_name OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( DROP option , DROP option )"),
	}
	RunTests(t, tests)
}
