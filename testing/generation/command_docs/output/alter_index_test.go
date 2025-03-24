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
		Parses("ALTER INDEX name SET TABLESPACE tablespace_name"),
		Parses("ALTER INDEX IF EXISTS name SET TABLESPACE tablespace_name"),
		Parses("ALTER INDEX name ATTACH PARTITION index_name"),
		Parses("ALTER INDEX name DEPENDS ON EXTENSION extension_name"),
		Parses("ALTER INDEX name NO DEPENDS ON EXTENSION extension_name"),
		Parses("ALTER INDEX name SET ( fillfactor )"),
		Parses("ALTER INDEX IF EXISTS name SET ( fillfactor )"),
		Parses("ALTER INDEX name SET ( fillfactor = value )"),
		Parses("ALTER INDEX IF EXISTS name SET ( fillfactor = value )"),
		Parses("ALTER INDEX name SET ( fillfactor , fillfactor )"),
		Parses("ALTER INDEX IF EXISTS name SET ( fillfactor , fillfactor )"),
		Parses("ALTER INDEX name SET ( fillfactor = value , fillfactor )"),
		Parses("ALTER INDEX IF EXISTS name SET ( fillfactor = value , fillfactor )"),
		Parses("ALTER INDEX name SET ( fillfactor , fillfactor = value )"),
		Parses("ALTER INDEX IF EXISTS name SET ( fillfactor , fillfactor = value )"),
		Parses("ALTER INDEX name SET ( fillfactor = value , fillfactor = value )"),
		Parses("ALTER INDEX IF EXISTS name SET ( fillfactor = value , fillfactor = value )"),
		Parses("ALTER INDEX name RESET ( fillfactor )"),
		Parses("ALTER INDEX IF EXISTS name RESET ( fillfactor )"),
		Parses("ALTER INDEX name RESET ( fillfactor , fillfactor )"),
		Parses("ALTER INDEX IF EXISTS name RESET ( fillfactor , fillfactor )"),
		Parses("ALTER INDEX name ALTER 1 SET STATISTICS 1"),
		Parses("ALTER INDEX IF EXISTS name ALTER 1 SET STATISTICS 1"),
		Parses("ALTER INDEX name ALTER COLUMN 1 SET STATISTICS 1"),
		Parses("ALTER INDEX IF EXISTS name ALTER COLUMN 1 SET STATISTICS 1"),
		Parses("ALTER INDEX ALL IN TABLESPACE name SET TABLESPACE new_tablespace"),
		Parses("ALTER INDEX ALL IN TABLESPACE name OWNED BY role_name SET TABLESPACE new_tablespace"),
		Parses("ALTER INDEX ALL IN TABLESPACE name OWNED BY role_name , role_name SET TABLESPACE new_tablespace"),
		Parses("ALTER INDEX ALL IN TABLESPACE name SET TABLESPACE new_tablespace NOWAIT"),
		Parses("ALTER INDEX ALL IN TABLESPACE name OWNED BY role_name SET TABLESPACE new_tablespace NOWAIT"),
		Parses("ALTER INDEX ALL IN TABLESPACE name OWNED BY role_name , role_name SET TABLESPACE new_tablespace NOWAIT"),
	}
	
	RunTests(t, tests)
	// RewriteTests(t, tests, "alter_index_test.go")
}
