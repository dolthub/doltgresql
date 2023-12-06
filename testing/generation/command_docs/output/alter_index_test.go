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

func TestAlterIndex(t *testing.T) {
	tests := []QueryParses{
		Parses("ALTER INDEX name RENAME TO new_name"),
		Parses("ALTER INDEX IF EXISTS name RENAME TO new_name"),
		Unimplemented("ALTER INDEX name SET TABLESPACE tablespace_name"),
		Unimplemented("ALTER INDEX IF EXISTS name SET TABLESPACE tablespace_name"),
		Unimplemented("ALTER INDEX name ATTACH PARTITION index_name"),
		Unimplemented("ALTER INDEX name DEPENDS ON EXTENSION extension_name"),
		Unimplemented("ALTER INDEX name NO DEPENDS ON EXTENSION extension_name"),
		Unimplemented("ALTER INDEX name SET ( fillfactor )"),
		Unimplemented("ALTER INDEX IF EXISTS name SET ( fillfactor )"),
		Unimplemented("ALTER INDEX name SET ( fillfactor = value )"),
		Unimplemented("ALTER INDEX IF EXISTS name SET ( fillfactor = value )"),
		Unimplemented("ALTER INDEX name SET ( fillfactor , fillfactor )"),
		Unimplemented("ALTER INDEX IF EXISTS name SET ( fillfactor , fillfactor )"),
		Unimplemented("ALTER INDEX name SET ( fillfactor = value , fillfactor )"),
		Unimplemented("ALTER INDEX IF EXISTS name SET ( fillfactor = value , fillfactor )"),
		Unimplemented("ALTER INDEX name SET ( fillfactor , fillfactor = value )"),
		Unimplemented("ALTER INDEX IF EXISTS name SET ( fillfactor , fillfactor = value )"),
		Unimplemented("ALTER INDEX name SET ( fillfactor = value , fillfactor = value )"),
		Unimplemented("ALTER INDEX IF EXISTS name SET ( fillfactor = value , fillfactor = value )"),
		Unimplemented("ALTER INDEX name RESET ( fillfactor )"),
		Unimplemented("ALTER INDEX IF EXISTS name RESET ( fillfactor )"),
		Unimplemented("ALTER INDEX name RESET ( fillfactor , fillfactor )"),
		Unimplemented("ALTER INDEX IF EXISTS name RESET ( fillfactor , fillfactor )"),
		Unimplemented("ALTER INDEX name ALTER column_number SET STATISTICS 1"),
		Unimplemented("ALTER INDEX IF EXISTS name ALTER column_number SET STATISTICS 1"),
		Unimplemented("ALTER INDEX name ALTER COLUMN column_number SET STATISTICS 1"),
		Unimplemented("ALTER INDEX IF EXISTS name ALTER COLUMN column_number SET STATISTICS 1"),
		Unimplemented("ALTER INDEX ALL IN TABLESPACE name SET TABLESPACE new_tablespace"),
		Unimplemented("ALTER INDEX ALL IN TABLESPACE name OWNED BY role_name SET TABLESPACE new_tablespace"),
		Unimplemented("ALTER INDEX ALL IN TABLESPACE name OWNED BY role_name , role_name SET TABLESPACE new_tablespace"),
		Unimplemented("ALTER INDEX ALL IN TABLESPACE name SET TABLESPACE new_tablespace NOWAIT"),
		Unimplemented("ALTER INDEX ALL IN TABLESPACE name OWNED BY role_name SET TABLESPACE new_tablespace NOWAIT"),
		Unimplemented("ALTER INDEX ALL IN TABLESPACE name OWNED BY role_name , role_name SET TABLESPACE new_tablespace NOWAIT"),
	}
	RunTests(t, tests)
}
