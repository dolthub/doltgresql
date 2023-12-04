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

func TestCreateTransform(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("CREATE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name , TO SQL WITH FUNCTION to_sql_function_name )"),
		Unimplemented("CREATE OR REPLACE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name , TO SQL WITH FUNCTION to_sql_function_name )"),
		Unimplemented("CREATE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name ( argument_type ) , TO SQL WITH FUNCTION to_sql_function_name )"),
		Unimplemented("CREATE OR REPLACE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name ( argument_type ) , TO SQL WITH FUNCTION to_sql_function_name )"),
		Unimplemented("CREATE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name ( argument_type , argument_type ) , TO SQL WITH FUNCTION to_sql_function_name )"),
		Unimplemented("CREATE OR REPLACE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name ( argument_type , argument_type ) , TO SQL WITH FUNCTION to_sql_function_name )"),
		Unimplemented("CREATE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name , TO SQL WITH FUNCTION to_sql_function_name ( argument_type ) )"),
		Unimplemented("CREATE OR REPLACE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name , TO SQL WITH FUNCTION to_sql_function_name ( argument_type ) )"),
		Unimplemented("CREATE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name ( argument_type ) , TO SQL WITH FUNCTION to_sql_function_name ( argument_type ) )"),
		Unimplemented("CREATE OR REPLACE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name ( argument_type ) , TO SQL WITH FUNCTION to_sql_function_name ( argument_type ) )"),
		Unimplemented("CREATE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name ( argument_type , argument_type ) , TO SQL WITH FUNCTION to_sql_function_name ( argument_type ) )"),
		Unimplemented("CREATE OR REPLACE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name ( argument_type , argument_type ) , TO SQL WITH FUNCTION to_sql_function_name ( argument_type ) )"),
		Unimplemented("CREATE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name , TO SQL WITH FUNCTION to_sql_function_name ( argument_type , argument_type ) )"),
		Unimplemented("CREATE OR REPLACE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name , TO SQL WITH FUNCTION to_sql_function_name ( argument_type , argument_type ) )"),
		Unimplemented("CREATE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name ( argument_type ) , TO SQL WITH FUNCTION to_sql_function_name ( argument_type , argument_type ) )"),
		Unimplemented("CREATE OR REPLACE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name ( argument_type ) , TO SQL WITH FUNCTION to_sql_function_name ( argument_type , argument_type ) )"),
		Unimplemented("CREATE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name ( argument_type , argument_type ) , TO SQL WITH FUNCTION to_sql_function_name ( argument_type , argument_type ) )"),
		Unimplemented("CREATE OR REPLACE TRANSFORM FOR type_name LANGUAGE lang_name ( FROM SQL WITH FUNCTION from_sql_function_name ( argument_type , argument_type ) , TO SQL WITH FUNCTION to_sql_function_name ( argument_type , argument_type ) )"),
	}
	RunTests(t, tests)
}
