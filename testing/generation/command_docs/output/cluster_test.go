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

func TestCluster(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("CLUSTER table_name"),
		Unimplemented("CLUSTER VERBOSE table_name"),
		Unimplemented("CLUSTER table_name USING index_name"),
		Unimplemented("CLUSTER VERBOSE table_name USING index_name"),
		Unimplemented("CLUSTER ( VERBOSE ) table_name"),
		Unimplemented("CLUSTER ( VERBOSE true ) table_name"),
		Unimplemented("CLUSTER ( VERBOSE , VERBOSE ) table_name"),
		Unimplemented("CLUSTER ( VERBOSE true , VERBOSE ) table_name"),
		Unimplemented("CLUSTER ( VERBOSE , VERBOSE true ) table_name"),
		Unimplemented("CLUSTER ( VERBOSE true , VERBOSE true ) table_name"),
		Unimplemented("CLUSTER ( VERBOSE ) table_name USING index_name"),
		Unimplemented("CLUSTER ( VERBOSE true ) table_name USING index_name"),
		Unimplemented("CLUSTER ( VERBOSE , VERBOSE ) table_name USING index_name"),
		Unimplemented("CLUSTER ( VERBOSE true , VERBOSE ) table_name USING index_name"),
		Unimplemented("CLUSTER ( VERBOSE , VERBOSE true ) table_name USING index_name"),
		Unimplemented("CLUSTER ( VERBOSE true , VERBOSE true ) table_name USING index_name"),
		Unimplemented("CLUSTER"),
		Unimplemented("CLUSTER VERBOSE"),
	}
	RunTests(t, tests)
}
