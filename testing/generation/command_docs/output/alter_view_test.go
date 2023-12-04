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

func TestAlterView(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("ALTER VIEW name ALTER column_name SET DEFAULT expression"),
		Unimplemented("ALTER VIEW IF EXISTS name ALTER column_name SET DEFAULT expression"),
		Unimplemented("ALTER VIEW name ALTER COLUMN column_name SET DEFAULT expression"),
		Unimplemented("ALTER VIEW IF EXISTS name ALTER COLUMN column_name SET DEFAULT expression"),
		Unimplemented("ALTER VIEW name ALTER column_name DROP DEFAULT"),
		Unimplemented("ALTER VIEW IF EXISTS name ALTER column_name DROP DEFAULT"),
		Unimplemented("ALTER VIEW name ALTER COLUMN column_name DROP DEFAULT"),
		Unimplemented("ALTER VIEW IF EXISTS name ALTER COLUMN column_name DROP DEFAULT"),
		Unimplemented("ALTER VIEW name OWNER TO new_owner"),
		Unimplemented("ALTER VIEW IF EXISTS name OWNER TO new_owner"),
		Unimplemented("ALTER VIEW name OWNER TO CURRENT_ROLE"),
		Unimplemented("ALTER VIEW IF EXISTS name OWNER TO CURRENT_ROLE"),
		Unimplemented("ALTER VIEW name OWNER TO CURRENT_USER"),
		Unimplemented("ALTER VIEW IF EXISTS name OWNER TO CURRENT_USER"),
		Unimplemented("ALTER VIEW name OWNER TO SESSION_USER"),
		Unimplemented("ALTER VIEW IF EXISTS name OWNER TO SESSION_USER"),
		Unimplemented("ALTER VIEW name RENAME column_name TO new_column_name"),
		Unimplemented("ALTER VIEW IF EXISTS name RENAME column_name TO new_column_name"),
		Unimplemented("ALTER VIEW name RENAME COLUMN column_name TO new_column_name"),
		Unimplemented("ALTER VIEW IF EXISTS name RENAME COLUMN column_name TO new_column_name"),
		Parses("ALTER VIEW name RENAME TO new_name"),
		Parses("ALTER VIEW IF EXISTS name RENAME TO new_name"),
		Parses("ALTER VIEW name SET SCHEMA new_schema"),
		Parses("ALTER VIEW IF EXISTS name SET SCHEMA new_schema"),
		Unimplemented("ALTER VIEW name SET ( view_option_name )"),
		Unimplemented("ALTER VIEW IF EXISTS name SET ( view_option_name )"),
		Unimplemented("ALTER VIEW name SET ( view_option_name = view_option_value )"),
		Unimplemented("ALTER VIEW IF EXISTS name SET ( view_option_name = view_option_value )"),
		Unimplemented("ALTER VIEW name SET ( view_option_name , view_option_name )"),
		Unimplemented("ALTER VIEW IF EXISTS name SET ( view_option_name , view_option_name )"),
		Unimplemented("ALTER VIEW name SET ( view_option_name = view_option_value , view_option_name )"),
		Unimplemented("ALTER VIEW IF EXISTS name SET ( view_option_name = view_option_value , view_option_name )"),
		Unimplemented("ALTER VIEW name SET ( view_option_name , view_option_name = view_option_value )"),
		Unimplemented("ALTER VIEW IF EXISTS name SET ( view_option_name , view_option_name = view_option_value )"),
		Unimplemented("ALTER VIEW name SET ( view_option_name = view_option_value , view_option_name = view_option_value )"),
		Unimplemented("ALTER VIEW IF EXISTS name SET ( view_option_name = view_option_value , view_option_name = view_option_value )"),
		Unimplemented("ALTER VIEW name RESET ( view_option_name )"),
		Unimplemented("ALTER VIEW IF EXISTS name RESET ( view_option_name )"),
		Unimplemented("ALTER VIEW name RESET ( view_option_name , view_option_name )"),
		Unimplemented("ALTER VIEW IF EXISTS name RESET ( view_option_name , view_option_name )"),
	}
	RunTests(t, tests)
}
