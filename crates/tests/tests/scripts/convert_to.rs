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
fn test_convert_to() {
    run_scripts(&[
        ScriptTest {
            name: "convert_to",
            set_up_script: &[
                "CREATE TABLE supported_encodings (name TEXT PRIMARY KEY);",
                "INSERT INTO supported_encodings VALUES ('SQL_ASCII'),('SQLASCII'),('EUC_JP'),('EUC_KR'),('UTF8'),('utf-8'),('UNICODE'),('LATIN1'),('ISO_8859_1'),('LATIN2'),('ISO_8859_2'),('LATIN3'),('ISO_8859_3'),('LATIN4'),('ISO_8859_4'),('LATIN5'),('ISO_8859_9'),('LATIN6'),('ISO_8859_10'),('LATIN7'),('ISO_8859_13'),('LATIN8'),('ISO_8859_14'),('LATIN9'),('ISO_8859_15'),('LATIN10'),('ISO_8859_16'),('WIN1256'),('WIN1258'),('WIN866'),('ALT'),('WIN874'),('KOI8R'),('WIN1251'),('WIN'),('WIN1252'),('ISO_8859_5'),('ISO_8859_6'),('ISO_8859_7'),('ISO_8859_8'),('WIN1250'),('WIN1253'),('WIN1254'),('WIN1255'),('WIN1257'),('KOI8U'),('SJIS'),('SHIFT_JIS'),('BIG5'),('GBK'),('UHC'),('GB18030'),('WINDOWS-866'),('WINDOWS-874'),('WINDOWS-1250'),('WINDOWS-1251'),('WINDOWS-1252'),('WINDOWS-1253'),('WINDOWS-1254'),('WINDOWS-1255'),('WINDOWS-1256'),('WINDOWS-1257'),('WINDOWS-1258');",
            ],
            assertions: &[
                ScriptTestAssertion {
                    query: "SELECT count(*), bool_and(pg_char_to_encoding(name) >= 0), bool_and(encode(convert_to('x', name::name), 'hex') = '78') FROM supported_encodings;",
                    expected: Expected::Rows {
                        columns: &[Column("count", INT8), Column("bool_and", BOOL), Column("bool_and", BOOL)],
                        rows: &[
                            &[T("63"), T("t"), T("t")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode(convert_to('x', 'UTF8'), 'hex');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T("78")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode(convert_to('café', 'utf-8'), 'hex');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T("636166c3a9")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode(convert_to('日本', 'EUC_JP'), 'hex');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T("c6fccbdc")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode(convert_to('한', 'UHC'), 'hex'), encode(convert_to('Ж', 'ALT'), 'hex'), encode(convert_to('Ж', 'WIN'), 'hex'), encode(convert_to('Ж', 'ISO_8859_5'), 'hex');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT), Column("encode", TEXT), Column("encode", TEXT), Column("encode", TEXT)],
                        rows: &[
                            &[T("c7d1"), T("86"), T("c6"), T("b6")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode(convert_to('Ж', 'KOI8R'), 'hex'), encode(convert_to('日本', 'SJIS'), 'hex'), encode(convert_to('한', 'EUC_KR'), 'hex'), encode(convert_to('한', 'UHC'), 'hex'), encode(convert_to('中文', 'BIG5'), 'hex'), encode(convert_to('中文', 'GBK'), 'hex'), encode(convert_to('😄', 'GB18030'), 'hex');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT), Column("encode", TEXT), Column("encode", TEXT), Column("encode", TEXT), Column("encode", TEXT), Column("encode", TEXT), Column("encode", TEXT)],
                        rows: &[
                            &[T("f6"), T("93fa967b"), T("c7d1"), T("c7d1"), T("a4a4a4e5"), T("d6d0cec4"), T("9439fd30")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode(convert_to('€', 'WIN1252'), 'hex'), encode(convert_to('€', 'WINDOWS-1252'), 'hex'), encode(convert_to('Ж', 'WIN866'), 'hex'), encode(convert_to('Ж', 'WINDOWS-866'), 'hex');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT), Column("encode", TEXT), Column("encode", TEXT), Column("encode", TEXT)],
                        rows: &[
                            &[T("80"), T("80"), T("86"), T("86")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode(convert_to('café', 'LATIN1'), 'hex'), encode(convert_to('café', 'ISO_8859_1'), 'hex'), encode(convert_to('日本', 'SJIS'), 'hex'), encode(convert_to('日本', 'SHIFT_JIS'), 'hex');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT), Column("encode", TEXT), Column("encode", TEXT), Column("encode", TEXT)],
                        rows: &[
                            &[T("636166e9"), T("636166e9"), T("93fa967b"), T("93fa967b")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_char_to_encoding('ALT'), pg_char_to_encoding('WIN'), pg_char_to_encoding('UHC'), pg_char_to_encoding('ISO_8859_5'), pg_char_to_encoding('EUC_CN');",
                    expected: Expected::Rows {
                        columns: &[Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4)],
                        rows: &[
                            &[T("20"), T("23"), T("38"), T("25"), T("2")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT pg_char_to_encoding('WINDOWS-866'), pg_char_to_encoding('WINDOWS-874'), pg_char_to_encoding('WINDOWS-1250'), pg_char_to_encoding('WINDOWS-1251'), pg_char_to_encoding('WINDOWS-1252'), pg_char_to_encoding('WINDOWS-1253'), pg_char_to_encoding('WINDOWS-1254'), pg_char_to_encoding('WINDOWS-1255'), pg_char_to_encoding('WINDOWS-1256'), pg_char_to_encoding('WINDOWS-1257'), pg_char_to_encoding('WINDOWS-1258');",
                    expected: Expected::Rows {
                        columns: &[Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4), Column("pg_char_to_encoding", INT4)],
                        rows: &[
                            &[T("20"), T("21"), T("29"), T("23"), T("24"), T("30"), T("31"), T("32"), T("18"), T("33"), T("19")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT encode(convert_to('', 'UTF8'), 'hex');",
                    expected: Expected::Rows {
                        columns: &[Column("encode", TEXT)],
                        rows: &[
                            &[T("")],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT convert_to(NULL, 'UTF8'), convert_to('x', NULL);",
                    expected: Expected::Rows {
                        columns: &[Column("convert_to", BYTEA), Column("convert_to", BYTEA)],
                        rows: &[
                            &[Null, Null],
                        ],
                        tag: "SELECT 1",
                    },
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT convert_to('€', 'LATIN1');",
                    expected: Expected::Error(Diagnostic { code: "22P05", message: r#"character with byte sequence 0xe2 0x82 0xac in encoding "UTF8" has no equivalent in encoding "LATIN1""#, ..E }),
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT convert_to('x', 'EUC_CN');",
                    expected: Expected::Tag("SELECT 1"),
                    flow: Flow::Exec,
                    ..A
                },
                ScriptTestAssertion {
                    query: "SELECT convert_to('x', 'bogus');",
                    expected: Expected::Error(Diagnostic { code: "22023", message: r#"invalid destination encoding name "bogus""#, ..E }),
                    ..A
                },
            ],
            ..S
        },
    ]);
}
