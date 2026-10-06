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

use harness::oid::*;
use harness::pgx::Time;
use harness::plan::PlanFact;
use harness::script::Cell::{Any, Null, Text as T};
use harness::script::{A, BindVar, Column, Diagnostic, E, Expected, Flow, N, S, ScriptTest, ScriptTestAssertion, USER_DEFINED, run_scripts, run_scripts_repeated};
use harness::wire::{Datum, F, Field, Fields, PGX_STARTUP, Receive, Send, Step, W, WireTest, run_wire_tests};

#[test]
fn test_foreach_slice() {
    run_scripts(&[
        ScriptTest {
            name: "foreach array slices",
            set_up_script: &[
                "CREATE TABLE slice_inputs (id int PRIMARY KEY, a int[]);",
                "INSERT INTO slice_inputs VALUES (1,ARRAY[[1,NULL],[3,4]]),(2,ARRAY[[5,6]]);",
                "CREATE FUNCTION qa_slice(input int[]) RETURNS int LANGUAGE plpgsql AS $$ DECLARE row int[]; total int := 0; BEGIN FOREACH row SLICE 1 IN ARRAY input LOOP total := total + cardinality(row); END LOOP; RETURN total; END $$;",
                r#"CREATE FUNCTION qa_slice_quoted(input int[]) RETURNS int LANGUAGE plpgsql AS $$ DECLARE "row" int[]; total int := 0; BEGIN FOREACH "row" SLICE 1 IN ARRAY input LOOP total := total + cardinality("row"); END LOOP; RETURN total; END $$;"#,
                "CREATE FUNCTION row_totals(a int[]) RETURNS bigint[] LANGUAGE plpgsql AS $$ DECLARE r int[]; totals bigint[] := ARRAY[]::bigint[]; BEGIN FOREACH r SLICE 1 IN ARRAY a LOOP totals := array_append(totals,(SELECT sum(v) FROM unnest(r) AS u(v))); END LOOP; RETURN totals; END $$;",
                r#"CREATE FUNCTION slice_control(a int[]) RETURNS int[] LANGUAGE plpgsql AS $$
                DECLARE r int[]; totals int[] := ARRAY[]::int[];
                BEGIN
                    FOREACH r SLICE 1 IN ARRAY a LOOP
                        a := ARRAY[[99]];
                        CONTINUE WHEN r[1] = 1;
                        totals := array_append(totals, r[1]);
                        EXIT WHEN r[1] = 5;
                    END LOOP;
                    RETURN totals;
                END $$;"#,
                "CREATE FUNCTION scalar_slice(a int[]) RETURNS int LANGUAGE plpgsql AS $$ DECLARE r int; BEGIN FOREACH r SLICE 1 IN ARRAY a LOOP END LOOP; RETURN r; END $$;",
                "CREATE FUNCTION planes(a int[]) RETURNS int[] LANGUAGE plpgsql AS $$ DECLARE r int[]; totals int[] := ARRAY[]::int[]; BEGIN FOREACH r SLICE 2 IN ARRAY a LOOP totals := array_append(totals,cardinality(r)); END LOOP; RETURN totals; END $$;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT row_totals(NULL::int[]);",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "FOREACH expression must not be null", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_totals(ARRAY[]::int[]);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "slice dimension (1) is out of the valid range 0..0", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_totals(ARRAY[[1,2,3],[4,5,6]]);",
                    expected: Expected::Rows {
                        columns: &[Column("row_totals", INT8_ARRAY)],
                        rows: &[
                            &[T("{6,15}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_totals(ARRAY[[[1,2],[3,4]],[[5,6],[7,8]]]);",
                    expected: Expected::Rows {
                        columns: &[Column("row_totals", INT8_ARRAY)],
                        rows: &[
                            &[T("{3,7,11,15}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT planes(ARRAY[[[1,2],[3,4]],[[5,6],[7,8]]]);",
                    expected: Expected::Rows {
                        columns: &[Column("planes", INT4_ARRAY)],
                        rows: &[
                            &[T("{4,4}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT planes(ARRAY[1,2]);",
                    expected: Expected::Error(Diagnostic { code: "2202E", message: "slice dimension (2) is out of the valid range 0..1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT qa_slice(ARRAY[[1,2],[3,4]]),qa_slice_quoted(ARRAY[[1,2],[3,4]]);",
                    expected: Expected::Rows {
                        columns: &[Column("qa_slice", INT4), Column("qa_slice_quoted", INT4)],
                        rows: &[
                            &[T("4"), T("4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id,row_totals(a),planes(a) FROM slice_inputs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("row_totals", INT8_ARRAY), Column("planes", INT4_ARRAY)],
                        rows: &[
                            &[T("1"), T("{1,7}"), T("{4}")],
                            &[T("2"), T("{11}"), T("{2}")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_totals((SELECT a FROM slice_inputs WHERE id=1)),planes((SELECT a FROM slice_inputs WHERE id=2));",
                    expected: Expected::Rows {
                        columns: &[Column("row_totals", INT8_ARRAY), Column("planes", INT4_ARRAY)],
                        rows: &[
                            &[T("{1,7}"), T("{2}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT row_totals(ARRAY[NULL,NULL]::int[]),planes(ARRAY[[NULL,NULL]]::int[]);",
                    expected: Expected::Rows {
                        columns: &[Column("row_totals", INT8_ARRAY), Column("planes", INT4_ARRAY)],
                        rows: &[
                            &[T("{NULL}"), T("{2}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT slice_control(ARRAY[[1,2],[3,4],[5,6],[7,8]]);",
                    expected: Expected::Rows {
                        columns: &[Column("slice_control", INT4_ARRAY)],
                        rows: &[
                            &[T("{3,5}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT scalar_slice(ARRAY[[1,2]]);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: "FOREACH ... SLICE loop variable must be of an array type", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
