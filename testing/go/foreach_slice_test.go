// Copyright 2026 Dolthub, Inc.
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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestForeachSlice(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "foreach array slices",
		SetUpScript: []string{
			`CREATE FUNCTION row_totals(a int[]) RETURNS bigint[] LANGUAGE plpgsql AS $$ DECLARE r int[]; totals bigint[] := ARRAY[]::bigint[]; BEGIN FOREACH r SLICE 1 IN ARRAY a LOOP totals := array_append(totals,(SELECT sum(v) FROM unnest(r) AS u(v))); END LOOP; RETURN totals; END $$;`,
			`CREATE FUNCTION planes(a int[]) RETURNS int[] LANGUAGE plpgsql AS $$ DECLARE r int[]; totals int[] := ARRAY[]::int[]; BEGIN FOREACH r SLICE 2 IN ARRAY a LOOP totals := array_append(totals,cardinality(r)); END LOOP; RETURN totals; END $$;`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:           "SELECT row_totals(NULL::int[]);",
				ExpectedErr:     "must not be null",
				ExpectedErrCode: "22004",
			},
			{
				Query:           "SELECT row_totals(ARRAY[]::int[]);",
				ExpectedErr:     "slice dimension (1) is out of the valid range 0..0",
				ExpectedErrCode: "2202E",
			},

			{
				Query:    "SELECT row_totals(ARRAY[[1,2,3],[4,5,6]]);",
				Expected: []sql.Row{{"{6,15}"}},
			},
			{
				Query:    "SELECT row_totals(ARRAY[[[1,2],[3,4]],[[5,6],[7,8]]]);",
				Expected: []sql.Row{{"{3,7,11,15}"}},
			},
			{
				Query:    "SELECT planes(ARRAY[[[1,2],[3,4]],[[5,6],[7,8]]]);",
				Expected: []sql.Row{{"{4,4}"}},
			},
			{
				Query:           "SELECT planes(ARRAY[1,2]);",
				ExpectedErr:     "slice dimension (2) is out of the valid range 0..1",
				ExpectedErrCode: "2202E",
			},
		},
	}})
}
