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
fn test_pgvector() {
    run_scripts(&[
        ScriptTest {
            name: "pgvector vector type",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ' [ 1.5, -0.02 , 3 ] '::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1.5,-0.02,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1e38,-1e-38]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1e+38,-1e-38]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[100000,1000000,1234567]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[100000,1e+06,1.234567e+06]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[1,]""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT 'abc'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "abc""#, detail: r#"Vector contents must start with "["."#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[1,2""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[NaN]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "NaN not allowed in vector", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[Infinity,1]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "infinite value not allowed in vector", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1e39]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""1e39" is out of range for type vector"#, position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pgvector vector functions",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT vector_dims('[1,2,3]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("vector_dims", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_norm('[3,4]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("vector_norm", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('[1,2,3]'::vector, '[4,5,6]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8)],
                        rows: &[
                            &[T("5.196152422706632")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('[1,2]'::vector, '[3,4]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,2]'::vector, '[2,4]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,0]'::vector, '[0,0]'::vector)::text;",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", TEXT)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('[1,2]'::vector, '[4,7]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('[3,4]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("[0.6,0.8]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('[0,0]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("[0,0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT binary_quantize('[1,-1,0]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("binary_quantize", BIT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::vector, 2, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[2,3,4]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::vector, 4, 100);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[4,5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::vector, 6, 1);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pgvector vector operators",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector <-> '[4,5,6]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("5.196152422706632")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector <#> '[3,4]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("-11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector <=> '[2,4]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector <+> '[4,7]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector + '[3,4]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[4,6]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[5,6]'::vector - '[1,2]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[4,4]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector * '[3,4]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[3,8]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector || '[3]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector < '[1,3]'::vector, '[1,2]'::vector = '[1,2]'::vector, '[1,2]'::vector >= '[1,3]'::vector, '[1,2]'::vector != '[1,2]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL), Column("?column?", BOOL)],
                        rows: &[
                            &[T("t"), T("t"), T("f"), T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector < '[1,2,3]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector + '[1,2,3]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different vector dimensions 2 and 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[3e38,3e38]'::vector + '[3e38,3e38]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value out of range: overflow", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pgvector halfvec",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '[0.1,1]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[0.099975586,1]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[65504,-65504]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[65504,-65504]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[65520]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""65520" is out of range for type halfvec"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1e-8]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::halfvec <-> '[4,6]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::halfvec <#> '[3,4]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("-11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::halfvec <=> '[2,4]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::halfvec <+> '[4,7]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::halfvec + '[3,4]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[4,6]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('[3,4]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("[0.60009766,0.7998047]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT binary_quantize('[1,-1,0]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("binary_quantize", BIT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::halfvec, 2, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[2,3,4]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_dims('[1,2,3]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("vector_dims", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_norm('[3,4]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_norm", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pgvector sparsevec",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '{1:1.5,3:2}/5'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1.5,3:2}/5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{3:1,1:2}/5'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:2,3:1}/5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:0,2:3}/4'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{2:3}/4")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/7'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{}/7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{6:1}/5'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec index out of bounds", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{0:1}/5'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec index out of bounds", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,1:2}/5'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec indices must not contain duplicates", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1}/0'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec must have at least 1 dimension", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_norm('{1:3,2:4}/2'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_norm", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2}/3'::sparsevec <-> '{1:4,3:6}/3'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2}/2'::sparsevec <#> '{1:3,2:4}/2'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("-11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2}/2'::sparsevec <=> '{1:2,2:4}/2'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2}/2'::sparsevec <+> '{1:4,2:7}/2'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("8")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('{1:3,2:4}/3'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("{1:0.6,2:0.8}/3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pgvector bit distances",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '101'::bit(3) <~> '111'::bit(3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '101'::bit(3) <%> '111'::bit(3);",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("0.33333333333333337")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('101', '111');",
                    expected: Expected::Rows {
                        columns: &[Column("hamming_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('101', '111');",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("0.33333333333333337")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('000', '000');",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('10', '111');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different bit lengths 2 and 3", ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pgvector casts",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[0.1,1]'::halfvec::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[0.099975586,1]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,0,2]'::vector::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1,3:2}/3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,3:2}/3'::sparsevec::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,0,2]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,0,2]'::halfvec::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1,3:2}/3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1.5}/2'::sparsevec::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[1.5,0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,2,3]::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1.5,2.5]::float4[]::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("[1.5,2.5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1.5,2.5]::float8[]::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("[1.5,2.5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1.5,2.5]::numeric[]::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("[1.5,2.5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector::float4[];",
                    expected: Expected::Rows {
                        columns: &[Column("float4", FLOAT4_ARRAY)],
                        rows: &[
                            &[T("{1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,2]::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,0,2]::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1,3:2}/3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pgvector vectors in tables",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE items (id INTEGER PRIMARY KEY, embedding vector);",
                "INSERT INTO items VALUES (1, '[1,2]'), (2, '[4,7]'), (3, '[0,0]');",
                "CREATE TABLE hitems (id INTEGER PRIMARY KEY, embedding halfvec);",
                "INSERT INTO hitems VALUES (1, '[1,2]'), (2, '[4,7]'), (3, '[0,0]');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT embedding FROM items ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("embedding", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2]")],
                            &[T("[4,7]")],
                            &[T("[0,0]")],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM items ORDER BY embedding <-> '[1,1]' LIMIT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, embedding <=> '[1,1]' AS d FROM items WHERE id != 3 ORDER BY d;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("d", FLOAT8)],
                        rows: &[
                            &[T("2"), T("0.03523617876226781")],
                            &[T("1"), T("0.05131670194948623")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(embedding), sum(embedding) FROM items;",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED), Column("sum", USER_DEFINED)],
                        rows: &[
                            &[T("[1.6666666,3]"), T("[5,9]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(embedding) FROM items WHERE id > 100;",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(embedding) FROM items WHERE id > 100;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", USER_DEFINED)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(embedding), sum(embedding) FROM hitems;",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED), Column("sum", USER_DEFINED)],
                        rows: &[
                            &[T("[1.6669922,3]"), T("[5,9]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pgvector vectors outside the public schema",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE SCHEMA vfx;",
                "CREATE TABLE vfx.docs (id INT PRIMARY KEY, embedding vector(3) NOT NULL, src INT REFERENCES vfx.docs (id));",
                "INSERT INTO vfx.docs VALUES (1, '[1,2,3]', NULL), (2, '[4,5,6]', 1), (3, '[10,10,10]', NULL);",
                "CREATE TABLE vfx.q (id INT PRIMARY KEY, v public.vector(2));",
                "INSERT INTO vfx.q VALUES (1, '[1,1]');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id FROM vfx.docs ORDER BY embedding <-> '[1,2,4]' LIMIT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT v FROM vfx.q;",
                    expected: Expected::Rows {
                        columns: &[Column("v", USER_DEFINED)],
                        rows: &[
                            &[T("[1,1]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO vfx, public;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_dims(embedding) FROM docs WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("vector_dims", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SET search_path TO vfx;",
                    expected: Expected::Tag("SET"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_dims(embedding) FROM docs WHERE id = 1;",
                    expected: Expected::Error(Diagnostic { code: "42883", message: "function vector_dims(public.vector) does not exist", hint: "No function matches the given name and argument types. You might need to add explicit type casts.", position: 8, ..E }),
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pgvector distance operator with a bound parameter",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE items (id INT PRIMARY KEY, embedding vector(3));",
                "INSERT INTO items VALUES (1, '[1,2,3]'), (2, '[4,5,6]'), (3, '[10,10,10]');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id FROM items ORDER BY embedding <-> $1 LIMIT 2;",
                    bind_vars: &[BindVar::Str("[1,2,4]")],
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("2")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, embedding <-> $1 AS d FROM items ORDER BY d LIMIT 2;",
                    bind_vars: &[BindVar::Str("[1,2,4]")],
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("d", FLOAT8)],
                        rows: &[
                            &[T("1"), T("1")],
                            &[T("2"), T("4.69041575982343")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pgvector copy from csv",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE docs (id INT PRIMARY KEY, embedding vector(3) NOT NULL, src INT);",
                "CREATE TABLE hdocs (id INT PRIMARY KEY, embedding halfvec(3) NOT NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "COPY docs FROM STDIN (FORMAT CSV, HEADER TRUE);",
                    expected: Expected::Tag("COPY 3"),
                    copy_from_stdin_file: "csv-load-vector.sql",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id, src FROM docs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("src", INT4)],
                        rows: &[
                            &[T("1"), Null],
                            &[T("2"), T("1")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM docs ORDER BY embedding <-> '[1,2,4]' LIMIT 2;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                            &[T("3")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "COPY hdocs FROM STDIN (FORMAT CSV, HEADER TRUE);",
                    expected: Expected::Tag("COPY 2"),
                    copy_from_stdin_file: "csv-load-halfvec.sql",
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT embedding FROM hdocs ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("embedding", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[0.5,0.25,0.125]")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pgvector fixture schema shape",
            set_up_script: &[
                "BEGIN;",
                "CREATE EXTENSION IF NOT EXISTS vector;",
                "CREATE SCHEMA vf;",
                "CREATE TYPE vf.record_status AS ENUM ('active', 'archived');",
                "CREATE TABLE vf.workspaces (id uuid PRIMARY KEY, name text NOT NULL UNIQUE);",
                r#"CREATE TABLE vf.documents (
					id uuid PRIMARY KEY,
					workspace_id uuid NOT NULL REFERENCES vf.workspaces (id) ON DELETE CASCADE,
					status vf.record_status NOT NULL,
					fixture_source_id uuid REFERENCES vf.documents (id),
					embedding vector(4) NOT NULL
				);"#,
                "CREATE INDEX documents_filter_idx ON vf.documents (workspace_id, status, id);",
                "COMMIT;",
                "INSERT INTO vf.workspaces VALUES ('11111111-1111-1111-1111-111111111111', 'ws1'), ('22222222-2222-2222-2222-222222222222', 'ws2');",
                r#"INSERT INTO vf.documents VALUES
					('aaaaaaaa-0000-0000-0000-000000000001', '11111111-1111-1111-1111-111111111111', 'active', NULL, '[1,0,0,0]'),
					('aaaaaaaa-0000-0000-0000-000000000002', '11111111-1111-1111-1111-111111111111', 'archived', 'aaaaaaaa-0000-0000-0000-000000000001', '[1,0.01,0,0]'),
					('aaaaaaaa-0000-0000-0000-000000000003', '22222222-2222-2222-2222-222222222222', 'active', NULL, '[0,1,0,0]');"#,
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id FROM vf.documents WHERE workspace_id = '11111111-1111-1111-1111-111111111111' AND status = 'active' ORDER BY embedding <=> '[1,0.02,0,0]' LIMIT 1;",
                    expected: Expected::Rows {
                        columns: &[Column("id", UUID)],
                        rows: &[
                            &[T("aaaaaaaa-0000-0000-0000-000000000001")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "DELETE FROM vf.workspaces WHERE name = 'ws1';",
                    expected: Expected::Tag("DELETE 1"),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM vf.documents;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "pgvector catalog tables",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT extname, extrelocatable, extversion FROM pg_catalog.pg_extension WHERE extname = 'vector';",
                    expected: Expected::Rows {
                        columns: &[Column("extname", NAME), Column("extrelocatable", BOOL), Column("extversion", TEXT)],
                        rows: &[
                            &[T("vector"), T("t"), T("0.8.6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name, default_version, installed_version, comment FROM pg_catalog.pg_available_extensions WHERE name = 'vector';",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME), Column("default_version", TEXT), Column("installed_version", TEXT), Column("comment", TEXT)],
                        rows: &[
                            &[T("vector"), T("0.8.6"), T("0.8.6"), T("vector data type and ivfflat and hnsw access methods")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT name, version, installed, superuser, trusted, relocatable, schema, requires FROM pg_catalog.pg_available_extension_versions WHERE name = 'vector';",
                    expected: Expected::Rows {
                        columns: &[Column("name", NAME), Column("version", TEXT), Column("installed", BOOL), Column("superuser", BOOL), Column("trusted", BOOL), Column("relocatable", BOOL), Column("schema", NAME), Column("requires", NAME_ARRAY)],
                        rows: &[
                            &[T("vector"), T("0.8.6"), T("t"), T("t"), T("f"), T("t"), Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM pg_catalog.pg_proc WHERE proname IN ('l2_distance', 'inner_product', 'cosine_distance', 'l1_distance', 'l2_norm', 'l2_normalize', 'binary_quantize', 'subvector');",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("21")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "committing with the extension installed",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                // Doltgres-specific: Postgres cannot run this, so the Go server's output is expected.
                ScriptTestAssertion {
                    query: "SELECT count(*) FROM (SELECT dolt_commit('-Am', 'installed pgvector')) sq;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8)],
                        rows: &[
                            &[T("1")],
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

#[test]
fn test_pgvector_bit() {
    run_scripts(&[
        ScriptTest {
            name: "bit",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('111', '111');",
                    expected: Expected::Rows {
                        columns: &[Column("hamming_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('111', '110');",
                    expected: Expected::Rows {
                        columns: &[Column("hamming_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('111', '100');",
                    expected: Expected::Rows {
                        columns: &[Column("hamming_distance", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('111', '000');",
                    expected: Expected::Rows {
                        columns: &[Column("hamming_distance", FLOAT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('10101010101010101010', '01010101010101010101');",
                    expected: Expected::Rows {
                        columns: &[Column("hamming_distance", FLOAT8)],
                        rows: &[
                            &[T("20")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101', '101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101');",
                    expected: Expected::Rows {
                        columns: &[Column("hamming_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101', '010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010');",
                    expected: Expected::Rows {
                        columns: &[Column("hamming_distance", FLOAT8)],
                        rows: &[
                            &[T("513")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('110000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000011', '100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001');",
                    expected: Expected::Rows {
                        columns: &[Column("hamming_distance", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('', '');",
                    expected: Expected::Rows {
                        columns: &[Column("hamming_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('111', '00');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different bit lengths 3 and 2", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('111', '000'::varbit(4));",
                    expected: Expected::Rows {
                        columns: &[Column("hamming_distance", FLOAT8)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT hamming_distance('111', '0000'::varbit(4));",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different bit lengths 3 and 4", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('1111', '1111');",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('1111', '1110');",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("0.25")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('1111', '1100');",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("0.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('1111', '1000');",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("0.75")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('1111', '0000');",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('1100', '1000');",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("0.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('10101010101010101010', '01010101010101010101');",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101', '101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101');",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101', '010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010101010');",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('110000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000011', '100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001');",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("0.5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('', '');",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('1111', '000');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different bit lengths 4 and 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('1111', '0000'::varbit(5));",
                    expected: Expected::Rows {
                        columns: &[Column("jaccard_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT jaccard_distance('1111', '00000'::varbit(5));",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different bit lengths 4 and 5", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_pgvector_btree() {
    run_scripts(&[
        ScriptTest {
            name: "btree vector",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE t (val vector(3));",
                "INSERT INTO t (val) VALUES ('[0,0,0]'), ('[1,2,3]'), ('[1,1,1]'), (NULL);",
                "CREATE INDEX ON t (val);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t WHERE val = '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t ORDER BY val;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[0,0,0]")],
                            &[T("[1,1,1]")],
                            &[T("[1,2,3]")],
                            &[Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "btree halfvec",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE t (val halfvec(3));",
                "INSERT INTO t (val) VALUES ('[0,0,0]'), ('[1,2,3]'), ('[1,1,1]'), (NULL);",
                "CREATE INDEX ON t (val);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t WHERE val = '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t ORDER BY val;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("[0,0,0]")],
                            &[T("[1,1,1]")],
                            &[T("[1,2,3]")],
                            &[Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "btree sparsevec",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE t (val sparsevec(3));",
                "INSERT INTO t (val) VALUES ('{}/3'), ('{1:1,2:2,3:3}/3'), ('{1:1,2:1,3:1}/3'), (NULL);",
                "CREATE INDEX ON t (val);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT * FROM t WHERE val = '{1:1,2:2,3:3}/3';",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1,2:2,3:3}/3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT * FROM t ORDER BY val;",
                    expected: Expected::Rows {
                        columns: &[Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("{}/3")],
                            &[T("{1:1,2:1,3:1}/3")],
                            &[T("{1:1,2:2,3:3}/3")],
                            &[Null],
                        ],
                        tag: "SELECT 4",
                    },
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_pgvector_cast() {
    run_scripts(&[
        ScriptTest {
            name: "cast",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,2,3]::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1.0,2.0,3.0]::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,2,3]::float4[]::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,2,3]::float8[]::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,2,3]::numeric[]::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector::real[];",
                    expected: Expected::Rows {
                        columns: &[Column("float4", FLOAT4_ARRAY)],
                        rows: &[
                            &[T("{1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,2,3}'::real[]::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,2,3}'::real[]::vector(3);",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,2,3}'::real[]::vector(2);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{NULL}'::real[]::vector;",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "array must not contain nulls", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{NaN}'::real[]::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "NaN not allowed in vector", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{Infinity}'::real[]::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "infinite value not allowed in vector", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{-Infinity}'::real[]::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "infinite value not allowed in vector", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}'::real[]::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{{1}}'::real[]::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "array must be 1-D", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,2,3}'::double precision[]::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,2,3}'::double precision[]::vector(3);",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,2,3}'::double precision[]::vector(2);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{4e38,-4e38}'::double precision[]::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "infinite value not allowed in vector", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1e-46,-1e-46}'::double precision[]::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[0,-0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector::halfvec(3);",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector::halfvec(2);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[65520]'::vector::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""65520" is out of range for type halfvec"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1e-8]'::vector::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec::vector(3);",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec::vector(2);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,2,3}'::real[]::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,2,3}'::real[]::halfvec(3);",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,2,3}'::real[]::halfvec(2);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{65520,-65520}'::real[]::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""65520" is out of range for type halfvec"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1e-8,-1e-8}'::real[]::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[0,-0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[0,1.5,0,3.5,0]'::vector::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{2:1.5,4:3.5}/5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[0,1.5,0,3.5,0]'::vector::sparsevec(5);",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{2:1.5,4:3.5}/5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[0,1.5,0,3.5,0]'::vector::sparsevec(4);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 4 dimensions, not 5", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{2:1.5,4:3.5}/5'::sparsevec::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[0,1.5,0,3.5,0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{2:1.5,4:3.5}/5'::sparsevec::vector(5);",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[0,1.5,0,3.5,0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{2:1.5,4:3.5}/5'::sparsevec::vector(4);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 4 dimensions, not 5", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/16001'::sparsevec::vector;",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "vector cannot have more than 16000 dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[0,1.5,0,3.5,0]'::halfvec::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{2:1.5,4:3.5}/5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[0,1.5,0,3.5,0]'::halfvec::sparsevec(5);",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{2:1.5,4:3.5}/5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[0,1.5,0,3.5,0]'::halfvec::sparsevec(4);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 4 dimensions, not 5", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{2:1.5,4:3.5}/5'::sparsevec::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[0,1.5,0,3.5,0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{2:1.5,4:3.5}/5'::sparsevec::halfvec(5);",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[0,1.5,0,3.5,0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{2:1.5,4:3.5}/5'::sparsevec::halfvec(4);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 4 dimensions, not 5", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/16001'::sparsevec::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "halfvec cannot have more than 16000 dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:65520}/1'::sparsevec::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""65520" is out of range for type halfvec"#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1e-8}/1'::sparsevec::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,0,2,0,3,0]::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1,3:2,5:3}/6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1.0,0.0,2.0,0.0,3.0,0.0]::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1,3:2,5:3}/6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,0,2,0,3,0]::float4[]::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1,3:2,5:3}/6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,0,2,0,3,0]::float8[]::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1,3:2,5:3}/6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,0,2,0,3,0]::numeric[]::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("array", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1,3:2,5:3}/6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,0,2,0,3,0}'::real[]::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1,3:2,5:3}/6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,0,2,0,3,0}'::real[]::sparsevec(6);",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1,3:2,5:3}/6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,0,2,0,3,0}'::real[]::sparsevec(5);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 5 dimensions, not 6", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{NULL}'::real[]::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22004", message: "array must not contain nulls", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{NaN}'::real[]::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "NaN not allowed in sparsevec", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{Infinity}'::real[]::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "infinite value not allowed in sparsevec", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{-Infinity}'::real[]::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "infinite value not allowed in sparsevec", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}'::real[]::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec must have at least 1 dimension", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{{1}}'::real[]::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "array must be 1-D", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(n)::vector FROM generate_series(1, 16001) n;",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "vector cannot have more than 16000 dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_to_vector(array_agg(n), 16001, false) FROM generate_series(1, 16001) n;",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "vector cannot have more than 16000 dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(n)::halfvec FROM generate_series(1, 16001) n;",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "halfvec cannot have more than 16000 dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_agg(n)::sparsevec FROM generate_series(1, 16001) n;",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "sparsevec cannot have more than 16000 non-zero elements", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ARRAY[1,2,3] = ARRAY[1,2,3];",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
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

#[test]
fn test_pgvector_halfvec() {
    run_scripts(&[
        ScriptTest {
            name: "halfvec",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[-1,-2,-3]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[-1,-2,-3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1.,2.,3.]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ' [ 1,  2 ,    3  ] '::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1.23456]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[1.234375]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[hello,1]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type halfvec: "[hello,1]""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[NaN,1]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "NaN not allowed in halfvec", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[Infinity,1]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "infinite value not allowed in halfvec", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[-Infinity,1]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "infinite value not allowed in halfvec", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[65519,-65519]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[65504,-65504]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[65520,-65520]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""65520" is out of range for type halfvec"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1e-8,-1e-8]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[0,-0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[4e38,1]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""4e38" is out of range for type halfvec"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1e-46,1]'::halfvec;",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[0,1]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type halfvec: "[1,2,3""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]9'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type halfvec: "[1,2,3]9""#, detail: "Junk after closing right brace.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1,2,3'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type halfvec: "1,2,3""#, detail: r#"Vector contents must start with "["."#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ''::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type halfvec: """#, detail: r#"Vector contents must start with "["."#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '['::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type halfvec: "[""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[ '::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type halfvec: "[ ""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[,'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type halfvec: "[,""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "halfvec must have at least 1 dimension", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[ ]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "halfvec must have at least 1 dimension", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[,]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type halfvec: "[,]""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type halfvec: "[1,]""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1a]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type halfvec: "[1a]""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,,3]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type halfvec: "[1,,3]""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1, ,3]'::halfvec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type halfvec: "[1, ,3]""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec(3);",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec(2);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec(3, 2);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "invalid type modifier", position: 19, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec('a');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 19, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec(0);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "dimensions for type halfvec must be at least 1", position: 19, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec(16001);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "dimensions for type halfvec cannot exceed 16000", position: 19, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT unnest('{"[1,2,3]", "[4,5,6]"}'::halfvec[]);"#,
                    expected: Expected::Rows {
                        columns: &[Column("unnest", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[4,5,6]")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"[1,2,3]"}'::halfvec(2)[];"#,
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec + '[4,5,6]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[5,7,9]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[65519]'::halfvec + '[65519]';",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value out of range: overflow", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::halfvec + '[3]';",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different halfvec dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec - '[4,5,6]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[-3,-3,-3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[-65519]'::halfvec - '[65519]';",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value out of range: overflow", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::halfvec - '[3]';",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different halfvec dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec * '[4,5,6]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[4,10,18]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[65519]'::halfvec * '[65519]';",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value out of range: overflow", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1e-7]'::halfvec * '[1e-7]';",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value out of range: underflow", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::halfvec * '[3]';",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different halfvec dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec || '[4,5]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3,4,5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill(0, ARRAY[16000])::halfvec || '[1]';",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "halfvec cannot have more than 16000 dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec < '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec < '[1,2]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec <= '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec <= '[1,2]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec = '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec = '[1,2]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec != '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec != '[1,2]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec >= '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec >= '[1,2]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec > '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::halfvec > '[1,2]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_cmp('[1,2,3]', '[1,2,3]');",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec_cmp", INT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_cmp('[1,2,3]', '[0,0,0]');",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec_cmp", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_cmp('[0,0,0]', '[1,2,3]');",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec_cmp", INT4)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_cmp('[1,2]', '[1,2,3]');",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec_cmp", INT4)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_cmp('[1,2,3]', '[1,2]');",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec_cmp", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_cmp('[1,2]', '[2,3,4]');",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec_cmp", INT4)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_cmp('[2,3]', '[1,2,3]');",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec_cmp", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_dims('[1,2,3]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("vector_dims", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT round(l2_norm('[1,1]'::halfvec)::numeric, 5);",
                    expected: Expected::Rows {
                        columns: &[Column("round", NUMERIC)],
                        rows: &[
                            &[T("1.41421")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_norm('[3,4]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_norm", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_norm('[0,1]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_norm", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_norm('[0,0]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_norm", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_norm('[2]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_norm", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('[0,0]'::halfvec, '[3,4]');",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('[0,0]'::halfvec, '[0,1]');",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('[1,2]'::halfvec, '[3]');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different halfvec dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('[1,1,1,1,1,1,1,1,1]'::halfvec, '[1,1,1,1,1,1,1,4,5]');",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[0,0]'::halfvec <-> '[3,4]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('[1,2]'::halfvec, '[3,4]');",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('[1,2]'::halfvec, '[3]');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different halfvec dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('[65504]'::halfvec, '[65504]');",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("4290774016")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('[1,1,1,1,1,1,1,1,1]'::halfvec, '[1,2,3,4,5,6,7,8,9]');",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("45")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::halfvec <#> '[3,4]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("-11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,2]'::halfvec, '[2,4]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,2]'::halfvec, '[0,0]')::text;",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", TEXT)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,1]'::halfvec, '[1,1]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,0]'::halfvec, '[0,2]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,1]'::halfvec, '[-1,-1]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,2]'::halfvec, '[3]');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different halfvec dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,1]'::halfvec, '[1.1,1.1]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,1]'::halfvec, '[-1.1,-1.1]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,2,3,4,5,6,7,8,9]'::halfvec, '[1,2,3,4,5,6,7,8,9]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,2,3,4,5,6,7,8,9]'::halfvec, '[-1,-2,-3,-4,-5,-6,-7,-8,-9]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::halfvec <=> '[2,4]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('[0,0]'::halfvec, '[3,4]');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('[0,0]'::halfvec, '[0,1]');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('[1,2]'::halfvec, '[3]');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different halfvec dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('[1,2,3,4,5,6,7,8,9]'::halfvec, '[1,2,3,4,5,6,7,8,9]');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('[1,2,3,4,5,6,7,8,9]'::halfvec, '[0,3,2,5,4,7,6,9,8]');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[0,0]'::halfvec <+> '[3,4]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('[3,4]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("[0.60009766,0.7998047]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('[3,0]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("[1,0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('[0,0.1]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("[0,1]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('[0,0]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("[0,0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('[65504]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("[1]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT binary_quantize('[1,0,-1]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("binary_quantize", BIT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT binary_quantize('[0,0.1,-0.2,-0.3,0.4,0.5,0.6,-0.7,0.8,-0.9,1]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("binary_quantize", BIT)],
                        rows: &[
                            &[T("01001110101")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT binary_quantize('[1,2,3,-4,5,6,-7,8,1,-2,-3,4,5,-6,7,8,-1,2,3]'::halfvec);",
                    expected: Expected::Rows {
                        columns: &[Column("binary_quantize", BIT)],
                        rows: &[
                            &[T("1110110110011011011")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::halfvec, 1, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::halfvec, 3, 2);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[3,4]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::halfvec, -1, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[1]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::halfvec, 3, 9);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[3,4,5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::halfvec, 1, 0);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "halfvec must have at least 1 dimension", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::halfvec, 3, -1);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "halfvec must have at least 1 dimension", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::halfvec, -1, 2);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "halfvec must have at least 1 dimension", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::halfvec, 2147483647, 10);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "halfvec must have at least 1 dimension", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::halfvec, 3, 2147483647);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[3,4,5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::halfvec, -2147483644, 2147483647);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v) FROM unnest(ARRAY['[1,2,3]'::halfvec, '[3,5,7]']) v;",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED)],
                        rows: &[
                            &[T("[2,3.5,5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v) FROM unnest(ARRAY['[1,2,3]'::halfvec, '[3,5,7]', NULL]) v;",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED)],
                        rows: &[
                            &[T("[2,3.5,5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v) FROM unnest(ARRAY[]::halfvec[]) v;",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v) FROM unnest(ARRAY['[1,2]'::halfvec, '[3]']) v;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v) FROM unnest(ARRAY['[65504]'::halfvec, '[65504]']) v;",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED)],
                        rows: &[
                            &[T("[65504]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_avg('{2,2,4,6}');",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec_avg", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_avg('{0}');",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec_avg", USER_DEFINED)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_avg('{1}');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "halfvec must have at least 1 dimension", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_avg('{{2,2,4,6}}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "halfvec_avg: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_avg('{NULL,2,4,6}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "halfvec_avg: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_avg('{}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "halfvec_avg: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_avg(array_agg(n)) FROM generate_series(1, 16002) n;",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "halfvec cannot have more than 16000 dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_accum('{0}', '[1,2,3]');",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec_accum", FLOAT8_ARRAY)],
                        rows: &[
                            &[T("{1,1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_accum('{0,0,0,0}', '[1,2,3]');",
                    expected: Expected::Rows {
                        columns: &[Column("halfvec_accum", FLOAT8_ARRAY)],
                        rows: &[
                            &[T("{1,1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_accum('{{0}}', '[1,2,3]');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "halfvec_accum: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_accum('{NULL}', '[1,2,3]');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "halfvec_accum: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_accum('{}', '[1,2,3]');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "halfvec_accum: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT halfvec_accum('{0,0}', '[1,2,3]');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 1 dimensions, not 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(v) FROM unnest(ARRAY['[1,2,3]'::halfvec, '[3,5,7]']) v;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", USER_DEFINED)],
                        rows: &[
                            &[T("[4,7,10]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(v) FROM unnest(ARRAY['[1,2,3]'::halfvec, '[3,5,7]', NULL]) v;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", USER_DEFINED)],
                        rows: &[
                            &[T("[4,7,10]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(v) FROM unnest(ARRAY[]::halfvec[]) v;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", USER_DEFINED)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(v) FROM unnest(ARRAY['[1,2]'::halfvec, '[3]']) v;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different halfvec dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(v) FROM unnest(ARRAY['[65504]'::halfvec, '[65504]']) v;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value out of range: overflow", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}

#[test]
fn test_pgvector_sparsevec() {
    run_scripts(&[
        ScriptTest {
            name: "sparsevec",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '{1:1.5,3:3.5}/5'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1.5,3:3.5}/5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:-2,3:-4}/5'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:-2,3:-4}/5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:2.,3:4.}/5'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:2,3:4}/5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ' { 1 : 1.5 ,  3  :  3.5  } / 5 '::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1.5,3:3.5}/5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1.23456}/1'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1.23456}/1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:hello,2:1}/2'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{1:hello,2:1}/2""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:NaN,2:1}/2'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "NaN not allowed in sparsevec", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:Infinity,2:1}/2'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "infinite value not allowed in sparsevec", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:-Infinity,2:1}/2'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "infinite value not allowed in sparsevec", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1.5e38,2:-1.5e38}/2'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1.5e+38,2:-1.5e+38}/2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1.5e+38,2:-1.5e+38}/2'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1.5e+38,2:-1.5e+38}/2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1.5e-38,2:-1.5e-38}/2'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1.5e-38,2:-1.5e-38}/2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:4e38,2:1}/2'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""4e38" is out of range for type sparsevec"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:-4e38,2:1}/2'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""-4e38" is out of range for type sparsevec"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1e-46,2:1}/2'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""1e-46" is out of range for type sparsevec"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:-1e-46,2:1}/2'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""-1e-46" is out of range for type sparsevec"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ''::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: """#, detail: r#"Vector contents must start with "{"."#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{ '::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{ ""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{:'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{:""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{,'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{,""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{}""#, detail: "Unexpected end of input.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{}/""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/1'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{}/1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/1a'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{}/1a""#, detail: "Junk after closing.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{ }/1'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{}/1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{:}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{:}/1""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{,}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{,}/1""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1,}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{1,}/1""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{:1}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{:1}/1""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{1:}/1""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1a:1}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{1a:1}/1""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1a}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{1:1a}/1""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type sparsevec: "{1:1,}/1""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:0,2:1,3:0}/3'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{2:1}/3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{2:1,1:1}/2'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1,2:1}/2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,1:1}/2'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec indices must not contain duplicates", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:1,1:1}/2'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec indices must not contain duplicates", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/5'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{}/5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/-1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec must have at least 1 dimension", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/1000000000'::sparsevec;",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{}/1000000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/1000000001'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "sparsevec cannot have more than 1000000000 dimensions", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/2147483648'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "sparsevec cannot have more than 1000000000 dimensions", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/-2147483649'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec must have at least 1 dimension", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/9223372036854775808'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "sparsevec cannot have more than 1000000000 dimensions", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/-9223372036854775809'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec must have at least 1 dimension", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{2147483647:1}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec index out of bounds", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{2147483648:1}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec index out of bounds", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{-2147483648:1}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec index out of bounds", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{-2147483649:1}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec index out of bounds", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{0:1}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec index out of bounds", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{2:1}/1'::sparsevec;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "sparsevec index out of bounds", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/3'::sparsevec(3);",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec", USER_DEFINED)],
                        rows: &[
                            &[T("{}/3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/3'::sparsevec(2);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/3'::sparsevec(3, 2);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "invalid type modifier", position: 16, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/3'::sparsevec('a');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 16, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/3'::sparsevec(0);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "dimensions for type sparsevec must be at least 1", position: 16, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/3'::sparsevec(1000000001);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "dimensions for type sparsevec cannot exceed 1000000000", position: 16, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2,3:3}/3'::sparsevec < '{1:1,2:2,3:3}/3';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2,3:3}/3'::sparsevec < '{1:1,2:2}/2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2,3:3}/3'::sparsevec <= '{1:1,2:2,3:3}/3';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2,3:3}/3'::sparsevec <= '{1:1,2:2}/2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2,3:3}/3'::sparsevec = '{1:1,2:2,3:3}/3';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2,3:3}/3'::sparsevec = '{1:1,2:2}/2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2,3:3}/3'::sparsevec != '{1:1,2:2,3:3}/3';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2,3:3}/3'::sparsevec != '{1:1,2:2}/2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2,3:3}/3'::sparsevec >= '{1:1,2:2,3:3}/3';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2,3:3}/3'::sparsevec >= '{1:1,2:2}/2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2,3:3}/3'::sparsevec > '{1:1,2:2,3:3}/3';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2,3:3}/3'::sparsevec > '{1:1,2:2}/2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sparsevec_cmp('{1:1,2:2,3:3}/3', '{1:1,2:2,3:3}/3');",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec_cmp", INT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sparsevec_cmp('{1:1,2:2,3:3}/3', '{}/3');",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec_cmp", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sparsevec_cmp('{}/3', '{1:1,2:2,3:3}/3');",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec_cmp", INT4)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sparsevec_cmp('{1:1,2:2}/2', '{1:1,2:2,3:3}/3');",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec_cmp", INT4)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sparsevec_cmp('{1:1,2:2,3:3}/3', '{1:1,2:2}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec_cmp", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sparsevec_cmp('{1:1,2:2}/2', '{1:2,2:3,3:4}/3');",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec_cmp", INT4)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sparsevec_cmp('{1:2,2:3}/2', '{1:1,2:2,3:3}/3');",
                    expected: Expected::Rows {
                        columns: &[Column("sparsevec_cmp", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT round(l2_norm('{1:1,2:1}/2'::sparsevec)::numeric, 5);",
                    expected: Expected::Rows {
                        columns: &[Column("round", NUMERIC)],
                        rows: &[
                            &[T("1.41421")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_norm('{1:3,2:4}/2'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_norm", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_norm('{2:1}/2'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_norm", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_norm('{1:3e37,2:4e37}/2'::sparsevec)::real;",
                    expected: Expected::Rows {
                        columns: &[Column("l2_norm", FLOAT4)],
                        rows: &[
                            &[T("5e+37")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_norm('{}/2'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_norm", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_norm('{1:2}/1'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_norm", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('{}/2'::sparsevec, '{1:3,2:4}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('{1:3}/2'::sparsevec, '{2:4}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('{2:4}/2'::sparsevec, '{1:3}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('{1:3,2:4}/2'::sparsevec, '{}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('{}/2'::sparsevec, '{2:1}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/2'::sparsevec <-> '{1:3,2:4}/2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('{1:1,2:2}/2'::sparsevec, '{1:2,2:4}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("10")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('{1:1,2:2}/2'::sparsevec, '{1:3}/1');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different sparsevec dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('{1:1,3:3}/4'::sparsevec, '{2:2,4:4}/4');",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('{2:2,4:4}/4'::sparsevec, '{1:1,3:3}/4');",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('{1:1,3:3,5:5}/5'::sparsevec, '{2:4,3:6,4:8}/5');",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("18")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('{1:1}/2'::sparsevec, '{}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('{}/2'::sparsevec, '{1:1}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('{1:3e38}/1'::sparsevec, '{1:3e38}/1');",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2}/2'::sparsevec <#> '{1:3,2:4}/2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("-11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('{1:1,2:2}/2'::sparsevec, '{1:2,2:4}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('{1:1,2:2}/2'::sparsevec, '{}/2')::text;",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", TEXT)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('{1:1,2:1}/2'::sparsevec, '{1:1,2:1}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('{1:1}/2'::sparsevec, '{2:2}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('{1:1,2:1}/2'::sparsevec, '{1:-1,2:-1}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('{1:2}/2'::sparsevec, '{2:2}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('{2:2}/2'::sparsevec, '{1:2}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('{1:1,2:2}/2'::sparsevec, '{1:3}/1');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different sparsevec dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('{1:1,2:1}/2'::sparsevec, '{1:1.1,2:1.1}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('{1:1,2:1}/2'::sparsevec, '{1:-1.1,2:-1.1}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('{1:3e38}/1'::sparsevec, '{1:3e38}/1')::text;",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", TEXT)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('{}/1'::sparsevec, '{}/1')::text;",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", TEXT)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{1:1,2:2}/2'::sparsevec <=> '{1:2,2:4}/2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('{}/2'::sparsevec, '{1:3,2:4}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('{}/2'::sparsevec, '{2:1}/2');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('{1:1,2:2}/2'::sparsevec, '{1:3}/1');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different sparsevec dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('{1:3e38}/1'::sparsevec, '{1:-3e38}/1');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('{1:1,3:3,5:5,7:7}/8'::sparsevec, '{2:2,4:4,6:6,8:8}/8');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("36")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('{1:1,3:3,5:5,7:7,9:9}/9'::sparsevec, '{2:2,4:4,6:6,8:8}/9');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("45")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '{}/2'::sparsevec <+> '{1:3,2:4}/2';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('{1:3,2:4}/2'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("{1:0.6,2:0.8}/2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('{1:3}/2'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1}/2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('{2:0.1}/2'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("{2:1}/2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('{}/2'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("{}/2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('{1:3e38}/1'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1}/1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('{1:3e38,2:1e-37}/2'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("{1:1}/2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('{2:3e37,4:3e-37,6:4e37,8:4e-37}/9'::sparsevec);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("{2:0.6,6:0.8}/9")],
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

#[test]
fn test_pgvector_storage() {
    run_scripts(&[
        ScriptTest {
            name: "wire binary format",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT encode(vector_send('[1,2,3]'::vector), 'hex');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T("000300003f8000004000000040400000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode(halfvec_send('[1,2,3]'::halfvec), 'hex');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T("000300003c0040004200")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode(sparsevec_send('{1:1.5,3:-2}/100'::sparsevec), 'hex');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T("00000064000000020000000000000000000000023fc00000c0000000")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "vector storage round-trip",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE t (id INT4 PRIMARY KEY, val vector(700));",
                "INSERT INTO t VALUES (1, '[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99]'), (2, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, val FROM t ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99]")],
                            &[T("2"), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t WHERE val = '[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99]';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_norm(val) FROM t WHERE id = 1;",
                    expected: Expected::Rows {
                        columns: &[Column("vector_norm", FLOAT8)],
                        rows: &[
                            &[T("1516.0639828186672")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "halfvec storage round-trip",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE t (id INT4 PRIMARY KEY, val halfvec(1100));",
                "INSERT INTO t VALUES (1, '[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99]'), (2, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, val FROM t ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99]")],
                            &[T("2"), Null],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t WHERE val = '[0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99]';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
            ],
            ..S
        },
        ScriptTest {
            name: "sparsevec storage round-trip",
            set_up_script: &[
                "CREATE EXTENSION vector;",
                "CREATE TABLE t (id INT4 PRIMARY KEY, val sparsevec(1000));",
                "INSERT INTO t VALUES (1, '{1:1.5,2:2.5,3:3.5,4:4.5,5:5.5,6:6.5,7:7.5,8:8.5,9:0.5,10:1.5,11:2.5,12:3.5,13:4.5,14:5.5,15:6.5,16:7.5,17:8.5,18:0.5,19:1.5,20:2.5,21:3.5,22:4.5,23:5.5,24:6.5,25:7.5,26:8.5,27:0.5,28:1.5,29:2.5,30:3.5,31:4.5,32:5.5,33:6.5,34:7.5,35:8.5,36:0.5,37:1.5,38:2.5,39:3.5,40:4.5,41:5.5,42:6.5,43:7.5,44:8.5,45:0.5,46:1.5,47:2.5,48:3.5,49:4.5,50:5.5,51:6.5,52:7.5,53:8.5,54:0.5,55:1.5,56:2.5,57:3.5,58:4.5,59:5.5,60:6.5,61:7.5,62:8.5,63:0.5,64:1.5,65:2.5,66:3.5,67:4.5,68:5.5,69:6.5,70:7.5,71:8.5,72:0.5,73:1.5,74:2.5,75:3.5,76:4.5,77:5.5,78:6.5,79:7.5,80:8.5,81:0.5,82:1.5,83:2.5,84:3.5,85:4.5,86:5.5,87:6.5,88:7.5,89:8.5,90:0.5,91:1.5,92:2.5,93:3.5,94:4.5,95:5.5,96:6.5,97:7.5,98:8.5,99:0.5,100:1.5,101:2.5,102:3.5,103:4.5,104:5.5,105:6.5,106:7.5,107:8.5,108:0.5,109:1.5,110:2.5,111:3.5,112:4.5,113:5.5,114:6.5,115:7.5,116:8.5,117:0.5,118:1.5,119:2.5,120:3.5,121:4.5,122:5.5,123:6.5,124:7.5,125:8.5,126:0.5,127:1.5,128:2.5,129:3.5,130:4.5,131:5.5,132:6.5,133:7.5,134:8.5,135:0.5,136:1.5,137:2.5,138:3.5,139:4.5,140:5.5,141:6.5,142:7.5,143:8.5,144:0.5,145:1.5,146:2.5,147:3.5,148:4.5,149:5.5,150:6.5,151:7.5,152:8.5,153:0.5,154:1.5,155:2.5,156:3.5,157:4.5,158:5.5,159:6.5,160:7.5,161:8.5,162:0.5,163:1.5,164:2.5,165:3.5,166:4.5,167:5.5,168:6.5,169:7.5,170:8.5,171:0.5,172:1.5,173:2.5,174:3.5,175:4.5,176:5.5,177:6.5,178:7.5,179:8.5,180:0.5,181:1.5,182:2.5,183:3.5,184:4.5,185:5.5,186:6.5,187:7.5,188:8.5,189:0.5,190:1.5,191:2.5,192:3.5,193:4.5,194:5.5,195:6.5,196:7.5,197:8.5,198:0.5,199:1.5,200:2.5,201:3.5,202:4.5,203:5.5,204:6.5,205:7.5,206:8.5,207:0.5,208:1.5,209:2.5,210:3.5,211:4.5,212:5.5,213:6.5,214:7.5,215:8.5,216:0.5,217:1.5,218:2.5,219:3.5,220:4.5,221:5.5,222:6.5,223:7.5,224:8.5,225:0.5,226:1.5,227:2.5,228:3.5,229:4.5,230:5.5,231:6.5,232:7.5,233:8.5,234:0.5,235:1.5,236:2.5,237:3.5,238:4.5,239:5.5,240:6.5,241:7.5,242:8.5,243:0.5,244:1.5,245:2.5,246:3.5,247:4.5,248:5.5,249:6.5,250:7.5,251:8.5,252:0.5,253:1.5,254:2.5,255:3.5,256:4.5,257:5.5,258:6.5,259:7.5,260:8.5,261:0.5,262:1.5,263:2.5,264:3.5,265:4.5,266:5.5,267:6.5,268:7.5,269:8.5,270:0.5,271:1.5,272:2.5,273:3.5,274:4.5,275:5.5,276:6.5,277:7.5,278:8.5,279:0.5,280:1.5,281:2.5,282:3.5,283:4.5,284:5.5,285:6.5,286:7.5,287:8.5,288:0.5,289:1.5,290:2.5,291:3.5,292:4.5,293:5.5,294:6.5,295:7.5,296:8.5,297:0.5,298:1.5,299:2.5,300:3.5}/1000'), (2, '{}/1000'), (3, NULL);",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT id, val FROM t ORDER BY id;",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4), Column("val", USER_DEFINED)],
                        rows: &[
                            &[T("1"), T("{1:1.5,2:2.5,3:3.5,4:4.5,5:5.5,6:6.5,7:7.5,8:8.5,9:0.5,10:1.5,11:2.5,12:3.5,13:4.5,14:5.5,15:6.5,16:7.5,17:8.5,18:0.5,19:1.5,20:2.5,21:3.5,22:4.5,23:5.5,24:6.5,25:7.5,26:8.5,27:0.5,28:1.5,29:2.5,30:3.5,31:4.5,32:5.5,33:6.5,34:7.5,35:8.5,36:0.5,37:1.5,38:2.5,39:3.5,40:4.5,41:5.5,42:6.5,43:7.5,44:8.5,45:0.5,46:1.5,47:2.5,48:3.5,49:4.5,50:5.5,51:6.5,52:7.5,53:8.5,54:0.5,55:1.5,56:2.5,57:3.5,58:4.5,59:5.5,60:6.5,61:7.5,62:8.5,63:0.5,64:1.5,65:2.5,66:3.5,67:4.5,68:5.5,69:6.5,70:7.5,71:8.5,72:0.5,73:1.5,74:2.5,75:3.5,76:4.5,77:5.5,78:6.5,79:7.5,80:8.5,81:0.5,82:1.5,83:2.5,84:3.5,85:4.5,86:5.5,87:6.5,88:7.5,89:8.5,90:0.5,91:1.5,92:2.5,93:3.5,94:4.5,95:5.5,96:6.5,97:7.5,98:8.5,99:0.5,100:1.5,101:2.5,102:3.5,103:4.5,104:5.5,105:6.5,106:7.5,107:8.5,108:0.5,109:1.5,110:2.5,111:3.5,112:4.5,113:5.5,114:6.5,115:7.5,116:8.5,117:0.5,118:1.5,119:2.5,120:3.5,121:4.5,122:5.5,123:6.5,124:7.5,125:8.5,126:0.5,127:1.5,128:2.5,129:3.5,130:4.5,131:5.5,132:6.5,133:7.5,134:8.5,135:0.5,136:1.5,137:2.5,138:3.5,139:4.5,140:5.5,141:6.5,142:7.5,143:8.5,144:0.5,145:1.5,146:2.5,147:3.5,148:4.5,149:5.5,150:6.5,151:7.5,152:8.5,153:0.5,154:1.5,155:2.5,156:3.5,157:4.5,158:5.5,159:6.5,160:7.5,161:8.5,162:0.5,163:1.5,164:2.5,165:3.5,166:4.5,167:5.5,168:6.5,169:7.5,170:8.5,171:0.5,172:1.5,173:2.5,174:3.5,175:4.5,176:5.5,177:6.5,178:7.5,179:8.5,180:0.5,181:1.5,182:2.5,183:3.5,184:4.5,185:5.5,186:6.5,187:7.5,188:8.5,189:0.5,190:1.5,191:2.5,192:3.5,193:4.5,194:5.5,195:6.5,196:7.5,197:8.5,198:0.5,199:1.5,200:2.5,201:3.5,202:4.5,203:5.5,204:6.5,205:7.5,206:8.5,207:0.5,208:1.5,209:2.5,210:3.5,211:4.5,212:5.5,213:6.5,214:7.5,215:8.5,216:0.5,217:1.5,218:2.5,219:3.5,220:4.5,221:5.5,222:6.5,223:7.5,224:8.5,225:0.5,226:1.5,227:2.5,228:3.5,229:4.5,230:5.5,231:6.5,232:7.5,233:8.5,234:0.5,235:1.5,236:2.5,237:3.5,238:4.5,239:5.5,240:6.5,241:7.5,242:8.5,243:0.5,244:1.5,245:2.5,246:3.5,247:4.5,248:5.5,249:6.5,250:7.5,251:8.5,252:0.5,253:1.5,254:2.5,255:3.5,256:4.5,257:5.5,258:6.5,259:7.5,260:8.5,261:0.5,262:1.5,263:2.5,264:3.5,265:4.5,266:5.5,267:6.5,268:7.5,269:8.5,270:0.5,271:1.5,272:2.5,273:3.5,274:4.5,275:5.5,276:6.5,277:7.5,278:8.5,279:0.5,280:1.5,281:2.5,282:3.5,283:4.5,284:5.5,285:6.5,286:7.5,287:8.5,288:0.5,289:1.5,290:2.5,291:3.5,292:4.5,293:5.5,294:6.5,295:7.5,296:8.5,297:0.5,298:1.5,299:2.5,300:3.5}/1000")],
                            &[T("2"), T("{}/1000")],
                            &[T("3"), Null],
                        ],
                        tag: "SELECT 3",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT id FROM t WHERE val = '{1:1.5,2:2.5,3:3.5,4:4.5,5:5.5,6:6.5,7:7.5,8:8.5,9:0.5,10:1.5,11:2.5,12:3.5,13:4.5,14:5.5,15:6.5,16:7.5,17:8.5,18:0.5,19:1.5,20:2.5,21:3.5,22:4.5,23:5.5,24:6.5,25:7.5,26:8.5,27:0.5,28:1.5,29:2.5,30:3.5,31:4.5,32:5.5,33:6.5,34:7.5,35:8.5,36:0.5,37:1.5,38:2.5,39:3.5,40:4.5,41:5.5,42:6.5,43:7.5,44:8.5,45:0.5,46:1.5,47:2.5,48:3.5,49:4.5,50:5.5,51:6.5,52:7.5,53:8.5,54:0.5,55:1.5,56:2.5,57:3.5,58:4.5,59:5.5,60:6.5,61:7.5,62:8.5,63:0.5,64:1.5,65:2.5,66:3.5,67:4.5,68:5.5,69:6.5,70:7.5,71:8.5,72:0.5,73:1.5,74:2.5,75:3.5,76:4.5,77:5.5,78:6.5,79:7.5,80:8.5,81:0.5,82:1.5,83:2.5,84:3.5,85:4.5,86:5.5,87:6.5,88:7.5,89:8.5,90:0.5,91:1.5,92:2.5,93:3.5,94:4.5,95:5.5,96:6.5,97:7.5,98:8.5,99:0.5,100:1.5,101:2.5,102:3.5,103:4.5,104:5.5,105:6.5,106:7.5,107:8.5,108:0.5,109:1.5,110:2.5,111:3.5,112:4.5,113:5.5,114:6.5,115:7.5,116:8.5,117:0.5,118:1.5,119:2.5,120:3.5,121:4.5,122:5.5,123:6.5,124:7.5,125:8.5,126:0.5,127:1.5,128:2.5,129:3.5,130:4.5,131:5.5,132:6.5,133:7.5,134:8.5,135:0.5,136:1.5,137:2.5,138:3.5,139:4.5,140:5.5,141:6.5,142:7.5,143:8.5,144:0.5,145:1.5,146:2.5,147:3.5,148:4.5,149:5.5,150:6.5,151:7.5,152:8.5,153:0.5,154:1.5,155:2.5,156:3.5,157:4.5,158:5.5,159:6.5,160:7.5,161:8.5,162:0.5,163:1.5,164:2.5,165:3.5,166:4.5,167:5.5,168:6.5,169:7.5,170:8.5,171:0.5,172:1.5,173:2.5,174:3.5,175:4.5,176:5.5,177:6.5,178:7.5,179:8.5,180:0.5,181:1.5,182:2.5,183:3.5,184:4.5,185:5.5,186:6.5,187:7.5,188:8.5,189:0.5,190:1.5,191:2.5,192:3.5,193:4.5,194:5.5,195:6.5,196:7.5,197:8.5,198:0.5,199:1.5,200:2.5,201:3.5,202:4.5,203:5.5,204:6.5,205:7.5,206:8.5,207:0.5,208:1.5,209:2.5,210:3.5,211:4.5,212:5.5,213:6.5,214:7.5,215:8.5,216:0.5,217:1.5,218:2.5,219:3.5,220:4.5,221:5.5,222:6.5,223:7.5,224:8.5,225:0.5,226:1.5,227:2.5,228:3.5,229:4.5,230:5.5,231:6.5,232:7.5,233:8.5,234:0.5,235:1.5,236:2.5,237:3.5,238:4.5,239:5.5,240:6.5,241:7.5,242:8.5,243:0.5,244:1.5,245:2.5,246:3.5,247:4.5,248:5.5,249:6.5,250:7.5,251:8.5,252:0.5,253:1.5,254:2.5,255:3.5,256:4.5,257:5.5,258:6.5,259:7.5,260:8.5,261:0.5,262:1.5,263:2.5,264:3.5,265:4.5,266:5.5,267:6.5,268:7.5,269:8.5,270:0.5,271:1.5,272:2.5,273:3.5,274:4.5,275:5.5,276:6.5,277:7.5,278:8.5,279:0.5,280:1.5,281:2.5,282:3.5,283:4.5,284:5.5,285:6.5,286:7.5,287:8.5,288:0.5,289:1.5,290:2.5,291:3.5,292:4.5,293:5.5,294:6.5,295:7.5,296:8.5,297:0.5,298:1.5,299:2.5,300:3.5}/1000';",
                    expected: Expected::Rows {
                        columns: &[Column("id", INT4)],
                        rows: &[
                            &[T("1")],
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

#[test]
fn test_pgvector_vector_type() {
    run_scripts(&[
        ScriptTest {
            name: "vector_type",
            set_up_script: &[
                "CREATE EXTENSION vector;",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[-1,-2,-3]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[-1,-2,-3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1.,2.,3.]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ' [ 1,  2 ,    3  ] '::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1.23456]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1.23456]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[hello,1]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[hello,1]""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[NaN,1]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "NaN not allowed in vector", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[Infinity,1]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "infinite value not allowed in vector", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[-Infinity,1]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "infinite value not allowed in vector", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1.5e38,-1.5e38]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1.5e+38,-1.5e+38]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1.5e+38,-1.5e+38]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1.5e+38,-1.5e+38]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1.5e-38,-1.5e-38]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1.5e-38,-1.5e-38]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[4e38,1]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""4e38" is out of range for type vector"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[-4e38,1]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: r#""-4e38" is out of range for type vector"#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1e-46,1]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[0,1]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[-1e-46,1]'::vector;",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[-0,1]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[1,2,3""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]9'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[1,2,3]9""#, detail: "Junk after closing right brace.", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '1,2,3'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "1,2,3""#, detail: r#"Vector contents must start with "["."#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT ''::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: """#, detail: r#"Vector contents must start with "["."#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '['::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[ '::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[ ""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[,'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[,""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[ ]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[,]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[,]""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[1,]""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1a]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[1a]""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,,3]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[1,,3]""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1, ,3]'::vector;",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type vector: "[1, ,3]""#, position: 8, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector(3);",
                    expected: Expected::Rows {
                        columns: &[Column("vector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector(2);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector(3, 2);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "invalid type modifier", position: 19, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector('a');",
                    expected: Expected::Error(Diagnostic { code: "22P02", message: r#"invalid input syntax for type integer: "a""#, position: 19, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector(0);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "dimensions for type vector must be at least 1", position: 19, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector(16001);",
                    expected: Expected::Error(Diagnostic { code: "22023", message: "dimensions for type vector cannot exceed 16000", position: 19, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT unnest('{"[1,2,3]", "[4,5,6]"}'::vector[]);"#,
                    expected: Expected::Rows {
                        columns: &[Column("unnest", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                            &[T("[4,5,6]")],
                        ],
                        tag: "SELECT 2",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: r#"SELECT '{"[1,2,3]"}'::vector(2)[];"#,
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector + '[4,5,6]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[5,7,9]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[3e38]'::vector + '[3e38]';",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value out of range: overflow", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector + '[3]';",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different vector dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector - '[4,5,6]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[-3,-3,-3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[-3e38]'::vector - '[3e38]';",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value out of range: overflow", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector - '[3]';",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different vector dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector * '[4,5,6]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[4,10,18]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1e37]'::vector * '[1e37]';",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value out of range: overflow", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1e-37]'::vector * '[1e-37]';",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value out of range: underflow", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector * '[3]';",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different vector dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector || '[4,5]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3,4,5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT array_fill(0, ARRAY[16000])::vector || '[1]';",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "vector cannot have more than 16000 dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector < '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector < '[1,2]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector <= '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector <= '[1,2]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector = '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector = '[1,2]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector != '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector != '[1,2]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector >= '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector >= '[1,2]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector > '[1,2,3]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("f")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2,3]'::vector > '[1,2]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", BOOL)],
                        rows: &[
                            &[T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_cmp('[1,2,3]', '[1,2,3]');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_cmp", INT4)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_cmp('[1,2,3]', '[0,0,0]');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_cmp", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_cmp('[0,0,0]', '[1,2,3]');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_cmp", INT4)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_cmp('[1,2]', '[1,2,3]');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_cmp", INT4)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_cmp('[1,2,3]', '[1,2]');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_cmp", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_cmp('[1,2]', '[2,3,4]');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_cmp", INT4)],
                        rows: &[
                            &[T("-1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_cmp('[2,3]', '[1,2,3]');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_cmp", INT4)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_dims('[1,2,3]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("vector_dims", INT4)],
                        rows: &[
                            &[T("3")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT round(vector_norm('[1,1]')::numeric, 5);",
                    expected: Expected::Rows {
                        columns: &[Column("round", NUMERIC)],
                        rows: &[
                            &[T("1.41421")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_norm('[3,4]');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_norm", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_norm('[0,1]');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_norm", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_norm('[3e37,4e37]')::real;",
                    expected: Expected::Rows {
                        columns: &[Column("vector_norm", FLOAT4)],
                        rows: &[
                            &[T("5e+37")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_norm('[0,0]');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_norm", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_norm('[2]');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_norm", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('[0,0]'::vector, '[3,4]');",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('[0,0]'::vector, '[0,1]');",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('[1,2]'::vector, '[3]');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different vector dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('[3e38]'::vector, '[-3e38]');",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_distance('[1,1,1,1,1,1,1,1,1]'::vector, '[1,1,1,1,1,1,1,4,5]');",
                    expected: Expected::Rows {
                        columns: &[Column("l2_distance", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[0,0]'::vector <-> '[3,4]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("5")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('[1,2]'::vector, '[3,4]');",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('[1,2]'::vector, '[3]');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different vector dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('[3e38]'::vector, '[3e38]');",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT inner_product('[1,1,1,1,1,1,1,1,1]'::vector, '[1,2,3,4,5,6,7,8,9]');",
                    expected: Expected::Rows {
                        columns: &[Column("inner_product", FLOAT8)],
                        rows: &[
                            &[T("45")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector <#> '[3,4]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("-11")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,2]'::vector, '[2,4]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,2]'::vector, '[0,0]')::text;",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", TEXT)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,1]'::vector, '[1,1]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,0]'::vector, '[0,2]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,1]'::vector, '[-1,-1]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,2]'::vector, '[3]');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different vector dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,1]'::vector, '[1.1,1.1]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,1]'::vector, '[-1.1,-1.1]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[3e38]'::vector, '[3e38]')::text;",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", TEXT)],
                        rows: &[
                            &[T("NaN")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,2,3,4,5,6,7,8,9]'::vector, '[1,2,3,4,5,6,7,8,9]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT cosine_distance('[1,2,3,4,5,6,7,8,9]'::vector, '[-1,-2,-3,-4,-5,-6,-7,-8,-9]');",
                    expected: Expected::Rows {
                        columns: &[Column("cosine_distance", FLOAT8)],
                        rows: &[
                            &[T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[1,2]'::vector <=> '[2,4]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('[0,0]'::vector, '[3,4]');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('[0,0]'::vector, '[0,1]');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("1")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('[1,2]'::vector, '[3]');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different vector dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('[3e38]'::vector, '[-3e38]');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("Infinity")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('[1,2,3,4,5,6,7,8,9]'::vector, '[1,2,3,4,5,6,7,8,9]');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("0")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l1_distance('[1,2,3,4,5,6,7,8,9]'::vector, '[0,3,2,5,4,7,6,9,8]');",
                    expected: Expected::Rows {
                        columns: &[Column("l1_distance", FLOAT8)],
                        rows: &[
                            &[T("9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT '[0,0]'::vector <+> '[3,4]';",
                    expected: Expected::Rows {
                        columns: &[Column("?column?", FLOAT8)],
                        rows: &[
                            &[T("7")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('[3,4]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("[0.6,0.8]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('[3,0]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("[1,0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('[0,0.1]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("[0,1]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('[0,0]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("[0,0]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT l2_normalize('[3e38]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("l2_normalize", USER_DEFINED)],
                        rows: &[
                            &[T("[1]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT binary_quantize('[1,0,-1]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("binary_quantize", BIT)],
                        rows: &[
                            &[T("100")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT binary_quantize('[0,0.1,-0.2,-0.3,0.4,0.5,0.6,-0.7,0.8,-0.9,1]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("binary_quantize", BIT)],
                        rows: &[
                            &[T("01001110101")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT binary_quantize('[1,2,3,-4,5,6,-7,8,1,-2,-3,4,5,-6,7,8,-1,2,3]'::vector);",
                    expected: Expected::Rows {
                        columns: &[Column("binary_quantize", BIT)],
                        rows: &[
                            &[T("1110110110011011011")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::vector, 1, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::vector, 3, 2);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[3,4]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::vector, -1, 3);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[1]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::vector, 3, 9);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[3,4,5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::vector, 1, 0);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::vector, 3, -1);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::vector, -1, 2);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::vector, 2147483647, 10);",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::vector, 3, 2147483647);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[3,4,5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT subvector('[1,2,3,4,5]'::vector, -2147483644, 2147483647);",
                    expected: Expected::Rows {
                        columns: &[Column("subvector", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v) FROM unnest(ARRAY['[1,2,3]'::vector, '[3,5,7]']) v;",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED)],
                        rows: &[
                            &[T("[2,3.5,5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v) FROM unnest(ARRAY['[1,2,3]'::vector, '[3,5,7]', NULL]) v;",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED)],
                        rows: &[
                            &[T("[2,3.5,5]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v) FROM unnest(ARRAY[]::vector[]) v;",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v) FROM unnest(ARRAY['[1,2]'::vector, '[3]']) v;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 2 dimensions, not 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT avg(v) FROM unnest(ARRAY['[3e38]'::vector, '[3e38]']) v;",
                    expected: Expected::Rows {
                        columns: &[Column("avg", USER_DEFINED)],
                        rows: &[
                            &[T("[3e+38]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_avg('{2,2,4,6}');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_avg", USER_DEFINED)],
                        rows: &[
                            &[T("[1,2,3]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_avg('{0}');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_avg", USER_DEFINED)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_avg('{1}');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_avg('{{2,2,4,6}}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "vector_avg: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_avg('{NULL,2,4,6}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "vector_avg: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_avg('{}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "vector_avg: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_avg(array_agg(n)) FROM generate_series(1, 16002) n;",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "vector cannot have more than 16000 dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_accum('{0}', '[1,2,3]');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_accum", FLOAT8_ARRAY)],
                        rows: &[
                            &[T("{1,1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_accum('{0,0,0,0}', '[1,2,3]');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_accum", FLOAT8_ARRAY)],
                        rows: &[
                            &[T("{1,1,2,3}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_accum('{{0}}', '[1,2,3]');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "vector_accum: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_accum('{NULL}', '[1,2,3]');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "vector_accum: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_accum('{}', '[1,2,3]');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "vector_accum: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_accum('{0,0}', '[1,2,3]');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "expected 1 dimensions, not 3", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_combine('{1,2}', '{3,4}');",
                    expected: Expected::Rows {
                        columns: &[Column("vector_combine", FLOAT8_ARRAY)],
                        rows: &[
                            &[T("{4,6}")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_combine('{{1,2}}', '{3,4}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "vector_combine: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_combine('{1,2}', '{{3,4}}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "vector_combine: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_combine('{NULL,2}', '{3,4}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "vector_combine: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_combine('{1,2}', '{3,NULL}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "vector_combine: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_combine('{}', '{0}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "vector_combine: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_combine('{0}', '{}');",
                    expected: Expected::Error(Diagnostic { code: "XX000", message: "vector_combine: expected state array", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_combine('{0}', '{0}');",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "vector must have at least 1 dimension", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_combine('{0}', (SELECT array_agg(n) FROM generate_series(1, 16002) n));",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "vector cannot have more than 16000 dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_combine((SELECT array_agg(n) FROM generate_series(1, 16002) n), '{0}');",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "vector cannot have more than 16000 dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT vector_combine((SELECT array_agg(n) FROM generate_series(1, 16002) n), (SELECT array_agg(n) FROM generate_series(1, 16002) n));",
                    expected: Expected::Error(Diagnostic { code: "54000", message: "vector cannot have more than 16000 dimensions", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(v) FROM unnest(ARRAY['[1,2,3]'::vector, '[3,5,7]']) v;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", USER_DEFINED)],
                        rows: &[
                            &[T("[4,7,10]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(v) FROM unnest(ARRAY['[1,2,3]'::vector, '[3,5,7]', NULL]) v;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", USER_DEFINED)],
                        rows: &[
                            &[T("[4,7,10]")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(v) FROM unnest(ARRAY[]::vector[]) v;",
                    expected: Expected::Rows {
                        columns: &[Column("sum", USER_DEFINED)],
                        rows: &[
                            &[Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(v) FROM unnest(ARRAY['[1,2]'::vector, '[3]']) v;",
                    expected: Expected::Error(Diagnostic { code: "22000", message: "different vector dimensions 2 and 1", ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT sum(v) FROM unnest(ARRAY['[3e38]'::vector, '[3e38]']) v;",
                    expected: Expected::Error(Diagnostic { code: "22003", message: "value out of range: overflow", ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
