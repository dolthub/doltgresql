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

func TestCreateTablespace(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("CREATE TABLESPACE tablespace_name LOCATION ' directory '"),
		Unimplemented("CREATE TABLESPACE tablespace_name OWNER new_owner LOCATION ' directory '"),
		Unimplemented("CREATE TABLESPACE tablespace_name OWNER CURRENT_ROLE LOCATION ' directory '"),
		Unimplemented("CREATE TABLESPACE tablespace_name OWNER CURRENT_USER LOCATION ' directory '"),
		Unimplemented("CREATE TABLESPACE tablespace_name OWNER SESSION_USER LOCATION ' directory '"),
		Unimplemented("CREATE TABLESPACE tablespace_name LOCATION ' directory ' WITH ( tablespace_option = value )"),
		Unimplemented("CREATE TABLESPACE tablespace_name OWNER new_owner LOCATION ' directory ' WITH ( tablespace_option = value )"),
		Unimplemented("CREATE TABLESPACE tablespace_name OWNER CURRENT_ROLE LOCATION ' directory ' WITH ( tablespace_option = value )"),
		Unimplemented("CREATE TABLESPACE tablespace_name OWNER CURRENT_USER LOCATION ' directory ' WITH ( tablespace_option = value )"),
		Unimplemented("CREATE TABLESPACE tablespace_name OWNER SESSION_USER LOCATION ' directory ' WITH ( tablespace_option = value )"),
		Unimplemented("CREATE TABLESPACE tablespace_name LOCATION ' directory ' WITH ( tablespace_option = value , value )"),
		Unimplemented("CREATE TABLESPACE tablespace_name OWNER new_owner LOCATION ' directory ' WITH ( tablespace_option = value , value )"),
		Unimplemented("CREATE TABLESPACE tablespace_name OWNER CURRENT_ROLE LOCATION ' directory ' WITH ( tablespace_option = value , value )"),
		Unimplemented("CREATE TABLESPACE tablespace_name OWNER CURRENT_USER LOCATION ' directory ' WITH ( tablespace_option = value , value )"),
		Unimplemented("CREATE TABLESPACE tablespace_name OWNER SESSION_USER LOCATION ' directory ' WITH ( tablespace_option = value , value )"),
	}
	RunTests(t, tests)
}
