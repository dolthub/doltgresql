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

func TestAlterOperator(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) OWNER TO new_owner"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) OWNER TO new_owner"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) OWNER TO CURRENT_ROLE"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) OWNER TO CURRENT_ROLE"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) OWNER TO CURRENT_USER"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) OWNER TO CURRENT_USER"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) OWNER TO SESSION_USER"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) OWNER TO SESSION_USER"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET SCHEMA new_schema"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET SCHEMA new_schema"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( RESTRICT = res_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( RESTRICT = res_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( RESTRICT = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( RESTRICT = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( JOIN = join_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( JOIN = join_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( JOIN = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( JOIN = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( RESTRICT = res_proc , RESTRICT = res_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( RESTRICT = res_proc , RESTRICT = res_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( RESTRICT = NONE , RESTRICT = res_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( RESTRICT = NONE , RESTRICT = res_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( JOIN = join_proc , RESTRICT = res_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( JOIN = join_proc , RESTRICT = res_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( JOIN = NONE , RESTRICT = res_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( JOIN = NONE , RESTRICT = res_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( RESTRICT = res_proc , RESTRICT = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( RESTRICT = res_proc , RESTRICT = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( RESTRICT = NONE , RESTRICT = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( RESTRICT = NONE , RESTRICT = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( JOIN = join_proc , RESTRICT = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( JOIN = join_proc , RESTRICT = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( JOIN = NONE , RESTRICT = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( JOIN = NONE , RESTRICT = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( RESTRICT = res_proc , JOIN = join_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( RESTRICT = res_proc , JOIN = join_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( RESTRICT = NONE , JOIN = join_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( RESTRICT = NONE , JOIN = join_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( JOIN = join_proc , JOIN = join_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( JOIN = join_proc , JOIN = join_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( JOIN = NONE , JOIN = join_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( JOIN = NONE , JOIN = join_proc )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( RESTRICT = res_proc , JOIN = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( RESTRICT = res_proc , JOIN = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( RESTRICT = NONE , JOIN = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( RESTRICT = NONE , JOIN = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( JOIN = join_proc , JOIN = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( JOIN = join_proc , JOIN = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( left_type , right_type ) SET ( JOIN = NONE , JOIN = NONE )"),
		Unimplemented("ALTER OPERATOR @@ ( NONE , right_type ) SET ( JOIN = NONE , JOIN = NONE )"),
	}
	RunTests(t, tests)
}
