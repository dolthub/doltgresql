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

func TestFetch(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("FETCH cursor_name"),
		Unimplemented("FETCH NEXT cursor_name"),
		Unimplemented("FETCH PRIOR cursor_name"),
		Unimplemented("FETCH FIRST cursor_name"),
		Unimplemented("FETCH LAST cursor_name"),
		Unimplemented("FETCH ABSOLUTE count cursor_name"),
		Unimplemented("FETCH RELATIVE count cursor_name"),
		Unimplemented("FETCH count cursor_name"),
		Unimplemented("FETCH ALL cursor_name"),
		Unimplemented("FETCH FORWARD cursor_name"),
		Unimplemented("FETCH FORWARD count cursor_name"),
		Unimplemented("FETCH FORWARD ALL cursor_name"),
		Unimplemented("FETCH BACKWARD cursor_name"),
		Unimplemented("FETCH BACKWARD count cursor_name"),
		Unimplemented("FETCH BACKWARD ALL cursor_name"),
		Unimplemented("FETCH FROM cursor_name"),
		Unimplemented("FETCH NEXT FROM cursor_name"),
		Unimplemented("FETCH PRIOR FROM cursor_name"),
		Unimplemented("FETCH FIRST FROM cursor_name"),
		Unimplemented("FETCH LAST FROM cursor_name"),
		Unimplemented("FETCH ABSOLUTE count FROM cursor_name"),
		Unimplemented("FETCH RELATIVE count FROM cursor_name"),
		Unimplemented("FETCH count FROM cursor_name"),
		Unimplemented("FETCH ALL FROM cursor_name"),
		Unimplemented("FETCH FORWARD FROM cursor_name"),
		Unimplemented("FETCH FORWARD count FROM cursor_name"),
		Unimplemented("FETCH FORWARD ALL FROM cursor_name"),
		Unimplemented("FETCH BACKWARD FROM cursor_name"),
		Unimplemented("FETCH BACKWARD count FROM cursor_name"),
		Unimplemented("FETCH BACKWARD ALL FROM cursor_name"),
		Unimplemented("FETCH IN cursor_name"),
		Unimplemented("FETCH NEXT IN cursor_name"),
		Unimplemented("FETCH PRIOR IN cursor_name"),
		Unimplemented("FETCH FIRST IN cursor_name"),
		Unimplemented("FETCH LAST IN cursor_name"),
		Unimplemented("FETCH ABSOLUTE count IN cursor_name"),
		Unimplemented("FETCH RELATIVE count IN cursor_name"),
		Unimplemented("FETCH count IN cursor_name"),
		Unimplemented("FETCH ALL IN cursor_name"),
		Unimplemented("FETCH FORWARD IN cursor_name"),
		Unimplemented("FETCH FORWARD count IN cursor_name"),
		Unimplemented("FETCH FORWARD ALL IN cursor_name"),
		Unimplemented("FETCH BACKWARD IN cursor_name"),
		Unimplemented("FETCH BACKWARD count IN cursor_name"),
		Unimplemented("FETCH BACKWARD ALL IN cursor_name"),
	}
	RunTests(t, tests)
}
