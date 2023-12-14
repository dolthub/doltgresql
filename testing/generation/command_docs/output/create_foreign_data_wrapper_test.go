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

func TestCreateForeignDataWrapper(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("CREATE FOREIGN DATA WRAPPER name"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name HANDLER handler_function"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name NO HANDLER"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name VALIDATOR validator_function"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name NO VALIDATOR"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name HANDLER handler_function OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name NO HANDLER OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name VALIDATOR validator_function OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name HANDLER handler_function VALIDATOR validator_function OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name NO HANDLER VALIDATOR validator_function OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name NO VALIDATOR OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name HANDLER handler_function NO VALIDATOR OPTIONS ( option ' value ' , option ' value ' )"),
		Unimplemented("CREATE FOREIGN DATA WRAPPER name NO HANDLER NO VALIDATOR OPTIONS ( option ' value ' , option ' value ' )"),
	}
	RunTests(t, tests)
}
