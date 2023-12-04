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

func TestAlterTextSearchConfiguration(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name ADD MAPPING FOR token_type WITH dictionary_name"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name ADD MAPPING FOR token_type , token_type WITH dictionary_name"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name ADD MAPPING FOR token_type WITH dictionary_name , dictionary_name"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name ADD MAPPING FOR token_type , token_type WITH dictionary_name , dictionary_name"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name ALTER MAPPING FOR token_type WITH dictionary_name"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name ALTER MAPPING FOR token_type , token_type WITH dictionary_name"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name ALTER MAPPING FOR token_type WITH dictionary_name , dictionary_name"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name ALTER MAPPING FOR token_type , token_type WITH dictionary_name , dictionary_name"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name ALTER MAPPING REPLACE old_dictionary WITH new_dictionary"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name ALTER MAPPING FOR token_type REPLACE old_dictionary WITH new_dictionary"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name ALTER MAPPING FOR token_type , token_type REPLACE old_dictionary WITH new_dictionary"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name DROP MAPPING FOR token_type"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name DROP MAPPING IF EXISTS FOR token_type"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name DROP MAPPING FOR token_type , token_type"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name DROP MAPPING IF EXISTS FOR token_type , token_type"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name RENAME TO new_name"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name OWNER TO new_owner"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name OWNER TO CURRENT_ROLE"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name OWNER TO CURRENT_USER"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name OWNER TO SESSION_USER"),
		Unimplemented("ALTER TEXT SEARCH CONFIGURATION name SET SCHEMA new_schema"),
	}
	RunTests(t, tests)
}
