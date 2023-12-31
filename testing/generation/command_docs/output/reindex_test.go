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

func TestReindex(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("REINDEX INDEX name"),
		Unimplemented("REINDEX ( CONCURRENTLY ) INDEX name"),
		Unimplemented("REINDEX ( CONCURRENTLY true ) INDEX name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace ) INDEX name"),
		Unimplemented("REINDEX ( VERBOSE ) INDEX name"),
		Unimplemented("REINDEX ( VERBOSE true ) INDEX name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY ) INDEX name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY ) INDEX name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY ) INDEX name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY ) INDEX name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY ) INDEX name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY true ) INDEX name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY true ) INDEX name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY true ) INDEX name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY true ) INDEX name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY true ) INDEX name"),
		Unimplemented("REINDEX ( CONCURRENTLY , TABLESPACE new_tablespace ) INDEX name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , TABLESPACE new_tablespace ) INDEX name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , TABLESPACE new_tablespace ) INDEX name"),
		Unimplemented("REINDEX ( VERBOSE , TABLESPACE new_tablespace ) INDEX name"),
		Unimplemented("REINDEX ( VERBOSE true , TABLESPACE new_tablespace ) INDEX name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE ) INDEX name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE ) INDEX name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE ) INDEX name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE ) INDEX name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE ) INDEX name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE true ) INDEX name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE true ) INDEX name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE true ) INDEX name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE true ) INDEX name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE true ) INDEX name"),
		Unimplemented("REINDEX TABLE name"),
		Unimplemented("REINDEX ( CONCURRENTLY ) TABLE name"),
		Unimplemented("REINDEX ( CONCURRENTLY true ) TABLE name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace ) TABLE name"),
		Unimplemented("REINDEX ( VERBOSE ) TABLE name"),
		Unimplemented("REINDEX ( VERBOSE true ) TABLE name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY ) TABLE name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY ) TABLE name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY ) TABLE name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY ) TABLE name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY ) TABLE name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY true ) TABLE name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY true ) TABLE name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY true ) TABLE name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY true ) TABLE name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY true ) TABLE name"),
		Unimplemented("REINDEX ( CONCURRENTLY , TABLESPACE new_tablespace ) TABLE name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , TABLESPACE new_tablespace ) TABLE name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , TABLESPACE new_tablespace ) TABLE name"),
		Unimplemented("REINDEX ( VERBOSE , TABLESPACE new_tablespace ) TABLE name"),
		Unimplemented("REINDEX ( VERBOSE true , TABLESPACE new_tablespace ) TABLE name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE ) TABLE name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE ) TABLE name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE ) TABLE name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE ) TABLE name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE ) TABLE name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE true ) TABLE name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE true ) TABLE name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE true ) TABLE name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE true ) TABLE name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE true ) TABLE name"),
		Unimplemented("REINDEX SCHEMA name"),
		Unimplemented("REINDEX ( CONCURRENTLY ) SCHEMA name"),
		Unimplemented("REINDEX ( CONCURRENTLY true ) SCHEMA name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace ) SCHEMA name"),
		Unimplemented("REINDEX ( VERBOSE ) SCHEMA name"),
		Unimplemented("REINDEX ( VERBOSE true ) SCHEMA name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY ) SCHEMA name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY ) SCHEMA name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY ) SCHEMA name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY ) SCHEMA name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY ) SCHEMA name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY true ) SCHEMA name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY true ) SCHEMA name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY true ) SCHEMA name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY true ) SCHEMA name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY true ) SCHEMA name"),
		Unimplemented("REINDEX ( CONCURRENTLY , TABLESPACE new_tablespace ) SCHEMA name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , TABLESPACE new_tablespace ) SCHEMA name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , TABLESPACE new_tablespace ) SCHEMA name"),
		Unimplemented("REINDEX ( VERBOSE , TABLESPACE new_tablespace ) SCHEMA name"),
		Unimplemented("REINDEX ( VERBOSE true , TABLESPACE new_tablespace ) SCHEMA name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE ) SCHEMA name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE ) SCHEMA name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE ) SCHEMA name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE ) SCHEMA name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE ) SCHEMA name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE true ) SCHEMA name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE true ) SCHEMA name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE true ) SCHEMA name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE true ) SCHEMA name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE true ) SCHEMA name"),
		Unimplemented("REINDEX DATABASE name"),
		Unimplemented("REINDEX ( CONCURRENTLY ) DATABASE name"),
		Unimplemented("REINDEX ( CONCURRENTLY true ) DATABASE name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace ) DATABASE name"),
		Unimplemented("REINDEX ( VERBOSE ) DATABASE name"),
		Unimplemented("REINDEX ( VERBOSE true ) DATABASE name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY ) DATABASE name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY ) DATABASE name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY ) DATABASE name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY ) DATABASE name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY ) DATABASE name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY true ) DATABASE name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY true ) DATABASE name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY true ) DATABASE name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY true ) DATABASE name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY true ) DATABASE name"),
		Unimplemented("REINDEX ( CONCURRENTLY , TABLESPACE new_tablespace ) DATABASE name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , TABLESPACE new_tablespace ) DATABASE name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , TABLESPACE new_tablespace ) DATABASE name"),
		Unimplemented("REINDEX ( VERBOSE , TABLESPACE new_tablespace ) DATABASE name"),
		Unimplemented("REINDEX ( VERBOSE true , TABLESPACE new_tablespace ) DATABASE name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE ) DATABASE name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE ) DATABASE name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE ) DATABASE name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE ) DATABASE name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE ) DATABASE name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE true ) DATABASE name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE true ) DATABASE name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE true ) DATABASE name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE true ) DATABASE name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE true ) DATABASE name"),
		Unimplemented("REINDEX SYSTEM name"),
		Unimplemented("REINDEX ( CONCURRENTLY ) SYSTEM name"),
		Unimplemented("REINDEX ( CONCURRENTLY true ) SYSTEM name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace ) SYSTEM name"),
		Unimplemented("REINDEX ( VERBOSE ) SYSTEM name"),
		Unimplemented("REINDEX ( VERBOSE true ) SYSTEM name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY ) SYSTEM name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY ) SYSTEM name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY ) SYSTEM name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY ) SYSTEM name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY ) SYSTEM name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY true ) SYSTEM name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY true ) SYSTEM name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY true ) SYSTEM name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY true ) SYSTEM name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY true ) SYSTEM name"),
		Unimplemented("REINDEX ( CONCURRENTLY , TABLESPACE new_tablespace ) SYSTEM name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , TABLESPACE new_tablespace ) SYSTEM name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , TABLESPACE new_tablespace ) SYSTEM name"),
		Unimplemented("REINDEX ( VERBOSE , TABLESPACE new_tablespace ) SYSTEM name"),
		Unimplemented("REINDEX ( VERBOSE true , TABLESPACE new_tablespace ) SYSTEM name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE ) SYSTEM name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE ) SYSTEM name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE ) SYSTEM name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE ) SYSTEM name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE ) SYSTEM name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE true ) SYSTEM name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE true ) SYSTEM name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE true ) SYSTEM name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE true ) SYSTEM name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE true ) SYSTEM name"),
		Unimplemented("REINDEX INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY true ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY true ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY true ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY true ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY true ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , TABLESPACE new_tablespace ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , TABLESPACE new_tablespace ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , TABLESPACE new_tablespace ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , TABLESPACE new_tablespace ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , TABLESPACE new_tablespace ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE true ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE true ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE true ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE true ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE true ) INDEX CONCURRENTLY name"),
		Unimplemented("REINDEX TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY true ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY true ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY true ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY true ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY true ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , TABLESPACE new_tablespace ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , TABLESPACE new_tablespace ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , TABLESPACE new_tablespace ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , TABLESPACE new_tablespace ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , TABLESPACE new_tablespace ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE true ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE true ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE true ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE true ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE true ) TABLE CONCURRENTLY name"),
		Unimplemented("REINDEX SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY true ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY true ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY true ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY true ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY true ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , TABLESPACE new_tablespace ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , TABLESPACE new_tablespace ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , TABLESPACE new_tablespace ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , TABLESPACE new_tablespace ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , TABLESPACE new_tablespace ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE true ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE true ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE true ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE true ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE true ) SCHEMA CONCURRENTLY name"),
		Unimplemented("REINDEX DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY true ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY true ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY true ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY true ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY true ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , TABLESPACE new_tablespace ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , TABLESPACE new_tablespace ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , TABLESPACE new_tablespace ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , TABLESPACE new_tablespace ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , TABLESPACE new_tablespace ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE true ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE true ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE true ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE true ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE true ) DATABASE CONCURRENTLY name"),
		Unimplemented("REINDEX SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , CONCURRENTLY true ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , CONCURRENTLY true ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , CONCURRENTLY true ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , CONCURRENTLY true ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , CONCURRENTLY true ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , TABLESPACE new_tablespace ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , TABLESPACE new_tablespace ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , TABLESPACE new_tablespace ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , TABLESPACE new_tablespace ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , TABLESPACE new_tablespace ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY , VERBOSE true ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( CONCURRENTLY true , VERBOSE true ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( TABLESPACE new_tablespace , VERBOSE true ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE , VERBOSE true ) SYSTEM CONCURRENTLY name"),
		Unimplemented("REINDEX ( VERBOSE true , VERBOSE true ) SYSTEM CONCURRENTLY name"),
	}
	RunTests(t, tests)
}
