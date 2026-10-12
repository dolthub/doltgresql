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

mod common;

use std::sync::{Arc, Barrier};

use common::*;
use driver::client::{Db, Value};
use driver::runner::Env;
use harness::pgx::Arg;

/// build_wide_create_table returns a CREATE TABLE with an id key and num_cols columns of the given type.
fn build_wide_create_table(table_name: &str, num_cols: usize, col_type: &str) -> String {
    let columns: String = (0..num_cols).map(|i| format!(", c{i} {col_type}")).collect();
    format!("CREATE TABLE {table_name} (id BIGINT PRIMARY KEY{columns})")
}

/// build_wide_insert returns an INSERT of one row, with make_val returning each column's SQL literal.
fn build_wide_insert(table_name: &str, row_id: i64, num_cols: usize, make_val: impl Fn(usize) -> String) -> String {
    let values: String = (0..num_cols).map(|i| format!(", {}", make_val(i))).collect();
    format!("INSERT INTO {table_name} VALUES ({row_id}{values})")
}

/// build_wide_text_insert returns a parameterized INSERT of one row with a test text argument per column.
fn build_wide_text_insert(table_name: &str, row: i64, num_cols: usize, size: usize) -> (String, Vec<Arg>) {
    let mut args = vec![Arg::Int(row)];
    args.extend((0..num_cols as i64).map(|col| Arg::Str(make_test_text(row * num_cols as i64 + col, size))));
    let placeholders: Vec<String> = (1..=args.len()).map(|i| format!("${i}")).collect();
    (format!("INSERT INTO {table_name} VALUES ({})", placeholders.join(",")), args)
}

/// verify_after_gc runs the check, garbage collects, reconnects, and runs the check again.
fn verify_after_gc(env: &mut Env, db: Db, check: impl Fn(&mut Db)) {
    let mut db = db;
    check(&mut db);
    let mut db = gc_and_reconnect(env, db);
    check(&mut db);
}

/// gc_and_reconnect garbage collects and returns a new connection.
fn gc_and_reconnect(env: &mut Env, mut db: Db) -> Db {
    exec(&mut db, "SELECT dolt_gc()", &[]);
    db.close();
    common::db(env)
}

/// opt returns an argument, or NULL.
fn opt(value: Option<Arg>) -> Arg {
    value.unwrap_or(Arg::Null)
}

#[test]
fn test_large_out_of_band_values_large_text() {
    const TEXT_SIZE: usize = 15_000;
    const NUM_ROWS: i64 = 20;
    let mut env = env();
    setup_test_server(&mut env, "large_text_values");
    let mut db = db(&mut env);
    exec(&mut db, "CREATE TABLE large_text (id BIGINT PRIMARY KEY, txt TEXT)", &[]);
    for i in 0..NUM_ROWS {
        exec(&mut db, "INSERT INTO large_text VALUES ($1, $2)", &[Arg::Int(i), Arg::Str(make_test_text(i, TEXT_SIZE))]);
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'insert large text rows')", &[]);
    verify_after_gc(&mut env, db, |db| {
        assert_eq!(query_int(db, "SELECT COUNT(*) FROM large_text", &[]), NUM_ROWS);
        let sql = "SELECT COUNT(*) FROM large_text WHERE octet_length(txt) = $1";
        assert_eq!(
            query_int(db, sql, &[Arg::Int(TEXT_SIZE as i64)]),
            NUM_ROWS,
            "every row should retain the full text length"
        );
        for id in [0, NUM_ROWS / 2, NUM_ROWS - 1] {
            let row = query_row(db, "SELECT txt FROM large_text WHERE id = $1", &[Arg::Int(id)]);
            assert_eq!(
                row,
                vec![Value::Text(make_test_text(id, TEXT_SIZE))],
                "row {id}: text content must survive storage and retrieval"
            );
        }
    });
    finish(env);
}

#[test]
fn test_large_out_of_band_values_large_blob() {
    const BLOB_SIZE: usize = 25_000;
    const NUM_ROWS: i64 = 20;
    let mut env = env();
    setup_test_server(&mut env, "large_blob_values");
    let mut db = db(&mut env);
    exec(&mut db, "CREATE TABLE large_blob (id BIGINT PRIMARY KEY, data BYTEA)", &[]);
    for i in 0..NUM_ROWS {
        exec(
            &mut db,
            "INSERT INTO large_blob VALUES ($1, $2)",
            &[Arg::Int(i), Arg::Bytes(make_test_blob_data(i, BLOB_SIZE))],
        );
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'insert large blob rows')", &[]);
    verify_after_gc(&mut env, db, |db| {
        assert_eq!(query_int(db, "SELECT COUNT(*) FROM large_blob", &[]), NUM_ROWS);
        let sql = "SELECT COUNT(*) FROM large_blob WHERE octet_length(data) = $1";
        assert_eq!(
            query_int(db, sql, &[Arg::Int(BLOB_SIZE as i64)]),
            NUM_ROWS,
            "every row should retain the full blob length"
        );
        for id in [0, NUM_ROWS / 2, NUM_ROWS - 1] {
            let row = query_row(db, "SELECT data FROM large_blob WHERE id = $1", &[Arg::Int(id)]);
            assert_eq!(
                row,
                vec![Value::Bytes(make_test_blob_data(id, BLOB_SIZE))],
                "row {id}: blob content must survive storage and retrieval"
            );
        }
    });
    finish(env);
}

#[test]
fn test_large_out_of_band_values_large_json() {
    const JSON_TARGET_SIZE: usize = 12_000;
    const NUM_ROWS: i64 = 20;
    let mut env = env();
    setup_test_server(&mut env, "large_json_values");
    let mut db = db(&mut env);
    let values: Vec<String> = (0..NUM_ROWS).map(|i| make_test_json_string(i, JSON_TARGET_SIZE)).collect();
    for value in &values {
        assert!(value.len() >= JSON_TARGET_SIZE, "generated JSON should meet the target size floor");
    }
    exec(&mut db, "CREATE TABLE large_json (id BIGINT PRIMARY KEY, doc JSON)", &[]);
    for (i, value) in values.iter().enumerate() {
        exec(&mut db, "INSERT INTO large_json VALUES ($1, $2)", &[Arg::Int(i as i64), Arg::Str(value.clone())]);
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'insert large json rows')", &[]);
    verify_after_gc(&mut env, db, |db| {
        assert_eq!(query_int(db, "SELECT COUNT(*) FROM large_json", &[]), NUM_ROWS);
        for id in [0, NUM_ROWS / 2, NUM_ROWS - 1] {
            let row = query_row(db, "SELECT doc FROM large_json WHERE id = $1", &[Arg::Int(id)]);
            let doc = row[0].to_go_string();
            assert_eq!(json_id(&doc), id, "row {id}: JSON id field must be preserved after storage");
        }
    });
    finish(env);
}

#[test]
fn test_large_out_of_band_values_mixed_large_columns() {
    const TEXT_SIZE: usize = 18_000;
    const BLOB_SIZE: usize = 22_000;
    const JSON_TARGET: usize = 11_000;
    const NUM_ROWS: i64 = 10;
    let mut env = env();
    setup_test_server(&mut env, "mixed_large_columns");
    let mut db = db(&mut env);
    exec(
        &mut db,
        "CREATE TABLE mixed_large (
				id       BIGINT PRIMARY KEY,
				txt      TEXT,
				bin_data BYTEA,
				doc      JSON,
				note     TEXT
			)",
        &[],
    );
    for i in 0..NUM_ROWS {
        exec(
            &mut db,
            "INSERT INTO mixed_large VALUES ($1, $2, $3, $4, $5)",
            &[
                Arg::Int(i),
                Arg::Str(make_test_text(i, TEXT_SIZE)),
                Arg::Bytes(make_test_blob_data(i, BLOB_SIZE)),
                Arg::Str(make_test_json_string(i, JSON_TARGET)),
                Arg::Str(make_test_text(i + 1000, 5_000)),
            ],
        );
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'insert mixed large column rows')", &[]);
    verify_after_gc(&mut env, db, |db| {
        assert_eq!(query_int(db, "SELECT COUNT(*) FROM mixed_large", &[]), NUM_ROWS);
        let sql = "SELECT COUNT(*) FROM mixed_large WHERE octet_length(txt) = $1 AND octet_length(bin_data) = $2 AND \
                   octet_length(note) = 5000";
        let args = [Arg::Int(TEXT_SIZE as i64), Arg::Int(BLOB_SIZE as i64)];
        assert_eq!(query_int(db, sql, &args), NUM_ROWS, "all large column lengths must be preserved");
        for id in [0, NUM_ROWS - 1] {
            let row = query_row(db, "SELECT txt, bin_data FROM mixed_large WHERE id = $1", &[Arg::Int(id)]);
            assert_eq!(
                row,
                vec![Value::Text(make_test_text(id, TEXT_SIZE)), Value::Bytes(make_test_blob_data(id, BLOB_SIZE))]
            );
        }
    });
    finish(env);
}

#[test]
fn test_large_out_of_band_values_concurrent_large_value_writes() {
    const TEXT_SIZE: usize = 12_000;
    const BLOB_SIZE: usize = 15_000;
    const NUM_WORKERS: i64 = 8;
    const ROWS_PER_WORKER: i64 = 10;
    let mut env = env();
    setup_test_server(&mut env, "concurrent_large_values");
    let mut db = db(&mut env);
    exec(
        &mut db,
        "CREATE TABLE large_concurrent (
				id   BIGINT PRIMARY KEY,
				txt  TEXT,
				data BYTEA
			)",
        &[],
    );
    exec(&mut db, "SELECT dolt_commit('-Am', 'create large_concurrent table')", &[]);
    let mut worker_dbs: Vec<Db> = (0..NUM_WORKERS).map(|_| common::db(&mut env)).collect();
    let start = Arc::new(Barrier::new(NUM_WORKERS as usize));
    let results: Vec<Result<(), String>> = std::thread::scope(|scope| {
        let handles: Vec<_> = worker_dbs
            .iter_mut()
            .enumerate()
            .map(|(w, conn)| {
                let start = start.clone();
                scope.spawn(move || -> Result<(), String> {
                    let w = w as i64;
                    start.wait();
                    for j in 0..ROWS_PER_WORKER {
                        let row_id = w * ROWS_PER_WORKER + j;
                        conn.exec_args(
                            "INSERT INTO large_concurrent VALUES ($1, $2, $3)",
                            &[
                                Arg::Int64(row_id),
                                Arg::Str(make_test_text(row_id, TEXT_SIZE)),
                                Arg::Bytes(make_test_blob_data(row_id, BLOB_SIZE)),
                            ],
                        )
                        .map_err(|e| format!("worker {w} insert row {j}: {e}"))?;
                    }
                    match conn.exec(&format!("SELECT dolt_commit('-Am', 'worker {w} inserts')"), &[]) {
                        Err(e) if !e.contains("nothing to commit") => Err(format!("worker {w} commit: {e}")),
                        _ => Ok(()),
                    }
                })
            })
            .collect();
        handles.into_iter().map(|h| h.join().unwrap()).collect()
    });
    for result in results {
        result.unwrap();
    }
    assert_eq!(
        query_int(&mut db, "SELECT COUNT(*) FROM large_concurrent", &[]),
        NUM_WORKERS * ROWS_PER_WORKER,
        "all rows from all workers must be present"
    );
    let sql = "SELECT COUNT(*) FROM large_concurrent WHERE octet_length(txt) = $1 AND octet_length(data) = $2";
    let args = [Arg::Int(TEXT_SIZE as i64), Arg::Int(BLOB_SIZE as i64)];
    assert_eq!(
        query_int(&mut db, sql, &args),
        NUM_WORKERS * ROWS_PER_WORKER,
        "all large values must have the correct size"
    );
    drop(worker_dbs);
    finish(env);
}

#[test]
#[ignore = "Doltgres panics (nil pointer dereference) handling a NUMERIC column holding the uint64 max value 18446744073709551615"]
fn test_type_diversity_integer_types() {
    let mut env = env();
    setup_test_server(&mut env, "integer_types_test");
    let mut db = db(&mut env);
    exec(
        &mut db,
        "CREATE TABLE int_types (
			id               BIGINT PRIMARY KEY,
			col_tinyint      SMALLINT,
			col_smallint     SMALLINT,
			col_mediumint    INTEGER,
			col_int          INTEGER,
			col_bigint       BIGINT,
			col_tinyint_u    SMALLINT,
			col_smallint_u   INTEGER,
			col_int_u        BIGINT,
			col_bigint_u     NUMERIC
		)",
        &[],
    );
    let rows: [[i128; 10]; 4] = [
        [0, -128, -32768, -8388608, -2147483648, -9223372036854775808, 0, 0, 0, 0],
        [1, 0, 0, 0, 0, 0, 0, 0, 0, 0],
        [2, 127, 32767, 8388607, 2147483647, 9223372036854775807, 255, 65535, 4294967295, 18446744073709551615],
        [3, 42, 1000, 100000, 1000000, 1000000000000, 200, 50000, 2000000000, 9000000000000000000],
    ];
    for r in rows {
        let mut args: Vec<Arg> = r[..6].iter().map(|v| Arg::Int64(*v as i64)).collect();
        args.extend(r[6..].iter().map(|v| Arg::Uint64(*v as u64)));
        exec(&mut db, "INSERT INTO int_types VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)", &args);
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'integer types')", &[]);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM int_types", &[]), rows.len() as i64);
    let sql = "SELECT COUNT(*) FROM int_types WHERE col_tinyint = -128 AND col_bigint = -9223372036854775808";
    assert_eq!(query_int(&mut db, sql, &[]), 1, "min boundary values must be stored exactly");
    let row = query_row(&mut db, "SELECT col_bigint_u::text FROM int_types WHERE col_tinyint = 127", &[]);
    assert_eq!(
        row,
        vec![Value::Text("18446744073709551615".into())],
        "max unsigned bigint boundary value must be stored exactly"
    );
    let mut db = gc_and_reconnect(&mut env, db);
    assert_eq!(
        query_int(&mut db, "SELECT COUNT(*) FROM int_types", &[]),
        rows.len() as i64,
        "integer rows must survive GC"
    );
    finish(env);
}

#[test]
#[allow(clippy::approx_constant)]
fn test_type_diversity_floating_point_and_decimal() {
    let mut env = env();
    setup_test_server(&mut env, "float_types_test");
    let mut db = db(&mut env);
    exec(
        &mut db,
        "CREATE TABLE float_types (
			id          BIGINT PRIMARY KEY,
			col_float   REAL,
			col_double  DOUBLE PRECISION,
			col_dec     DECIMAL(30,10)
		)",
        &[],
    );
    let rows: [(i64, f32, f64, &str); 6] = [
        (0, 0.0, 0.0, "0.0000000000"),
        (1, 1.5, 1.5, "1.5000000000"),
        (2, -1.5, -1.5, "-1.5000000000"),
        (3, 3.14159, 3.14159265358979, "3.1415926536"),
        (4, 1e10, 1e15, "99999999999999999999.9999999999"),
        (5, -1e10, -1e15, "-99999999999999999999.9999999999"),
    ];
    for (id, f, d, dec) in rows {
        let args = [Arg::Int(id), Arg::Float32(f), Arg::Float64(d), Arg::Str(dec.to_string())];
        exec(&mut db, "INSERT INTO float_types VALUES ($1,$2,$3,$4)", &args);
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'float types')", &[]);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM float_types", &[]), rows.len() as i64);
    let sql = "SELECT COUNT(*) FROM float_types WHERE col_dec = 3.1415926536";
    assert_eq!(query_int(&mut db, sql, &[]), 1, "decimal value must be stored and retrieved exactly");
    let mut db = gc_and_reconnect(&mut env, db);
    assert_eq!(
        query_int(&mut db, "SELECT COUNT(*) FROM float_types", &[]),
        rows.len() as i64,
        "float rows must survive GC"
    );
    finish(env);
}

#[test]
fn test_type_diversity_string_types() {
    let mut env = env();
    setup_test_server(&mut env, "string_types_test");
    let mut db = db(&mut env);
    exec(
        &mut db,
        "CREATE TABLE string_types (
			id            BIGINT PRIMARY KEY,
			col_char      CHAR(100),
			col_varchar   VARCHAR(2000),
			col_text      TEXT,
			col_medtext   TEXT,
			col_longtext  TEXT
		)",
        &[],
    );
    let text = |s: String| Some(Arg::Str(s));
    let values: Vec<(i64, [Option<Arg>; 5])> = vec![
        (0, [text("".into()), text("".into()), text("".into()), text("".into()), text("".into())]),
        (
            1,
            [
                text("hello".into()),
                text("world".into()),
                text("short text".into()),
                text("medium text".into()),
                text("long text".into()),
            ],
        ),
        (
            2,
            [
                text("日本語テスト".into()),
                text("Ünïcödé strïng wïth vàrïöüs chàrs: αβγδεζηθ ℕ ℤ ℚ ℝ ℂ".into()),
                text("中文测试内容-".repeat(100)),
                text("한국어 테스트 데이터-".repeat(500)),
                text("العربية اختبار البيانات-".repeat(1000)),
            ],
        ),
        (
            3,
            [
                text("x".repeat(100)),
                text("v".repeat(2000)),
                text(make_test_text(300, 12_000)),
                text(make_test_text(301, 50_000)),
                text(make_test_text(302, 200_000)),
            ],
        ),
        (4, [None, None, None, None, None]),
    ];
    let count = values.len() as i64;
    for (id, columns) in values {
        let mut args = vec![Arg::Int(id)];
        args.extend(columns.into_iter().map(opt));
        exec(&mut db, "INSERT INTO string_types VALUES ($1,$2,$3,$4,$5,$6)", &args);
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'string types')", &[]);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM string_types", &[]), count);
    let sql = "SELECT octet_length(col_longtext) FROM string_types WHERE id = 3";
    assert_eq!(query_int(&mut db, sql, &[]), 200_000, "200 KB TEXT value must be stored and retrieved intact");
    let sql = "SELECT COUNT(*) FROM string_types WHERE col_char IS NULL AND col_varchar IS NULL";
    assert_eq!(query_int(&mut db, sql, &[]), 1, "NULL string columns must be stored as NULL");
    let row = query_row(&mut db, "SELECT col_char FROM string_types WHERE id = 2", &[]);
    assert_eq!(
        row[0].to_go_string().trim_end_matches(' '),
        "日本語テスト",
        "unicode CHAR value must round-trip unchanged"
    );
    let mut db = gc_and_reconnect(&mut env, db);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM string_types", &[]), count, "string rows must survive GC");
    finish(env);
}

#[test]
fn test_type_diversity_binary_types() {
    let mut env = env();
    setup_test_server(&mut env, "binary_types_test");
    let mut db = db(&mut env);
    exec(
        &mut db,
        "CREATE TABLE binary_types (
			id            BIGINT PRIMARY KEY,
			col_binary    BYTEA,
			col_varbinary BYTEA,
			col_blob      BYTEA,
			col_longblob  BYTEA
		)",
        &[],
    );
    let bytes = |b: Vec<u8>| Some(Arg::Bytes(b));
    let values: Vec<(i64, [Option<Arg>; 4])> = vec![
        (0, [bytes(vec![0; 32]), bytes(Vec::new()), bytes(Vec::new()), bytes(Vec::new())]),
        (
            1,
            [
                bytes(make_test_blob_data(1, 32)),
                bytes(make_test_blob_data(2, 500)),
                bytes(make_test_blob_data(3, 8_000)),
                bytes(make_test_blob_data(4, 30_000)),
            ],
        ),
        (
            2,
            [
                bytes(make_test_blob_data(5, 32)),
                bytes(make_test_blob_data(6, 2000)),
                bytes(make_test_blob_data(7, 65_000)),
                bytes(make_test_blob_data(8, 500_000)),
            ],
        ),
        (3, [None, None, None, None]),
    ];
    let count = values.len() as i64;
    for (id, columns) in values {
        let mut args = vec![Arg::Int(id)];
        args.extend(columns.into_iter().map(opt));
        exec(&mut db, "INSERT INTO binary_types VALUES ($1,$2,$3,$4,$5)", &args);
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'binary types')", &[]);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM binary_types", &[]), count);
    let sql = "SELECT octet_length(col_longblob) FROM binary_types WHERE id = 2";
    assert_eq!(query_int(&mut db, sql, &[]), 500_000, "500 KB bytea value must be stored and retrieved intact");
    let row = query_row(&mut db, "SELECT col_longblob FROM binary_types WHERE id = 2", &[]);
    assert_eq!(
        row,
        vec![Value::Bytes(make_test_blob_data(8, 500_000))],
        "500 KB bytea round-trip must be byte-perfect"
    );
    let mut db = gc_and_reconnect(&mut env, db);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM binary_types", &[]), count, "binary rows must survive GC");
    finish(env);
}

#[test]
fn test_type_diversity_date_time_types() {
    let mut env = env();
    setup_test_server(&mut env, "datetime_types_test");
    let mut db = db(&mut env);
    exec(
        &mut db,
        "CREATE TABLE datetime_types (
			id           BIGINT PRIMARY KEY,
			col_date     DATE,
			col_time     TIME,
			col_datetime TIMESTAMP(6),
			col_ts       TIMESTAMP(6),
			col_year     INTEGER
		)",
        &[],
    );
    let s = |v: &str| Some(Arg::Str(v.to_string()));
    let rows: Vec<(i64, [Option<Arg>; 5])> = vec![
        (0, [s("1000-01-01"), s("00:00:00"), s("1000-01-01 00:00:00.000000"), None, Some(Arg::Int(1901))]),
        (
            1,
            [
                s("2024-06-15"),
                s("00:00:00"),
                s("2024-06-15 12:30:45.123456"),
                s("2024-06-15 12:30:45.123456"),
                Some(Arg::Int(2024)),
            ],
        ),
        (
            2,
            [
                s("9999-12-31"),
                s("23:59:59"),
                s("9999-12-31 23:59:59.999999"),
                s("2038-01-19 03:14:07.000000"),
                Some(Arg::Int(2155)),
            ],
        ),
        (3, [None, None, None, None, None]),
    ];
    let count = rows.len() as i64;
    for (id, columns) in rows {
        let mut args = vec![Arg::Int(id)];
        args.extend(columns.into_iter().map(opt));
        exec(&mut db, "INSERT INTO datetime_types VALUES ($1,$2,$3,$4,$5,$6)", &args);
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'datetime types')", &[]);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM datetime_types", &[]), count);
    let sql = "SELECT COUNT(*) FROM datetime_types WHERE col_date = '1000-01-01'";
    assert_eq!(query_int(&mut db, sql, &[]), 1, "minimum DATE value must round-trip correctly");
    let sql = "SELECT COUNT(*) FROM datetime_types WHERE col_date IS NULL AND col_time IS NULL";
    assert_eq!(query_int(&mut db, sql, &[]), 1, "NULL date/time values must be stored as NULL");
    let mut db = gc_and_reconnect(&mut env, db);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM datetime_types", &[]), count, "datetime rows must survive GC");
    finish(env);
}

#[test]
fn test_type_diversity_special_types() {
    let mut env = env();
    setup_test_server(&mut env, "special_types_test");
    let mut db = db(&mut env);
    exec(
        &mut db,
        "CREATE TABLE special_types (
			id        BIGINT PRIMARY KEY,
			col_bool  BOOLEAN,
			col_json  JSON,
			col_enum  VARCHAR(32),
			col_set   VARCHAR(64)
		)",
        &[],
    );
    let s = |v: &str| Some(Arg::Str(v.to_string()));
    let rows: Vec<(i64, [Option<Arg>; 4])> = vec![
        (0, [Some(Arg::Bool(false)), s("{}"), s("alpha"), s("red")]),
        (1, [Some(Arg::Bool(true)), s(r#"{"key": "value", "num": 42, "arr": [1,2,3]}"#), s("beta"), s("red,green")]),
        (2, [None, Some(Arg::Str(make_test_json_string(200, 15_000))), s("gamma"), s("red,green,blue")]),
        (3, [Some(Arg::Bool(false)), s("null"), s("delta"), s("red,green,blue,yellow")]),
        (4, [None, None, None, None]),
    ];
    let count = rows.len() as i64;
    for (id, columns) in rows {
        let mut args = vec![Arg::Int(id)];
        args.extend(columns.into_iter().map(opt));
        exec(&mut db, "INSERT INTO special_types VALUES ($1,$2,$3,$4,$5)", &args);
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'special types')", &[]);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM special_types", &[]), count);
    let row = query_row(&mut db, "SELECT col_json FROM special_types WHERE id = 2", &[]);
    let object: serde_json::Map<String, serde_json::Value> = serde_json::from_str(&row[0].to_go_string()).unwrap();
    assert!(!object.is_empty(), "large JSON value must be a non-empty object");
    let sql = "SELECT COUNT(*) FROM special_types WHERE col_enum = 'gamma' AND (',' || col_set || ',') LIKE '%,blue,%'";
    assert_eq!(query_int(&mut db, sql, &[]), 1, "ENUM and SET values must round-trip correctly");
    let sql = "SELECT COUNT(*) FROM special_types WHERE col_enum IS NULL AND col_set IS NULL";
    assert_eq!(query_int(&mut db, sql, &[]), 1, "NULL ENUM and SET values must be stored as NULL");
    let mut db = gc_and_reconnect(&mut env, db);
    assert_eq!(
        query_int(&mut db, "SELECT COUNT(*) FROM special_types", &[]),
        count,
        "special-type rows must survive GC"
    );
    finish(env);
}

#[test]
#[allow(clippy::approx_constant)]
fn test_type_diversity_null_values_across_all_types() {
    let mut env = env();
    setup_test_server(&mut env, "null_types_test");
    let mut db = db(&mut env);
    exec(
        &mut db,
        "CREATE TABLE nullable_types (
			id     BIGINT PRIMARY KEY,
			n_int  INTEGER,
			n_dbl  DOUBLE PRECISION,
			n_dec  DECIMAL(10,4),
			n_str  VARCHAR(255),
			n_txt  TEXT,
			n_blob BYTEA,
			n_date DATE,
			n_ts   TIMESTAMP(3),
			n_json JSON
		)",
        &[],
    );
    exec(&mut db, "INSERT INTO nullable_types VALUES (0, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL)", &[]);
    let args = [
        Arg::Int(1),
        Arg::Int(42),
        Arg::Float64(3.14),
        Arg::Str("2.7183".into()),
        Arg::Str("hello".into()),
        Arg::Str(make_test_text(99, 5000)),
        Arg::Bytes(make_test_blob_data(99, 5000)),
        Arg::Str("2024-03-15".into()),
        Arg::Str("2024-03-15 10:00:00.000".into()),
        Arg::Str(r#"{"x": 1}"#.into()),
    ];
    exec(&mut db, "INSERT INTO nullable_types VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)", &args);
    for i in 2..20i64 {
        let even = i % 2 == 0;
        let args = [
            Arg::Int(i),
            if even { Arg::Int(i * 100) } else { Arg::Null },
            if even { Arg::Str(format!("value-{i}")) } else { Arg::Null },
            if even { Arg::Str(make_test_text(i, 2000)) } else { Arg::Null },
            if even { Arg::Str("2024-01-01".into()) } else { Arg::Null },
        ];
        exec(&mut db, "INSERT INTO nullable_types (id, n_int, n_str, n_txt, n_date) VALUES ($1,$2,$3,$4,$5)", &args);
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'nullable types')", &[]);
    let sql = "SELECT COUNT(*) FROM nullable_types WHERE n_int IS NULL AND n_str IS NULL AND n_blob IS NULL";
    assert!(query_int(&mut db, sql, &[]) > 0, "rows with all-NULL values must be stored and counted correctly");
    let mut db = gc_and_reconnect(&mut env, db);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM nullable_types", &[]), 20, "all nullable rows must survive GC");
    finish(env);
}

#[test]
fn test_wide_table_many_int_columns() {
    const NUM_COLS: usize = 500;
    const NUM_ROWS: i64 = 30;
    let mut env = env();
    setup_test_server(&mut env, "wide_int_table");
    let mut db = db(&mut env);
    exec(&mut db, &build_wide_create_table("wide_int", NUM_COLS, "BIGINT"), &[]);
    for row in 0..NUM_ROWS {
        exec(&mut db, &build_wide_insert("wide_int", row, NUM_COLS, |col| (row * (col as i64 + 1)).to_string()), &[]);
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'wide int table')", &[]);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM wide_int", &[]), NUM_ROWS);
    let sql = "SELECT c250 FROM wide_int WHERE id = 5";
    assert_eq!(query_int(&mut db, sql, &[]), 5 * 251, "cell value at (row=5, col=c250) must be exact");
    let mut db = gc_and_reconnect(&mut env, db);
    assert_eq!(
        query_int(&mut db, "SELECT COUNT(*) FROM wide_int", &[]),
        NUM_ROWS,
        "wide int table must survive GC intact"
    );
    assert_eq!(query_int(&mut db, sql, &[]), 5 * 251, "cell value must be preserved after GC");
    finish(env);
}

#[test]
fn test_wide_table_many_varchar_columns() {
    const NUM_COLS: usize = 8;
    const COL_WIDTH: usize = 300;
    const NUM_ROWS: i64 = 20;
    let mut env = env();
    setup_test_server(&mut env, "wide_varchar_table");
    let mut db = db(&mut env);
    exec(&mut db, &build_wide_create_table("wide_varchar", NUM_COLS, &format!("VARCHAR({COL_WIDTH})")), &[]);
    for row in 0..NUM_ROWS {
        let stmt = build_wide_insert("wide_varchar", row, NUM_COLS, |col| {
            let prefix = format!("r{row:03}c{col:03}-");
            format!("'{}'", &prefix.repeat(COL_WIDTH.div_ceil(prefix.len()))[..COL_WIDTH])
        });
        exec(&mut db, &stmt, &[]);
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'wide varchar table')", &[]);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM wide_varchar", &[]), NUM_ROWS);
    let sql = format!(
        "SELECT COUNT(*) FROM wide_varchar WHERE octet_length(c0) = {COL_WIDTH} AND octet_length(c{}) = {COL_WIDTH}",
        NUM_COLS - 1
    );
    assert_eq!(query_int(&mut db, &sql, &[]), NUM_ROWS, "all VARCHAR cells must retain their full width");
    let cell = query_row(&mut db, "SELECT c4 FROM wide_varchar WHERE id = 10", &[])[0].to_go_string();
    assert!(cell.starts_with("r010c004-"), "VARCHAR cell content must encode the correct (row, col) position");
    let mut db = gc_and_reconnect(&mut env, db);
    assert_eq!(
        query_int(&mut db, "SELECT COUNT(*) FROM wide_varchar", &[]),
        NUM_ROWS,
        "wide varchar table must survive GC intact"
    );
    finish(env);
}

#[test]
fn test_wide_table_many_text_columns() {
    const NUM_COLS: usize = 100;
    const COL_DATA_SIZE: usize = 2_000;
    const NUM_ROWS: i64 = 10;
    const CHECK_ROW: i64 = 3;
    const CHECK_COL: i64 = 47;
    let mut env = env();
    setup_test_server(&mut env, "wide_text_table");
    let mut db = db(&mut env);
    exec(&mut db, &build_wide_create_table("wide_text", NUM_COLS, "TEXT"), &[]);
    for row in 0..NUM_ROWS {
        let (stmt, args) = build_wide_text_insert("wide_text", row, NUM_COLS, COL_DATA_SIZE);
        exec(&mut db, &stmt, &args);
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'wide text table')", &[]);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM wide_text", &[]), NUM_ROWS);
    let sql = format!(
        "SELECT COUNT(*) FROM wide_text WHERE octet_length(c0) = {COL_DATA_SIZE} AND octet_length(c{}) = {COL_DATA_SIZE}",
        NUM_COLS - 1
    );
    assert_eq!(query_int(&mut db, &sql, &[]), NUM_ROWS, "all TEXT cells must retain their content length");
    let check = format!("SELECT c{CHECK_COL} FROM wide_text WHERE id = {CHECK_ROW}");
    let expected = vec![Value::Text(make_test_text(CHECK_ROW * NUM_COLS as i64 + CHECK_COL, COL_DATA_SIZE))];
    assert_eq!(
        query_row(&mut db, &check, &[]),
        expected,
        "TEXT cell content must be byte-perfect after storage and retrieval"
    );
    let mut db = gc_and_reconnect(&mut env, db);
    assert_eq!(
        query_int(&mut db, "SELECT COUNT(*) FROM wide_text", &[]),
        NUM_ROWS,
        "wide text table must survive GC intact"
    );
    assert_eq!(query_row(&mut db, &check, &[]), expected, "TEXT cell content must be preserved through GC");
    finish(env);
}

#[test]
fn test_wide_table_extremely_wide_row() {
    const NUM_COLS: usize = 200;
    const COL_DATA_SIZE: usize = 5_000;
    const NUM_ROWS: i64 = 3;
    const CHECK_ROW: i64 = 1;
    const CHECK_COL: i64 = 100;
    let mut env = env();
    setup_test_server(&mut env, "extreme_wide_row");
    let mut db = db(&mut env);
    exec(&mut db, &build_wide_create_table("extreme_wide", NUM_COLS, "TEXT"), &[]);
    for row in 0..NUM_ROWS {
        let (stmt, args) = build_wide_text_insert("extreme_wide", row, NUM_COLS, COL_DATA_SIZE);
        exec(&mut db, &stmt, &args);
    }
    exec(&mut db, "SELECT dolt_commit('-Am', 'extreme wide rows')", &[]);
    assert_eq!(query_int(&mut db, "SELECT COUNT(*) FROM extreme_wide", &[]), NUM_ROWS);
    let sql = format!(
        "SELECT COUNT(*) FROM extreme_wide WHERE octet_length(c0) = {COL_DATA_SIZE} AND octet_length(c{}) = {COL_DATA_SIZE}",
        NUM_COLS - 1
    );
    assert_eq!(query_int(&mut db, &sql, &[]), NUM_ROWS, "extremely wide rows must fully preserve all column data");
    let mut db = gc_and_reconnect(&mut env, db);
    assert_eq!(
        query_int(&mut db, "SELECT COUNT(*) FROM extreme_wide", &[]),
        NUM_ROWS,
        "extremely wide rows must survive GC"
    );
    let check = format!("SELECT c{CHECK_COL} FROM extreme_wide WHERE id = {CHECK_ROW}");
    let expected = vec![Value::Text(make_test_text(CHECK_ROW * NUM_COLS as i64 + CHECK_COL, COL_DATA_SIZE))];
    assert_eq!(query_row(&mut db, &check, &[]), expected, "extreme wide row TEXT cell must be byte-perfect after GC");
    finish(env);
}
