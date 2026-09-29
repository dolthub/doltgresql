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

// TestForeachSlice checks slice-loop creation, iteration, and invalid inputs.
func TestForeachSlice(t *testing.T) {
	RunScripts(t, []ScriptTest{{
		Name: "foreach array slices",
		SetUpScript: []string{
			`CREATE TABLE slice_inputs (id int PRIMARY KEY, a int[]);`,
			`INSERT INTO slice_inputs VALUES (1,ARRAY[[1,NULL],[3,4]]),(2,ARRAY[[5,6]]);`,
			`CREATE FUNCTION qa_slice(input int[]) RETURNS int LANGUAGE plpgsql AS $$ DECLARE row int[]; total int := 0; BEGIN FOREACH row SLICE 1 IN ARRAY input LOOP total := total + cardinality(row); END LOOP; RETURN total; END $$;`,
			`CREATE FUNCTION qa_slice_quoted(input int[]) RETURNS int LANGUAGE plpgsql AS $$ DECLARE "row" int[]; total int := 0; BEGIN FOREACH "row" SLICE 1 IN ARRAY input LOOP total := total + cardinality("row"); END LOOP; RETURN total; END $$;`,
			`CREATE FUNCTION row_totals(a int[]) RETURNS bigint[] LANGUAGE plpgsql AS $$ DECLARE r int[]; totals bigint[] := ARRAY[]::bigint[]; BEGIN FOREACH r SLICE 1 IN ARRAY a LOOP totals := array_append(totals,(SELECT sum(v) FROM unnest(r) AS u(v))); END LOOP; RETURN totals; END $$;`,
			`CREATE FUNCTION slice_control(a int[]) RETURNS int[] LANGUAGE plpgsql AS $$
                DECLARE r int[]; totals int[] := ARRAY[]::int[];
                BEGIN
                    FOREACH r SLICE 1 IN ARRAY a LOOP
                        a := ARRAY[[99]];
                        CONTINUE WHEN r[1] = 1;
                        totals := array_append(totals, r[1]);
                        EXIT WHEN r[1] = 5;
                    END LOOP;
                    RETURN totals;
                END $$;`,
			`CREATE FUNCTION scalar_slice(a int[]) RETURNS int LANGUAGE plpgsql AS $$ DECLARE r int; BEGIN FOREACH r SLICE 1 IN ARRAY a LOOP END LOOP; RETURN r; END $$;`,
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

			{
				Query:    "SELECT qa_slice(ARRAY[[1,2],[3,4]]),qa_slice_quoted(ARRAY[[1,2],[3,4]]);",
				Expected: []sql.Row{{4, 4}},
			},
			{
				Query:    "SELECT id,row_totals(a),planes(a) FROM slice_inputs ORDER BY id;",
				Expected: []sql.Row{{1, "{1,7}", "{4}"}, {2, "{11}", "{2}"}},
			},
			{
				Query:    "SELECT row_totals((SELECT a FROM slice_inputs WHERE id=1)),planes((SELECT a FROM slice_inputs WHERE id=2));",
				Expected: []sql.Row{{"{1,7}", "{2}"}},
			},
			{
				Query:    "SELECT row_totals(ARRAY[NULL,NULL]::int[]),planes(ARRAY[[NULL,NULL]]::int[]);",
				Expected: []sql.Row{{"{NULL}", "{2}"}},
			},

			{
				Query:    "SELECT slice_control(ARRAY[[1,2],[3,4],[5,6],[7,8]]);",
				Expected: []sql.Row{{"{3,5}"}},
			},
			{
				Query:           "SELECT scalar_slice(ARRAY[[1,2]]);",
				ExpectedErr:     "FOREACH ... SLICE loop variable must be of an array type",
				ExpectedErrCode: "42804",
			},
		},
	}})
}
