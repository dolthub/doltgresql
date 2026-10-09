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
fn test_issues_wire() {
    run_wire_tests(&[
        WireTest {
            name: "Issue #2546: string literal described as invalid OID 705",
            steps: &[
                Step::Send(&[
                    Send::Query("SELECT 'foo';"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "?column?", attnum: 0, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("foo")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "Issue #3097: bind parameters described as unknown OID, breaking pgx",
            set_up_script: &[
                "CREATE TABLE g_arr (id INT4 PRIMARY KEY, v INT4, vals INT4[]);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "s1", query: "SELECT count(*) FROM g_arr WHERE $1 = ANY(vals)", parameter_oids: &[] },
                    Send::Describe(b'S', "s1"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[INT4]),
                    Receive::RowDescription(&[Field { name: "count", attnum: 0, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "s2", query: "SELECT count(*) FROM g_arr WHERE $1 = ANY(SELECT v FROM g_arr)", parameter_oids: &[] },
                    Send::Describe(b'S', "s2"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[INT4]),
                    Receive::RowDescription(&[Field { name: "count", attnum: 0, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "s3", query: "SELECT count(*) FROM g_arr WHERE vals[$1] = 2", parameter_oids: &[] },
                    Send::Describe(b'S', "s3"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[INT4]),
                    Receive::RowDescription(&[Field { name: "count", attnum: 0, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "s4", query: "SELECT count(*) FROM g_arr WHERE coalesce($1, v) = 1", parameter_oids: &[] },
                    Send::Describe(b'S', "s4"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[INT4]),
                    Receive::RowDescription(&[Field { name: "count", attnum: 0, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "s5", query: "SELECT count(*) FROM g_arr WHERE greatest($1, v) = 1", parameter_oids: &[] },
                    Send::Describe(b'S', "s5"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[INT4]),
                    Receive::RowDescription(&[Field { name: "count", attnum: 0, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "Issue #2557: unescaped newlines in JSON output",
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: r#"SELECT '{"v":"a\\nb"}'::jsonb;"#, parameter_oids: &[] },
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#"{"v": "a\\nb"}"#)]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: r#"SELECT $${"v":"a\\nb"}$$::jsonb;"#, parameter_oids: &[] },
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#"{"v": "a\\nb"}"#)]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name3", query: r#"SELECT $${"v":"a\\\nb"}$$::jsonb;"#, parameter_oids: &[] },
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name3", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#"{"v": "a\\\nb"}"#)]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name4", query: r#"select json '{ "a":  "dollar \\u0024 character" }' ->> 'a' as not_an_escape;"#, parameter_oids: &[] },
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name4", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#"dollar \u0024 character"#)]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "Issue #3111 (set-returning functions and max1Row spooling)",
            set_up_script: &[
                "CREATE TABLE bug14 (a integer, b integer);",
                "CREATE INDEX bug14_ab ON bug14 (a, b);",
                "CREATE TABLE arrtbl (pk integer PRIMARY KEY, arr integer[]);",
                "INSERT INTO arrtbl VALUES (1, '{10,20,30}'), (2, '{40}');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("SELECT unnest(indkey) FROM pg_index WHERE indexrelid = 'bug14_ab'::regclass;"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "unnest", attnum: 0, type_oid: INT2, size: 2, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("1")]),
                    Receive::DataRow(&[Datum::Text("2")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Query("SELECT unnest(arr) FROM arrtbl WHERE pk = 1;"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "unnest", attnum: 0, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("10")]),
                    Receive::DataRow(&[Datum::Text("20")]),
                    Receive::DataRow(&[Datum::Text("30")]),
                    Receive::CommandComplete("SELECT 3"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_result_batch_mixed_formats() {
    run_wire_tests(&[
        WireTest {
            name: "mixed formats across result batch boundary",
            set_up_script: &[
                "CREATE TABLE result_batch_mixed_test (id INT PRIMARY KEY, value TEXT)",
                "INSERT INTO result_batch_mixed_test VALUES (1, 'value-001'),(2, 'value-002'),(3, 'value-003'),(4, 'value-004'),(5, 'value-005'),(6, 'value-006'),(7, 'value-007'),(8, 'value-008'),(9, 'value-009'),(10, 'value-010'),(11, 'value-011'),(12, 'value-012'),(13, 'value-013'),(14, 'value-014'),(15, 'value-015'),(16, 'value-016'),(17, NULL),(18, 'value-018'),(19, 'value-019'),(20, 'value-020'),(21, 'value-021'),(22, 'value-022'),(23, 'value-023'),(24, 'value-024'),(25, 'value-025'),(26, 'value-026'),(27, 'value-027'),(28, 'value-028'),(29, 'value-029'),(30, 'value-030'),(31, 'value-031'),(32, 'value-032'),(33, 'value-033'),(34, NULL),(35, 'value-035'),(36, 'value-036'),(37, 'value-037'),(38, 'value-038'),(39, 'value-039'),(40, 'value-040'),(41, 'value-041'),(42, 'value-042'),(43, 'value-043'),(44, 'value-044'),(45, 'value-045'),(46, 'value-046'),(47, 'value-047'),(48, 'value-048'),(49, 'value-049'),(50, 'value-050'),(51, NULL),(52, 'value-052'),(53, 'value-053'),(54, 'value-054'),(55, 'value-055'),(56, 'value-056'),(57, 'value-057'),(58, 'value-058'),(59, 'value-059'),(60, 'value-060'),(61, 'value-061'),(62, 'value-062'),(63, 'value-063'),(64, 'value-064'),(65, 'value-065'),(66, 'value-066'),(67, 'value-067'),(68, NULL),(69, 'value-069'),(70, 'value-070'),(71, 'value-071'),(72, 'value-072'),(73, 'value-073'),(74, 'value-074'),(75, 'value-075'),(76, 'value-076'),(77, 'value-077'),(78, 'value-078'),(79, 'value-079'),(80, 'value-080'),(81, 'value-081'),(82, 'value-082'),(83, 'value-083'),(84, 'value-084'),(85, NULL),(86, 'value-086'),(87, 'value-087'),(88, 'value-088'),(89, 'value-089'),(90, 'value-090'),(91, 'value-091'),(92, 'value-092'),(93, 'value-093'),(94, 'value-094'),(95, 'value-095'),(96, 'value-096'),(97, 'value-097'),(98, 'value-098'),(99, 'value-099'),(100, 'value-100'),(101, 'value-101'),(102, NULL),(103, 'value-103'),(104, 'value-104'),(105, 'value-105'),(106, 'value-106'),(107, 'value-107'),(108, 'value-108'),(109, 'value-109'),(110, 'value-110'),(111, 'value-111'),(112, 'value-112'),(113, 'value-113'),(114, 'value-114'),(115, 'value-115'),(116, 'value-116'),(117, 'value-117'),(118, 'value-118'),(119, NULL),(120, 'value-120'),(121, 'value-121'),(122, 'value-122'),(123, 'value-123'),(124, 'value-124'),(125, 'value-125'),(126, 'value-126'),(127, 'value-127'),(128, 'value-128'),(129, NULL)",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "result_batch_mixed", query: "SELECT id, value FROM result_batch_mixed_test ORDER BY id", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "result_batch_mixed", parameter_formats: &[], parameters: &[], result_formats: &[1, 0] },
                    Send::Execute("", 0),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{1}"), Datum::Text("value-001")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{2}"), Datum::Text("value-002")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{3}"), Datum::Text("value-003")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{4}"), Datum::Text("value-004")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{5}"), Datum::Text("value-005")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{6}"), Datum::Text("value-006")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{7}"), Datum::Text("value-007")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{8}"), Datum::Text("value-008")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\t"), Datum::Text("value-009")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\n"), Datum::Text("value-010")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{b}"), Datum::Text("value-011")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{c}"), Datum::Text("value-012")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\r"), Datum::Text("value-013")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{e}"), Datum::Text("value-014")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{f}"), Datum::Text("value-015")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{10}"), Datum::Text("value-016")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{11}"), Datum::Null]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{12}"), Datum::Text("value-018")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{13}"), Datum::Text("value-019")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{14}"), Datum::Text("value-020")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{15}"), Datum::Text("value-021")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{16}"), Datum::Text("value-022")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{17}"), Datum::Text("value-023")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{18}"), Datum::Text("value-024")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{19}"), Datum::Text("value-025")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{1a}"), Datum::Text("value-026")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{1b}"), Datum::Text("value-027")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{1c}"), Datum::Text("value-028")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{1d}"), Datum::Text("value-029")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{1e}"), Datum::Text("value-030")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{1f}"), Datum::Text("value-031")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0 "), Datum::Text("value-032")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0!"), Datum::Text("value-033")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\""), Datum::Null]),
                    Receive::DataRow(&[Datum::Text("\0\0\0#"), Datum::Text("value-035")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0$"), Datum::Text("value-036")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0%"), Datum::Text("value-037")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0&"), Datum::Text("value-038")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0'"), Datum::Text("value-039")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0("), Datum::Text("value-040")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0)"), Datum::Text("value-041")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0*"), Datum::Text("value-042")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0+"), Datum::Text("value-043")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0,"), Datum::Text("value-044")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0-"), Datum::Text("value-045")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0."), Datum::Text("value-046")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0/"), Datum::Text("value-047")]),
                    Receive::DataRow(&[Datum::Text("\0\0\00"), Datum::Text("value-048")]),
                    Receive::DataRow(&[Datum::Text("\0\0\01"), Datum::Text("value-049")]),
                    Receive::DataRow(&[Datum::Text("\0\0\02"), Datum::Text("value-050")]),
                    Receive::DataRow(&[Datum::Text("\0\0\03"), Datum::Null]),
                    Receive::DataRow(&[Datum::Text("\0\0\04"), Datum::Text("value-052")]),
                    Receive::DataRow(&[Datum::Text("\0\0\05"), Datum::Text("value-053")]),
                    Receive::DataRow(&[Datum::Text("\0\0\06"), Datum::Text("value-054")]),
                    Receive::DataRow(&[Datum::Text("\0\0\07"), Datum::Text("value-055")]),
                    Receive::DataRow(&[Datum::Text("\0\0\08"), Datum::Text("value-056")]),
                    Receive::DataRow(&[Datum::Text("\0\0\09"), Datum::Text("value-057")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0:"), Datum::Text("value-058")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0;"), Datum::Text("value-059")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0<"), Datum::Text("value-060")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0="), Datum::Text("value-061")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0>"), Datum::Text("value-062")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0?"), Datum::Text("value-063")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0@"), Datum::Text("value-064")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0A"), Datum::Text("value-065")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0B"), Datum::Text("value-066")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0C"), Datum::Text("value-067")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0D"), Datum::Null]),
                    Receive::DataRow(&[Datum::Text("\0\0\0E"), Datum::Text("value-069")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0F"), Datum::Text("value-070")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0G"), Datum::Text("value-071")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0H"), Datum::Text("value-072")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0I"), Datum::Text("value-073")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0J"), Datum::Text("value-074")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0K"), Datum::Text("value-075")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0L"), Datum::Text("value-076")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0M"), Datum::Text("value-077")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0N"), Datum::Text("value-078")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0O"), Datum::Text("value-079")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0P"), Datum::Text("value-080")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0Q"), Datum::Text("value-081")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0R"), Datum::Text("value-082")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0S"), Datum::Text("value-083")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0T"), Datum::Text("value-084")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0U"), Datum::Null]),
                    Receive::DataRow(&[Datum::Text("\0\0\0V"), Datum::Text("value-086")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0W"), Datum::Text("value-087")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0X"), Datum::Text("value-088")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0Y"), Datum::Text("value-089")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0Z"), Datum::Text("value-090")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0["), Datum::Text("value-091")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\\"), Datum::Text("value-092")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0]"), Datum::Text("value-093")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0^"), Datum::Text("value-094")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0_"), Datum::Text("value-095")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0`"), Datum::Text("value-096")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0a"), Datum::Text("value-097")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0b"), Datum::Text("value-098")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0c"), Datum::Text("value-099")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0d"), Datum::Text("value-100")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0e"), Datum::Text("value-101")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0f"), Datum::Null]),
                    Receive::DataRow(&[Datum::Text("\0\0\0g"), Datum::Text("value-103")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0h"), Datum::Text("value-104")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0i"), Datum::Text("value-105")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0j"), Datum::Text("value-106")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0k"), Datum::Text("value-107")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0l"), Datum::Text("value-108")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0m"), Datum::Text("value-109")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0n"), Datum::Text("value-110")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0o"), Datum::Text("value-111")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0p"), Datum::Text("value-112")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0q"), Datum::Text("value-113")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0r"), Datum::Text("value-114")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0s"), Datum::Text("value-115")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0t"), Datum::Text("value-116")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0u"), Datum::Text("value-117")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0v"), Datum::Text("value-118")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0w"), Datum::Null]),
                    Receive::DataRow(&[Datum::Text("\0\0\0x"), Datum::Text("value-120")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0y"), Datum::Text("value-121")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0z"), Datum::Text("value-122")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0{"), Datum::Text("value-123")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0|"), Datum::Text("value-124")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0}"), Datum::Text("value-125")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0~"), Datum::Text("value-126")]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{7f}"), Datum::Text("value-127")]),
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 128]), Datum::Text("value-128")]),
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 129]), Datum::Null]),
                    Receive::CommandComplete("SELECT 129"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_wire_types_receiving() {
    run_wire_tests(&[
        WireTest {
            name: "BIT receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 BIT(65), v2 BIT(3));",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1, 1], parameters: &[Datum::Bytes(&[0, 0, 0, 65, 170, 35, 108, 179, 21, 106, 150, 172, 128]), Datum::Bytes(&[0, 0, 0, 3, 160])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("10101010001000110110110010110011000101010110101010010110101011001"), Datum::Text("101")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "BIT VARYING receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 BIT VARYING, v2 BIT VARYING(5));",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[0, 0, 0, 18, 149, 102, 64]), Datum::Bytes(&[0, 0, 0, 3, 192])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("100101010110011001"), Datum::Text("110")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "BOOL receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 BOOL, v2 BOOL);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text("\u{1}"), Datum::Text("\0")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("t"), Datum::Text("f")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "BYTEA receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 BYTEA, v2 BYTEA);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[222, 173, 190, 239]), Datum::Bytes(&[192, 255, 238])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#"\xdeadbeef"#), Datum::Text(r#"\xc0ffee"#)]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "CID receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 CID, v2 CID);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text("\0\0\0\u{6}"), Datum::Text("\0\0\0\u{7}")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("6"), Datum::Text("7")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: r#""char" receiving binary format"#,
            set_up_script: &[
                r#"CREATE TABLE test (v1 "char", v2 "char");"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text("1"), Datum::Text("v")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("1"), Datum::Text("v")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "DATE receiving binary format",
            set_up_script: &[
                "SET datestyle TO 'ISO, YMD';",
                "CREATE TABLE test (v1 DATE, v2 DATE);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[255, 255, 254, 154]), Datum::Text("\0\0$\u{16}")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[255, 255, 254, 154]), Datum::Text("\0\0$\u{16}")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "ENUM receiving binary format",
            set_up_script: &[
                "CREATE TYPE enumType AS ENUM ('eval1', 'eval2', 'eval3');",
                "CREATE TABLE test (v1 enumType, v2 enumType);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text("eval1"), Datum::Text("eval3")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("eval1"), Datum::Text("eval3")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "FLOAT4 receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 FLOAT4, v2 FLOAT4);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[191, 0, 0, 0]), Datum::Bytes(&[65, 208, 32, 0])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("-0.5"), Datum::Text("26.015625")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "FLOAT8 receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 FLOAT8, v2 FLOAT8);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[191, 224, 0, 0, 0, 0, 0, 0]), Datum::Text("@:\u{4}\0\0\0\0\0")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("-0.5"), Datum::Text("26.015625")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INT2 receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 INT2, v2 INT2);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text("\0\u{3}"), Datum::Text("1f")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("3"), Datum::Text("12646")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INT2VECTOR receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 INT2VECTOR, v2 INT2VECTOR);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text("\0\0\0\u{1}\0\0\0\0\0\0\0\u{15}\0\0\0\u{4}\0\0\0\0\0\0\0\u{2}\0\u{1}\0\0\0\u{2}\0\u{2}\0\0\0\u{2}\0\u{4}\0\0\0\u{2}\0\u{5}"), Datum::Bytes(&[0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 21, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 2, 0, 5, 0, 0, 0, 2, 0, 87, 0, 0, 0, 2, 3, 223])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("1 2 4 5"), Datum::Text("5 87 991")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INT4 receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 INT4, v2 INT4);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[255, 255, 255, 251]), Datum::Bytes(&[0, 54, 154, 89])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("-5"), Datum::Text("3578457")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INT8 receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 INT8, v2 INT8);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[255, 255, 255, 255, 255, 255, 255, 212]), Datum::Bytes(&[0, 0, 2, 88, 88, 7, 187, 113])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("-44"), Datum::Text("2578457279345")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INTERVAL receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 INTERVAL, v2 INTERVAL);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[0, 0, 0, 0, 3, 147, 135, 0, 0, 0, 0, 0, 0, 0, 0, 0]), Datum::Bytes(&[0, 0, 0, 111, 185, 177, 134, 8, 0, 0, 2, 188, 0, 0, 0, 39])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 0, 3, 147, 135, 0, 0, 0, 0, 0, 0, 0, 0, 0]), Datum::Bytes(&[0, 0, 0, 111, 185, 177, 134, 8, 0, 0, 2, 188, 0, 0, 0, 39])]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "JSON receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 JSON, v2 JSON);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text(r#"{"key1": {"key": "value"}}"#), Datum::Text("{}")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#"{"key1": {"key": "value"}}"#), Datum::Text("{}")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "JSONB receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 JSONB, v2 JSONB);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text("\u{1}{\"key1\": {\"key\": [2, 3]}}"), Datum::Text("\u{1}[]")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#"{"key1": {"key": [2, 3]}}"#), Datum::Text("[]")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "NAME receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 NAME, v2 NAME);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text(""), Datum::Text("abc")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(""), Datum::Text("abc")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "NUMERIC receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 NUMERIC, v2 NUMERIC(5,2), v3 NUMERIC(14,5));",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2, $3);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[0, 8, 0, 1, 0, 0, 0, 21, 4, 211, 28, 64, 17, 215, 33, 217, 12, 152, 30, 7, 21, 203, 35, 40]), Datum::Bytes(&[0, 2, 0, 0, 0, 0, 0, 2, 0, 235, 26, 44]), Datum::Bytes(&[0, 2, 0, 0, 0, 0, 0, 5, 16, 182, 0, 90])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("12357232.456786653224768755799"), Datum::Text("235.67"), Datum::Text("4278.00900")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "OID receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 OID, v2 OID);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text("\0\0\0\u{1}"), Datum::Bytes(&[148, 8, 88, 129])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("1"), Datum::Text("2483574913")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "OIDVECTOR receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 OIDVECTOR, v2 OIDVECTOR);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 26, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 4, 210, 0, 0, 0, 4, 0, 0, 9, 185]), Datum::Bytes(&[0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 26, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 9, 179, 0, 0, 0, 4, 0, 0, 2, 62, 0, 0, 0, 4, 0, 0, 3, 145])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("1234 2489"), Datum::Text("2483 574 913")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "RECORD receiving binary format",
            set_up_script: &[
                "CREATE TABLE pre1 (v1 TEXT, v2 INT8, v3 NUMERIC(6,1));",
                "CREATE TABLE pre2 (v1 VARCHAR, v2 OID, v3 BOOL);",
                "CREATE TABLE test (v1 pre1, v2 pre2);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[0, 0, 0, 3, 0, 0, 0, 25, 0, 0, 0, 3, 97, 98, 99, 0, 0, 0, 20, 0, 0, 0, 8, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 6, 164, 0, 0, 0, 14, 0, 3, 0, 1, 0, 0, 0, 1, 0, 1, 9, 41, 23, 112]), Datum::Text("\0\0\0\u{3}\0\0\u{4}\u{13}\0\0\0\u{3}def\0\0\0\u{1a}\0\0\0\u{4}\0\0\0\u{2}\0\0\0\u{10}\0\0\0\u{1}\u{1}")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("(abc,1,12345.6)"), Datum::Text("(def,2,t)")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "REGTYPE receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 REGTYPE, v2 REGTYPE);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[0, 0, 6, 164]), Datum::Text("\0\0\0\u{19}")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("numeric"), Datum::Text("text")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TEXT receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 TEXT, v2 TEXT);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text(""), Datum::Text("abc")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(""), Datum::Text("abc")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TEXT ARRAY receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 TEXT[], v2 TEXT[]);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text("\0\0\0\0\0\0\0\0\0\0\0\u{19}"), Datum::Text("\0\0\0\u{1}\0\0\0\0\0\0\0\u{19}\0\0\0\u{3}\0\0\0\u{1}\0\0\0\u{1}a\0\0\0\u{2}bb\0\0\0\u{3}ccc")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("{}"), Datum::Text("{a,bb,ccc}")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TID receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 TID, v2 TID);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text("\0\0\0\u{c}\0\""), Datum::Text("\0\0\08\0N")], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("(12,34)"), Datum::Text("(56,78)")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TIME receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 TIME, v2 TIME);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[0, 0, 0, 17, 250, 171, 177, 0]), Datum::Bytes(&[0, 0, 0, 10, 57, 214, 4, 0])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("21:27:00"), Datum::Text("12:12:00")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TIMETZ receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 TIMETZ, v2 TIMETZ);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[0, 0, 0, 17, 250, 171, 177, 0, 0, 0, 112, 128]), Datum::Bytes(&[0, 0, 0, 10, 57, 214, 4, 0, 0, 0, 112, 128])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("21:27:00-08"), Datum::Text("12:12:00-08")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TIMESTAMP receiving binary format",
            set_up_script: &[
                "SET datestyle TO 'Postgres, MDY';",
                "CREATE TABLE test (v1 TIMESTAMP, v2 TIMESTAMP);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[0, 2, 62, 228, 207, 3, 128, 0]), Datum::Bytes(&[0, 2, 94, 46, 160, 114, 138, 136])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("Sun Jan 12 00:00:00 2020"), Datum::Text("Sat Feb 13 04:05:06.789 2021")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TIMESTAMPTZ receiving binary format",
            set_up_script: &[
                "SET datestyle TO 'Postgres, MDY';",
                "CREATE TABLE test (v1 TIMESTAMPTZ, v2 TIMESTAMPTZ);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[0, 2, 62, 235, 131, 160, 160, 0]), Datum::Bytes(&[0, 2, 94, 52, 126, 124, 6, 136])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("Sun Jan 12 00:00:00 2020 PST"), Datum::Text("Sat Feb 13 03:05:06.789 2021 PST")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "UUID receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 UUID, v2 UUID);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[253, 171, 240, 61, 155, 33, 69, 49, 185, 0, 198, 246, 207, 248, 56, 108]), Datum::Bytes(&[7, 48, 121, 28, 192, 221, 73, 114, 158, 114, 17, 234, 217, 49, 122, 90])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("fdabf03d-9b21-4531-b900-c6f6cff8386c"), Datum::Text("0730791c-c0dd-4972-9e72-11ead9317a5a")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "VARCHAR receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 VARCHAR, v2 VARCHAR(5));",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text(""), Datum::Text(r#"a",c"#)], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(""), Datum::Text(r#"a",c"#)]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "XID receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 XID, v2 XID);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text("\0\0\0\u{1}"), Datum::Bytes(&[148, 8, 88, 129])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("1"), Datum::Text("2483574913")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "XML receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 XML, v2 XML);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Text("<a>x</a>"), Datum::Text(r#"<?xml version="1.0" encoding="UTF-8"?><b c="1">y&amp;z</b>"#)], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("<a>x</a>"), Datum::Text(r#"<b c="1">y&amp;z</b>"#)]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "XML ARRAY receiving binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 XML[], v2 XML[]);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name1", query: "INSERT INTO test VALUES ($1, $2);", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name1", parameter_formats: &[1], parameters: &[Datum::Bytes(&[0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 142, 0, 0, 0, 2, 0, 0, 0, 1, 0, 0, 0, 4, 60, 97, 47, 62, 0, 0, 0, 8, 60, 98, 62, 120, 60, 47, 98, 62]), Datum::Bytes(&[0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 142, 0, 0, 0, 2, 0, 0, 0, 1, 255, 255, 255, 255, 0, 0, 0, 4, 60, 99, 47, 62])], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::CommandComplete("INSERT 0 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Parse { name: "stmt_name2", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Bind { portal: "", statement: "stmt_name2", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("{<a/>,<b>x</b>}"), Datum::Text("{NULL,<c/>}")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}

#[test]
fn test_wire_types_sending() {
    run_wire_tests(&[
        WireTest {
            name: "Smoke Test",
            set_up_script: &[
                "CREATE TABLE test (pk INT4 PRIMARY KEY);",
                "INSERT INTO test VALUES (7);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Query("SELECT * FROM test;"),
                ]),
                Step::Receive(&[
                    Receive::RowDescription(&[Field { name: "pk", attnum: 1, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::DataRow(&[Datum::Text("7")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "BIT returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 BIT(8), v2 BIT(3));",
                "INSERT INTO test VALUES (B'11011010', '101');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: BIT, size: -1, typmod: 8, format: 0 }, Field { name: "v2", attnum: 2, type_oid: BIT, size: -1, typmod: 3, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("11011010"), Datum::Text("101")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "BIT returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 BIT(65), v2 BIT(3));",
                "INSERT INTO test VALUES (B'10101010001000110110110010110011000101010110101010010110101011001', '101');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: BIT, size: -1, typmod: 65, format: 0 }, Field { name: "v2", attnum: 2, type_oid: BIT, size: -1, typmod: 3, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 65, 170, 35, 108, 179, 21, 106, 150, 172, 128]), Datum::Bytes(&[0, 0, 0, 3, 160])]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "BIT VARYING returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 BIT VARYING, v2 BIT VARYING(5));",
                "INSERT INTO test VALUES (B'100101010110011001', '110');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: VARBIT, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: VARBIT, size: -1, typmod: 5, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("100101010110011001"), Datum::Text("110")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "BIT VARYING returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 BIT VARYING, v2 BIT VARYING(5));",
                "INSERT INTO test VALUES (B'100101010110011001', '110');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: VARBIT, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: VARBIT, size: -1, typmod: 5, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 18, 149, 102, 64]), Datum::Bytes(&[0, 0, 0, 3, 192])]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "BOOL returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 BOOL, v2 BOOL);",
                "INSERT INTO test VALUES (true, false);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: BOOL, size: 1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: BOOL, size: 1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("t"), Datum::Text("f")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "BOOL returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 BOOL, v2 BOOL);",
                "INSERT INTO test VALUES (true, false);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: BOOL, size: 1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: BOOL, size: 1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\u{1}"), Datum::Text("\0")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "BPCHAR returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 BPCHAR, v2 BPCHAR(7));",
                "INSERT INTO test VALUES ('', 'abc'), ('more text', 'text');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: BPCHAR, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: BPCHAR, size: -1, typmod: 11, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(""), Datum::Text("abc    ")]),
                    Receive::DataRow(&[Datum::Text("more text"), Datum::Text("text   ")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "BPCHAR returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 BPCHAR, v2 BPCHAR(7));",
                "INSERT INTO test VALUES ('', 'abc'), ('more text', 'text');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: BPCHAR, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: BPCHAR, size: -1, typmod: 11, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(""), Datum::Text("abc    ")]),
                    Receive::DataRow(&[Datum::Text("more text"), Datum::Text("text   ")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "BYTEA returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 BYTEA, v2 BYTEA);",
                r#"INSERT INTO test VALUES ('', E'\\xDEADBEEF'), ('\xC0FFEE', NULL);"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: BYTEA, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: BYTEA, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#"\x"#), Datum::Text(r#"\xdeadbeef"#)]),
                    Receive::DataRow(&[Datum::Text(r#"\xc0ffee"#), Datum::Null]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "BYTEA returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 BYTEA, v2 BYTEA);",
                r#"INSERT INTO test VALUES ('', E'\\xDEADBEEF'), ('\xC0FFEE', NULL);"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: BYTEA, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: BYTEA, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(""), Datum::Bytes(&[222, 173, 190, 239])]),
                    Receive::DataRow(&[Datum::Bytes(&[192, 255, 238]), Datum::Null]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: r#""char" returning text format"#,
            set_up_script: &[
                r#"CREATE TABLE test (v1 "char", v2 "char");"#,
                "INSERT INTO test VALUES ('123', 'v');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: CHAR, size: 1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: CHAR, size: 1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("1"), Datum::Text("v")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: r#""char" returning binary format"#,
            set_up_script: &[
                r#"CREATE TABLE test (v1 "char", v2 "char");"#,
                "INSERT INTO test VALUES ('123', 'v');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: CHAR, size: 1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: CHAR, size: 1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("1"), Datum::Text("v")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "CID returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 CID, v2 CID);",
                "INSERT INTO test VALUES ('4', '5');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: CID, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: CID, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("4"), Datum::Text("5")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "CID returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 CID, v2 CID);",
                "INSERT INTO test VALUES ('6', '7');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: CID, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: CID, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{6}"), Datum::Text("\0\0\0\u{7}")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "DATE returning text format",
            set_up_script: &[
                "SET datestyle TO 'ISO, YMD';",
                "CREATE TABLE test (v1 DATE, v2 DATE);",
                "INSERT INTO test VALUES ('1999-01-08', 'April 17, 2025');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: DATE, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: DATE, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("01-08-1999"), Datum::Text("04-17-2025")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "DATE returning binary format",
            set_up_script: &[
                "SET datestyle TO 'ISO, YMD';",
                "CREATE TABLE test (v1 DATE, v2 DATE);",
                "INSERT INTO test VALUES ('1999-01-08', 'April 17, 2025');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: DATE, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: DATE, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[255, 255, 254, 154]), Datum::Text("\0\0$\u{16}")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "ENUM returning text format",
            set_up_script: &[
                "CREATE TYPE enumType AS ENUM ('eval1', 'eval2', 'eval3');",
                "CREATE TABLE test (v1 enumType, v2 enumType);",
                "INSERT INTO test VALUES ('eval1', 'eval3');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: 0, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: 0, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("eval1"), Datum::Text("eval3")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "ENUM returning binary format",
            set_up_script: &[
                "CREATE TYPE enumType AS ENUM ('eval1', 'eval2', 'eval3');",
                "CREATE TABLE test (v1 enumType, v2 enumType);",
                "INSERT INTO test VALUES ('eval1', 'eval3');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: 0, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: 0, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("eval1"), Datum::Text("eval3")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "FLOAT4 returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 FLOAT4, v2 FLOAT4);",
                "INSERT INTO test VALUES (-0.5, 26.015625);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: FLOAT4, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: FLOAT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("-0.5"), Datum::Text("26.015625")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "FLOAT4 returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 FLOAT4, v2 FLOAT4);",
                "INSERT INTO test VALUES (-0.5, 26.015625);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: FLOAT4, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: FLOAT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[191, 0, 0, 0]), Datum::Bytes(&[65, 208, 32, 0])]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "FLOAT8 returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 FLOAT8, v2 FLOAT8);",
                "INSERT INTO test VALUES (-0.5, 26.015625);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: FLOAT8, size: 8, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: FLOAT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("-0.5"), Datum::Text("26.015625")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "FLOAT8 returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 FLOAT8, v2 FLOAT8);",
                "INSERT INTO test VALUES (-0.5, 26.015625);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: FLOAT8, size: 8, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: FLOAT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[191, 224, 0, 0, 0, 0, 0, 0]), Datum::Text("@:\u{4}\0\0\0\0\0")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INT2 returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 INT2, v2 INT2);",
                "INSERT INTO test VALUES (3, 12646);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: INT2, size: 2, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: INT2, size: 2, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("3"), Datum::Text("12646")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INT2 returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 INT2, v2 INT2);",
                "INSERT INTO test VALUES (3, 12646);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: INT2, size: 2, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: INT2, size: 2, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\u{3}"), Datum::Text("1f")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INT2VECTOR returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 int2vector, v2 int2vector);",
                "INSERT INTO test VALUES ('1 2 4 5', '5 87 991');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: INT2VECTOR, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: INT2VECTOR, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("1 2 4 5"), Datum::Text("5 87 991")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INT2VECTOR returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 int2vector, v2 int2vector);",
                "INSERT INTO test VALUES ('1 2 4 5', '5 87 991');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: INT2VECTOR, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: INT2VECTOR, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{1}\0\0\0\0\0\0\0\u{15}\0\0\0\u{4}\0\0\0\0\0\0\0\u{2}\0\u{1}\0\0\0\u{2}\0\u{2}\0\0\0\u{2}\0\u{4}\0\0\0\u{2}\0\u{5}"), Datum::Bytes(&[0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 21, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 2, 0, 5, 0, 0, 0, 2, 0, 87, 0, 0, 0, 2, 3, 223])]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INT4 returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 INT4, v2 INT4);",
                "INSERT INTO test VALUES (-5, 3578457);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: INT4, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("-5"), Datum::Text("3578457")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INT4 returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 INT4, v2 INT4);",
                "INSERT INTO test VALUES (-5, 3578457);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: INT4, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: INT4, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[255, 255, 255, 251]), Datum::Bytes(&[0, 54, 154, 89])]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INT8 returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 INT8, v2 INT8);",
                "INSERT INTO test VALUES (-44, 2578457279345);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: INT8, size: 8, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("-44"), Datum::Text("2578457279345")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INT8 returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 INT8, v2 INT8);",
                "INSERT INTO test VALUES (-44, 2578457279345);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: INT8, size: 8, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: INT8, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[255, 255, 255, 255, 255, 255, 255, 212]), Datum::Bytes(&[0, 0, 2, 88, 88, 7, 187, 113])]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INTERVAL returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 INTERVAL, v2 INTERVAL);",
                "INSERT INTO test VALUES ('@ 1 minute', '2 years 15 months 100 weeks 99 hours 123456789 milliseconds');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: INTERVAL, size: 16, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: INTERVAL, size: 16, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("@ 1 min"), Datum::Text("@ 3 years 3 mons 700 days 133 hours 17 mins 36.789 secs")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "INTERVAL returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 INTERVAL, v2 INTERVAL);",
                "INSERT INTO test VALUES ('@ 1 minute', '2 years 15 months 100 weeks 99 hours 123456789 milliseconds');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: INTERVAL, size: 16, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: INTERVAL, size: 16, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 0, 3, 147, 135, 0, 0, 0, 0, 0, 0, 0, 0, 0]), Datum::Bytes(&[0, 0, 0, 111, 185, 177, 134, 8, 0, 0, 2, 188, 0, 0, 0, 39])]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "JSON returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 JSON, v2 JSON, v3 INT4);",
                r#"INSERT INTO test VALUES ('{"key1": {"key": "value"}}', '{}', 1), ('{"key1": {"key": [2, 3]}}', '[]', 2);"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT v1, v2 FROM test ORDER BY v3;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: JSON, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: JSON, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#"{"key1": {"key": "value"}}"#), Datum::Text("{}")]),
                    Receive::DataRow(&[Datum::Text(r#"{"key1": {"key": [2, 3]}}"#), Datum::Text("[]")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "JSON returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 JSON, v2 JSON, v3 INT4);",
                r#"INSERT INTO test VALUES ('{"key1": {"key": "value"}}', '{}', 1), ('{"key1": {"key": [2, 3]}}', '[]', 2);"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT v1, v2 FROM test ORDER BY v3;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: JSON, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: JSON, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#"{"key1": {"key": "value"}}"#), Datum::Text("{}")]),
                    Receive::DataRow(&[Datum::Text(r#"{"key1": {"key": [2, 3]}}"#), Datum::Text("[]")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "JSONB returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 JSONB, v2 JSONB, v3 INT4);",
                r#"INSERT INTO test VALUES ('{"key1": {"key": "value"}}', '{}', 1), ('{"key1": {"key": [2, 3]}}', '[]', 2);"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT v1, v2 FROM test ORDER BY v3;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: JSONB, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: JSONB, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#"{"key1": {"key": "value"}}"#), Datum::Text("{}")]),
                    Receive::DataRow(&[Datum::Text(r#"{"key1": {"key": [2, 3]}}"#), Datum::Text("[]")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "JSONB returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 JSONB, v2 JSONB, v3 INT4);",
                r#"INSERT INTO test VALUES ('{"key1": {"key": "value"}}', '{}', 1), ('{"key1": {"key": [2, 3]}}', '[]', 2);"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT v1, v2 FROM test ORDER BY v3;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: JSONB, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: JSONB, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\u{1}{\"key1\": {\"key\": \"value\"}}"), Datum::Text("\u{1}{}")]),
                    Receive::DataRow(&[Datum::Text("\u{1}{\"key1\": {\"key\": [2, 3]}}"), Datum::Text("\u{1}[]")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "NAME returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 NAME, v2 NAME);",
                r#"INSERT INTO test VALUES ('', 'abc'), (NULL, 'a",c');"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: NAME, size: 64, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: NAME, size: 64, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Null, Datum::Text(r#"a",c"#)]),
                    Receive::DataRow(&[Datum::Text(""), Datum::Text("abc")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "NAME returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 NAME, v2 NAME);",
                r#"INSERT INTO test VALUES ('', 'abc'), (NULL, 'a",c');"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: NAME, size: 64, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: NAME, size: 64, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Null, Datum::Text(r#"a",c"#)]),
                    Receive::DataRow(&[Datum::Text(""), Datum::Text("abc")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "NUMERIC returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 NUMERIC, v2 NUMERIC(5,2), v3 NUMERIC(14,5));",
                "INSERT INTO test VALUES (0, -0.1, NULL), (12357232.456786653224768755799, 235.67, 4278.009), ('Infinity', 'NaN', 'NaN'), ('-Infinity', '0.05', '0.1045678');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: NUMERIC, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: NUMERIC, size: -1, typmod: 327686, format: 0 }, Field { name: "v3", attnum: 3, type_oid: NUMERIC, size: -1, typmod: 917513, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("-Infinity"), Datum::Text("0.05"), Datum::Text("0.10457")]),
                    Receive::DataRow(&[Datum::Text("0"), Datum::Text("-0.10"), Datum::Null]),
                    Receive::DataRow(&[Datum::Text("12357232.456786653224768755799"), Datum::Text("235.67"), Datum::Text("4278.00900")]),
                    Receive::DataRow(&[Datum::Text("Infinity"), Datum::Text("NaN"), Datum::Text("NaN")]),
                    Receive::CommandComplete("SELECT 4"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "NUMERIC returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 NUMERIC, v2 NUMERIC(5,2), v3 NUMERIC(14,5));",
                "INSERT INTO test VALUES (0, -0.1, NULL), (12357232.456786653224768755799, 235.67, 4278.009), ('Infinity', 'NaN', 'NaN'), ('-Infinity', '0.05', '0.1045678');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: NUMERIC, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: NUMERIC, size: -1, typmod: 327686, format: 0 }, Field { name: "v3", attnum: 3, type_oid: NUMERIC, size: -1, typmod: 917513, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 0, 240, 0, 0, 32]), Datum::Bytes(&[0, 1, 255, 255, 0, 0, 0, 2, 1, 244]), Datum::Bytes(&[0, 2, 255, 255, 0, 0, 0, 5, 4, 21, 27, 88])]),
                    Receive::DataRow(&[Datum::Text("\0\0\0\0\0\0\0\0"), Datum::Bytes(&[0, 1, 255, 255, 64, 0, 0, 2, 3, 232]), Datum::Null]),
                    Receive::DataRow(&[Datum::Bytes(&[0, 8, 0, 1, 0, 0, 0, 21, 4, 211, 28, 64, 17, 215, 33, 217, 12, 152, 30, 7, 21, 203, 35, 40]), Datum::Bytes(&[0, 2, 0, 0, 0, 0, 0, 2, 0, 235, 26, 44]), Datum::Bytes(&[0, 2, 0, 0, 0, 0, 0, 5, 16, 182, 0, 90])]),
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 0, 208, 0, 0, 32]), Datum::Bytes(&[0, 0, 0, 0, 192, 0, 0, 0]), Datum::Bytes(&[0, 0, 0, 0, 192, 0, 0, 0])]),
                    Receive::CommandComplete("SELECT 4"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "OID returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 OID, v2 OID);",
                "INSERT INTO test VALUES (1, 2483574913);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: OID, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: OID, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("1"), Datum::Text("2483574913")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "OID returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 OID, v2 OID);",
                "INSERT INTO test VALUES (1, 2483574913);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: OID, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: OID, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{1}"), Datum::Bytes(&[148, 8, 88, 129])]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "OIDVECTOR returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 oidvector, v2 oidvector);",
                "INSERT INTO test VALUES ('1234 2489', '2483 574 913');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: OIDVECTOR, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: OIDVECTOR, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("1234 2489"), Datum::Text("2483 574 913")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "OIDVECTOR returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 oidvector, v2 oidvector);",
                "INSERT INTO test VALUES ('1234 2489', '2483 574 913');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: OIDVECTOR, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: OIDVECTOR, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 26, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 4, 210, 0, 0, 0, 4, 0, 0, 9, 185]), Datum::Bytes(&[0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 26, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 4, 0, 0, 9, 179, 0, 0, 0, 4, 0, 0, 2, 62, 0, 0, 0, 4, 0, 0, 3, 145])]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "RECORD returning text format",
            set_up_script: &[
                "CREATE TABLE pre1 (v1 TEXT, v2 INT8, v3 NUMERIC(6,1));",
                "CREATE TABLE pre2 (v1 VARCHAR, v2 OID, v3 BOOL);",
                "CREATE TABLE test (v1 pre1, v2 pre2);",
                "INSERT INTO test VALUES (ROW('abc'::TEXT, 1::INT8, '12345.6'::NUMERIC(6,1)), ROW('def'::VARCHAR, 2::OID, 't'::BOOL));",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: 0, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: 0, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("(abc,1,12345.6)"), Datum::Text("(def,2,t)")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "RECORD returning binary format",
            set_up_script: &[
                "CREATE TABLE pre1 (v1 TEXT, v2 INT8, v3 NUMERIC(6,1));",
                "CREATE TABLE pre2 (v1 VARCHAR, v2 OID, v3 BOOL);",
                "CREATE TABLE test (v1 pre1, v2 pre2);",
                "INSERT INTO test VALUES (ROW('abc'::TEXT, 1::INT8, '12345.6'::NUMERIC(6,1)), ROW('def'::VARCHAR, 2::OID, 't'::BOOL));",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: 0, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: 0, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 3, 0, 0, 0, 25, 0, 0, 0, 3, 97, 98, 99, 0, 0, 0, 20, 0, 0, 0, 8, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 6, 164, 0, 0, 0, 14, 0, 3, 0, 1, 0, 0, 0, 1, 0, 1, 9, 41, 23, 112]), Datum::Text("\0\0\0\u{3}\0\0\u{4}\u{13}\0\0\0\u{3}def\0\0\0\u{1a}\0\0\0\u{4}\0\0\0\u{2}\0\0\0\u{10}\0\0\0\u{1}\u{1}")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "REGTYPE returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 REGTYPE, v2 REGTYPE);",
                "INSERT INTO test VALUES ('numeric'::REGTYPE, 'text'::REGTYPE);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: REGTYPE, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: REGTYPE, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("numeric"), Datum::Text("text")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "REGTYPE returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 REGTYPE, v2 REGTYPE);",
                "INSERT INTO test VALUES ('numeric'::REGTYPE, 'text'::REGTYPE);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: REGTYPE, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: REGTYPE, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 6, 164]), Datum::Text("\0\0\0\u{19}")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TEXT returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 TEXT, v2 TEXT);",
                r#"INSERT INTO test VALUES ('', 'abc'), (NULL, 'a",c');"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TEXT, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Null, Datum::Text(r#"a",c"#)]),
                    Receive::DataRow(&[Datum::Text(""), Datum::Text("abc")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TEXT returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 TEXT, v2 TEXT);",
                r#"INSERT INTO test VALUES ('', 'abc'), (NULL, 'a",c');"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TEXT, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TEXT, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Null, Datum::Text(r#"a",c"#)]),
                    Receive::DataRow(&[Datum::Text(""), Datum::Text("abc")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TEXT ARRAY returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 TEXT[], v2 TEXT[]);",
                "INSERT INTO test VALUES (ARRAY[]::text[], ARRAY['a','bb','ccc']), (NULL, ARRAY['dd',NULL,'ee']);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TEXT_ARRAY, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TEXT_ARRAY, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("{}"), Datum::Text("{a,bb,ccc}")]),
                    Receive::DataRow(&[Datum::Null, Datum::Text("{dd,NULL,ee}")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TEXT ARRAY returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 TEXT[], v2 TEXT[]);",
                "INSERT INTO test VALUES (ARRAY[]::text[], ARRAY['a','bb','ccc']), (NULL, ARRAY['dd',NULL,'ee']);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TEXT_ARRAY, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TEXT_ARRAY, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\0\0\0\0\0\0\0\0\u{19}"), Datum::Text("\0\0\0\u{1}\0\0\0\0\0\0\0\u{19}\0\0\0\u{3}\0\0\0\u{1}\0\0\0\u{1}a\0\0\0\u{2}bb\0\0\0\u{3}ccc")]),
                    Receive::DataRow(&[Datum::Null, Datum::Bytes(&[0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 25, 0, 0, 0, 3, 0, 0, 0, 1, 0, 0, 0, 2, 100, 100, 255, 255, 255, 255, 0, 0, 0, 2, 101, 101])]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TID returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 TID, v2 TID);",
                "INSERT INTO test VALUES ('(44,55)', '(66,77)');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TID, size: 6, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TID, size: 6, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("(44,55)"), Datum::Text("(66,77)")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TID returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 TID, v2 TID);",
                "INSERT INTO test VALUES ('(12,34)', '(56,78)');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TID, size: 6, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TID, size: 6, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{c}\0\""), Datum::Text("\0\0\08\0N")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TIME returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 TIME, v2 TIME);",
                "INSERT INTO test VALUES ('0:0', '04:05:06.789'), ('09:27 PM', '12:12');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TIME, size: 8, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TIME, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("00:00:00"), Datum::Text("04:05:06.789")]),
                    Receive::DataRow(&[Datum::Text("21:27:00"), Datum::Text("12:12:00")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TIME returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 TIME, v2 TIME);",
                "INSERT INTO test VALUES ('0:0', '04:05:06.789'), ('09:27 PM', '12:12');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TIME, size: 8, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TIME, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\0\0\0\0\0"), Datum::Bytes(&[0, 0, 0, 3, 108, 151, 202, 136])]),
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 17, 250, 171, 177, 0]), Datum::Bytes(&[0, 0, 0, 10, 57, 214, 4, 0])]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TIMETZ returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 TIMETZ, v2 TIMETZ);",
                "INSERT INTO test VALUES ('0:0 PST', '04:05:06.789 MST'), ('09:27 PM CST', '12:12 EST');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TIMETZ, size: 12, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TIMETZ, size: 12, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("00:00:00-08"), Datum::Text("04:05:06.789-07")]),
                    Receive::DataRow(&[Datum::Text("21:27:00-06"), Datum::Text("12:12:00-05")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TIMETZ returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 TIMETZ, v2 TIMETZ);",
                "INSERT INTO test VALUES ('0:0 PST', '04:05:06.789 PDT'), ('09:27 PM CST', '12:12 EST');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TIMETZ, size: 12, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TIMETZ, size: 12, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 112, 128]), Datum::Bytes(&[0, 0, 0, 3, 108, 151, 202, 136, 0, 0, 98, 112])]),
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 17, 250, 171, 177, 0, 0, 0, 84, 96]), Datum::Bytes(&[0, 0, 0, 10, 57, 214, 4, 0, 0, 0, 70, 80])]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TIMESTAMP returning text format",
            set_up_script: &[
                "SET datestyle TO 'Postgres, MDY';",
                "CREATE TABLE test (v1 TIMESTAMP, v2 TIMESTAMP);",
                "INSERT INTO test VALUES ('2020-01-12 00:00:00', '2021-02-13 04:05:06.789'), ('2022-03-14 10:11:12', '2023-04-15 11:12:13');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TIMESTAMP, size: 8, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TIMESTAMP, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("Sun Jan 12 00:00:00 2020"), Datum::Text("Sat Feb 13 04:05:06.789 2021")]),
                    Receive::DataRow(&[Datum::Text("Mon Mar 14 10:11:12 2022"), Datum::Text("Sat Apr 15 11:12:13 2023")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TIMESTAMP returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 TIMESTAMP, v2 TIMESTAMP);",
                "INSERT INTO test VALUES ('2020-01-12 00:00:00', '2021-02-13 04:05:06.789'), ('2022-03-14 10:11:12', '2023-04-15 11:12:13');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TIMESTAMP, size: 8, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TIMESTAMP, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[0, 2, 62, 228, 207, 3, 128, 0]), Datum::Bytes(&[0, 2, 94, 46, 160, 114, 138, 136])]),
                    Receive::DataRow(&[Datum::Bytes(&[0, 2, 125, 41, 171, 38, 208, 0]), Datum::Bytes(&[0, 2, 156, 92, 204, 93, 29, 64])]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TIMESTAMPTZ returning text format",
            set_up_script: &[
                "SET datestyle TO 'Postgres, MDY';",
                "CREATE TABLE test (v1 TIMESTAMPTZ, v2 TIMESTAMPTZ);",
                "INSERT INTO test VALUES ('2020-01-12 00:00:00 PST', '2021-02-13 04:05:06.789 MST'), ('2022-03-14 10:11:12 CST', '2023-04-15 11:12:13 EST');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TIMESTAMPTZ, size: 8, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TIMESTAMPTZ, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("Sun Jan 12 00:00:00 2020 PST"), Datum::Text("Sat Feb 13 03:05:06.789 2021 PST")]),
                    Receive::DataRow(&[Datum::Text("Mon Mar 14 09:11:12 2022 PDT"), Datum::Text("Sat Apr 15 09:12:13 2023 PDT")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "TIMESTAMPTZ returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 TIMESTAMPTZ, v2 TIMESTAMPTZ);",
                "INSERT INTO test VALUES ('2020-01-12 00:00:00 PST', '2021-02-13 04:05:06.789 MST'), ('2022-03-14 10:11:12 CST', '2023-04-15 11:12:13 EST');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: TIMESTAMPTZ, size: 8, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: TIMESTAMPTZ, size: 8, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[0, 2, 62, 235, 131, 160, 160, 0]), Datum::Bytes(&[0, 2, 94, 52, 126, 124, 6, 136])]),
                    Receive::DataRow(&[Datum::Bytes(&[0, 2, 125, 46, 178, 156, 168, 0]), Datum::Bytes(&[0, 2, 156, 96, 253, 63, 81, 64])]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "UUID returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 UUID, v2 UUID);",
                "INSERT INTO test VALUES ('fdabf03d-9b21-4531-b900-c6f6cff8386c', '0730791c-c0dd-4972-9e72-11ead9317a5a');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: UUID, size: 16, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: UUID, size: 16, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("fdabf03d-9b21-4531-b900-c6f6cff8386c"), Datum::Text("0730791c-c0dd-4972-9e72-11ead9317a5a")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "UUID returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 UUID, v2 UUID);",
                "INSERT INTO test VALUES ('fdabf03d-9b21-4531-b900-c6f6cff8386c', '0730791c-c0dd-4972-9e72-11ead9317a5a');",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: UUID, size: 16, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: UUID, size: 16, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[253, 171, 240, 61, 155, 33, 69, 49, 185, 0, 198, 246, 207, 248, 56, 108]), Datum::Bytes(&[7, 48, 121, 28, 192, 221, 73, 114, 158, 114, 17, 234, 217, 49, 122, 90])]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "VARCHAR returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 VARCHAR, v2 VARCHAR(5));",
                r#"INSERT INTO test VALUES ('', 'abc'), (NULL, 'a",c');"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: VARCHAR, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: VARCHAR, size: -1, typmod: 9, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Null, Datum::Text(r#"a",c"#)]),
                    Receive::DataRow(&[Datum::Text(""), Datum::Text("abc")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "VARCHAR returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 VARCHAR, v2 VARCHAR(5));",
                r#"INSERT INTO test VALUES ('', 'abc'), (NULL, 'a",c');"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test ORDER BY v1 NULLS FIRST;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: VARCHAR, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: VARCHAR, size: -1, typmod: 9, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Null, Datum::Text(r#"a",c"#)]),
                    Receive::DataRow(&[Datum::Text(""), Datum::Text("abc")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "XID returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 XID, v2 XID);",
                "INSERT INTO test VALUES (1::TEXT::XID, 2483574913::TEXT::XID);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: XID, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: XID, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("1"), Datum::Text("2483574913")]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "XID returning binary format",
            set_up_script: &[
                "CREATE TABLE test (v1 XID, v2 XID);",
                "INSERT INTO test VALUES (1::TEXT::XID, 2483574913::TEXT::XID);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT * FROM test;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: XID, size: 4, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: XID, size: 4, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("\0\0\0\u{1}"), Datum::Bytes(&[148, 8, 88, 129])]),
                    Receive::CommandComplete("SELECT 1"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "XML returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 XML, v2 XML, v3 INT4);",
                r#"INSERT INTO test VALUES ('<a>x</a>', '', 1), (NULL, '<b c="1">y&amp;z</b>', 2);"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT v1, v2 FROM test ORDER BY v3;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: XML, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: XML, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("<a>x</a>"), Datum::Text("")]),
                    Receive::DataRow(&[Datum::Null, Datum::Text(r#"<b c="1">y&amp;z</b>"#)]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "XML returning binary format",
            set_up_script: &[
                "SET client_encoding TO 'UTF8';",
                "CREATE TABLE test (v1 XML, v2 XML, v3 INT4);",
                r#"INSERT INTO test VALUES ('<a>x</a>', '', 1), (NULL, '<b c="1">y&amp;z</b>', 2);"#,
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT v1, v2 FROM test ORDER BY v3;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: XML, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: XML, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text(r#"<?xml version="1.0" encoding="WIN1252"?><a>x</a>"#), Datum::Text(r#"<?xml version="1.0" encoding="WIN1252"?>"#)]),
                    Receive::DataRow(&[Datum::Null, Datum::Text(r#"<?xml version="1.0" encoding="WIN1252"?><b c="1">y&amp;z</b>"#)]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "XML ARRAY returning text format",
            set_up_script: &[
                "CREATE TABLE test (v1 XML[], v2 XML[], v3 INT4);",
                "INSERT INTO test VALUES ('{}', ARRAY['<a/>'::xml, '<b>x</b>', '<c/>'], 1), (NULL, ARRAY['<d/>'::xml, NULL, '<e/>'], 2);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT v1, v2 FROM test ORDER BY v3;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: XML_ARRAY, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: XML_ARRAY, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[0] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Text("{}"), Datum::Text("{<a/>,<b>x</b>,<c/>}")]),
                    Receive::DataRow(&[Datum::Null, Datum::Text("{<d/>,NULL,<e/>}")]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
        WireTest {
            name: "XML ARRAY returning binary format",
            set_up_script: &[
                "SET client_encoding TO 'UTF8';",
                "CREATE TABLE test (v1 XML[], v2 XML[], v3 INT4);",
                "INSERT INTO test VALUES ('{}', ARRAY['<a/>'::xml, '<b>x</b>', '<c/>'], 1), (NULL, ARRAY['<d/>'::xml, NULL, '<e/>'], 2);",
            ],
            steps: &[
                Step::Send(&[
                    Send::Parse { name: "stmt_name", query: "SELECT v1, v2 FROM test ORDER BY v3;", parameter_oids: &[] },
                    Send::Describe(b'S', "stmt_name"),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::ParseComplete,
                    Receive::ParameterDescription(&[]),
                    Receive::RowDescription(&[Field { name: "v1", attnum: 1, type_oid: XML_ARRAY, size: -1, typmod: -1, format: 0 }, Field { name: "v2", attnum: 2, type_oid: XML_ARRAY, size: -1, typmod: -1, format: 0 }]),
                    Receive::ReadyForQuery(b'I'),
                ]),
                Step::Send(&[
                    Send::Bind { portal: "", statement: "stmt_name", parameter_formats: &[], parameters: &[], result_formats: &[1] },
                    Send::Execute("", 0),
                    Send::Close(b'P', ""),
                    Send::Sync,
                ]),
                Step::Receive(&[
                    Receive::BindComplete,
                    Receive::DataRow(&[Datum::Bytes(&[0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 142]), Datum::Bytes(&[0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 142, 0, 0, 0, 3, 0, 0, 0, 1, 0, 0, 0, 44, 60, 63, 120, 109, 108, 32, 118, 101, 114, 115, 105, 111, 110, 61, 34, 49, 46, 48, 34, 32, 101, 110, 99, 111, 100, 105, 110, 103, 61, 34, 87, 73, 78, 49, 50, 53, 50, 34, 63, 62, 60, 97, 47, 62, 0, 0, 0, 48, 60, 63, 120, 109, 108, 32, 118, 101, 114, 115, 105, 111, 110, 61, 34, 49, 46, 48, 34, 32, 101, 110, 99, 111, 100, 105, 110, 103, 61, 34, 87, 73, 78, 49, 50, 53, 50, 34, 63, 62, 60, 98, 62, 120, 60, 47, 98, 62, 0, 0, 0, 44, 60, 63, 120, 109, 108, 32, 118, 101, 114, 115, 105, 111, 110, 61, 34, 49, 46, 48, 34, 32, 101, 110, 99, 111, 100, 105, 110, 103, 61, 34, 87, 73, 78, 49, 50, 53, 50, 34, 63, 62, 60, 99, 47, 62])]),
                    Receive::DataRow(&[Datum::Null, Datum::Bytes(&[0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 142, 0, 0, 0, 3, 0, 0, 0, 1, 0, 0, 0, 44, 60, 63, 120, 109, 108, 32, 118, 101, 114, 115, 105, 111, 110, 61, 34, 49, 46, 48, 34, 32, 101, 110, 99, 111, 100, 105, 110, 103, 61, 34, 87, 73, 78, 49, 50, 53, 50, 34, 63, 62, 60, 100, 47, 62, 255, 255, 255, 255, 0, 0, 0, 44, 60, 63, 120, 109, 108, 32, 118, 101, 114, 115, 105, 111, 110, 61, 34, 49, 46, 48, 34, 32, 101, 110, 99, 111, 100, 105, 110, 103, 61, 34, 87, 73, 78, 49, 50, 53, 50, 34, 63, 62, 60, 101, 47, 62])]),
                    Receive::CommandComplete("SELECT 2"),
                    Receive::CloseComplete,
                    Receive::ReadyForQuery(b'I'),
                ]),
            ],
            ..W
        },
    ]);
}
