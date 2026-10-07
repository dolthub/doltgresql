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
fn test_pgvector_index_ddl() {
    run_scripts(&[
        ScriptTest {
            name: "every operator class creates an index under both methods",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE tv (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL, s sparsevec(3) NOT NULL, t text NOT NULL, vnodim vector NOT NULL, big vector(2001) NOT NULL, hbig halfvec(4001) NOT NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE INDEX idx1 ON tv USING hnsw (v vector_l2_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx2 ON tv USING hnsw (v vector_ip_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx3 ON tv USING hnsw (v vector_cosine_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx4 ON tv USING hnsw (v vector_l1_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx5 ON tv USING ivfflat (v vector_l2_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    notices: &[Diagnostic { code: "00000", message: "ivfflat index created with little data", detail: "This will cause low recall.", hint: "Drop the index until the table has more data.", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx6 ON tv USING ivfflat (v vector_ip_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    notices: &[Diagnostic { code: "00000", message: "ivfflat index created with little data", detail: "This will cause low recall.", hint: "Drop the index until the table has more data.", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx7 ON tv USING ivfflat (v vector_cosine_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    notices: &[Diagnostic { code: "00000", message: "ivfflat index created with little data", detail: "This will cause low recall.", hint: "Drop the index until the table has more data.", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx8 ON tv USING hnsw (h halfvec_l2_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx9 ON tv USING hnsw (h halfvec_ip_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx10 ON tv USING hnsw (h halfvec_cosine_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx11 ON tv USING hnsw (h halfvec_l1_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx12 ON tv USING ivfflat (h halfvec_l2_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    notices: &[Diagnostic { code: "00000", message: "ivfflat index created with little data", detail: "This will cause low recall.", hint: "Drop the index until the table has more data.", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx13 ON tv USING ivfflat (v vector_l1_ops);",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"operator class "vector_l1_ops" does not exist for access method "ivfflat""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "default operator classes",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE tv (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL, s sparsevec(3) NOT NULL, t text NOT NULL, vnodim vector NOT NULL, big vector(2001) NOT NULL, hbig halfvec(4001) NOT NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING ivfflat (v);",
                    expected: Expected::Tag("CREATE INDEX"),
                    notices: &[Diagnostic { code: "00000", message: "ivfflat index created with little data", detail: "This will cause low recall.", hint: "Drop the index until the table has more data.", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v);",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"data type vector has no default operator class for access method "hnsw""#, hint: "You must specify an operator class for the index or define a default operator class for the data type.", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING ivfflat (h);",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"data type halfvec has no default operator class for access method "ivfflat""#, hint: "You must specify an operator class for the index or define a default operator class for the data type.", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "operator class errors",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE tv (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL, s sparsevec(3) NOT NULL, t text NOT NULL, vnodim vector NOT NULL, big vector(2001) NOT NULL, hbig halfvec(4001) NOT NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v foo_ops);",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"operator class "foo_ops" does not exist for access method "hnsw""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING ivfflat (s sparsevec_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"operator class "sparsevec_l2_ops" does not exist for access method "ivfflat""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (s sparsevec_l2_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    flow: Flow::Exec,
                    skip: Some("Dolt's vector indexes cannot hold sparsevec values"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (h vector_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"operator class "vector_l2_ops" does not accept data type halfvec"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (t vector_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"operator class "vector_l2_ops" does not accept data type text"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "index restrictions",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE tv (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL, s sparsevec(3) NOT NULL, t text NOT NULL, vnodim vector NOT NULL, big vector(2001) NOT NULL, hbig halfvec(4001) NOT NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE UNIQUE INDEX ON tv USING hnsw (v vector_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"access method "hnsw" does not support unique indexes"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) INCLUDE (id);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"access method "hnsw" does not support included columns"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops, h halfvec_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"access method "hnsw" does not support multicolumn indexes"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WHERE id > 3;",
                    expected: Expected::Tag("CREATE INDEX"),
                    flow: Flow::Exec,
                    skip: Some("Dolt's vector indexes cannot be partial"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw ((l2_normalize(v)) vector_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "column does not have dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING gin (v);",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"data type vector has no default operator class for access method "gin""#, hint: "You must specify an operator class for the index or define a default operator class for the data type.", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "column dimension requirements",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE tv (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL, s sparsevec(3) NOT NULL, t text NOT NULL, vnodim vector NOT NULL, big vector(2001) NOT NULL, hbig halfvec(4001) NOT NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (vnodim vector_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "column does not have dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (big vector_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "column cannot have more than 2000 dimensions for hnsw index", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (hbig halfvec_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "column cannot have more than 4000 dimensions for hnsw index", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING ivfflat (big vector_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "column cannot have more than 2000 dimensions for ivfflat index", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING ivfflat (hbig halfvec_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "column cannot have more than 4000 dimensions for ivfflat index", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "storage parameters",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE tv (id INT4 PRIMARY KEY, v vector(3) NOT NULL, h halfvec(3) NOT NULL, s sparsevec(3) NOT NULL, t text NOT NULL, vnodim vector NOT NULL, big vector(2001) NOT NULL, hbig halfvec(4001) NOT NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE INDEX idx1 ON tv USING hnsw (v vector_l2_ops) WITH (m = 16, ef_construction = 64);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX idx2 ON tv USING ivfflat (v vector_l2_ops) WITH (lists = 100);",
                    expected: Expected::Tag("CREATE INDEX"),
                    notices: &[Diagnostic { code: "00000", message: "ivfflat index created with little data", detail: "This will cause low recall.", hint: "Drop the index until the table has more data.", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (m = 1);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"value 1 out of bounds for option "m""#, detail: r#"Valid values are between "2" and "100"."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (m = 101);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"value 101 out of bounds for option "m""#, detail: r#"Valid values are between "2" and "100"."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (ef_construction = 3);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"value 3 out of bounds for option "ef_construction""#, detail: r#"Valid values are between "4" and "1000"."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (ef_construction = 1001);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"value 1001 out of bounds for option "ef_construction""#, detail: r#"Valid values are between "4" and "1000"."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (m = 40, ef_construction = 64);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "ef_construction must be greater than or equal to 2 * m", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (m = 40);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "ef_construction must be greater than or equal to 2 * m", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (foo = 1);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"unrecognized parameter "foo""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (lists = 100);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"unrecognized parameter "lists""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING ivfflat (v vector_l2_ops) WITH (m = 16);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"unrecognized parameter "m""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING ivfflat (v vector_l2_ops) WITH (lists = 0);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"value 0 out of bounds for option "lists""#, detail: r#"Valid values are between "1" and "32768"."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING ivfflat (v vector_l2_ops) WITH (lists = 32769);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"value 32769 out of bounds for option "lists""#, detail: r#"Valid values are between "1" and "32768"."#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (m = 'abc');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"invalid value for integer option "m": abc"#, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "index maintenance",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE items (id INT4 PRIMARY KEY, v vector(3) NOT NULL);",
                "INSERT INTO items VALUES (1, '[1,1,1]'), (2, '[2,2,2]'), (3, '[3,3,3]');",
                "CREATE INDEX items_idx ON items USING hnsw (v vector_cosine_ops);",
                "INSERT INTO items VALUES (4, '[4,4,4]');",
                "UPDATE items SET v = '[9,9,9]' WHERE id = 2;",
                "DELETE FROM items WHERE id = 1;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, v FROM items ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", USER_DEFINED)],
                        rows: &[
                            &[T("2"), T("[9,9,9]")],
                            &[T("3"), T("[3,3,3]")],
                            &[T("4"), T("[4,4,4]")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM items WHERE v = '[9,9,9]';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "index maintenance with out-of-line vectors",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE items (id INT4 PRIMARY KEY, v vector(700) NOT NULL);",
                "INSERT INTO items VALUES (1, '[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99]'), (2, '[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0]');",
                "CREATE INDEX items_idx ON items USING hnsw (v vector_l2_ops);",
                "INSERT INTO items VALUES (3, '[2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1]');",
                "UPDATE items SET v = '[3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2]' WHERE id = 1;",
                "DELETE FROM items WHERE id = 2;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, v FROM items ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("v", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("[3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2]")],
                            &[T("3"), T("[2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1]")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_vector_index_rules() {
    run_scripts(&[
        ScriptTest {
            name: "vector index validation",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE EXTENSION vector;",
                    expected: Expected::Tag("CREATE EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE tv (id INT PRIMARY KEY, v vector(3), h halfvec(3), t TEXT, nodim vector, big vector(2001), hbig halfvec(4001));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING ivfflat (v);",
                    expected: Expected::Tag("CREATE INDEX"),
                    notices: &[Diagnostic { code: "00000", message: "ivfflat index created with little data", detail: "This will cause low recall.", hint: "Drop the index until the table has more data.", ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING ivfflat (v vector_l1_ops);",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"operator class "vector_l1_ops" does not exist for access method "ivfflat""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v);",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"data type vector has no default operator class for access method "hnsw""#, hint: "You must specify an operator class for the index or define a default operator class for the data type.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v foo_ops);",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"operator class "foo_ops" does not exist for access method "hnsw""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (h vector_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"operator class "vector_l2_ops" does not accept data type halfvec"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (t vector_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "42804", message: r#"operator class "vector_l2_ops" does not accept data type text"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE UNIQUE INDEX ON tv USING hnsw (v vector_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"access method "hnsw" does not support unique indexes"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) INCLUDE (id);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"access method "hnsw" does not support included columns"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops, h halfvec_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "0A000", message: r#"access method "hnsw" does not support multicolumn indexes"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw ((l2_normalize(v)) vector_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "column does not have dimensions", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (nodim vector_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "column does not have dimensions", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (big vector_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "column cannot have more than 2000 dimensions for hnsw index", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING ivfflat (hbig halfvec_l2_ops);",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "column cannot have more than 4000 dimensions for ivfflat index", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (m = 1);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"value 1 out of bounds for option "m""#, detail: r#"Valid values are between "2" and "100"."#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (m = 40);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "ef_construction must be greater than or equal to 2 * m", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (foo = 1);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"unrecognized parameter "foo""#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING hnsw (v vector_l2_ops) WITH (m = 'abc');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"invalid value for integer option "m": abc"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING ivfflat (v vector_l2_ops) WITH (lists = 0);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"value 0 out of bounds for option "lists""#, detail: r#"Valid values are between "1" and "32768"."#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING gin (v);",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"data type vector has no default operator class for access method "gin""#, hint: "You must specify an operator class for the index or define a default operator class for the data type.", ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ON tv USING nosuchmethod (v);",
                    expected: Expected::Error(Diagnostic { code: "42704", message: r#"access method "nosuchmethod" does not exist"#, ..E }),
                    flow: Flow::Query,
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX hv ON tv USING hnsw (v vector_cosine_ops) WITH (m = 16, ef_construction = 64);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX IF NOT EXISTS hv ON tv USING hnsw (v vector_cosine_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    notices: &[Diagnostic { code: "42P07", message: r#"relation "hv" already exists, skipping"#, ..N }],
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'tv' AND indexdef LIKE '%hnsw%' ORDER BY indexname;",
                    expected: Expected::Rows {
                        columns: &[Column("indexname", NAME), Column("indexdef", TEXT)],
                        rows: &[
                            &[T("hv"), T("CREATE INDEX hv ON public.tv USING hnsw (v vector_cosine_ops) WITH (m='16', ef_construction='64')")],
                            &[T("tv_v_idx"), T("CREATE INDEX tv_v_idx ON public.tv USING hnsw (v vector_l2_ops)")],
                        ],
                        tag: "SELECT 2",
                    },
                    skip: Some("Dolt stores only a vector index's distance, so every vector index renders with hnsw and no storage parameters"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_get_indexdef('hv'::regclass), pg_get_indexdef('hv'::regclass, 1, false), pg_get_indexdef(0);",
                    expected: Expected::Rows {
                        columns: &[Column("pg_get_indexdef", TEXT), Column("pg_get_indexdef", TEXT), Column("pg_get_indexdef", TEXT)],
                        rows: &[
                            &[T("CREATE INDEX hv ON public.tv USING hnsw (v vector_cosine_ops) WITH (m='16', ef_construction='64')"), T("v"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    skip: Some("Dolt stores only a vector index's distance, so every vector index renders with hnsw and no storage parameters"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT c.relname, a.amname FROM pg_class c JOIN pg_am a ON a.oid = c.relam WHERE c.relname = 'hv';",
                    expected: Expected::Rows {
                        columns: &[Column("relname", NAME), Column("amname", NAME)],
                        rows: &[
                            &[T("hv"), T("hnsw")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP INDEX hv;",
                    expected: Expected::Tag("DROP INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DROP INDEX tv_v_idx;",
                    expected: Expected::Tag("DROP INDEX"),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "vector index maintenance",
            assertions: &[
                ScriptTestAssertion {
                    query: "CREATE EXTENSION vector;",
                    expected: Expected::Tag("CREATE EXTENSION"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE TABLE items (id INT PRIMARY KEY, v vector(3), h halfvec(2));",
                    expected: Expected::Tag("CREATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO items VALUES (1, '[1,1,1]', '[1,0]'), (2, '[2,2,2]', '[0,1]'), (3, '[1,2,3]', '[1,1]'), (4, NULL, NULL);",
                    expected: Expected::Tag("INSERT 0 4"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX il2 ON items USING hnsw (v vector_l2_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX icos ON items USING hnsw (v vector_cosine_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "CREATE INDEX ihalf ON items USING hnsw (h halfvec_l1_ops);",
                    expected: Expected::Tag("CREATE INDEX"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM items ORDER BY v <-> '[1,1,2]' LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                            &[T("3")],
                        ],
                        tag: "SELECT 3",
                    },
                    skip: Some("Doltgres always answers ORDER BY distance LIMIT with the vector index, whose search orders the tied rows 2 and 3 as Dolt's does, while Postgres sorts this small table in row order"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO items VALUES (5, '[1,1,2]', '[0,0]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "UPDATE items SET v = '[9,9,9]' WHERE id = 1;",
                    expected: Expected::Tag("UPDATE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM items WHERE id = 2;",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM items ORDER BY v <-> '[1,1,2]' LIMIT 3;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("5")],
                            &[T("3")],
                            &[T("1")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM items ORDER BY v <=> '[1,1,2]' LIMIT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("5")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM items ORDER BY h <+> '[0,0]' LIMIT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("5")],
                            &[T("1")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "TRUNCATE items;",
                    expected: Expected::Tag("TRUNCATE TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM items;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "INSERT INTO items VALUES (6, '[3,3,3]', '[3,3]');",
                    expected: Expected::Tag("INSERT 0 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM items ORDER BY v <-> '[0,0,0]';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "ALTER TABLE items ADD COLUMN note TEXT;",
                    expected: Expected::Tag("ALTER TABLE"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, note FROM items ORDER BY v <-> '[0,0,0]';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("note", TEXT)],
                        rows: &[
                            &[T("6"), Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}
