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

func TestCreateUserMapping(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("CREATE USER MAPPING FOR user_name SERVER server_name"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR user_name SERVER server_name"),
		Unimplemented("CREATE USER MAPPING FOR USER SERVER server_name"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR USER SERVER server_name"),
		Unimplemented("CREATE USER MAPPING FOR CURRENT_ROLE SERVER server_name"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR CURRENT_ROLE SERVER server_name"),
		Unimplemented("CREATE USER MAPPING FOR CURRENT_USER SERVER server_name"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR CURRENT_USER SERVER server_name"),
		Unimplemented("CREATE USER MAPPING FOR PUBLIC SERVER server_name"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR PUBLIC SERVER server_name"),
		Unimplemented("CREATE USER MAPPING FOR user_name SERVER server_name OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR user_name SERVER server_name OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE USER MAPPING FOR USER SERVER server_name OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR USER SERVER server_name OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR CURRENT_ROLE SERVER server_name OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR CURRENT_USER SERVER server_name OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR PUBLIC SERVER server_name OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE USER MAPPING FOR user_name SERVER server_name OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR user_name SERVER server_name OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE USER MAPPING FOR USER SERVER server_name OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR USER SERVER server_name OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE USER MAPPING FOR CURRENT_ROLE SERVER server_name OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR CURRENT_ROLE SERVER server_name OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE USER MAPPING FOR CURRENT_USER SERVER server_name OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR CURRENT_USER SERVER server_name OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE USER MAPPING FOR PUBLIC SERVER server_name OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE USER MAPPING IF NOT EXISTS FOR PUBLIC SERVER server_name OPTIONS ( option ' value ' , option ' value ' )"),
	}
	RunTests(t, tests)
}
