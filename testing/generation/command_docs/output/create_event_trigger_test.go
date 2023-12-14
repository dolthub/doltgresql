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

func TestCreateEventTrigger(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("CREATE EVENT TRIGGER name ON event EXECUTE FUNCTION function_name ( )"),
		Unimplemented("CREATE EVENT TRIGGER name ON event WHEN filter_variable IN ( 'Active' ) EXECUTE FUNCTION function_name ( )"),
		Unimplemented("CREATE EVENT TRIGGER name ON event WHEN filter_variable IN ( 'Active' , 'Active' ) EXECUTE FUNCTION function_name ( )"),
		Unimplemented("CREATE EVENT TRIGGER name ON event WHEN filter_variable IN ( 'Active' ) AND filter_variable IN ( 'Active' ) EXECUTE FUNCTION function_name ( )"),
		Unimplemented("CREATE EVENT TRIGGER name ON event WHEN filter_variable IN ( 'Active' , 'Active' ) AND filter_variable IN ( 'Active' ) EXECUTE FUNCTION function_name ( )"),
		Unimplemented("CREATE EVENT TRIGGER name ON event WHEN filter_variable IN ( 'Active' ) AND filter_variable IN ( 'Active' , 'Active' ) EXECUTE FUNCTION function_name ( )"),
		Unimplemented("CREATE EVENT TRIGGER name ON event WHEN filter_variable IN ( 'Active' , 'Active' ) AND filter_variable IN ( 'Active' , 'Active' ) EXECUTE FUNCTION function_name ( )"),
		Unimplemented("CREATE EVENT TRIGGER name ON event EXECUTE PROCEDURE function_name ( )"),
		Unimplemented("CREATE EVENT TRIGGER name ON event WHEN filter_variable IN ( 'Active' ) EXECUTE PROCEDURE function_name ( )"),
		Unimplemented("CREATE EVENT TRIGGER name ON event WHEN filter_variable IN ( 'Active' , 'Active' ) EXECUTE PROCEDURE function_name ( )"),
		Unimplemented("CREATE EVENT TRIGGER name ON event WHEN filter_variable IN ( 'Active' ) AND filter_variable IN ( 'Active' ) EXECUTE PROCEDURE function_name ( )"),
		Unimplemented("CREATE EVENT TRIGGER name ON event WHEN filter_variable IN ( 'Active' , 'Active' ) AND filter_variable IN ( 'Active' ) EXECUTE PROCEDURE function_name ( )"),
		Unimplemented("CREATE EVENT TRIGGER name ON event WHEN filter_variable IN ( 'Active' ) AND filter_variable IN ( 'Active' , 'Active' ) EXECUTE PROCEDURE function_name ( )"),
		Unimplemented("CREATE EVENT TRIGGER name ON event WHEN filter_variable IN ( 'Active' , 'Active' ) AND filter_variable IN ( 'Active' , 'Active' ) EXECUTE PROCEDURE function_name ( )"),
	}
	RunTests(t, tests)
}
