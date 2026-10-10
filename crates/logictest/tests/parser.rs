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

use std::path::Path;

use logictest::parser::{Condition, Record, RecordType, parse_test_file};
use logictest::results::parse_log;

fn data(name: &str) -> String {
    format!("{}/tests/data/{name}", env!("CARGO_MANIFEST_DIR"))
}

fn record(record_type: RecordType, line_num: usize, query: &str) -> Record {
    Record {
        record_type,
        expect_error: false,
        conditions: Vec::new(),
        schema: String::new(),
        sort_mode: String::new(),
        query: query.replace('\n', ""),
        line_num,
        result: Vec::new(),
        label: String::new(),
        hash_threshold: 8,
    }
}

fn query(line_num: usize, schema: &str, sort_mode: &str, sql: &str, result: &[&str]) -> Record {
    Record {
        schema: schema.to_string(),
        sort_mode: sort_mode.to_string(),
        result: result.iter().map(|r| r.to_string()).collect(),
        hash_threshold: 16,
        ..record(RecordType::Query, line_num, sql)
    }
}

fn only(engine: &str) -> Condition {
    Condition { only: true, engine: engine.to_string() }
}

fn skip(engine: &str) -> Condition {
    Condition { only: false, engine: engine.to_string() }
}

#[test]
fn parse_file() {
    let records = parse_test_file(Path::new(&data("select1.test"))).unwrap();
    let expected = vec![
        record(RecordType::Statement, 2, "CREATE TABLE t1(a INTEGER, b INTEGER, c INTEGER, d INTEGER, e INTEGER)"),
        record(RecordType::Statement, 5, "INSERT INTO t1(e,c,b,d,a) VALUES(103,102,100,101,104)"),
        Record {
            expect_error: true,
            ..record(RecordType::Statement, 8, "INSERT INTO t1(a,c,d,e,b) VALUES(107,106,108,109,105)")
        },
        record(RecordType::Halt, 11, ""),
        Record {
            hash_threshold: 8,
            ..query(
                14,
                "I",
                "nosort",
                "SELECT CASE WHEN c>(SELECT avg(c) FROM t1) THEN a*2 ELSE b*10 END\n  FROM t1\n ORDER BY 1",
                &["30 values hashing to 3c13dee48d9356ae19af2515e05e6b54"],
            )
        },
        Record {
            label: "label-1".to_string(),
            ..query(
                29,
                "II",
                "nosort",
                "SELECT a+b*2+c*3+d*4+e*5,\n       (a+b+c+d+e)/5\n  FROM t1\n ORDER BY 1,2",
                &["60 values hashing to 808146289313018fce25f1a280bd8c30"],
            )
        },
        Record { conditions: vec![only("mysql")], hash_threshold: 16, ..record(RecordType::Halt, 37, "") },
        Record {
            conditions: vec![only("mysql")],
            ..query(
                41,
                "IIIII",
                "rowsort",
                "SELECT a+b*2+c*3+d*4+e*5,\n       CASE WHEN a<b-3 THEN 111 WHEN a<=b THEN 222\n        WHEN a<b+3 \
                 THEN 333 ELSE 444 END,\n       abs(b-c),\n       (a+b+c+d+e)/5,\n       a+b*2+c*3\n  FROM t1\n WHERE \
                 (e>c OR e<d)\n   AND d>e\n   AND EXISTS(SELECT 1 FROM t1 AS x WHERE x.b<t1.b)\n ORDER BY 4,2,1,3,5",
                &["1", "2", "3", "4", "5"],
            )
        },
        Record {
            conditions: vec![skip("mssql")],
            ..query(
                62,
                "II",
                "nosort",
                "SELECT a-b,\n       CASE WHEN a<b-3 THEN 111 WHEN a<=b THEN 222\n        WHEN a<b+3 THEN 333 ELSE \
                 444 END\n  FROM t1\n WHERE c>d\n   AND b>c\n ORDER BY 2,1",
                &["-3", "222", "-3", "222", "-1", "222", "-1", "222"],
            )
        },
        Record {
            hash_threshold: 16,
            ..record(
                RecordType::Statement,
                80,
                "CREATE TABLE t1(\n  a1 INTEGER,\n  b1 INTEGER,\n  c1 INTEGER,\n  d1 INTEGER,\n  e1 INTEGER,\n  x1 \
                 VARCHAR(30)\n)",
            )
        },
        Record {
            label: "join-4-1".to_string(),
            ..query(
                90,
                "TTTT",
                "valuesort",
                "SELECT x29,x31,x51,x55\n  FROM t51,t29,t31,t55\n  WHERE a51=b31\n    AND a29=6\n    AND a29=b51\n    \
                 AND b55=a31",
                &["table t29 row 6", "table t31 row 9", "table t51 row 5", "table t55 row 4"],
            )
        },
        Record {
            conditions: vec![skip("mysql"), skip("mssql"), skip("oracle")],
            ..query(106, "I", "nosort", "SELECT 1 FROM t1 WHERE 1.0 IN ()", &[])
        },
    ];
    assert_eq!(records, expected);
}

#[test]
fn record_methods() {
    let strings = |values: &[&str]| values.iter().map(|v| v.to_string()).collect::<Vec<_>>();
    let valuesort = query(90, "TTTT", "valuesort", "", &["w", "x", "y", "z"]);
    assert_eq!(valuesort.num_results(), 4);
    assert!(!valuesort.is_hash_result());
    assert!(valuesort.should_execute_for_engine("mysql"));
    assert_eq!(valuesort.sort_results(strings(&["c", "a", "d", "b"])).unwrap(), strings(&["a", "b", "c", "d"]));

    let skipped = Record { conditions: vec![skip("mssql")], ..query(62, "II", "nosort", "", &["1", "2"]) };
    assert!(skipped.should_execute_for_engine("mysql"));
    assert!(!skipped.should_execute_for_engine("mssql"));
    assert_eq!(skipped.sort_results(strings(&["c", "b", "a"])).unwrap(), strings(&["c", "b", "a"]));

    let rowsort = Record { conditions: vec![only("mysql")], ..query(41, "IIIII", "rowsort", "", &[]) };
    assert!(rowsort.should_execute_for_engine("mysql"));
    assert!(!rowsort.should_execute_for_engine("postgresql"));
    let rows = strings(&[
        "c", "a", "z", "e", "g", "a", "j", "k", "e", "3", "d", "b", "w", "q", "g", "c", "a", "z", "e", "f", "b", "l",
        "2", "foo", "m", "c", "a", "z", "e", "f",
    ]);
    let sorted = strings(&[
        "a", "j", "k", "e", "3", "b", "l", "2", "foo", "m", "c", "a", "z", "e", "f", "c", "a", "z", "e", "f", "c", "a",
        "z", "e", "g", "d", "b", "w", "q", "g",
    ]);
    assert_eq!(rowsort.sort_results(rows).unwrap(), sorted);

    let hashed = query(29, "II", "nosort", "", &["60 values hashing to 808146289313018fce25f1a280bd8c30"]);
    assert_eq!(hashed.num_results(), 60);
    assert!(hashed.is_hash_result());
    assert_eq!(hashed.hash(), "808146289313018fce25f1a280bd8c30");

    let skipped =
        Record { conditions: vec![skip("mysql"), skip("mssql"), skip("oracle")], ..query(106, "I", "nosort", "", &[]) };
    assert!(!skipped.should_execute_for_engine("mysql"));
    assert!(!skipped.should_execute_for_engine("mssql"));
    assert!(skipped.should_execute_for_engine("postgresql"));
}

#[test]
fn parse_result_file() {
    let entries = parse_log(&std::fs::read_to_string(data("resultlog.txt")).unwrap()).unwrap();
    let summary: Vec<(usize, &str, u64, &str, &str)> = entries
        .iter()
        .map(|e| (e.line_num, e.query.as_str(), e.duration_ms, e.result, e.error_message.as_str()))
        .collect();
    assert!(entries.iter().all(|e| e.test_file == "evidence/in1.test"));
    assert_eq!(
        summary,
        vec![
            (25, "SELECT 1 IN ()", 213654, "skipped", ""),
            (30, "SELECT 1 IN (2)", 789321, "ok", ""),
            (35, "SELECT 1 IN (2,3,4,5,6,7,8,9)", 123445, "ok", ""),
            (41, "SELECT 1 NOT IN ()", 9807843, "skipped", ""),
            (46, "SELECT 1 NOT IN (2)", 34121, "ok", ""),
            (51, "SELECT 1 NOT IN (2,3,4,5,6,7,8,9)", 2123, "ok", ""),
            (57, "SELECT null IN ()", 21456998, "skipped", ""),
            (63, "SELECT null NOT IN ()", 395874, "skipped", ""),
            (68, "CREATE TABLE t1(x INTEGER)", 87838293, "not ok", "Unexpected error no primary key columns"),
            (72, "SELECT 1 IN t1", 98321, "skipped", ""),
        ]
    );
}
