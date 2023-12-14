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

func TestMove(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("MOVE cursor_name"),
		Unimplemented("MOVE NEXT PRIOR cursor_name"),
		Unimplemented("MOVE FIRST cursor_name"),
		Unimplemented("MOVE LAST cursor_name"),
		Unimplemented("MOVE ABSOLUTE count cursor_name"),
		Unimplemented("MOVE RELATIVE count cursor_name"),
		Unimplemented("MOVE count cursor_name"),
		Unimplemented("MOVE ALL cursor_name"),
		Unimplemented("MOVE FORWARD cursor_name"),
		Unimplemented("MOVE FORWARD count cursor_name"),
		Unimplemented("MOVE FORWARD ALL cursor_name"),
		Unimplemented("MOVE BACKWARD cursor_name"),
		Unimplemented("MOVE BACKWARD count cursor_name"),
		Unimplemented("MOVE BACKWARD ALL cursor_name"),
		Unimplemented("MOVE FROM cursor_name"),
		Unimplemented("MOVE NEXT PRIOR FROM cursor_name"),
		Unimplemented("MOVE FIRST FROM cursor_name"),
		Unimplemented("MOVE LAST FROM cursor_name"),
		Unimplemented("MOVE ABSOLUTE count FROM cursor_name"),
		Unimplemented("MOVE RELATIVE count FROM cursor_name"),
		Unimplemented("MOVE count FROM cursor_name"),
		Unimplemented("MOVE ALL FROM cursor_name"),
		Unimplemented("MOVE FORWARD FROM cursor_name"),
		Unimplemented("MOVE FORWARD count FROM cursor_name"),
		Unimplemented("MOVE FORWARD ALL FROM cursor_name"),
		Unimplemented("MOVE BACKWARD FROM cursor_name"),
		Unimplemented("MOVE BACKWARD count FROM cursor_name"),
		Unimplemented("MOVE BACKWARD ALL FROM cursor_name"),
		Unimplemented("MOVE IN cursor_name"),
		Unimplemented("MOVE NEXT PRIOR IN cursor_name"),
		Unimplemented("MOVE FIRST IN cursor_name"),
		Unimplemented("MOVE LAST IN cursor_name"),
		Unimplemented("MOVE ABSOLUTE count IN cursor_name"),
		Unimplemented("MOVE RELATIVE count IN cursor_name"),
		Unimplemented("MOVE count IN cursor_name"),
		Unimplemented("MOVE ALL IN cursor_name"),
		Unimplemented("MOVE FORWARD IN cursor_name"),
		Unimplemented("MOVE FORWARD count IN cursor_name"),
		Unimplemented("MOVE FORWARD ALL IN cursor_name"),
		Unimplemented("MOVE BACKWARD IN cursor_name"),
		Unimplemented("MOVE BACKWARD count IN cursor_name"),
		Unimplemented("MOVE BACKWARD ALL IN cursor_name"),
	}
	RunTests(t, tests)
}
