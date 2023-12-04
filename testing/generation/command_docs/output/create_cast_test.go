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

func TestCreateCast(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITH FUNCTION function_name"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITH FUNCTION function_name ( argument_type )"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITH FUNCTION function_name ( argument_type , argument_type )"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITH FUNCTION function_name AS ASSIGNMENT"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITH FUNCTION function_name ( argument_type ) AS ASSIGNMENT"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITH FUNCTION function_name ( argument_type , argument_type ) AS ASSIGNMENT"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITH FUNCTION function_name AS IMPLICIT"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITH FUNCTION function_name ( argument_type ) AS IMPLICIT"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITH FUNCTION function_name ( argument_type , argument_type ) AS IMPLICIT"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITHOUT FUNCTION"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITHOUT FUNCTION AS ASSIGNMENT"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITHOUT FUNCTION AS IMPLICIT"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITH INOUT"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITH INOUT AS ASSIGNMENT"),
		Unimplemented("CREATE CAST ( source_type AS target_type ) WITH INOUT AS IMPLICIT"),
	}
	RunTests(t, tests)
}
