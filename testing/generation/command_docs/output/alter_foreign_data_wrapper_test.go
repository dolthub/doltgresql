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

func TestAlterForeignDataWrapper(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name VALIDATOR validator_function"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO VALIDATOR"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OPTIONS ( ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OPTIONS ( SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OPTIONS ( DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OPTIONS ( ADD option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OPTIONS ( SET option ' value ' , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OPTIONS ( DROP option , ADD option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OPTIONS ( ADD option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OPTIONS ( SET option ' value ' , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OPTIONS ( DROP option , SET option ' value ' )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OPTIONS ( ADD option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OPTIONS ( SET option ' value ' , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OPTIONS ( DROP option , DROP option )"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OWNER TO new_owner"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OWNER TO CURRENT_ROLE"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OWNER TO CURRENT_USER"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name OWNER TO SESSION_USER"),
		Unimplemented("ALTER FOREIGN DATA WRAPPER name RENAME TO new_name"),
	}
	RunTests(t, tests)
}
